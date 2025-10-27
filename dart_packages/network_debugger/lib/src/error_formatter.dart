import 'dart:async';
import 'dart:io';

/// Formats exceptions into user-friendly error messages.
///
/// This class provides static methods to convert technical exceptions
/// into clear, actionable error messages that help users understand
/// what went wrong and how to fix it.
class ErrorFormatter {
  /// Formats any exception into a user-friendly message.
  ///
  /// This is the main entry point that delegates to specific formatters
  /// based on exception type. Falls back to generic message if no
  /// specific formatter matches.
  ///
  /// Example:
  /// ```dart
  /// try {
  ///   await downloadFile(url);
  /// } catch (e) {
  ///   print(ErrorFormatter.format(e));
  /// }
  /// ```
  static String format(dynamic error) {
    if (error == null) {
      return 'Unknown error occurred';
    }

    final errorString = error.toString();

    // Network errors
    if (error is SocketException || errorString.contains('SocketException')) {
      return formatNetworkError(error);
    }
    if (error is HttpException || errorString.contains('HttpException') ||
        errorString.contains('HTTP 404') || errorString.contains('HTTP 403') ||
        errorString.contains('HTTP 5')) {
      return formatHttpError(error);
    }
    if (error is TimeoutException || errorString.contains('TimeoutException')) {
      return formatTimeoutError(error);
    }

    // Platform errors
    if (error is UnsupportedError || errorString.contains('Unsupported')) {
      return formatPlatformError(error);
    }

    // Permission errors
    if (errorString.contains('Permission denied') ||
        errorString.contains('Access denied')) {
      return formatPermissionError(error);
    }

    // Checksum errors
    if (errorString.contains('ChecksumValidationException') ||
        errorString.contains('checksum')) {
      return formatChecksumError(error);
    }

    // GitHub API errors
    if (errorString.contains('GitHubReleaseException')) {
      return formatGitHubError(error);
    }

    // Archive extraction errors
    if (errorString.contains('Archive') || errorString.contains('extract')) {
      return formatArchiveError(error);
    }

    // Process errors
    if (errorString.contains('ProcessException')) {
      return formatProcessError(error);
    }

    // Cache errors
    if (errorString.contains('CacheException')) {
      return formatCacheError(error);
    }

    // Retry exhausted
    if (errorString.contains('RetryExhaustedException')) {
      return formatRetryExhaustedError(error);
    }

    // Generic fallback
    return formatGenericError(error);
  }

  /// Formats network connection errors.
  static String formatNetworkError(dynamic error) {
    return '''
Network connection error: Unable to connect to the internet.

Possible solutions:
  • Check your internet connection
  • Verify firewall settings aren't blocking the connection
  • Try again in a few moments
  • If behind a proxy, ensure proxy settings are configured

Technical details: $error
''';
  }

  /// Formats HTTP errors (non-200 status codes).
  static String formatHttpError(dynamic error) {
    final errorString = error.toString();

    if (errorString.contains('404') || errorString.contains('Not Found')) {
      return '''
Resource not found (HTTP 404).

This usually means:
  • The binary for your platform is not available in this release
  • The release version you specified doesn't exist
  • The GitHub repository structure has changed

Try:
  • Omit the --binary-version flag to use the latest version
  • Check available releases at: github.com/cherrypick-agency/flutter_network_debugger/releases

Technical details: $error
''';
    }

    if (errorString.contains('403') || errorString.contains('Forbidden')) {
      return '''
Access forbidden (HTTP 403).

This might be caused by:
  • GitHub API rate limiting (60 requests/hour for unauthenticated users)
  • Geographic restrictions
  • Repository access permissions

Try:
  • Wait an hour and try again
  • Set GITHUB_TOKEN environment variable to increase rate limit

Technical details: $error
''';
    }

    if (errorString.contains('500') || errorString.contains('502') ||
        errorString.contains('503') || errorString.contains('504')) {
      return '''
Server error (HTTP 5xx).

The remote server is experiencing issues. This is temporary.

Try:
  • Wait a few minutes and try again
  • Check GitHub status at: githubstatus.com

Technical details: $error
''';
    }

    return '''
HTTP error occurred while downloading.

Technical details: $error
''';
  }

  /// Formats timeout errors.
  static String formatTimeoutError(dynamic error) {
    return '''
Operation timed out.

The operation took too long to complete. This can happen with:
  • Slow internet connection
  • Large file downloads
  • Server not responding

Try:
  • Check your internet speed
  • Try again when you have a better connection
  • If the problem persists, the server might be down

Technical details: $error
''';
  }

  /// Formats platform/architecture errors.
  static String formatPlatformError(dynamic error) {
    final errorString = error.toString();

    return '''
Platform not supported.

Your platform/architecture is not currently supported.

Current platform: $errorString

Supported platforms:
  • macOS (Intel/ARM)
  • Linux (amd64/arm64)
  • Windows (32-bit/64-bit/ARM)

If you believe this is an error, please report it at:
https://github.com/cherrypick-agency/flutter_network_debugger/issues

Technical details: $error
''';
  }

