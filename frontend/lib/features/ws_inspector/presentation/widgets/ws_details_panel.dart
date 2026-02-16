import 'dart:async';
import 'package:flutter/material.dart';
import '../../../../theme/context_ext.dart';
import '../../../../core/di/di.dart';
import '../../../inspector/application/stores/home_ui_store.dart';
import 'frames_timeline/frames_timeline.dart';
import 'frames_timeline/frames_timeline_legend.dart';
import '../../../../widgets/json_viewer.dart';
import '../../../../widgets/common_search_bar.dart';
import 'searchable_text_rich.dart';
import 'firebase_event_data.dart';
import 'firebase_event_title.dart';
import 'inline_json_spans.dart';
import 'dart:convert';
import 'package:flutter/services.dart';

String _fmtTime(String ts) {
  try {
    final dt = DateTime.parse(ts).toLocal();
    String two(int v) => v < 10 ? '0$v' : '$v';
    return '${two(dt.hour)}:${two(dt.minute)}:${two(dt.second)}';
  } catch (_) {
    return ts;
  }
}

// Atomic pending focus state to prevent race conditions (Bug #12 fix)
class _PendingFocus {
  final String frameId;
  final int localIndex;
  final int navSeq;
  _PendingFocus(this.frameId, this.localIndex, this.navSeq);
}

// Local search state for a single frame
class _LocalSearchState {
  _LocalSearchState()
    : controller = TextEditingController(),
      focus = FocusNode();
  bool show = false;
  final TextEditingController controller;
  final FocusNode focus;
  bool matchCase = false;
  bool wholeWord = false;
  bool useRegex = false;
  int focusedIndex = 0;
  List<GlobalKey> keys = const [];
  void dispose() {
    controller.dispose();
    focus.dispose();
  }
}

class WsDetailsPanel extends StatefulWidget {
  const WsDetailsPanel({
    super.key,
    required this.frames,
    required this.events,
    required this.opcodeFilter,
    required this.directionFilter,
    required this.namespaceCtrl,
    required this.onChangeOpcode,
    required this.onChangeDirection,
    required this.hideHeartbeats,
    required this.onToggleHeartbeats,
    required this.isClosed,
    this.closedAt,
    this.error,
    this.onCloseFullscreen,
    this.fbOpFilter = 'all',
    this.fbStatusFilter = 'all',
    this.fbPathFilter = '',
    this.onChangeFbOp,
    this.onChangeFbStatus,
    this.onChangeFbPath,
  });
  final List<dynamic> frames;
  final List<dynamic> events;
  final String opcodeFilter;
  final String directionFilter;
  final TextEditingController namespaceCtrl;
  final void Function(String) onChangeOpcode;
  final void Function(String) onChangeDirection;
  final bool hideHeartbeats;
  final void Function(bool) onToggleHeartbeats;
  final bool isClosed;
  final DateTime? closedAt;
  final String? error;
  final VoidCallback? onCloseFullscreen;
  // Firebase RTDB filters
  final String fbOpFilter;
  final String fbStatusFilter;
  final String fbPathFilter;
  final void Function(String)? onChangeFbOp;
  final void Function(String)? onChangeFbStatus;
  final void Function(String)? onChangeFbPath;

  @override
  State<WsDetailsPanel> createState() => _WsDetailsPanelState();
}

class _WsDetailsPanelState extends State<WsDetailsPanel> {
  final _ui = sl<HomeUiStore>();

  // Кеш типа сессии, чтобы не парсить JSON на каждый вызов _frameMatches
  bool? _isFirebaseCache;
  List<dynamic>? _isFirebaseCacheFrames;

  bool get _isFirebaseSession {
    if (_isFirebaseCacheFrames == widget.frames && _isFirebaseCache != null) {
      return _isFirebaseCache!;
    }
    _isFirebaseCacheFrames = widget.frames;
    _isFirebaseCache = _detectFirebase();
    return _isFirebaseCache!;
  }

  bool _detectFirebase() {
    for (final item in widget.frames) {
      final frame = item as Map<String, dynamic>;
      final preview = (frame['preview'] ?? '').toString().trim();
      if (!(preview.startsWith('{') || preview.startsWith('['))) continue;
      try {
        final decoded = jsonDecode(preview);
        if (decoded is Map<String, dynamic> &&
            decoded['type'] == 'firebase_database') {
          return true;
        }
      } catch (_) {
        // skip
      }
    }
    return false;
  }

  bool get _pretty => _ui.wsPretty.value;
  set _pretty(bool v) => _ui.setWsPretty(v);
  bool get _tree => _ui.wsTree.value;
  set _tree(bool v) => _ui.setWsTree(v);
  bool get _showTimeline => _ui.wsShowTimeline.value;
  set _showTimeline(bool v) => _ui.setWsShowTimeline(v);
  final ScrollController _listCtrl = ScrollController();
  // TEMPORARILY DISABLED: Auto-scroll logic
  // bool _stickToBottom = true;
  // // Last known frames list length (for auto-scroll)
  // int _lastFramesLen = 0;
  // // Flag that after next frame we should scroll down
  // bool _autoScrollPending = false;
  DateTimeRange? _brushRange;

  // Guard flag to prevent concurrent reindexing (Bug #3 fix)
  bool _isReindexing = false;

  // Cached filtered frames list for scroll methods
  List<Map<String, dynamic>> _cachedVisibleFrames = [];

  // Update cached visible frames (call whenever filters or frames change)
  void _updateVisibleFramesCache() {
    _cachedVisibleFrames = _computeVisibleFrames();
    // Clear focused index cache when visible frames change (Bug #1 & #2 fix)
    _frameLocalFocusedIndex.clear();
    // Clear match keys cache when visible frames change (Bug #8 fix)
    _frameMatchKeys.clear();
  }

  @override
  void initState() {
    super.initState();
    _updateVisibleFramesCache();
    // TEMPORARILY DISABLED: Auto-scroll logic
    // _listCtrl.addListener(_onListScroll);
  }

  // TEMPORARILY DISABLED: Auto-scroll logic
  // void _onListScroll() {
  //   if (!_listCtrl.hasClients) return;
  //   final pos = _listCtrl.position;
  //   // small threshold to avoid jerking scroll with barely visible "tail"
  //   const threshold = 48.0;
  //   _stickToBottom = (pos.maxScrollExtent - pos.pixels) <= threshold;
  // }

  // Unique frame key to avoid confusing elements with same id in different directions
  String _frameKeyOf(Map<String, dynamic> f) {
    final id = (f['id'] ?? '').toString();
    final dir = (f['direction'] ?? '').toString();
    final ts = (f['ts'] ?? '').toString();
    return '$id|$dir|$ts';
  }

  // Global search across all frames
  bool _showGlobalSearch = false;
  final TextEditingController _searchCtrl = TextEditingController();
  final FocusNode _searchFocus = FocusNode();
  bool _matchCase = false;
  bool _wholeWord = false;
  bool _useRegex = false;
  int _globalFocusedIndex = 0;
  int _globalTotalMatches = 0;
  final Map<String, int> _frameMatchCounts = <String, int>{};
  final Map<String, List<GlobalKey>> _frameMatchKeys =
      <String, List<GlobalKey>>{};
  // Cache: frameKey → local focused index (-1 if not focused)
  final Map<String, int> _frameLocalFocusedIndex = <String, int>{};
  // Bug #12 fix: Atomic pending focus to prevent race conditions
  _PendingFocus? _pendingFocus;
  final Map<String, _PendingFocus> _pendingLocalFocus =
      <String, _PendingFocus>{};
  // Track which frame was expanded by search navigation (to collapse on next nav)
  String? _searchExpandedFid;

  int _searchNavSeq = 0;
  bool _suspendNestedScrollSync = false;

  final Map<String, GlobalKey> _frameTileKeys = <String, GlobalKey>{};
  GlobalKey _tileKeyFor(String id) =>
      _frameTileKeys.putIfAbsent(id, () => GlobalKey());

  final Map<String, ScrollController> _frameInnerScrollControllers =
      <String, ScrollController>{};
  ScrollController _innerScrollControllerFor(String id) =>
      _frameInnerScrollControllers.putIfAbsent(id, () => ScrollController());

