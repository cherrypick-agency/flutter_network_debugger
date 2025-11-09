//go:build !darwin
// +build !darwin

package helper

import "fmt"

type stubInstaller struct{}

func NewInstaller() Installer {
	return &stubInstaller{}
}

func (i *stubInstaller) IsInstalled() bool {
	return false
}

func (i *stubInstaller) Install(helperBinaryPath string) error {
	return fmt.Errorf("helper tool not supported on this platform")
}

func (i *stubInstaller) Uninstall() error {
	return fmt.Errorf("helper tool not supported on this platform")
}

func (i *stubInstaller) GetVersion() string {
	return ""
}
