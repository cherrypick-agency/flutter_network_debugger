import 'dart:io';

/// Reads environment variable value.
///
/// Returns null if the variable is unset or empty.
String? readEnvVar(String key) {
  try {
    final v = Platform.environment[key];
    if (v == null || v.trim().isEmpty) return null;
    return v;
  } catch (_) {
    return null;
  }
}
