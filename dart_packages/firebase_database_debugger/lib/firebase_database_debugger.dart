/// Библиотека для перехвата и отправки операций Firebase Realtime Database
/// в network-debugger для визуальной отладки.
///
/// Оборачивает стандартные [DatabaseReference] и [Query] в отладочные обёртки,
/// которые логируют каждую операцию (get, set, update, remove, listen)
/// и батчами отправляют фреймы в ingest API дебаггера.
library firebase_database_debugger;

import 'dart:async';
import 'dart:convert';
import 'dart:math';

import 'package:firebase_database/firebase_database.dart';
import 'package:http/http.dart' as http;

import 'src/firebase_database_debugger_config.dart';
import 'src/firebase_ingest_client.dart';
import 'src/models.dart';

export 'src/firebase_database_debugger_config.dart';

/// Основной класс дебаггера Firebase Realtime Database.
///
/// Создаёт отладочные обёртки над [DatabaseReference] и [Query],
/// буферизует фреймы и периодически отправляет их в ingest API.
///
/// Пример использования:
/// ```dart
/// final debugger = FirebaseDatabaseDebugger(
///   config: FirebaseDatabaseDebuggerConfig(
///     debuggerBaseUrl: 'http://localhost:9092',
///   ),
/// );
/// final ref = debugger.ref(FirebaseDatabase.instance.ref('users/123'));
/// await ref.set({'name': 'John'});
/// ```
class FirebaseDatabaseDebugger {
  /// Создаёт экземпляр дебаггера.
  ///
  /// [config] — конфигурация подключения и батчинга.
  /// [httpClient] — опциональный HTTP-клиент (удобно для тестов).
  FirebaseDatabaseDebugger({
    required FirebaseDatabaseDebuggerConfig config,
    http.Client? httpClient,
  })  : _config = config,
        _client = FirebaseIngestClient(config: config, httpClient: httpClient),
        _runId = _makeRunId();

  final FirebaseDatabaseDebuggerConfig _config;
  final FirebaseIngestClient _client;
  final String _runId;
  final Map<String, _SessionBuffer> _buffers = <String, _SessionBuffer>{};
  Timer? _flushTimer;
  int _seq = 0;

  /// Оборачивает [reference] в отладочную обёртку [DebugDatabaseReference].
  ///
  /// Все операции через обёртку будут логироваться и отправляться в дебаггер.
  DebugDatabaseReference ref(DatabaseReference reference) {
    return DebugDatabaseReference._(owner: this, inner: reference);
  }

  /// Оборачивает [value] (Firebase Query) в отладочную обёртку [DebugQuery].
  ///
  /// Все операции через обёртку будут логироваться и отправляться в дебаггер.
  DebugQuery query(Query value) {
    return DebugQuery._(owner: this, inner: value);
  }

  /// Принудительно отправляет все накопленные фреймы в дебаггер.
  Future<void> flush() async {
    final sessions = _buffers.values.toList(growable: false);
    _buffers.clear();
    for (final session in sessions) {
      await _flushSession(session);
    }
  }

  /// Останавливает таймер, отправляет оставшиеся фреймы и закрывает HTTP-клиент.
  ///
  /// Вызывайте при завершении работы приложения или когда дебаггер больше не нужен.
  Future<void> dispose() async {
    _flushTimer?.cancel();
    _flushTimer = null;
    await flush();
    _client.dispose();
  }

  /// Логирует одну операцию Firebase и добавляет фрейм в буфер.
  ///
  /// [path] — путь в базе данных (например, `/users/123`).
  /// [op] — тип операции (`set`, `get`, `update`, `remove`, `onValue` и т.д.).
  /// [direction] — направление: `client->upstream` или `upstream->client`.
  /// [payload] — данные операции (значение, длительность и т.д.).
  /// [query] — строка запроса (если операция через Query).
  /// [ok] — успешна ли операция.
  /// [error] — объект ошибки (если операция завершилась с ошибкой).
  Future<void> logOperation({
    required String path,
    required String op,
    required String direction,
    required dynamic payload,
    String? query,
    bool ok = true,
    Object? error,
  }) async {
    if (!_config.enabled) return;
    final now = DateTime.now().toUtc();
    final (sessionPath, sessionQuery) =
        _sessionTargetParts(path: path, query: query);
    final sessionId = _sessionIdFor(path: sessionPath, query: sessionQuery);
    final target = _targetFor(path: sessionPath, query: sessionQuery);

    final (preview, body, bodyEncoding) = _buildFramePayload(
      op: op,
      path: path,
      query: query,
      payload: payload,
      ok: ok,
      error: error,
      ts: now,
    );

    final frame = FirebaseIngestFrame(
      id: _nextFrameId(op),
      ts: now,
      direction: direction,
      opcode: 'text',
      preview: preview,
      body: body,
      bodyEncoding: bodyEncoding,
    );

    final buffer = _buffers.putIfAbsent(
      sessionId,
      () => _SessionBuffer(
        session: FirebaseIngestSession(
          id: sessionId,
          target: target,
          captureId: _config.captureId,
        ),
      ),
    );
    buffer.frames.add(frame);
    if (buffer.frames.length >= _config.maxBatchFrames) {
      await _flushSession(buffer);
      _buffers.remove(sessionId);
      return;
    }
    _scheduleFlush();
  }

