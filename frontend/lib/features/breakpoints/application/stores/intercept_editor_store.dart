import 'package:mobx/mobx.dart';

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

  @action
  void setItem(InterceptItem? it) {
    item = it;
  }

  Future<void> continueRequest({
    String? method,
    String? url,
    Map<String, List<String>>? headers,
    String? bodyBase64,
    bool drop = false,
  }) async {
    if (item == null) return;
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
  }

  Future<void> continueResponse({
    int? status,
    Map<String, List<String>>? headers,
    String? bodyBase64,
  }) async {
    if (item == null) return;
    await _repo.continueResponse(
      item!.id,
      ResponseDecision(
        action: 'continue',
        status: status,
        headers: headers,
        bodyBase64: bodyBase64,
      ),
    );
  }

  Future<void> cancel() async {
    if (item == null) return;
    await _repo.cancel(item!.id);
  }
}
