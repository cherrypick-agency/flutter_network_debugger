package extism

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

// TestNewExtismExecutor_WithCache tests executor creation with cache
func TestNewExtismExecutor_WithCache(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "wasm-cache")
	executor, err := NewExtismExecutor(cacheDir)
	if err != nil {
		t.Fatalf("NewExtismExecutor() error = %v, want nil", err)
	}
	defer executor.Close()

	if executor.poolManager == nil {
		t.Error("Expected poolManager to be initialized")
	}

	if executor.compilationCache == nil {
		t.Error("Expected compilationCache to be initialized")
	}
}

// TestNewExtismExecutor_WithoutCache tests executor creation without cache
func TestNewExtismExecutor_WithoutCache(t *testing.T) {
	executor, err := NewExtismExecutor("")
	if err != nil {
		t.Fatalf("NewExtismExecutor() error = %v, want nil", err)
	}
	defer executor.Close()

	if executor.compilationCache != nil {
		t.Error("Expected compilationCache to be nil when cacheDir is empty")
	}
}

// TestExtismExecutor_Runtime tests Runtime() method
func TestExtismExecutor_Runtime(t *testing.T) {
	executor, err := NewExtismExecutor("")
	if err != nil {
		t.Fatalf("NewExtismExecutor() error = %v", err)
	}
	defer executor.Close()

	if executor.Runtime() != domain.RuntimeExtism {
		t.Errorf("Runtime() = %v, want %v", executor.Runtime(), domain.RuntimeExtism)
	}
}

// TestExtismExecutor_Validate_EmptyCode tests Validate with empty code
func TestExtismExecutor_Validate_EmptyCode(t *testing.T) {
	executor, err := NewExtismExecutor("")
	if err != nil {
		t.Fatalf("NewExtismExecutor() error = %v", err)
	}
	defer executor.Close()

	ctx := context.Background()
	script := domain.Script{
		Code: []byte{},
	}

	err = executor.Validate(ctx, script)
	if err != nil {
		t.Errorf("Validate() with empty code should return nil, got: %v", err)
	}
}

// TestExtismExecutor_Validate_InvalidWASM tests Validate with invalid WASM
func TestExtismExecutor_Validate_InvalidWASM(t *testing.T) {
	executor, err := NewExtismExecutor("")
	if err != nil {
		t.Fatalf("NewExtismExecutor() error = %v", err)
	}
	defer executor.Close()

	ctx := context.Background()
	script := domain.Script{
		Code: []byte("invalid wasm"),
	}

	err = executor.Validate(ctx, script)
	if err == nil {
		t.Error("Validate() with invalid WASM should return error")
	}
}

// TestExtismExecutor_Validate_ValidWASM tests Validate with valid WASM
func TestExtismExecutor_Validate_ValidWASM(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)
	ctx := context.Background()
	script := createSuccessScript(t)

	err := executor.Validate(ctx, script)
	if err != nil {
		t.Errorf("Validate() with valid WASM should return nil, got: %v", err)
	}
}

// TestExtismExecutor_Execute_Success tests Execute with successful script
func TestExtismExecutor_Execute_Success(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)
	ctx := context.Background()

	script := createSuccessScript(t)
	req := domain.ExecutionRequest{
		Script: script,
		Input:  []byte(`{"test": "data"}`),
	}

	result, err := executor.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	if result.Error != "" {
		t.Errorf("Execute() result.Error = %q, want empty", result.Error)
	}

	if result.Duration == 0 {
		t.Error("Execute() result.Duration should be > 0")
	}
}

