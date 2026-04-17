//go:build darwin
// +build darwin

package helper

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

const (
	helperLabel       = "com.networkdebugger.helper"
	helperInstallPath = "/Library/PrivilegedHelperTools/com.networkdebugger.helper"
	plistPath         = "/Library/LaunchDaemons/com.networkdebugger.helper.plist"
	helperVersion     = "1.0.0"
)

var (
	darwinExecCommand        = exec.Command
	darwinExecCommandContext = exec.CommandContext
)

var plistContent = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>com.networkdebugger.helper</string>
	<key>ProgramArguments</key>
	<array>
		<string>/Library/PrivilegedHelperTools/com.networkdebugger.helper</string>
	</array>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>StandardOutPath</key>
	<string>/var/log/network-debugger-helper.log</string>
	<key>StandardErrorPath</key>
	<string>/var/log/network-debugger-helper.err</string>
</dict>
</plist>`

// darwinInstaller - installer for macOS
type darwinInstaller struct{}

// NewInstaller - create installer for the current platform
func NewInstaller() Installer {
	return &darwinInstaller{}
}

// IsInstalled - check if helper is installed
func (i *darwinInstaller) IsInstalled() bool {
	// Check that plist file exists
	if _, err := os.Stat(plistPath); os.IsNotExist(err) {
		return false
	}

	// Check that binary exists
	if _, err := os.Stat(helperInstallPath); os.IsNotExist(err) {
		return false
	}

	return true
}

// Install - install helper tool
func (i *darwinInstaller) Install(helperBinaryPath string) error {
	// Check that source binary exists
	if _, err := os.Stat(helperBinaryPath); os.IsNotExist(err) {
		return fmt.Errorf("helper binary not found: %s", helperBinaryPath)
	}

	// Create temporary plist file
	tmpPlist := "/tmp/com.networkdebugger.helper.plist"
	if err := os.WriteFile(tmpPlist, []byte(plistContent), 0644); err != nil {
		return fmt.Errorf("failed to create temp plist: %w", err)
	}
	defer os.Remove(tmpPlist)

	// Installation command (with password prompt via osascript)
	script := fmt.Sprintf(`do shell script "cp '%s' '%s' && chmod 755 '%s' && cp '%s' '%s' && chmod 644 '%s' && launchctl load -w '%s'" with administrator privileges`,
		helperBinaryPath, helperInstallPath, helperInstallPath,
		tmpPlist, plistPath, plistPath,
		plistPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd := darwinExecCommandContext(ctx, "osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("installation timeout: command took too long")
		}
		return fmt.Errorf("installation failed: %w (output: %s)", err, string(output))
	}

	return nil
}

// Uninstall - uninstall helper tool
func (i *darwinInstaller) Uninstall() error {
	// Unload daemon (doesn't require password if daemon is already running as root)
	cmd := darwinExecCommand("launchctl", "unload", plistPath)
	_ = cmd.Run() // ignore error if already unloaded

	// Delete files (requires password)
	script := fmt.Sprintf(`do shell script "rm -f '%s' '%s'" with administrator privileges`,
		plistPath, helperInstallPath)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cmd = darwinExecCommandContext(ctx, "osascript", "-e", script)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("uninstallation timeout: command took too long")
		}
		return fmt.Errorf("uninstallation failed: %w (output: %s)", err, string(output))
	}

	return nil
}

// GetVersion - get helper version
func (i *darwinInstaller) GetVersion() string {
	if !i.IsInstalled() {
		return ""
	}

	// Try to get version via ping
	client := NewHelperClient()
	defer client.Close()

	if client.IsRunning() {
		return helperVersion // hardcoded for now
	}

	return ""
}