  /// Formats permission/access denied errors.
  static String formatPermissionError(dynamic error) {
    return '''
Permission denied.

The application doesn't have sufficient permissions to perform this operation.

Common causes:
  • Cannot write to cache directory
  • Cannot execute the downloaded binary
  • Directory/file is read-only

Try:
  • Check permissions on cache directory
  • On Unix systems: chmod +x on the binary if needed
  • Run with appropriate privileges
  • Check antivirus/security software isn't blocking execution

Cache location varies by platform:
  • macOS/Linux: ~/.cache/network_debugger/
  • Windows: %LOCALAPPDATA%\\network_debugger\\Cache\\

Technical details: $error
''';
  }

  /// Formats checksum validation errors.
  static String formatChecksumError(dynamic error) {
    return '''
Security validation failed: Checksum mismatch.

The downloaded file's checksum doesn't match the expected value.
This could indicate:
  • File corruption during download
  • Security issue (file was tampered with)
  • Network transmission error

For security reasons, the file will not be used.

What to do:
  1. Clear the cache: network_debugger --clear-cache (if implemented)
  2. Try downloading again
  3. If problem persists, report at:
     https://github.com/cherrypick-agency/flutter_network_debugger/issues

To skip validation (NOT RECOMMENDED):
  • Use skipChecksumValidation parameter (API only)

Technical details: $error
''';
  }

  /// Formats GitHub API errors.
  static String formatGitHubError(dynamic error) {
    return '''
GitHub API error.

Failed to fetch release information from GitHub.

Common causes:
  • Release not found
  • API rate limiting
  • Repository moved or renamed
  • Temporary GitHub outage

Try:
  • Check your internet connection
  • Verify the release exists
  • Wait if rate limited (resets hourly)
  • Check GitHub status: githubstatus.com

Technical details: $error
''';
  }

  /// Formats archive extraction errors.
  static String formatArchiveError(dynamic error) {
    return '''
Failed to extract archive.

The downloaded archive could not be extracted.

Possible causes:
  • Corrupted download
  • Unsupported archive format
  • Disk space full
  • Permission issues

Try:
  • Clear cache and download again
  • Check available disk space
  • Verify you have write permissions

Technical details: $error
''';
  }

  /// Formats process execution errors.
  static String formatProcessError(dynamic error) {
    final errorString = error.toString();

    if (errorString.contains('port') || errorString.contains('ADDR')) {
      return '''
Failed to start debugger: Port already in use.

The specified port is already occupied by another process.

Try:
  • Use a different port: network_debugger --port 8080
  • Find and stop the process using the port
  • On Unix: lsof -i :9091
  • On Windows: netstat -ano | findstr :9091

Technical details: $error
''';
    }

    if (errorString.contains('not found') || errorString.contains('binary')) {
      return '''
Failed to start debugger: Binary not found or not executable.

The debugger binary could not be started.

Try:
  • Clear cache and download again
  • On Unix systems: ensure binary has execute permissions
  • Check antivirus isn't blocking the file

Technical details: $error
''';
    }

    return '''
Failed to start debugger process.

The debugger process could not be started.

Try:
  • Clear cache: rm -rf ~/.cache/network_debugger/ (macOS/Linux)
  • Download the binary again
  • Check system resources (RAM, disk space)
  • Review security software settings

Technical details: $error
''';
  }

  /// Formats cache-related errors.
  static String formatCacheError(dynamic error) {
    return '''
Cache operation failed.

An error occurred while accessing the cache directory.

Common causes:
  • Insufficient disk space
  • Permission issues
  • Corrupted cache files

Try:
  • Check available disk space
  • Verify write permissions
  • Clear cache manually and try again:
    macOS/Linux: rm -rf ~/.cache/network_debugger/
    Windows: del /s /q %LOCALAPPDATA%\\network_debugger\\Cache\\

Technical details: $error
''';
  }

  /// Formats retry exhausted errors.
  static String formatRetryExhaustedError(dynamic error) {
    return '''
Operation failed after multiple retry attempts.

The operation was retried several times but continued to fail.

This suggests a persistent problem:
  • Network connectivity issues
  • Server is down
  • Resource unavailable

Try:
  • Check your internet connection
  • Verify the resource exists
  • Try again later
  • Check service status

Technical details: $error
''';
  }

  /// Formats generic/unknown errors.
  static String formatGenericError(dynamic error) {
    return '''
An unexpected error occurred.

Error: $error

If this problem persists, please report it at:
https://github.com/cherrypick-agency/flutter_network_debugger/issues

Include:
  • Your operating system and version
  • Command you ran
  • Full error message above
  • Whether --verbose flag shows more details
''';
  }

  /// Extracts a short, one-line error message.
  ///
  /// Useful for logging or displaying errors inline.
  static String formatShort(dynamic error) {
    if (error == null) return 'Unknown error';

    final errorString = error.toString();

    if (error is SocketException || errorString.contains('SocketException')) {
      return 'Network connection error';
    }
    if (error is TimeoutException) {
      return 'Operation timed out';
    }
    if (errorString.contains('404')) {
      return 'Resource not found (HTTP 404)';
    }
    if (errorString.contains('403')) {
      return 'Access forbidden - possible rate limit (HTTP 403)';
    }
    if (errorString.contains('ChecksumValidationException')) {
      return 'Checksum validation failed - file may be corrupted';
    }
    if (errorString.contains('Permission denied')) {
      return 'Permission denied';
    }
    if (errorString.contains('UnsupportedError')) {
      return 'Platform not supported';
    }

    // Truncate long messages
    if (errorString.length > 100) {
      return '${errorString.substring(0, 97)}...';
    }

    return errorString;
  }
}
