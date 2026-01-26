# Troubleshooting Guide

Common issues and solutions when using Network Debugger packages.

## Connection Issues

### Proxy Server Not Reachable

**Symptoms:**
- Connection timeout
- "Connection refused" errors
- Requests fail silently

**Solutions:**

1. **Verify proxy server is running:**
   ```bash
   curl http://localhost:9091/healthz
   # Should return: {"status":"ok"}
   ```

2. **Check port availability:**
   ```bash
   lsof -i :9091
   ```

3. **For Android Emulator, use correct IP:**
   ```dart
   // Wrong
   final proxyUrl = 'http://localhost:9091';
   
   // Correct
   final proxyUrl = 'http://10.0.2.2:9091';
   ```

4. **For physical devices, use network IP:**
   ```dart
   final proxyUrl = 'http://192.168.1.100:9091';
   ```

### Android Cleartext Traffic Blocked

**Symptoms:**
- `CLEARTEXT communication not permitted`
- Requests fail on Android 9+

**Solution:**

Add to `android/app/src/main/AndroidManifest.xml`:
```xml
<application
    android:usesCleartextTraffic="true"
    ...>
```

### SSL/TLS Certificate Errors

**Symptoms:**
- `CERTIFICATE_VERIFY_FAILED`
- `HandshakeException`

**Solutions:**

1. **Enable bad certificate bypass (development only):**
   ```dart
   DioDebugger.interceptor(
     proxyBaseUrl: 'http://localhost:9091',
     allowBadCertificates: true,
   );
   ```

2. **Or use HTTPS with valid certificate for proxy server**

## WebSocket Issues

### WebSocket Connection Fails

**Symptoms:**
- WebSocket upgrade fails
- Connection closes immediately
- `WebSocketException`

**Solutions:**

1. **Verify WebSocket endpoint:**
   ```bash
   # Test with wscat
   wscat -c "ws://localhost:9091/wsproxy?_target=wss://echo.websocket.events"
   ```

2. **Check scheme conversion:**
   - `http://` proxy -> `ws://` WebSocket
   - `https://` proxy -> `wss://` WebSocket

3. **Verify target URL encoding:**
   ```dart
   // The _target must be properly URL-encoded
   final config = WebSocketDebugger.attach(
     baseUrl: 'wss://example.com/ws?token=abc',  // Query params OK
     proxyBaseUrl: 'http://localhost:9091',
   );
   ```

### Socket.IO Connection Issues

**Symptoms:**
- "Incompatible server version" error
- Connection timeout
- Handshake fails

**Solutions:**

1. **Check Socket.IO version compatibility:**
   | socket_io_debugger | socket_io_client | Server Version |
   | ------------------ | ---------------- | -------------- |
   | 1.0.0+ | ^3.1.4 | v4.7+ |
   | ^0.1.0 | ^2.0.3 | v2.*/v3.*/v4.6 |

2. **Verify Engine.IO parameters:**
   ```dart
   final config = SocketIoDebugger.attach(...);
   print(config.query['_target']);
   // Should contain: EIO=4&transport=websocket
   ```

3. **Check explicit port in target URL:**
   ```dart
   // socket_io_client bug: needs explicit port
   // Target should be: https://example.com:443/socket.io/...
   // Not: https://example.com/socket.io/...
   ```

## Configuration Issues

### Environment Variables Not Working

**Symptoms:**
- ENV variables ignored
- Default values used instead

**Solutions:**

1. **Web platform doesn't support ENV:**
   ```bash
   # Use --dart-define instead
   flutter build web --dart-define=SOCKET_PROXY=http://localhost:9091
   ```

2. **Verify ENV is exported:**
   ```bash
   # Wrong
   SOCKET_PROXY=http://localhost:9091 flutter run
   
   # Correct
   export SOCKET_PROXY=http://localhost:9091
   flutter run
   ```

### Proxy Disabled Unexpectedly

**Symptoms:**
- Traffic not going through proxy
- Sessions not appearing in UI

**Solutions:**

1. **Check enabled flag:**
   ```dart
   DioDebugger.interceptor(
     proxyBaseUrl: 'http://localhost:9091',
     enabled: true,  // Explicitly enable
   );
   ```

2. **Check ENV/dart-define:**
   ```bash
   # This disables the proxy
   --dart-define=SOCKET_PROXY_ENABLED=false
   ```

3. **Check mode:**
   ```dart
   // mode: 'none' disables proxy
   mode: 'reverse',  // Use reverse or forward
   ```

## Traffic Not Appearing in UI

### Sessions Not Recording

**Symptoms:**
- Requests succeed but don't appear in Web UI
- Empty session list

**Solutions:**

1. **Verify proxy is intercepting:**
   - Check proxy server logs
   - Add debug logging:
   ```dart
   // Debug output shows proxy URL
   print('Proxy URL: ${config.connectUrl}');
   ```

2. **Check bypass rules:**
   ```dart
   // These paths are skipped
   DioDebugger.interceptor(
     skipPaths: ['/health'],  // Remove if testing /health
     skipHosts: ['localhost'],
   );
   ```

3. **Refresh Web UI or check filter settings**

### Wrong Session Type

**Symptoms:**
- HTTP requests showing as WebSocket
- Sessions miscategorized

**Solution:**

Use correct proxy endpoint:
- HTTP: `/httpproxy`
- WebSocket: `/wsproxy`

```dart
// HTTP
DioDebugger.interceptor(
  proxyHttpPath: '/httpproxy',  // Not /wsproxy
);

// WebSocket
WebSocketDebugger.attach(
  proxyPath: '/wsproxy',  // Not /httpproxy
);
```

## Performance Issues

### Slow Requests

**Symptoms:**
- Requests slower than direct connection
- High latency

**Solutions:**

1. **Check proxy server resources**
2. **Disable proxy in production:**
   ```dart
   enabled: kDebugMode,
   ```

3. **Use skip rules for high-frequency requests:**
   ```dart
   DioDebugger.interceptor(
     skipPaths: ['/metrics', '/health'],
   );
   ```

### Memory Issues

**Symptoms:**
- High memory usage
- App slowdown over time

**Solutions:**

1. **Clear sessions periodically:**
   ```bash
   curl -X DELETE http://localhost:9092/_api/v1/sessions
   ```

2. **Limit session retention on proxy server**

## Common Error Messages

### "connect error: It seems you are trying to reach a Socket.IO server in v2.x with a v3.x client"

**Cause:** Socket.IO version mismatch

**Solution:** Use compatible versions (see Socket.IO section above)

### "URI port must be non-negative"

**Cause:** socket_io_client bug with missing port

**Solution:** Update to socket_io_debugger 1.0.0+ which includes fix

### "HttpOverrides not available"

**Cause:** Trying to use forward mode on Web

**Solution:** Use reverse mode on Web platform

## Getting Help

If you're still experiencing issues:

1. **Check proxy server logs:**
   ```bash
   network_debugger  # Logs show incoming requests
   ```

2. **Enable debug logging in packages**

3. **Check GitHub Issues:** [github.com/your-repo/issues](https://github.com/your-repo/issues)

4. **Provide:**
   - Platform (iOS/Android/Web/Desktop)
   - Package version
   - Proxy server version
   - Error message/stack trace
   - Minimal reproduction code
