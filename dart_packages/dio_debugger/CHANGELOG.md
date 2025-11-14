# Changelog

## 0.1.5
- Feature: Added `resetCaptureOnHotRestart` parameter to `DioDebugger.attach()`.
  When set to `true`, automatically clears previous sessions and starts a new capture
  on each hot restart, making it easy to separate traffic from different app runs.
- Implementation: Uses `_resetCapture=true` query parameter in the first proxied request
  for zero-latency capture reset (no separate HTTP call).
- Backend: Added `_resetCapture` query parameter support in `/httpproxy` handler
  for atomic capture reset operation.
- Docs: Added "Auto-reset capture on hot restart" section to README.
- Docs: Updated example with `resetCaptureOnHotRestart` usage.

## 0.1.4
- Fix: eliminated double `?` and encoded `%3F` artifacts in query when building target URL.
  Now `ReverseProxyInterceptor` normalizes keys (strips leading `?`) and
  builds query string via `Uri.replace(query: ...)`, preventing errors
  like `??page=...` and `?%3Fpage=...` in rare cases.
- Internal: non-breaking changes, API unchanged.

## 0.1.3
- Docs: Added "Part of network_debugger ecosystem" reference
- Docs: Added "Starting the Proxy" section with setup instructions
- Docs: Added cross-references to network_debugger package
- Docs: Improved documentation structure and clarity

## 0.1.2
- Feature: `attach` now falls back to `dio.options.baseUrl` when `upstreamBaseUrl` is not provided.
- Docs: updated README with configuration precedence and a simplified quick start example.
- Docs: replaced specific URLs with abstract ones (`https://api.example.test`).
- Docs: removed base URL duplication in the example (extracted to a variable).

## 0.1.1
- Docs: minor fixes and small improvements in examples.

## 0.1.0
- Initial release.
- Adds reverse/forward proxy support for Dio via a single attach call.
- Supports env and --dart-define configuration, with allow/skip filters.


