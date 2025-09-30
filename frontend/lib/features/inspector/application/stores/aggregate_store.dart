import 'package:mobx/mobx.dart';
import '../usecases/list_aggregate.dart';
import '../../../../core/network/error_utils.dart';
import '../../../../core/notifications/notifications_service.dart';
import '../../../../core/di/di.dart';

part 'aggregate_store.g.dart';

class AggregateStore = _AggregateStore with _$AggregateStore;

abstract class _AggregateStore with Store {
  _AggregateStore(this._listAgg);
  final ListAggregateUseCase _listAgg;

  @observable
  ObservableList<Map<String, dynamic>> groups = ObservableList.of([]);

  @observable
  bool loading = false;

  @action
  void clear() {
    groups = ObservableList.of([]);
  }

  @action
  Future<void> load({String groupBy = 'domain'}) async {
    if (loading) return;
    loading = true;
    try {
      final res = await _listAgg(groupBy: groupBy);
      groups = ObservableList.of(res);
    } catch (e) {
      final msg = resolveErrorMessage(e);
      sl<NotificationsService>().error(msg.title, msg.description);
    } finally {
      loading = false;
    }
  }
}


