# Network Debugger - Dart Packages Documentation

This documentation covers the Dart client packages for integrating with the Network Debugger proxy server.

## Overview

Network Debugger is a local proxy server that intercepts and records HTTP and WebSocket traffic from your Dart/Flutter applications. These packages provide seamless integration with popular Dart networking libraries.

## Available Packages

| Package | Purpose | Library Integration |
| ------- | ------- | ------------------- |
| [dio_debugger](./packages/dio_debugger.md) | HTTP debugging | `package:dio` |
| [http_debugger](./packages/http_debugger.md) | HTTP debugging | `package:http`, `dart:io HttpClient` |
| [web_socket_debugger](./packages/web_socket_debugger.md) | WebSocket debugging | `package:web_socket` |
| [web_socket_channel_debugger](./packages/web_socket_channel_debugger.md) | WebSocket debugging | `package:web_socket_channel` |
| [socket_io_debugger](./packages/socket_io_debugger.md) | Socket.IO debugging | `package:socket_io_client` |

## Documentation Index

- [Quick Start Guide](./quick-start.md) - Get up and running in 5 minutes
- [Desktop Setup Guide](./desktop-setup.md) - Install and run desktop app
- [Configuration Guide](./configuration.md) - Environment variables, dart-define, and runtime config
- [Proxy Modes](./proxy-modes.md) - Understanding reverse vs forward proxy
- [Platform Support](./platform-support.md) - dart:io vs Web platform differences
- [Troubleshooting](./troubleshooting.md) - Common issues and solutions
- [API Reference](./api-reference.md) - Complete API documentation

## Quick Links

### Starting the Proxy Server

```bash
# Install CLI globally
dart pub global activate network_debugger

# Start proxy server
network_debugger

# Default ports:
# - Proxy: http://localhost:9091
# - Web UI: http://localhost:9092
```

### Basic Integration Example

```dart
import 'package:dio/dio.dart';
import 'package:dio_debugger/dio_debugger.dart';

void main() {
  final dio = Dio();
  
  // Add debugger interceptor
  dio.interceptors.add(
    DioDebugger.interceptor(
      proxyBaseUrl: 'http://localhost:9091',
    ),
  );
  
  // All requests now go through the proxy
  dio.get('https://api.example.com/users');
}
```

## Architecture

```
┌─────────────────────┐
│   Flutter/Dart App  │
│  ┌───────────────┐  │
│  │ dio_debugger  │  │
│  │ http_debugger │  │
│  │ ws_debugger   │  │
│  └───────┬───────┘  │
└──────────┼──────────┘
           │
           ▼
┌─────────────────────┐
│  Network Debugger   │
│    Proxy Server     │
│  (localhost:9091)   │
└──────────┬──────────┘
           │
           ▼
┌─────────────────────┐
│   Upstream Server   │
│  (api.example.com)  │
└─────────────────────┘
```

## Requirements

- Dart SDK >= 3.0.0
- Flutter >= 3.10.0 (for Flutter projects)
- Go >= 1.21 (for running the proxy server)

## License

MIT
