import 'package:flutter/material.dart';
import '../../../../theme/context_ext.dart';
import '../../../../widgets/json_viewer.dart';

/// Текстовый виджет с подсветкой совпадений и якорями для навигации
class SearchableTextRich extends StatelessWidget {
  const SearchableTextRich({
    super.key,
    required this.text,
    required this.search,
    this.style,
  });

  final String text;
  final JsonSearchConfig search;
  final TextStyle? style;

  @override
  Widget build(BuildContext context) {
    final base = style ?? context.appText.monospace;
    final Color hl = context.appColors.warning.withValues(alpha: 0.35);
    final Color hlFocus = context.appColors.warning.withValues(alpha: 0.55);

    int matchCounter = 0;
    final List<GlobalKey> keys = [];

    final spans = _splitWithHighlights(
      text,
      baseStyle: base,
      highlight: hl,
      highlightFocused: hlFocus,
      matchCounter: () => matchCounter,
      incMatchCounter: () { matchCounter++; },
      keys: keys,
    );

    if (search.onRebuilt != null) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        search.onRebuilt!(matchCounter, keys);
      });
    }

    return SelectableText.rich(TextSpan(children: spans));
  }

  List<InlineSpan> _splitWithHighlights(
    String srcText, {
    required TextStyle baseStyle,
    required Color highlight,
    required Color highlightFocused,
    required int Function() matchCounter,
    required void Function() incMatchCounter,
    required List<GlobalKey> keys,
  }) {
    final cfg = search;
    final query = cfg.query;
    if (query.isEmpty) {
      return [TextSpan(text: srcText, style: baseStyle)];
    }
    final List<InlineSpan> out = [];
    if (cfg.useRegex) {
      RegExp? re;
      try {
        re = RegExp(query, caseSensitive: cfg.matchCase);
      } catch (_) {
        return [TextSpan(text: srcText, style: baseStyle)];
      }
      int last = 0;
      Iterable<RegExpMatch> matches = re.allMatches(srcText);
      bool isWordChar(String ch) {
        final code = ch.codeUnitAt(0);
        final isAZ = (code >= 65 && code <= 90) || (code >= 97 && code <= 122);
        final is09 = (code >= 48 && code <= 57);
        return isAZ || is09 || ch == '_';
      }
      for (final m in matches) {
        final s = m.start;
        final e = m.end;
        if (cfg.wholeWord) {
          final left = s - 1 >= 0 ? srcText.substring(s - 1, s) : null;
          final right = e < srcText.length ? srcText.substring(e, e + 1) : null;
          final leftOk = left == null || !isWordChar(left);
          final rightOk = right == null || !isWordChar(right);
          if (!(leftOk && rightOk)) {
            continue;
          }
        }
        if (s > last) {
          out.add(TextSpan(text: srcText.substring(last, s), style: baseStyle));
        }
        final key = GlobalKey();
        keys.add(key);
        out.add(WidgetSpan(child: SizedBox(key: key, width: 0, height: 0)));
        final isFocused = matchCounter() == cfg.focusedIndex;
        out.add(TextSpan(
          text: srcText.substring(s, e),
          style: baseStyle.copyWith(backgroundColor: isFocused ? highlightFocused : highlight),
        ));
        incMatchCounter();
        last = e;
      }
      if (last < srcText.length) {
        out.add(TextSpan(text: srcText.substring(last), style: baseStyle));
      }
      return out;
    }
    final String src = cfg.matchCase ? srcText : srcText.toLowerCase();
    final String q = cfg.matchCase ? query : query.toLowerCase();
    int start = 0;
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
      final right = (idx + q.length) < src.length ? src.substring(idx + q.length, idx + q.length + 1) : null;
      final leftOk = left == null || !isWordChar(left);
      final rightOk = right == null || !isWordChar(right);
      if (leftOk && rightOk) return idx;
      int nextStart = idx + 1;
      while (true) {
        idx = src.indexOf(q, nextStart);
        if (idx < 0) return -1;
        final l = idx - 1 >= 0 ? src.substring(idx - 1, idx) : null;
        final r = (idx + q.length) < src.length ? src.substring(idx + q.length, idx + q.length + 1) : null;
        final lo = l == null || !isWordChar(l);
        final ro = r == null || !isWordChar(r);
        if (lo && ro) return idx;
        nextStart = idx + 1;
      }
    }

    while (true) {
      final idx = indexOfNext(start);
      if (idx < 0) {
        if (start < srcText.length) {
          out.add(TextSpan(text: srcText.substring(start), style: baseStyle));
        }
        break;
      }
      if (idx > start) {
        out.add(TextSpan(text: srcText.substring(start, idx), style: baseStyle));
      }
      final key = GlobalKey();
      keys.add(key);
      out.add(WidgetSpan(child: SizedBox(key: key, width: 0, height: 0)));
      final isFocused = matchCounter() == cfg.focusedIndex;
      out.add(TextSpan(
        text: srcText.substring(idx, idx + q.length),
        style: baseStyle.copyWith(backgroundColor: isFocused ? highlightFocused : highlight),
      ));
      incMatchCounter();
      start = idx + q.length;
    }

    return out;
  }
}


