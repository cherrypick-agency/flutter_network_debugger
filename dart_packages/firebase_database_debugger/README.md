# firebase_database_debugger

<p align="left">
  <a href="https://pub.dev/packages/firebase_database_debugger"><img src="https://img.shields.io/pub/v/firebase_database_debugger.svg" alt="pub version" /></a>
  <a href="https://pub.dev/packages/firebase_database_debugger/score"><img src="https://img.shields.io/pub/likes/firebase_database_debugger" alt="likes" /></a>
  <a href="https://pub.dev/packages/firebase_database_debugger/score"><img src="https://img.shields.io/pub/points/firebase_database_debugger" alt="pub points" /></a>
</p>

## Getting started with Network Debugger

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

> This package is for integrating Network Debugger with **Firebase Realtime Database**.

> Part of the [network_debugger](https://pub.dev/packages/network_debugger) ecosystem

Wrapper around `firebase_database` that intercepts all RTDB operations (set, get, update, remove, onValue, queries) and sends them to the Network Debugger UI as structured events. Each operation appears as a frame with colored badges (op, path, ok/error) and expandable JSON payload.

## Features

- Drop-in wrappers: `DebugDatabaseReference` and `DebugQuery`
- Intercepts: `set`, `get`, `update`, `remove`, `onValue`, `onChildAdded/Changed/Removed`
- Structured events with operation type, path, payload, duration, and error info
- Session grouping by path depth (configurable)
- Automatic batching and flushing to minimize overhead
- Large payload handling (base64 body spill for payloads > 16 KB)
- Error tracking with `PERMISSION_DENIED` and other Firebase errors
- Works alongside other debugger packages (dio_debugger, web_socket_debugger, etc.)

## Installation

Add to your `pubspec.yaml`:

```yaml
dependencies:
  firebase_database: ^11.1.6
  firebase_database_debugger: ^0.1.1
```

## Quick start

```dart
import 'package:firebase_database/firebase_database.dart';
import 'package:flutter/foundation.dart';
import 'package:firebase_database_debugger/firebase_database_debugger.dart';

final db = FirebaseDatabase.instance;
final debugger = FirebaseDatabaseDebugger(
  config: FirebaseDatabaseDebuggerConfig(
    enabled: kDebugMode,
  ),
);

// Wrap your DatabaseReference:
Future<void> example() async {
  final ref = debugger.ref(db.ref('users/alice'));

  await ref.set({'name': 'Alice', 'age': 30});   // logged as SET
  final snap = await ref.get();                    // logged as GET
  await ref.update({'age': 31});                   // logged as UPDATE
  await ref.remove();                              // logged as REMOVE

  // Listeners are also tracked:
  ref.onValue.listen((event) {
    print(event.snapshot.value);                   // logged as onValue
  });
}
```

### Session grouping

By default, each unique path gets its own session. Use `sessionPathDepth` to group related operations:

```dart
FirebaseDatabaseDebuggerConfig(
  // Group by first 2 path segments:
  // /users/alice/profile and /users/alice/settings → one session "/users/alice"
  sessionPathDepth: 2,
)
```

| `sessionPathDepth` | Behavior |
| --- | --- |
| `-1` (default) | One session per full path |
| `0` | Single session for entire database |
| `N` (1, 2, ...) | Session per path prefix of depth N |

### Query debugging

```dart
final query = debugger.query(
  db.ref('messages').orderByChild('timestamp').limitToLast(50),
);

final snap = await query.get();  // logged as query_get
query.onValue.listen((event) {}); // logged as onValue
```

### Advanced options

```dart
FirebaseDatabaseDebuggerConfig(
  // Optional override:
  // debuggerBaseUrl: 'http://192.168.1.100:9092',
  databaseUrl: 'https://my-project.firebaseio.com',
  enabled: kDebugMode,
  sessionPathDepth: 2,
  flushInterval: Duration(milliseconds: 300),
  maxBatchFrames: 50,
  previewBodyThresholdBytes: 32 * 1024,  // 32 KB before base64 spill
)
```

You can also configure defaults via `--dart-define`:

```bash
flutter run \
  --dart-define=FIREBASE_DATABASE_DEBUGGER_BASE_URL=http://10.0.2.2:9092 \
  --dart-define=FIREBASE_DATABASE_DEBUGGER_ENABLED=true
```

Supported aliases:
- `FIREBASE_DATABASE_DEBUGGER_BASE_URL` or `FIREBASE_DEBUGGER_BASE_URL`
- `FIREBASE_DATABASE_DEBUGGER_ENABLED` or `FIREBASE_DEBUGGER_ENABLED`

### Cleanup

```dart
// Flush pending events and close the session
await debugger.dispose();
```

## How it works

The package wraps `DatabaseReference` and `Query` objects. Each operation is intercepted, executed on the real Firebase SDK, and then a structured event is sent to the Network Debugger backend via the ingest API (`POST /_api/v1/ingest/firebase_database`).

Events are batched and flushed periodically to minimize network overhead. The UI displays them as "RTDB events" with colored badges for operation type, path, and status.

## Notes

- The debugger adds minimal overhead — events are batched and sent asynchronously.
- If the Network Debugger backend is not running, events are silently dropped.
- Set `enabled: false` (or use `kDebugMode`) to completely disable in production.
- Works with Firebase Emulator — just point `databaseUrl` to your emulator URL.
- `debuggerBaseUrl` defaults to `http://10.0.2.2:9092` on Android emulator,
  and `http://localhost:9092` on other platforms.

## License

MIT