  /// Помечает сессию как закрытую (например, при отписке от слушателя).
  ///
  /// [path] — путь в базе данных.
  /// [query] — строка запроса (если есть).
  /// [error] — текст ошибки при закрытии (если есть).
  void markSessionClosed({required String path, String? query, String? error}) {
    if (!_config.enabled) return;
    final (sessionPath, sessionQuery) =
        _sessionTargetParts(path: path, query: query);
    final sessionId = _sessionIdFor(path: sessionPath, query: sessionQuery);
    final target = _targetFor(path: sessionPath, query: sessionQuery);
    final buffer = _buffers.putIfAbsent(
      sessionId,
      () => _SessionBuffer(
        session: FirebaseIngestSession(
          id: sessionId,
          target: target,
          captureId: _config.captureId,
        ),
      ),
    );
    buffer.close = true;
    if ((error ?? '').trim().isNotEmpty) {
      buffer.error = error!.trim();
    }
    _scheduleFlush();
  }

  Future<void> _flushSession(_SessionBuffer session) async {
    if (session.frames.isEmpty && !session.close) return;
    final payload = FirebaseIngestRequest(
      session: session.session,
      frames: session.frames,
      close: session.close,
      error: session.error,
    );
    await _client.ingest(payload);
  }

  void _scheduleFlush() {
    _flushTimer ??= Timer(_config.flushInterval, () async {
      _flushTimer = null;
      await flush();
    });
  }

  String _sessionIdFor({required String path, String? query}) {
    final target = _targetFor(path: path, query: query);
    final base = 'fb-${_stableHash64(target)}';
    if (!_config.includeRunIdInSessionId) return base;
    return '$base-$_runId';
  }

  String _targetFor({required String path, String? query}) {
    final normalizedPath = _normalizePath(path);
    final base = _config.databaseUrl.trim();
    final withSlash = base.endsWith('/') ? base : '$base/';
    final fullPath = normalizedPath.startsWith('/')
        ? normalizedPath.substring(1)
        : normalizedPath;
    if ((query ?? '').trim().isEmpty) {
      return '$withSlash$fullPath';
    }
    return '$withSlash$fullPath?${query!.trim()}';
  }

  String _nextFrameId(String op) {
    _seq++;
    final micros = DateTime.now().microsecondsSinceEpoch;
    return 'fr-$op-$micros-$_seq';
  }

  String _normalizePath(String path) {
    final p = path.trim();
    if (p.isEmpty) return '/';
    return p.startsWith('/') ? p : '/$p';
  }

  (String, String?) _sessionTargetParts({
    required String path,
    required String? query,
  }) {
    final normalized = _normalizePath(path);
    final depth = _config.sessionPathDepth;
    if (depth < 0) {
      return (normalized, query);
    }
    if (depth == 0) {
      return ('/', null);
    }
    final parts = normalized
        .split('/')
        .where((s) => s.trim().isNotEmpty)
        .toList(growable: false);
    if (parts.isEmpty) return ('/', null);
    final take = depth.clamp(1, parts.length);
    final prefix = '/${parts.take(take).join('/')}';
    return (prefix, null);
  }

  static String _makeRunId() {
    // 8 hex chars — достаточно чтобы не конфликтовать между запусками
    Random rnd;
    try {
      rnd = Random.secure();
    } catch (_) {
      rnd = Random();
    }
    final v = rnd.nextInt(0xffffffff);
    return v.toRadixString(16).padLeft(8, '0');
  }

  String _stableHash64(String input) {
    const int fnvOffset = 0xcbf29ce484222325;
    const int fnvPrime = 0x100000001b3;
    var hash = fnvOffset;
    final bytes = utf8.encode(input);
    for (final b in bytes) {
      hash ^= b;
      hash = (hash * fnvPrime) & 0xffffffffffffffff;
    }
    final hex = hash.toRadixString(16).padLeft(16, '0');
    return hex;
  }

