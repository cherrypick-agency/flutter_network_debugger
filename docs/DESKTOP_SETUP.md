# Network Debugger - Desktop Application Setup

## Overview

Network Debugger provides native desktop apps for:
- **macOS** (Intel x86_64 and Apple Silicon arm64) — DMG installer
- **Windows** (64-bit) — ZIP archive with `install.bat`
- **Linux** (64-bit) — tar.gz and deb packages

The desktop app runs both the Flutter UI and the Go proxy server as a single
application.

## Architecture

```
┌─────────────────────────────────────┐
│   Flutter Desktop App (UI)          │
│                                      │
│  ┌────────────────────────────────┐ │
│  │  BootstrapApp                  │ │
│  │  - Startup Dialog              │ │
│  │  - Port Configuration          │ │
│  │  - Auto-Update Check           │ │
│  └────────┬───────────────────────┘ │
│           │                          │
│           ▼                          │
│  ┌────────────────────────────────┐ │
│  │  GoServerManager               │ │
│  │  - Launch subprocess           │ │
│  │  - Health monitoring           │ │
│  │  - Graceful shutdown           │ │
│  └────────┬───────────────────────┘ │
│           │                          │
└───────────┼──────────────────────────┘
            │ Process.start()
            ▼
┌─────────────────────────────────────┐
│   Go Server (subprocess)             │
│   - Forward Proxy                    │
│   - HTTP API                         │
│   - WebSocket connections            │
│   - Session storage                  │
└─────────────────────────────────────┘
```

## Development requirements

### General
- Flutter SDK 3.x or newer
- Go 1.22.x or newer
- Git

### macOS
- macOS 11 (Big Sur) or newer
- Xcode Command Line Tools
- `brew install create-dmg` (optional, for nicer DMG)

### Windows
- Windows 10 or newer
- Visual Studio 2022 or Visual Studio Build Tools
- PowerShell 5.1 or newer

### Linux
- Ubuntu 20.04+ or a similar distro
- GTK 3.0 development headers
- Required packages:
  ```bash
  sudo apt-get install \
    clang cmake ninja-build pkg-config \
    libgtk-3-dev liblzma-dev libstdc++-12-dev
  ```

## Local builds

### macOS

```bash
# Enable desktop support (first time only)
cd frontend
flutter create --platforms=macos .

# Build DMG
cd ..
chmod +x scripts/package-macos.sh
VERSION=1.0.0 ./scripts/package-macos.sh

# Output: dist/NetworkDebugger-1.0.0-macos-{arch}.dmg
```

### Windows

```powershell
# Enable desktop support (first time only)
cd frontend
flutter create --platforms=windows .

# Build ZIP
cd ..
.\scripts\package-windows.ps1 -Version "1.0.0"

# Output: dist\NetworkDebugger-1.0.0-windows-amd64.zip
```

### Linux

```bash
# Enable desktop support (first time only)
cd frontend
flutter create --platforms=linux .

# Build tar.gz and deb
cd ..
chmod +x scripts/package-linux.sh
VERSION=1.0.0 ARCH=amd64 ./scripts/package-linux.sh

# Output:
# - dist/NetworkDebugger-1.0.0-linux-amd64.tar.gz
# - dist/network-debugger_1.0.0_amd64.deb
```

## CI/CD Pipeline

### GitHub Actions Workflow

Workflow `.github/workflows/build-desktop.yml` runs automatically:

1. **Triggers:**
   - Push to `main`
   - Pull requests
   - Version tags (`v*.*.*`)
   - Manual dispatch

2. **Jobs:**
   - `build-macos`: builds DMG for macOS (x86_64 and arm64)
   - `build-windows`: builds ZIP for Windows (amd64)
   - `build-linux`: builds tar.gz and deb for Linux (amd64)
   - `release`: on tag push, creates a GitHub Release with artifacts
   - `summary`: shows build status

3. **Artifacts:**
   - retained for 7 days for non-release builds
   - attached to GitHub Releases for version tags

### Creating a release

```bash
# Full release process: docs/RELEASING.md
#
# TL;DR:
# 1) Update CHANGELOG.md and frontend/pubspec.yaml
# 2) make fmt && make test
# 3) make release VERSION=X.Y.Z
# 4) GitHub Actions will build desktop and create a GitHub Release
```

## Installation

### macOS

1. Download the DMG for your architecture:
   - Intel: `NetworkDebugger-*-macos-x86_64.dmg`
   - Apple Silicon: `NetworkDebugger-*-macos-arm64.dmg`

2. Open the DMG and drag the app into Applications

3. On first launch: System Preferences → Security & Privacy → "Open Anyway"

### Windows

1. Download `NetworkDebugger-*-windows-amd64.zip`

2. Extract the ZIP anywhere

3. Run `install.bat` (creates shortcuts on Desktop and Start Menu)

