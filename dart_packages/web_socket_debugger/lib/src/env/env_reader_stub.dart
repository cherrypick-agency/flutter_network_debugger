/// Stub for reading environment variables in web environments.
///
/// In environments without dart:io, environment variables are unavailable,
/// so always returns `null`.
///
/// [key] - environment variable name (ignored).
///
/// Always returns `null`.
String? readEnvVar(String key) => null;
