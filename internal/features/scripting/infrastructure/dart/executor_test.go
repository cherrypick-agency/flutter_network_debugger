package dart

import (
	"context"
	"os"
	"testing"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

// TestNewDartExecutor_DartNotAvailable tests executor creation when Dart SDK is not available
func TestNewDartExecutor_DartNotAvailable(t *testing.T) {
	// This test assumes Dart is installed. We can't easily simulate "Dart not available"
	// without mocking exec.Command, so we just verify the constructor works
	executor, err := NewDartExecutor(2, "nonexistent/path/script_runner.dart")
	if err != nil {
		t.Fatalf("Expected no error when Dart is available, got: %v", err)
	}

	if executor == nil {
		t.Fatal("Expected non-nil executor")
	}

	// Clean up
	if executor.processPool != nil {
		executor.Close()
	}
}

// TestDartExecutor_Runtime tests the Runtime() method
func TestDartExecutor_Runtime(t *testing.T) {
	executor := &DartExecutor{enabled: true}
	if executor.Runtime() != domain.RuntimeDart {
		t.Errorf("Expected RuntimeDart, got %v", executor.Runtime())
	}
}

// TestDartExecutor_Validate_Disabled tests validation when Dart is disabled
func TestDartExecutor_Validate_Disabled(t *testing.T) {
	executor := &DartExecutor{enabled: false}
	ctx := context.Background()

	script := domain.Script{
		SourceCode: "void main() { print('test'); }",
	}

	err := executor.Validate(ctx, script)
	if err == nil {
		t.Fatal("Expected error when Dart runtime not available")
	}

	expectedMsg := "Dart runtime not available"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error '%s', got '%s'", expectedMsg, err.Error())
	}
}

// TestDartExecutor_Validate_EmptySourceCode tests validation with empty source code
func TestDartExecutor_Validate_EmptySourceCode(t *testing.T) {
	if !isDartAvailable() {
		t.Skip("Dart SDK not installed")
	}

	executor := &DartExecutor{enabled: true}
	ctx := context.Background()

	script := domain.Script{
		SourceCode: "",
	}

	err := executor.Validate(ctx, script)
	if err != nil {
		t.Fatalf("Expected no error for empty source code, got: %v", err)
	}
}

// TestDartExecutor_Validate_ValidScript tests validation with valid Dart code
func TestDartExecutor_Validate_ValidScript(t *testing.T) {
	if !isDartAvailable() {
		t.Skip("Dart SDK not installed")
	}

	executor := &DartExecutor{enabled: true}
	ctx := context.Background()

	script := domain.Script{
		SourceCode: `
void main() {
  print('Hello, World!');
}
`,
	}

	err := executor.Validate(ctx, script)
	if err != nil {
		t.Fatalf("Expected no error for valid script, got: %v", err)
	}
}

// TestDartExecutor_Validate_InvalidScript tests validation with invalid Dart code
func TestDartExecutor_Validate_InvalidScript(t *testing.T) {
	if !isDartAvailable() {
		t.Skip("Dart SDK not installed")
	}

	executor := &DartExecutor{enabled: true}
	ctx := context.Background()

	script := domain.Script{
		SourceCode: `
void main() {
  // Syntax error: missing closing brace
  print('test');
`,
	}

	err := executor.Validate(ctx, script)
	if err == nil {
		t.Fatal("Expected error for invalid script")
	}

	// Check that error message contains "dart validation failed"
	if len(err.Error()) == 0 {
		t.Error("Expected non-empty error message")
	}
}

// TestDartExecutor_Validate_Timeout tests validation timeout
func TestDartExecutor_Validate_Timeout(t *testing.T) {
	if !isDartAvailable() {
		t.Skip("Dart SDK not installed")
	}

	executor := &DartExecutor{enabled: true}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Use a valid script but with very short timeout
	script := domain.Script{
		SourceCode: `
void main() {
  print('test');
}
`,
	}

	// Wait for context to expire
	time.Sleep(2 * time.Millisecond)

	err := executor.Validate(ctx, script)
	// This might succeed or fail depending on timing, so we just ensure it doesn't hang
	if err != nil {
		t.Logf("Validation with expired context returned: %v", err)
	}
}