4. The app will be installed to `%LOCALAPPDATA%\NetworkDebugger`

### Linux

#### Using .deb (Ubuntu/Debian):
```bash
sudo dpkg -i network-debugger_*_amd64.deb
network-debugger
```

#### Using tar.gz (any distro):
```bash
# Extract the archive
tar -xzf NetworkDebugger-*-linux-amd64.tar.gz
cd NetworkDebugger-*

# Install
./install.sh

# Run
network-debugger
```

## Usage

### First run

On launch you will see a **Startup Dialog** to configure ports:

```
┌─────────────────────────────────────┐
│ Network Debugger - Configuration    │
├─────────────────────────────────────┤
│                                     │
│  API Server Port:    [9092]         │
│  Forward Proxy Port: [9093]         │
│                                     │
│  ℹ️ These ports must be available   │
│  and different from each other      │
│                                     │
├─────────────────────────────────────┤
│           [Cancel]  [Start]         │
└─────────────────────────────────────┘
```

**Settings:**
- **API Server Port**: port for UI and REST API (default: 9092)
- **Forward Proxy Port**: port for forward proxy (default: 9093)

**Validation:**
- ports must be in range 1024-65535
- ports must be different
- ports must be available

After clicking **Start**:
1. The Go server starts with the configured ports
2. The health endpoint (`/_health`) is checked
3. Flutter UI connects to the server
4. The app is ready to use

### Settings persistence

Selected ports are stored in SharedPreferences:
- macOS: `~/Library/Preferences/com.belieflab.networkDebugger`
- Windows: Registry `HKCU\Software\belieflab\network-debugger`
- Linux: `~/.local/share/network-debugger/shared_preferences.json`

Saved values are used on the next launch.

### Changing ports

To change ports after installation:
1. Settings → Server Settings → Restart with different ports
2. Or delete saved preferences and restart

## Auto-updates

The desktop app automatically checks for updates via the GitHub Releases API.

### How it works

1. **On startup** (no more than once per hour)
2. Fetch latest release via API
3. Compare against current version (semantic versioning)
4. If a newer version exists → show a dialog

### Update Dialog

```
┌─────────────────────────────────────┐
│ 🔄 Update Available                 │
├─────────────────────────────────────┤
│                                     │
│  New Version: v1.0.1  Size: 45 MB   │
│                                     │
│  What's New:                        │
│  • Fixed critical bug               │
│  • Added new feature                │
│  • Performance improvements         │
│                                     │
├─────────────────────────────────────┤
│  [Skip]  [Later]  [Download] ⬇️     │
└─────────────────────────────────────┘
```

**Actions:**
- **Download Update**: Opens the GitHub Release page in the browser
- **Skip This Version**: Don't show this release again
- **Remind Me Later**: Show again on the next launch

See [AUTO_UPDATE.md](AUTO_UPDATE.md) for details.

## Troubleshooting

### macOS: "App is damaged and can't be opened"

```bash
xattr -cr "/Applications/Network Debugger.app"
```

### Windows: "Windows protected your PC"

1. Click "More info"
2. Click "Run anyway"

Or run the installer as Administrator.

### Linux: Port permission denied (< 1024)

```bash
# Use ports >= 1024 (e.g. 9092/9093)
# Or grant capability:
sudo setcap 'cap_net_bind_service=+ep' ~/.local/share/network-debugger/resources/server_linux_amd64
```

### Go server does not start

**Check:**
1. Ports are not taken: `lsof -i :9092` (macOS/Linux) or
   `netstat -ano | findstr 9092` (Windows)
2. Server binary exists in Resources:
   - macOS: `Network Debugger.app/Contents/Resources/server_darwin_*`
   - Windows: `resources\server_windows_amd64.exe`
   - Linux: `resources/server_linux_amd64`
3. Binary executable: `chmod +x <binary>`

**Logs:**
- macOS: Console.app → Filter: "network-debugger"
- Windows: Event Viewer → Application logs
- Linux: `journalctl -f | grep network-debugger`

### Flutter assets are missing / not loading

The app bundle must include Flutter assets.

**Check:**
```bash
# After build you should have:
frontend/build/macos/Build/Products/Release/Network Debugger.app/Contents/Frameworks/App.framework/Resources/flutter_assets/

# It should contain:
- assets/
- fonts/
- packages/
```

## Development

### Local development run

#### Option 1: Separate processes (recommended for dev)

Terminal 1 - Go server:
```bash
cd cmd/network-debugger
go run . --api-port 9092 --proxy-port 9093
```

Terminal 2 - Flutter desktop:
```bash
cd frontend
flutter run -d macos  # or -d windows / -d linux
```

#### Option 2: Full desktop integration

