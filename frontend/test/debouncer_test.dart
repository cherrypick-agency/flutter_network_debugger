import 'package:flutter_test/flutter_test.dart';
import 'package:fake_async/fake_async.dart';
import 'package:frontend/core/utils/debouncer.dart';

void main() {
  group('Debouncer', () {
    test('executes action once after delay', () {
      // Arrange
      int calls = 0;
      final d = Debouncer(const Duration(milliseconds: 200));

      // Act
      fakeAsync((async) {
        d.run(() => calls++);
        async.elapse(const Duration(milliseconds: 199));
        // Assert before timer fires
        expect(calls, 0);

        async.elapse(const Duration(milliseconds: 1));
        expect(calls, 1);
      });
    });

    test('last call wins (previous is cancelled)', () {
      // Arrange
      final d = Debouncer(const Duration(milliseconds: 100));
      int value = 0;

      // Act
      fakeAsync((async) {
        d.run(() => value = 1);
        async.elapse(const Duration(milliseconds: 60));
        d.run(() => value = 2);
        async.elapse(const Duration(milliseconds: 39));
        // Assert - too early
        expect(value, 0);

        // Wait for full interval for the second call
        async.elapse(const Duration(milliseconds: 61));
        // Assert - only the second call was executed
        expect(value, 2);
      });
    });

    test('dispose cancels scheduled action', () {
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
