# Guide

Network Debugger captures HTTP, WebSocket, Socket.IO, and Firebase RTDB traffic from your Flutter/Dart app and displays it in a real-time UI.

## Getting Started

1. **[Quick Start Guide](./network_debugger_workspace/quick-start.md)** — install the CLI, add a package, and see your first captured request in under 2 minutes.
2. **[Desktop Setup](./network_debugger_workspace/desktop-setup.md)** — download the native macOS / Windows / Linux app.

## Pick Your Package

Choose the package that matches your networking stack:

| Stack | Package | Install |
|---|---|---|
| **Dio** | [dio_debugger](./network_debugger_workspace/packages/dio_debugger.md) | `dart pub add dio_debugger` |
| **package:http** | [http_debugger](./network_debugger_workspace/packages/http_debugger.md) | `dart pub add http_debugger` |
| **WebSocket** | [web_socket_debugger](./network_debugger_workspace/packages/web_socket_debugger.md) | `dart pub add web_socket_debugger` |
| **WebSocketChannel** | [web_socket_channel_debugger](./network_debugger_workspace/packages/web_socket_channel_debugger.md) | `dart pub add web_socket_channel_debugger` |
| **Socket.IO** | [socket_io_debugger](./network_debugger_workspace/packages/socket_io_debugger.md) | `dart pub add socket_io_debugger` |
| **Firebase RTDB** | [firebase_database_debugger](./firebase-database.md) | `dart pub add firebase_database_debugger` |

## Configuration & Modes

- **[Configuration Guide](./network_debugger_workspace/configuration.md)** — all settings, environment variables, and `--dart-define` options.
- **[Proxy Modes](./network_debugger_workspace/proxy-modes.md)** — reverse vs forward proxy, when to use which.
- **[Platform Support](./network_debugger_workspace/platform-support.md)** — what works on iOS, Android, Web, and Desktop.

## Reference

- **[API Reference](/api/)** — generated class and method documentation for every package.
- **[Troubleshooting](./network_debugger_workspace/troubleshooting.md)** — common issues and solutions.
