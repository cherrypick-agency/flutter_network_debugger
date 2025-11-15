import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../../../../../theme/context_ext.dart';

/// Copyable key-value item widget
///
/// Displays a key-value pair with hover state and copy button.
/// Used for query parameters, headers, cookies, etc.
class CopyableKeyValueItem extends StatefulWidget {
  const CopyableKeyValueItem({
    super.key,
    required this.name,
    required this.value,
  });

  final String name;
  final String value;

  @override
  State<CopyableKeyValueItem> createState() => _CopyableKeyValueItemState();
}

class _CopyableKeyValueItemState extends State<CopyableKeyValueItem> {
  bool _hover = false;
  bool _iconHover = false;

  @override
  Widget build(BuildContext context) {
    final nameStyle = Theme.of(context).textTheme.bodySmall?.copyWith(
      fontFamily: 'monospace',
      fontWeight: FontWeight.w600,
    );
    final valueStyle = Theme.of(context).textTheme.bodySmall?.copyWith(
      fontFamily: 'monospace',
      color: context.appColors.textSecondary,
    );
    final iconSize = valueStyle?.fontSize ?? 12;
    final iconColor = valueStyle?.color;

    return MouseRegion(
      onEnter: (_) => setState(() => _hover = true),
      onExit: (_) => setState(() => _hover = false),
      child: SelectableText.rich(
        TextSpan(
          children: [
            TextSpan(text: '${widget.name}: ', style: nameStyle),
            TextSpan(text: widget.value, style: valueStyle),
            if (_hover)
              WidgetSpan(
                alignment: PlaceholderAlignment.baseline,
                baseline: TextBaseline.alphabetic,
                child: MouseRegion(
                  cursor: SystemMouseCursors.click,
                  onEnter: (_) => setState(() => _iconHover = true),
                  onExit: (_) => setState(() => _iconHover = false),
                  child: GestureDetector(
                    onTap: () {
                      Clipboard.setData(
                        ClipboardData(text: '${widget.name}: ${widget.value}'),
                      );
                    },
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 2,
                        vertical: 1,
                      ),
                      margin: const EdgeInsets.only(left: 6),
                      decoration: BoxDecoration(
                        color: _iconHover
                            ? Theme.of(
                                context,
                              ).colorScheme.primary.withValues(alpha: 0.12)
                            : Colors.transparent,
                        borderRadius: BorderRadius.circular(4),
                      ),
                      child: Icon(
                        Icons.copy,
                        size: iconSize,
                        color: _iconHover
                            ? Theme.of(context).colorScheme.primary
                            : iconColor,
                      ),
                    ),
                  ),
                ),
              ),
          ],
        ),
      ),
    );
  }
}

/// Header item widget with sensitive data masking and reveal functionality
///
/// Displays HTTP headers with:
/// - Automatic masking of sensitive data (Authorization, Cookie, tokens, etc.)
/// - Reveal/hide toggle button for masked values
/// - Copy button for copying header value
class HeaderItem extends StatefulWidget {
  const HeaderItem({
    super.key,
    required this.name,
    required this.value,
    this.raw,
  });

  final String name;
  final String value;
  final String? raw;

  @override
  State<HeaderItem> createState() => _HeaderItemState();
}

class _HeaderItemState extends State<HeaderItem> {
  bool _hover = false;
  bool _visibilityIconHover = false; // FIX: Separate state for visibility icon
  bool _copyIconHover = false; // FIX: Separate state for copy icon
  bool _reveal = false;

  /// Check if this header contains sensitive data
  bool get _isSensitive {
    final lname = widget.name.toLowerCase();
    return lname == 'authorization' ||
        lname == 'cookie' ||
        lname == 'set-cookie' ||
        lname.contains('token') ||
        lname.contains('secret') ||
        lname.contains('api-key') ||
        lname.contains('apikey');
  }

  @override
  Widget build(BuildContext context) {
    final nameStyle = Theme.of(context).textTheme.bodySmall?.copyWith(
      fontFamily: 'monospace',
      fontWeight: FontWeight.w600,
    );
    final valueStyle = Theme.of(context).textTheme.bodySmall?.copyWith(
      fontFamily: 'monospace',
      color: context.appColors.textSecondary,
    );
    final iconSize = valueStyle?.fontSize ?? 12;
    final iconColor = valueStyle?.color;

    // Use raw value if available, otherwise use masked value
    final raw = widget.raw ?? widget.value;
    final shownValue = (_isSensitive && !_reveal) ? '***' : raw;

    return MouseRegion(
      onEnter: (_) => setState(() => _hover = true),
      onExit: (_) => setState(() => _hover = false),
      child: SelectableText.rich(
        TextSpan(
          children: [
            TextSpan(text: '${widget.name}: ', style: nameStyle),
            TextSpan(text: shownValue, style: valueStyle),
            if (_hover)
              WidgetSpan(
                alignment: PlaceholderAlignment.baseline,
                baseline: TextBaseline.alphabetic,
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    // Reveal/Hide button for sensitive headers
                    if (_isSensitive)
                      MouseRegion(
                        cursor: SystemMouseCursors.click,
                        onEnter: (_) =>
                            setState(() => _visibilityIconHover = true),
                        onExit: (_) =>
                            setState(() => _visibilityIconHover = false),
                        child: GestureDetector(
                          onTap: () {
                            setState(() {
                              _reveal = !_reveal;
                            });
                          },
                          child: Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: 2,
                              vertical: 1,
                            ),
                            margin: const EdgeInsets.only(left: 6),
                            decoration: BoxDecoration(
                              color: _visibilityIconHover
                                  ? Theme.of(context).colorScheme.primary
                                        .withValues(alpha: 0.12)
                                  : Colors.transparent,
                              borderRadius: BorderRadius.circular(4),
                            ),
                            child: Icon(
                              _reveal ? Icons.visibility_off : Icons.visibility,
                              size: iconSize,
                              color: _visibilityIconHover
                                  ? Theme.of(context).colorScheme.primary
                                  : iconColor,
                            ),
                          ),
                        ),
                      ),
                    // Copy button
                    MouseRegion(
                      cursor: SystemMouseCursors.click,
                      onEnter: (_) => setState(() => _copyIconHover = true),
                      onExit: (_) => setState(() => _copyIconHover = false),
                      child: GestureDetector(
                        onTap: () {
                          Clipboard.setData(
                            ClipboardData(text: '${widget.name}: $raw'),
                          );
                        },
                        child: Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 2,
                            vertical: 1,
                          ),
                          margin: const EdgeInsets.only(left: 6),
                          decoration: BoxDecoration(
                            color: _copyIconHover
                                ? Theme.of(
                                    context,
                                  ).colorScheme.primary.withValues(alpha: 0.12)
                                : Colors.transparent,
                            borderRadius: BorderRadius.circular(4),
                          ),
                          child: Icon(
                            Icons.copy,
                            size: iconSize,
                            color: _copyIconHover
                                ? Theme.of(context).colorScheme.primary
                                : iconColor,
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
          ],
        ),
      ),
    );
  }
}
