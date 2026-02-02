# Changelog

## 0.2.2

- Reverse proxy now preserves query parameters from upstreamBaseUrl (merged with request query)

## 0.2.1

- Repo hygiene: moved e2e tests to dedicated internal package
- Documentation fixes

## 0.2.0

- Added comprehensive e2e tests with real Go proxy
- Added `HttpDebuggerClient.wrap()` for package:http integration
- Added `HttpDebugger.runZonedWithReverseProxy()` for scoped execution
- Fixed Stack Overflow issue in HttpOverrides
- Improved documentation
- Code comments translated to English

## 0.1.0

- Initial release
- Global HttpOverrides for dart:io
- Forward and reverse proxy modes
- Skip/allow rules for paths, hosts, methods
