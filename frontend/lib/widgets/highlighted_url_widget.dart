import 'package:flutter/material.dart';
import '../theme/app_theme.dart';
import '../theme/context_ext.dart';

/// Reusable widget for displaying URL with color highlighting of parts
/// and optional search highlighting
class HighlightedUrlWidget extends StatelessWidget {
  final String url;
  final String? searchQuery;
  final String? suppressHost;
  final TextStyle? baseStyle;
  final int? maxLines;
  final TextOverflow? overflow;
  final bool selectable;

  const HighlightedUrlWidget({
    super.key,
    required this.url,
    this.searchQuery,
    this.suppressHost,
    this.baseStyle,
    this.maxLines,
    this.overflow,
    this.selectable = false,
  });

  @override
  Widget build(BuildContext context) {
    final colors = context.appColors;
    final style =
        baseStyle ??
        Theme.of(
          context,
        ).textTheme.bodyMedium?.copyWith(fontFamily: 'monospace') ??
        const TextStyle(fontFamily: 'monospace');

    // 1. Format URL with suppressHost consideration
    final formattedUrl = _formatDisplayUrl(url, suppressHost: suppressHost);

    // 2. Parse URL and create colored spans
    final spans = _parseUrlWithSearch(
      formattedUrl,
      colors,
      style,
      searchQuery?.trim(),
    );

    // 3. Return widget
    if (selectable) {
      return SelectableText.rich(
        TextSpan(children: spans),
        maxLines: maxLines,
        style: style,
      );
    } else {
      return RichText(
        text: TextSpan(children: spans),
        maxLines: maxLines,
        overflow: overflow ?? TextOverflow.clip,
      );
    }
  }

  /// Formats URL for display with suppressHost consideration
  String _formatDisplayUrl(String raw, {String? suppressHost}) {
    try {
      final u = Uri.parse(raw);
      // If all elements have the same domain — hide scheme/domain/port
      if (suppressHost != null &&
          suppressHost.isNotEmpty &&
          u.host == suppressHost) {
        final String path = u.path.startsWith('/')
            ? u.path.substring(1)
            : u.path;
        final String q = u.query.isNotEmpty ? '?${u.query}' : '';
        final String frag = u.fragment.isNotEmpty ? '#${u.fragment}' : '';
        return '/$path$q$frag';
      }
      // Otherwise — keep as is, but normalize double encoding
      final String auth = u.hasAuthority
          ? '${u.host}${u.hasPort ? ':${u.port}' : ''}'
          : '';
      final String base = u.scheme.isNotEmpty ? '${u.scheme}://$auth' : auth;
      final String q = u.query.isNotEmpty ? '?${u.query}' : '';
      final String frag = u.fragment.isNotEmpty ? '#${u.fragment}' : '';
      return '$base${u.path}$q$frag';
    } catch (_) {
      return raw;
    }
  }

  /// Parses URL, applies color highlighting and search highlighting
  List<InlineSpan> _parseUrlWithSearch(
    String url,
    AppColors colors,
    TextStyle baseStyle,
    String? query,
  ) {
    final spans = <InlineSpan>[];

    try {
      final uri = Uri.parse(url);

      // 1. Protocol (if present) - base color
      if (uri.scheme.isNotEmpty) {
        _addSpansWithHighlight(
          spans,
          '${uri.scheme}://',
          baseStyle,
          query,
          colors,
        );
      }

      // 2. Domain (host + port) - blue
      if (uri.host.isNotEmpty) {
        final hostPart = uri.hasPort ? '${uri.host}:${uri.port}' : uri.host;
        _addSpansWithHighlight(
          spans,
          hostPart,
          baseStyle.copyWith(color: colors.primary),
          query,
          colors,
        );
      }

      // 3. Path - green
      if (uri.path.isNotEmpty) {
        _addSpansWithHighlight(
          spans,
          uri.path,
          baseStyle.copyWith(color: colors.success),
          query,
          colors,
        );
      }

      // 4. Query parameters - yellow
      if (uri.query.isNotEmpty) {
        _addSpansWithHighlight(
          spans,
          '?${uri.query}',
          baseStyle.copyWith(color: colors.warning),
          query,
          colors,
        );
      }

      // 5. Fragment (if present) - base color
      if (uri.fragment.isNotEmpty) {
        _addSpansWithHighlight(
          spans,
          '#${uri.fragment}',
          baseStyle,
          query,
          colors,
        );
      }
    } catch (e) {
      // If URL parsing failed, show as is
      _addSpansWithHighlight(spans, url, baseStyle, query, colors);
    }

    return spans;
  }

  /// Adds spans for text with search highlighting
  void _addSpansWithHighlight(
    List<InlineSpan> spans,
    String text,
    TextStyle style,
    String? query,
    AppColors colors,
  ) {
    // If no search query, just add text with color
    if (query == null || query.isEmpty) {
      spans.add(TextSpan(text: text, style: style));
      return;
    }

    // Otherwise search for matches and highlight them with yellow background
    final String src = text;
    final String srcLower = src.toLowerCase();
    final String qLower = query.toLowerCase();
    int start = 0;
    final Color highlightBg = colors.warning.withValues(alpha: 0.35);

    while (true) {
      final idx = srcLower.indexOf(qLower, start);
      if (idx < 0) {
        if (start < src.length) {
          spans.add(TextSpan(text: src.substring(start), style: style));
        }
        break;
      }
      if (idx > start) {
        spans.add(TextSpan(text: src.substring(start, idx), style: style));
      }
      spans.add(
        TextSpan(
          text: src.substring(idx, idx + query.length),
          style: style.copyWith(backgroundColor: highlightBg),
        ),
      );
      start = idx + query.length;
    }
  }
}
