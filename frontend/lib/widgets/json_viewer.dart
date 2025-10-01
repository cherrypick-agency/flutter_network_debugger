import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter/gestures.dart';
import 'package:flutter/services.dart';
import '../core/hotkeys/hotkeys_service.dart';
import 'common_search_bar.dart';
import '../core/di/di.dart';
// import 'package:json_tree_viewer/json_tree_viewer.dart' as jtv; // no longer used directly
import '../theme/context_ext.dart';

/// Универсальный JSON-виджет поверх json_tree_viewer
class JsonViewer extends StatefulWidget {
  const JsonViewer({
    super.key,
    required this.jsonString,
    this.forceTree = false,
    this.treeHeight = 280,
  });
  final String jsonString;
  final bool forceTree;
  final double treeHeight;

  @override
  State<JsonViewer> createState() => _JsonViewerState();
}

class _JsonViewerState extends State<JsonViewer> {
  bool _showSearch = false;
  final TextEditingController _searchCtrl = TextEditingController();
  final FocusNode _searchFocusNode = FocusNode();
  bool _matchCase = false;
  bool _wholeWord = false;
  bool _useRegex = false;
  int _focusedIndex = 0;
  List<GlobalKey> _matchKeys = const [];

  @override
  void dispose() {
    _searchCtrl.dispose();
    _searchFocusNode.dispose();
    super.dispose();
  }

  void _handleMatchesRebuilt(int count, List<GlobalKey> keys) {
    setState(() {
      _matchKeys = keys;
      if (_matchKeys.isEmpty) {
        _focusedIndex = 0;
      } else if (_focusedIndex >= _matchKeys.length) {
        _focusedIndex = 0;
      }
    });
  }

  void _gotoNext() {
    if (_matchKeys.isEmpty) return;
    setState(() {
      _focusedIndex = (_focusedIndex + 1) % _matchKeys.length;
    });
    _scrollToFocused();
  }

  void _gotoPrev() {
    if (_matchKeys.isEmpty) return;
    setState(() {
      _focusedIndex =
          (_focusedIndex - 1) < 0 ? _matchKeys.length - 1 : _focusedIndex - 1;
    });
    _scrollToFocused();
  }

