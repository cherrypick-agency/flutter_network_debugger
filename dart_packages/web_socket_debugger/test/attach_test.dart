import 'package:test/test.dart';
import 'package:web_socket_debugger/web_socket_debugger.dart';

void main() {
  group('WebSocketDebugger.attach', () {
    test('disabled returns base URL as-is', () {
      final cfg = WebSocketDebugger.attach(
        baseUrl: 'wss://ws.postman-echo.com/raw',
        enabled: false,
      );
      expect(cfg.useForwardOverrides, isFalse);
      expect(cfg.query, isEmpty);
      expect(cfg.connectUrl.toString(), 'wss://ws.postman-echo.com/raw');
    });

    test('forward keeps URL and enables overrides', () {
      final cfg = WebSocketDebugger.attach(
        baseUrl: 'wss://ws.postman-echo.com/raw',
        proxyBaseUrl: 'http://localhost:9091',
        mode: 'forward',
        enabled: true,
      );
      expect(cfg.useForwardOverrides, isTrue);
      expect(cfg.connectUrl.toString(), 'wss://ws.postman-echo.com/raw');
      expect(cfg.query, isEmpty);
      expect(cfg.httpClientFactory, isNotNull);
    });

    test('reverse builds ws:// proxy URL', () {
      final cfg = WebSocketDebugger.attach(
        baseUrl: 'wss://ws.postman-echo.com/raw',
        proxyBaseUrl: 'http://localhost:9091',
        proxyPath: '/wsproxy',
        mode: 'reverse',
        enabled: true,
      );
      expect(cfg.connectUrl.toString(), 'ws://localhost:9091/wsproxy');
      expect(cfg.query.containsKey('_target'), isTrue);
    });
  });
}
