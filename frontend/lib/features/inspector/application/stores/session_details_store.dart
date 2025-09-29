import 'package:mobx/mobx.dart';
import 'dart:async';
import 'dart:convert';
import 'package:http/http.dart' as http;
// removed duplicate di import below
import 'package:app_http_client/application/app_http_client.dart' as http_client;
import '../../domain/entities/frame.dart';
import '../../domain/entities/event.dart';
import '../usecases/list_frames.dart';
import '../usecases/list_events.dart';
import '../../../../core/network/error_utils.dart';
import '../../../../core/notifications/notifications_service.dart';
import '../../../../core/di/di.dart';

part 'session_details_store.g.dart';

class SessionDetailsStore = _SessionDetailsStore with _$SessionDetailsStore;

abstract class _SessionDetailsStore with Store {
  _SessionDetailsStore(this._listFrames, this._listEvents);
  final ListFramesUseCase _listFrames;
  final ListEventsUseCase _listEvents;

  http.Client? _sseClient;
  StreamSubscription<List<int>>? _sseSub;
  bool _sseConnecting = false;

  @observable
  String? sessionId;

  @observable
  ObservableList<Frame> frames = ObservableList.of([]);

  @observable
  ObservableList<EventEntity> events = ObservableList.of([]);

  @observable
  bool loading = false;

  @action
  Future<void> open(String id) async {
    sessionId = id;
    frames.clear();
    events.clear();
    await Future.wait([
      loadMoreFrames(),
      loadMoreEvents(),
    ]);
    _startSSE();
  }

  @action
  Future<void> loadMoreFrames() async {
    if (sessionId == null || loading) return;
    loading = true;
    try {
      final from = frames.isNotEmpty ? frames.last.id : null;
      final res = await _listFrames(sessionId!, from: from, limit: 100);
      frames.addAll(res);
    } catch (e, st) {
      final msg = resolveErrorMessage(e, st);
      sl<NotificationsService>().errorFromResolved(msg);
    } finally {
      loading = false;
    }
  }

  @action
  Future<void> loadMoreEvents() async {
    if (sessionId == null || loading) return;
    loading = true;
    try {
      final from = events.isNotEmpty ? events.last.id : null;
      final res = await _listEvents(sessionId!, from: from, limit: 100);
      events.addAll(res);
    } catch (e, st) {
      final msg = resolveErrorMessage(e, st);
      sl<NotificationsService>().errorFromResolved(msg);
    } finally {
      loading = false;
    }
  }

  void _startSSE() {
    if (_sseConnecting || sessionId == null) return;
    _sseConnecting = true;
    // cleanup previous
    _sseSub?.cancel();
    _sseClient?.close();
    _sseSub = null;
    _sseClient = http.Client();
    final base = sl<http_client.AppHttpClient>().defaultHost;
    final b = base.endsWith('/') ? base.substring(0, base.length - 1) : base;
    final url = Uri.parse('$b/_api/v1/sessions_stream/${sessionId}');
    final req = http.Request('GET', url);
    req.headers['Accept'] = 'text/event-stream';
    _sseClient!.send(req).then((res){
      _sseConnecting = false;
      var event = '';
      final decoder = const Utf8Decoder();
      String buffer = '';
      _sseSub = res.stream.listen((chunk){
        buffer += decoder.convert(chunk);
        // process by lines
        int idx;
        while ((idx = buffer.indexOf('\n')) != -1) {
          final line = buffer.substring(0, idx).trimRight();
          buffer = buffer.substring(idx + 1);
          if (line.isEmpty) {
            continue; // ignore stray blanks
          }
          if (line.startsWith('event: ')) {
            event = line.substring(7).trim();
            // next line(s) expected to be raw JSON (no 'data:')
            continue;
          }
          // Treat line as JSON payload; enc.Encode writes one line per event
          try {
            final data = jsonDecode(line);
            _handleSSE(event, data);
          } catch (_) {}
        }
      }, onError: (_){}, onDone: (){ _sseConnecting = false; });
    }).catchError((_){ _sseConnecting = false; });
  }

  void _handleSSE(String event, dynamic data) {
    if (sessionId == null) return;
    if (event == 'frames') {
      final List<dynamic> arr = (data is List) ? data : [data];
      for (final f in arr) {
        try {
          frames.add(Frame(
            id: (f['id'] ?? '').toString(),
            ts: DateTime.tryParse((f['ts'] ?? '').toString()) ?? DateTime.now(),
            direction: (f['direction'] ?? '').toString(),
            opcode: (f['opcode'] ?? '').toString(),
            size: int.tryParse((f['size'] ?? '0').toString()) ?? 0,
            preview: (f['preview'] ?? '').toString(),
          ));
        } catch (_) {}
      }
    } else if (event == 'events') {
      final List<dynamic> arr = (data is List) ? data : [data];
      for (final e in arr) {
        try {
          events.add(EventEntity(
            id: (e['id'] ?? '').toString(),
            ts: DateTime.tryParse((e['ts'] ?? '').toString()) ?? DateTime.now(),
            namespace: (e['namespace'] ?? '').toString(),
            event: (e['event'] ?? e['name'] ?? '').toString(),
            ackId: int.tryParse((e['ackId'] ?? '').toString()),
            argsPreview: (e['argsPreview'] ?? '').toString(),
          ));
        } catch (_) {}
      }
    }
  }
}


