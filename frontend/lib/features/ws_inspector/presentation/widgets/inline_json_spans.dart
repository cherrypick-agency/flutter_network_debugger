import 'package:flutter/material.dart';

import '../../../../theme/context_ext.dart';

/// Рекурсивно строит цветные InlineSpan для JSON-значения.
/// Ключи, строки, числа, bool, null — каждый со своим цветом.
List<InlineSpan> buildInlineJsonSpans(BuildContext context, dynamic node) {
  final base = context.appText.body;
  final punct = base;
  final keyStyle = base.copyWith(color: context.appColors.primary);
  final stringStyle = base.copyWith(color: context.appColors.success);
  final numberStyle = base.copyWith(color: context.appColors.warning);
  final boolStyle = base.copyWith(color: context.appColors.warning);
  final nullStyle = base.copyWith(color: context.appColors.danger);

  List<InlineSpan> build(dynamic n) {
    final List<InlineSpan> out = [];
    void add(String s, TextStyle st) => out.add(TextSpan(text: s, style: st));

    if (n is Map) {
      add('{', punct);
      int i = 0;
      final last = n.length - 1;
      for (final e in n.entries) {
        add('"${e.key}"', keyStyle);
        add(': ', punct);
        out.addAll(build(e.value));
        if (i != last) add(', ', punct);
        i++;
      }
      add('}', punct);
      return out;
    }
    if (n is List) {
      add('[', punct);
      for (int i = 0; i < n.length; i++) {
        out.addAll(build(n[i]));
        if (i != n.length - 1) add(', ', punct);
      }
      add(']', punct);
      return out;
    }
    if (n is String) {
      add('"$n"', stringStyle);
      return out;
    }
    if (n is num) {
      add(n.toString(), numberStyle);
      return out;
    }
    if (n is bool) {
      add(n ? 'true' : 'false', boolStyle);
      return out;
    }
    if (n == null) {
      add('null', nullStyle);
      return out;
    }
    add(n.toString(), punct);
    return out;
  }

  return build(node);
}
