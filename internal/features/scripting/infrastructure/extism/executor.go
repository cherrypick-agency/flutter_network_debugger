package extism

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	extism "github.com/extism/go-sdk"
	"github.com/tetratelabs/wazero"

	"network-debugger/internal/features/scripting/domain"
)

// ExtismExecutor implements domain.ScriptExecutor for WASM-based scripts
// Supports: Rust, Go (TinyGo), JavaScript (AssemblyScript), C, C++, and more
type ExtismExecutor struct {
	poolManager      *PoolManager
	compilationCache wazero.CompilationCache

	// Goroutine leak tracking (cannot cancel plugin.Call() - Extism limitation)
	runawayGoroutines atomic.Int64 // Count of goroutines that continue after timeout
	totalTimeouts     atomic.Int64 // Total number of timeouts for rate calculation
}

// NewExtismExecutor creates a new Extism executor with host functions and compilation cache
// cacheDir: directory for persistent compilation cache (e.g., "./cache/wasm")
// If cacheDir is empty, cache is disabled (in-memory only)
func NewExtismExecutor(cacheDir string) (*ExtismExecutor, error) {
	var cache wazero.CompilationCache
	if cacheDir != "" {
		// Create persistent compilation cache on disk
		var err error
		cache, err = wazero.NewCompilationCacheWithDir(cacheDir)
		if err != nil {
			return nil, fmt.Errorf("failed to create compilation cache: %w", err)
		}
		log.Printf("[Extism] Compilation cache enabled at: %s", cacheDir)
	} else {
		log.Printf("[Extism] Compilation cache disabled (in-memory only)")
	}

	executor := &ExtismExecutor{
		compilationCache: cache,
	}

	// Create pool manager with plugin creation function
	executor.poolManager = NewPoolManager(executor.createPlugin)

	return executor, nil
}

// Execute runs a WASM script with the given input using plugin pooling
func (e *ExtismExecutor) Execute(ctx context.Context, req domain.ExecutionRequest) (domain.ExecutionResult, error) {
	start := time.Now()

	// Get pool for script
	pool := e.poolManager.GetPool(req.Script)

	// Acquire plugin instance from pool (blocks if at concurrency limit)
	instance, err := pool.Acquire(ctx)
	if err != nil {
		return domain.ExecutionResult{}, fmt.Errorf("failed to acquire plugin: %w", err)
	}

	// Track execution error for pool metrics.
	// Important: on timeout/cancel we do NOT return the instance to the pool because it may be
	// in an undefined state. This behavior is implemented inside PluginPool.Release().
	var execErr error
	defer func() {
		pool.Release(instance, execErr)
	}()

	// Apply timeout from script config
	execCtx, cancel := context.WithTimeout(ctx, time.Duration(req.Script.Config.TimeoutMs)*time.Millisecond)
	defer cancel()

	// Execute plugin main function
	var output []byte
	done := make(chan error, 1)

	// WARNING: This goroutine continues executing even after timeout!
	// plugin.Call() cannot be cancelled (Extism limitation).
	// This is acceptable because:
	// 1. done channel is buffered (doesn't block goroutine)
	// 2. Plugin instance remains valid until Release() completes
	// 3. Pool's Close() waits for all active executions
	// 4. We track runaway goroutines for monitoring
	// CRITICAL: Use atomic.Bool for thread-safe access from multiple goroutines
	var timedOut atomic.Bool
	go func() {
		var callErr error
		var exitCode uint32
		// Extism v1.7.1+ API returns 3 values: exitCode, output, error
		exitCode, output, callErr = instance.plugin.Call("process", req.Input)
		// Exit code from WASM module (0 = success, non-zero = error)
		// Currently not used - we rely on the error return value instead
		_ = exitCode

		// If goroutine completes after timeout, decrement runaway counter
		if timedOut.Load() {
			remaining := e.runawayGoroutines.Add(-1)
			recordRunawayGoroutine(remaining)
			log.Printf("[Extism] Runaway goroutine completed (remaining: %d)", remaining)
		}

		done <- callErr
	}()

	// Wait for execution or timeout
	select {
	case <-execCtx.Done():
		duration := time.Since(start)
		execErr = execCtx.Err() // Track error for circuit breaker
		if execCtx.Err() == context.DeadlineExceeded {
			// Track runaway goroutine (plugin.Call() cannot be cancelled)
			timedOut.Store(true)
			runaway := e.runawayGoroutines.Add(1)
			e.totalTimeouts.Add(1)
			log.Printf("[Extism] Script timeout - goroutine leak (total runaway: %d, timeouts: %d)",
				runaway, e.totalTimeouts.Load())

			// Update Prometheus metrics
			recordRunawayGoroutine(runaway)
			recordTimeout()
			recordExecution(req.Script.ID, req.Script.Name, duration.Seconds(), "timeout")

			return domain.ExecutionResult{
				Duration: duration,
				Error:    fmt.Sprintf("script timeout after %dms", req.Script.Config.TimeoutMs),
			}, nil
		}
		return domain.ExecutionResult{
			Duration: duration,
			Error:    "script canceled",
		}, nil

	case err := <-done:
		duration := time.Since(start)
		if err != nil {
			execErr = err // Track error for circuit breaker
			recordExecution(req.Script.ID, req.Script.Name, duration.Seconds(), "error")
			return domain.ExecutionResult{
				Duration: duration,
				Error:    err.Error(),
			}, nil
		}

		recordExecution(req.Script.ID, req.Script.Name, duration.Seconds(), "success")
		return domain.ExecutionResult{
			Output:   output,
			Duration: duration,
		}, nil
	}
}

