import 'package:flutter/material.dart';
import '../../../../theme/app_theme.dart';
import '../../../../theme/context_ext.dart';

/// Виджет для отображения URL с цветовой подсветкой частей
class HighlightedUrlText extends StatelessWidget {
  final String url;
  final TextStyle? baseStyle;

  const HighlightedUrlText({super.key, required this.url, this.baseStyle});

  @override
  Widget build(BuildContext context) {
    final colors = context.appColors;
    final style = baseStyle ?? context.appText.subtitle;

    // Парсим URL
    final spans = _parseUrl(url, colors, style);

    return SelectableText.rich(TextSpan(children: spans));
  }

  /// Парсит URL и возвращает список TextSpan с разными цветами
  List<TextSpan> _parseUrl(String url, AppColors colors, TextStyle baseStyle) {
    final spans = <TextSpan>[];

    try {
      final uri = Uri.parse(url);

      // 1. Протокол (если есть) - базовый цвет
      if (uri.scheme.isNotEmpty) {
        spans.add(TextSpan(text: '${uri.scheme}://', style: baseStyle));
      }

      // 2. Домен (host + port) - голубой
      if (uri.host.isNotEmpty) {
        final hostPart = uri.hasPort ? '${uri.host}:${uri.port}' : uri.host;
        spans.add(
          TextSpan(
            text: hostPart,
            style: baseStyle.copyWith(color: colors.primary),
          ),
        );
      }

      // 3. Путь - зелёный
      if (uri.path.isNotEmpty) {
        spans.add(
          TextSpan(
            text: uri.path,
            style: baseStyle.copyWith(color: colors.success),
          ),
        );
      }

      // 4. Query параметры - жёлтый
      if (uri.query.isNotEmpty) {
        spans.add(
          TextSpan(
            text: '?${uri.query}',
            style: baseStyle.copyWith(color: colors.warning),
          ),
        );
      }

      // 5. Fragment (если есть) - базовый цвет
      if (uri.fragment.isNotEmpty) {
        spans.add(TextSpan(text: '#${uri.fragment}', style: baseStyle));
      }
    } catch (e) {
      // Если URL не удалось распарсить, показываем как есть
      spans.add(TextSpan(text: url, style: baseStyle));
    }

    return spans;
  }
}