// TestExtismExecutor_Execute_Timeout tests Execute with timeout
func TestExtismExecutor_Execute_Timeout(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)

	script := createSuccessScript(t)
	script.Config.TimeoutMs = 1

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	req := domain.ExecutionRequest{
		Script: script,
		Input:  []byte(`{"test": "data"}`),
	}

	time.Sleep(20 * time.Millisecond)

	result, err := executor.Execute(ctx, req)
	// Execute может вернуть ошибку при таймауте на этапе Acquire
	if err != nil {
		// Это нормально - таймаут может произойти при получении плагина из пула
		if !contains(err.Error(), "deadline exceeded") && !contains(err.Error(), "timeout") {
			t.Errorf("Execute() error = %v, expected timeout-related error", err)
		}
		return
	}

	// Или вернуть результат с ошибкой
	if result.Error == "" {
		t.Error("Execute() with timeout should return error in result or as error")
	}
}

// TestExtismExecutor_Execute_Cancellation tests Execute with cancellation
func TestExtismExecutor_Execute_Cancellation(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)

	script := createSuccessScript(t)
	ctx, cancel := context.WithCancel(context.Background())

	req := domain.ExecutionRequest{
		Script: script,
		Input:  []byte(`{"test": "data"}`),
	}

	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()

	result, err := executor.Execute(ctx, req)
	// Execute может вернуть ошибку при отмене на этапе Acquire
	if err != nil {
		// Это нормально - отмена может произойти при получении плагина из пула
		if !contains(err.Error(), "canceled") && !contains(err.Error(), "cancelled") {
			t.Errorf("Execute() error = %v, expected cancellation-related error", err)
		}
		return
	}

	// Или вернуть результат с ошибкой
	if result.Error == "" {
		t.Error("Execute() with cancellation should return error in result or as error")
	}
}

// TestExtismExecutor_RemovePlugin tests RemovePlugin
func TestExtismExecutor_RemovePlugin(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)
	script := createSuccessScript(t)

	pool := executor.poolManager.GetPool(script)
	if pool == nil {
		t.Fatal("Expected pool to be created")
	}

	executor.RemovePlugin(script.ID)

	ctx := context.Background()
	_, err := pool.Acquire(ctx)
	if err == nil {
		t.Error("Expected error after RemovePlugin")
	}
}

// TestExtismExecutor_WarmUpPlugins tests WarmUpPlugins
func TestExtismExecutor_WarmUpPlugins(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)
	ctx := context.Background()

	scripts := []domain.Script{
		createSuccessScript(t),
		createSuccessScript(t),
	}

	err := executor.WarmUpPlugins(ctx, scripts)
	if err != nil {
		t.Errorf("WarmUpPlugins() error = %v, want nil", err)
	}
}

// TestExtismExecutor_WarmUpPlugins_Empty tests WarmUpPlugins with empty list
func TestExtismExecutor_WarmUpPlugins_Empty(t *testing.T) {
	executor, err := NewExtismExecutor("")
	if err != nil {
		t.Fatalf("NewExtismExecutor() error = %v", err)
	}
	defer executor.Close()

	ctx := context.Background()
	err = executor.WarmUpPlugins(ctx, []domain.Script{})
	if err != nil {
		t.Errorf("WarmUpPlugins() with empty list error = %v, want nil", err)
	}
}

// TestExtismExecutor_WarmUpPlugins_Cancellation tests WarmUpPlugins with cancellation
func TestExtismExecutor_WarmUpPlugins_Cancellation(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)
	ctx, cancel := context.WithCancel(context.Background())

	scripts := []domain.Script{
		createSuccessScript(t),
	}

	cancel()

	err := executor.WarmUpPlugins(ctx, scripts)
	if err == nil {
		t.Error("WarmUpPlugins() with cancelled context should return error")
	}
}

// TestExtismExecutor_WarmUpPlugins_DisabledScripts tests WarmUpPlugins skips disabled scripts
func TestExtismExecutor_WarmUpPlugins_DisabledScripts(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)
	ctx := context.Background()

	script := createSuccessScript(t)
	script.Enabled = false

	scripts := []domain.Script{script}

	err := executor.WarmUpPlugins(ctx, scripts)
	if err != nil {
		t.Errorf("WarmUpPlugins() error = %v, want nil", err)
	}
}

