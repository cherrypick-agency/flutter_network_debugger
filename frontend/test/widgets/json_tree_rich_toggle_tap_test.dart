import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:frontend/theme/app_theme.dart';
import 'package:frontend/widgets/json_viewer.dart';

void main() {
  testWidgets(
    'JsonTreeRich: тап по строке раскрывает/сворачивает вложенный объект',
    (tester) async {
      await tester.pumpWidget(
        MaterialApp(
          theme: buildLightTheme(),
          home: const Scaffold(
            body: JsonTreeRich(
              data: {
                'account': {
                  'data': [1],
                },
              },
              search: JsonSearchConfig(
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

      Finder textWith(String needle) => find.byWidgetPredicate((w) {
        if (w is! Text) return false;
        final t = w.textSpan?.toPlainText() ?? w.data ?? '';
        return t.contains(needle);
      });

      // Элемент массива изначально скрыт (data не топ-уровень => свернут)
      expect(textWith('0:'), findsNothing);

      // Тап по строке '"data": {… 1}' раскрывает
      await tester.tap(textWith('"data":'));
      await tester.pump();
      expect(textWith('0:'), findsOneWidget);

      // Тап по закрывающей скобке тоже сворачивает
      await tester.tap(textWith(']'));
      await tester.pump();
      expect(textWith('0:'), findsNothing);
    },
  );
}
