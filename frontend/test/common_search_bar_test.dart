import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:frontend/widgets/common_search_bar.dart';
import 'package:frontend/theme/app_theme.dart';

void main() {
  Widget _wrap(Widget child) => MaterialApp(
    theme: buildLightTheme(),
    home: Scaffold(body: Center(child: child)),
  );

  testWidgets(
    'CommonSearchBar: onChanged and navigation disabled when canNavigate=false',
    (tester) async {
      // Arrange
      final ctrl = TextEditingController();
      int changes = 0;
      int next = 0;
      int prev = 0;

      final bar = CommonSearchBar(
        controller: ctrl,
        focusNode: FocusNode(),
        countText: '0/0',
        matchCase: false,
        wholeWord: false,
        useRegex: false,
        canNavigate: false,
        onChanged: () => changes++,
        onNext: () => next++,
        onPrev: () => prev++,
        onClose: () {},
        onToggleMatchCase: () {},
        onToggleWholeWord: () {},
        onToggleRegex: () {},
      );

      // Act
      await tester.pumpWidget(_wrap(bar));
      await tester.enterText(find.byType(TextField), 'abc');

      // Assert
      expect(changes, greaterThan(0));

      // Press next/prev buttons - should not trigger
      await tester.tap(find.byIcon(Icons.keyboard_arrow_down));
      await tester.tap(find.byIcon(Icons.keyboard_arrow_up));
      await tester.pump();
      expect(next, 0);
      expect(prev, 0);
    },
  );

  testWidgets('CommonSearchBar: hotkeys Enter/Shift+Enter/Escape', (
    tester,
  ) async {
    // Arrange
    final ctrl = TextEditingController();
    int next = 0;
    int prev = 0;
    int closed = 0;

    final bar = CommonSearchBar(
      controller: ctrl,
      countText: '1/3',
      matchCase: false,
      wholeWord: false,
      useRegex: false,
      canNavigate: true,
      onChanged: () {},
      onNext: () => next++,
      onPrev: () => prev++,
      onClose: () => closed++,
      onToggleMatchCase: () {},
      onToggleWholeWord: () {},
      onToggleRegex: () {},
    );

    // Act
    await tester.pumpWidget(_wrap(bar));
    await tester.pump();

    // Enter => next
    await tester.sendKeyEvent(LogicalKeyboardKey.enter);
    // Shift+Enter => prev
    await tester.sendKeyDownEvent(LogicalKeyboardKey.shift);
    await tester.sendKeyEvent(LogicalKeyboardKey.enter);
    await tester.sendKeyUpEvent(LogicalKeyboardKey.shift);
    // Escape => close
    await tester.sendKeyEvent(LogicalKeyboardKey.escape);

    // Assert
    expect(next, 1);
    expect(prev, 1);
    expect(closed, 1);
  });
}
