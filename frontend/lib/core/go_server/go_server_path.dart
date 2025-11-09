/// Platform-agnostic interface for getting Go server path
/// Uses conditional imports to select the right implementation
library;

import 'go_server_path_stub.dart' if (dart.library.io) 'go_server_path_io.dart';

/// Returns the path to the Go server binary for the current platform
/// Returns null if the binary is not found
Future<String?> getGoServerPath() async {
  return getGoServerPathImpl();
}
