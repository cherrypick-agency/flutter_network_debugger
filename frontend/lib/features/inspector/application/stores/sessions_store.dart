import 'package:mobx/mobx.dart';
import '../../domain/entities/session.dart';
import '../usecases/list_sessions.dart';
import '../../../../core/network/error_utils.dart';
import '../../../../core/di/di.dart';
import '../../../../core/notifications/notifications_service.dart';
import 'home_ui_store.dart';

part 'sessions_store.g.dart';

class SessionsStore = _SessionsStore with _$SessionsStore;

abstract class _SessionsStore with Store {
  _SessionsStore(this._listSessions);
  final ListSessionsUseCase _listSessions;

  @observable
  ObservableList<Session> items = ObservableList.of([]);

  @observable
  bool loading = false;

  @action
  void clear() {
    items = ObservableList.of([]);
    // Сессии очистили — мета по ним больше не нужна, чтобы не ловить странные артефакты в деталях.
    try {
      sl<HomeUiStore>().httpMeta.clear();
    } catch (_) {}
  }

  @action
  void applyInit(List<Session> data) {
    items = ObservableList.of(data);
    _syncHttpMetaFromSessions(data);
  }

  @action
  void upsert(Session s) {
    final idx = items.indexWhere((e) => e.id == s.id);
    if (idx >= 0) {
      items[idx] = s;
    } else {
      items.add(s);
    }
    _syncHttpMetaFromSession(s);
  }

  @action
  void removeById(String id) {
    items.removeWhere((e) => e.id == id);
    try {
      sl<HomeUiStore>().httpMeta.remove(id);
    } catch (_) {}
  }

  @action
  Future<void> load({
    String? q,
    String? target,
    List<String>? tags,
    Set<String>? types,
    Set<String>? statusGroups,
  }) async {
    if (loading) return;
    loading = true;
    try {
      final res = await _listSessions(q: q, target: target, tags: tags);
      items = ObservableList.of(res);
    } catch (e, st) {
      final msg = resolveErrorMessage(e, st);
      sl<NotificationsService>().errorFromResolved(msg);
    } finally {
      loading = false;
    }
  }

  void _syncHttpMetaFromSessions(List<Session> sessions) {
    for (final s in sessions) {
      _syncHttpMetaFromSession(s);
    }
  }

  void _syncHttpMetaFromSession(Session s) {
    final meta = s.httpMeta;
    if (meta == null || meta.isEmpty) return;
    try {
      // HomeUiStore используется в деталях и некоторых фильтрах как "кэш" меты.
      // Если сервер уже прислал httpMeta — кладём сразу, чтобы UI был стабильнее.
      sl<HomeUiStore>().httpMeta[s.id] = Map<String, dynamic>.from(meta);
    } catch (_) {}
  }
}