  // Controllers for programmatic expand/collapse of tiles
  // BUG 2 fix: Use ExpansibleController for programmatic expansion
  final Map<String, ExpansibleController> _frameTileControllers =
      <String, ExpansibleController>{};
  ExpansibleController _tileControllerFor(String id) =>
      _frameTileControllers.putIfAbsent(id, () => ExpansibleController());

  // Local search at frame level
  final Map<String, _LocalSearchState> _localSearch =
      <String, _LocalSearchState>{};
  _LocalSearchState _localFor(String id) =>
      _localSearch.putIfAbsent(id, () => _LocalSearchState());
  void _cleanupFrameCaches() {
    final current = widget.frames
        .cast<Map<String, dynamic>>()
        .map((f) => _frameKeyOf(f))
        .toSet();
    _frameTileControllers.removeWhere((k, controller) {
      if (!current.contains(k)) {
        controller.dispose();
        return true;
      }
      return false;
    });
    _localSearch.removeWhere((k, state) {
      if (!current.contains(k)) {
        state.dispose();
        return true;
      }
      return false;
    });
    _frameInnerScrollControllers.removeWhere((k, controller) {
      if (!current.contains(k)) {
        controller.dispose();
        return true;
      }
      return false;
    });
    _frameTileKeys.removeWhere((k, _) => !current.contains(k));
    _frameMatchCounts.removeWhere((k, _) => !current.contains(k));
    _frameMatchKeys.removeWhere((k, _) => !current.contains(k));
    _frameLocalFocusedIndex.removeWhere((k, _) => !current.contains(k));
    _pendingLocalFocus.removeWhere((k, _) => !current.contains(k));
    if (_pendingFocus != null && !current.contains(_pendingFocus!.frameId)) {
      _pendingFocus = null;
    }
  }

  void _localGotoNext(String id) {
    final s = _localSearch[id];
    if (s == null || s.keys.isEmpty) return;
    final navSeq = ++_searchNavSeq;
    setState(() {
      s.focusedIndex = (s.focusedIndex + 1) % s.keys.length;
      _pendingLocalFocus[id] = _PendingFocus(id, s.focusedIndex, navSeq);
    });
  }

  void _localGotoPrev(String id) {
    final s = _localSearch[id];
    if (s == null || s.keys.isEmpty) return;
    final navSeq = ++_searchNavSeq;
    setState(() {
      s.focusedIndex = (s.focusedIndex - 1) < 0
          ? s.keys.length - 1
          : s.focusedIndex - 1;
      _pendingLocalFocus[id] = _PendingFocus(id, s.focusedIndex, navSeq);
    });
  }

  bool _frameMatches(Map<String, dynamic> f) {
    if (_isFirebaseSession) {
      // Firebase RTDB: свои фильтры, WS-шные не применяем
      final preview = (f['preview'] ?? '').toString();
      final fb = FirebaseEventData.tryParse(preview);
      if (fb != null) {
        if (widget.fbOpFilter != 'all' && fb.op != widget.fbOpFilter) {
          return false;
        }
        if (widget.fbStatusFilter == 'ok' && !fb.ok) return false;
        if (widget.fbStatusFilter == 'error' && fb.ok) return false;
        if (widget.fbPathFilter.isNotEmpty &&
            !fb.path.contains(widget.fbPathFilter)) {
          return false;
        }
      }
    } else {
      // WS фильтры
      if (widget.opcodeFilter != 'all' &&
          (f['opcode']?.toString() ?? '') != widget.opcodeFilter) {
        return false;
      }
      if (widget.directionFilter != 'all' &&
          (f['direction']?.toString() ?? '') != widget.directionFilter) {
        return false;
      }
    }
    return true;
  }

  // events sidebar temporarily disabled

  bool _isHeartbeat(Map<String, dynamic> f) {
    final opcode = (f['opcode'] ?? '').toString();
    final preview = (f['preview'] ?? '').toString();
    final size = (f['size'] ?? 0).toString();
    final isWsPingPong = opcode == 'ping' || opcode == 'pong';
    final isEnginePingPong =
        opcode == 'text' && (preview == '2' || preview == '3') && size == '1';
    return isWsPingPong || isEnginePingPong;
  }

  // Compute visible frames applying all filters
  List<Map<String, dynamic>> _computeVisibleFrames() {
    return widget.frames.cast<Map<String, dynamic>>().where((f) {
      if (!_frameMatches(f)) return false;
      if (_brushRange != null) {
        try {
          final ts = DateTime.parse((f['ts'] ?? '').toString());
          if (ts.isBefore(_brushRange!.start) || ts.isAfter(_brushRange!.end))
            return false;
        } catch (_) {
          return false;
        }
      }
      if (widget.hideHeartbeats && _isHeartbeat(f)) return false;
      return true;
    }).toList();
  }

  // TEMPORARILY DISABLED: Auto-scroll logic
  // If new frames appeared — scroll down if necessary,
  // but only if user was already at the end of the list
  // void _maybeAutoScrollToBottomOnNewFrames() {
  //   if (widget.frames.length > _lastFramesLen) {
  //     _autoScrollPending = _stickToBottom;
  //     WidgetsBinding.instance.addPostFrameCallback((_) {
  //       if (!mounted) return;
  //       if (_autoScrollPending && _listCtrl.hasClients) {
  //         final max = _listCtrl.position.maxScrollExtent;
  //         if (max > 0) {
  //           _listCtrl.animateTo(
  //             max,
  //             duration: const Duration(milliseconds: 120),
  //             curve: Curves.easeOut,
  //           );
  //         }
  //       }
  //       _lastFramesLen = widget.frames.length;
  //       _autoScrollPending = false;
  //     });
  //   } else {
  //     _lastFramesLen = widget.frames.length;
  //   }
  // }

  @override
  void didUpdateWidget(covariant WsDetailsPanel oldWidget) {
    super.didUpdateWidget(oldWidget);
    // TEMPORARILY DISABLED: Auto-scroll logic
    // _maybeAutoScrollToBottomOnNewFrames();
    // Clean up caches when frames list changes
    if (oldWidget.frames != widget.frames) {
      _cleanupFrameCaches();
    }
    // Update cache when frames or any filters change
    if (oldWidget.frames != widget.frames ||
        oldWidget.hideHeartbeats != widget.hideHeartbeats ||
        oldWidget.opcodeFilter != widget.opcodeFilter ||
        oldWidget.directionFilter != widget.directionFilter ||
        oldWidget.fbOpFilter != widget.fbOpFilter ||
        oldWidget.fbStatusFilter != widget.fbStatusFilter ||
        oldWidget.fbPathFilter != widget.fbPathFilter) {
      _updateVisibleFramesCache();
      // Reindex search when visible frames change
      if (_showGlobalSearch && _searchCtrl.text.trim().isNotEmpty) {
        _reindexGlobalMatches();
      }
    }
  }

  void _scrollToFrame(String frameId) {
    final navSeq = ++_searchNavSeq;
    unawaited(_scrollToFrameForSearch(frameId, navSeq: navSeq));
  }

  Future<void> _scrollToFrameForSearch(
    String frameId, {
    required int navSeq,
  }) async {
    final visibleFrames = _cachedVisibleFrames;

    int idx = -1;
    String? resolvedFrameKey;
    for (int i = 0; i < visibleFrames.length; i++) {
      final f = visibleFrames[i];
      final compositeKey = _frameKeyOf(f);
      if (compositeKey == frameId || (f['id'] ?? '').toString() == frameId) {
        idx = i;
        resolvedFrameKey = compositeKey;
        break;
      }
    }

    if (idx < 0 || !_listCtrl.hasClients) return;

    final frameKey = resolvedFrameKey ?? frameId;
    final tileKey = _tileKeyFor(frameKey);

    final pos = _listCtrl.position;
    final max = pos.maxScrollExtent;

    const estimatedItemExtent = 64.0;
    final estimatedTarget = (idx * estimatedItemExtent).toDouble().clamp(
      0.0,
      max,
    );

    Future<void> animateTo(double offset, Duration duration) async {
      if (!_listCtrl.hasClients) return;
      try {
        await _listCtrl.animateTo(
          offset,
          duration: duration,
          curve: Curves.easeOutCubic,
        );
      } catch (_) {}
    }

    await animateTo(estimatedTarget, const Duration(milliseconds: 220));
    if (!mounted || navSeq != _searchNavSeq) return;

    for (int attempt = 0; attempt < 10; attempt++) {
      if (!mounted || navSeq != _searchNavSeq) return;

      await WidgetsBinding.instance.endOfFrame;
      if (!mounted || navSeq != _searchNavSeq) return;

      final ctx = tileKey.currentContext;
      if (ctx != null) {
        try {
          await Scrollable.ensureVisible(
            ctx,
            duration: const Duration(milliseconds: 140),
            curve: Curves.easeOutCubic,
            alignment: 0.08,
          );
        } catch (_) {}
        return;
      }

      final step = pos.viewportDimension * 0.85;
      final cur = pos.pixels;
      final dir = cur < estimatedTarget ? 1.0 : -1.0;
      final next = (cur + dir * step).clamp(0.0, max);
      if ((next - cur).abs() < 1) return;
      await animateTo(next, const Duration(milliseconds: 160));
    }
  }

