import 'package:flutter/material.dart';

import '../../../../ws_inspector/presentation/widgets/ws_details_panel.dart';
import '../../../../http_inspector/presentation/widgets/http_details_panel.dart';
import '../../../application/stores/home_ui_store.dart';
import '../../../../../core/di/di.dart';

// Компонент-обёртка над табами деталей сессии.
// Делает логику выбора вкладок компактной и не тянет state в родителя.
class DetailsTabs extends StatelessWidget {
  const DetailsTabs({
    super.key,
    required this.showWs,
    required this.showHttp,
    required this.frames,
    required this.events,
    required this.selectedSessionId,
    required this.httpMeta,
    required this.opcodeFilter,
    required this.directionFilter,
    required this.namespaceCtrl,
    required this.onChangeOpcode,
    required this.onChangeDirection,
    required this.hideHeartbeats,
    required this.onToggleHeartbeats,
  });

  // Какие вкладки показывать
  final bool showWs;
  final bool showHttp;

  // Данные
  final List<Map<String, dynamic>> frames;
  final List<Map<String, dynamic>> events;
  final String? selectedSessionId;
  final Map<String, dynamic>? httpMeta;

  // Фильтры WS
  final String opcodeFilter;
  final String directionFilter;
  final TextEditingController namespaceCtrl;
  final ValueChanged<String> onChangeOpcode;
  final ValueChanged<String> onChangeDirection;
  final bool hideHeartbeats;
  final ValueChanged<bool> onToggleHeartbeats;

  @override
  Widget build(BuildContext context) {
    final ui = sl<HomeUiStore>();
    final tabsCount =
        (() {
          if (showWs && showHttp) return 2;
          if (showWs || showHttp) return 1;
          return 2; // дефолт на два, как было ранее
        })();

    final tabBar =
        (showWs && showHttp)
            ? const TabBar(tabs: [Tab(text: 'WebSocket'), Tab(text: 'HTTP')])
            : const SizedBox.shrink();

    final initialIndex =
        (() {
          if (!showWs || !showHttp) return 0;
          final sid = selectedSessionId;
          if (sid == null) return 0;
          final saved = ui.getSessionTab(sid);
          if (saved == 'http') return 1;
          return 0;
        })();

    return DefaultTabController(
      initialIndex: initialIndex,
      length: tabsCount,
      child: Column(
        children: [
          tabBar,
          Expanded(
            child: Builder(
              builder: (context) {
                if (showWs && showHttp) {
                  return TabBarView(children: [_buildWs(), _buildHttp()]);
                }
                if (showWs) return _buildWs();
                return _buildHttp();
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildWs() {
    // Сохраняем выбранную вкладку как 'ws'
    final sid = selectedSessionId;
    if (sid != null) {
      sl<HomeUiStore>().setSessionTab(sid, 'ws');
    }
    return WsDetailsPanel(
      frames: frames,
      events: events,
      opcodeFilter: opcodeFilter,
      directionFilter: directionFilter,
      namespaceCtrl: namespaceCtrl,
      onChangeOpcode: onChangeOpcode,
      onChangeDirection: onChangeDirection,
      hideHeartbeats: hideHeartbeats,
      onToggleHeartbeats: onToggleHeartbeats,
    );
  }

  Widget _buildHttp() {
    // Сохраняем выбранную вкладку как 'http'
    final sid = selectedSessionId;
    if (sid != null) {
      sl<HomeUiStore>().setSessionTab(sid, 'http');
    }
    return HttpDetailsPanel(
      sessionId: selectedSessionId,
      frames: frames,
      httpMeta: httpMeta,
    );
  }
}
