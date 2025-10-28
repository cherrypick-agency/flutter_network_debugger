import 'package:test/test.dart';
import 'package:web_socket_channel_debugger/web_socket_channel_debugger.dart';

void main() {
  group('WebSocketChannelDebugger.attach', () {
    test('disabled returns base URL as-is', () {
      final cfg = WebSocketChannelDebugger.attach(
        baseUrl: 'wss://echo.websocket.events',
        enabled: false,
      );
      expect(cfg.useForwardOverrides, isFalse);
      expect(cfg.query, isEmpty);
      expect(cfg.connectUrl.toString(), 'wss://echo.websocket.events');
    });

    test('reverse builds proxy URL with _target', () {
      final cfg = WebSocketChannelDebugger.attach(
        baseUrl: 'wss://echo.websocket.events',
        proxyBaseUrl: 'http://localhost:9091',
        proxyPath: '/wsproxy',
        mode: 'reverse',
        enabled: true,
      );
      expect(cfg.useForwardOverrides, isFalse);
      expect(cfg.connectUrl.toString(), 'ws://localhost:9091/wsproxy');
      expect(cfg.query.containsKey('_target'), isTrue);
      final target = cfg.query['_target'] as String;
      expect(target.startsWith('wss://echo.websocket.events'), isTrue);
    });

    test('forward keeps URL and enables overrides', () {
      final cfg = WebSocketChannelDebugger.attach(
        baseUrl: 'wss://echo.websocket.events',
        proxyBaseUrl: 'http://localhost:9091',
        mode: 'forward',
        enabled: true,
      );
      expect(cfg.useForwardOverrides, isTrue);
      expect(cfg.connectUrl.toString(), 'wss://echo.websocket.events');
      expect(cfg.query, isEmpty);
      expect(cfg.httpClientFactory, isNotNull);
    });
  });
}