  // (header is rendered via _buildTitleRich)

  dynamic _decodeNestedJsonStrings(dynamic v) {
    if (v is Map) {
      final out = <String, dynamic>{};
      v.forEach((k, val) {
        out[k.toString()] = _decodeNestedJsonStrings(
          _maybeDecodeJsonString(val),
        );
      });
      return out;
    }
    if (v is List) {
      return v
          .map((e) => _decodeNestedJsonStrings(_maybeDecodeJsonString(e)))
          .toList();
    }
    return v;
  }

  dynamic _maybeDecodeJsonString(dynamic val) {
    if (val is String) {
      final s = val.trim();
      if (s.isNotEmpty && (s.startsWith('{') || s.startsWith('['))) {
        try {
          final parsed = jsonDecode(s);
          // Only interested in objects/arrays - don't touch string primitives
          if (parsed is Map || parsed is List) {
            return _decodeNestedJsonStrings(parsed);
          }
        } catch (_) {}
      }
    }
    return val;
  }

  // Try to get normalized object for header (if preview is JSON)
  dynamic _tryDecodeNormalizedForHeader(String preview) {
    final t = preview.trim();
    if (t.isEmpty) return null;
    if (!(t.startsWith('{') || t.startsWith('['))) return null;
    try {
      final decoded = jsonDecode(t);
      return _decodeNestedJsonStrings(decoded);
    } catch (_) {
      return null;
    }
  }

  // Single-line JSON render with token highlighting for header
  Widget _buildTitleRich(BuildContext context, String preview) {
    // 1) Pure JSON entirely
    final wholeObj = _tryDecodeNormalizedForHeader(preview);
    if (wholeObj != null) {
      final spans = _buildInlineJsonSpans(context, wholeObj);
      return Text.rich(
        TextSpan(children: spans),
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        style: context.appText.body,
      );
    }
    // 2) JSON as part of a string (e.g., socket.io '42/ns,[...]')
    final jsonPart = _extractJsonPayload(preview);
    if (jsonPart != null) {
      try {
        final decoded = jsonDecode(jsonPart);
        final normalized = _decodeNestedJsonStrings(decoded);
        final idx = preview.indexOf(jsonPart);
        final before = idx > 0 ? preview.substring(0, idx) : '';
        final afterIdx = idx + jsonPart.length;
        final after = (afterIdx < preview.length)
            ? preview.substring(afterIdx)
            : '';
        final List<InlineSpan> children = [];
        if (before.isNotEmpty) {
          children.add(TextSpan(text: before, style: context.appText.body));
        }
        children.addAll(_buildInlineJsonSpans(context, normalized));
        if (after.isNotEmpty) {
          children.add(TextSpan(text: after, style: context.appText.body));
        }
        return Text.rich(
          TextSpan(children: children),
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
          style: context.appText.body,
        );
      } catch (_) {
        // if parsing fails - fall back to default render
      }
    }
    // 3) Not JSON - just text
    return Text(
      preview,
      maxLines: 1,
      overflow: TextOverflow.ellipsis,
      style: context.appText.body,
    );
  }

  List<InlineSpan> _buildInlineJsonSpans(BuildContext context, dynamic node) =>
      buildInlineJsonSpans(context, node);

  void _reindexGlobalMatches() {
    // Guard against concurrent reindexing (Bug #3 fix)
    if (_isReindexing) return;
    _isReindexing = true;

    try {
      final query = _searchCtrl.text.trim();
      _globalTotalMatches = 0;
      _frameLocalFocusedIndex.clear();
      if (query.isEmpty) {
        _frameMatchCounts.clear();
        // Clear match keys when query is empty (Bug #9 fix)
        _frameMatchKeys.clear();
        setState(() {
          _globalFocusedIndex = 0;
        });
        return;
      }
      // Use cached visible frames for consistent filtering
      final visibleFrames = _cachedVisibleFrames;
      int acc = 0;
      for (final fm in visibleFrames) {
        final preview = (fm['preview'] ?? '').toString();
        final extractedJson = _extractJsonPayload(preview);
        final computedCnt = (extractedJson != null && (_pretty || _tree))
            ? _countMatchesIn(extractedJson)
            : _countMatchesIn(preview);
        final frameKey = _frameKeyOf(fm);
        final effectiveCnt = computedCnt;
        if (effectiveCnt > 0) {
          _frameMatchCounts[frameKey] = effectiveCnt;
          _globalTotalMatches += effectiveCnt;
          // Pre-compute local focused index for this frame
          final end = acc + effectiveCnt - 1;
          if (_globalFocusedIndex >= acc && _globalFocusedIndex <= end) {
            _frameLocalFocusedIndex[frameKey] = _globalFocusedIndex - acc;
          }
          acc += effectiveCnt;
        } else {
          _frameMatchCounts.remove(frameKey);
        }
      }
      if (_globalTotalMatches == 0) {
        setState(() {
          _globalFocusedIndex = 0;
        });
      } else if (_globalFocusedIndex >= _globalTotalMatches) {
        setState(() {
          _globalFocusedIndex = 0;
        });
      } else {
        setState(() {});
      }
    } finally {
      _isReindexing = false;
    }
  }

  int _countMatchesIn(String text) {
    final q = _searchCtrl.text.trim();
    if (q.isEmpty) return 0;
    if (_useRegex) {
      // Bug #24 fix: In regex mode, user controls word boundaries via \b
      // Don't apply manual wholeWord checks - just use the regex as-is
      RegExp? re;
      try {
        re = RegExp(q, caseSensitive: _matchCase);
      } catch (_) {
        return 0;
      }
      return re.allMatches(text).length;
    }
    final src = _matchCase ? text : text.toLowerCase();
    final query = _matchCase ? q : q.toLowerCase();
    int from = 0;
    int c = 0;
    bool isWordChar(String ch) {
      final code = ch.codeUnitAt(0);
      final isAZ = (code >= 65 && code <= 90) || (code >= 97 && code <= 122);
      final is09 = (code >= 48 && code <= 57);
      return isAZ || is09 || ch == '_';
    }

    while (true) {
      final idx = src.indexOf(query, from);
      if (idx < 0) break;
      if (_wholeWord) {
        final left = idx - 1 >= 0 ? src.substring(idx - 1, idx) : null;
        final right = (idx + query.length) < src.length
            ? src.substring(idx + query.length, idx + query.length + 1)
            : null;
        final leftOk = left == null || !isWordChar(left);
        final rightOk = right == null || !isWordChar(right);
        if (!(leftOk && rightOk)) {
          from = idx + 1;
          continue;
        }
      }
      c++;
      from = idx + query.length;
    }
    return c;
  }

  (String?, int) _resolveGlobalIndexToFrame(int gIndex) {
    int acc = 0;
    // Use cached visible frames for consistent filtering
    final visibleFrames = _cachedVisibleFrames;
    for (final f in visibleFrames) {
      final idStr = _frameKeyOf(f);
      final cnt = _frameMatchCounts[idStr] ?? 0;
      if (cnt <= 0) continue;
      final end = acc + cnt - 1;
      if (gIndex >= acc && gIndex <= end) {
        final local = gIndex - acc;
        return (idStr, local);
      }
      acc += cnt;
    }
    return (null, 0);
  }

