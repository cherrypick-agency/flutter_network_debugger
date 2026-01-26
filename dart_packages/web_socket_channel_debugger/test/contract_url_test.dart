import 'package:test/test.dart';
import 'package:web_socket_channel_debugger/web_socket_channel_debugger.dart';

void main() {
  group('WebSocketChannelDebugger URL contract', () {
    test('reverse: _target keeps ws/wss, proxy url uses ws scheme', () {
      final cfg = WebSocketChannelDebugger.attach(
        baseUrl: 'wss://echo.websocket.events',
        proxyBaseUrl: 'http://localhost:9091',
        proxyPath: '/wsproxy',
        mode: 'reverse',
        enabled: true,
      );
      expect(cfg.connectUrl.toString(), 'ws://localhost:9091/wsproxy');
      final target = cfg.query['_target'] as String;
      expect(target.startsWith('wss://echo.websocket.events'), isTrue);
    });

    test('reverse: proxyPath normalization strips scheme and query', () {
      final cfg = WebSocketChannelDebugger.attach(
        baseUrl: 'ws://example.com/ws',
        proxyBaseUrl: 'http://localhost:9091',
        proxyPath: 'http://localhost:9091/wsproxy?bad=1',
        mode: 'reverse',
        enabled: true,
      );
      expect(cfg.connectUrl.path, '/wsproxy');
      expect(cfg.connectUrl.toString(), 'ws://localhost:9091/wsproxy');
    });
  });
}
