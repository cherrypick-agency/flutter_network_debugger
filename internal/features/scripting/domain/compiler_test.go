package domain

import (
	"testing"
)

// Composer 1.
func TestCompileRequest(t *testing.T) {
	req := CompileRequest{
		SourceCode:   "fn main() {}",
		Dependencies: map[string]string{"Cargo.toml": "[package]"},
		Language:     "rust",
		Optimize:     true,
		Debug:        false,
	}

	if req.SourceCode == "" {
		t.Error("SourceCode should not be empty")
	}

	if len(req.Dependencies) == 0 {
		t.Error("Dependencies should not be empty")
	}

	if req.Language == "" {
		t.Error("Language should not be empty")
	}

	if !req.Optimize {
		t.Error("Optimize should be true")
	}

	if req.Debug {
		t.Error("Debug should be false")
	}
}

// Composer 1.
func TestCompileResult(t *testing.T) {
	result := CompileResult{
		WASMBinary: []byte{1, 2, 3, 4, 5},
		Logs:       []string{"log1", "log2"},
		Warnings:   []string{"warning1"},
		Duration:   1000,
		WASMSize:   5,
	}

	if len(result.WASMBinary) == 0 {
		t.Error("WASMBinary should not be empty")
	}

	if len(result.Logs) == 0 {
		t.Error("Logs should not be empty")
	}

	if len(result.Warnings) == 0 {
		t.Error("Warnings should not be empty")
	}

	if result.Duration == 0 {
		t.Error("Duration should not be zero")
	}

	if result.WASMSize == 0 {
		t.Error("WASMSize should not be zero")
	}
}
