package extism

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

// testWASMFixtures contains paths to test WASM files
// Using path relative to this file's location (../../../../)
var testWASMFixtures = struct {
	Noop      string
	AddHeader string
	Failing   string
	Timeout   string
}{
	Noop:      "../../../../e2e/testdata/scripts/wasm/noop.wasm",
	AddHeader: "../../../../e2e/testdata/scripts/wasm/add_header.wasm",
	Failing:   "../../../../e2e/testdata/scripts/wasm/failing.wasm",
	Timeout:   "../../../../e2e/testdata/scripts/wasm/timeout.wasm",
}

// createTestScript creates a test script with WASM code
func createTestScript(t *testing.T, wasmPath string) domain.Script {
	t.Helper()

	// Load WASM file
	code, err := os.ReadFile(wasmPath)
	if err != nil {
		t.Fatalf("failed to load WASM file %s: %v", wasmPath, err)
	}

	return domain.Script{
		ID:          fmt.Sprintf("test-script-%d", time.Now().UnixNano()),
		Name:        "Test Script",
		Description: "Test script for pooling",
		Runtime:     domain.RuntimeExtism,
		Code:        code,
		Language:    "rust",
		TriggerType: domain.TriggerRequest,
		Priority:    10,
		Enabled:     true,
		Config: domain.ScriptConfig{
			TimeoutMs:     5000,
			MemoryLimitMB: 10,
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// createSuccessScript creates a script that always succeeds
func createSuccessScript(t *testing.T) domain.Script {
	t.Helper()
	return createTestScript(t, testWASMFixtures.Noop)
}

// createFailingScript creates a script that compiles but panics on execution
func createFailingScript(t *testing.T) domain.Script {
	t.Helper()
	return createTestScript(t, testWASMFixtures.Failing)
}

// createTimeoutScript creates a script that spins until the runtime deadline is hit.
func createTimeoutScript(t *testing.T) domain.Script {
	t.Helper()
	return createTestScript(t, testWASMFixtures.Timeout)
}

// createSlowScript creates a script that takes time to execute
func createSlowScript(t *testing.T, delayMs int) domain.Script {
	t.Helper()
	// For now, use noop.wasm - in real scenario we'd need a custom slow WASM
	return createTestScript(t, testWASMFixtures.Noop)
}

// createTestExecutor creates ExtismExecutor for testing
func createTestExecutor(t *testing.T) *ExtismExecutor {
	t.Helper()

	// Use temporary cache directory for tests
	cacheDir := filepath.Join(t.TempDir(), "wasm-cache")

	executor, err := NewExtismExecutor(cacheDir)
	if err != nil {
		t.Fatalf("failed to create test executor: %v", err)
	}

	t.Cleanup(func() {
		_ = executor.Close()
	})

	return executor
}

// createTestPool creates PluginPool for testing
func createTestPool(t *testing.T, script domain.Script) *PluginPool {
	t.Helper()

	executor := createTestExecutor(t)

	pool := NewPluginPool(script, executor.createPlugin)

	t.Cleanup(func() {
		_ = pool.Close()
	})

	return pool
}

// executeScript executes a script and returns result
// Uses default context (no timeout) for most tests
func executeScript(t *testing.T, pool *PluginPool, input []byte) ([]byte, error) {
	t.Helper()
	return executeScriptWithContext(t, pool, input, context.Background())
}

// executeScriptWithContext is like executeScript but allows custom context (e.g., with timeout)
func executeScriptWithContext(t *testing.T, pool *PluginPool, input []byte, ctx context.Context) ([]byte, error) {
	t.Helper()

	instance, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}

	var execErr error
	defer func() {
		pool.Release(instance, execErr)
	}()

	var output []byte
	_, output, execErr = instance.plugin.CallWithContext(ctx, "process", input)

	return output, execErr
}

// waitForCircuitOpen waits for circuit breaker to open (max 5 seconds)
func waitForCircuitOpen(t *testing.T, pool *PluginPool) {
	t.Helper()

	start := time.Now()
	for time.Since(start) < 5*time.Second {
		metrics := pool.GetMetrics()
		if metrics.CircuitState == CircuitOpen {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}

	t.Fatal("circuit breaker did not open within 5 seconds")
}

// waitForCircuitHalfOpen waits for circuit breaker to enter half-open state
func waitForCircuitHalfOpen(t *testing.T, pool *PluginPool, timeout time.Duration) {
	t.Helper()

	start := time.Now()
	for time.Since(start) < timeout {
		metrics := pool.GetMetrics()
		if metrics.CircuitState == CircuitHalfOpen {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	t.Fatalf("circuit breaker did not enter half-open state within %v", timeout)
}

// waitForCleanup waits for at least one cleanup cycle to complete
func waitForCleanup(t *testing.T) {
	t.Helper()
	// CleanupInterval is 1 minute, wait a bit longer to ensure cycle completed
	time.Sleep(70 * time.Second)
}

// assertMetrics asserts expected metrics values
func assertMetrics(t *testing.T, pool *PluginPool, expected PoolMetrics) {
	t.Helper()

	metrics := pool.GetMetrics()

	if expected.Hits > 0 && metrics.Hits != expected.Hits {
		t.Errorf("expected hits=%d, got %d", expected.Hits, metrics.Hits)
	}

	if expected.Misses > 0 && metrics.Misses != expected.Misses {
		t.Errorf("expected misses=%d, got %d", expected.Misses, metrics.Misses)
	}

	if expected.Evictions > 0 && metrics.Evictions != expected.Evictions {
		t.Errorf("expected evictions=%d, got %d", expected.Evictions, metrics.Evictions)
	}

	if expected.CircuitState != metrics.CircuitState {
		t.Errorf("expected circuitState=%d, got %d", expected.CircuitState, metrics.CircuitState)
	}

	if expected.IdleCount > 0 && metrics.IdleCount != expected.IdleCount {
		t.Errorf("expected idleCount=%d, got %d", expected.IdleCount, metrics.IdleCount)
	}

	if expected.ActiveCount > 0 && metrics.ActiveCount != expected.ActiveCount {
		t.Errorf("expected activeCount=%d, got %d", expected.ActiveCount, metrics.ActiveCount)
	}
}

// createMinimalWASM creates a minimal valid WASM module for testing
// This is used when we need a WASM that's valid but we don't care about its behavior
func createMinimalWASM() []byte {
	// Minimal WASM module (just magic number + version)
	// This is NOT a valid executable WASM, but useful for some tests
	// For real tests, use testWASMFixtures
	wasmMagic := []byte{0x00, 0x61, 0x73, 0x6D}   // \0asm
	wasmVersion := []byte{0x01, 0x00, 0x00, 0x00} // version 1
	return append(wasmMagic, wasmVersion...)
}

// b64Encode encodes bytes to base64 string (helper for tests)
func b64Encode(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

// b64Decode decodes base64 string to bytes (helper for tests)
func b64Decode(t *testing.T, s string) []byte {
	t.Helper()
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		t.Fatalf("failed to decode base64: %v", err)
	}
	return data
}

// skipIfNoWASMFixtures skips test if WASM fixtures are not available
func skipIfNoWASMFixtures(t *testing.T) {
	t.Helper()

	if _, err := os.Stat(testWASMFixtures.Noop); os.IsNotExist(err) {
		wd, _ := os.Getwd()
		t.Skipf("WASM fixtures not available at %s (cwd: %s), skipping test", testWASMFixtures.Noop, wd)
	}
}
