import 'dart:io';

/// Reads the environment variable value for the given key.
///
/// Returns the value of environment variable [key] or `null` if the variable
/// is not set or an error occurred while reading.
String? readEnvVar(String key) {
  try {
    return Platform.environment[key];
  } catch (_) {
    return null;
  }
}
