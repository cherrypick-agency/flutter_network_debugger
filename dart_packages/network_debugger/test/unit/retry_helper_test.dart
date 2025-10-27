import 'dart:async';
import 'package:test/test.dart';
import 'package:network_debugger/src/retry_helper.dart';

void main() {
  group('RetryHelper', () {
    test('execute succeeds on first attempt', () async {
      final helper = RetryHelper();
      var attempts = 0;

      final result = await helper.execute(() async {
        attempts++;
        return 'success';
      });

      expect(result, equals('success'));
      expect(attempts, equals(1));
    });

    test('execute retries on exception', () async {
      final helper = RetryHelper(
        config: const RetryConfig(
          maxAttempts: 3,
          initialDelay: Duration(milliseconds: 10),
        ),
      );
      var attempts = 0;

      final result = await helper.execute(() async {
        attempts++;
        if (attempts < 3) {
          throw Exception('Temporary error');
        }
        return 'success';
      });

      expect(result, equals('success'));
      expect(attempts, equals(3));
    });

    test('execute throws after max attempts', () async {
      final helper = RetryHelper(
        config: const RetryConfig(
          maxAttempts: 2,
          initialDelay: Duration(milliseconds: 10),
        ),
      );
      var attempts = 0;

      expect(
        () => helper.execute(() async {
          attempts++;
          throw Exception('Persistent error');
        }),
        throwsException,
      );

      await Future.delayed(const Duration(milliseconds: 100));
      expect(attempts, equals(3)); // Initial + 2 retries
    });

    test('execute respects retryIf condition', () async {
      final helper = RetryHelper(
        config: const RetryConfig(
          maxAttempts: 3,
          initialDelay: Duration(milliseconds: 10),
        ),
      );
      var attempts = 0;

      expect(
        () => helper.execute(
          () async {
            attempts++;
            throw FormatException('Not retryable');
          },
          retryIf: (e) => e is TimeoutException, // Only retry TimeoutException
        ),
        throwsA(isA<FormatException>()),
      );

      await Future.delayed(const Duration(milliseconds: 50));
      expect(attempts, equals(1)); // Should not retry
    });

    test('execute calls onRetry callback', () async {
      final helper = RetryHelper(
        config: const RetryConfig(
          maxAttempts: 2,
          initialDelay: Duration(milliseconds: 10),
        ),
      );
      var attempts = 0;
      final retryAttempts = <int>[];

      await helper.execute(
        () async {
          attempts++;
          if (attempts < 3) {
            throw Exception('Retry me');
          }
          return 'success';
        },
        onRetry: (attempt, error) {
          retryAttempts.add(attempt);
        },
      );

      expect(retryAttempts, equals([1, 2]));
    });

    test('RetryConfig network defaults are reasonable', () {
      const config = RetryConfig.network;

      expect(config.maxAttempts, equals(3));
      expect(config.initialDelay, equals(const Duration(seconds: 1)));
      expect(config.maxDelay, equals(const Duration(seconds: 30)));
      expect(config.backoffMultiplier, equals(2.0));
      expect(config.useJitter, isTrue);
    });

    test('RetryConfig aggressive retries more', () {
      const config = RetryConfig.aggressive;

      expect(config.maxAttempts, greaterThan(RetryConfig.network.maxAttempts));
    });

    test('RetryConfig conservative retries less', () {
      const config = RetryConfig.conservative;

      expect(config.maxAttempts, lessThan(RetryConfig.network.maxAttempts));
    });

    test('executeNetworkOperation retries on network errors', () async {
      final helper = RetryHelper(
        config: const RetryConfig(
          maxAttempts: 2,
          initialDelay: Duration(milliseconds: 10),
        ),
      );
      var attempts = 0;

      final result = await helper.executeNetworkOperation(() async {
        attempts++;
        if (attempts < 2) {
          throw TimeoutException('Network timeout');
        }
        return 'success';
      });

      expect(result, equals('success'));
      expect(attempts, equals(2));
    });

    test('delay increases exponentially', () async {
      final helper = RetryHelper(
        config: const RetryConfig(
          maxAttempts: 3,
          initialDelay: Duration(milliseconds: 100),
          backoffMultiplier: 2.0,
          useJitter: false, // Disable jitter for predictable testing
        ),
      );

      final delays = <Duration>[];
      var attempts = 0;

      final startTime = DateTime.now();

      try {
        await helper.execute(
          () async {
            if (attempts > 0) {
              delays.add(DateTime.now().difference(startTime));
            }
            attempts++;
            throw Exception('Keep failing');
          },
        );
      } catch (e) {
        // Expected to fail
      }

      // Verify exponential backoff (with some tolerance for timing)
      expect(delays.length, greaterThanOrEqualTo(2));

      // First delay should be around 100ms
      expect(delays[0].inMilliseconds, greaterThan(50));
      expect(delays[0].inMilliseconds, lessThan(200));

      // Second delay should be around 100ms + 200ms = 300ms total
      expect(delays[1].inMilliseconds, greaterThan(250));
      expect(delays[1].inMilliseconds, lessThan(450));
    });

    test('delay respects maxDelay', () async {
      final helper = RetryHelper(
        config: const RetryConfig(
          maxAttempts: 10,
          initialDelay: Duration(seconds: 1),
          maxDelay: Duration(seconds: 2),
          backoffMultiplier: 10.0,
          useJitter: false,
        ),
      );

      var attempts = 0;
      final startTime = DateTime.now();
      DateTime? lastAttemptTime;

      try {
        await helper.execute(
          () async {
            if (attempts > 0 && attempts < 5) {
              final now = DateTime.now();
              if (lastAttemptTime != null) {
                final delay = now.difference(lastAttemptTime!);
                // Delay should not exceed maxDelay
                expect(delay.inMilliseconds, lessThanOrEqualTo(2500));
              }
              lastAttemptTime = now;
            }
            attempts++;
            throw Exception('Keep failing');
          },
        );
      } catch (e) {
        // Expected to fail
      }
    });
  });
}
