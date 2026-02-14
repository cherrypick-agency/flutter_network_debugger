# Changelog

## 0.2.3

- Improved debugger process logging defaults for terminal usage (better readability when stdout/stderr are piped)
- Internal: refactored `DebuggerProcess` to use `dart:io` namespace imports

## 0.2.2

- Added `--github-token` CLI option (alternative to `GITHUB_TOKEN`) to mitigate GitHub API rate limits
- Updated package/CLI version to 0.2.2

## 0.2.1

- Fixed CLI version mismatch and improved CLI UX
- Added support for GITHUB_TOKEN when calling GitHub API (rate limit mitigation)
- Improved checksum lookup logic (supports consolidated checksum files)

## 0.2.0

- Improved binary download reliability
- Better error messages
- Documentation improvements

## 0.1.3

- Fixed binary path resolution on Windows
- Improved caching mechanism

## 0.1.2

- Added automatic binary updates
- SHA256 checksum verification

## 0.1.1

- Bug fixes for process management

## 0.1.0

- Initial release
- Automatic binary download and caching
- Process management for network-debugger server
- Cross-platform support (macOS, Linux, Windows)
