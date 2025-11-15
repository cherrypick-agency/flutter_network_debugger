import 'package:flutter/material.dart';
import '../../../../theme/context_ext.dart';
import 'frames_timeline/frames_timeline.dart';
import 'frames_timeline/frames_timeline_legend.dart';
import '../../../../widgets/json_viewer.dart';
import '../../../../widgets/common_search_bar.dart';
import 'searchable_text_rich.dart';
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
  _PendingFocus(this.frameId, this.localIndex);
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

  @override
  State<WsDetailsPanel> createState() => _WsDetailsPanelState();
}

class _WsDetailsPanelState extends State<WsDetailsPanel> {
  bool _pretty = true;
  bool _tree = false;
  bool _showTimeline = true;
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
  //   // небольшой порог, чтобы не дёргать прокрутку при едва заметном «хвосте»
  //   const threshold = 48.0;
  //   _stickToBottom = (pos.maxScrollExtent - pos.pixels) <= threshold;
  // }

  // Уникальный ключ фрейма, чтобы не путать элементы с одинаковыми id в разных направлениях
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

  // Keys for stable widget identity (not cached - created fresh each build)
  ValueKey<String> _tileKeyFor(String id) => ValueKey<String>('tile_$id');

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
    _frameMatchCounts.removeWhere((k, _) => !current.contains(k));
    _frameMatchKeys.removeWhere((k, _) => !current.contains(k));
    _frameLocalFocusedIndex.removeWhere((k, _) => !current.contains(k));
    if (_pendingFocus != null && !current.contains(_pendingFocus!.frameId)) {
      _pendingFocus = null;
    }
  }

  void _localGotoNext(String id) {
    final s = _localSearch[id];
    if (s == null || s.keys.isEmpty) return;
    setState(() {
      s.focusedIndex = (s.focusedIndex + 1) % s.keys.length;
    });
    final ctx = s.keys[s.focusedIndex].currentContext;
    if (ctx != null) {
      Scrollable.ensureVisible(
        ctx,
        duration: const Duration(milliseconds: 220),
        curve: Curves.easeOutCubic,
        alignment: 0.1,
      );
    }
  }

  void _localGotoPrev(String id) {
    final s = _localSearch[id];
    if (s == null || s.keys.isEmpty) return;
    setState(() {
      s.focusedIndex = (s.focusedIndex - 1) < 0
          ? s.keys.length - 1
          : s.focusedIndex - 1;
    });
    final ctx = s.keys[s.focusedIndex].currentContext;
    if (ctx != null) {
      Scrollable.ensureVisible(
        ctx,
        duration: const Duration(milliseconds: 220),
        curve: Curves.easeOutCubic,
        alignment: 0.1,
      );
    }
  }

  bool _frameMatches(Map<String, dynamic> f) {
    if (widget.opcodeFilter != 'all' &&
        (f['opcode']?.toString() ?? '') != widget.opcodeFilter)
      return false;
    if (widget.directionFilter != 'all' &&
        (f['direction']?.toString() ?? '') != widget.directionFilter)
      return false;
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
        oldWidget.directionFilter != widget.directionFilter) {
      _updateVisibleFramesCache();
      // Reindex search when visible frames change
      if (_showGlobalSearch && _searchCtrl.text.trim().isNotEmpty) {
        _reindexGlobalMatches();
      }
    }
  }

  void _scrollToFrame(String frameId) {
    // Use cached visible frames for consistency with displayed list
    final visibleFrames = _cachedVisibleFrames;

    final idx = visibleFrames.indexWhere((f) {
      final compositeKey = _frameKeyOf(f);
      return compositeKey == frameId || (f['id'] ?? '').toString() == frameId;
    });

    if (idx < 0 || !_listCtrl.hasClients) return;

    final estimatedItemExtent = 80.0; // Average height including heartbeats
    final target = (idx * estimatedItemExtent).toDouble();
    _listCtrl.animateTo(
      target.clamp(0, _listCtrl.position.maxScrollExtent),
      duration: const Duration(milliseconds: 220),
      curve: Curves.easeOutCubic,
    );
  }

  // (заголовок рендерим через _buildTitleRich)

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
          // Интересуют только объекты/массивы — примитивы как строки не трогаем
          if (parsed is Map || parsed is List) {
            return _decodeNestedJsonStrings(parsed);
          }
        } catch (_) {}
      }
    }
    return val;
  }

  // Пытаемся получить нормализованный объект для заголовка (если preview — JSON)
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

  // Однострочный рендер JSON с подсветкой токенов для заголовка
  Widget _buildTitleRich(BuildContext context, String preview) {
    // 1) Чистый JSON целиком
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
    // 2) JSON как часть строки (например, socket.io '42/ns,[...]')
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
        // если вдруг не распарсили — упадём на дефолтный рендер
      }
    }
    // 3) Не JSON — просто текст
    return Text(
      preview,
      maxLines: 1,
      overflow: TextOverflow.ellipsis,
      style: context.appText.body,
    );
  }

  List<InlineSpan> _buildInlineJsonSpans(BuildContext context, dynamic node) {
    final base = context.appText.body;
    final punct = base;
    final keyStyle = base.copyWith(color: context.appColors.primary);
    final stringStyle = base.copyWith(color: context.appColors.success);
    final numberStyle = base.copyWith(color: context.appColors.warning);
    final boolStyle = base.copyWith(color: context.appColors.warning);
    final nullStyle = base.copyWith(color: context.appColors.danger);

    List<InlineSpan> build(dynamic n) {
      final List<InlineSpan> out = [];
      void add(String s, TextStyle st) => out.add(TextSpan(text: s, style: st));

      if (n is Map) {
        add('{', punct);
        int i = 0;
        final last = n.length - 1;
        for (final e in n.entries) {
          add('"${e.key}"', keyStyle);
          add(': ', punct);
          out.addAll(build(e.value));
          if (i != last) add(', ', punct);
          i++;
        }
        add('}', punct);
        return out;
      }
      if (n is List) {
        add('[', punct);
        for (int i = 0; i < n.length; i++) {
          out.addAll(build(n[i]));
          if (i != n.length - 1) add(', ', punct);
        }
        add(']', punct);
        return out;
      }
      if (n is String) {
        add('"$n"', stringStyle);
        return out;
      }
      if (n is num) {
        add(n.toString(), numberStyle);
        return out;
      }
      if (n is bool) {
        add(n ? 'true' : 'false', boolStyle);
        return out;
      }
      if (n == null) {
        add('null', nullStyle);
        return out;
      }
      add(n.toString(), punct);
      return out;
    }

    return build(node);
  }

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
    // Clear focused index cache before resolving (handles null case)
    _frameLocalFocusedIndex.clear();
    final (fid, local) = _resolveGlobalIndexToFrame(gIndex);
    if (fid == null) return;
    // Update local focused index cache for highlighting
    _frameLocalFocusedIndex[fid] = local;
    // BUG 2 fix: Expand tile and wait for it to render before scrolling
    _tileControllerFor(fid).expand();
    // Wait for tile expansion to complete and widgets to build
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      // Scroll to tile after expansion
      _scrollToFrame(fid);
      // Defer focus to internal match until keys are ready
      _pendingFocus = _PendingFocus(fid, local);
    });
  }

  void _gotoNext() {
    if (_globalTotalMatches <= 0) return;
    setState(() {
      _globalFocusedIndex = (_globalFocusedIndex + 1) % _globalTotalMatches;
    });
    _focusGlobal(_globalFocusedIndex);
  }

  void _gotoPrev() {
    if (_globalTotalMatches <= 0) return;
    setState(() {
      _globalFocusedIndex = (_globalFocusedIndex - 1) < 0
          ? _globalTotalMatches - 1
          : _globalFocusedIndex - 1;
    });
    _focusGlobal(_globalFocusedIndex);
  }

  @override
  void dispose() {
    _searchCtrl.dispose();
    _searchFocus.dispose();
    _listCtrl.dispose();
    for (final s in _localSearch.values) {
      s.dispose();
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
    if (pending != null && pending.frameId == frameId) {
      final local = pending.localIndex;
      if (local >= 0 && local < keys.length) {
        final ctx = keys[local].currentContext;
        if (ctx != null) {
          WidgetsBinding.instance.addPostFrameCallback((_) {
            if (!mounted) return;
            try {
              final scrollable = Scrollable.maybeOf(ctx);
              if (scrollable == null) {
                WidgetsBinding.instance.addPostFrameCallback((_) {
                  if (!mounted) return;
                  final ctx2 = keys[local].currentContext;
                  if (ctx2 != null && Scrollable.maybeOf(ctx2) != null) {
                    Scrollable.ensureVisible(
                      ctx2,
                      duration: const Duration(milliseconds: 220),
                      curve: Curves.easeOutCubic,
                      alignment: 0.1,
                    );
                  }
                });
              } else {
                Scrollable.ensureVisible(
                  ctx,
                  duration: const Duration(milliseconds: 220),
                  curve: Curves.easeOutCubic,
                  alignment: 0.1,
                );
              }
            } catch (_) {}
          });
        }
      }
      _pendingFocus = null;
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
                  // hover: только подсветка на таймлайне, без скролла списка
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
        // Timeline and search above header
        timelineSection,
        Expanded(
          child: _Card(
            title: 'Frames',
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
                      final extractedJson = _extractJsonPayload(preview);
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
                        leading: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            icon,
                            const SizedBox(width: 6),
                            Text(opcode),
                          ],
                        ),
                        title: _buildTitleRich(context, preview),
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
                        trailing: Text(
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
                                        setState(() {
                                          local.keys = keys;
                                          if (local.focusedIndex >=
                                              keys.length) {
                                            local.focusedIndex = 0;
                                          }
                                        });
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
                                      if (_tree) {
                                        return JsonTreeRich(
                                          data: normalized,
                                          search: cfg,
                                        );
                                      }
                                      if (_pretty) {
                                        return JsonPrettyRich(
                                          data: normalized,
                                          search: cfg,
                                        );
                                      }
                                    } catch (_) {}
                                  }
                                  return SearchableTextRich(
                                    text: preview,
                                    search: cfg,
                                    style: context.appText.monospace,
                                  );
                                },
                              );

                              // BUG 1 fix: Wrap content in scrollable container
                              // so Stack doesn't scroll, only content inside scrolls
                              final content = ConstrainedBox(
                                constraints: const BoxConstraints(
                                  maxHeight: 400, // Max height before scrolling
                                ),
                                child: SingleChildScrollView(
                                  child: Container(
                                    alignment: Alignment.centerLeft,
                                    padding: const EdgeInsets.all(8),
                                    child: contentWidget,
                                  ),
                                ),
                              );

                              return Stack(
                                children: [
                                  content,
                                  Positioned(
                                    top: 6,
                                    left: 6,
                                    right: 6,
                                    child: Align(
                                      alignment: Alignment.topRight,
                                      child: Row(
                                        mainAxisSize: MainAxisSize.min,
                                        children: [
                                          // Кнопка копирования содержимого фрейма
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
                                          // Поиск по содержимому фрейма
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
                                                    .addPostFrameCallback((_) {
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
  final String title;
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
                  child: Text(
                    title,
                    style: Theme.of(context).textTheme.titleMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
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
  void _openFilters(BuildContext context) {
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
