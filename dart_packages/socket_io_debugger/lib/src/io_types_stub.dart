/// Stub types for platforms without `dart:io`.
///
/// This package supports Web builds. On Web there is no `dart:io`, but we still
/// want the public API to compile. Forward-proxy mode is IO-only anyway.
typedef HttpClient = Object;
