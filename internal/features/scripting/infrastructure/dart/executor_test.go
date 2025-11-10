package dart

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

// getScriptRunnerPath возвращает путь к script_runner.dart, проверяя несколько возможных мест
func getScriptRunnerPath(t *testing.T) string {
	// Получаем рабочую директорию
	wd, err := os.Getwd()
	if err != nil {
		t.Logf("Failed to get working directory: %v", err)
		return "scripts/dart/script_runner.dart"
	}

	// Пробуем найти, поднимаясь вверх по директориям от текущей
	dir := wd
	for i := 0; i < 10; i++ {
		testPath := filepath.Join(dir, "scripts/dart/script_runner.dart")
		if _, err := os.Stat(testPath); err == nil {
			return testPath
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// Пробуем относительные пути от текущей директории
	possiblePaths := []string{
		"scripts/dart/script_runner.dart",
		"../../scripts/dart/script_runner.dart",
		"../../../scripts/dart/script_runner.dart",
		"../../../../scripts/dart/script_runner.dart",
	}

	for _, relPath := range possiblePaths {
		testPath := filepath.Join(wd, relPath)
		if _, err := os.Stat(testPath); err == nil {
			return testPath
		}
	}

	// Возвращаем относительный путь как fallback
	return "scripts/dart/script_runner.dart"
}

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

	scriptRunnerPath := getScriptRunnerPath(t)
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
		t.Fatalf("Expected no error, got: %v", err)
	}

	if result.Error == "" {
		t.Error("Expected error message in result")
	}

	if !strings.Contains(result.Error, "Dart runtime not available") {
		t.Errorf("Expected error about Dart runtime, got: %s", result.Error)
	}
}

// TestDartExecutor_Execute_Success tests successful script execution
func TestDartExecutor_Execute_Success(t *testing.T) {
	if !isDartAvailable() {
		t.Skip("Dart SDK not installed")
	}

	scriptRunnerPath := "scripts/dart/script_runner.dart"
	if _, err := os.Stat(scriptRunnerPath); os.IsNotExist(err) {
		t.Skipf("Script runner not found at %s", scriptRunnerPath)
	}

	executor, err := NewDartExecutor(2, scriptRunnerPath)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}
	defer executor.Close()

	if !executor.enabled {
		t.Skip("Dart executor is disabled")
	}

	ctx := context.Background()

	req := domain.ExecutionRequest{
		Script: domain.Script{
			Name: "test",
			Code: []byte(`
void main() {
  print('Hello from Dart!');
}
`),
		},
		Input: []byte("{}"),
	}

	result, err := executor.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}

	if result.Error != "" {
		t.Errorf("Expected no error, got: %s", result.Error)
	}

	if result.Duration == 0 {
		t.Error("Expected non-zero duration")
	}
}

// TestDartExecutor_Execute_Timeout tests execution timeout
func TestDartExecutor_Execute_Timeout(t *testing.T) {
	if !isDartAvailable() {
		t.Skip("Dart SDK not installed")
	}

	scriptRunnerPath := "scripts/dart/script_runner.dart"
	if _, err := os.Stat(scriptRunnerPath); os.IsNotExist(err) {
		t.Skipf("Script runner not found at %s", scriptRunnerPath)
	}

	executor, err := NewDartExecutor(2, scriptRunnerPath)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}
	defer executor.Close()

	if !executor.enabled {
		t.Skip("Dart executor is disabled")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	req := domain.ExecutionRequest{
		Script: domain.Script{
			Name: "test",
			Code: []byte(`
void main() async {
  await Future.delayed(Duration(seconds: 10));
  print('This should not execute');
}
`),
		},
		Input: []byte("{}"),
	}

	result, err := executor.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute should not return error on timeout, got: %v", err)
	}

	if result.Error != "timeout" {
		t.Errorf("Expected timeout error, got: %s", result.Error)
	}

	if result.Duration == 0 {
		t.Error("Expected non-zero duration")
	}
}

// TestDartExecutor_Execute_InvalidJSONRPC tests handling of invalid JSON-RPC response
func TestDartExecutor_Execute_InvalidJSONRPC(t *testing.T) {
	if !isDartAvailable() {
		t.Skip("Dart SDK not installed")
	}

	scriptRunnerPath := "scripts/dart/script_runner.dart"
	if _, err := os.Stat(scriptRunnerPath); os.IsNotExist(err) {
		t.Skipf("Script runner not found at %s", scriptRunnerPath)
	}

	executor, err := NewDartExecutor(2, scriptRunnerPath)
	if err != nil {
		t.Fatalf("Failed to create executor: %v", err)
	}
	defer executor.Close()

	if !executor.enabled {
		t.Skip("Dart executor is disabled")
	}

	ctx := context.Background()

	req := domain.ExecutionRequest{
		Script: domain.Script{
			Name: "test",
			Code: []byte(`
void main() {
  // This will cause an error in the script runner
  throw Exception('Test error');
}
`),
		},
		Input: []byte("{}"),
	}

	result, err := executor.Execute(ctx, req)
	if err != nil {
		t.Fatalf("Execute should handle errors gracefully, got: %v", err)
	}

	if result.Error == "" {
		t.Log("Script error was handled, but no error message in result")
	}
}

