import 'package:mobx/mobx.dart';

import '../../../../core/di/di.dart';
import '../../../../core/notifications/notifications_service.dart';
import '../../domain/entities/intercept_item.dart';
import '../../domain/entities/decisions.dart';
import '../../domain/repositories/breakpoints_repository.dart';

part 'intercept_editor_store.g.dart';

class InterceptEditorStore = _InterceptEditorStore with _$InterceptEditorStore;

abstract class _InterceptEditorStore with Store {
  final BreakpointsRepository _repo;

  _InterceptEditorStore(this._repo);

  @observable
  InterceptItem? item;

  @observable
  bool submitting = false;

  @observable
  String? lastError;

  @action
  void setItem(InterceptItem? it) {
    item = it;
  }

  @action
  Future<void> continueRequest({
    String? method,
    String? url,
    Map<String, List<String>>? headers,
    String? bodyBase64,
    bool drop = false,
  }) async {
    if (item == null) return;
    submitting = true;
    lastError = null;
    try {
      await _repo.continueRequest(
        item!.id,
        RequestDecision(
          action: drop ? 'drop' : 'continue',
          method: method,
          url: url,
          headers: headers,
          bodyBase64: bodyBase64,
        ),
      );
    } catch (e) {
      lastError = e.toString();
      sl<NotificationsService>().error('Continue request', e.toString());
      rethrow;
    } finally {
      submitting = false;
    }
  }

  @action
  Future<void> continueResponse({
    int? status,
    Map<String, List<String>>? headers,
    String? bodyBase64,
  }) async {
    if (item == null) return;
    submitting = true;
    lastError = null;
    try {
      await _repo.continueResponse(
        item!.id,
        ResponseDecision(
          action: 'continue',
          status: status,
          headers: headers,
          bodyBase64: bodyBase64,
        ),
      );
    } catch (e) {
      lastError = e.toString();
      sl<NotificationsService>().error('Continue response', e.toString());
      rethrow;
    } finally {
      submitting = false;
    }
  }

  @action
  Future<void> cancel() async {
    if (item == null) return;
    submitting = true;
    lastError = null;
    try {
      await _repo.cancel(item!.id);
    } catch (e) {
      lastError = e.toString();
      sl<NotificationsService>().error('Cancel', e.toString());
      rethrow;
    } finally {
      submitting = false;
    }
  }
}
