# Changelog

## 0.1.1

- Add platform-aware default `debuggerBaseUrl`:
  - Android emulator: `http://10.0.2.2:9092`
  - Other platforms: `http://localhost:9092`
- Make `debuggerBaseUrl` optional in `FirebaseDatabaseDebuggerConfig`
- Add URL normalization for `debuggerBaseUrl` (supports values like `localhost:9092`)
- Add `--dart-define` support for defaults:
  - `FIREBASE_DATABASE_DEBUGGER_BASE_URL` / `FIREBASE_DEBUGGER_BASE_URL`
  - `FIREBASE_DATABASE_DEBUGGER_ENABLED` / `FIREBASE_DEBUGGER_ENABLED`
- Improve README examples (no `late`, clearer default behavior)

## 0.1.0

- Initial release
- `DebugDatabaseReference` wrapper for `DatabaseReference` (set, get, update, remove, onValue)
- `DebugQuery` wrapper for `Query` (get, onValue, onChildAdded/Changed/Removed)
- `FirebaseDatabaseDebuggerConfig` with session grouping by path depth
- Automatic batching and periodic flushing of events
- Large payload handling with base64 body spill
- Error tracking (PERMISSION_DENIED, etc.)
- Structured event format with operation type, path, payload, and timing
