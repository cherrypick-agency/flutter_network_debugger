/// Stub for reading environment variables on platforms without dart:io.
///
/// On web platforms environment variables are not available, so always
/// returns `null`.
String? readEnvVar(String key) => null;
