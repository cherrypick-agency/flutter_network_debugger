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
  // Last known frames list length (for auto-scroll)
  int _lastFramesLen = 0;
  // Flag that after next frame we should scroll down
  bool _autoScrollPending = false;
  DateTimeRange? _brushRange;
  String? _expandedId;

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
  String? _pendingFocusFrameId;
  int _pendingFocusLocalIndex = 0;

  // Keys of tile containers for precise ensureVisible in external list
  final Map<String, GlobalKey> _frameTileKeys = <String, GlobalKey>{};
  GlobalKey _tileKeyFor(String id) =>
      _frameTileKeys.putIfAbsent(id, () => GlobalKey());

  // Local search at frame level
  final Map<String, _LocalSearchState> _localSearch =
      <String, _LocalSearchState>{};
  _LocalSearchState _localFor(String id) =>
      _localSearch.putIfAbsent(id, () => _LocalSearchState());
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
        duration: const Duration(milliseconds: 200),
        alignment: 0.1,
      );
    }
  }

  void _localGotoPrev(String id) {
    final s = _localSearch[id];
    if (s == null || s.keys.isEmpty) return;
    setState(() {
      s.focusedIndex =
          (s.focusedIndex - 1) < 0 ? s.keys.length - 1 : s.focusedIndex - 1;
    });
    final ctx = s.keys[s.focusedIndex].currentContext;
    if (ctx != null) {
      Scrollable.ensureVisible(
        ctx,
        duration: const Duration(milliseconds: 200),
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

  // If new frames appeared — scroll down if necessary,
  // but only if user was already at the end of the list
  void _maybeAutoScrollToBottomOnNewFrames() {
    if (widget.frames.length > _lastFramesLen) {
      bool atBottom = false;
      if (_listCtrl.hasClients) {
        try {
          final pos = _listCtrl.position;
          // small tolerance to not jerk if almost at bottom
          atBottom = pos.pixels >= (pos.maxScrollExtent - 16);
        } catch (_) {}
      } else {
        // if scroll not attached yet — consider user not at bottom
        atBottom = false;
      }
      _autoScrollPending = atBottom;
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (!mounted) return;
        if (_autoScrollPending && _listCtrl.hasClients) {
          final max = _listCtrl.position.maxScrollExtent;
          if (max > 0) {
            _listCtrl.animateTo(
              max,
              duration: const Duration(milliseconds: 120),
              curve: Curves.easeOut,
            );
          }
        }
        _lastFramesLen = widget.frames.length;
        _autoScrollPending = false;
      });
    } else {
      _lastFramesLen = widget.frames.length;
    }
  }

  @override
  void didUpdateWidget(covariant WsDetailsPanel oldWidget) {
    super.didUpdateWidget(oldWidget);
    _maybeAutoScrollToBottomOnNewFrames();
  }

  void _scrollToFrame(String frameId) {
    // 1) Rough scroll by index to build widget in screen area
    final idx = widget.frames.indexWhere(
      (e) => (e as Map)['id']?.toString() == frameId,
    );
    if (idx < 0 || !_listCtrl.hasClients) return;
    final estimatedItemExtent = 56.0;
    final target = (idx * estimatedItemExtent).toDouble();
    _listCtrl
        .animateTo(
          target.clamp(0, _listCtrl.position.maxScrollExtent),
          duration: const Duration(milliseconds: 200),
          curve: Curves.easeOut,
        )
        .whenComplete(() {
          if (!mounted) return;
          // 2) Precise adjustment: ensureVisible by tile container key
          WidgetsBinding.instance.addPostFrameCallback((_) {
            final key = _tileKeyFor(frameId);
            final ctx = key.currentContext;
            if (ctx != null) {
              try {
                Scrollable.ensureVisible(
                  ctx,
                  duration: const Duration(milliseconds: 180),
                  alignment: 0.1,
                );
              } catch (_) {}
            }
          });
        });
  }

  void _reindexGlobalMatches() {
    final query = _searchCtrl.text.trim();
    _frameMatchCounts.clear();
    _globalTotalMatches = 0;
    if (query.isEmpty) {
      setState(() {
        _globalFocusedIndex = 0;
      });
      return;
    }
    for (final f in widget.frames) {
      final fm = f as Map<String, dynamic>;
      if (!_frameMatches(fm)) continue;
      if (widget.hideHeartbeats && _isHeartbeat(fm)) continue;
      if (_brushRange != null) {
        try {
          final ts = DateTime.parse((fm['ts'] ?? '').toString());
          if (ts.isBefore(_brushRange!.start) || ts.isAfter(_brushRange!.end))
            continue;
        } catch (_) {}
      }
      final preview = (fm['preview'] ?? '').toString();
      final extractedJson = _extractJsonPayload(preview);
      int cnt = 0;
      if (extractedJson != null && (_pretty || _tree)) {
        cnt = _countMatchesIn(extractedJson);
      } else {
        cnt = _countMatchesIn(preview);
      }
      if (cnt > 0) {
        final idStr = (fm['id'] ?? '').toString();
        _frameMatchCounts[idStr] = cnt;
        _globalTotalMatches += cnt;
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
  }

  int _countMatchesIn(String text) {
    final q = _searchCtrl.text.trim();
    if (q.isEmpty) return 0;
    if (_useRegex) {
      RegExp? re;
      try {
        re = RegExp(q, caseSensitive: _matchCase);
      } catch (_) {
        return 0;
      }
      int c = 0;
      for (final m in re.allMatches(text)) {
        if (_wholeWord) {
          bool isWordChar(String ch) {
            final code = ch.codeUnitAt(0);
            final isAZ =
                (code >= 65 && code <= 90) || (code >= 97 && code <= 122);
            final is09 = (code >= 48 && code <= 57);
            return isAZ || is09 || ch == '_';
          }

          final s = m.start;
          final e = m.end;
          final left = s - 1 >= 0 ? text.substring(s - 1, s) : null;
          final right = e < text.length ? text.substring(e, e + 1) : null;
          final leftOk = left == null || !isWordChar(left);
          final rightOk = right == null || !isWordChar(right);
          if (!(leftOk && rightOk)) {
            continue;
          }
        }
        c++;
      }
      return c;
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
        final right =
            (idx + query.length) < src.length
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
    for (final f in widget.frames) {
      final idStr = ((f as Map)['id'] ?? '').toString();
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
    final (fid, local) = _resolveGlobalIndexToFrame(gIndex);
    if (fid == null) return;
    setState(() {
      _expandedId = fid;
    });
    // Scroll to tile (rough + precise adjustment)
    _scrollToFrame(fid);
    // Always defer focus to internal match until keys are ready
    _pendingFocusFrameId = fid;
    _pendingFocusLocalIndex = local;
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
      _globalFocusedIndex =
          (_globalFocusedIndex - 1) < 0
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
    super.dispose();
  }

  void _onChildMatches(String frameId, int count, List<GlobalKey> keys) {
    if (!_showGlobalSearch)
      return; // local search should not overwrite global keys
    _frameMatchKeys[frameId] = keys;
    // if waiting for focus on this specific frame — try to navigate
    if (_pendingFocusFrameId == frameId) {
      final local = _pendingFocusLocalIndex;
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
                      duration: const Duration(milliseconds: 200),
                      alignment: 0.1,
                    );
                  }
                });
              } else {
                Scrollable.ensureVisible(
                  ctx,
                  duration: const Duration(milliseconds: 200),
                  alignment: 0.1,
                );
              }
            } catch (_) {}
          });
        }
      }
      _pendingFocusFrameId = null;
    }
  }

  @override
  Widget build(BuildContext context) {
    final timelineSection = Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4),
      child:
          (_showTimeline)
              ? Column(
                children: [
                  Builder(
                    builder: (context) {
                      final framesList =
                          widget.frames.cast<Map<String, dynamic>>();
                      final tlFrames =
                          framesList.where((f) {
                            if (!_frameMatches(f)) return false;
                            if (widget.hideHeartbeats && _isHeartbeat(f))
                              return false;
                            return true;
                          }).toList();
                      return FramesTimeline(
                        frames: tlFrames,
                        height: 50,
                        onFrameTap: _scrollToFrame,
                        onBrushChanged: (r) {
                          setState(() {
                            _brushRange = r;
                          });
                        },
                        onFrameHover: (id) {
                          final idx = widget.frames.indexWhere(
                            (e) => (e as Map)['id']?.toString() == id,
                          );
                          if (idx >= 0 && _listCtrl.hasClients) {
                            final estimatedItemExtent = 56.0;
                            final target =
                                (idx * estimatedItemExtent).toDouble();
                            _listCtrl.jumpTo(
                              target.clamp(
                                0,
                                _listCtrl.position.maxScrollExtent,
                              ),
                            );
                          }
                        },
                      );
                    },
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
                    itemCount: widget.frames.length,
                    itemBuilder: (_, i) {
                      final f = widget.frames[i] as Map<String, dynamic>;
                      if (!_frameMatches(f)) {
                        return const SizedBox.shrink();
                      }
                      if (_brushRange != null) {
                        try {
                          final ts = DateTime.parse((f['ts'] ?? '').toString());
                          if (ts.isBefore(_brushRange!.start) ||
                              ts.isAfter(_brushRange!.end)) {
                            return const SizedBox.shrink();
                          }
                        } catch (_) {}
                      }
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
                        color:
                            isDown
                                ? context.appColors.success
                                : context.appColors.primary,
                      );
                      if (widget.hideHeartbeats && isHeartbeat) {
                        return const SizedBox.shrink();
                      }
                      if (isHeartbeat) {
                        final label =
                            isWsPingPong
                                ? opcode.toUpperCase()
                                : (preview == '2' ? 'PING' : 'PONG');
                        return ListTile(
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
                            style: Theme.of(
                              context,
                            ).textTheme.labelSmall?.copyWith(
                              color: context.appColors.textSecondary,
                            ),
                          ),
                        );
                      }
                      final idStr = (f['id'] ?? '').toString();
                      final isExpanded = idStr == _expandedId;
                      // Local focusedIndex for this frame
                      int localFocusedIndex = -1;
                      if (_globalTotalMatches > 0 &&
                          _frameMatchCounts.isNotEmpty) {
                        int acc = 0;
                        for (final ff in widget.frames) {
                          final fid = ((ff as Map)['id'] ?? '').toString();
                          final cnt = _frameMatchCounts[fid] ?? 0;
                          if (cnt <= 0) continue;
                          final end = acc + cnt - 1;
                          if (fid == idStr &&
                              _globalFocusedIndex >= acc &&
                              _globalFocusedIndex <= end) {
                            localFocusedIndex = _globalFocusedIndex - acc;
                            break;
                          }
                          acc += cnt;
                        }
                      }
                      return ExpansionTile(
                        key: ValueKey(
                          'frame_${idStr}_${isExpanded ? 'open' : 'closed'}',
                        ),
                        initiallyExpanded: isExpanded,
                        tilePadding: const EdgeInsets.symmetric(horizontal: 8),
                        dense: true,
                        leading: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            icon,
                            const SizedBox(width: 6),
                            Text(f['opcode'].toString()),
                          ],
                        ),
                        title: Text(
                          preview,
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                          style: context.appText.body,
                        ),
                        subtitle: Row(
                          children: [
                            Expanded(
                              child: Text(
                                '${f['size']} B',
                                style: Theme.of(
                                  context,
                                ).textTheme.labelSmall?.copyWith(
                                  color: context.appColors.textSecondary,
                                ),
                              ),
                            ),
                          ],
                        ),
                        trailing: Text(
                          ts,
                          style: Theme.of(
                            context,
                          ).textTheme.labelSmall?.copyWith(
                            color: context.appColors.textSecondary,
                          ),
                        ),
                        onExpansionChanged: (open) {
                          setState(() {
                            _expandedId =
                                open
                                    ? idStr
                                    : (_expandedId == idStr
                                        ? null
                                        : _expandedId);
                          });
                        },
                        children: [
                          Builder(
                            builder: (context) {
                              final local = _localFor(idStr);
                              final String activeQuery =
                                  local.show
                                      ? local.controller.text.trim()
                                      : (_showGlobalSearch
                                          ? _searchCtrl.text.trim()
                                          : '');
                              final bool activeMatchCase =
                                  local.show ? local.matchCase : _matchCase;
                              final bool activeWholeWord =
                                  local.show ? local.wholeWord : _wholeWord;
                              final bool activeUseRegex =
                                  local.show ? local.useRegex : _useRegex;
                              final int activeFocusedIndex =
                                  local.show
                                      ? local.focusedIndex
                                      : (localFocusedIndex < 0
                                          ? 0
                                          : localFocusedIndex);

                              final content = Container(
                                alignment: Alignment.centerLeft,
                                padding: const EdgeInsets.all(8),
                                child: Builder(
                                  builder: (context) {
                                    final cfg = JsonSearchConfig(
                                      query: activeQuery,
                                      matchCase: activeMatchCase,
                                      wholeWord: activeWholeWord,
                                      useRegex: activeUseRegex,
                                      focusedIndex: activeFocusedIndex,
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
                                          _onChildMatches(idStr, count, keys);
                                          if ((_frameMatchCounts[idStr] ??
                                                  -1) !=
                                              count) {
                                            _frameMatchCounts[idStr] = count;
                                            _reindexGlobalMatches();
                                          }
                                        }
                                      },
                                    );
                                    if (extractedJson != null) {
                                      if (_tree) {
                                        return JsonTreeRich(
                                          data: jsonDecode(extractedJson),
                                          search: cfg,
                                        );
                                      }
                                      if (_pretty) {
                                        return JsonPrettyRich(
                                          data: jsonDecode(extractedJson),
                                          search: cfg,
                                        );
                                      }
                                    }
                                    return SearchableTextRich(
                                      text: preview,
                                      search: cfg,
                                      style: context.appText.monospace,
                                    );
                                  },
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
                                      child:
                                          local.show
                                              ? CommonSearchBar(
                                                controller: local.controller,
                                                focusNode: local.focus,
                                                countText:
                                                    local.keys.isEmpty
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
                                                onNext:
                                                    () => _localGotoNext(idStr),
                                                onPrev:
                                                    () => _localGotoPrev(idStr),
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
                                              : IconButton(
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
    final countText =
        _globalTotalMatches == 0
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
          _globalFocusedIndex = 0;
          _reindexGlobalMatches();
          setState(() {});
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
                  Align(
                    alignment: Alignment.centerRight,
                    child: ElevatedButton(
                      onPressed: () {
                        onChangeOpcode(localOpcode);
                        onChangeDirection(localDirection);
                        onToggleHeartbeats(localHideHeartbeats);
                        Navigator.of(context).pop();
                      },
                      child: const Text('Apply'),
                    ),
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
