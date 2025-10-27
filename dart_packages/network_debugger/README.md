# network_debugger

Dart package that automatically downloads, caches, and launches the `network-debugger-web` binary for easy network debugging.

## Features

- Automatic binary download from GitHub releases
- Smart caching with version management
- Platform detection (Windows, macOS, Linux with various architectures)
- Download progress tracking
- Simple process management
- Automatic browser opening to web UI

## Installation

Add to your `pubspec.yaml`:

```yaml
dependencies:
  network_debugger: ^0.1.0
```

## Quick Start

```dart
import 'package:network_debugger/network_debugger.dart';

void main() async {
  // Launch debugger (downloads if needed, uses cache otherwise)
  final debugger = await NetworkDebugger.launch(
    port: 9091,
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

## License

MIT
