import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:frontend/widgets/json_viewer.dart';
import 'package:frontend/theme/app_theme.dart';

void main() {
  Widget _wrap(Widget child) => MaterialApp(
    theme: buildLightTheme(),
    home: Scaffold(body: child),
  );

  testWidgets('JsonViewer: invalid JSON - shows raw text', (tester) async {
    // Arrange
    const raw = 'not json';

    // Act
    await tester.pumpWidget(_wrap(const JsonViewer(jsonString: raw)));

    // Assert
    expect(find.text(raw), findsOneWidget);
  });

  testWidgets('JsonPrettyRich: counts and highlights matches', (tester) async {
    // Arrange
    int count = -1;
    final cfg = JsonSearchConfig(
      query: 'id',
      matchCase: false,
      wholeWord: false,
      focusedIndex: 0,
      useRegex: false,
      onRebuilt: (c, _) => count = c,
    );
    final data = {'id': 1, 'name': 'idid'};

    // Act
    await tester.pumpWidget(_wrap(JsonPrettyRich(data: data, search: cfg)));
    await tester.pump();

    // Assert: "id" in key and twice in value "idid" => 3
    expect(count, 3);
  });

  testWidgets('JsonViewer: valid JSON renders prettified content', (
    tester,
  ) async {
    // Arrange
    const src = '{"id":1}';

    // Act
    await tester.pumpWidget(_wrap(const JsonViewer(jsonString: src)));

    // Assert: expect to see the key
    expect(find.textContaining('"id"'), findsWidgets);
  });
}
