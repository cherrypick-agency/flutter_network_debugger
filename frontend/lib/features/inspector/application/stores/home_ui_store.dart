import 'package:mobx/mobx.dart' as mobx;
import 'package:flutter/material.dart';

// UI store for the inspector main page
class HomeUiStore {
  HomeUiStore() {
    selectedSessionId = mobx.Observable<String?>(null);
    hoveredSessionId = mobx.Observable<String?>(null);
    showFilters = mobx.Observable<bool>(false);
    hideHeartbeats = mobx.Observable<bool>(false);
    wfFitAll = mobx.Observable<bool>(true);
    showTimeline = mobx.Observable<bool>(true);
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
    includePaused = mobx.Observable<bool>(true);
    recentWindowEnabled = mobx.Observable<bool>(false);
    recentWindowMinutes = mobx.Observable<int>(5);
    // WS frames view state
    wsPretty = mobx.Observable<bool>(true);
    wsTree = mobx.Observable<bool>(false);
    wsShowTimeline = mobx.Observable<bool>(true);
    // Quick filters below the timeline
    quickTypes = mobx.ObservableSet.of(<String>{});
    quickStatusGroups = mobx.ObservableSet.of(<String>{});
    // Firebase RTDB filters
    fbOpFilter = mobx.Observable<String>('all');
    fbStatusFilter = mobx.Observable<String>('all');
    fbPathFilter = mobx.Observable<String>('');
  }

  late final mobx.Observable<String?> selectedSessionId;
  late final mobx.Observable<String?> hoveredSessionId;
  late final mobx.Observable<bool> showFilters;
  late final mobx.Observable<bool> hideHeartbeats;
  late final mobx.Observable<bool> wfFitAll;
  late final mobx.Observable<bool> showTimeline;
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
  late final mobx.Observable<bool> recentWindowEnabled;
  late final mobx.Observable<int> recentWindowMinutes;
  // timestamp when recording was paused (to filter out new sessions)
  late final mobx.Observable<DateTime?> pausedSince =
      mobx.Observable<DateTime?>(null);
  // WS frames view state
  late final mobx.Observable<bool> wsPretty;
  late final mobx.Observable<bool> wsTree;
  late final mobx.Observable<bool> wsShowTimeline;
  // Quick filters: content/protocol types and status groups
  late final mobx.ObservableSet<String>
  quickTypes; // http, https, ws, json, form, ...
  late final mobx.ObservableSet<String> quickStatusGroups; // 1xx..5xx
  // Firebase RTDB filters
  late final mobx.Observable<String> fbOpFilter; // all | set | get | ...
  late final mobx.Observable<String> fbStatusFilter; // all | ok | error
  late final mobx.Observable<String> fbPathFilter; // substring match

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
  void setShowTimeline(bool v) =>
      mobx.runInAction(() => showTimeline.value = v);
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
  void setIsRecording(bool v) {
    // debug log
    // ignore: avoid_print
    print('[HomeUiStore] setIsRecording -> $v (was ${isRecording.value})');
    mobx.runInAction(() => isRecording.value = v);
  }

  void toggleRecording() =>
      mobx.runInAction(() => isRecording.value = !isRecording.value);

  void setCaptureScope(String v) =>
      mobx.runInAction(() => captureScope.value = v);
  void setIncludePaused(bool v) =>
      mobx.runInAction(() => includePaused.value = v);
  void setPausedSince(DateTime? v) =>
      mobx.runInAction(() => pausedSince.value = v);
  void setRecentWindowEnabled(bool v) =>
      mobx.runInAction(() => recentWindowEnabled.value = v);
  void setRecentWindowMinutes(int v) =>
      mobx.runInAction(() => recentWindowMinutes.value = v);

  // WS frames view
  void setWsPretty(bool v) => mobx.runInAction(() => wsPretty.value = v);
  void setWsTree(bool v) => mobx.runInAction(() => wsTree.value = v);
  void setWsShowTimeline(bool v) =>
      mobx.runInAction(() => wsShowTimeline.value = v);

  // Quick filters
  void toggleQuickType(String key) => mobx.runInAction(() {
    if (quickTypes.contains(key)) {
      quickTypes.remove(key);
    } else {
      quickTypes.add(key);
    }
  });

  void toggleQuickStatusGroup(String key) => mobx.runInAction(() {
    if (quickStatusGroups.contains(key)) {
      quickStatusGroups.remove(key);
    } else {
      quickStatusGroups.add(key);
    }
  });

  void clearQuickFilters() => mobx.runInAction(() {
    quickTypes.clear();
    quickStatusGroups.clear();
  });

  // Firebase RTDB filters
  void setFbOpFilter(String v) => mobx.runInAction(() => fbOpFilter.value = v);
  void setFbStatusFilter(String v) =>
      mobx.runInAction(() => fbStatusFilter.value = v);
  void setFbPathFilter(String v) =>
      mobx.runInAction(() => fbPathFilter.value = v);
}
