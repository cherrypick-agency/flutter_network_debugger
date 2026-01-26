import 'package:socket_io_debugger/socket_io_debugger.dart';
import 'package:test/test.dart';

void main() {
  group('SocketIoDebugger.attach', () {
    test('reverse: builds wsproxy config with engine.io target', () {
      final cfg = SocketIoDebugger.attach(
        baseUrl: 'https://example.com',
        path: '/socket.io/',
        proxyBaseUrl: 'http://localhost:9091',
        proxyHttpPath: '/wsproxy',
        enabled: true,
      );

      expect(cfg.useForwardOverrides, isFalse);
      expect(cfg.effectiveBaseUrl, 'http://localhost:9091');
      expect(cfg.effectivePath, '/wsproxy');
      expect(cfg.query.containsKey('_target'), isTrue);
      expect(
        (cfg.query['_target'] as String).startsWith('https://example.com:443/'),
        isTrue,
      );
      expect(
        (cfg.query['_target'] as String).contains('transport=websocket'),
        isTrue,
      );
    });

    test('reverse: does not leak query into path (no %3F... in path)', () {
      final cfg = SocketIoDebugger.attach(
        baseUrl: 'https://example.com',
        path: '/socket.io/',
        proxyBaseUrl: 'http://localhost:9091',
        proxyHttpPath: '/wsproxy?bad=1',
        enabled: true,
      );

      // proxyHttpPath должен оставаться "путём", без query-хвоста.
      // Если сюда попадает '?', дальше в URL может появиться мусор типа '%3F_target'.
      expect(cfg.effectivePath.contains('?'), isFalse);
    });
  });
}
