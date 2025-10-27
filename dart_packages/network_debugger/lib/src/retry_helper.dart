import 'dart:async';
import 'dart:math';

/// Configuration for retry behavior.
class RetryConfig {
  /// Maximum number of retry attempts (excluding initial attempt).
  final int maxAttempts;

  /// Initial delay before first retry.
  final Duration initialDelay;

  /// Maximum delay between retries.
  final Duration maxDelay;

  /// Multiplier for exponential backoff.
  final double backoffMultiplier;

  /// Whether to add random jitter to delays.
  final bool useJitter;

  const RetryConfig({
    this.maxAttempts = 3,
    this.initialDelay = const Duration(seconds: 1),
    this.maxDelay = const Duration(seconds: 30),
    this.backoffMultiplier = 2.0,
    this.useJitter = true,
  });

  /// Default configuration for network operations.
  static const network = RetryConfig(
    maxAttempts: 3,
    initialDelay: Duration(seconds: 1),
    maxDelay: Duration(seconds: 30),
    backoffMultiplier: 2.0,
    useJitter: true,
  );

  /// Aggressive retry configuration for critical operations.
  static const aggressive = RetryConfig(
    maxAttempts: 5,
    initialDelay: Duration(milliseconds: 500),
    maxDelay: Duration(seconds: 60),
    backoffMultiplier: 2.0,
    useJitter: true,
  );

  /// Conservative retry configuration.
  static const conservative = RetryConfig(
    maxAttempts: 2,
    initialDelay: Duration(seconds: 2),
    maxDelay: Duration(seconds: 10),
    backoffMultiplier: 1.5,
    useJitter: false,
  );
}

/// Helper class for retrying operations with exponential backoff.
class RetryHelper {
  final RetryConfig config;
  final Random _random = Random();

  RetryHelper({RetryConfig? config}) : config = config ?? RetryConfig.network;

  /// Executes an operation with retry logic.
  ///
  /// The operation will be retried up to [config.maxAttempts] times if it throws.
  /// Only exceptions matching [retryIf] will trigger a retry.
  ///
  /// Example:
  /// ```dart
  /// final helper = RetryHelper();
  /// final result = await helper.execute(
  ///   () async => await downloadFile(url),
  ///   retryIf: (e) => e is SocketException || e is HttpException,
  ///   onRetry: (attempt, error) => print('Retry $attempt: $error'),
  /// );
  /// ```
  Future<T> execute<T>(
    Future<T> Function() operation, {
    bool Function(Exception)? retryIf,
    void Function(int attempt, Exception error)? onRetry,
  }) async {
    Exception? lastException;

    for (var attempt = 0; attempt <= config.maxAttempts; attempt++) {
      try {
        return await operation();
      } on Exception catch (e) {
        lastException = e;

        // Check if we should retry this exception
        if (retryIf != null && !retryIf(e)) {
          rethrow;
        }

        // If this was the last attempt, don't retry
        if (attempt >= config.maxAttempts) {
          rethrow;
        }

        // Notify about retry
        if (onRetry != null) {
          onRetry(attempt + 1, e);
        }

        // Calculate delay for next attempt
        final delay = _calculateDelay(attempt);
        await Future.delayed(delay);
      }
    }

    // This should never be reached, but satisfies the analyzer
    throw lastException!;
  }

  /// Calculates delay for the given attempt number.
  Duration _calculateDelay(int attempt) {
    // Calculate exponential backoff: initialDelay * (backoffMultiplier ^ attempt)
    final exponentialDelay = config.initialDelay.inMilliseconds *
        pow(config.backoffMultiplier, attempt);

    // Cap at max delay
    var delayMs = min(exponentialDelay, config.maxDelay.inMilliseconds.toDouble());

    // Add jitter if enabled (0-50% random variation)
    if (config.useJitter) {
      final jitter = _random.nextDouble() * 0.5; // 0-50%
      delayMs = delayMs * (1 + jitter);
    }

    return Duration(milliseconds: delayMs.round());
  }

  /// Convenience method for common retry patterns.
  ///
  /// Automatically retries on network-related exceptions:
  /// - SocketException
  /// - HttpException
  /// - TimeoutException
  Future<T> executeNetworkOperation<T>(
    Future<T> Function() operation, {
    void Function(int attempt, Exception error)? onRetry,
  }) {
    return execute(
      operation,
      retryIf: (e) =>
          e.toString().contains('SocketException') ||
          e.toString().contains('HttpException') ||
          e is TimeoutException,
      onRetry: onRetry,
    );
  }
}

/// Exception thrown when all retry attempts fail.
class RetryExhaustedException implements Exception {
  final String message;
  final Exception lastException;
  final int attempts;

  RetryExhaustedException(
    this.message,
    this.lastException,
    this.attempts,
  );

  @override
  String toString() =>
      'RetryExhaustedException: $message (after $attempts attempts). Last error: $lastException';
}