  void _focusGlobal(int gIndex) {
    if (_globalTotalMatches <= 0) return;

    final navSeq = ++_searchNavSeq;
    _frameLocalFocusedIndex.clear();

    final (fid, local) = _resolveGlobalIndexToFrame(gIndex);
    if (fid == null) return;

    _globalFocusedIndex = gIndex;
    _frameLocalFocusedIndex[fid] = local;

    final keys = _frameMatchKeys[fid];
    final hasLiveAnchor =
        keys != null &&
        local >= 0 &&
        local < keys.length &&
        keys[local].currentContext != null;

    if (_searchExpandedFid != null && _searchExpandedFid != fid) {
      _tileControllerFor(_searchExpandedFid!).collapse();
    }
    _searchExpandedFid = fid;

    _pendingFocus = _PendingFocus(fid, local, navSeq);
    setState(() {});

    if (hasLiveAnchor) {
      return;
    }

    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted || navSeq != _searchNavSeq) return;
      unawaited(() async {
        await _scrollToFrameForSearch(fid, navSeq: navSeq);
        if (!mounted || navSeq != _searchNavSeq) return;

        final controller = _tileControllerFor(fid);
        controller.expand();

        await WidgetsBinding.instance.endOfFrame;
        if (!mounted || navSeq != _searchNavSeq) return;

        controller.expand();
      }());
    });
  }

  Future<void> _scrollToMatchInFrame(
    String frameId,
    GlobalKey key, {
    required int navSeq,
  }) async {
    if (!mounted || navSeq != _searchNavSeq) return;

    Future<void> ensureFrameVisible(Duration duration) async {
      final tileKey = _tileKeyFor(frameId);
      var ctx = tileKey.currentContext;
      if (ctx == null) {
        await _scrollToFrameForSearch(frameId, navSeq: navSeq);
        if (!mounted || navSeq != _searchNavSeq) return;
        ctx = tileKey.currentContext;
      }
      if (ctx == null) return;

      final ro = ctx.findRenderObject();
      if (ro == null) return;

      if (_listCtrl.hasClients) {
        try {
          await _listCtrl.position.ensureVisible(
            ro,
            duration: duration,
            curve: Curves.easeOutCubic,
            alignment: 0.08,
          );
        } catch (_) {}
      } else {
        try {
          await Scrollable.ensureVisible(
            ctx,
            duration: duration,
            curve: Curves.easeOutCubic,
            alignment: 0.08,
          );
        } catch (_) {}
      }
    }

    Future<void> ensureMatchVisible(Duration duration) async {
      final ctx = key.currentContext;
      if (ctx == null) return;
      final ro = ctx.findRenderObject();
      if (ro == null) return;

      final innerCtrl = _innerScrollControllerFor(frameId);
      if (!innerCtrl.hasClients) {
        await WidgetsBinding.instance.endOfFrame;
        if (!mounted || navSeq != _searchNavSeq) return;
      }
      if (!innerCtrl.hasClients) return;
      try {
        await innerCtrl.position.ensureVisible(
          ro,
          duration: duration,
          curve: Curves.easeOutCubic,
          alignment: 0.28,
        );
      } catch (_) {}
    }

    void correct(Duration duration) {
      if (!mounted || navSeq != _searchNavSeq) return;
      unawaited(ensureFrameVisible(duration));
      unawaited(ensureMatchVisible(duration));
    }

    _suspendNestedScrollSync = true;
    try {
      // Сначала делаем видимым сам фрейм, потом — конкретное совпадение внутри него.
      await ensureFrameVisible(const Duration(milliseconds: 220));
      if (!mounted || navSeq != _searchNavSeq) return;

      await WidgetsBinding.instance.endOfFrame;
      if (!mounted || navSeq != _searchNavSeq) return;

      await ensureMatchVisible(const Duration(milliseconds: 180));
      if (!mounted || navSeq != _searchNavSeq) return;

      WidgetsBinding.instance.addPostFrameCallback((_) {
        correct(const Duration(milliseconds: 120));
      });

      Future<void>.delayed(const Duration(milliseconds: 260), () {
        correct(const Duration(milliseconds: 120));
      });
    } finally {
      Future<void>.delayed(const Duration(milliseconds: 500), () {
        if (!mounted) return;
        if (navSeq != _searchNavSeq) return;
        _suspendNestedScrollSync = false;
      });
    }
  }

  void _gotoNext() {
    if (_globalTotalMatches <= 0) return;
    final next = (_globalFocusedIndex + 1) % _globalTotalMatches;
    _focusGlobal(next);
  }

  void _gotoPrev() {
    if (_globalTotalMatches <= 0) return;
    final prev = (_globalFocusedIndex - 1) < 0
        ? _globalTotalMatches - 1
        : _globalFocusedIndex - 1;
    _focusGlobal(prev);
  }

  @override
  void dispose() {
    _searchCtrl.dispose();
    _searchFocus.dispose();
    _listCtrl.dispose();
    for (final s in _localSearch.values) {
      s.dispose();
    }
    for (final controller in _frameInnerScrollControllers.values) {
      controller.dispose();
    }
    // Dispose all ExpansibleControllers to prevent memory leaks
    for (final controller in _frameTileControllers.values) {
      controller.dispose();
    }
    super.dispose();
  }

  void _onChildMatches(String frameId, int count, List<GlobalKey> keys) {
    if (!_showGlobalSearch) {
      return; // local search should not overwrite global keys
    }
    _frameMatchKeys[frameId] = keys;
    // if waiting for focus on this specific frame — try to navigate
    final pending = _pendingFocus;
    if (pending != null &&
        pending.frameId == frameId &&
        pending.navSeq == _searchNavSeq) {
      _pendingFocus = null;
      final local = pending.localIndex;
      if (local >= 0 && local < keys.length) {
        unawaited(
          _scrollToMatchInFrame(frameId, keys[local], navSeq: pending.navSeq),
        );
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    // Use pre-computed cached visible frames
    final visibleFrames = _cachedVisibleFrames;

    final timelineSection = Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4),
      child: (_showTimeline)
          ? Column(
              children: [
                FramesTimeline(
                  frames: visibleFrames,
                  height: 50,
                  onFrameTap: _scrollToFrame,
                  onBrushChanged: (r) {
                    _brushRange = r;
                    _updateVisibleFramesCache();
                    // Reindex search when brush filter changes
                    if (_showGlobalSearch &&
                        _searchCtrl.text.trim().isNotEmpty) {
                      _reindexGlobalMatches();
                    }
                    setState(() {});
                  },
                  // hover: only highlight on timeline, no list scrolling
                ),
                const Align(
                  alignment: Alignment.centerLeft,
                  child: FramesTimelineLegend(),
                ),
                const SizedBox(height: 4),
                if (_showGlobalSearch) _buildGlobalSearchBar(context),
              ],
            )
          : (_showGlobalSearch
                ? Column(
                    children: [
                      _buildGlobalSearchBar(context),
                      const SizedBox(height: 8),
                    ],
                  )
                : const SizedBox.shrink()),
    );

    return Column(
      children: [
        // Closed / error banner
        if (widget.isClosed)
          _WsClosedBanner(closedAt: widget.closedAt, error: widget.error),
        // Timeline and search above header
        timelineSection,
        Expanded(
          child: _Card(
            title: Text(_isFirebaseSession ? 'RTDB events' : 'Frames'),
            actions: [
              FilterChip(
                label: const Text('Pretty', style: TextStyle(fontSize: 12)),
                selected: _pretty && !_tree,
                onSelected: (v) {
                  setState(() {
                    _pretty = v;
                    if (v) _tree = false;
                    // Reindex search if active (match counts depend on pretty/tree)
                    if (_showGlobalSearch &&
                        _searchCtrl.text.trim().isNotEmpty) {
                      _reindexGlobalMatches();
                    }
                  });
                },
              ),
              const SizedBox(width: 6),
              FilterChip(
                label: const Text('Tree', style: TextStyle(fontSize: 12)),
                selected: _tree,
                onSelected: (v) {
                  setState(() {
                    _tree = v;
                    if (v) _pretty = false;
                    // Reindex search if active (match counts depend on pretty/tree)
                    if (_showGlobalSearch &&
                        _searchCtrl.text.trim().isNotEmpty) {
                      _reindexGlobalMatches();
                    }
                  });
                },
              ),
              const SizedBox(width: 6),
              Builder(
                builder: (context) {
                  final c = context.appColors;
                  final cs = Theme.of(context).colorScheme;
                  final sel = _showTimeline;
                  return FilterChip(
                    avatar: Icon(
                      Icons.timeline,
                      size: 14,
                      color: sel ? c.primary : c.textSecondary,
                    ),
                    label: const Text(
                      'Timeline',
                      style: TextStyle(fontSize: 12),
                    ),
                    selected: sel,
                    showCheckmark: false,
                    shape: const StadiumBorder(),
                    side: BorderSide(
                      color: sel ? c.primary : c.border,
                      width: sel ? 1.5 : 1,
                    ),
                    selectedColor: cs.primary.withOpacity(0.18),
                    backgroundColor: cs.surface,
                    onSelected: (v) {
                      setState(() {
                        _showTimeline = v;
                      });
                    },
                  );
                },
              ),
              const SizedBox(width: 6),
              // Global search
              if (!_showGlobalSearch)
                IconButton(
                  tooltip: 'Search',
                  icon: const Icon(Icons.search, size: 18),
                  onPressed: () {
                    setState(() {
                      _showGlobalSearch = true;
                    });
                    WidgetsBinding.instance.addPostFrameCallback((_) {
                      _searchFocus.requestFocus();
                    });
                  },
                ),
              IconButton(
                tooltip: 'Filters',
                icon: const Icon(Icons.filter_list, size: 18),
                onPressed: () => widget._openFilters(context),
              ),
              if (widget.onCloseFullscreen != null)
                IconButton(
                  tooltip: 'Exit fullscreen',
                  icon: const Icon(Icons.close_fullscreen, size: 18),
                  onPressed: widget.onCloseFullscreen,
                )
              else
                IconButton(
                  tooltip: 'Fullscreen',
                  icon: const Icon(Icons.open_in_full, size: 18),
                  onPressed: () => _openFullscreen(context),
                ),
            ],
            child: Column(
              children: [
                Expanded(
                  child: ListView.builder(
                    controller: _listCtrl,
                    itemCount: visibleFrames.length,
                    itemBuilder: (_, i) {
                      final f = visibleFrames[i];
                      final preview = (f['preview'] ?? '').toString();
                      final fbEvent = FirebaseEventData.tryParse(preview);
                      final String? extractedJson = fbEvent?.payload != null
                          ? jsonEncode(fbEvent!.payload)
                          : _extractJsonPayload(preview);
                      final dir = (f['direction'] ?? '').toString();
                      final isDown = dir == 'upstream->client';

                      final ts = _fmtTime((f['ts'] ?? '').toString());
                      final opcode = (f['opcode'] ?? '').toString();
                      final size = (f['size'] ?? 0).toString();
                      final isWsPingPong = opcode == 'ping' || opcode == 'pong';
                      final isEnginePingPong =
                          opcode == 'text' &&
                          (preview == '2' || preview == '3') &&
                          size == '1';
                      final isHeartbeat = isWsPingPong || isEnginePingPong;
                      final icon = Icon(
                        isDown ? Icons.south : Icons.north,
                        size: isHeartbeat ? 10 : 16,
                        color: isDown
                            ? context.appColors.success
                            : context.appColors.primary,
                      );
                      if (isHeartbeat) {
                        final label = isWsPingPong
                            ? opcode.toUpperCase()
                            : (preview == '2' ? 'PING' : 'PONG');
                        return ListTile(
                          key: ValueKey('heartbeat_${_frameKeyOf(f)}'),
                          dense: true,
                          contentPadding: const EdgeInsets.symmetric(
                            horizontal: 8,
                          ),
                          leading: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              icon,
                              const SizedBox(width: 6),
                              Text(label),
                            ],
                          ),
                          trailing: Text(
                            ts,
                            style: Theme.of(context).textTheme.labelSmall
                                ?.copyWith(
                                  color: context.appColors.textSecondary,
                                ),
                          ),
                        );
                      }
                      final frameKey = _frameKeyOf(f);
                      // Use pre-computed local focused index (O(1) lookup)
                      final localFocusedIndex =
                          _frameLocalFocusedIndex[frameKey] ?? -1;

                      return ExpansionTile(
                        key: _tileKeyFor(frameKey),
                        controller: _tileControllerFor(frameKey),
                        tilePadding: const EdgeInsets.symmetric(horizontal: 8),
                        dense: true,
                        leading: fbEvent != null
                            ? icon
                            : Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  icon,
                                  const SizedBox(width: 6),
                                  Text(opcode),
                                ],
                              ),
                        title: Builder(
                          builder: (context) {
                            if (fbEvent == null) {
                              return _buildTitleRich(context, preview);
                            }
                            final local = _localFor(frameKey);
                            final String activeQuery = local.show
                                ? local.controller.text.trim()
                                : (_showGlobalSearch
                                      ? _searchCtrl.text.trim()
                                      : '');
                            final bool activeMatchCase = local.show
                                ? local.matchCase
                                : _matchCase;
                            final bool activeWholeWord = local.show
                                ? local.wholeWord
                                : _wholeWord;
                            final bool activeUseRegex = local.show
                                ? local.useRegex
                                : _useRegex;
                            final int activeFocusedIndex = local.show
                                ? local.focusedIndex
                                : (localFocusedIndex < 0
                                      ? 0
                                      : localFocusedIndex);

                            final cfg = JsonSearchConfig(
                              query: activeQuery,
                              matchCase: activeMatchCase,
                              wholeWord: activeWholeWord,
                              useRegex: activeUseRegex,
                              focusedIndex: activeFocusedIndex,
                              anchorScope: null,
                              onRebuilt: null,
                            );

                            final previewText =
                                FirebaseEventTitle.buildCompactPreview(
                                  fbEvent.payload,
                                );

                            return FirebaseEventTitle(
                              event: fbEvent,
                              contentPreview: previewText,
                              timestamp: ts,
                              search: cfg,
                            );
                          },
                        ),
                        subtitle: Row(
                          children: [
                            Expanded(
                              child: Text(
                                '$size B',
                                style: Theme.of(context).textTheme.labelSmall
                                    ?.copyWith(
                                      color: context.appColors.textSecondary,
                                    ),
                              ),
                            ),
                          ],
                        ),
                        trailing: fbEvent != null
                            ? null
                            : Text(
                                ts,
                                style: Theme.of(context).textTheme.labelSmall
                                    ?.copyWith(
                                      color: context.appColors.textSecondary,
                                    ),
                              ),
                        children: [
                          Builder(
                            builder: (context) {
                              final local = _localFor(frameKey);
                              final String activeQuery = local.show
                                  ? local.controller.text.trim()
                                  : (_showGlobalSearch
                                        ? _searchCtrl.text.trim()
                                        : '');
                              final bool activeMatchCase = local.show
                                  ? local.matchCase
                                  : _matchCase;
                              final bool activeWholeWord = local.show
                                  ? local.wholeWord
                                  : _wholeWord;
                              final bool activeUseRegex = local.show
                                  ? local.useRegex
                                  : _useRegex;
                              final int activeFocusedIndex = local.show
                                  ? local.focusedIndex
                                  : (localFocusedIndex < 0
                                        ? 0
                                        : localFocusedIndex);

                              final contentWidget = Builder(
                                builder: (context) {
                                  final cfg = JsonSearchConfig(
                                    query: activeQuery,
                                    matchCase: activeMatchCase,
                                    wholeWord: activeWholeWord,
                                    useRegex: activeUseRegex,
                                    focusedIndex: activeFocusedIndex,
                                    anchorScope: 'ws:$frameKey',
                                    onRebuilt: (count, keys) {
                                      if (local.show) {
                                        final pending =
                                            _pendingLocalFocus[frameKey];
                                        final clampedIndex = keys.isEmpty
                                            ? 0
                                            : local.focusedIndex.clamp(
                                                0,
                                                keys.length - 1,
                                              );

                                        final shouldUpdateUi =
                                            local.keys.length != keys.length ||
                                            local.focusedIndex != clampedIndex;
                                        if (shouldUpdateUi) {
                                          setState(() {
                                            local.keys = keys;
                                            local.focusedIndex = clampedIndex;
                                          });
                                        } else {
                                          local.keys = keys;
                                        }

                                        if (pending != null &&
                                            pending.navSeq == _searchNavSeq) {
                                          _pendingLocalFocus.remove(frameKey);
                                          final idx = pending.localIndex.clamp(
                                            0,
                                            keys.length - 1,
                                          );
                                          if (keys.isNotEmpty) {
                                            unawaited(
                                              _scrollToMatchInFrame(
                                                frameKey,
                                                keys[idx],
                                                navSeq: pending.navSeq,
                                              ),
                                            );
                                          }
                                        }
                                      } else if (_showGlobalSearch) {
                                        _onChildMatches(frameKey, count, keys);
                                        // Bug #7 fix: Don't call _reindexGlobalMatches() here
                                        // to avoid infinite loop. The count from child widget
                                        // is the same as what _reindexGlobalMatches computes.
                                        // Just update the map directly.
                                        if ((_frameMatchCounts[frameKey] ??
                                                -1) !=
                                            count) {
                                          _frameMatchCounts[frameKey] = count;
                                        }
                                      }
                                    },
                                  );
                                  if (extractedJson != null) {
                                    try {
                                      final parsed = jsonDecode(extractedJson);
                                      final normalized =
                                          _decodeNestedJsonStrings(parsed);
                                      // Firebase: внутри показываем payload, а не envelope.
                                      final effective =
                                          fbEvent?.payload ?? normalized;
                                      if (_tree) {
                                        return JsonTreeRich(
                                          data: effective,
                                          search: cfg,
                                        );
                                      }
                                      if (_pretty) {
                                        return JsonPrettyRich(
                                          data: effective,
                                          search: cfg,
                                        );
                                      }
                                    } catch (_) {}
                                  }
                                  final contentText = fbEvent?.payload != null
                                      ? extractedJson!
                                      : preview;
                                  return SearchableTextRich(
                                    text: contentText,
                                    search: cfg,
                                    style: context.appText.monospace,
                                  );
                                },
                              );

                              // Limit frame height and scroll only JSON content,
                              // copy/search panel floats above the content.
                              return ConstrainedBox(
                                constraints: const BoxConstraints(
                                  maxHeight: 400, // maximum, then scroll
                                ),
                                child: Stack(
                                  children: [
                                    NotificationListener<ScrollNotification>(
                                      onNotification: (notification) {
                                        if (_suspendNestedScrollSync) {
                                          return false;
                                        }
                                        final metrics = notification.metrics;
                                        if (notification
                                                is ScrollUpdateNotification &&
                                            notification.scrollDelta != null &&
                                            metrics.axis == Axis.vertical) {
                                          final delta =
                                              notification.scrollDelta!;
                                          if (_listCtrl.hasClients) {
                                            final parentPos =
                                                _listCtrl.position;
                                            // Scrolled frame content to bottom
                                            // and continuing to scroll down -
                                            // transfer scroll to parent list.
                                            if (metrics.pixels >=
                                                    metrics.maxScrollExtent &&
                                                delta > 0) {
                                              final target =
                                                  (parentPos.pixels + delta)
                                                      .clamp(
                                                        0.0,
                                                        parentPos
                                                            .maxScrollExtent,
                                                      );
                                              if (target != parentPos.pixels) {
                                                _listCtrl.jumpTo(target);
                                              }
                                            }
                                            // Scrolled frame content to top
                                            // and continuing to scroll up -
                                            // transfer scroll to parent.
                                            if (metrics.pixels <= 0 &&
                                                delta < 0) {
                                              final target =
                                                  (parentPos.pixels + delta)
                                                      .clamp(
                                                        0.0,
                                                        parentPos
                                                            .maxScrollExtent,
                                                      );
                                              if (target != parentPos.pixels) {
                                                _listCtrl.jumpTo(target);
                                              }
                                            }
                                          }
                                        }
                                        return false;
                                      },
                                      child: SingleChildScrollView(
                                        controller: _innerScrollControllerFor(
                                          frameKey,
                                        ),
                                        primary: false,
                                        child: Container(
                                          alignment: Alignment.centerLeft,
                                          // small internal padding, without extra empty space at top
                                          padding: const EdgeInsets.all(8),
                                          child: contentWidget,
                                        ),
                                      ),
                                    ),
                                    Positioned(
                                      top: 6,
                                      left: 6,
                                      right: 6,
                                      child: Align(
                                        alignment: Alignment.topRight,
                                        child: Row(
                                          mainAxisSize: MainAxisSize.min,
                                          children: [
                                            IconButton(
                                              tooltip: 'Copy content',
                                              icon: const Icon(
                                                Icons.copy,
                                                size: 18,
                                              ),
                                              onPressed: () {
                                                final toCopy =
                                                    extractedJson ?? preview;
                                                Clipboard.setData(
                                                  ClipboardData(text: toCopy),
                                                );
                                              },
                                            ),
                                            const SizedBox(width: 6),
                                            if (local.show)
                                              CommonSearchBar(
                                                controller: local.controller,
                                                focusNode: local.focus,
                                                countText: local.keys.isEmpty
                                                    ? '0/0'
                                                    : '${local.focusedIndex + 1}/${local.keys.length}',
                                                matchCase: local.matchCase,
                                                wholeWord: local.wholeWord,
                                                useRegex: local.useRegex,
                                                canNavigate:
                                                    local.keys.isNotEmpty,
                                                onChanged: () {
                                                  setState(() {
                                                    local.focusedIndex = 0;
                                                  });
                                                },
                                                onNext: () =>
                                                    _localGotoNext(frameKey),
                                                onPrev: () =>
                                                    _localGotoPrev(frameKey),
                                                onClose: () {
                                                  setState(() {
                                                    local.show = false;
                                                    local.controller.clear();
                                                    local.focusedIndex = 0;
                                                    local.keys = const [];
                                                    local.focus.unfocus();
                                                  });
                                                },
                                                onToggleMatchCase: () {
                                                  setState(() {
                                                    local.matchCase =
                                                        !local.matchCase;
                                                  });
                                                },
                                                onToggleWholeWord: () {
                                                  setState(() {
                                                    local.wholeWord =
                                                        !local.wholeWord;
                                                  });
                                                },
                                                onToggleRegex: () {
                                                  setState(() {
                                                    local.useRegex =
                                                        !local.useRegex;
                                                  });
                                                },
                                              )
                                            else
                                              IconButton(
                                                tooltip: 'Search in frame',
                                                icon: const Icon(
                                                  Icons.search,
                                                  size: 18,
                                                ),
                                                onPressed: () {
                                                  setState(() {
                                                    local.show = true;
                                                  });
                                                  WidgetsBinding.instance
                                                      .addPostFrameCallback((
                                                        _,
                                                      ) {
                                                        local.focus
                                                            .requestFocus();
                                                      });
                                                },
                                              ),
                                          ],
                                        ),
                                      ),
                                    ),
                                  ],
                                ),
                              );
                            },
                          ),
                        ],
                      );
                    },
                  ),
                ),
              ],
            ),
          ),
        ),
        /*
      const VerticalDivider(width: 1),
      SizedBox(width: 200, child: _Card(
        title: 'Events',
        child: ListView.builder(
          itemCount: widget.events.length,
          itemBuilder: (_, i) {
            final e = widget.events[i] as Map<String, dynamic>;
            if (!_eventMatches(e)) { return const SizedBox.shrink(); }
            final args = (e['argsPreview'] ?? '').toString();
            return ExpansionTile(
              tilePadding: const EdgeInsets.symmetric(horizontal: 8),
              dense: true,
              title: Text('${e['namespace']} ${e['event']}', style: context.appText.subtitle),
              subtitle: Text(args, maxLines: 1, overflow: TextOverflow.ellipsis, style: context.appText.body),
              trailing: e['ackId'] != null ? Text('#${e['ackId']}') : null,
              children: [
                Container(alignment: Alignment.centerLeft, padding: const EdgeInsets.all(8), child: JsonViewer(jsonString: args)),
              ],
            );
          },
        ),
      )),
      */
      ],
    );
  }

  void _openFullscreen(BuildContext context) {
    showDialog<void>(
      context: context,
      builder: (_) => _WsFramesFullscreenDialog(
        frames: widget.frames,
        events: widget.events,
        initialOpcodeFilter: widget.opcodeFilter,
        initialDirectionFilter: widget.directionFilter,
        initialNamespaceText: widget.namespaceCtrl.text,
        initialHideHeartbeats: widget.hideHeartbeats,
        isClosed: widget.isClosed,
        closedAt: widget.closedAt,
        error: widget.error,
        initialFbOpFilter: widget.fbOpFilter,
        initialFbStatusFilter: widget.fbStatusFilter,
        initialFbPathFilter: widget.fbPathFilter,
      ),
    );
  }

  Widget _buildGlobalSearchBar(BuildContext context) {
    final countText = _globalTotalMatches == 0
        ? '0/0'
        : '${(_globalFocusedIndex + 1)}/${_globalTotalMatches}';
    return Padding(
      padding: const EdgeInsets.only(bottom: 8),
      child: CommonSearchBar(
        controller: _searchCtrl,
        focusNode: _searchFocus,
        countText: countText,
        matchCase: _matchCase,
        wholeWord: _wholeWord,
        useRegex: _useRegex,
        canNavigate: _globalTotalMatches > 0,
        onChanged: () {
          setState(() {
            _globalFocusedIndex = 0;
          });
          _reindexGlobalMatches();
        },
        onNext: _gotoNext,
        onPrev: _gotoPrev,
        onClose: () {
          // Collapse tile that was expanded by search
          if (_searchExpandedFid != null) {
            _tileControllerFor(_searchExpandedFid!).collapse();
            _searchExpandedFid = null;
          }
          setState(() {
            _showGlobalSearch = false;
            _searchCtrl.clear();
            _frameMatchCounts.clear();
            _frameMatchKeys.clear();
            _globalFocusedIndex = 0;
            _globalTotalMatches = 0;
          });
        },
        onToggleMatchCase: () {
          setState(() {
            _matchCase = !_matchCase;
          });
          _reindexGlobalMatches();
        },
        onToggleWholeWord: () {
          setState(() {
            _wholeWord = !_wholeWord;
          });
          _reindexGlobalMatches();
        },
        onToggleRegex: () {
          setState(() {
            _useRegex = !_useRegex;
          });
          _reindexGlobalMatches();
        },
      ),
    );
  }
}

