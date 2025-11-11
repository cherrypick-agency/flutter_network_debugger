//go:build !darwin
// +build !darwin

package helper

import (
	"testing"
)

// Composer 1.
func TestNewHelperClient_OtherPlatform(t *testing.T) {
	client := NewHelperClient()

	if client == nil {
		t.Fatal("NewHelperClient returned nil")
	}

	stub, ok := client.(*stubHelperClient)
	if !ok {
		t.Fatal("NewHelperClient should return stubHelperClient on non-darwin platforms")
	}

	if stub == nil {
		t.Fatal("stubHelperClient is nil")
	}
}

// Composer 1.
func TestStubHelperClient_IsRunning(t *testing.T) {
	client := &stubHelperClient{}

	if client.IsRunning() {
		t.Error("IsRunning() should return false on non-darwin platforms")
	}
}

// Composer 1.
func TestStubHelperClient_DetectProcess(t *testing.T) {
	client := &stubHelperClient{}

	info, err := client.DetectProcess(8080)

	if err == nil {
		t.Error("DetectProcess() should return error on non-darwin platforms")
	}

	if info != nil {
		t.Error("DetectProcess() should return nil ProcessInfo on non-darwin platforms")
	}
}

// Composer 1.
func TestStubHelperClient_ExtractIcon(t *testing.T) {
	client := &stubHelperClient{}

	icon, err := client.ExtractIcon(1234)

	if err == nil {
		t.Error("ExtractIcon() should return error on non-darwin platforms")
	}

	if icon != nil {
		t.Error("ExtractIcon() should return nil AppIcon on non-darwin platforms")
	}
}

// Composer 1.
func TestStubHelperClient_Ping(t *testing.T) {
	client := &stubHelperClient{}

	err := client.Ping()

	if err == nil {
		t.Error("Ping() should return error on non-darwin platforms")
	}
}

// Composer 1.
func TestStubHelperClient_Close(t *testing.T) {
	client := &stubHelperClient{}

	err := client.Close()

	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}
