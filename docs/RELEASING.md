# Releasing (desktop)

This repository publishes **desktop apps only** to GitHub Releases:
macOS (DMG), Windows (ZIP), Linux (DEB and/or tar.gz).

Web builds are intentionally excluded from CD. It keeps the pipeline fast and
avoids confusion about which artifact users should download.

## What “version” means here

- **Git tag**: `vX.Y.Z` (SemVer).
- **Frontend (Flutter)**: `frontend/pubspec.yaml` → `version: X.Y.Z+N`.
  - `X.Y.Z` is the release version.
  - `N` is the build number (monotonically increasing; required on some
    platforms).

The app reads `currentVersion` via `PackageInfo.fromPlatform()`, so the single
source of truth is **pubspec + build parameters**.

## Where to bump versions before a release

- **`CHANGELOG.md` (repo root)**: add a new version section.
- **`frontend/pubspec.yaml`**: update `version: X.Y.Z+N`.
- **Optional (if publishing packages)**:
  - `dart_packages/*/pubspec.yaml`
  - `dart_packages/*/CHANGELOG.md`

## Changelog: keep it useful

Format: Keep a Changelog.

Guidelines:
- Write user-facing changes, not internal implementation details.
- Group changes into sections: `Added`, `Changed`, `Fixed`, `Deprecated`,
  `Removed`, `Security`.
- If there are breaking changes, call them out at the top of the release
  section.

## How to release (recommended)

1) **Update `CHANGELOG.md`**: add `## [X.Y.Z] - YYYY-MM-DD`.

2) **Run formatting and tests**:

```bash
make fmt
make test
```

If needed, run Flutter/Dart tests separately in `frontend/` and `dart_packages/`.

3) **Create the release via Makefile** (it creates a commit + tag and pushes):

```bash
make release VERSION=X.Y.Z
```

Important:
- working tree must be clean;
- version must be exactly `X.Y.Z`.

4) **CI does the rest**: the tag triggers the desktop workflow and creates a
GitHub Release.

## What goes into GitHub Releases (and how notes are generated)

Release notes are generated automatically:
- the `CHANGELOG.md` section for `X.Y.Z` is used;
- a “Download” table is appended with per-OS artifacts.

If release notes are missing anything, fix `CHANGELOG.md` (it’s the source of
truth).

## Post-release checklist

- The release has artifacts:
  - macOS: `NetworkDebugger-X.Y.Z-macos-arm64.dmg` and
    `NetworkDebugger-X.Y.Z-macos-x86_64.dmg`
  - Windows: `NetworkDebugger-X.Y.Z-windows-amd64.zip`
  - Linux: `network-debugger_X.Y.Z_amd64.deb` and/or
    `NetworkDebugger-X.Y.Z-linux-amd64.tar.gz`
- The release description contains the “Download” table at the bottom.
- Desktop app detects the update (on a machine with an older version).

