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
    required this.wsClosed,
    required this.wsClosedAt,
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
  final bool wsClosed;
  final DateTime? wsClosedAt;

  @override
  Widget build(BuildContext context) {
    final ui = sl<HomeUiStore>();
    // Если выбрана HTTP‑сессия — показываем сразу HTTP без вкладок.
    // Если выбрана WS‑сессия — показываем WS без вкладок.
    // Хранение выбранной вкладки оставляем ради совместимости истории.
    if (showHttp && !showWs) {
      final sid = selectedSessionId;
      if (sid != null) {
        ui.setSessionTab(sid, 'http');
      }
      return _buildHttp();
    }
    if (showWs && !showHttp) {
      final sid = selectedSessionId;
      if (sid != null) {
        ui.setSessionTab(sid, 'ws');
      }
      return _buildWs();
    }
    // Фолбэк на случай неконсистентного состояния.
    return const SizedBox.shrink();
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
      isClosed: wsClosed,
      closedAt: wsClosedAt,
    );
  }

  Widget _buildHttp() {
    return HttpDetailsPanel(
      sessionId: selectedSessionId,
      frames: frames,
      httpMeta: httpMeta,
    );
  }
}
