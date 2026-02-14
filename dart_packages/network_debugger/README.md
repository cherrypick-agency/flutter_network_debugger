# network_debugger

## Quick start (recommended)

1) CLI (WEB UI in browser) — fastest way to start

```bash
dart pub global activate network_debugger
network_debugger
```

This starts the proxy and opens the UI in your browser:
- UI: `http://localhost:9092/`
- Proxy base (HTTP/WebSocket forward): `http://localhost:9091`

2) Desktop App (Native GUI)

Download the desktop application from
[GitHub Releases](https://github.com/cherrypick-agency/flutter_network_debugger/releases).
It bundles the proxy server and UI.

![Screen recording (2025-10-02)](https://github.com/user-attachments/assets/43044ece-e6b4-4702-80bc-0584e844c042)

Free tool for debugging HTTP and WebSocket traffic which is much better than
the built-in Flutter Network DevTools.

Suitable for local development and test environments. Has web interface (opens in browser), desktop and CLI.

## Features

What it can do
- Intercept and view HTTP and WebSocket traffic
- Waterfall timeline of requests
- grouping by domain/route
- Filters: method, status, MIME, minimum duration, by headers
- Convenient search with highlighting
- HTTP details: headers (with sensitive data masking), body (pretty/JSON tree), TTFB/Total
- CORS/Cache hints, cookies and TLS summary
- WebSocket details: events/frames, pings/pongs, payload preview
- HAR export
- Artificial response delay (useful for simulating "slow networks")
- Record/stop and records management
- HTML preview
- Form Data (show files) For example Flutter devtools don't show at all
- You can proxy only app requests or all OS requests (forward proxy)
- Crossplatform

## Quick Start (CLI)

### Installation

```bash
# Install globally from pub.dev
dart pub global activate network_debugger

# OR install locally from path
cd dart_packages/network_debugger
dart pub global activate --source path .
```

### Usage

```bash
# Launch debugger (downloads binary if needed, opens browser automatically)
network_debugger

# Custom port
network_debugger --port 8080

# Without opening browser
network_debugger --no-browser

# Specific binary version
network_debugger --binary-version v1.0.0

# Show version
network_debugger --version

# Show help
network_debugger --help
```

Press `Ctrl+C` to stop the debugger.

## How It Works

1. Detects your platform (OS + architecture)
2. Checks local cache (`~/.cache/network_debugger/` or platform-specific)
3. Downloads from GitHub releases if not cached
4. Extracts and caches the binary
5. Launches the process
6. Returns a `DebuggerInstance` for process management

## Cache Location

- **macOS/Linux**: `~/.cache/network_debugger/`
- **Windows**: `%LOCALAPPDATA%\network_debugger\Cache\`

Each version is stored separately for easy version switching.


## Programmatic usage (Dart API)

You can also use the package programmatically in your Dart code:

### Installation

Add to your `pubspec.yaml`:

```yaml
dependencies:
  network_debugger: ^0.2.3
```

### Usage

```dart
import 'package:network_debugger/network_debugger.dart';

void main() async {
  // Launch debugger (downloads if needed, uses cache otherwise)
  final debugger = await NetworkDebugger.launch(
    port: 9092,
    onProgress: (received, total) {
      print('Download progress: ${(received / total * 100).toStringAsFixed(1)}%');
    },
  );

  print('Network debugger running at: ${debugger.url}');

  // When done...
  await debugger.stop();
}
```

## Advanced Usage

### Specify Version

```dart
final debugger = await NetworkDebugger.launch(
  version: 'v1.0.0',  // or null for latest
);
```

### Custom Configuration

```dart
final debugger = await NetworkDebugger.launch(
  port: 8080,
  autoOpenBrowser: false,  // Don't open browser automatically
  onProgress: (received, total) {
    final percent = (received / total * 100).toStringAsFixed(1);
    print('Downloading: $percent%');
  },
);
```

### Cache Management

```dart
// Clear all cached binaries
await NetworkDebugger.clearCache();

// Clear specific version
await NetworkDebugger.clearCache(version: 'v1.0.0');
```


## License

MIT
