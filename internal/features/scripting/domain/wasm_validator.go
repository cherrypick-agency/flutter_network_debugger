package domain

// WASMValidator is the port/interface for WASM module validation
// This follows Clean Architecture - this is a PORT that adapters implement
type WASMValidator interface {
	// ValidateWASM validates that a WASM binary exports required functions
	// Returns nil if valid, error with details if invalid
	// The error message should include helpful fix suggestions
	ValidateWASM(wasmBytes []byte, language string) error
}
