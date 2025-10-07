import 'package:flutter/material.dart';
import 'package:flutter_mobx/flutter_mobx.dart';
import 'package:provider/provider.dart';

import '../../../../../../core/di/di.dart';
import '../../../../application/stores/home_ui_store.dart';
import '../../../../application/stores/sessions_store.dart';
import '../../../../application/stores/session_details_store.dart';
import '../../../widgets/details/details_tabs.dart';

class DetailsContainer extends StatelessWidget {
  const DetailsContainer({super.key, required this.namespaceCtrl});
  final TextEditingController namespaceCtrl;

  @override
  Widget build(BuildContext context) {
    final ui = sl<HomeUiStore>();
    if (ui.selectedSessionId.value == null) return const SizedBox.shrink();
    bool selIsWs = true;
    bool selIsHttp = true;
    final items = context.watch<SessionsStore>().items.toList();
    Map<String, dynamic>? meta;
    String? kind;
    for (final s in items) {
      if (s.id == ui.selectedSessionId.value) {
        meta = s.httpMeta?.cast<String, dynamic>();
        kind = s.kind;
        break;
      }
    }
    final method = (meta?['method'] ?? '').toString();
    final isWs = (kind == 'ws') || (method.isEmpty && kind == null);
    selIsWs = isWs;
    selIsHttp = !isWs;

    return Observer(
      builder: (_) {
        final details = context.watch<SessionDetailsStore>();
        final frames =
            details.frames
                .map(
                  (f) => {
                    'id': f.id,
                    'ts': f.ts.toIso8601String(),
                    'direction': f.direction,
                    'opcode': f.opcode,
                    'size': f.size,
                    'preview': f.preview,
                  },
                )
                .toList();
        final events =
            details.events
                .map(
                  (e) => {
                    'id': e.id,
                    'ts': e.ts.toIso8601String(),
                    'namespace': e.namespace,
                    'event': e.event,
                    'ackId': e.ackId,
                    'argsPreview': e.argsPreview,
                  },
                )
                .toList();
        return DetailsTabs(
          showWs: selIsWs,
          showHttp: selIsHttp,
          frames: frames.cast<Map<String, dynamic>>(),
          events: events.cast<Map<String, dynamic>>(),
          selectedSessionId: ui.selectedSessionId.value,
          httpMeta: sl<HomeUiStore>().httpMeta[ui.selectedSessionId.value],
          opcodeFilter: sl<HomeUiStore>().opcodeFilter.value,
          directionFilter: sl<HomeUiStore>().directionFilter.value,
          namespaceCtrl: namespaceCtrl,
          onChangeOpcode: (v) {
            sl<HomeUiStore>().setOpcodeFilter(v);
          },
          onChangeDirection: (v) {
            sl<HomeUiStore>().setDirectionFilter(v);
          },
          hideHeartbeats: sl<HomeUiStore>().hideHeartbeats.value,
          onToggleHeartbeats: (v) {
            sl<HomeUiStore>().setHideHeartbeats(v);
          },
        );
      },
    );
  }
}