```bash
# 1. Build Go binary
cd cmd/network-debugger
go build -o ../../frontend/macos/Runner/Resources/server_darwin_arm64 .

# 2. Run Flutter desktop
cd ../../frontend
flutter run -d macos
```

### Debugging

#### Flutter DevTools

```bash
cd frontend
flutter run -d macos --observatory-port=9090
# Then open DevTools in the browser
```

#### Go delve debugger

```bash
cd cmd/network-debugger
dlv debug . -- --api-port 9092 --proxy-port 9093
```

### Testing packaging scripts

```bash
# macOS
VERSION=dev ./scripts/package-macos.sh
open dist/*.dmg

# Windows
.\scripts\package-windows.ps1 -Version "dev"
Expand-Archive dist\*.zip -DestinationPath dist\test

# Linux
VERSION=dev ./scripts/package-linux.sh
tar -xzf dist/*.tar.gz -C dist/
```

## Project structure

```
go-proxy/
├── cmd/network-debugger/         # Go server entry point
│   └── main.go                    # CLI flags: --api-port, --proxy-port, --data-dir
├── frontend/
│   ├── lib/
│   │   ├── core/
│   │   │   ├── desktop/
│   │   │   │   └── desktop_bootstrap.dart  # Desktop initialization
│   │   │   ├── go_server/
│   │   │   │   ├── go_server_manager.dart  # Process management
│   │   │   │   ├── go_server_path_io.dart  # Binary path resolution
│   │   │   │   └── go_server_path_stub.dart
│   │   │   └── update/
│   │   │       ├── update_service.dart      # GitHub Releases API
│   │   │       ├── update_dialog.dart       # Update UI
│   │   │       └── update_info.dart         # Version comparison
│   │   ├── features/startup/
│   │   │   └── startup_dialog.dart          # Port configuration dialog
│   │   └── main.dart                         # Entry point with BootstrapApp
│   ├── macos/                     # macOS specific
│   ├── windows/                   # Windows specific
│   └── linux/                     # Linux specific
├── scripts/
│   ├── package-macos.sh           # macOS DMG builder
│   ├── package-windows.ps1        # Windows ZIP builder
│   └── package-linux.sh           # Linux tar.gz/deb builder
├── .github/workflows/
│   └── build-desktop.yml          # CI/CD for desktop builds
└── docs/
    ├── DESKTOP_SETUP.md           # This file
    └── AUTO_UPDATE.md             # Auto-update documentation
```

## Best Practices

### Versioning

Always keep versions consistent:
1. `frontend/pubspec.yaml`: `version: 1.0.1+2`
2. Git tag: `v1.0.1`

### Changelog

Use conventional commits:
- `feat:` - new features
- `fix:` - bug fixes
- `chore:` - maintenance work
- `docs:` - documentation

### Testing before release

```bash
# 1. Build installers (prefer CI for true multi-platform)
# 2. Test installation:
# - macOS: open DMG, install, run
# - Windows: extract ZIP, run install.bat, run
# - Linux: install deb/tar.gz, run
#
# 3. Verify update detection:
# - install an older version
# - create a draft release with a newer version
# - start the app and confirm the update dialog
#
# 4. Create the real release:
git tag v1.0.1 && git push origin v1.0.1
```

## FAQ

**Q: Can we cross-compile everything from one machine?**

A: Partially. Go supports cross-compilation, but Flutter desktop requires a
native toolchain per platform. Use GitHub Actions for multi-platform builds.

**Q: How to add code signing?**

A:
- macOS: `codesign --deep --force --verify --verbose --sign "Developer ID" "Network Debugger.app"`
- Windows: use `signtool.exe` with a certificate
- Linux: code signing is usually not required

**Q: Can we build a .pkg/.exe installer instead of DMG/ZIP?**

A: Yes:
- macOS: use `pkgbuild` for a `.pkg` installer
- Windows: use Inno Setup or NSIS for an `.exe` installer
- Linux: Snap, Flatpak, or AppImage

**Q: How to update Go dependencies in the desktop app?**

A:
```bash
# Rebuild Go binary
cd cmd/network-debugger
go mod tidy
go build -o <destination> .

# Run the packaging script
cd ../..
./scripts/package-<platform>.sh
```

**Q: Can I run multiple app instances?**

A: Yes, but make sure they use different ports. The startup dialog helps to
avoid conflicts.

## Links

- [Flutter Desktop](https://docs.flutter.dev/desktop)
- [Go Cross-compilation](https://go.dev/doc/install/source#environment)
- [GitHub Releases API](https://docs.github.com/en/rest/releases/releases)
- [Semantic Versioning](https://semver.org/)
- [macOS App Distribution](https://developer.apple.com/documentation/xcode/distributing-your-app-for-beta-testing-and-releases)
- [Windows App Packaging](https://docs.microsoft.com/en-us/windows/msix/desktop/desktop-to-uwp-packaging-dot-net)
