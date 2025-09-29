import 'package:mobx/mobx.dart';
import '../../domain/entities/session.dart';
import '../usecases/list_sessions.dart';
import '../../../../core/network/error_utils.dart';
import '../../../../core/di/di.dart';
import '../../../../core/notifications/notifications_service.dart';

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
  Future<void> load({String? q, String? target}) async {
    if (loading) return;
    loading = true;
    try {
      final res = await _listSessions(q: q, target: target);
      items = ObservableList.of(res);
    } catch (e, st) {
      final msg = resolveErrorMessage(e, st);
      sl<NotificationsService>().errorFromResolved(msg);
    } finally {
      loading = false;
    }
  }
}