class _Card extends StatelessWidget {
  const _Card({required this.title, required this.child, this.actions});
  final Widget title;
  final Widget child;
  final List<Widget>? actions;
  @override
  Widget build(BuildContext context) {
    return Card(
      elevation: 1,
      margin: const EdgeInsets.all(8),
      child: Padding(
        padding: const EdgeInsets.all(8),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Expanded(
                  child: DefaultTextStyle.merge(
                    style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                    child: title,
                  ),
                ),
                if (actions != null) ...actions!,
              ],
            ),
            const SizedBox(height: 8),
            Expanded(child: child),
          ],
        ),
      ),
    );
  }
}

extension on WsDetailsPanel {
  bool get _isFirebaseExt {
    for (final item in frames) {
      final frame = item as Map<String, dynamic>;
      final preview = (frame['preview'] ?? '').toString().trim();
      if (!(preview.startsWith('{') || preview.startsWith('['))) continue;
      try {
        final decoded = jsonDecode(preview);
        if (decoded is Map<String, dynamic> &&
            decoded['type'] == 'firebase_database') {
          return true;
        }
      } catch (_) {
        // skip
      }
    }
    return false;
  }

  void _openFilters(BuildContext context) {
    if (_isFirebaseExt) {
      _openFirebaseFilters(context);
    } else {
      _openWsFilters(context);
    }
  }

