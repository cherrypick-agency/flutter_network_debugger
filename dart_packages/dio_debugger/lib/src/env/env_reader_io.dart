import 'dart:io';

String? readEnvVar(String key) {
  try {
    return Platform.environment[key];
  } catch (_) {
    return null;
  }
}
