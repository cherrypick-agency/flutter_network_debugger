import 'dart:io';

/// Runs a function with overridden HTTP client.
///
/// Used for configuring proxy via HttpOverrides.
T runWithHttpOverrides<T>(
    HttpClient Function() clientFactory, T Function() action) {
  return HttpOverrides.runZoned(
    action,
    createHttpClient: (SecurityContext? _) => clientFactory(),
  );
}
