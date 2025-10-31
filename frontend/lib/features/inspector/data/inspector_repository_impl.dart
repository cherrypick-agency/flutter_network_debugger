import 'package:app_http_client/application/app_http_client.dart';
import '../../../core/di/di.dart';
import '../../../core/network/error_utils.dart';
import '../../inspector/application/stores/home_ui_store.dart';
import '../domain/repositories/inspector_repository.dart';
import '../domain/entities/session.dart';
import '../domain/entities/frame.dart';
import '../domain/entities/event.dart';

class InspectorRepositoryImpl implements InspectorRepository {
  InspectorRepositoryImpl(this._api);
  final AppHttpClient _api;

  @override
  Future<List<Session>> listSessions({
    String? q,
    String? target,
    Set<String>? types,
    Set<String>? statusGroups,
  }) async {
    try {
      // Build capture scope params via HomeUiStore
      String scope = '';
      bool includePaused = false;
      try {
        final ui = sl<HomeUiStore>();
        scope = ui.captureScope.value;
        includePaused = ui.includePaused.value;
      } catch (_) {}
      // Собираем быстрые фильтры напрямую из UI-стора, чтобы не плодить параметры
      Set<String> uiTypes = const {};
      Set<String> uiStatusGroups = const {};
      try {
        final ui = sl<HomeUiStore>();
        uiTypes = ui.quickTypes.toSet();
        uiStatusGroups = ui.quickStatusGroups.toSet();
      } catch (_) {}

      final res = await _api.get(
        host: null,
        path: '/_api/v1/sessions',
        query: {
          'limit': '100',
          if (q != null && q.isNotEmpty) 'q': q,
          if (target != null && target.isNotEmpty) '_target': target,
          if (scope == 'all') 'captures': 'all',
          if (scope != 'all') 'captureId': 'current',
          if (includePaused) 'includeUnassigned': 'true',
          if ((types != null && types.isNotEmpty) || uiTypes.isNotEmpty)
            'types': (types ?? uiTypes).join(','),
          if ((statusGroups != null && statusGroups.isNotEmpty) ||
              uiStatusGroups.isNotEmpty)
            'status': (statusGroups ?? uiStatusGroups).join(','),
          // Включаем глубокую проверку GraphQL только при явном выборе
          if (((types ?? uiTypes).contains('graphql'))) 'scan': 'graphql',
        },
      );
      final data =
          (res.data is Map<String, dynamic>)
              ? (res.data as Map<String, dynamic>)
              : {};
      final items = ((data['items'] as List?) ?? const [])
          .whereType<Map<String, dynamic>>()
          .toList(growable: false);
      return items
          .map(
            (m) => Session(
              id: m['id'] as String,
              target: m['target'] as String,
              clientAddr: m['clientAddr'] as String?,
              startedAt:
                  m['startedAt'] != null
                      ? DateTime.tryParse(m['startedAt'].toString())
                      : null,
              closedAt:
                  m['closedAt'] != null
                      ? DateTime.tryParse(m['closedAt'].toString())
                      : null,
              error: m['error']?.toString(),
              kind: m['kind']?.toString(),
              httpMeta: (m['httpMeta'] as Map?)?.cast<String, dynamic>(),
              sizes: (m['sizes'] as Map?)?.cast<String, dynamic>(),
              isSocketIo:
                  (() {
                    try {
                      final ev = (m['events'] as Map?)?.cast<String, dynamic>();
                      final sio =
                          int.tryParse((ev?['sio'] ?? '0').toString()) ?? 0;
                      return sio > 0;
                    } catch (_) {
                      return false;
                    }
                  })(),
            ),
          )
          .toList(growable: false);
    } catch (e) {
      // Пробрасываем только сообщение; UI решит, как показывать
      throw resolveErrorMessage(e);
    }
  }

  @override
  Future<List<Frame>> listFrames(
    String sessionId, {
    String? from,
    int limit = 100,
  }) async {
    try {
      final res = await _api.get(
        path: '/_api/v1/sessions/$sessionId/frames',
        query: {
          'limit': '$limit',
          if (from != null && from.isNotEmpty) 'from': from,
        },
      );
      final data =
          (res.data is Map<String, dynamic>)
              ? (res.data as Map<String, dynamic>)
              : {};
      final items = ((data['items'] as List?) ?? const [])
          .whereType<Map<String, dynamic>>()
          .toList(growable: false);
      return items
          .map(
            (m) => Frame(
              id: m['id'] as String,
              ts: DateTime.parse(m['ts'] as String),
              direction: m['direction'] as String,
              opcode: m['opcode'] as String,
              size: (m['size'] as num).toInt(),
              preview: m['preview'] as String,
            ),
          )
          .toList(growable: false);
    } catch (e) {
      throw resolveErrorMessage(e);
    }
  }

  @override
  Future<List<EventEntity>> listEvents(
    String sessionId, {
    String? from,
    int limit = 100,
  }) async {
    try {
      final res = await _api.get(
        path: '/_api/v1/sessions/$sessionId/events',
        query: {
          'limit': '$limit',
          if (from != null && from.isNotEmpty) 'from': from,
        },
      );
      final data =
          (res.data is Map<String, dynamic>)
              ? (res.data as Map<String, dynamic>)
              : {};
      final items = ((data['items'] as List?) ?? const [])
          .whereType<Map<String, dynamic>>()
          .toList(growable: false);
      return items
          .map(
            (m) => EventEntity(
              id: m['id'] as String,
              ts: DateTime.parse(m['ts'] as String),
              namespace: (m['namespace'] ?? '').toString(),
              event: (m['event'] ?? '').toString(),
              ackId: m['ackId'] == null ? null : (m['ackId'] as num).toInt(),
              argsPreview: (m['argsPreview'] ?? '').toString(),
            ),
          )
          .toList(growable: false);
    } catch (e) {
      throw resolveErrorMessage(e);
    }
  }

  @override
  Future<List<Map<String, dynamic>>> aggregateSessions({
    String groupBy = 'domain',
  }) async {
    final res = await _api.get(
      path: '/_api/v1/sessions/aggregate',
      query: {'groupBy': groupBy},
    );
    final data =
        (res.data is Map<String, dynamic>)
            ? (res.data as Map<String, dynamic>)
            : {};
    final groups =
        (data['groups'] as List<dynamic>? ?? const [])
            .cast<Map<String, dynamic>>();
    return groups;
  }
}
