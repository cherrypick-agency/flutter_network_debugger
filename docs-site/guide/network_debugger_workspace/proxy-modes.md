# Proxy Modes

Network Debugger supports two proxy modes: **Reverse Proxy** and **Forward Proxy**. Understanding the difference helps you choose the right mode for your use case.

## Overview

| Mode | How it Works | Platform Support | Use Case |
| ---- | ------------ | ---------------- | -------- |
| Reverse | Client connects to proxy, proxy forwards to upstream | All platforms | Default, recommended |
| Forward | Client connects to upstream, traffic routes through proxy | dart:io only | Special cases |

## Reverse Proxy Mode (Default)

In reverse proxy mode, your app connects to the proxy server, and the proxy forwards requests to the actual upstream server.

### How It Works

```
┌──────────┐    ┌─────────────────────────────────────────────────┐    ┌──────────┐
│          │    │                   Proxy Server                  │    │          │
│   App    │───>│  /httpproxy?_target=https://api.example.com/... │───>│ Upstream │
│          │    │                                                 │    │          │
└──────────┘    └─────────────────────────────────────────────────┘    └──────────┘
```

### Request Flow

1. App makes request to `http://proxy:9091/httpproxy?_target=https://api.example.com/users`
2. Proxy extracts target URL from `_target` parameter
3. Proxy forwards request to `https://api.example.com/users`
4. Proxy returns response to app

### Configuration

```dart
// HTTP (Dio)
DioDebugger.interceptor(
  proxyBaseUrl: 'http://localhost:9091',
  mode: 'reverse',  // Default
);

// WebSocket
WebSocketDebugger.attach(
  baseUrl: 'wss://example.com/ws',
  proxyBaseUrl: 'http://localhost:9091',
  mode: 'reverse',  // Default
);
```

### Advantages

- Works on all platforms (iOS, Android, Web, Desktop)
- No system-level configuration needed
- Simple URL rewriting
- Reliable interception

### Limitations

- Requires URL modification (handled automatically by packages)
- Target URL visible in query parameter

## Forward Proxy Mode

In forward proxy mode, your app connects directly to the upstream server, but traffic is routed through the proxy at the system/HTTP client level.

### How It Works

```
┌──────────┐    ┌─────────────────┐    ┌──────────┐
│          │    │                 │    │          │
│   App    │───>│  Proxy Server   │───>│ Upstream │
│          │    │  (transparent)  │    │          │
└──────────┘    └─────────────────┘    └──────────┘
     │
     └── Uses HttpOverrides.global to route traffic
```

### Request Flow

1. App makes request to `https://api.example.com/users`
2. `HttpOverrides` intercepts and routes through proxy
3. Proxy forwards request to upstream
4. Proxy returns response to app

### Configuration

```dart
// HTTP (Dio) - requires HttpOverrides setup
DioDebugger.interceptor(
  proxyBaseUrl: 'http://localhost:9091',
  mode: 'forward',
);

// Must run in zone with overrides
final config = WebSocketDebugger.attach(
  baseUrl: 'wss://example.com/ws',
  proxyBaseUrl: 'http://localhost:9091',
  mode: 'forward',
);

if (config.useForwardOverrides) {
  HttpOverrides.runZoned(
    () async {
      final socket = await WebSocketDebugger.connect(config: config);
      // Use socket...
    },
    createHttpClient: (_) => config.httpClientFactory!() as HttpClient,
  );
}
```

### Advantages

- Original URLs preserved in requests
- More transparent interception
- Works with libraries that don't support URL modification

### Limitations

- **dart:io only** - Does not work on Web platform
- Requires `HttpOverrides` setup
- May conflict with other HttpOverrides

## Mode: None (Bypass)

Setting mode to `none` disables the proxy entirely. Useful for conditional debugging.

```dart
DioDebugger.interceptor(
  proxyBaseUrl: 'http://localhost:9091',
  mode: 'none',  // Bypass proxy completely
);
```

## Choosing the Right Mode

### Use Reverse Proxy (Default) When:

- Building for Web platform
- Building cross-platform apps
- Want simple, reliable interception
- Don't need transparent URL preservation

### Use Forward Proxy When:

- Building dart:io only apps (mobile, desktop)
- Need transparent URL preservation
- Working with libraries that construct URLs internally
- Need system-level traffic interception

## WebSocket Proxy Behavior

For WebSocket connections, the proxy modes work similarly:

### Reverse Mode

```
Client -> ws://proxy:9091/wsproxy?_target=wss://example.com/ws -> Proxy -> wss://example.com/ws
```

### Forward Mode

```
Client -> wss://example.com/ws -> (via HttpOverrides) -> Proxy -> wss://example.com/ws
```

## Socket.IO Specific Behavior

Socket.IO uses Engine.IO for transport, which requires special handling:

1. The `_target` parameter includes Engine.IO query params: `EIO=4&transport=websocket`
2. The proxy normalizes `ws://` to `http://` scheme for Engine.IO handshake
3. Port 80/443 is explicitly included to work around `socket_io_client` bugs

Example target URL:
```
http://example.com:443/socket.io/?EIO=4&transport=websocket
```

## Debugging Mode Selection

To see which mode is active, enable debug logging:

```dart
// Debug logging shows mode selection
// Output: [WscDebugger] mode=reverse base=wss://example.com proxy=http://localhost:9091
```

## Next Steps

- [Platform Support](./platform-support.md) - Platform-specific considerations
- [Configuration Guide](./configuration.md) - Advanced configuration options
- [Troubleshooting](./troubleshooting.md) - Mode-related issues
