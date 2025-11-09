import 'dart:io';
import 'dart:async';

/// Статус проверки портов
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

/// Проверяет доступность API порта через health endpoint
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

/// Проверяет доступность Proxy порта через TCP подключение
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

/// Проверяет статус обоих портов
Future<PortStatus> checkPorts({
  required int apiPort,
  required int proxyPort,
}) async {
  // Проверяем оба порта параллельно для скорости
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
