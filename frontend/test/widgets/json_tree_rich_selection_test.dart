import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:frontend/widgets/json_viewer.dart';
import 'package:frontend/theme/app_theme.dart';

void main() {
  testWidgets(
    'JsonTreeRich: в строках нет recognizer у TextSpan (выделение не блокируется)',
    (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          theme: buildLightTheme(),
          home: Scaffold(
            body: JsonTreeRich(
              data: const {
                'account': {'id': 1, 'name': 'test'},
              },
              search: const JsonSearchConfig(
                query: '',
                matchCase: false,
                wholeWord: false,
                focusedIndex: 0,
                onRebuilt: null,
              ),
            ),
          ),
        ),
      );

      bool hasRecognizer(InlineSpan span) {
        if (span is TextSpan) {
          if (span.recognizer != null) return true;
          final children = span.children;
          if (children != null) {
            for (final c in children) {
              if (hasRecognizer(c)) return true;
            }
          }
        }
        return false;
      }

      expect(find.byType(SelectionArea), findsOneWidget);

      final texts = tester.widgetList<Text>(find.byType(Text));
      expect(texts, isNotEmpty);

      for (final t in texts) {
        final span = t.textSpan;
        if (span == null) continue;
        expect(hasRecognizer(span), isFalse);
      }
    },
  );
}