// TestExtismExecutor_WarmUpPlugins_NonExtismScripts tests WarmUpPlugins skips non-Extism scripts
func TestExtismExecutor_WarmUpPlugins_NonExtismScripts(t *testing.T) {
	executor, err := NewExtismExecutor("")
	if err != nil {
		t.Fatalf("NewExtismExecutor() error = %v", err)
	}
	defer executor.Close()

	ctx := context.Background()

	script := domain.Script{
		ID:      "test",
		Runtime: domain.RuntimeDart,
		Enabled: true,
	}

	scripts := []domain.Script{script}

	err = executor.WarmUpPlugins(ctx, scripts)
	if err != nil {
		t.Errorf("WarmUpPlugins() error = %v, want nil", err)
	}
}

// TestExtismExecutor_GetMetrics tests GetMetrics
func TestExtismExecutor_GetMetrics(t *testing.T) {
	executor, err := NewExtismExecutor("")
	if err != nil {
		t.Fatalf("NewExtismExecutor() error = %v", err)
	}
	defer executor.Close()

	metrics := executor.GetMetrics()

	if metrics.RunawayGoroutines < 0 {
		t.Error("GetMetrics() RunawayGoroutines should be >= 0")
	}

	if metrics.TotalTimeouts < 0 {
		t.Error("GetMetrics() TotalTimeouts should be >= 0")
	}
}

// TestExtismExecutor_Close tests Close
func TestExtismExecutor_Close(t *testing.T) {
	cacheDir := filepath.Join(t.TempDir(), "wasm-cache")
	executor, err := NewExtismExecutor(cacheDir)
	if err != nil {
		t.Fatalf("NewExtismExecutor() error = %v", err)
	}

	err = executor.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

// TestExtismExecutor_CreatePlugin tests createPlugin
func TestExtismExecutor_CreatePlugin(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)
	ctx := context.Background()
	script := createSuccessScript(t)

	plugin, err := executor.createPlugin(ctx, script)
	if err != nil {
		t.Fatalf("createPlugin() error = %v, want nil", err)
	}

	if plugin == nil {
		t.Error("createPlugin() should return non-nil plugin")
	}

	plugin.Close(ctx)
}

// TestExtismExecutor_CreatePlugin_WithAllowedHosts tests createPlugin with AllowedHosts
func TestExtismExecutor_CreatePlugin_WithAllowedHosts(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)
	ctx := context.Background()
	script := createSuccessScript(t)
	script.Config.AllowedHosts = []string{"api.example.com"}

	plugin, err := executor.createPlugin(ctx, script)
	if err != nil {
		t.Fatalf("createPlugin() error = %v, want nil", err)
	}

	if plugin == nil {
		t.Error("createPlugin() should return non-nil plugin")
	}

	plugin.Close(ctx)
}

// TestExtismExecutor_CreatePlugin_WithMemoryLimit tests createPlugin with MemoryLimit
func TestExtismExecutor_CreatePlugin_WithMemoryLimit(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)
	ctx := context.Background()
	script := createSuccessScript(t)
	script.Config.MemoryLimitMB = 20

	plugin, err := executor.createPlugin(ctx, script)
	if err != nil {
		t.Fatalf("createPlugin() error = %v, want nil", err)
	}

	if plugin == nil {
		t.Error("createPlugin() should return non-nil plugin")
	}

	plugin.Close(ctx)
}

// TestExtismExecutor_Execute_WithEmptyInput tests Execute with empty input
func TestExtismExecutor_Execute_WithEmptyInput(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)
	ctx := context.Background()

	script := createSuccessScript(t)
	req := domain.ExecutionRequest{
		Script: script,
		Input:  []byte(""),
	}

	result, err := executor.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	if result.Duration == 0 {
		t.Error("Execute() result.Duration should be > 0")
	}
}

