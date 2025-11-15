// Platform-agnostic file operations facade
// Uses conditional exports to select the correct implementation:
// - Web: dart:html (file_operations_web.dart)
// - Non-web: Stubs (file_operations_stub.dart)

export 'file_operations_web.dart'
    if (dart.library.io) 'file_operations_stub.dart';
