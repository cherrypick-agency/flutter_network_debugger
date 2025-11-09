import 'package:mobx/mobx.dart';
import '../../domain/entities/frame.dart';
import '../../domain/entities/event.dart';
import '../usecases/list_frames.dart';
import '../usecases/list_events.dart';
import '../../../../core/network/error_utils.dart';
import '../../../../core/notifications/notifications_service.dart';
import '../../../../core/di/di.dart';
import '../../../../core/realtime/socket_io_service.dart';

part 'session_details_store.g.dart';

class SessionDetailsStore = _SessionDetailsStore with _$SessionDetailsStore;

abstract class _SessionDetailsStore with Store {
  _SessionDetailsStore(this._listFrames, this._listEvents);
  final ListFramesUseCase _listFrames;
  final ListEventsUseCase _listEvents;

  SocketIoService? _socketIo;

  @observable
  String? sessionId;

  @observable
  ObservableList<Frame> frames = ObservableList.of([]);

  @observable
  ObservableList<EventEntity> events = ObservableList.of([]);

  @observable
  bool loading = false;

  @observable
  String? loadError;

  @action
  Future<void> open(String id) async {
    sessionId = id;
    frames.clear();
    events.clear();
    loadError = null;
    await Future.wait([loadMoreFrames(), loadMoreEvents()]);
    _subscribeToSession();
  }

  @action
  void dispose() {
    _unsubscribeFromSession();
  }

  @action
  Future<void> loadMoreFrames() async {
    if (sessionId == null || loading) return;
    loading = true;
    loadError = null;
    try {
      final from = frames.isNotEmpty ? frames.last.id : null;
      final res = await _listFrames(sessionId!, from: from, limit: 100);
      frames.addAll(res);
    } catch (e, st) {
      final msg = resolveErrorMessage(e, st);
      loadError = msg.description;
      sl<NotificationsService>().errorFromResolved(msg);
    } finally {
      loading = false;
    }
  }

  @action
  Future<void> loadMoreEvents() async {
    if (sessionId == null || loading) return;
    loading = true;
    loadError = null;
    try {
      final from = events.isNotEmpty ? events.last.id : null;
      final res = await _listEvents(sessionId!, from: from, limit: 100);
      events.addAll(res);
    } catch (e, st) {
      final msg = resolveErrorMessage(e, st);
      loadError = msg.description;
      sl<NotificationsService>().errorFromResolved(msg);
    } finally {
      loading = false;
    }
  }

  void _subscribeToSession() {
    if (sessionId == null) return;

    // Получаем синглтон Socket.IO сервиса
    _socketIo = sl<SocketIoService>();

    // Подписываемся на события
    _socketIo?.on('session:frames', _handleFrames);
    _socketIo?.on('session:events', _handleEvents);
    _socketIo?.on('session:http', _handleHttp);

    // Отправляем подписку на сервер
    _socketIo?.emit('session:subscribe', {'sessionId': sessionId});
  }

  void _unsubscribeFromSession() {
    if (sessionId == null || _socketIo == null) return;

    // Отписываемся от событий
    _socketIo?.off('session:frames', _handleFrames);
    _socketIo?.off('session:events', _handleEvents);
    _socketIo?.off('session:http', _handleHttp);

    // Отправляем отписку на сервер
    _socketIo?.emit('session:unsubscribe', {'sessionId': sessionId});
    _socketIo = null;
  }

  void _handleFrames(dynamic data) {
    if (sessionId == null) return;
    final List<dynamic> arr = (data is List) ? data : [data];
    for (final f in arr) {
      try {
        frames.add(
          Frame(
            id: (f['id'] ?? '').toString(),
            ts: DateTime.tryParse((f['ts'] ?? '').toString()) ?? DateTime.now(),
            direction: (f['direction'] ?? '').toString(),
            opcode: (f['opcode'] ?? '').toString(),
            size: int.tryParse((f['size'] ?? '0').toString()) ?? 0,
            preview: (f['preview'] ?? '').toString(),
          ),
        );
      } catch (_) {}
    }
  }

  void _handleEvents(dynamic data) {
    if (sessionId == null) return;
    final List<dynamic> arr = (data is List) ? data : [data];
    for (final e in arr) {
      try {
        events.add(
          EventEntity(
            id: (e['id'] ?? '').toString(),
            ts: DateTime.tryParse((e['ts'] ?? '').toString()) ?? DateTime.now(),
            namespace: (e['namespace'] ?? '').toString(),
            event: (e['event'] ?? e['name'] ?? '').toString(),
            ackId: int.tryParse((e['ackId'] ?? '').toString()),
            argsPreview: (e['argsPreview'] ?? '').toString(),
          ),
        );
      } catch (_) {}
    }
  }

  void _handleHttp(dynamic data) {
    // HTTP транзакции можно добавить позже если нужно
    // Пока оставляем заглушку
  }
}