  (String, String?, String?) _buildFramePayload({
    required String op,
    required String path,
    required String? query,
    required dynamic payload,
    required bool ok,
    required Object? error,
    required DateTime ts,
  }) {
    final envelope = <String, dynamic>{
      'type': 'firebase_database',
      'op': op,
      'path': path,
      'query': query,
      'ok': ok,
      'error': error?.toString(),
      'ts': ts.toIso8601String(),
    };
    if (payload != null) {
      envelope['payload'] = payload;
    }
    final preview = jsonEncode(envelope);
    if (preview.length <= _config.previewBodyThresholdBytes) {
      return (preview, null, null);
    }
    final bodyBytes = utf8.encode(preview);
    final bodyBase64 = base64Encode(bodyBytes);
    final compactPreview = jsonEncode(<String, dynamic>{
      ...envelope,
      'bodySpilled': true,
      'payload': null,
    });
    return (compactPreview, bodyBase64, 'base64');
  }
}

/// Отладочная обёртка над [DatabaseReference].
///
/// Проксирует все операции (get, set, update, remove, onValue)
/// к оригинальному [DatabaseReference] и логирует их через [FirebaseDatabaseDebugger].
class DebugDatabaseReference {
  DebugDatabaseReference._({
    required FirebaseDatabaseDebugger owner,
    required DatabaseReference inner,
  })  : _owner = owner,
        _inner = inner;

  final FirebaseDatabaseDebugger _owner;
  final DatabaseReference _inner;

  /// Оригинальный [DatabaseReference] без обёртки.
  DatabaseReference get raw => _inner;

  /// Путь в базе данных, например `/users/123`.
  String get path => _inner.path;

  /// Возвращает отладочную ссылку на дочерний узел по [childPath].
  DebugDatabaseReference child(String childPath) {
    return DebugDatabaseReference._(
      owner: _owner,
      inner: _inner.child(childPath),
    );
  }

  /// Записывает [value] по текущему пути и логирует операцию.
  Future<void> set(Object? value) async {
    final startedAt = DateTime.now().toUtc();
    try {
      await _inner.set(value);
      await _owner.logOperation(
        path: path,
        op: 'set',
        direction: 'client->upstream',
        payload: <String, dynamic>{
          'value': value,
          'durationMs':
              DateTime.now().toUtc().difference(startedAt).inMilliseconds,
        },
      );
    } catch (e) {
      await _owner.logOperation(
        path: path,
        op: 'set',
        direction: 'client->upstream',
        payload: <String, dynamic>{'value': value},
        ok: false,
        error: e,
      );
      rethrow;
    }
  }

  /// Обновляет поля по текущему пути (merge) и логирует операцию.
  Future<void> update(Map<String, Object?> value) async {
    final startedAt = DateTime.now().toUtc();
    try {
      await _inner.update(value);
      await _owner.logOperation(
        path: path,
        op: 'update',
        direction: 'client->upstream',
        payload: <String, dynamic>{
          'value': value,
          'durationMs':
              DateTime.now().toUtc().difference(startedAt).inMilliseconds,
        },
      );
    } catch (e) {
      await _owner.logOperation(
        path: path,
        op: 'update',
        direction: 'client->upstream',
        payload: <String, dynamic>{'value': value},
        ok: false,
        error: e,
      );
      rethrow;
    }
  }

  /// Удаляет данные по текущему пути и логирует операцию.
  Future<void> remove() async {
    try {
      await _inner.remove();
      await _owner.logOperation(
        path: path,
        op: 'remove',
        direction: 'client->upstream',
        payload: null,
      );
    } catch (e) {
      await _owner.logOperation(
        path: path,
        op: 'remove',
        direction: 'client->upstream',
        payload: null,
        ok: false,
        error: e,
      );
      rethrow;
    }
  }

  /// Читает данные по текущему пути и логирует результат.
  Future<DataSnapshot> get() async {
    final startedAt = DateTime.now().toUtc();
    try {
      final result = await _inner.get();
      await _owner.logOperation(
        path: path,
        op: 'get',
        direction: 'client->upstream',
        payload: <String, dynamic>{
          'exists': result.exists,
          'value': result.value,
          'durationMs':
              DateTime.now().toUtc().difference(startedAt).inMilliseconds,
        },
      );
      return result;
    } catch (e) {
      await _owner.logOperation(
        path: path,
        op: 'get',
        direction: 'client->upstream',
        payload: null,
        ok: false,
        error: e,
      );
      rethrow;
    }
  }

  /// Подписка на изменения значения по текущему пути.
  ///
  /// Каждое обновление логируется как фрейм `onValue` с направлением
  /// `upstream->client`.
  Stream<DatabaseEvent> get onValue {
    _owner.logOperation(
      path: path,
      op: 'listen_start',
      direction: 'client->upstream',
      payload: <String, dynamic>{'event': 'onValue'},
    );
    return _inner.onValue.map((event) {
      _owner.logOperation(
        path: path,
        op: 'onValue',
        direction: 'upstream->client',
        payload: <String, dynamic>{
          'exists': event.snapshot.exists,
          'value': event.snapshot.value,
        },
      );
      return event;
    }).handleError((Object error) {
      _owner.logOperation(
        path: path,
        op: 'onValue_error',
        direction: 'upstream->client',
        payload: null,
        ok: false,
        error: error,
      );
    });
  }
}

