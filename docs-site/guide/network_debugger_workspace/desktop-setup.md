# Desktop Setup Guide

This guide explains how to install and run the **Network Debugger desktop app**
on macOS, Windows, and Linux.

For full build and release internals, see the extended document:

- [Full Desktop Setup (repository doc)](https://github.com/cherrypick-agency/flutter_network_debugger/blob/main/docs/DESKTOP_SETUP.md)

## Download

Get desktop artifacts from
[GitHub Releases](https://github.com/cherrypick-agency/flutter_network_debugger/releases).

## Installation

### macOS

1. Download the DMG for your architecture:
   - Intel: `NetworkDebugger-*-macos-x86_64.dmg`
   - Apple Silicon: `NetworkDebugger-*-macos-arm64.dmg`
2. Open DMG and drag app to **Applications**
3. If macOS blocks first run, open:
   **System Settings → Privacy & Security → Open Anyway**

### Windows

1. Download `NetworkDebugger-*-windows-amd64.zip`
2. Extract archive
3. Run `install.bat`
4. App is installed to `%LOCALAPPDATA%\NetworkDebugger`

### Linux

#### Debian/Ubuntu (.deb)

```bash
sudo dpkg -i network-debugger_*_amd64.deb
network-debugger
```

#### Any distro (tar.gz)

```bash
tar -xzf NetworkDebugger-*-linux-amd64.tar.gz
cd NetworkDebugger-*
./install.sh
network-debugger
```

## First launch

On first start, configure:

- API Server Port (default `9092`)
- Forward Proxy Port (default `9091`)

Ports must be available and different from each other.

## Verify setup

1. Launch desktop app
2. Start your mobile/web/desktop client with debugger package integration
3. Confirm sessions appear in the UI

## Related docs

- [Quick Start Guide](./quick-start.md)
- [Platform Support](./platform-support.md)
- [Troubleshooting](./troubleshooting.md)
