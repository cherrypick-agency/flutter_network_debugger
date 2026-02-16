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
    return Observer(
      builder: (_) {
        final selectedId = ui.selectedSessionId.value;
        if (selectedId == null) return const SizedBox.shrink();

        final items = context.watch<SessionsStore>().items.toList();
        Map<String, dynamic>? meta;
        String? kind;
        bool wsClosed = false;
        DateTime? wsClosedAt;
        String? wsError;
        for (final s in items) {
          if (s.id == selectedId) {
            meta = s.httpMeta?.cast<String, dynamic>();
            kind = s.kind;
            wsClosed = s.closedAt != null;
            wsClosedAt = s.closedAt;
            wsError = s.error;
            break;
          }
        }

        final method = (meta?['method'] ?? '').toString();
        final isFirebase = kind == 'firebase_database';
        final isWs =
            isFirebase || (kind == 'ws') || (method.isEmpty && kind == null);

        final details = context.watch<SessionDetailsStore>();
        final frames = details.frames
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
        final events = details.events
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
          showWs: isWs,
          showHttp: !isWs,
          frames: frames.cast<Map<String, dynamic>>(),
          events: events.cast<Map<String, dynamic>>(),
          selectedSessionId: selectedId,
          httpMeta: sl<HomeUiStore>().httpMeta[selectedId],
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
          wsClosed: wsClosed,
          wsClosedAt: wsClosedAt,
          wsError: wsError,
          fbOpFilter: ui.fbOpFilter.value,
          fbStatusFilter: ui.fbStatusFilter.value,
          fbPathFilter: ui.fbPathFilter.value,
          onChangeFbOp: (v) => ui.setFbOpFilter(v),
          onChangeFbStatus: (v) => ui.setFbStatusFilter(v),
          onChangeFbPath: (v) => ui.setFbPathFilter(v),
        );
      },
    );
  }
}
