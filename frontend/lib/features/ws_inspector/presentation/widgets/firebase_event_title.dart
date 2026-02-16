import 'dart:convert';

import 'package:flutter/material.dart';

import '../../../../theme/app_theme.dart';
import '../../../../theme/context_ext.dart';
import '../../../../widgets/json_viewer.dart';
import 'firebase_event_data.dart';
import 'inline_json_spans.dart';
import 'text_search_highlighter.dart';

/// Заголовок Firebase-события: бейджи + компактный контент + время.
///
/// Требования:
/// - всё в одну строку
/// - после бейджей идёт контент (payload), с подсветкой совпадений поиска
/// - время вторично: если места мало, оно режется первым
class FirebaseEventTitle extends StatelessWidget {
  const FirebaseEventTitle({
    super.key,
    required this.event,
    required this.contentPreview,
    required this.timestamp,
    required this.search,
  });

  final FirebaseEventData event;
  final String contentPreview;
  final String timestamp;
  final JsonSearchConfig search;

  @override
  Widget build(BuildContext context) {
    final colors = context.appColors;
    final labelStyle = Theme.of(
      context,
    ).textTheme.labelSmall?.copyWith(fontWeight: FontWeight.w600, fontSize: 11);

    return Row(
      children: [
        _Chip(
          label: event.op.toUpperCase(),
          color: _opColor(event.op, colors),
          textStyle: labelStyle,
        ),
        const SizedBox(width: 4),
        Flexible(
          flex: 0,
          child: _Chip(
            label: event.path,
            color: colors.primary,
            textStyle: labelStyle,
            maxWidth: 240,
          ),
        ),
        const SizedBox(width: 4),
        if (!event.ok && (event.error?.isNotEmpty ?? false))
          Flexible(
            flex: 0,
            child: _Chip(
              label: event.error!,
              color: colors.danger,
              textStyle: labelStyle,
              maxWidth: 160,
            ),
          )
        else
          _Chip(label: 'OK', color: colors.success, textStyle: labelStyle),
        const SizedBox(width: 8),
        Expanded(
          child: _SingleLineHighlightedText(
            text: contentPreview,
            payload: event.payload,
            search: search,
          ),
        ),
        const SizedBox(width: 8),
        // Время не должно ломать строку — если места мало, оно режется.
        ConstrainedBox(
          constraints: const BoxConstraints(maxWidth: 72),
          child: Text(
            timestamp,
            maxLines: 1,
            overflow: TextOverflow.ellipsis,
            textAlign: TextAlign.right,
            style: Theme.of(
              context,
            ).textTheme.labelSmall?.copyWith(color: colors.textSecondary),
          ),
        ),
      ],
    );
  }

  Color _opColor(String op, AppColors colors) {
    return switch (op) {
      'set' || 'update' => colors.warning,
      'remove' => colors.danger,
      'get' ||
      'onValue' ||
      'onChildAdded' ||
      'onChildChanged' ||
      'onChildRemoved' => colors.primary,
      _ => colors.textSecondary,
    };
  }

  static String buildCompactPreview(dynamic payload, {int maxChars = 220}) {
    if (payload == null) return '';
    try {
      final txt = jsonEncode(payload);
      final oneLine = txt.replaceAll('\n', ' ').replaceAll('\r', ' ').trim();
      if (oneLine.length <= maxChars) return oneLine;
      return '${oneLine.substring(0, maxChars)}…';
    } catch (_) {
      final s = payload.toString();
      if (s.length <= maxChars) return s;
      return '${s.substring(0, maxChars)}…';
    }
  }
}

/// Компактный однострочный превью payload с JSON syntax highlighting
/// и подсветкой совпадений поиска поверх.
class _SingleLineHighlightedText extends StatelessWidget {
  const _SingleLineHighlightedText({
    required this.text,
    required this.payload,
    required this.search,
  });

  final String text;
  final dynamic payload;
  final JsonSearchConfig search;

  @override
  Widget build(BuildContext context) {
    if (text.isEmpty) {
      return Text(
        'no payload',
        maxLines: 1,
        style: context.appText.body.copyWith(
          color: context.appColors.textSecondary,
          fontStyle: FontStyle.italic,
        ),
      );
    }

    // Если есть поисковый запрос — используем текстовую подсветку совпадений
    if (search.query.isNotEmpty) {
      final base = context.appText.body;
      final Color hl = context.appColors.warning.withValues(alpha: 0.28);
      final Color hlFocus = context.appColors.warning.withValues(alpha: 0.45);

      final res = const TextSearchHighlighter().buildSpans(
        text: text,
        search: search,
        baseStyle: base,
        highlight: hl,
        highlightFocused: hlFocus,
        includeAnchors: false,
      );

      return Text.rich(
        TextSpan(children: res.spans),
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        style: base,
      );
    }

    // Без поиска — JSON syntax highlighting (цветные ключи/значения)
    if (payload != null) {
      final spans = buildInlineJsonSpans(context, payload);
      return Text.rich(
        TextSpan(children: spans),
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
        style: context.appText.body,
      );
    }

    return Text(
      text,
      maxLines: 1,
      overflow: TextOverflow.ellipsis,
      style: context.appText.body,
    );
  }
}

class _Chip extends StatelessWidget {
  const _Chip({
    required this.label,
    required this.color,
    this.textStyle,
    this.maxWidth,
  });

  final String label;
  final Color color;
  final TextStyle? textStyle;
  final double? maxWidth;

  @override
  Widget build(BuildContext context) {
    return Container(
      constraints: maxWidth != null
          ? BoxConstraints(maxWidth: maxWidth!)
          : null,
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.15),
        borderRadius: BorderRadius.circular(4),
        border: Border.all(color: color.withValues(alpha: 0.3), width: 0.5),
      ),
      child: Text(
        label,
        style: textStyle?.copyWith(color: color),
        maxLines: 1,
        overflow: TextOverflow.ellipsis,
      ),
    );
  }
}
