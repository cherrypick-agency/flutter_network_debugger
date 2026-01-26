import 'package:test/test.dart';
import 'package:web_socket_debugger/web_socket_debugger.dart';

void main() {
  group('WebSocketDebugger URL contract', () {
    test('reverse: _target keeps ws/wss, proxy url uses ws scheme', () {
      final cfg = WebSocketDebugger.attach(
        baseUrl: 'wss://example.com/raw',
        proxyBaseUrl: 'http://localhost:9091',
        proxyPath: '/wsproxy',
        mode: 'reverse',
        enabled: true,
      );
      expect(cfg.connectUrl.toString(), 'ws://localhost:9091/wsproxy');
      final target = cfg.query['_target'] as String;
      expect(target, 'wss://example.com/raw');
    });

    test('reverse: http/https baseUrl converted to ws/wss target', () {
      final a = WebSocketDebugger.attach(
        baseUrl: 'http://example.com/ws',
        proxyBaseUrl: 'http://localhost:9091',
        proxyPath: '/wsproxy',
        mode: 'reverse',
        enabled: true,
      );
      expect(a.query['_target'], 'ws://example.com/ws');

      final b = WebSocketDebugger.attach(
        baseUrl: 'https://example.com/ws',
        proxyBaseUrl: 'http://localhost:9091',
        proxyPath: '/wsproxy',
        mode: 'reverse',
        enabled: true,
      );
      expect(b.query['_target'], 'wss://example.com/ws');
    });

    test('reverse: proxyPath normalization strips scheme and query', () {
      final cfg = WebSocketDebugger.attach(
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
