library http_debugger;

// Export platform-specific implementation: on IO - working,
// on Web and others - stub without side effects.
export 'src/overrides_stub.dart' if (dart.library.io) 'src/overrides_io.dart';
export 'src/http_client_wrapper.dart';
