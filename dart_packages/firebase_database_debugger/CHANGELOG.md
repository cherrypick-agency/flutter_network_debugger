# Changelog

## 0.1.0

- Initial release
- `DebugDatabaseReference` wrapper for `DatabaseReference` (set, get, update, remove, onValue)
- `DebugQuery` wrapper for `Query` (get, onValue, onChildAdded/Changed/Removed)
- `FirebaseDatabaseDebuggerConfig` with session grouping by path depth
- Automatic batching and periodic flushing of events
- Large payload handling with base64 body spill
- Error tracking (PERMISSION_DENIED, etc.)
- Structured event format with operation type, path, payload, and timing
