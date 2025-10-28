import 'dart:io';

T runWithHttpOverrides<T>(
    HttpClient Function() clientFactory, T Function() action) {
  return HttpOverrides.runZoned(
    action,
    createHttpClient: (SecurityContext? _) => clientFactory(),
  );
}