  void _scrollToFocused() {
    if (_matchKeys.isEmpty) return;
    final key = _matchKeys[_focusedIndex];
    final ctx = key.currentContext;
    if (ctx != null) {
      Scrollable.ensureVisible(
        ctx,
        duration: const Duration(milliseconds: 200),
        alignment: 0.1,
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    dynamic data;
    try {
      data = jsonDecode(widget.jsonString);
    } catch (_) {
      return SelectableText(
        widget.jsonString,
        style: context.appText.monospace,
      );
    }
    return LayoutBuilder(
      builder: (context, constraints) {
        if (widget.forceTree) {
          final searchCfgTree = JsonSearchConfig(
            query: _searchCtrl.text.trim(),
            matchCase: _matchCase,
            wholeWord: _wholeWord,
            useRegex: _useRegex,
            focusedIndex: _focusedIndex,
            onRebuilt: _handleMatchesRebuilt,
          );
          final content = JsonTreeRich(data: data, search: searchCfgTree);
          if (!constraints.hasBoundedHeight) {
            return SizedBox(
              height: widget.treeHeight,
              child: Stack(
                children: [
                  Positioned.fill(child: SingleChildScrollView(child: content)),
                  Positioned(
                    top: 6,
                    left: 6,
                    right: 6,
                    child: Align(
                      alignment: Alignment.topRight,
                      child:
                          _showSearch
                              ? _buildSearchBar(context)
                              : _buildSearchButton(context),
                    ),
                  ),
                ],
              ),
            );
          }
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
                      _showSearch
                          ? _buildSearchBar(context)
                          : _buildSearchButton(context),
                ),
              ),
            ],
          );
        }

        final searchCfg = JsonSearchConfig(
          query: _showSearch ? _searchCtrl.text.trim() : '',
          matchCase: _matchCase,
          wholeWord: _wholeWord,
          useRegex: _useRegex,
          focusedIndex: _focusedIndex,
          onRebuilt: _handleMatchesRebuilt,
        );

        if (!constraints.hasBoundedHeight) {
          // Внутренний скролл и закрепленная сверху панель поиска
          return SizedBox(
            height: widget.treeHeight,
            child: Stack(
              children: [
                Positioned.fill(
                  child: SingleChildScrollView(
                    child: JsonPrettyRich(data: data, search: searchCfg),
                  ),
                ),
                Positioned(
                  top: 6,
                  left: 6,
                  right: 6,
                  child: Align(
                    alignment: Alignment.topRight,
                    child:
                        _showSearch
                            ? _buildSearchBar(context)
                            : _buildSearchButton(context),
                  ),
                ),
              ],
            ),
          );
        }
        // Обычный режим с закрепленной панелью по центру сверху
        return Stack(
          children: [
            JsonPrettyRich(data: data, search: searchCfg),
            Positioned(
              top: 6,
              left: 0,
              right: 0,
              child: Center(
                child:
                    _showSearch
                        ? _buildSearchBar(context)
                        : _buildSearchButton(context),
              ),
            ),
          ],
        );
      },
    );
  }

  Widget _buildSearchButton(BuildContext context) {
    return Material(
      color: Colors.transparent,
      child: IconButton(
        iconSize: 18,
        padding: const EdgeInsets.all(4),
        constraints: const BoxConstraints(minWidth: 28, minHeight: 28),
        tooltip: 'Search',
        onPressed: () {
          setState(() => _showSearch = true);
          WidgetsBinding.instance.addPostFrameCallback((_) {
            _searchFocusNode.requestFocus();
          });
        },
        icon: const Icon(Icons.search, size: 18),
      ),
    );
  }

  Widget _buildSearchBar(BuildContext context) {
    final countText =
        _matchKeys.isEmpty
            ? '0/0'
            : '${_focusedIndex + 1}/${_matchKeys.length}';
    final hk = sl<HotkeysService>();
    void _closeSearch() {
      setState(() {
        _showSearch = false;
        _searchCtrl.clear();
        _focusedIndex = 0;
        _matchKeys = const [];
        _searchFocusNode.unfocus();
      });
    }

    final handlers = hk.buildHandlers({
      'jsonSearch.next': _gotoNext,
      'jsonSearch.prev': _gotoPrev,
      'jsonSearch.close': _closeSearch,
    });
    return CallbackShortcuts(
      bindings: handlers,
      child: CommonSearchBar(
        controller: _searchCtrl,
        focusNode: _searchFocusNode,
        countText: countText,
        matchCase: _matchCase,
        wholeWord: _wholeWord,
        useRegex: _useRegex,
        canNavigate: _matchKeys.isNotEmpty,
        onChanged:
            () => setState(() {
              _focusedIndex = 0;
            }),
        onNext: _gotoNext,
        onPrev: _gotoPrev,
        onClose: _closeSearch,
        onToggleMatchCase: () => setState(() => _matchCase = !_matchCase),
        onToggleWholeWord: () => setState(() => _wholeWord = !_wholeWord),
        onToggleRegex: () => setState(() => _useRegex = !_useRegex),
      ),
    );
  }
}

// Kept for reference; pretty/tree rich modes cover tree rendering now

/// Расширенный tree-view с подсветкой и автоматическим раскрытием совпадений
class JsonTreeRich extends StatefulWidget {
  const JsonTreeRich({required this.data, required this.search});
  final dynamic data;
  final JsonSearchConfig search;

  @override
  State<JsonTreeRich> createState() => _JsonTreeRichState();
}

class _JsonTreeRichState extends State<JsonTreeRich> {
  final Set<String> _userExpanded = <String>{};

