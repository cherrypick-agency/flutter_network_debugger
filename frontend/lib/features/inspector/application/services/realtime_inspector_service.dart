import '../../../../core/di/di.dart';
import '../../../../core/realtime/socket_io_service.dart';
import '../../domain/entities/session.dart';
import '../stores/sessions_store.dart';
import '../stores/home_ui_store.dart';
import '../../../filters/application/stores/sessions_filters_store.dart';
import '../stores/aggregate_store.dart';

class RealtimeInspectorService {
  late final SocketIoService _sio = sl<SocketIoService>();
  bool _inited = false;
  bool _fallbackUsed = false;

  Future<void> init() async {
    if (_inited) {
      print('[RealtimeInspector] init() called but already inited');
      return;
    }
    print('[RealtimeInspector] init() starting, calling _sio.connect()...');
    await _sio.connect();
    _inited = true;
    print('[RealtimeInspector] init() completed');
  }

  Future<void> resubscribeWithCurrentFilters() async {
    print('[RealtimeInspector] 🔄 resubscribeWithCurrentFilters() called');
    final ui = sl<HomeUiStore>();
    final filters = sl<SessionsFiltersStore>();
    final payload = {
      'q': ui.sessionSearchQuery.value,
      'target': filters.target.trim(),
      'types': ui.quickTypes.toList(),
      'status': ui.quickStatusGroups.toList(),
      'captureScope': ui.captureScope.value, // current | all
      'includeUnassigned': ui.includePaused.value,
      'limit': 1000,
      'groupBy': 'domain',
      if (filters.selectedTags.isNotEmpty)
        'tags': filters.selectedTags.toList(),
    };
    print('[RealtimeInspector] Payload prepared: $payload');
    // Если запись выключена и «в паузе не показывать», то отбрасываем все
    // сессии без captureId на клиенте (доп. защита на случай гонок/дефолтов сервера)
    final bool dropUnassignedWhilePaused =
        !ui.isRecording.value && !ui.includePaused.value;
    final DateTime? pausedSince = ui.pausedSince.value;
    final bool dropStartedAfterPaused =
        pausedSince != null && !ui.includePaused.value;

    // ВАЖНО: Сначала создаем socket (если еще нет), ПОТОМ регистрируем handlers
    print('[RealtimeInspector] Calling init() to ensure socket exists...');
    await init();
    print('[RealtimeInspector] Socket exists, now registering handlers...');

    print('[RealtimeInspector] Removing old handlers...');
    _sio.off('sessions:init');
    _sio.off('sessions:upsert');
    _sio.off('sessions:remove');
    _sio.off('aggregate:update');
    _sio.off('connect_error');
    _sio.off('disconnect');
    _sio.off('connect');

    print('[RealtimeInspector] Registering new handlers...');

    _sio.on('sessions:init', (data) {
      print(
        '[RealtimeInspector] 📥 sessions:init handler CALLED with data: $data',
      );
      try {
        final rawItems = (data['items'] as List? ?? const [])
            .whereType<Map>()
            .cast<Map<String, dynamic>>();
        Iterable<Map<String, dynamic>> filteredRaw = rawItems;
        if (dropUnassignedWhilePaused) {
          filteredRaw = filteredRaw.where((m) => m['captureId'] != null);
        }
        if (dropStartedAfterPaused) {
          filteredRaw = filteredRaw.where((m) {
            final ts = m['startedAt'];
            if (ts == null) return true;
            try {
              final t = DateTime.tryParse(ts.toString());
              if (t == null) return true;
              return !t.isAfter(pausedSince);
            } catch (_) {
              return true;
            }
          });
        }
        final list = filteredRaw.map(_mapSession).toList(growable: false);
        print(
          '[RealtimeInspector] sessions:init parsed ${list.length} sessions',
        );
        sl<SessionsStore>().applyInit(list);
        final groups = (data['aggregate'] as List? ?? const [])
            .whereType<Map>()
            .cast<Map<String, dynamic>>()
            .toList(growable: false);
        print(
          '[RealtimeInspector] sessions:init parsed ${groups.length} aggregates',
        );
        sl<AggregateStore>().applyAggregate(groups);
        print('[RealtimeInspector] sessions:init processing completed');
      } catch (e) {
        print('[RealtimeInspector] ❌ sessions:init error: $e');
      }
    });
    _sio.on('sessions:upsert', (data) {
      print('[RealtimeInspector] 📥 sessions:upsert handler CALLED');
      try {
        if (data is Map) {
          if (dropUnassignedWhilePaused) {
            final cap = data['captureId'];
            if (cap == null) {
              // пропускаем «непривязанные» во время паузы
              return;
            }
          }
          if (dropStartedAfterPaused) {
            final ts = data['startedAt'];
            if (ts != null) {
              try {
                final t = DateTime.tryParse(ts.toString());
                if (t != null && t.isAfter(pausedSince)) {
                  return;
                }
              } catch (_) {}
            }
          }
          sl<SessionsStore>().upsert(_mapSession(data.cast<String, dynamic>()));
          print('[RealtimeInspector] sessions:upsert processed');
        }
      } catch (e) {
        print('[RealtimeInspector] ❌ sessions:upsert error: $e');
      }
    });
    _sio.on('sessions:remove', (data) {
      print('[RealtimeInspector] 📥 sessions:remove handler CALLED');
      try {
        if (data is Map && data['id'] is String) {
          sl<SessionsStore>().removeById(data['id'] as String);
          print('[RealtimeInspector] sessions:remove processed');
        }
      } catch (e) {
        print('[RealtimeInspector] ❌ sessions:remove error: $e');
      }
    });
    _sio.on('aggregate:update', (data) {
      print('[RealtimeInspector] 📥 aggregate:update handler CALLED');
      try {
        final groups = (data as List? ?? const [])
            .whereType<Map>()
            .cast<Map<String, dynamic>>()
            .toList(growable: false);
        sl<AggregateStore>().applyAggregate(groups);
        print(
          '[RealtimeInspector] aggregate:update processed ${groups.length} groups',
        );
      } catch (e) {
        print('[RealtimeInspector] ❌ aggregate:update error: $e');
      }
    });

    // Простой фолбэк: если сокет недоступен — разово подхватим REST
    _sio.on('connect_error', (_) {
      print('[RealtimeInspector] 📥 connect_error handler CALLED');
      _restFallbackOnce();
    });
    _sio.on('disconnect', (_) {
      print('[RealtimeInspector] 📥 disconnect handler CALLED');
      _restFallbackOnce();
    });

    // Обеспечим подписку при переподключении
    _sio.on('connect', (_) {
      print(
        '[RealtimeInspector] 📥 connect handler CALLED, emitting sessions:subscribe',
      );
      _fallbackUsed = false; // сбросим флаг после успешного переподключения
      _sio.emit('sessions:subscribe', payload);
    });

    print('[RealtimeInspector] All handlers registered');

    // Only emit immediately if already connected, otherwise wait for connect handler
    if (_sio.connected) {
      print(
        '[RealtimeInspector] Already connected, emitting sessions:subscribe immediately...',
      );
      _sio.emit('sessions:subscribe', payload);
    } else {
      print(
        '[RealtimeInspector] Not connected yet, will emit when connect handler fires',
      );
    }
    print('[RealtimeInspector] resubscribeWithCurrentFilters() completed');
  }

