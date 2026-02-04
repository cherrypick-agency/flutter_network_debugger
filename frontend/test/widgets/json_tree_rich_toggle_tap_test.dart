import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:frontend/theme/app_theme.dart';
import 'package:frontend/widgets/json_viewer.dart';

void main() {
  testWidgets('JsonTreeRich: tap on row expands/collapses nested object', (
    tester,
  ) async {
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

    // Array element is initially hidden (data is not top-level => collapsed)
    expect(textWith('0:'), findsNothing);

    // Tap on '"data": {… 1}' row expands it
    await tester.tap(textWith('"data":'));
    await tester.pump();
    expect(textWith('0:'), findsOneWidget);

    // Tap on closing bracket also collapses
    await tester.tap(textWith(']'));
    await tester.pump();
    expect(textWith('0:'), findsNothing);
  });
}
