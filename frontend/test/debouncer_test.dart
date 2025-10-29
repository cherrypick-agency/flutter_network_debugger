import 'package:flutter_test/flutter_test.dart';
import 'package:fake_async/fake_async.dart';
import 'package:frontend/core/utils/debouncer.dart';

void main() {
  group('Debouncer', () {
    test('выполняет действие один раз после задержки', () {
      // Arrange
      int calls = 0;
      final d = Debouncer(const Duration(milliseconds: 200));

      // Act
      fakeAsync((async) {
        d.run(() => calls++);
        async.elapse(const Duration(milliseconds: 199));
        // Assert до наступления таймера
        expect(calls, 0);

        async.elapse(const Duration(milliseconds: 1));
        expect(calls, 1);
      });
    });

    test('последний вызов побеждает (предыдущий отменяется)', () {
      // Arrange
      final d = Debouncer(const Duration(milliseconds: 100));
      int value = 0;

      // Act
      fakeAsync((async) {
        d.run(() => value = 1);
        async.elapse(const Duration(milliseconds: 60));
        d.run(() => value = 2);
        async.elapse(const Duration(milliseconds: 39));
        // Assert — ещё рано
        expect(value, 0);

        // Дождёмся полного интервала для второго вызова
        async.elapse(const Duration(milliseconds: 61));
        // Assert — сработал только второй вызов
        expect(value, 2);
      });
    });

    test('dispose отменяет запланированное действие', () {
      // Arrange
      final d = Debouncer(const Duration(milliseconds: 100));
      int calls = 0;

      // Act
      fakeAsync((async) {
        d.run(() => calls++);
        d.dispose();
        async.elapse(const Duration(milliseconds: 200));

        // Assert
        expect(calls, 0);
      });
    });
  });
}
