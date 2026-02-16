import 'dart:io';

/// Reads environment variable value in IO environment.
///
/// Gets the value by key. Returns `null` if not found or empty.
/// Read errors are handled and also return `null`.
///
/// [key] - environment variable name.
///
/// Returns the value or `null` if unset or empty.
String? readEnvVar(String key) {
  try {
    final v = Platform.environment[key];
    if (v == null || v.trim().isEmpty) return null;
    return v;
  } catch (_) {
    return null;
  }
}
