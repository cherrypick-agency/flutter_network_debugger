import 'package:socket_io_debugger/socket_io_debugger.dart';
import 'package:test/test.dart';

void main() {
  group('SocketIoDebugger edge cases', () {
    // Test: enabled: false returns original URL unchanged
    test('enabled: false returns original URL', () {
      final cfg = SocketIoDebugger.attach(
        baseUrl: 'https://example.com',
        path: '/socket.io/',
        proxyBaseUrl: 'http://localhost:9091',
        enabled: false,
      );
      expect(cfg.effectiveBaseUrl, 'https://example.com');
      expect(cfg.effectivePath, '/socket.io/');
      expect(cfg.query, isEmpty);
      expect(cfg.useForwardOverrides, isFalse);
    });

    // Test: _target contains explicit port (fix for socket_io_client bug)
    test('_target contains explicit port 443 for https', () {
      final cfg = SocketIoDebugger.attach(
        baseUrl: 'https://example.com',
        path: '/socket.io/',
        proxyBaseUrl: 'http://localhost:9091',
        proxyHttpPath: '/wsproxy',
        enabled: true,
      );
      final target = cfg.query['_target'] as String;
      expect(target, contains('example.com:443'));
    });

    // Test: _target contains explicit port 80 for http
    test('_target contains explicit port 80 for http', () {
      final cfg = SocketIoDebugger.attach(
        baseUrl: 'http://example.com',
        path: '/socket.io/',
        proxyBaseUrl: 'http://localhost:9091',
        proxyHttpPath: '/wsproxy',
        enabled: true,
      );
      final target = cfg.query['_target'] as String;
      expect(target, contains('example.com:80'));
    });

    // Test: custom port in baseUrl preserved
    test('custom port in baseUrl preserved', () {
      final cfg = SocketIoDebugger.attach(
        baseUrl: 'https://example.com:8443',
        path: '/socket.io/',
        proxyBaseUrl: 'http://localhost:9091',
        proxyHttpPath: '/wsproxy',
        enabled: true,
      );
      final target = cfg.query['_target'] as String;
      expect(target, contains('example.com:8443'));
    });

    // Test: namespace in baseUrl preserved in effectiveBaseUrl
    test('namespace in baseUrl preserved', () {
      final cfg = SocketIoDebugger.attach(
        baseUrl: 'https://example.com/chat',
        path: '/socket.io/',
        proxyBaseUrl: 'http://localhost:9091',
        proxyHttpPath: '/wsproxy',
        enabled: true,
      );
      expect(cfg.effectiveBaseUrl, contains('/chat'));
    });

    // Test: empty proxyBaseUrl returns original
    test('empty proxyBaseUrl returns original URL', () {
      final cfg = SocketIoDebugger.attach(
        baseUrl: 'https://example.com',
        path: '/socket.io/',
        proxyBaseUrl: '',
        enabled: true,
      );
      expect(cfg.effectiveBaseUrl, 'https://example.com');
      expect(cfg.query, isEmpty);
    });

    // Test: proxyHttpPath with double slashes gets normalized
    test('proxyHttpPath with double slashes gets normalized', () {
      final cfg = SocketIoDebugger.attach(
        baseUrl: 'https://example.com',
        path: '/socket.io/',
        proxyBaseUrl: 'http://localhost:9091',
        proxyHttpPath: '//wsproxy//test//',
        enabled: true,
      );
      expect(cfg.effectivePath, '/wsproxy/test/');
    });

    // Test: proxyHttpPath without leading slash gets normalized
    test('proxyHttpPath without leading slash gets normalized', () {
      final cfg = SocketIoDebugger.attach(
        baseUrl: 'https://example.com',
        path: '/socket.io/',
        proxyBaseUrl: 'http://localhost:9091',
        proxyHttpPath: 'wsproxy',
        enabled: true,
      );
      expect(cfg.effectivePath, '/wsproxy');
    });

    // Test: proxyHttpPath with query param gets trimmed
    test('proxyHttpPath with query param gets trimmed', () {
      final cfg = SocketIoDebugger.attach(
        baseUrl: 'https://example.com',
        path: '/socket.io/',
        proxyBaseUrl: 'http://localhost:9091',
        proxyHttpPath: '/wsproxy?bad=1',
        enabled: true,
      );
      expect(cfg.effectivePath, '/wsproxy');
      expect(cfg.effectivePath.contains('?'), isFalse);
    });

    // Test: _target contains EIO=4&transport=websocket
    test('_target contains Engine.IO parameters', () {
      final cfg = SocketIoDebugger.attach(
        baseUrl: 'https://example.com',
        path: '/socket.io/',
        proxyBaseUrl: 'http://localhost:9091',
        proxyHttpPath: '/wsproxy',
        enabled: true,
      );
      final target = cfg.query['_target'] as String;
      expect(target, contains('EIO=4'));
      expect(target, contains('transport=websocket'));
    });

    // Test: custom path passed to _target
    test('custom path passed to _target', () {
      final cfg = SocketIoDebugger.attach(
        baseUrl: 'https://example.com',
        path: '/my-custom-io/',
        proxyBaseUrl: 'http://localhost:9091',
        proxyHttpPath: '/wsproxy',
        enabled: true,
      );
      final target = cfg.query['_target'] as String;
      expect(target, contains('/my-custom-io/'));
    });

    // Test: ws/wss scheme in baseUrl converts to http/https for _target
    test('ws/wss scheme converts to http/https for _target', () {
      final cfgWs = SocketIoDebugger.attach(
        baseUrl: 'ws://example.com',
        path: '/socket.io/',
        proxyBaseUrl: 'http://localhost:9091',
        proxyHttpPath: '/wsproxy',
        enabled: true,
      );
      final targetWs = cfgWs.query['_target'] as String;
      expect(targetWs, startsWith('http://'));

      final cfgWss = SocketIoDebugger.attach(
        baseUrl: 'wss://example.com',
        path: '/socket.io/',
        proxyBaseUrl: 'http://localhost:9091',
        proxyHttpPath: '/wsproxy',
        enabled: true,
      );
      final targetWss = cfgWss.query['_target'] as String;
      expect(targetWss, startsWith('https://'));
    });
  });
}