// Runtime returns the runtime type
func (e *ExtismExecutor) Runtime() domain.ScriptRuntime {
	return domain.RuntimeExtism
}

// Validate checks if the script is valid WASM with required exports
func (e *ExtismExecutor) Validate(ctx context.Context, script domain.Script) error {
	// Skip validation if Code is empty (script will be compiled later)
	if len(script.Code) == 0 {
		return nil
	}

	// First, validate WASM exports (fast check without creating plugin)
	if err := ValidateWASMWithLanguage(ctx, script.Code, script.Language); err != nil {
		return err
	}

	// Then try to create a plugin to validate the WASM module can be instantiated
	plugin, err := e.createPlugin(ctx, script)
	if err != nil {
		return err
	}

	// Close the plugin since we only created it for validation
	{
		closeCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if closeErr := plugin.Close(closeCtx); closeErr != nil {
			return fmt.Errorf("failed to close plugin after validation: %w", closeErr)
		}
	}

	return nil
}

// Close cleans up all plugin pools and the compilation cache
func (e *ExtismExecutor) Close() error {
	// Close all plugin pools
	e.poolManager.CloseAll()

	// Close compilation cache to flush to disk
	ctx := context.Background()
	if e.compilationCache != nil {
		if err := e.compilationCache.Close(ctx); err != nil {
			log.Printf("[Extism] Warning: failed to close compilation cache: %v", err)
		}
	}

	return nil
}

// createPlugin creates a new plugin for the given script (used by PluginPool)
func (e *ExtismExecutor) createPlugin(ctx context.Context, script domain.Script) (*extism.Plugin, error) {
	// Validate script has compiled code
	if len(script.Code) == 0 {
		log.Printf("[Extism] Cannot create plugin for script '%s' (ID: %s): empty code", script.Name, script.ID)
		return nil, fmt.Errorf("script '%s' has empty code, compilation required", script.Name)
	}

	log.Printf("[Extism] Creating plugin for script '%s' (ID: %s, Code size: %d bytes, Language: %s)",
		script.Name, script.ID, len(script.Code), script.Language)

	// Create host functions with script-specific AllowedHosts
	hostFns := createHostFunctions(script.Config.AllowedHosts)

	// Create new plugin
	manifest := extism.Manifest{
		Wasm: []extism.Wasm{
			extism.WasmData{Data: script.Code},
		},
		Memory: &extism.ManifestMemory{
			MaxPages: uint32(uint64(script.Config.MemoryLimitMB) * 16), // 64KB per page
		},
		AllowedHosts: script.Config.AllowedHosts,
	}

	// Configure plugin with compilation cache
	config := extism.PluginConfig{
		EnableWasi: false, // Security: disable WASI by default
	}

	// Apply compilation cache if available
	if e.compilationCache != nil {
		config.RuntimeConfig = wazero.NewRuntimeConfig().WithCompilationCache(e.compilationCache)
		log.Printf("[Extism] Using compilation cache for script %s", script.Name)
	}

	plugin, err := extism.NewPlugin(ctx, manifest, config, hostFns)
	if err != nil {
		log.Printf("[Extism] Failed to create plugin for script '%s' (ID: %s): %v", script.Name, script.ID, err)
		return nil, fmt.Errorf("failed to create plugin: %w", err)
	}

	log.Printf("[Extism] Successfully created plugin for script '%s' (ID: %s, Language: %s)", script.Name, script.ID, script.Language)

	return plugin, nil
}

// RemovePlugin invalidates the plugin pool for a script (useful when script is updated)
func (e *ExtismExecutor) RemovePlugin(scriptID string) {
	e.poolManager.InvalidateScript(scriptID)
}

// WarmUpPlugins pre-compiles and pools all enabled scripts at startup
// This eliminates cold start penalty for first HTTP requests
func (e *ExtismExecutor) WarmUpPlugins(ctx context.Context, scripts []domain.Script) error {
	if len(scripts) == 0 {
		return nil
	}

	log.Printf("[Extism] Warming up %d plugin(s)...", len(scripts))
	start := time.Now()

	var successCount, failCount int

	for _, script := range scripts {
		// Check context before expensive operations
		select {
		case <-ctx.Done():
			log.Printf("[Extism] Warm-up cancelled: %v", ctx.Err())
			return ctx.Err()
		default:
		}

		// Only warm up enabled scripts
		if !script.Enabled {
			continue
		}

		// Skip non-Extism scripts
		if script.Runtime != domain.RuntimeExtism {
			continue
		}

		// Get pool for script (creates pool if needed)
		pool := e.poolManager.GetPool(script)

		// Pre-compile and pool one instance
		instance, err := pool.Acquire(ctx)
		if err != nil {
			log.Printf("[Extism] Failed to warm up plugin '%s': %v", script.Name, err)
			failCount++
			continue
		}

		// Immediately release back to pool (no error since we didn't execute)
		pool.Release(instance, nil)
		successCount++
	}

	duration := time.Since(start)
	log.Printf("[Extism] Warm-up complete: %d success, %d failed (took %v)",
		successCount, failCount, duration)

	return nil
}

// GetMetrics returns executor-level metrics (goroutine leaks, timeouts, etc.)
func (e *ExtismExecutor) GetMetrics() ExecutorMetrics {
	runaway := e.runawayGoroutines.Load()
	recordRunawayGoroutine(runaway)

	return ExecutorMetrics{
		RunawayGoroutines: runaway,
		TotalTimeouts:     e.totalTimeouts.Load(),
	}
}

// ExecutorMetrics contains executor-level statistics
type ExecutorMetrics struct {
	RunawayGoroutines int64 // Goroutines that continue after timeout
	TotalTimeouts     int64 // Total number of script timeouts
}
