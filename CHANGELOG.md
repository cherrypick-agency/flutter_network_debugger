# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-02-16

### Added
- **Firebase Realtime Database debugger** — new `firebase_database_debugger` package for intercepting and inspecting Firebase RTDB operations
- **Error filtering in sessions** — filter requests by error status in the session list
- **`--github-token` option** — pass GitHub token to `network_debugger` CLI for private release downloads

### Changed
- Updated Dart workspace configuration and SDK constraints across all packages
- Improved documentation: README, desktop setup guide, and breakpoints feature description

### Fixed
- Debugger process imports and logging configuration cleanup

## [0.2.2] - 2026-02-13

### Fixed
- GitHub Release description no longer includes the changelog title line (`[X.Y.Z] - YYYY-MM-DD`) — only the section content is shown.

## [0.2.1] - 2026-02-13

### Added
- **Release docs** — standardized process (`docs/RELEASING.md`) and clearer “how to download” sections
- **Dedicated docs pages** — moved Settings and CLI sessions mode into `docs/` with links from the main README
- **Download tables in GitHub Releases** — release notes now include a clear per-OS download table (desktop + web server binaries)

### Changed
- **Quick start** — CLI (web UI in browser) is now the first and fastest way to start; desktop app is the second option
- **CI/CD artifacts** — GitHub Releases include both desktop installers and `network-debugger-web_*` web server binaries

## [0.2.0] - 2026-02-05

### Added
- **Intercept rule management** — new options UI and improved rule configuration for breakpoints
- **Mapping feature DI integration** — mapping is now wired into the dependency injection container with enhanced intercept queue management
- **`--open-browser` flag** — automatically opens the browser when starting the web server
- **Visual density settings** — configurable UI density with custom font preview in settings
- **Dart packages published** — hygiene pass and patch releases for all dart_packages

### Fixed
- Translated all UI strings and code comments from Russian to English for consistency
- Fixed context leak in tests
- Resolved race condition in WebSocket monitor hub test

### Changed
- Bumped all dart_packages versions for publication readiness

## [0.1.13] - 2025-12-29

### Added
- **Selectable text in JSON Tree viewer** — text can now be selected and copied in the JSON tree view (Response → Tree)
- **Copy all headers button** — one-click copy of all headers in the Headers section (Request/Response → Info)

### Fixed
- JSON Tree: clicking a key/value now correctly expands/collapses nested objects
- JSON Tree: clicking a closing bracket `}` / `]` now collapses the node

## [0.1.12] - 2025-12-20

### Fixed
- **Auto-migration for all modes** — database migrations now run correctly regardless of the operating mode
- Removed unsupported `windows/386` from cross-compile targets and CI verification

## [0.1.11] - 2025-12-19

### Changed
- **Pure-Go SQLite driver** — switched to CGO-free SQLite for better Windows compatibility
- Added SQLite PRAGMA optimizations (WAL mode, busy_timeout) for improved concurrent access

### Fixed
- Fixed flaky WebSocket test timeout (increased from 500ms to 3s)
- Removed unsupported `windows/386` build target

## [0.1.10] - 2025-12-19

### Added
- **Accurate body truncation detection** — added `bodyRawSize` field to distinguish between truncated and complete response bodies
- Request and Response body tabs now display actual body sizes with truncation warnings

### Fixed
- Added `UpdateFrameBodyFile` method to all mock repositories for full test coverage

## [0.1.9] - 2025-12-17

### Added
- **Edit & Resend requests** — edit any captured HTTP request and resend it directly from the inspector
- **WASM script validation** — enhanced validation with detailed error messages
- **Hex viewer improvements** — better performance in the hex body renderer
- **Script pool optimization** — improved plugin pool management for memory efficiency

### Fixed
- Better error messages for script compilation and execution failures

## [0.1.8] - 2025-11-19

### Added
- **Automated release workflow** — `make release` for streamlined version bumps and releases
- Native Windows desktop builds in CI/CD

### Fixed
- Use `macos-15-large` runner instead of deprecated `macos-13` for Intel builds
- Switched to free `macos-15-intel` runner to reduce CI costs
- Build number now reads from `pubspec.yaml` instead of git commit count

## [0.1.6] - 2025-11-12

### Added
- **Auto-generated script names** — scripts get meaningful default names
- **Copy and repeat functionality** — copy requests and replay them from inspectors
- **App version display** — version shown in Server Configuration dialog

### Fixed
- Enable CGO for SQLite support in desktop builds
- Added `--no-tree-shake-icons` flag for web builds

## [0.1.4] - 2025-11-11

### Fixed
- Release display logic in Updates dialog
- Windows compilation errors in CI
- Debian package version format normalization

## [0.1.3] - 2025-11-11

### Fixed
- CI/CD stability improvements
- Windows test compatibility fixes
- Dart package formatting for CI compliance

## [0.1.0] - 2025-10-27

### Added
- **Initial release** of Network Debugger
- HTTP/HTTPS request interception and inspection
- WebSocket frame debugging
- Socket.IO message parsing
- Request breakpoints with conditional rules
- Request composer for crafting custom HTTP requests
- WASM-based scripting API for request/response modification
- Session management with import/export (HAR format)
- Forward proxy and MITM proxy support
- Process detection for identifying request origins
- Performance insights dashboard
- Custom fonts support
- Desktop apps for macOS, Windows, and Linux
- Web server mode with embedded Flutter UI
- Dart packages: `network_debugger`, `dio_debugger`, `http_debugger`, `socket_io_debugger`, `web_socket_debugger`, `web_socket_channel_debugger`