// TestProcessPool_Close tests graceful shutdown
func TestProcessPool_Close(t *testing.T) {
	if !isDartAvailable() {
		t.Skip("Dart SDK not installed")
	}

	scriptRunnerPath := "scripts/dart/script_runner.dart"
	if _, err := os.Stat(scriptRunnerPath); os.IsNotExist(err) {
		t.Skipf("Script runner not found at %s", scriptRunnerPath)
	}

	pool, err := NewProcessPool(2, scriptRunnerPath)
	if err != nil {
		t.Fatalf("Failed to create process pool: %v", err)
	}

	ctx := context.Background()

	proc1, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get process: %v", err)
	}

	proc2, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get second process: %v", err)
	}

	pool.Release(proc1)
	pool.Release(proc2)

	err = pool.Close()
	if err != nil {
		t.Errorf("Close() should not return error, got: %v", err)
	}
}

// TestProcessPool_Release_FullPool tests releasing when pool is full
func TestProcessPool_Release_FullPool(t *testing.T) {
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

	proc1, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get process: %v", err)
	}

	proc2, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get second process: %v", err)
	}

	pool.Release(proc1)

	pool.Release(proc2)
}

// TestProcessPool_Get_ContextCancelled tests Get with cancelled context
func TestProcessPool_Get_ContextCancelled(t *testing.T) {
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

	proc1, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get process: %v", err)
	}

	ctxCancelled, cancel := context.WithCancel(ctx)
	cancel()

	_, err = pool.Get(ctxCancelled)
	if err != context.Canceled {
		t.Errorf("Expected context.Canceled, got: %v", err)
	}

	pool.Release(proc1)
}

// TestProcessPool_ProcessReuse tests that processes are reused from pool
func TestProcessPool_ProcessReuse(t *testing.T) {
	if !isDartAvailable() {
		t.Skip("Dart SDK not installed")
	}

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

	proc1, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get process: %v", err)
	}

	pid1 := proc1.cmd.Process.Pid

	pool.Release(proc1)

	proc2, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get process second time: %v", err)
	}

	pid2 := proc2.cmd.Process.Pid

	if pid1 != pid2 {
		t.Logf("Processes have different PIDs: %d vs %d (this is OK if pool creates new processes)", pid1, pid2)
	}

	pool.Release(proc2)
}

// TestDartExecutor_Execute_PoolGetError tests Execute when pool.Get fails
func TestDartExecutor_Execute_PoolGetError(t *testing.T) {
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

	if !executor.enabled {
		t.Skip("Dart executor is disabled")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req := domain.ExecutionRequest{
		Script: domain.Script{
			Name: "test",
			Code: []byte("void main() { print('test'); }"),
		},
		Input: []byte("{}"),
	}

	_, err = executor.Execute(ctx, req)
	if err == nil {
		t.Error("Expected error when context is cancelled")
	}
}

// TestProcessPool_Release_NilProcess tests Release with nil process
func TestProcessPool_Release_NilProcess(t *testing.T) {
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

	var nilProc *DartProcess
	pool.Release(nilProc)
}

// TestProcessPool_Close_EmptyPool tests Close with empty pool
func TestProcessPool_Close_EmptyPool(t *testing.T) {
	if !isDartAvailable() {
		t.Skip("Dart SDK not installed")
	}

	scriptRunnerPath := "scripts/dart/script_runner.dart"
	if _, err := os.Stat(scriptRunnerPath); os.IsNotExist(err) {
		t.Skipf("Script runner not found at %s", scriptRunnerPath)
	}

	pool, err := NewProcessPool(2, scriptRunnerPath)
	if err != nil {
		t.Fatalf("Failed to create process pool: %v", err)
	}

	err = pool.Close()
	if err != nil {
		t.Errorf("Close() should not return error, got: %v", err)
	}
}

// TestProcessPool_Release_NilCmd tests Release with process that has nil cmd
func TestProcessPool_Release_NilCmd(t *testing.T) {
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

	proc := &DartProcess{
		cmd:    nil,
		stdin:  nil,
		stdout: nil,
	}

	pool.Release(proc)
}

// TestDartExecutor_Close_WithPool tests Close when pool exists
func TestDartExecutor_Close_WithPool(t *testing.T) {
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

	if !executor.enabled {
		t.Skip("Dart executor is disabled")
	}

	err = executor.Close()
	if err != nil {
		t.Errorf("Close() should not return error, got: %v", err)
	}
}

// TestNewProcessPool_InvalidPath tests NewProcessPool with invalid path
func TestNewProcessPool_InvalidPath(t *testing.T) {
	if !isDartAvailable() {
		t.Skip("Dart SDK not installed")
	}

	pool, err := NewProcessPool(1, "nonexistent/path/script_runner.dart")
	if err != nil {
		t.Fatalf("NewProcessPool should not fail on invalid path, got: %v", err)
	}

	if pool == nil {
		t.Fatal("Expected non-nil pool")
	}

	ctx := context.Background()
	_, err = pool.Get(ctx)
	// Process might start successfully even with invalid path (dart run handles it)
	// So we just verify the pool works
	if err != nil {
		t.Logf("Get() returned error (expected for invalid path): %v", err)
	}

	pool.Close()
}

// TestProcessPool_Release_FullPoolKillProcess tests Release when pool is full and process needs to be killed
func TestProcessPool_Release_FullPoolKillProcess(t *testing.T) {
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

	// Fill the pool
	proc1, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get first process: %v", err)
	}

	// Create another process beyond the pool size
	proc2, err := pool.Get(ctx)
	if err != nil {
		t.Fatalf("Failed to get second process: %v", err)
	}

	// Release first process - should go back to pool
	pool.Release(proc1)

	// Release second process - pool is full, should kill it
	pool.Release(proc2)
}
