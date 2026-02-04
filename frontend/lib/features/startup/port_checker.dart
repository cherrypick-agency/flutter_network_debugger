import 'dart:io';
import 'dart:async';

/// Port check status
class PortStatus {
  final bool apiRunning;
  final bool proxyRunning;
  final int apiPort;
  final int proxyPort;

  PortStatus({
    required this.apiRunning,
    required this.proxyRunning,
    required this.apiPort,
    required this.proxyPort,
  });

  bool get bothRunning => apiRunning && proxyRunning;
  bool get bothStopped => !apiRunning && !proxyRunning;
  bool get partialRunning => !bothRunning && !bothStopped;
}

/// Checks API port availability via health endpoint
Future<bool> checkApiPort(int port) async {
  try {
    final client = HttpClient();
    client.connectionTimeout = const Duration(seconds: 2);

    final request = await client.getUrl(
      Uri.parse('http://localhost:$port/_health'),
    );

    final response = await request.close().timeout(const Duration(seconds: 2));

    await response.drain();
    client.close();

    return response.statusCode == 200;
  } catch (e) {
    return false;
  }
}

/// Checks Proxy port availability via TCP connection
Future<bool> checkProxyPort(int port) async {
  try {
    final socket = await Socket.connect(
      'localhost',
      port,
      timeout: const Duration(seconds: 2),
    );
    await socket.close();
    return true;
  } catch (e) {
    return false;
  }
}

/// Checks status of both ports
Future<PortStatus> checkPorts({
  required int apiPort,
  required int proxyPort,
}) async {
  // Check both ports in parallel for speed
  final results = await Future.wait([
    checkApiPort(apiPort),
    checkProxyPort(proxyPort),
  ]);

  return PortStatus(
    apiRunning: results[0],
    proxyRunning: results[1],
    apiPort: apiPort,
    proxyPort: proxyPort,
  );
}
