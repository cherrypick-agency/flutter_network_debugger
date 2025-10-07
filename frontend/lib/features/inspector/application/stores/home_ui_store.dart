import 'package:mobx/mobx.dart' as mobx;
import 'package:flutter/material.dart';

// UI-стор для главной страницы инспектора
class HomeUiStore {
  HomeUiStore() {
    selectedSessionId = mobx.Observable<String?>(null);
    hoveredSessionId = mobx.Observable<String?>(null);
    showFilters = mobx.Observable<bool>(false);
    hideHeartbeats = mobx.Observable<bool>(false);
    wfFitAll = mobx.Observable<bool>(true);
    showSearch = mobx.Observable<bool>(false);
    since = mobx.Observable<DateTime?>(null);
    selectedRange = mobx.Observable<DateTimeRange?>(null);
    selectedDomains = mobx.ObservableSet.of(<String>{});
    httpMeta = mobx.ObservableMap.of(<String, Map<String, dynamic>>{});
    opcodeFilter = mobx.Observable<String>('all');
    directionFilter = mobx.Observable<String>('all');
    sessionTabById = mobx.ObservableMap.of(<String, String>{});
    sessionSearchQuery = mobx.Observable<String>('');
    isRecording = mobx.Observable<bool>(true);
    captureScope = mobx.Observable<String>('current'); // current | all
    includePaused = mobx.Observable<bool>(false);
  }

  late final mobx.Observable<String?> selectedSessionId;
  late final mobx.Observable<String?> hoveredSessionId;
  late final mobx.Observable<bool> showFilters;
  late final mobx.Observable<bool> hideHeartbeats;
  late final mobx.Observable<bool> wfFitAll;
  late final mobx.Observable<bool> showSearch;
  late final mobx.Observable<DateTime?> since;
  late final mobx.Observable<DateTimeRange?> selectedRange;
  late final mobx.ObservableSet<String> selectedDomains;
  late final mobx.ObservableMap<String, Map<String, dynamic>> httpMeta;
  late final mobx.Observable<String> opcodeFilter;
  late final mobx.Observable<String> directionFilter;
  late final mobx.ObservableMap<String, String> sessionTabById;
  late final mobx.Observable<String> sessionSearchQuery;
  late final mobx.Observable<bool> isRecording;
  late final mobx.Observable<String> captureScope;
  late final mobx.Observable<bool> includePaused;

  void setSelectedSessionId(String? v) =>
      mobx.runInAction(() => selectedSessionId.value = v);
  void setHoveredSessionId(String? v) =>
      mobx.runInAction(() => hoveredSessionId.value = v);
  void setShowFilters(bool v) => mobx.runInAction(() => showFilters.value = v);
  void toggleShowFilters() =>
      mobx.runInAction(() => showFilters.value = !showFilters.value);
  void setHideHeartbeats(bool v) =>
      mobx.runInAction(() => hideHeartbeats.value = v);
  void setWfFitAll(bool v) => mobx.runInAction(() => wfFitAll.value = v);
  void setShowSearch(bool v) => mobx.runInAction(() => showSearch.value = v);
  void setSince(DateTime? v) => mobx.runInAction(() => since.value = v);
  void setSelectedRange(DateTimeRange? v) =>
      mobx.runInAction(() => selectedRange.value = v);
  void clearSelectedRange() =>
      mobx.runInAction(() => selectedRange.value = null);
  void addDomain(String host) =>
      mobx.runInAction(() => selectedDomains.add(host));
  void removeDomain(String host) =>
      mobx.runInAction(() => selectedDomains.remove(host));
  void setOpcodeFilter(String v) =>
      mobx.runInAction(() => opcodeFilter.value = v);
  void setDirectionFilter(String v) =>
      mobx.runInAction(() => directionFilter.value = v);
  String? getSessionTab(String sessionId) => sessionTabById[sessionId];
  void setSessionTab(String sessionId, String tab) =>
      mobx.runInAction(() => sessionTabById[sessionId] = tab);
  void setSessionSearchQuery(String v) =>
      mobx.runInAction(() => sessionSearchQuery.value = v);
  void setIsRecording(bool v) => mobx.runInAction(() => isRecording.value = v);
  void toggleRecording() =>
      mobx.runInAction(() => isRecording.value = !isRecording.value);

  void setCaptureScope(String v) =>
      mobx.runInAction(() => captureScope.value = v);
  void setIncludePaused(bool v) =>
      mobx.runInAction(() => includePaused.value = v);
}
