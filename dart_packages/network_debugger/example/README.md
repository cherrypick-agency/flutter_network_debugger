# Examples

This directory contains examples of how to use the `network_debugger` package.

## Running the examples

### Basic launcher example

```bash
dart run example/main.dart
```

This will:
1. Display platform and cache information
2. Download the debugger binary (if not cached)
3. Launch the debugger on port 9092 (UI); proxy is on 9091
4. Keep it running for 30 seconds
5. Stop the debugger gracefully

## More examples

### Launch with custom port

```dart
final debugger = await NetworkDebugger.launch(
  port: 8080,
);
```

### Launch specific version

```dart
final debugger = await NetworkDebugger.launch(
  version: 'v1.0.0',
);
```

### Launch without auto-opening browser

```dart
final debugger = await NetworkDebugger.launch(
  autoOpenBrowser: false,
);
```

### Track download progress

```dart
final debugger = await NetworkDebugger.launch(
  onProgress: (received, total) {
    final percent = (received / total * 100).toStringAsFixed(1);
    print('Downloading: $percent%');
  },
);
```

### Clear cache

```dart
// Clear all cached versions
await NetworkDebugger.clearCache();

// Clear specific version
await NetworkDebugger.clearCache(version: 'v1.0.0');
```

### Check cache information

```dart
final versions = await NetworkDebugger.listCachedVersions();
final size = await NetworkDebugger.getCacheSizeFormatted();
final dir = await NetworkDebugger.getCacheDirectory();

print('Cached versions: $versions');
print('Cache size: $size');
print('Cache directory: $dir');
```
