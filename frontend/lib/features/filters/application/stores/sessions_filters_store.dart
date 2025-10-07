import 'package:mobx/mobx.dart';

part 'sessions_filters_store.g.dart';

// Храним значения фильтров сессий в одном Store, чтобы легко шарить через Provider
class SessionsFiltersStore = _SessionsFiltersStore with _$SessionsFiltersStore;

abstract class _SessionsFiltersStore with Store {
  @observable
  String target = '';

  @observable
  String httpMethod = 'any';

  @observable
  String httpStatus = 'any';

  @observable
  String httpMime = '';

  @observable
  int httpMinDurationMs = 0;

  @observable
  String groupBy = 'none';

  @observable
  String headerKey = '';

  @observable
  String headerVal = '';

  // Простой индикатор активности фильтров именно этой фичи (без доменов/диапазонов)
  @computed
  bool get hasActive =>
      target.trim().isNotEmpty ||
      httpMethod != 'any' ||
      httpStatus != 'any' ||
      httpMime.trim().isNotEmpty ||
      httpMinDurationMs > 0 ||
      groupBy != 'none' ||
      headerKey.trim().isNotEmpty ||
      headerVal.trim().isNotEmpty;

  @action
  void setTarget(String v) => target = v;

  @action
  void setHttpMethod(String v) => httpMethod = v;

  @action
  void setHttpStatus(String v) => httpStatus = v;

  @action
  void setHttpMime(String v) => httpMime = v;

  @action
  void setHttpMinDurationMs(int v) => httpMinDurationMs = v;

  @action
  void setGroupBy(String v) => groupBy = v;

  @action
  void setHeaderKey(String v) => headerKey = v;

  @action
  void setHeaderVal(String v) => headerVal = v;
}