/// Отладочная обёртка над Firebase [Query].
///
/// Проксирует операции чтения и подписки, логируя каждую
/// через [FirebaseDatabaseDebugger].
class DebugQuery {
  DebugQuery._({required FirebaseDatabaseDebugger owner, required Query inner})
      : _owner = owner,
        _inner = inner;

  final FirebaseDatabaseDebugger _owner;
  final Query _inner;

  /// Оригинальный [Query] без обёртки.
  Query get raw => _inner;

  /// Путь в базе данных, к которому привязан запрос.
  String get path => _inner.path;

  /// Выполняет запрос и возвращает снапшот данных, логируя результат.
  Future<DataSnapshot> get() async {
    final startedAt = DateTime.now().toUtc();
    try {
      final result = await _inner.get();
      await _owner.logOperation(
        path: path,
        op: 'query_get',
        direction: 'client->upstream',
        payload: <String, dynamic>{
          'exists': result.exists,
          'value': result.value,
          'durationMs':
              DateTime.now().toUtc().difference(startedAt).inMilliseconds,
        },
      );
      return result;
    } catch (e) {
      await _owner.logOperation(
        path: path,
        op: 'query_get',
        direction: 'client->upstream',
        payload: null,
        ok: false,
        error: e,
      );
      rethrow;
    }
  }

  /// Подписка на изменения значения по запросу.
  ///
  /// Каждое обновление логируется как фрейм `onValue`.
  Stream<DatabaseEvent> get onValue {
    _owner.logOperation(
      path: path,
      op: 'listen_start',
      direction: 'client->upstream',
      payload: <String, dynamic>{'event': 'onValue'},
    );
    return _inner.onValue.map((event) {
      _owner.logOperation(
        path: path,
        op: 'onValue',
        direction: 'upstream->client',
        payload: <String, dynamic>{
          'exists': event.snapshot.exists,
          'value': event.snapshot.value,
        },
      );
      return event;
    }).handleError((Object error) {
      _owner.logOperation(
        path: path,
        op: 'onValue_error',
        direction: 'upstream->client',
        payload: null,
        ok: false,
        error: error,
      );
    });
  }

  /// Подписка на добавление дочерних узлов.
  ///
  /// Каждый новый дочерний элемент логируется как фрейм `onChildAdded`.
  Stream<DatabaseEvent> get onChildAdded {
    _owner.logOperation(
      path: path,
      op: 'listen_start',
      direction: 'client->upstream',
      payload: <String, dynamic>{'event': 'onChildAdded'},
    );
    return _inner.onChildAdded.map((event) {
      _owner.logOperation(
        path: path,
        op: 'onChildAdded',
        direction: 'upstream->client',
        payload: <String, dynamic>{
          'key': event.snapshot.key,
          'value': event.snapshot.value,
        },
      );
      return event;
    });
  }

  /// Подписка на изменение дочерних узлов.
  ///
  /// Каждое изменение логируется как фрейм `onChildChanged`.
  Stream<DatabaseEvent> get onChildChanged {
    _owner.logOperation(
      path: path,
      op: 'listen_start',
      direction: 'client->upstream',
      payload: <String, dynamic>{'event': 'onChildChanged'},
    );
    return _inner.onChildChanged.map((event) {
      _owner.logOperation(
        path: path,
        op: 'onChildChanged',
        direction: 'upstream->client',
        payload: <String, dynamic>{
          'key': event.snapshot.key,
          'value': event.snapshot.value,
        },
      );
      return event;
    });
  }

  /// Подписка на удаление дочерних узлов.
  ///
  /// Каждое удаление логируется как фрейм `onChildRemoved`.
  Stream<DatabaseEvent> get onChildRemoved {
    _owner.logOperation(
      path: path,
      op: 'listen_start',
      direction: 'client->upstream',
      payload: <String, dynamic>{'event': 'onChildRemoved'},
    );
    return _inner.onChildRemoved.map((event) {
      _owner.logOperation(
        path: path,
        op: 'onChildRemoved',
        direction: 'upstream->client',
        payload: <String, dynamic>{
          'key': event.snapshot.key,
          'value': event.snapshot.value,
        },
      );
      return event;
    });
  }
}

/// Буфер фреймов для одной сессии, ожидающий отправки.
class _SessionBuffer {
  _SessionBuffer({required this.session});

  final FirebaseIngestSession session;
  final List<FirebaseIngestFrame> frames = <FirebaseIngestFrame>[];
  bool close = false;
  String? error;
}