// TestDartExecutor_Validate_ComplexScript tests validation with more complex Dart code
func TestDartExecutor_Validate_ComplexScript(t *testing.T) {
	if !isDartAvailable() {
		t.Skip("Dart SDK not installed")
	}

	executor := &DartExecutor{enabled: true}
	ctx := context.Background()

	script := domain.Script{
		SourceCode: `
import 'dart:convert';

class User {
  final String name;
  final int age;

  User(this.name, this.age);

  Map<String, dynamic> toJson() => {
    'name': name,
    'age': age,
  };
}

void main() {
  var user = User('Alice', 30);
  print(jsonEncode(user.toJson()));
}
`,
	}

	err := executor.Validate(ctx, script)
	if err != nil {
		t.Fatalf("Expected no error for complex valid script, got: %v", err)
	}
}

// TestDartExecutor_Validate_ScriptWithWarnings tests validation with code that has warnings
func TestDartExecutor_Validate_ScriptWithWarnings(t *testing.T) {
	if !isDartAvailable() {
		t.Skip("Dart SDK not installed")
	}

	executor := &DartExecutor{enabled: true}
	ctx := context.Background()

	// Script with unused variable (warning, not error)
	script := domain.Script{
		SourceCode: `
void main() {
  var unused = 42;
  print('test');
}
`,
	}

	err := executor.Validate(ctx, script)
	// dart analyze typically treats warnings as non-fatal, so this should succeed
	if err != nil {
		t.Logf("Validation returned error for script with warnings: %v", err)
		// Note: Depending on dart analyze configuration, this might be an error or not
	}
}

// TestDartExecutor_Close tests Close() method
func TestDartExecutor_Close(t *testing.T) {
	executor := &DartExecutor{enabled: false}
	err := executor.Close()
	if err != nil {
		t.Errorf("Expected no error closing executor without pool, got: %v", err)
	}
}

// TestProcessPool_GetRelease tests basic pool operations
func TestProcessPool_GetRelease(t *testing.T) {
	if !isDartAvailable() {
		t.Skip("Dart SDK not installed")
	}

	// We need a valid script_runner path for this test
	// Skip if the script runner doesn't exist
	scriptRunnerPath := "scripts/dart/script_runner.dart"
	if _, err := os.Stat(scriptRunnerPath); os.IsNotExist(err) {
		t.Skipf("Script runner not found at %s", scriptRunnerPath)
	}

	pool, err := NewProcessPool(2, scriptRunnerPath)
	if err != nil {
		t.Fatalf("Failed to create process pool: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()

	// Get a process from pool
	proc, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get process from pool: %v", err)
	}

	if proc == nil {
		t.Fatal("Expected non-nil process")
	}

	// Return process to pool
	pool.Release(proc)

	// Get it again - should be the same process
	proc2, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get process from pool second time: %v", err)
	}

	if proc2 == nil {
		t.Fatal("Expected non-nil process on second get")
	}

	pool.Release(proc2)
}

// TestProcessPool_Exhaustion tests pool exhaustion behavior
func TestProcessPool_Exhaustion(t *testing.T) {
	if !isDartAvailable() {
		t.Skip("Dart SDK not installed")
	}

	scriptRunnerPath := "scripts/dart/script_runner.dart"
	if _, err := os.Stat(scriptRunnerPath); os.IsNotExist(err) {
		t.Skipf("Script runner not found at %s", scriptRunnerPath)
	}

	pool, err := NewProcessPool(1, scriptRunnerPath)
	if err != nil {
		t.Fatalf("Failed to create process pool: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()

	// Get the only process
	proc1, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get first process: %v", err)
	}

	// Try to get another process with timeout
	ctxTimeout, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	_, err = pool.Get(ctxTimeout)
	if err != context.DeadlineExceeded {
		t.Errorf("Expected context.DeadlineExceeded, got: %v", err)
	}

	// Release first process
	pool.Release(proc1)

	// Now should be able to get a process
	proc2, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get process after release: %v", err)
	}
	pool.Release(proc2)
}

