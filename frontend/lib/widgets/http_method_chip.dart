import 'package:flutter/material.dart';

/// Get background color for HTTP method
Color getHttpMethodBgColor(BuildContext context, String method) {
  final m = method.toUpperCase();
  final cs = Theme.of(context).colorScheme;
  if (m == 'GET') return Colors.purple.withValues(alpha: 0.12);
  if (m == 'POST') return cs.primary.withValues(alpha: 0.12);
  if (m == 'PUT' || m == 'PATCH') return cs.tertiary.withValues(alpha: 0.12);
  if (m == 'DELETE') return cs.error.withValues(alpha: 0.12);
  return cs.surfaceContainerHighest;
}

/// Get text color for HTTP method
Color getHttpMethodFgColor(BuildContext context, String method) {
  final m = method.toUpperCase();
  final cs = Theme.of(context).colorScheme;
  if (m == 'GET') return Colors.purple;
  if (m == 'POST') return cs.primary;
  if (m == 'PUT' || m == 'PATCH') return cs.tertiary;
  if (m == 'DELETE') return cs.error;
  return cs.onSurfaceVariant;
}

/// Widget for displaying HTTP method with color highlighting
class HttpMethodChip extends StatelessWidget {
  final String method;
  final TextStyle? style;

  const HttpMethodChip({super.key, required this.method, this.style});

  @override
  Widget build(BuildContext context) {
    final bg = getHttpMethodBgColor(context, method);
    final fg = getHttpMethodFgColor(context, method);

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: bg,
        borderRadius: BorderRadius.circular(6),
      ),
      child: Text(
        method.toUpperCase(),
        style: (style ?? Theme.of(context).textTheme.labelSmall)?.copyWith(
          color: fg,
        ),
      ),
    );
  }
}
