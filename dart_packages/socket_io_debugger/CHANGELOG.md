# Changelog

## 1.1.1

- Fixed Web compilation (removed unconditional dart:io imports)
- Repo hygiene: moved e2e tests to dedicated internal package
- Documentation fixes

## 1.1.0

- Added comprehensive e2e tests with real Go Socket.IO upstream
- Added edge-case contract tests
- Fixed socket_io_client port bug (explicit port 80/443 in target URL)
- Replaced `print()` with `dart:developer.log` (stripped in release builds)
- Updated socket_io_client dependency to ^3.1.4
- Comprehensive documentation for path vs namespace
- Code comments translated to English

## 1.0.0

- Breaking: Updated for socket_io_client ^3.x and Socket.IO server v4.7+
- Engine.IO v4 parameters in target URL

## 0.1.0

- Initial release
- Reverse and forward proxy modes
- Support for socket_io_client ^2.x
