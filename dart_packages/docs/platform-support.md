# Platform Support

This guide covers platform-specific behavior and limitations of the Network Debugger packages.

## Platform Compatibility Matrix

| Feature | Android | iOS | Web | macOS | Windows | Linux |
| ------- | ------- | --- | --- | ----- | ------- | ----- |
| Reverse proxy mode | Yes | Yes | Yes | Yes | Yes | Yes |
| Forward proxy mode | Yes | Yes | No | Yes | Yes | Yes |
| Custom HTTP headers | Yes | Yes | No | Yes | Yes | Yes |
| Environment variables | Yes | Yes | No | Yes | Yes | Yes |
| Self-signed certificates | Yes | Yes | No | Yes | Yes | Yes |
| `--dart-define` | Yes | Yes | Yes | Yes | Yes | Yes |

## Web Platform

### Supported Features

- Reverse proxy mode
- `--dart-define` configuration
- Basic WebSocket connections

### Limitations

1. **No Forward Proxy Mode**
   - Web browsers don't support `HttpOverrides`
   - Use reverse proxy mode instead

2. **No Custom HTTP Headers for WebSocket**
   - Browser WebSocket API doesn't allow custom headers
   - Authentication must use query parameters or cookies

3. **No Environment Variables**
   - Use `--dart-define` instead:
   ```bash
   flutter build web --dart-define=SOCKET_PROXY=http://localhost:9091
   ```

4. **CORS Restrictions**
   - Proxy server must send appropriate CORS headers
   - Default proxy server includes CORS headers

### Web Configuration Example

```dart
// Use dart-define for Web
const proxyUrl = String.fromEnvironment(
  'SOCKET_PROXY',
  defaultValue: 'http://localhost:9091',
);

WebSocketChannelDebugger.attach(
  baseUrl: 'wss://example.com/ws',
  proxyBaseUrl: proxyUrl,
  mode: 'reverse',  // Only mode available on Web
);
```

## Android

### Special Considerations

1. **Emulator Localhost**
   - Android emulator cannot reach `localhost` on host
   - Use `10.0.2.2` instead:
   ```dart
   final proxyUrl = Platform.isAndroid 
     ? 'http://10.0.2.2:9091' 
     : 'http://localhost:9091';
   ```

2. **Cleartext Traffic (HTTP)**
   - Android 9+ blocks cleartext (HTTP) by default
   - Add to `AndroidManifest.xml`:
   ```xml
   <application
       android:usesCleartextTraffic="true"
       ...>
   ```
   - Or use HTTPS for proxy server

3. **Network Security Config**
   - For self-signed certificates, create `network_security_config.xml`:
   ```xml
   <?xml version="1.0" encoding="utf-8"?>
   <network-security-config>
       <domain-config cleartextTrafficPermitted="true">
           <domain includeSubdomains="true">10.0.2.2</domain>
       </domain-config>
   </network-security-config>
   ```

### Android Emulator Setup

```dart
import 'dart:io' show Platform;

String getProxyUrl() {
  if (Platform.isAndroid) {
    // Check if running in emulator
    return 'http://10.0.2.2:9091';
  }
  return 'http://localhost:9091';
}
```

## iOS

### Supported Features

All features work on iOS including:
- Both proxy modes
- Custom headers
- Environment variables
- Self-signed certificates (with flag)

### iOS Simulator

iOS Simulator can access `localhost` directly:

```dart
final proxyUrl = 'http://localhost:9091';
```

### Physical Device

For physical iOS devices, the proxy server must be accessible on the network:

```dart
// Use your machine's IP address
final proxyUrl = 'http://192.168.1.100:9091';
```

Start proxy server with network binding:

```bash
ADDR=0.0.0.0:9091 network_debugger
```

## macOS / Windows / Linux

### Desktop Platforms

All features work on desktop platforms:
- Both proxy modes
- Custom headers
- Environment variables
- Self-signed certificates

### Configuration

```dart
// Desktop platforms can use localhost directly
final proxyUrl = 'http://localhost:9091';

DioDebugger.interceptor(
  proxyBaseUrl: proxyUrl,
  allowBadCertificates: true,  // For self-signed certs
);
```

## Cross-Platform Configuration

### Recommended Approach

Create a helper function that handles platform differences:

```dart
import 'dart:io' show Platform;
import 'package:flutter/foundation.dart' show kIsWeb;

class ProxyConfig {
  static String get proxyUrl {
    if (kIsWeb) {
      return const String.fromEnvironment(
        'SOCKET_PROXY',
        defaultValue: 'http://localhost:9091',
      );
    }
    
    if (Platform.isAndroid) {
      return 'http://10.0.2.2:9091';
    }
    
    return 'http://localhost:9091';
  }
  
  static bool get isSupported {
    // Forward proxy not supported on Web
    if (kIsWeb) return true;  // Only reverse mode
    return true;
  }
}
```

### Usage

```dart
DioDebugger.interceptor(
  proxyBaseUrl: ProxyConfig.proxyUrl,
  mode: 'reverse',  // Works on all platforms
);
```

## Feature Detection

Check platform capabilities at runtime:

```dart
import 'package:flutter/foundation.dart' show kIsWeb;

void setupDebugger() {
  if (kIsWeb) {
    // Web: only reverse mode, no custom headers
    WebSocketChannelDebugger.attach(
      baseUrl: 'wss://example.com/ws',
      proxyBaseUrl: 'http://localhost:9091',
      mode: 'reverse',
    );
  } else {
    // Native: all features available
    final config = WebSocketChannelDebugger.attach(
      baseUrl: 'wss://example.com/ws',
      proxyBaseUrl: 'http://localhost:9091',
      mode: 'forward',
    );
    
    if (config.useForwardOverrides) {
      // Setup HttpOverrides...
    }
  }
}
```

## Next Steps

- [Configuration Guide](./configuration.md) - Platform-specific configuration
- [Troubleshooting](./troubleshooting.md) - Platform-specific issues
- [API Reference](./api-reference.md) - Complete API documentation
