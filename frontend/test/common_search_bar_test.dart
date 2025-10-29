import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_test/flutter_test.dart';

import 'package:frontend/widgets/common_search_bar.dart';

void main() {
  testWidgets('CommonSearchBar: actions, shortcuts, toggles', (tester) async {
    var changed = 0;
    var next = 0;
    var prev = 0;
    var closed = 0;
    var tMatch = 0;
    var tWhole = 0;
    var tRegex = 0;

    final ctrl = TextEditingController();

    await tester.pumpWidget(
      MaterialApp(
        home: Scaffold(
          body: Center(
            child: CommonSearchBar(
              controller: ctrl,
              countText: '0/0',
              matchCase: false,
              wholeWord: false,
              useRegex: false,
              canNavigate: true,
              onChanged: () => changed++,
              onNext: () => next++,
              onPrev: () => prev++,
              onClose: () => closed++,
              onToggleMatchCase: () => tMatch++,
              onToggleWholeWord: () => tWhole++,
              onToggleRegex: () => tRegex++,
            ),
          ),
        ),
      ),
    );

    // Ввод текста вызывает onChanged
    await tester.enterText(find.byType(TextField), 'abc');
    expect(changed > 0, isTrue);

    // onSubmitted заменён на Shortcut (Enter) => onNext
    await tester.sendKeyEvent(LogicalKeyboardKey.enter);
    await tester.pump();
    expect(next, greaterThanOrEqualTo(1));

    // Shift+Enter => Prev
    await tester.sendKeyDownEvent(LogicalKeyboardKey.shift);
    await tester.sendKeyEvent(LogicalKeyboardKey.enter);
    await tester.sendKeyUpEvent(LogicalKeyboardKey.shift);
    await tester.pump();
    expect(prev, greaterThanOrEqualTo(1));

    // Escape => Close
    await tester.sendKeyEvent(LogicalKeyboardKey.escape);
    await tester.pump();
    expect(closed, 1);

    // Кнопки навигации
    await tester.tap(find.byTooltip('Next match'));
    await tester.pump();
    expect(next, greaterThanOrEqualTo(2));

    await tester.tap(find.byTooltip('Previous match'));
    await tester.pump();
    expect(prev, greaterThanOrEqualTo(2));

    // Тогглы
    await tester.tap(find.byTooltip('Match case'));
    await tester.tap(find.byTooltip('Match whole word'));
    await tester.tap(find.byTooltip('Use regular expression'));
    await tester.pump();
    expect((tMatch, tWhole, tRegex), (1, 1, 1));

    // Кнопка закрытия
    await tester.tap(find.byTooltip('Close'));
    await tester.pump();
    expect(closed, 2);
  });
}


