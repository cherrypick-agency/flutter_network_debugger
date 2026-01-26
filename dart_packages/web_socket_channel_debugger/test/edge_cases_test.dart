import 'package:test/test.dart';
import 'package:web_socket_channel_debugger/web_socket_channel_debugger.dart';

void main() {
  group('WebSocketChannelDebugger edge cases', () {
    // Test: enabled: false returns original URL unchanged
    test('enabled: false returns original URL', () {
      final cfg = WebSocketChannelDebugger.attach(
        baseUrl: 'wss://example.com/ws',
        proxyBaseUrl: 'http://localhost:9091',
        enabled: false,
      );
      expect(cfg.connectUrl.toString(), 'wss://example.com/ws');
      expect(cfg.query, isEmpty);
      expect(cfg.useForwardOverrides, isFalse);
    });

    // Test: mode: none works as pass-through
    test('mode: none works as pass-through', () {
      final cfg = WebSocketChannelDebugger.attach(
        baseUrl: 'wss://example.com/ws',
        proxyBaseUrl: 'http://localhost:9091',
        mode: 'none',
        enabled: true,
      );
      expect(cfg.connectUrl.toString(), 'wss://example.com/ws');
      expect(cfg.query, isEmpty);
    });

    // Test: empty proxyBaseUrl in reverse mode returns original
    test('empty proxyBaseUrl returns original URL', () {
      final cfg = WebSocketChannelDebugger.attach(
        baseUrl: 'wss://example.com/ws',
        proxyBaseUrl: '',
        mode: 'reverse',
        enabled: true,
      );
      expect(cfg.connectUrl.toString(), 'wss://example.com/ws');
      expect(cfg.query, isEmpty);
    });

    // Test: baseUrl with query params preserved in _target
    test('baseUrl with query params preserved in _target', () {
      final cfg = WebSocketChannelDebugger.attach(
        baseUrl: 'wss://example.com/ws?token=abc&room=123',
        proxyBaseUrl: 'http://localhost:9091',
        proxyPath: '/wsproxy',
        mode: 'reverse',
        enabled: true,
      );
      final target = cfg.query['_target'] as String;
      expect(target, contains('token=abc'));
      expect(target, contains('room=123'));
    });

    // Test: proxyPath with double slashes gets normalized
    test('proxyPath with double slashes gets normalized', () {
      final cfg = WebSocketChannelDebugger.attach(
        baseUrl: 'wss://example.com/ws',
        proxyBaseUrl: 'http://localhost:9091',
        proxyPath: '//wsproxy//test//',
        mode: 'reverse',
        enabled: true,
      );
      expect(cfg.connectUrl.path, '/wsproxy/test/');
    });

    // Test: https proxy uses wss scheme for connection
    test('https proxy uses wss scheme', () {
      final cfg = WebSocketChannelDebugger.attach(
        baseUrl: 'wss://example.com/ws',
        proxyBaseUrl: 'https://proxy.example.com',
        proxyPath: '/wsproxy',
        mode: 'reverse',
        enabled: true,
      );
      expect(cfg.connectUrl.scheme, 'wss');
      expect(cfg.connectUrl.host, 'proxy.example.com');
    });

    // Test: proxyPath without leading slash gets normalized
    test('proxyPath without leading slash gets normalized', () {
      final cfg = WebSocketChannelDebugger.attach(
        baseUrl: 'wss://example.com/ws',
        proxyBaseUrl: 'http://localhost:9091',
        proxyPath: 'wsproxy',
        mode: 'reverse',
        enabled: true,
      );
      expect(cfg.connectUrl.path, '/wsproxy');
    });

    // Test: baseUrl with path preserved in _target
    test('baseUrl with path preserved in _target', () {
      final cfg = WebSocketChannelDebugger.attach(
        baseUrl: 'wss://example.com/chat/room1',
        proxyBaseUrl: 'http://localhost:9091',
        mode: 'reverse',
        enabled: true,
      );
      final target = cfg.query['_target'] as String;
      expect(target, contains('/chat/room1'));
    });

    // Test: http baseUrl converts to ws _target
    test('http baseUrl converts to ws _target', () {
      final cfg = WebSocketChannelDebugger.attach(
        baseUrl: 'http://example.com/ws',
        proxyBaseUrl: 'http://localhost:9091',
        mode: 'reverse',
        enabled: true,
      );
      final target = cfg.query['_target'] as String;
      expect(target, startsWith('ws://'));
    });

    // Test: https baseUrl converts to wss _target
    test('https baseUrl converts to wss _target', () {
      final cfg = WebSocketChannelDebugger.attach(
        baseUrl: 'https://example.com/ws',
        proxyBaseUrl: 'http://localhost:9091',
        mode: 'reverse',
        enabled: true,
      );
      final target = cfg.query['_target'] as String;
      expect(target, startsWith('wss://'));
    });
  });
}
