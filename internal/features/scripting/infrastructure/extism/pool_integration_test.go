package extism

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"testing"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

// TestFullIntegration tests complete lifecycle: create → execute → update → delete
func TestFullIntegration(t *testing.T) {
	skipIfNoWASMFixtures(t)

	script1 := createSuccessScript(t)
	script1.Name = "Integration Test Script V1"

	executor := createTestExecutor(t)
	manager := NewPoolManager(executor.createPlugin)
	defer manager.CloseAll()

	testInput := []byte(`{"test": "integration"}`)

	t.Log("Phase 1: Initial script execution (1000 requests)")

	// Phase 1: Execute original script 1000 times
	executePhase := func(script domain.Script, count int, phaseName string) {
		start := time.Now()
		var wg sync.WaitGroup
		errors := make(chan error, count)

		for i := 0; i < count; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()

				pool := manager.GetPool(script)
				_, err := executeScript(t, pool, testInput)
				if err != nil {
					errors <- fmt.Errorf("%s execution %d failed: %w", phaseName, id, err)
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// Check for errors
		errorCount := 0
		for err := range errors {
			t.Error(err)
			errorCount++
			if errorCount >= 10 {
				t.Fatal("too many errors, aborting")
			}
		}

		duration := time.Since(start)
		t.Logf("%s: %d requests completed in %v (%.0f req/s)",
			phaseName, count, duration, float64(count)/duration.Seconds())

		// Check metrics
		pool := manager.GetPool(script)
		metrics := pool.GetMetrics()
		hitRate := float64(metrics.Hits) / float64(metrics.Hits+metrics.Misses) * 100
		t.Logf("%s metrics: hits=%d misses=%d evictions=%d hitRate=%.1f%% idle=%d",
			phaseName, metrics.Hits, metrics.Misses, metrics.Evictions, hitRate, metrics.IdleCount)

		if hitRate < 70.0 {
			t.Errorf("%s: hit rate %.1f%% is too low (expected > 70%%)", phaseName, hitRate)
		}
	}

	executePhase(script1, 1000, "Phase 1")

	t.Log("Phase 2: Script update (simulating code change)")

	// Phase 2: Update script (simulate code change)
	// Use valid WASM code (can reuse same noop.wasm for testing)
	script2 := createSuccessScript(t)
	script2.ID = script1.ID // Keep same ID to simulate update
	script2.Name = "Integration Test Script V2"

	// Invalidate old pool
	manager.InvalidateScript(script1.ID)

	// Execute updated script
	executePhase(script2, 1000, "Phase 2")

	t.Log("Phase 3: Script delete")

	// Get reference to the pool before invalidation
	oldPool := manager.GetPool(script2)

	// Phase 3: Delete script
	manager.InvalidateScript(script2.ID)

	// Verify old pool is closed (should return error)
	ctx := context.Background()
	_, err := oldPool.Acquire(ctx)
	if err == nil {
		t.Error("expected error from old pool after invalidation, but it still works")
	} else {
		t.Logf("Old pool correctly closed after invalidation: %v", err)
	}

	t.Log("Full integration test completed successfully")
}

// TestMemoryStability tests memory stability under load (5000 executions)
func TestMemoryStability(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping memory stability test in short mode")
	}

	skipIfNoWASMFixtures(t)

	script := createSuccessScript(t)
	pool := createTestPool(t, script)

	testInput := []byte(`{"test": "memory"}`)

	// Force GC and measure baseline memory
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	t.Logf("Memory before: Alloc=%d MB, Sys=%d MB",
		memBefore.Alloc/1024/1024, memBefore.Sys/1024/1024)

	// Execute 5000 times
	t.Log("Executing 5000 requests...")
	executionCount := 5000
	batchSize := 100
	batches := executionCount / batchSize

	start := time.Now()

	for batch := 0; batch < batches; batch++ {
		var wg sync.WaitGroup
		for i := 0; i < batchSize; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, _ = executeScript(t, pool, testInput)
			}()
		}
		wg.Wait()

		// Log progress
		if (batch+1)%10 == 0 {
			progress := float64(batch+1) / float64(batches) * 100
			t.Logf("Progress: %.0f%% (%d/%d batches)", progress, batch+1, batches)
		}
	}

	duration := time.Since(start)
	t.Logf("Completed %d executions in %v (%.0f req/s)",
		executionCount, duration, float64(executionCount)/duration.Seconds())

	// Force GC and measure memory after
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	t.Logf("Memory after: Alloc=%d MB, Sys=%d MB",
		memAfter.Alloc/1024/1024, memAfter.Sys/1024/1024)

	// Calculate memory growth
	allocGrowth := int64(memAfter.Alloc) - int64(memBefore.Alloc)
	allocGrowthMB := allocGrowth / 1024 / 1024

	t.Logf("Memory growth: %d MB (Alloc)", allocGrowthMB)

	// Verify memory growth is acceptable (< 25MB for 5000 executions)
	// This accounts for compilation cache, runtime overhead, and GC timing
	maxAcceptableGrowthMB := int64(25)
	if allocGrowthMB > maxAcceptableGrowthMB {
		t.Errorf("Memory growth %d MB exceeds acceptable limit %d MB",
			allocGrowthMB, maxAcceptableGrowthMB)
	}

	// Check pool metrics
	metrics := pool.GetMetrics()
	hitRate := float64(metrics.Hits) / float64(metrics.Hits+metrics.Misses) * 100

	t.Logf("Final metrics: hits=%d misses=%d evictions=%d hitRate=%.1f%% idle=%d active=%d",
		metrics.Hits, metrics.Misses, metrics.Evictions, hitRate,
		metrics.IdleCount, metrics.ActiveCount)

	if hitRate < 95.0 {
		t.Logf("Warning: hit rate %.1f%% lower than expected (95%%+)", hitRate)
	}

	// Verify no instances leaked
	if metrics.ActiveCount > 0 {
		t.Errorf("Found %d active instances after all executions completed (should be 0)",
			metrics.ActiveCount)
	}

	if metrics.IdleCount > MaxIdlePerScript {
		t.Errorf("Found %d idle instances, exceeds limit %d",
			metrics.IdleCount, MaxIdlePerScript)
	}

	t.Log("Memory stability test passed")
}