  void _openWsFilters(BuildContext context) {
    showModalBottomSheet(
      context: context,
      builder: (_) {
        String localOpcode = opcodeFilter;
        String localDirection = directionFilter;
        bool localHideHeartbeats = hideHeartbeats;
        return StatefulBuilder(
          builder: (context, setState) {
            return Padding(
              padding: const EdgeInsets.all(12),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'WebSocket filters',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      const Text('Opcode:'),
                      const SizedBox(width: 8),
                      DropdownButton<String>(
                        value: localOpcode,
                        style: TextStyle(
                          color: Theme.of(context).colorScheme.onSurface,
                        ),
                        items: const [
                          DropdownMenuItem(value: 'all', child: Text('Any')),
                          DropdownMenuItem(value: 'text', child: Text('Text')),
                          DropdownMenuItem(
                            value: 'binary',
                            child: Text('Binary'),
                          ),
                          DropdownMenuItem(value: 'ping', child: Text('Ping')),
                          DropdownMenuItem(value: 'pong', child: Text('Pong')),
                          DropdownMenuItem(
                            value: 'close',
                            child: Text('Close'),
                          ),
                        ],
                        onChanged: (v) {
                          setState(() {
                            localOpcode = v ?? 'all';
                          });
                        },
                      ),
                      const SizedBox(width: 16),
                      const Text('Direction:'),
                      const SizedBox(width: 8),
                      DropdownButton<String>(
                        value: localDirection,
                        style: TextStyle(
                          color: Theme.of(context).colorScheme.onSurface,
                        ),
                        items: const [
                          DropdownMenuItem(value: 'all', child: Text('Any')),
                          DropdownMenuItem(
                            value: 'client->upstream',
                            child: Text('client->upstream'),
                          ),
                          DropdownMenuItem(
                            value: 'upstream->client',
                            child: Text('upstream->client'),
                          ),
                        ],
                        onChanged: (v) {
                          setState(() {
                            localDirection = v ?? 'all';
                          });
                        },
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  CheckboxListTile(
                    contentPadding: EdgeInsets.zero,
                    dense: true,
                    title: const Text('Hide heartbeats (ping/pong)'),
                    value: localHideHeartbeats,
                    onChanged: (v) {
                      setState(() {
                        localHideHeartbeats = v ?? false;
                      });
                    },
                  ),
                  const SizedBox(height: 8),
                  TextField(
                    controller: namespaceCtrl,
                    decoration: const InputDecoration(
                      labelText: 'Namespace contains (or ev=eventName)',
                    ),
                  ),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      TextButton(
                        onPressed: () {
                          setState(() {
                            localOpcode = 'all';
                            localDirection = 'all';
                            localHideHeartbeats = false;
                            namespaceCtrl.clear();
                          });
                          onChangeOpcode('all');
                          onChangeDirection('all');
                          onToggleHeartbeats(false);
                          Navigator.of(context).pop();
                        },
                        child: const Text('Reset'),
                      ),
                      const Spacer(),
                      ElevatedButton(
                        onPressed: () {
                          onChangeOpcode(localOpcode);
                          onChangeDirection(localDirection);
                          onToggleHeartbeats(localHideHeartbeats);
                          Navigator.of(context).pop();
                        },
                        child: const Text('Apply'),
                      ),
                    ],
                  ),
                ],
              ),
            );
          },
        );
      },
    );
  }

  void _openFirebaseFilters(BuildContext context) {
    // Собираем уникальные операции из текущих фреймов
    final ops = <String>{};
    for (final item in frames) {
      final frame = item as Map<String, dynamic>;
      final preview = (frame['preview'] ?? '').toString();
      final fb = FirebaseEventData.tryParse(preview);
      if (fb != null && fb.op.isNotEmpty) ops.add(fb.op);
    }
    final sortedOps = ops.toList()..sort();

    String localOp = fbOpFilter;
    String localStatus = fbStatusFilter;
    final pathCtrl = TextEditingController(text: fbPathFilter);

    showModalBottomSheet(
      context: context,
      builder: (_) {
        return StatefulBuilder(
          builder: (context, setState) {
            return Padding(
              padding: const EdgeInsets.all(12),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'RTDB filters',
                    style: Theme.of(context).textTheme.titleMedium,
                  ),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      const Text('Operation:'),
                      const SizedBox(width: 8),
                      DropdownButton<String>(
                        value: sortedOps.contains(localOp) || localOp == 'all'
                            ? localOp
                            : 'all',
                        style: TextStyle(
                          color: Theme.of(context).colorScheme.onSurface,
                        ),
                        items: [
                          const DropdownMenuItem(
                            value: 'all',
                            child: Text('Any'),
                          ),
                          ...sortedOps.map(
                            (op) =>
                                DropdownMenuItem(value: op, child: Text(op)),
                          ),
                        ],
                        onChanged: (v) {
                          setState(() {
                            localOp = v ?? 'all';
                          });
                        },
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  Row(
                    children: [
                      const Text('Status:'),
                      const SizedBox(width: 8),
                      DropdownButton<String>(
                        value: localStatus,
                        style: TextStyle(
                          color: Theme.of(context).colorScheme.onSurface,
                        ),
                        items: const [
                          DropdownMenuItem(value: 'all', child: Text('Any')),
                          DropdownMenuItem(value: 'ok', child: Text('OK')),
                          DropdownMenuItem(
                            value: 'error',
                            child: Text('Error'),
                          ),
                        ],
                        onChanged: (v) {
                          setState(() {
                            localStatus = v ?? 'all';
                          });
                        },
                      ),
                    ],
                  ),
                  const SizedBox(height: 8),
                  TextField(
                    controller: pathCtrl,
                    decoration: const InputDecoration(
                      labelText: 'Path contains',
                      hintText: '/users/alice',
                    ),
                  ),
                  const SizedBox(height: 12),
                  Row(
                    children: [
                      TextButton(
                        onPressed: () {
                          setState(() {
                            localOp = 'all';
                            localStatus = 'all';
                            pathCtrl.clear();
                          });
                          onChangeFbOp?.call('all');
                          onChangeFbStatus?.call('all');
                          onChangeFbPath?.call('');
                          Navigator.of(context).pop();
                        },
                        child: const Text('Reset'),
                      ),
                      const Spacer(),
                      ElevatedButton(
                        onPressed: () {
                          onChangeFbOp?.call(localOp);
                          onChangeFbStatus?.call(localStatus);
                          onChangeFbPath?.call(pathCtrl.text.trim());
                          Navigator.of(context).pop();
                        },
                        child: const Text('Apply'),
                      ),
                    ],
                  ),
                ],
              ),
            );
          },
        );
      },
    ).whenComplete(() => pathCtrl.dispose());
  }
}

