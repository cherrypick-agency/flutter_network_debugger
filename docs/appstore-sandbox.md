## App Store (Sandbox) and Auto-Integration

In Mac App Store, sandbox restrictions apply, preventing the app from:

- Installing certificates in System Keychain (requires privileges outside sandbox)
- Changing system network settings (enabling HTTP/HTTPS proxy globally)

This means that our auto-integration scenario (installing dev CA and enabling system proxy) is not available in MAS builds.

### What is possible in MAS

- Display step-by-step instructions and provide "Download CA (.crt)" button
- Open system settings (links/hints), but without changing parameters on behalf of the app
- Offer PAC file and connection instructions (manual user installation)

### Recommended scenario for MAS

1. User downloads CA (.crt) via UI
2. Opens Keychain Access → System → imports CA → Trust: Always Trust
3. Enables system HTTP/HTTPS proxy manually: System Settings → Network → Wi‑Fi → Proxies

Reverse rollback (disabling proxy/removing CA) is also performed manually by the user.

### For non-MAS (developer/enterprise)

In non-sandbox builds, automation can be performed via `osascript` (AppleScript) with admin confirmation. In the app this is implemented in UI (Integrations page) with buttons "Auto-setup", "Disable proxy", "Remove dev CA".


