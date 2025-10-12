import 'dart:io';

String? readEnvVar(String key) {
  try {
    final v = Platform.environment[key];
    if (v == null || v.trim().isEmpty) return null;
    return v;
  } catch (_) {
    return null;
  }
}
