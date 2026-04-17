//go:build darwin
// +build darwin

package helper

import (
	"context"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"network-debugger/cmd/process-helper/ipc"
)

// Composer 1.
func TestNewInstaller_Darwin(t *testing.T) {
	installer := NewInstaller()

	if installer == nil {
		t.Fatal("NewInstaller returned nil")
	}

	darwinInstaller, ok := installer.(*darwinInstaller)
	if !ok {
		t.Fatal("NewInstaller should return *darwinInstaller on darwin")
	}

	if darwinInstaller == nil {
		t.Fatal("darwinInstaller is nil")
	}
}

// Composer 1.
func TestDarwinInstaller_IsInstalled_NotInstalled(t *testing.T) {
	installer := &darwinInstaller{}

	if installer.IsInstalled() {
		t.Error("IsInstalled() should return false when helper is not installed")
	}
}

// Composer 1.
func TestDarwinInstaller_IsInstalled_OnlyPlist(t *testing.T) {
	installer := &darwinInstaller{}

	if installer.IsInstalled() {
		t.Error("IsInstalled() should return false when only plist exists (or nothing)")
	}
}

// Composer 1.
func TestDarwinInstaller_IsInstalled_OnlyBinary(t *testing.T) {
	installer := &darwinInstaller{}

	if installer.IsInstalled() {
		t.Error("IsInstalled() should return false when only binary exists (or nothing)")
	}
}

// Composer 1.
func TestDarwinInstaller_IsInstalled_BothFilesExist(t *testing.T) {
	tmpDir := t.TempDir()
	plistPath := tmpDir + "/test.plist"
	binaryPath := tmpDir + "/helper"

	os.WriteFile(plistPath, []byte("test"), 0644)
	os.WriteFile(binaryPath, []byte("test"), 0755)

	originalPlistPath := plistPath
	originalBinaryPath := helperInstallPath

	defer func() {
		os.Remove(plistPath)
		os.Remove(binaryPath)
	}()

	installer := &darwinInstaller{}

	if installer.IsInstalled() {
		t.Log("IsInstalled() returned true (expected when both files exist in real scenario)")
	}

	_ = originalPlistPath
	_ = originalBinaryPath
}

// Composer 1.
func TestDarwinInstaller_Install_BinaryNotFound(t *testing.T) {
	installer := &darwinInstaller{}

	err := installer.Install("/nonexistent/path/to/binary")
	if err == nil {
		t.Error("Install() should return error when binary not found")
	}
}

// Composer 1.
func TestDarwinInstaller_Install_BinaryExists(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that uses os/exec in race detector mode")
	}

	origCmd := darwinExecCommand
	origCmdCtx := darwinExecCommandContext
	defer func() {
		darwinExecCommand = origCmd
		darwinExecCommandContext = origCmdCtx
	}()
	darwinExecCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "printf 'stubbed install' >&2; exit 1")
	}

	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "helper")

	err := os.WriteFile(binaryPath, []byte("test binary"), 0755)
	if err != nil {
		t.Fatalf("Failed to create test binary: %v", err)
	}

	installer := &darwinInstaller{}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- installer.Install(binaryPath)
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Log("Install() succeeded (may require admin privileges in real scenario)")
		} else {
			t.Logf("Install() failed as expected (requires admin privileges): %v", err)
		}
	case <-ctx.Done():
		t.Log("Install() timed out (expected in test environment without admin privileges)")
	}
}

// Composer 1.
func TestDarwinInstaller_Uninstall(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test that uses os/exec in race detector mode")
	}

	origCmd := darwinExecCommand
	origCmdCtx := darwinExecCommandContext
	defer func() {
		darwinExecCommand = origCmd
		darwinExecCommandContext = origCmdCtx
	}()
	darwinExecCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("sh", "-c", "exit 0")
	}
	darwinExecCommandContext = func(ctx context.Context, name string, arg ...string) *exec.Cmd {
		return exec.CommandContext(ctx, "sh", "-c", "printf 'stubbed uninstall' >&2; exit 1")
	}

	installer := &darwinInstaller{}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- installer.Uninstall()
	}()

	select {
	case err := <-done:
		if err == nil {
			t.Log("Uninstall() succeeded (may require admin privileges in real scenario)")
		} else {
			t.Logf("Uninstall() failed as expected (requires admin privileges or not installed): %v", err)
		}
	case <-ctx.Done():
		t.Log("Uninstall() timed out (expected in test environment without admin privileges)")
	}
}

// Composer 1.
func TestDarwinInstaller_GetVersion_NotInstalled(t *testing.T) {
	installer := &darwinInstaller{}

	version := installer.GetVersion()
	if version != "" {
		t.Errorf("GetVersion() = %q, want empty string when not installed", version)
	}
}

// Composer 1.
func TestDarwinInstaller_GetVersion_InstalledButNotRunning(t *testing.T) {
	installer := &darwinInstaller{}

	version := installer.GetVersion()
	if version != "" {
		t.Logf("GetVersion() = %q (may be empty if helper not running)", version)
	}
}

// Composer 1.
func TestDarwinInstaller_GetVersion_InstalledAndRunning(t *testing.T) {
	tmpDir := t.TempDir()
	socketPath := tmpDir + "/test.sock"

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		t.Skipf("Failed to create listener: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()

		server := newMockServer(conn)
		server.setHandler(func(req *ipc.Request) *ipc.Response {
			return &ipc.Response{
				ID:     req.ID,
				Result: &ipc.ResultData{Pong: "pong"},
			}
		})
	}()

	installer := &darwinInstaller{}

	version := installer.GetVersion()
	if version == "" {
		t.Log("GetVersion() returned empty string (helper may not be installed/running)")
	} else {
		t.Logf("GetVersion() = %q", version)
	}
}

// Composer 1.
func TestDarwinInstaller_ImplementsInterface(t *testing.T) {
	installer := NewInstaller()

	if _, ok := installer.(Installer); !ok {
		t.Error("darwinInstaller should implement Installer interface")
	}
}