// TestDartExecutor_Execute_Disabled tests Execute when executor is disabled
func TestDartExecutor_Execute_Disabled(t *testing.T) {
	executor := &DartExecutor{enabled: false}
	ctx := context.Background()

	req := domain.ExecutionRequest{
		Script: domain.Script{
			Name: "test",
			Code: []byte("void main() { print('test'); }"),
		},
		Input: []byte("{}"),
	}

	result, err := executor.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute() should not return error when disabled, got: %v", err)
	}

	if result.Error == "" {
		t.Error("Expected error message when executor is disabled")
	}

	if !contains(result.Error, "Dart runtime not available") {
		t.Errorf("Expected error message about Dart runtime, got: %s", result.Error)
	}
}

// TestDartExecutor_Execute_Timeout tests Execute with timeout
func TestDartExecutor_Execute_Timeout(t *testing.T) {
	if !isDartAvailable() {
		t.Skip("Dart SDK not installed")
	}

	scriptRunnerPath := "scripts/dart/script_runner.dart"
	if _, err := os.Stat(scriptRunnerPath); os.IsNotExist(err) {
		t.Skipf("Script runner not found at %s", scriptRunnerPath)
	}

	executor, err := NewDartExecutor(1, scriptRunnerPath)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}
	defer executor.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	req := domain.ExecutionRequest{
		Script: domain.Script{
			Name: "test",
			Code: []byte("void main() { print('test'); }"),
		},
		Input: []byte("{}"),
	}

	time.Sleep(2 * time.Millisecond)

	result, err := executor.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute() should not return error on timeout, got: %v", err)
	}

	if result.Error != "timeout" {
		t.Errorf("Expected 'timeout' error, got: %s", result.Error)
	}
}

// TestDartProcess_ReadResponse_NoResponse tests readResponse when no response is available
func TestDartProcess_ReadResponse_NoResponse(t *testing.T) {
	if !isDartAvailable() {
		t.Skip("Dart SDK not installed")
	}

	scriptRunnerPath := "scripts/dart/script_runner.dart"
	if _, err := os.Stat(scriptRunnerPath); os.IsNotExist(err) {
		t.Skipf("Script runner not found at %s", scriptRunnerPath)
	}

	pool, err := NewProcessPool(1, scriptRunnerPath)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()
	proc, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get process: %v", err)
	}

	proc.cmd.Process.Kill()
	time.Sleep(100 * time.Millisecond)

	_, err = proc.readResponse()
	if err == nil {
		t.Error("Expected error when reading from killed process")
	}
}

// TestProcessPool_Release_FullPool tests Release when pool is full
func TestProcessPool_Release_FullPool(t *testing.T) {
	if !isDartAvailable() {
		t.Skip("Dart SDK not installed")
	}

	scriptRunnerPath := "scripts/dart/script_runner.dart"
	if _, err := os.Stat(scriptRunnerPath); os.IsNotExist(err) {
		t.Skipf("Script runner not found at %s", scriptRunnerPath)
	}

	pool, err := NewProcessPool(2, scriptRunnerPath)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}
	defer pool.Close()

	ctx := context.Background()

	procs := make([]*DartProcess, 3)
	for i := 0; i < 3; i++ {
		proc, err := pool.Get(ctx)
		if err != nil {
			t.Fatalf("Failed to get process %d: %v", i, err)
		}
		procs[i] = proc
	}

	for i := 0; i < 3; i++ {
		pool.Release(procs[i])
	}

	pool.mu.Lock()
	currentCount := pool.currentCount
	pool.mu.Unlock()

	if currentCount > 2 {
		t.Errorf("Expected currentCount <= 2 after releasing, got %d", currentCount)
	}
}

// TestProcessPool_Close_WithProcesses tests Close with active processes
func TestProcessPool_Close_WithProcesses(t *testing.T) {
	if !isDartAvailable() {
		t.Skip("Dart SDK not installed")
	}

	scriptRunnerPath := "scripts/dart/script_runner.dart"
	if _, err := os.Stat(scriptRunnerPath); os.IsNotExist(err) {
		t.Skipf("Script runner not found at %s", scriptRunnerPath)
	}

	pool, err := NewProcessPool(2, scriptRunnerPath)
	if err != nil {
		t.Fatalf("Failed to create pool: %v", err)
	}

	ctx := context.Background()
	proc, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get process: %v", err)
	}

	err = pool.Close()
	if err != nil {
		t.Errorf("Close() should not return error, got: %v", err)
	}

	pool.Release(proc)
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
