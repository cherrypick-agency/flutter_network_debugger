# Changelog

## 0.2.1

- Repo hygiene: moved e2e tests to dedicated internal package
- Documentation fixes

## 0.2.0

- Added comprehensive e2e tests with real Go proxy
- Added edge-case contract tests
- Replaced `print()` with `dart:developer.log` (stripped in release builds)
- Automatic http/https to ws/wss scheme conversion
- Improved URL normalization for proxy paths
- Comprehensive documentation
- Code comments translated to English

## 0.1.0

- Initial release
- Reverse and forward proxy modes
- Support for package:web_socket_channel
- Cross-platform (dart:io and Web)