  bool _containsQuery(String text) {
    final cfg = widget.search;
    if (cfg.query.isEmpty) return false;
    final src = cfg.matchCase ? text : text.toLowerCase();
    final q = cfg.matchCase ? cfg.query : cfg.query.toLowerCase();
    int idx = src.indexOf(q);
    if (idx < 0) return false;
    if (!cfg.wholeWord) return true;
    bool isWordChar(String ch) {
      final c = ch.codeUnitAt(0);
      final isAZ = (c >= 65 && c <= 90) || (c >= 97 && c <= 122);
      final is09 = (c >= 48 && c <= 57);
      return isAZ || is09 || ch == '_';
    }

    while (idx >= 0) {
      final left = idx - 1 >= 0 ? src.substring(idx - 1, idx) : null;
      final right =
          (idx + q.length) < src.length
              ? src.substring(idx + q.length, idx + q.length + 1)
              : null;
      final leftOk = left == null || !isWordChar(left);
      final rightOk = right == null || !isWordChar(right);
      if (leftOk && rightOk) return true;
      idx = src.indexOf(q, idx + 1);
    }
    return false;
  }

  void _collectAutoExpandPaths(dynamic node, String path, Set<String> out) {
    bool matchedHere = false;
    if (node is Map) {
      for (final e in node.entries) {
        final keyStr = e.key.toString();
        if (_containsQuery(keyStr)) {
          matchedHere = true;
        }
        final childPath = '$path.${e.key}';
        _collectAutoExpandPaths(e.value, childPath, out);
      }
    } else if (node is List) {
      for (int i = 0; i < node.length; i++) {
        final childPath = '$path[$i]';
        _collectAutoExpandPaths(node[i], childPath, out);
      }
    } else if (node is String) {
      if (_containsQuery(node)) matchedHere = true;
    } else if (node is num || node is bool) {
      if (_containsQuery(node.toString())) matchedHere = true;
    }
    if (matchedHere && path.isNotEmpty) {
      // expand all ancestors of path (split by . and [i])
      String p = path;
      while (p.isNotEmpty) {
        out.add(p);
        // remove last segment
        final idxDot = p.lastIndexOf('.');
        final idxBr = p.lastIndexOf('[');
        final cut = idxDot > idxBr ? idxDot : idxBr;
        if (cut <= 0) {
          p = '';
        } else {
          p = p.substring(0, cut);
        }
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final base = context.appText.monospace;
    final punct = base;
    final keyStyle = base.copyWith(color: context.appColors.primary);
    final stringStyle = base.copyWith(color: context.appColors.success);
    final numberStyle = base.copyWith(color: context.appColors.warning);
    final boolStyle = base.copyWith(color: context.appColors.warning);
    final nullStyle = base.copyWith(color: context.appColors.danger);
    final Color hl = context.appColors.warning.withValues(alpha: 0.35);
    final Color hlFocus = context.appColors.warning.withValues(alpha: 0.55);

    int matchCounter = 0;
    final List<GlobalKey> keys = [];

    final Set<String> autoExpand = <String>{};
    _collectAutoExpandPaths(
      (widget.data is Map || widget.data is Iterable)
          ? widget.data
          : {'value': widget.data},
      '',
      autoExpand,
    );

    bool isExpanded(String path) =>
        _userExpanded.contains(path) || autoExpand.contains(path);

    const double indentPx = 14;
    const double iconSize = 14;
    const double iconPad = 2;

    List<InlineSpan> _valueToSpans(
      dynamic node,
      TextStyle stringStyle,
      TextStyle numberStyle,
      TextStyle boolStyle,
      TextStyle nullStyle,
      Color highlight,
      Color highlightFocused,
      int Function() matchCounter,
      void Function() incMatchCounter,
      List<GlobalKey> keys,
    ) {
      if (node is String) {
        return _splitWithHighlights(
          '"$node"',
          baseStyle: stringStyle,
          highlight: highlight,
          highlightFocused: highlightFocused,
          matchCounter: matchCounter,
          incMatchCounter: incMatchCounter,
          keys: keys,
        );
      }
      if (node is num) {
        return _splitWithHighlights(
          node.toString(),
          baseStyle: numberStyle,
          highlight: highlight,
          highlightFocused: highlightFocused,
          matchCounter: matchCounter,
          incMatchCounter: incMatchCounter,
          keys: keys,
        );
      }
      if (node is bool) {
        return _splitWithHighlights(
          node ? 'true' : 'false',
          baseStyle: boolStyle,
          highlight: highlight,
          highlightFocused: highlightFocused,
          matchCounter: matchCounter,
          incMatchCounter: incMatchCounter,
          keys: keys,
        );
      }
      if (node == null) {
        return _splitWithHighlights(
          'null',
          baseStyle: nullStyle,
          highlight: highlight,
          highlightFocused: highlightFocused,
          matchCounter: matchCounter,
          incMatchCounter: incMatchCounter,
          keys: keys,
        );
      }
      return _splitWithHighlights(
        node.toString(),
        baseStyle: punct,
        highlight: highlight,
        highlightFocused: highlightFocused,
        matchCounter: matchCounter,
        incMatchCounter: incMatchCounter,
        keys: keys,
      );
    }

    List<Widget> buildNode(dynamic node, String path, int indent) {
      final List<Widget> out = [];
      if (node is Map) {
        int idx = 0;
        final last = node.length - 1;
        for (final entry in node.entries) {
          final childPath = '$path.${entry.key}';
          final value = entry.value;
          final container = value is Map || value is List;
          final expanded = container && isExpanded(childPath);
          out.add(
            Padding(
              padding: EdgeInsets.only(left: indent * indentPx),
              child: GestureDetector(
                behavior: HitTestBehavior.opaque,
                onTap: () {
                  if (!container) return;
                  setState(() {
                    if (expanded) {
                      _userExpanded.remove(childPath);
                    } else {
                      _userExpanded.add(childPath);
                    }
                  });
                },
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    if (container)
                      SizedBox(
                        width: 18,
                        height: 18,
                        child: IconButton(
                          padding: const EdgeInsets.all(iconPad),
                          constraints: const BoxConstraints(
                            minWidth: 18,
                            minHeight: 18,
                          ),
                          iconSize: iconSize,
                          onPressed: () {
                            setState(() {
                              if (expanded) {
                                _userExpanded.remove(childPath);
                              } else {
                                _userExpanded.add(childPath);
                              }
                            });
                          },
                          icon: Icon(
                            expanded ? Icons.expand_more : Icons.chevron_right,
                            size: iconSize,
                          ),
                        ),
                      )
                    else
                      const SizedBox(width: 18, height: 18),
                    Expanded(
                      child: RichText(
                        text: TextSpan(
                          children: [
                            TextSpan(
                              children: _splitWithHighlights(
                                '"${entry.key}": ',
                                baseStyle: keyStyle,
                                highlight: hl,
                                highlightFocused: hlFocus,
                                matchCounter: () => matchCounter,
                                incMatchCounter: () {
                                  matchCounter++;
                                },
                                keys: keys,
                              ),
                              recognizer:
                                  TapGestureRecognizer()
                                    ..onTap = () {
                                      if (!container) return;
                                      setState(() {
                                        if (expanded) {
                                          _userExpanded.remove(childPath);
                                        } else {
                                          _userExpanded.add(childPath);
                                        }
                                      });
                                    },
                            ),
                            if (!container)
                              ..._valueToSpans(
                                value,
                                stringStyle,
                                numberStyle,
                                boolStyle,
                                nullStyle,
                                hl,
                                hlFocus,
                                () => matchCounter,
                                () {
                                  matchCounter++;
                                },
                                keys,
                              )
                            else
                              TextSpan(
                                text:
                                    expanded
                                        ? (value is Map ? '{' : '[')
                                        : (value is Map
                                            ? '{… ${value.length}}'
                                            : '[… ${value.length}]'),
                                style: punct,
                              ),
                            TextSpan(
                              text: (expanded ? '' : (idx != last ? ',' : '')),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          );
          if (container && expanded) {
            out.addAll(buildNode(value, childPath, indent + 1));
            out.add(
              Padding(
                padding: EdgeInsets.only(left: indent * indentPx + 18),
                child: SelectableText.rich(
                  TextSpan(
                    text: value is Map ? '}' : ']',
                    style: punct,
                    children: [TextSpan(text: idx != last ? ',' : '')],
                  ),
                ),
              ),
            );
          }
          idx++;
        }
        return out;
      }
      if (node is List) {
        for (int i = 0; i < node.length; i++) {
          final childPath = '$path[$i]';
          final value = node[i];
          final container = value is Map || value is List;
          final expanded = container && isExpanded(childPath);
          out.add(
            Padding(
              padding: EdgeInsets.only(left: indent * indentPx),
              child: GestureDetector(
                behavior: HitTestBehavior.opaque,
                onTap: () {
                  if (!container) return;
                  setState(() {
                    if (expanded) {
                      _userExpanded.remove(childPath);
                    } else {
                      _userExpanded.add(childPath);
                    }
                  });
                },
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    if (container)
                      SizedBox(
                        width: 18,
                        height: 18,
                        child: IconButton(
                          padding: const EdgeInsets.all(iconPad),
                          constraints: const BoxConstraints(
                            minWidth: 18,
                            minHeight: 18,
                          ),
                          iconSize: iconSize,
                          onPressed: () {
                            setState(() {
                              if (expanded) {
                                _userExpanded.remove(childPath);
                              } else {
                                _userExpanded.add(childPath);
                              }
                            });
                          },
                          icon: Icon(
                            expanded ? Icons.expand_more : Icons.chevron_right,
                            size: iconSize,
                          ),
                        ),
                      )
                    else
                      const SizedBox(width: 18, height: 18),
                    Expanded(
                      child: RichText(
                        text: TextSpan(
                          children: [
                            TextSpan(
                              text: '$i: ',
                              style: punct,
                              recognizer:
                                  TapGestureRecognizer()
                                    ..onTap = () {
                                      if (!container) return;
                                      setState(() {
                                        if (expanded) {
                                          _userExpanded.remove(childPath);
                                        } else {
                                          _userExpanded.add(childPath);
                                        }
                                      });
                                    },
                            ),
                            if (!container)
                              ..._valueToSpans(
                                value,
                                stringStyle,
                                numberStyle,
                                boolStyle,
                                nullStyle,
                                hl,
                                hlFocus,
                                () => matchCounter,
                                () {
                                  matchCounter++;
                                },
                                keys,
                              )
                            else
                              TextSpan(
                                text:
                                    expanded
                                        ? (value is Map ? '{' : '[')
                                        : (value is Map
                                            ? '{… ${value.length}}'
                                            : '[… ${value.length}]'),
                                style: punct,
                              ),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          );
          if (container && expanded) {
            out.addAll(buildNode(value, childPath, indent + 1));
            out.add(
              Padding(
                padding: EdgeInsets.only(left: indent * indentPx + 18),
                child: SelectableText.rich(
                  TextSpan(text: value is Map ? '}' : ']', style: punct),
                ),
              ),
            );
          }
        }
        return out;
      }
      return out;
    }

    // Build tree using a single recursive dispatcher
    final dynamic root =
        (widget.data is Map || widget.data is Iterable)
            ? widget.data
            : {'value': widget.data};
    final List<Widget> children = buildNode(root, '', 0);

    final content = Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: children,
    );

    if (widget.search.onRebuilt != null) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        widget.search.onRebuilt!(matchCounter, keys);
      });
    }

    return content;
  }

  List<InlineSpan> _splitWithHighlights(
    String text, {
    required TextStyle baseStyle,
    required Color highlight,
    required Color highlightFocused,
    required int Function() matchCounter,
    required void Function() incMatchCounter,
    required List<GlobalKey> keys,
  }) {
    final query = widget.search.query;
    if (query.isEmpty) {
      return [TextSpan(text: text, style: baseStyle)];
    }
    final List<InlineSpan> out = [];
    if (widget.search.useRegex) {
      RegExp? re;
      try {
        re = RegExp(query, caseSensitive: widget.search.matchCase);
      } catch (_) {
        return [TextSpan(text: text, style: baseStyle)];
      }
      int last = 0;
      Iterable<RegExpMatch> matches = re.allMatches(text);
      bool isWordChar(String ch) {
        final code = ch.codeUnitAt(0);
        final isAZ = (code >= 65 && code <= 90) || (code >= 97 && code <= 122);
        final is09 = (code >= 48 && code <= 57);
        return isAZ || is09 || ch == '_';
      }

      for (final m in matches) {
        final s = m.start;
        final e = m.end;
        if (widget.search.wholeWord) {
          final left = s - 1 >= 0 ? text.substring(s - 1, s) : null;
          final right = e < text.length ? text.substring(e, e + 1) : null;
          final leftOk = left == null || !isWordChar(left);
          final rightOk = right == null || !isWordChar(right);
          if (!(leftOk && rightOk)) {
            continue;
          }
        }
        if (s > last) {
          out.add(TextSpan(text: text.substring(last, s), style: baseStyle));
        }
        final key = GlobalKey();
        keys.add(key);
        // Небольшой якорь по базовой линии — помогает ensureVisible прокрутить точно к совпадению
        out.add(
          WidgetSpan(
            alignment: PlaceholderAlignment.baseline,
            baseline: TextBaseline.alphabetic,
            child: SizedBox(key: key, width: 0, height: 1),
          ),
        );
        final isFocused = matchCounter() == widget.search.focusedIndex;
        out.add(
          TextSpan(
            text: text.substring(s, e),
            style: baseStyle.copyWith(
              backgroundColor: isFocused ? highlightFocused : highlight,
            ),
          ),
        );
        incMatchCounter();
        last = e;
      }
      if (last < text.length) {
        out.add(TextSpan(text: text.substring(last), style: baseStyle));
      }
      return out;
    }
    final String src = widget.search.matchCase ? text : text.toLowerCase();
    final String q = widget.search.matchCase ? query : query.toLowerCase();
    int start = 0;
    bool isWordChar(String ch) {
      final code = ch.codeUnitAt(0);
      final isAZ = (code >= 65 && code <= 90) || (code >= 97 && code <= 122);
      final is09 = (code >= 48 && code <= 57);
      return isAZ || is09 || ch == '_';
    }

    while (true) {
      final idx = src.indexOf(q, start);
      if (idx < 0) {
        if (start < text.length) {
          out.add(TextSpan(text: text.substring(start), style: baseStyle));
        }
        break;
      }
      if (widget.search.wholeWord) {
        final left = idx - 1 >= 0 ? src.substring(idx - 1, idx) : null;
        final right =
            (idx + q.length) < src.length
                ? src.substring(idx + q.length, idx + q.length + 1)
                : null;
        final leftOk = left == null || !isWordChar(left);
        final rightOk = right == null || !isWordChar(right);
        if (!(leftOk && rightOk)) {
          start = idx + 1;
          continue;
        }
      }
      if (idx > start) {
        out.add(TextSpan(text: text.substring(start, idx), style: baseStyle));
      }
      final key = GlobalKey();
      keys.add(key);
      out.add(
        WidgetSpan(
          alignment: PlaceholderAlignment.baseline,
          baseline: TextBaseline.alphabetic,
          child: SizedBox(key: key, width: 0, height: 1),
        ),
      );
      final isFocused = matchCounter() == widget.search.focusedIndex;
      out.add(
        TextSpan(
          text: text.substring(idx, idx + q.length),
          style: baseStyle.copyWith(
            backgroundColor: isFocused ? highlightFocused : highlight,
          ),
        ),
      );
      incMatchCounter();
      start = idx + q.length;
    }
    return out;
  }
}

class JsonPrettyRich extends StatelessWidget {
  const JsonPrettyRich({required this.data, this.search});
  final dynamic data;
  final JsonSearchConfig? search;

  @override
  Widget build(BuildContext context) {
    final base = context.appText.monospace;
    final punct = base;
    final keyStyle = base.copyWith(color: context.appColors.primary);
    final stringStyle = base.copyWith(color: context.appColors.success);
    final numberStyle = base.copyWith(color: context.appColors.warning);
    final boolStyle = base.copyWith(color: context.appColors.warning);
    final nullStyle = base.copyWith(color: context.appColors.danger);

    final Color hl = context.appColors.warning.withValues(alpha: 0.35);
    final Color hlFocus = context.appColors.warning.withValues(alpha: 0.55);

    int matchCounter = 0;
    final List<GlobalKey> keys = [];

    final spans = _buildSpans(
      context,
      data,
      0,
      punct: punct,
      keyStyle: keyStyle,
      stringStyle: stringStyle,
      numberStyle: numberStyle,
      boolStyle: boolStyle,
      nullStyle: nullStyle,
      highlight: hl,
      highlightFocused: hlFocus,
      matchCounter: () => matchCounter,
      incMatchCounter: () {
        matchCounter++;
      },
      keys: keys,
    );

    if (search?.onRebuilt != null) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        // Передаем наверх количество совпадений и якоря после завершения build
        search!.onRebuilt!(matchCounter, keys);
      });
    }

    return SelectableText.rich(TextSpan(children: spans));
  }

  List<InlineSpan> _buildSpans(
    BuildContext context,
    dynamic node,
    int indent, {
    required TextStyle punct,
    required TextStyle keyStyle,
    required TextStyle stringStyle,
    required TextStyle numberStyle,
    required TextStyle boolStyle,
    required TextStyle nullStyle,
    required Color highlight,
    required Color highlightFocused,
    required int Function() matchCounter,
    required void Function() incMatchCounter,
    required List<GlobalKey> keys,
  }) {
    final List<InlineSpan> out = [];
    final String ind = '  ' * indent;

    void addP(String s) => out.addAll(
      _splitWithHighlights(
        s,
        baseStyle: punct,
        highlight: highlight,
        highlightFocused: highlightFocused,
        matchCounter: matchCounter,
        incMatchCounter: incMatchCounter,
        keys: keys,
      ),
    );

    if (node is Map) {
      addP('{\n');
      int i = 0;
      final last = node.length - 1;
      for (final entry in node.entries) {
        addP('$ind  ');
        out.addAll(
          _splitWithHighlights(
            '"${entry.key}"',
            baseStyle: keyStyle,
            highlight: highlight,
            highlightFocused: highlightFocused,
            matchCounter: matchCounter,
            incMatchCounter: incMatchCounter,
            keys: keys,
          ),
        );
        addP(': ');
        out.addAll(
          _buildSpans(
            context,
            entry.value,
            indent + 1,
            punct: punct,
            keyStyle: keyStyle,
            stringStyle: stringStyle,
            numberStyle: numberStyle,
            boolStyle: boolStyle,
            nullStyle: nullStyle,
            highlight: highlight,
            highlightFocused: highlightFocused,
            matchCounter: matchCounter,
            incMatchCounter: incMatchCounter,
            keys: keys,
          ),
        );
        if (i != last) addP(',');
        addP('\n');
        i++;
      }
      addP('$ind}');
      return out;
    }
    if (node is List) {
      addP('[\n');
      for (int i = 0; i < node.length; i++) {
        addP('$ind  ');
        out.addAll(
          _buildSpans(
            context,
            node[i],
            indent + 1,
            punct: punct,
            keyStyle: keyStyle,
            stringStyle: stringStyle,
            numberStyle: numberStyle,
            boolStyle: boolStyle,
            nullStyle: nullStyle,
            highlight: highlight,
            highlightFocused: highlightFocused,
            matchCounter: matchCounter,
            incMatchCounter: incMatchCounter,
            keys: keys,
          ),
        );
        if (i != node.length - 1) addP(',');
        addP('\n');
      }
      addP('$ind]');
      return out;
    }
    if (node is String) {
      out.addAll(
        _splitWithHighlights(
          '"$node"',
          baseStyle: stringStyle,
          highlight: highlight,
          highlightFocused: highlightFocused,
          matchCounter: matchCounter,
          incMatchCounter: incMatchCounter,
          keys: keys,
        ),
      );
      return out;
    }
    if (node is num) {
      out.addAll(
        _splitWithHighlights(
          node.toString(),
          baseStyle: numberStyle,
          highlight: highlight,
          highlightFocused: highlightFocused,
          matchCounter: matchCounter,
          incMatchCounter: incMatchCounter,
          keys: keys,
        ),
      );
      return out;
    }
    if (node is bool) {
      out.addAll(
        _splitWithHighlights(
          node ? 'true' : 'false',
          baseStyle: boolStyle,
          highlight: highlight,
          highlightFocused: highlightFocused,
          matchCounter: matchCounter,
          incMatchCounter: incMatchCounter,
          keys: keys,
        ),
      );
      return out;
    }
    if (node == null) {
      out.addAll(
        _splitWithHighlights(
          'null',
          baseStyle: nullStyle,
          highlight: highlight,
          highlightFocused: highlightFocused,
          matchCounter: matchCounter,
          incMatchCounter: incMatchCounter,
          keys: keys,
        ),
      );
      return out;
    }
    out.addAll(
      _splitWithHighlights(
        node.toString(),
        baseStyle: punct,
        highlight: highlight,
        highlightFocused: highlightFocused,
        matchCounter: matchCounter,
        incMatchCounter: incMatchCounter,
        keys: keys,
      ),
    );
    return out;
  }

  List<InlineSpan> _splitWithHighlights(
    String text, {
    required TextStyle baseStyle,
    required Color highlight,
    required Color highlightFocused,
    required int Function() matchCounter,
    required void Function() incMatchCounter,
    required List<GlobalKey> keys,
  }) {
    final cfg = search;
    final query = cfg?.query ?? '';
    if (query.isEmpty) {
      return [TextSpan(text: text, style: baseStyle)];
    }

    final String src = cfg!.matchCase ? text : text.toLowerCase();
    final String q = cfg.matchCase ? query : query.toLowerCase();
    int start = 0;
    final List<InlineSpan> out = [];

    bool isWordChar(String ch) {
      final code = ch.codeUnitAt(0);
      final isAZ = (code >= 65 && code <= 90) || (code >= 97 && code <= 122);
      final is09 = (code >= 48 && code <= 57);
      return isAZ || is09 || ch == '_';
    }

    int indexOfNext(int from) {
      int idx = src.indexOf(q, from);
      if (idx < 0) return -1;
      if (!cfg.wholeWord) return idx;
      final left = idx - 1 >= 0 ? src.substring(idx - 1, idx) : null;
      final right =
          (idx + q.length) < src.length
              ? src.substring(idx + q.length, idx + q.length + 1)
              : null;
      final leftOk = left == null || !isWordChar(left);
      final rightOk = right == null || !isWordChar(right);
      if (leftOk && rightOk) return idx;
      int nextStart = idx + 1;
      while (true) {
        idx = src.indexOf(q, nextStart);
        if (idx < 0) return -1;
        final l = idx - 1 >= 0 ? src.substring(idx - 1, idx) : null;
        final r =
            (idx + q.length) < src.length
                ? src.substring(idx + q.length, idx + q.length + 1)
                : null;
        final lo = l == null || !isWordChar(l);
        final ro = r == null || !isWordChar(r);
        if (lo && ro) return idx;
        nextStart = idx + 1;
      }
    }

    while (true) {
      final idx = indexOfNext(start);
      if (idx < 0) {
        if (start < text.length) {
          out.add(TextSpan(text: text.substring(start), style: baseStyle));
        }
        break;
      }
      if (idx > start) {
        out.add(TextSpan(text: text.substring(start, idx), style: baseStyle));
      }
      final key = GlobalKey();
      keys.add(key);
      out.add(
        WidgetSpan(
          alignment: PlaceholderAlignment.baseline,
          baseline: TextBaseline.alphabetic,
          child: SizedBox(key: key, width: 0, height: 1),
        ),
      );
      final isFocused = matchCounter() == cfg.focusedIndex;
      out.add(
        TextSpan(
          text: text.substring(idx, idx + q.length),
          style: baseStyle.copyWith(
            backgroundColor: isFocused ? highlightFocused : highlight,
          ),
        ),
      );
      incMatchCounter();
      start = idx + q.length;
    }

    return out;
  }
}

class JsonSearchConfig {
  const JsonSearchConfig({
    required this.query,
    required this.matchCase,
    required this.wholeWord,
    this.useRegex = false,
    required this.focusedIndex,
    required this.onRebuilt,
  });
  final String query;
  final bool matchCase;
  final bool wholeWord;
  final bool useRegex;
  final int focusedIndex;
  final void Function(int count, List<GlobalKey> keys)? onRebuilt;
}

// Интенты удалены — используются в CommonSearchBar