  void _restFallbackOnce() {
    if (_fallbackUsed) return;
    _fallbackUsed = true;
    final ui = sl<HomeUiStore>();
    final filters = sl<SessionsFiltersStore>();
    // ignore: discarded_futures
    sl<SessionsStore>()
        .load(
          q: ui.sessionSearchQuery.value.trim(),
          target: filters.target.trim(),
          tags: filters.selectedTags.isNotEmpty
              ? filters.selectedTags.toList()
              : null,
        )
        .whenComplete(() {
          // ignore: discarded_futures
          sl<AggregateStore>().load(groupBy: 'domain');
        });
  }

  Session _mapSession(Map<String, dynamic> m) {
    DateTime? parseTime(dynamic v) {
      if (v == null) return null;
      try {
        return DateTime.tryParse(v.toString());
      } catch (_) {
        return null;
      }
    }

    return Session(
      id: (m['id'] ?? '').toString(),
      target: (m['target'] ?? '').toString(),
      clientAddr: (m['clientAddr']?.toString()),
      startedAt: parseTime(m['startedAt']),
      closedAt: parseTime(m['closedAt']),
      error: m['error']?.toString(),
      kind: m['kind']?.toString(),
      httpMeta: (m['httpMeta'] as Map?)?.cast<String, dynamic>(),
      sizes: (m['sizes'] as Map?)?.cast<String, dynamic>(),
      isSocketIo: false,
      processInfo: _parseProcessInfo(m['processInfo']),
    );
  }

  ProcessInfo? _parseProcessInfo(dynamic data) {
    if (data == null) return null;
    try {
      final m = data as Map<String, dynamic>;
      final pid = m['PID'] as int? ?? m['pid'] as int? ?? 0;
      final name = m['Name'] as String? ?? m['name'] as String? ?? '';

      if (pid == 0 || name.isEmpty) return null;

      return ProcessInfo(
        pid: pid,
        name: name,
        executablePath:
            m['ExecutablePath'] as String? ?? m['executablePath'] as String?,
        bundleId: m['BundleID'] as String? ?? m['bundleId'] as String?,
        icon: _parseProcessIcon(m['Icon'] ?? m['icon']),
        detectedAt: m['DetectedAt'] != null || m['detectedAt'] != null
            ? DateTime.tryParse((m['DetectedAt'] ?? m['detectedAt']).toString())
            : null,
      );
    } catch (e) {
      return null;
    }
  }

  ProcessIcon? _parseProcessIcon(dynamic data) {
    if (data == null) return null;
    try {
      final m = data as Map<String, dynamic>;
      final format = m['Format'] as String? ?? m['format'] as String? ?? '';
      final dataBase64 = m['Data'] as String? ?? m['data'] as String? ?? '';

      if (format.isEmpty || dataBase64.isEmpty) return null;

      return ProcessIcon(format: format, data: dataBase64);
    } catch (e) {
      return null;
    }
  }
}
