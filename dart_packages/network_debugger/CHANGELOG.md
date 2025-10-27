# Changelog

## 0.1.0

### Initial Release

**Core Features:**
- Automatic binary download from GitHub releases
- Cross-platform support (Windows, macOS, Linux with x64, ARM64, and 386 architectures)
- Smart binary caching with version management
- Process lifecycle management with graceful shutdown
- Automatic browser opening to web UI

**CLI Tool:**
- Command-line interface for easy launching: `network_debugger`
- Options: `--port`, `--no-browser`, `--binary-version`, `--verbose`, `--quiet`
- Interactive progress display

**Security & Reliability:**
- SHA256 checksum validation for downloaded binaries
- Automatic retry with exponential backoff
- Configurable timeout handling
- Network error handling with user-friendly messages

**Developer Experience:**
- Download progress callbacks
- Retry and checksum callbacks
- Comprehensive logging system with multiple log levels (debug, info, warning, error)
- Detailed error messages with troubleshooting hints
- Full API documentation and examples

**Testing:**
- Comprehensive unit test coverage (75+ tests)
- Integration tests for real-world scenarios
- CI/CD testing on Windows, macOS, and Linux
