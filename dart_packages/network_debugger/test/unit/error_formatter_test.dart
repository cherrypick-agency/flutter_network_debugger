import 'dart:async';
import 'dart:io';
import 'package:test/test.dart';
import 'package:network_debugger/src/error_formatter.dart' as error_fmt;

void main() {
  group('ErrorFormatter', () {
    test('format handles SocketException', () {
      final error = SocketException('Connection refused');
      final formatted = error_fmt.ErrorFormatter.format(error);

      expect(formatted, contains('Network connection error'));
      expect(formatted, contains('internet connection'));
    });

    test('format handles TimeoutException', () {
      final error = TimeoutException('Operation timed out');
      final formatted = error_fmt.ErrorFormatter.format(error);

      expect(formatted, contains('timed out'));
      expect(formatted, contains('Slow internet connection'));
    });

    test('format handles HTTP 404 errors', () {
      final error = Exception('HTTP 404 Not Found');
      final formatted = error_fmt.ErrorFormatter.format(error);

      expect(formatted, contains('404'));
      expect(formatted, contains('Resource not found'));
    });

    test('format handles HTTP 403 errors', () {
      final error = Exception('HTTP 403 Forbidden');
      final formatted = error_fmt.ErrorFormatter.format(error);

      expect(formatted, contains('403'));
      expect(formatted, contains('rate limit'));
    });

    test('format handles UnsupportedError', () {
      final error = UnsupportedError('Platform not supported');
      final formatted = error_fmt.ErrorFormatter.format(error);

      expect(formatted, contains('Platform not supported'));
      expect(formatted, contains('Supported platforms'));
    });

    test('format handles permission errors', () {
      final error = Exception('Permission denied');
      final formatted = error_fmt.ErrorFormatter.format(error);

      expect(formatted, contains('Permission denied'));
      expect(formatted, contains('permissions'));
    });

    test('format handles checksum errors', () {
      final error = Exception('ChecksumValidationException: mismatch');
      final formatted = error_fmt.ErrorFormatter.format(error);

      expect(formatted, contains('Checksum'));
      expect(formatted, contains('Security'));
    });

    test('format handles GitHub API errors', () {
      final error = Exception('GitHubReleaseException: failed to fetch');
      final formatted = error_fmt.ErrorFormatter.format(error);

      expect(formatted, contains('GitHub'));
      expect(formatted, contains('API'));
    });

    test('format handles archive extraction errors', () {
      final error = Exception('Failed to extract archive');
      final formatted = error_fmt.ErrorFormatter.format(error);

      expect(formatted, contains('extract'));
      expect(formatted, contains('archive'));
    });

    test('format handles process errors', () {
      final error = Exception('ProcessException: failed to start');
      final formatted = error_fmt.ErrorFormatter.format(error);

      expect(formatted, contains('process'));
      expect(formatted, contains('debugger'));
    });

    test('format handles port conflicts', () {
      final error = Exception('ProcessException: port already in use');
      final formatted = error_fmt.ErrorFormatter.format(error);

      expect(formatted, contains('port'));
      expect(formatted, contains('already'));
    });

    test('format handles cache errors', () {
      final error = Exception('CacheException: cannot write');
      final formatted = error_fmt.ErrorFormatter.format(error);

      expect(formatted, contains('Cache'));
      expect(formatted, contains('cache'));
    });

    test('format handles retry exhausted errors', () {
      final error = Exception('RetryExhaustedException: all attempts failed');
      final formatted = error_fmt.ErrorFormatter.format(error);

      expect(formatted, contains('retry'));
      expect(formatted, contains('multiple'));
    });

    test('format handles generic errors', () {
      final error = Exception('Some random error');
      final formatted = error_fmt.ErrorFormatter.format(error);

      expect(formatted, contains('error'));
      expect(formatted, isNotEmpty);
    });

    test('format handles null error', () {
      final formatted = error_fmt.ErrorFormatter.format(null);

      expect(formatted, contains('Unknown error'));
    });

    test('formatShort returns concise message for SocketException', () {
      final error = SocketException('Connection refused');
      final formatted = error_fmt.ErrorFormatter.formatShort(error);

      expect(formatted, equals('Network connection error'));
    });

    test('formatShort returns concise message for TimeoutException', () {
      final error = TimeoutException('Timed out');
      final formatted = error_fmt.ErrorFormatter.formatShort(error);

      expect(formatted, equals('Operation timed out'));
    });

    test('formatShort returns concise message for 404', () {
      final error = Exception('HTTP 404 Not Found');
      final formatted = error_fmt.ErrorFormatter.formatShort(error);

      expect(formatted, contains('404'));
    });

    test('formatShort returns concise message for 403', () {
      final error = Exception('HTTP 403 Forbidden');
      final formatted = error_fmt.ErrorFormatter.formatShort(error);

      expect(formatted, contains('403'));
      expect(formatted, contains('possible rate limit'));
    });

    test('formatShort truncates long messages', () {
      final longError = Exception('A' * 150);
      final formatted = error_fmt.ErrorFormatter.formatShort(longError);

      expect(formatted.length, lessThanOrEqualTo(100));
      expect(formatted, endsWith('...'));
    });

    test('formatShort handles null', () {
      final formatted = error_fmt.ErrorFormatter.formatShort(null);

      expect(formatted, equals('Unknown error'));
    });
  });
}
