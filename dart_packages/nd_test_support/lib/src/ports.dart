import 'dart:io';

/// Picks a free port on localhost.
///
/// Creates a temporary server, gets its port, and closes the server.
Future<int> pickFreePort() async {
  final s = await ServerSocket.bind(InternetAddress.loopbackIPv4, 0);
  final p = s.port;
  await s.close();
  return p;
}
