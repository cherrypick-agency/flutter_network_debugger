//go:build !darwin
// +build !darwin

package helper

import (
	"testing"
)

// Composer 1.
func TestNewInstaller_OtherPlatform(t *testing.T) {
	installer := NewInstaller()

	if installer == nil {
		t.Fatal("NewInstaller returned nil")
	}

	stub, ok := installer.(*stubInstaller)
	if !ok {
		t.Fatal("NewInstaller should return *stubInstaller on non-darwin platforms")
	}

	if stub == nil {
		t.Fatal("stubInstaller is nil")
	}
}

// Composer 1.
func TestStubInstaller_IsInstalled(t *testing.T) {
	installer := &stubInstaller{}

	if installer.IsInstalled() {
		t.Error("IsInstalled() should return false on non-darwin platforms")
	}
}

// Composer 1.
func TestStubInstaller_Install(t *testing.T) {
	installer := &stubInstaller{}

	err := installer.Install("/path/to/binary")
	if err == nil {
		t.Error("Install() should return error on non-darwin platforms")
	}

	expectedError := "helper tool not supported on this platform"
	if err.Error() != expectedError {
		t.Errorf("Install() error = %q, want %q", err.Error(), expectedError)
	}
}

// Composer 1.
func TestStubInstaller_Uninstall(t *testing.T) {
	installer := &stubInstaller{}

	err := installer.Uninstall()
	if err == nil {
		t.Error("Uninstall() should return error on non-darwin platforms")
	}

	expectedError := "helper tool not supported on this platform"
	if err.Error() != expectedError {
		t.Errorf("Uninstall() error = %q, want %q", err.Error(), expectedError)
	}
}

// Composer 1.
func TestStubInstaller_GetVersion(t *testing.T) {
	installer := &stubInstaller{}

	version := installer.GetVersion()
	if version != "" {
		t.Errorf("GetVersion() = %q, want empty string", version)
	}
}

// Composer 1.
func TestStubInstaller_ImplementsInterface(t *testing.T) {
	installer := NewInstaller()

	if _, ok := installer.(Installer); !ok {
		t.Error("stubInstaller should implement Installer interface")
	}
}