// TestExtismExecutor_Execute_WithLargeInput tests Execute with large input
func TestExtismExecutor_Execute_WithLargeInput(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)
	ctx := context.Background()

	script := createSuccessScript(t)
	largeInput := make([]byte, 100000)
	for i := range largeInput {
		largeInput[i] = byte(i % 256)
	}

	req := domain.ExecutionRequest{
		Script: script,
		Input:  largeInput,
	}

	result, err := executor.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	if result.Duration == 0 {
		t.Error("Execute() result.Duration should be > 0")
	}
}

// TestExtismExecutor_Execute_WithDifferentScripts tests Execute with different scripts
func TestExtismExecutor_Execute_WithDifferentScripts(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)
	ctx := context.Background()

	scripts := []domain.Script{
		createSuccessScript(t),
		createSuccessScript(t),
	}

	for i, script := range scripts {
		req := domain.ExecutionRequest{
			Script: script,
			Input:  []byte(`{"test": "data"}`),
		}

		result, err := executor.Execute(ctx, req)
		if err != nil {
			t.Fatalf("Execute() for script %d error = %v, want nil", i, err)
		}

		if result.Duration == 0 {
			t.Errorf("Execute() for script %d result.Duration should be > 0", i)
		}
	}
}

// TestExtismExecutor_Execute_WithZeroTimeout tests Execute with zero timeout
func TestExtismExecutor_Execute_WithZeroTimeout(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)
	ctx := context.Background()

	script := createSuccessScript(t)
	script.Config.TimeoutMs = 0

	req := domain.ExecutionRequest{
		Script: script,
		Input:  []byte(`{"test": "data"}`),
	}

	result, err := executor.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	if result.Duration == 0 {
		t.Error("Execute() result.Duration should be > 0")
	}
}

// TestExtismExecutor_Execute_WithVeryShortTimeout tests Execute with very short timeout
func TestExtismExecutor_Execute_WithVeryShortTimeout(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)

	script := createSuccessScript(t)
	script.Config.TimeoutMs = 1

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	req := domain.ExecutionRequest{
		Script: script,
		Input:  []byte(`{"test": "data"}`),
	}

	time.Sleep(100 * time.Millisecond)

	result, err := executor.Execute(ctx, req)
	if err != nil {
		if !contains(err.Error(), "deadline exceeded") && !contains(err.Error(), "timeout") {
			t.Errorf("Execute() error = %v, expected timeout-related error", err)
		}
		return
	}

	if result.Error == "" {
		t.Error("Execute() with very short timeout should return error in result or as error")
	}
}

// TestExtismExecutor_Execute_WithFailingScript tests Execute with failing script
func TestExtismExecutor_Execute_WithFailingScript(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)
	ctx := context.Background()

	script := createFailingScript(t)
	req := domain.ExecutionRequest{
		Script: script,
		Input:  []byte(`{"test": "data"}`),
	}

	result, err := executor.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute() error = %v, want nil", err)
	}

	if result.Error == "" {
		t.Error("Execute() with failing script should return error in result")
	}
}

// TestExtismExecutor_Execute_ConcurrentExecutions tests concurrent executions
func TestExtismExecutor_Execute_ConcurrentExecutions(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)
	ctx := context.Background()

	script := createSuccessScript(t)
	numConcurrent := 5

	results := make(chan domain.ExecutionResult, numConcurrent)
	errors := make(chan error, numConcurrent)

	for i := 0; i < numConcurrent; i++ {
		go func() {
			req := domain.ExecutionRequest{
				Script: script,
				Input:  []byte(`{"test": "data"}`),
			}

			result, err := executor.Execute(ctx, req)
			if err != nil {
				errors <- err
				return
			}
			results <- result
		}()
	}

	successCount := 0
	for i := 0; i < numConcurrent; i++ {
		select {
		case result := <-results:
			if result.Error == "" {
				successCount++
			}
		case <-errors:
		}
	}

	if successCount == 0 {
		t.Error("Expected at least one successful execution")
	}
}
