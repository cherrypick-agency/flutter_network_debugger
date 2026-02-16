import 'package:flutter/material.dart';
import 'package:url_launcher/url_launcher.dart';

/// Кликабельная инлайн-ссылка для использования внутри Text.rich / TextSpan
class InlineLink extends StatelessWidget {
  final String text;
  final String url;
  final TextStyle? style;

  const InlineLink({
    super.key,
    required this.text,
    required this.url,
    this.style,
  });

  @override
  Widget build(BuildContext context) {
    final cs = Theme.of(context).colorScheme;
    final tt = Theme.of(context).textTheme;
    final baseStyle = style ?? tt.bodyMedium;

    return MouseRegion(
      cursor: SystemMouseCursors.click,
      child: GestureDetector(
        onTap: () => launchUrl(Uri.parse(url)),
        child: Text(
          text,
          style: baseStyle?.copyWith(
            color: cs.primary,
            decoration: TextDecoration.underline,
          ),
        ),
      ),
    );
  }
}
