import 'package:flutter/material.dart';
import '../../../../theme/context_ext.dart';
import '../../../../widgets/json_viewer.dart';
import 'text_search_highlighter.dart';

/// Text widget with match highlighting and anchors for navigation
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

    final res = const TextSearchHighlighter().buildSpans(
      text: text,
      search: search,
      baseStyle: base,
      highlight: hl,
      highlightFocused: hlFocus,
      includeAnchors: true,
    );

    if (search.onRebuilt != null) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        search.onRebuilt!(res.matchCount, res.anchors);
      });
    }

    return SelectableText.rich(TextSpan(children: res.spans));
  }
}
