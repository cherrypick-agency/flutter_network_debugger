import 'dart:io';

/// Log levels in order of severity.
enum LogLevel {
  /// Debug messages for detailed troubleshooting.
  debug(0),

  /// Informational messages about normal operations.
  info(1),

  /// Warning messages about potential issues.
  warning(2),

  /// Error messages about failures.
  error(3),

  /// No logging.
  none(4);

  const LogLevel(this.severity);
  final int severity;
}

/// A simple logger with multiple log levels and optional color support.
///
/// Usage:
/// ```dart
/// // Set global log level
/// Logger.level = LogLevel.debug;
///
/// // Create logger for specific component
/// final logger = Logger('BinaryDownloader');
/// logger.debug('Starting download...');
/// logger.info('Downloaded 50%');
/// logger.warning('Slow connection detected');
/// logger.error('Download failed');
/// ```
class Logger {
  /// Global log level - messages below this level won't be shown.
  static LogLevel level = LogLevel.info;

  /// Whether to show timestamps in log messages.
  static bool showTimestamps = false;

  /// Whether to use colored output (automatically disabled on non-TTY).
  static bool useColors = stdout.hasTerminal;

  /// Component name for this logger instance.
  final String component;

  /// Creates a logger for a specific component.
  Logger(this.component);

  /// ANSI color codes for different log levels.
  static const _colors = {
    LogLevel.debug: '\x1B[36m', // Cyan
    LogLevel.info: '\x1B[32m', // Green
    LogLevel.warning: '\x1B[33m', // Yellow
    LogLevel.error: '\x1B[31m', // Red
  };

  static const _reset = '\x1B[0m';

  /// Logs a debug message.
  void debug(String message) => _log(LogLevel.debug, message);

  /// Logs an info message.
  void info(String message) => _log(LogLevel.info, message);

  /// Logs a warning message.
  void warning(String message) => _log(LogLevel.warning, message);

  /// Logs an error message.
  void error(String message, [Object? error, StackTrace? stackTrace]) {
    _log(LogLevel.error, message);
    if (error != null) {
      _log(LogLevel.error, 'Error: $error');
    }
    if (stackTrace != null && level.severity <= LogLevel.debug.severity) {
      _log(LogLevel.debug, 'Stack trace:\n$stackTrace');
    }
  }

  void _log(LogLevel messageLevel, String message) {
    // Check if message should be shown
    if (messageLevel.severity < level.severity) {
      return;
    }

    final buffer = StringBuffer();

    // Add timestamp if enabled
    if (showTimestamps) {
      final now = DateTime.now();
      buffer.write('[${_formatTimestamp(now)}] ');
    }

    // Add color if enabled
    if (useColors && _colors.containsKey(messageLevel)) {
      buffer.write(_colors[messageLevel]);
    }

    // Add log level
    buffer.write('[${messageLevel.name.toUpperCase().padRight(7)}]');

    // Add component name
    buffer.write(' [$component]');

    // Reset color before message
    if (useColors) {
      buffer.write(_reset);
    }

    // Add message
    buffer.write(' $message');

    // Output to appropriate stream
    if (messageLevel == LogLevel.error) {
      stderr.writeln(buffer.toString());
    } else {
      stdout.writeln(buffer.toString());
    }
  }

  String _formatTimestamp(DateTime dt) {
    return '${dt.hour.toString().padLeft(2, '0')}:'
        '${dt.minute.toString().padLeft(2, '0')}:'
        '${dt.second.toString().padLeft(2, '0')}.'
        '${(dt.millisecond ~/ 100).toString()}';
  }

  /// Sets the global log level from a string (case-insensitive).
  ///
  /// Valid values: 'debug', 'info', 'warning', 'error', 'none'
  /// Returns true if successful, false if invalid level name.
  static bool setLevelFromString(String levelName) {
    final normalized = levelName.toLowerCase();
    for (final level in LogLevel.values) {
      if (level.name == normalized) {
        Logger.level = level;
        return true;
      }
    }
    return false;
  }

  /// Gets the current log level as a string.
  static String getLevelString() {
    return level.name;
  }

  /// Creates a logger with verbose settings (debug level + timestamps).
  static void enableVerboseMode() {
    level = LogLevel.debug;
    showTimestamps = true;
  }

  /// Creates a logger with quiet settings (error level only).
  static void enableQuietMode() {
    level = LogLevel.error;
    showTimestamps = false;
  }

  /// Resets to default settings (info level, no timestamps).
  static void resetToDefaults() {
    level = LogLevel.info;
    showTimestamps = false;
    useColors = stdout.hasTerminal;
  }
}