bool _isJsonLocal(String s) {
  try {
    jsonDecode(s);
    return true;
  } catch (_) {
    return false;
  }
}

// Some frames contain protocol wrapper (socket.io) like '42/namespace,[...]' or '2'/'3'
// Try to safely extract JSON part if it exists
String? _extractJsonPayload(String preview) {
  final trimmed = preview.trim();
  if (trimmed.isEmpty) return null;
  // clean JSON
  if (trimmed.startsWith('{') || trimmed.startsWith('[')) {
    if (_isJsonLocal(trimmed)) return trimmed;
  }
  // socket.io payload: digits/code + optional namespace + comma + JSON array/object
  final idxBrace = trimmed.indexOf('[');
  final idxBraceObj = trimmed.indexOf('{');
  int idx = -1;
  if (idxBrace >= 0 && idxBraceObj >= 0) {
    idx = idxBrace < idxBraceObj ? idxBrace : idxBraceObj;
  } else {
    idx = idxBrace >= 0 ? idxBrace : idxBraceObj;
  }
  if (idx > 0) {
    final candidate = trimmed.substring(idx);
    if (_isJsonLocal(candidate)) return candidate;
  }
  return null;
}

class _WsFramesFullscreenDialog extends StatefulWidget {
  const _WsFramesFullscreenDialog({
    required this.frames,
    required this.events,
    required this.initialOpcodeFilter,
    required this.initialDirectionFilter,
    required this.initialNamespaceText,
    required this.initialHideHeartbeats,
    required this.isClosed,
    this.closedAt,
    this.error,
    this.initialFbOpFilter = 'all',
    this.initialFbStatusFilter = 'all',
    this.initialFbPathFilter = '',
  });
  final List<dynamic> frames;
  final List<dynamic> events;
  final String initialOpcodeFilter;
  final String initialDirectionFilter;
  final String initialNamespaceText;
  final bool initialHideHeartbeats;
  final bool isClosed;
  final DateTime? closedAt;
  final String? error;
  final String initialFbOpFilter;
  final String initialFbStatusFilter;
  final String initialFbPathFilter;

