# Network Debugger auto-update system

## Overview

Network Debugger uses a **lightweight custom implementation** for checking
updates via the GitHub Releases API.

**Why this approach:**
- ✅ Works on all platforms: macOS, Windows, Linux
- ✅ No dependency conflicts
- ✅ Simple GitHub Releases integration
- ✅ Full control over the process
- ✅ Minimal requirements (just HTTP + `url_launcher`)

## Architecture

```
┌─────────────────┐
│  BootstrapApp   │
│   (main.dart)   │
└────────┬────────┘
         │
         │ 1. Checks for updates on startup
         ▼
┌─────────────────┐
│  UpdateService  │──────► GitHub Releases API
└────────┬────────┘
         │
         │ 2. If a new version is available
         ▼
┌─────────────────┐
│  UpdateDialog   │
│  shows update   │
│  information    │
└────────┬────────┘
         │
         │ 3. User chooses an action
         ▼
    ┌────┴────────────────┐
    │                     │
Download             Skip/Later
    │                     │
    ▼                     ▼
Opens             Stores in
GitHub Release    SharedPreferences
```

## Configuration

### 1. Configure the GitHub repo in `main.dart`

Open `frontend/lib/main.dart` and verify the parameters:

```dart
await setupDI(
  baseUrl: 'http://localhost:9092',
  githubOwner: 'cherrypick-agency',
  githubRepo: 'flutter_network_debugger',
  currentVersion: packageInfo.version,
);
```

`currentVersion` comes from `PackageInfo.fromPlatform()`, so you don't need to
hardcode the version in code.

### 2. Bump the version in pubspec.yaml

For every release, update the version in `frontend/pubspec.yaml`:

```yaml
version: 1.0.1+2  # major.minor.patch+build
```
This value becomes `PackageInfo.version` and is used to compare against
`tag_name` from GitHub Releases.

## How update checks work

### 1. Automatic checks

The app checks for updates:
- **On app startup**
- **No more than once per hour** (cached)

### 2. GitHub Releases API

The app sends an HTTP request:

```
GET https://api.github.com/repos/OWNER/REPO/releases/latest
```

The response contains:
- `tag_name`: release version (e.g. `"v1.0.1"`)
- `body`: release notes / changelog
- `assets`: downloadable files

### 3. Platform asset selection

The app selects the right asset automatically:

| Platform  | Asset priority                  |
|-----------|---------------------------------|
| macOS     | `*.dmg`                         |
| Windows   | `*.msi`                         |
| Linux     | `*.AppImage` (if available) → `*.deb` → `*.tar.gz` |

### 4. Version comparison

Versions are compared using semantic versioning:
- `1.0.1` > `1.0.0` ✅
- `1.1.0` > `1.0.9` ✅
- `2.0.0` > `1.9.9` ✅

## User experience

### Update dialog

When a new version is available, the app shows a dialog:

```
┌─────────────────────────────────────────┐
│ 🔄 Update Available                     │
├─────────────────────────────────────────┤
│                                         │
│  New Version: v1.0.1     Size: 45.2 MB │
│                                         │
│  What's New:                            │
│  • Fixed critical bug                   │
│  • Added new feature                    │
│  • Performance improvements             │
│                                         │
│  ℹ️ Download from GitHub                 │
│                                         │
├─────────────────────────────────────────┤
│  [Skip] [Later] [Download Update] ⬇️    │
└─────────────────────────────────────────┘
```

### User actions

1. **Download Update** → Opens the GitHub Release page in the browser
2. **Skip This Version** → Don't show this release again
3. **Remind Me Later** → Show again on the next start

## Skip version

Skipped versions are stored in SharedPreferences:

```dart
// User clicked "Skip This Version"
await _updateService.skipVersion('1.0.1');

// Reset (for testing)
await _updateService.clearSkippedVersion();
```

## Caching

Update check results are cached for 1 hour:

```dart
// Force check (ignores cache)
await _updateService.checkForUpdates(forceCheck: true);
```

## GitHub Releases requirements

### File naming

Release assets must have the expected names/extensions:

```
NetworkDebugger-1.0.1-macos-arm64.dmg       ✅
NetworkDebugger-1.0.1-macos-x86_64.dmg      ✅
NetworkDebugger-1.0.1-windows-amd64.zip     ✅
network-debugger_1.0.1_amd64.deb            ✅
NetworkDebugger-1.0.1-linux-amd64.tar.gz    ✅
NetworkDebugger-1.0.1-linux-amd64.deb       ✅
```

### Tags

Release tag should match:

```
v1.0.0   ✅
v1.0.1   ✅
1.0.0    ✅ (also works)
release-1.0.0  ❌
```

### Release notes

Release body is displayed in the dialog as “What’s new”:

```markdown
## What's New

### Features
- Added auto-update functionality
- Improved startup dialog

### Bug Fixes
- Fixed memory leak in session monitor
- Corrected port validation

### Performance
- Optimized websocket handling
```

## Testing

### Local testing

1. Create a test release in GitHub
2. Set an older version in code to simulate an update being available:
   ```dart
   currentVersion: '0.9.0',
   ```
3. Run the app — the update dialog should appear

### Verify the API manually

```bash
curl -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/OWNER/REPO/releases/latest
```

## Disabling auto-updates

For development you can temporarily disable checks:

```dart
// Add this at the beginning of _checkForUpdates():
if (kDebugMode) {
  return; // Skip update check in debug mode
}
```

## Troubleshooting

### Updates are not being checked

**Problem:** No internet connectivity or GitHub API is unavailable

**Fix:** The app fails gracefully. Check logs:

```dart
Logger.root.level = Level.FINE;
Logger.root.onRecord.listen((record) {
  print('${record.level.name}: ${record.time}: ${record.message}');
});
```

### Wrong version is detected

**Problem:** `currentVersion` does not match the build version from pubspec

**Fix:** Make sure `frontend/pubspec.yaml` is bumped and the app is built with
the correct version (the app reads `PackageInfo` at runtime)

### Asset is not found

**Problem:** Asset name/extension does not match what the app expects

**Fix:** Ensure the release contains `.dmg`, `.zip`, `.deb` / `.tar.gz` assets

## Best Practices

1. **Always keep versions consistent:**
   - `frontend/pubspec.yaml` → `version: X.Y.Z+N`
   - GitHub Release tag → `vX.Y.Z`

2. **Write good release notes:**
   - Group changes (Features, Bug Fixes, etc.)
   - Be specific
   - Call out breaking changes

3. **Test before publishing:**
   - Create a draft release
   - Verify update detection
   - Publish only after validation

4. **Versioning rules:**
   - Patch (0.0.x): Bug fixes
   - Minor (0.x.0): New features
   - Major (x.0.0): Breaking changes

## Future improvements

Potential improvements:

- [ ] **Automatic installation** (currently only opens the browser)
- [ ] **Delta updates** (download only changes)
- [ ] **In-app download** with progress UI
- [ ] **Background updates**
- [ ] **Automatic restart** after install
- [ ] **Rollback mechanism**

## See also

- [GitHub Releases API Documentation](https://docs.github.com/en/rest/releases/releases)
- [Semantic Versioning](https://semver.org/)
- [Flutter Desktop Documentation](https://docs.flutter.dev/desktop)