  @override
  State<_WsFramesFullscreenDialog> createState() =>
      _WsFramesFullscreenDialogState();
}

class _WsFramesFullscreenDialogState extends State<_WsFramesFullscreenDialog> {
  late String _opcodeFilter;
  late String _directionFilter;
  late bool _hideHeartbeats;
  late TextEditingController _namespaceCtrl;
  late String _fbOpFilter;
  late String _fbStatusFilter;
  late String _fbPathFilter;

  @override
  void initState() {
    super.initState();
    _opcodeFilter = widget.initialOpcodeFilter;
    _directionFilter = widget.initialDirectionFilter;
    _hideHeartbeats = widget.initialHideHeartbeats;
    _namespaceCtrl = TextEditingController(text: widget.initialNamespaceText);
    _fbOpFilter = widget.initialFbOpFilter;
    _fbStatusFilter = widget.initialFbStatusFilter;
    _fbPathFilter = widget.initialFbPathFilter;
  }

  @override
  void dispose() {
    _namespaceCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Dialog.fullscreen(
      child: Scaffold(
        appBar: AppBar(
          title: const Text('Frames'),
          leading: IconButton(
            icon: const Icon(Icons.close),
            onPressed: () => Navigator.of(context).pop(),
          ),
        ),
        body: WsDetailsPanel(
          frames: widget.frames,
          events: widget.events,
          opcodeFilter: _opcodeFilter,
          directionFilter: _directionFilter,
          namespaceCtrl: _namespaceCtrl,
          onChangeOpcode: (v) => setState(() => _opcodeFilter = v),
          onChangeDirection: (v) => setState(() => _directionFilter = v),
          hideHeartbeats: _hideHeartbeats,
          onToggleHeartbeats: (v) => setState(() => _hideHeartbeats = v),
          isClosed: widget.isClosed,
          closedAt: widget.closedAt,
          error: widget.error,
          onCloseFullscreen: () => Navigator.of(context).pop(),
          fbOpFilter: _fbOpFilter,
          fbStatusFilter: _fbStatusFilter,
          fbPathFilter: _fbPathFilter,
          onChangeFbOp: (v) => setState(() => _fbOpFilter = v),
          onChangeFbStatus: (v) => setState(() => _fbStatusFilter = v),
          onChangeFbPath: (v) => setState(() => _fbPathFilter = v),
        ),
      ),
    );
  }
}

class _WsClosedBanner extends StatelessWidget {
  const _WsClosedBanner({required this.closedAt, this.error});
  final DateTime? closedAt;
  final String? error;

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final hasError = error != null && error!.isNotEmpty;
    final timeStr = closedAt != null
        ? _fmtTime(closedAt!.toIso8601String())
        : null;

    final label = StringBuffer('Closed');
    if (timeStr != null) {
      label.write(' at $timeStr');
    }

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      margin: const EdgeInsets.fromLTRB(8, 4, 8, 0),
      decoration: BoxDecoration(
        color: cs.error.withValues(alpha: 0.08),
        borderRadius: BorderRadius.circular(6),
        border: Border(left: BorderSide(color: cs.error, width: 2)),
      ),
      child: Row(
        children: [
          Icon(
            hasError ? Icons.error_outline : Icons.info_outline,
            size: 14,
            color: cs.error,
          ),
          const SizedBox(width: 6),
          Expanded(
            child: Tooltip(
              message: hasError ? error! : '',
              child: Text(
                hasError ? '$label — $error' : '$label',
                style: TextStyle(fontSize: 11, color: cs.error),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),
          ),
        ],
      ),
    );
  }
}

class _JsonToggleRow extends StatefulWidget {
  const _JsonToggleRow({required this.json});
  final String json;
  @override
  State<_JsonToggleRow> createState() => _JsonToggleRowState();
}

class _JsonToggleRowState extends State<_JsonToggleRow> {
  bool pretty = true;
  bool tree = false;
  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Wrap(
          spacing: 8,
          children: [
            FilterChip(
              label: const Text('Pretty', style: TextStyle(fontSize: 12)),
              selected: pretty && !tree,
              onSelected: (v) {
                setState(() {
                  tree = false;
                  pretty = true;
                });
              },
            ),
            FilterChip(
              label: const Text('Tree', style: TextStyle(fontSize: 12)),
              selected: tree,
              onSelected: (v) {
                setState(() {
                  tree = v;
                  pretty = !v;
                });
              },
            ),
            TextButton.icon(
              style: TextButton.styleFrom(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
                textStyle: const TextStyle(fontSize: 12),
              ),
              onPressed: () {
                Clipboard.setData(ClipboardData(text: widget.json));
              },
              icon: const Icon(Icons.copy, size: 16),
              label: const Text('Copy'),
            ),
          ],
        ),
        const SizedBox(height: 6),
        // Content
        if (tree)
          JsonViewer(jsonString: widget.json, forceTree: true)
        else
          JsonViewer(jsonString: widget.json),
      ],
    );
  }
}
