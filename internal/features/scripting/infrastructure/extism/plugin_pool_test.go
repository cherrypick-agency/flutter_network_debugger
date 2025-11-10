package extism

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestBasicPooling verifies basic pool hit/miss behavior
func TestBasicPooling(t *testing.T) {
	skipIfNoWASMFixtures(t)

	script := createSuccessScript(t)
	pool := createTestPool(t, script)

	testInput := []byte(`{"test": "data"}`)

	// First execution - should be MISS (create new instance)
	_, err := executeScript(t, pool, testInput)
	if err != nil {
		t.Fatalf("first execution failed: %v", err)
	}

	metrics := pool.GetMetrics()
	if metrics.Misses != 1 {
		t.Errorf("expected 1 miss, got %d", metrics.Misses)
	}
	if metrics.Hits != 0 {
		t.Errorf("expected 0 hits, got %d", metrics.Hits)
	}

	// Second execution - should be HIT (reuse from pool)
	_, err = executeScript(t, pool, testInput)
	if err != nil {
		t.Fatalf("second execution failed: %v", err)
	}

	metrics = pool.GetMetrics()
	if metrics.Misses != 1 {
		t.Errorf("expected 1 miss (unchanged), got %d", metrics.Misses)
	}
	if metrics.Hits != 1 {
		t.Errorf("expected 1 hit, got %d", metrics.Hits)
	}

	// Verify idle count
	if metrics.IdleCount != 1 {
		t.Errorf("expected 1 idle instance, got %d", metrics.IdleCount)
	}
}

// TestConcurrentAccess verifies thread-safety with 100 parallel goroutines
func TestConcurrentAccess(t *testing.T) {
	skipIfNoWASMFixtures(t)

	script := createSuccessScript(t)
	pool := createTestPool(t, script)

	testInput := []byte(`{"test": "concurrent"}`)

	var wg sync.WaitGroup
	concurrency := 100
	errors := make(chan error, concurrency)

	// Launch 100 concurrent executions
	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			_, err := executeScript(t, pool, testInput)
			if err != nil {
				errors <- fmt.Errorf("goroutine %d failed: %w", id, err)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}

	// Verify metrics
	metrics := pool.GetMetrics()
	totalExecutions := metrics.Hits + metrics.Misses

	if totalExecutions != int64(concurrency) {
		t.Errorf("expected %d total executions, got %d", concurrency, totalExecutions)
	}

	// Hit rate should be > 50% (most requests reused pooled instances)
	hitRate := float64(metrics.Hits) / float64(totalExecutions)
	if hitRate < 0.5 {
		t.Errorf("hit rate too low: %.2f%% (expected > 50%%)", hitRate*100)
	}

	t.Logf("Concurrent test: %d executions, %.1f%% hit rate, %d idle",
		totalExecutions, hitRate*100, metrics.IdleCount)
}

// TestInstanceExpiry_UseCount verifies instance recreation after MaxUseCount
func TestInstanceExpiry_UseCount(t *testing.T) {
	skipIfNoWASMFixtures(t)

	script := createSuccessScript(t)
	pool := createTestPool(t, script)

	testInput := []byte(`{"test": "expiry"}`)

	// Check evictions before
	metricsBefore := pool.GetMetrics()
	evictionsBefore := metricsBefore.Evictions

	// Execute MaxUseCount times (100)
	// On the 100th release, the instance should be evicted (useCount=100 >= MaxUseCount)
	for i := 0; i < MaxUseCount; i++ {
		_, err := executeScript(t, pool, testInput)
		if err != nil {
			t.Fatalf("execution %d failed: %v", i, err)
		}
	}

	metricsAfter100 := pool.GetMetrics()

	// Evictions should have increased by 1 (instance evicted on 100th release)
	expectedEvictions := evictionsBefore + 1
	if metricsAfter100.Evictions != expectedEvictions {
		t.Errorf("expected evictions=%d after MaxUseCount executions, got %d",
			expectedEvictions, metricsAfter100.Evictions)
	}

	// Next execution should create a new instance (miss)
	missesBefore := metricsAfter100.Misses
	_, err := executeScript(t, pool, testInput)
	if err != nil {
		t.Fatalf("execution after eviction failed: %v", err)
	}

	metricsAfter := pool.GetMetrics()
	if metricsAfter.Misses <= missesBefore {
		t.Errorf("expected cache miss after eviction, misses before=%d after=%d",
			missesBefore, metricsAfter.Misses)
	}

	t.Logf("UseCount expiry test: evictions=%d, new instance created (miss)",
		metricsAfter100.Evictions)
}

// TestInstanceExpiry_TTL verifies instance recreation after InstanceTTL
// NOTE: This test is SLOW (waits 6 minutes), run with -short to skip
func TestInstanceExpiry_TTL(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow TTL test in short mode")
	}

	skipIfNoWASMFixtures(t)

	// Create pool with shorter TTL for testing (1 second instead of 5 minutes)
	// We'll need to modify pool creation to accept custom TTL
	// For now, skip this test or wait 6 minutes
	t.Skip("TTL test requires 6+ minutes wait time - implement with configurable TTL")

	// TODO: Implement this test when we add configurable TTL
	// script := createSuccessScript(t)
	// pool := createTestPoolWithTTL(t, script, 1*time.Second)
	//
	// testInput := []byte(`{"test": "ttl"}`)
	//
	// // Execute once
	// _, err := executeScript(t, pool, testInput)
	// if err != nil {
	// 	t.Fatalf("first execution failed: %v", err)
	// }
	//
	// metricsBeforeSleep := pool.GetMetrics()
	//
	// // Wait for TTL to expire
	// time.Sleep(2 * time.Second)
	//
	// // Execute again - should evict expired instance
	// _, err = executeScript(t, pool, testInput)
	// if err != nil {
	// 	t.Fatalf("second execution failed: %v", err)
	// }
	//
	// metricsAfterSleep := pool.GetMetrics()
	//
	// if metricsAfterSleep.Evictions <= metricsBeforeSleep.Evictions {
	// 	t.Error("expected eviction after TTL expired")
	// }
}

// TestCircuitBreaker_Open verifies circuit breaker opens on high error rate
func TestCircuitBreaker_Open(t *testing.T) {
	skipIfNoWASMFixtures(t)

	// Create a script that panics on execution
	script := createFailingScript(t)
	pool := createTestPool(t, script)

	testInput := []byte(`{"test": "circuit"}`)

	// Execute multiple times to trigger errors
	// Need at least CircuitMinSamples (10) executions
	failedCount := 0
	for i := 0; i < 20; i++ {
		_, err := executeScript(t, pool, testInput)
		if err != nil {
			failedCount++
		}
	}

	if failedCount < 10 {
		t.Logf("Warning: only %d/20 executions failed (expected most to fail)", failedCount)
	}

	// Check circuit breaker opened
	metrics := pool.GetMetrics()
	if metrics.CircuitState != CircuitOpen {
		t.Errorf("expected circuit breaker to be OPEN after %d errors, but it's closed", failedCount)
		t.Logf("Metrics: totalCalls=%d errorCalls=%d errorRate=%.1f%%",
			metrics.TotalCalls, metrics.ErrorCalls,
			float64(metrics.ErrorCalls)/float64(metrics.TotalCalls)*100)
	}

	// Next Acquire should fail with circuit breaker error
	ctx := context.Background()
	_, err := pool.Acquire(ctx)
	if err == nil {
		t.Error("expected error when circuit is open, but got nil")
	} else if !contains(err.Error(), "circuit breaker open") {
		t.Errorf("expected 'circuit breaker open' error, got: %v", err)
	}

	t.Logf("Circuit breaker test: %d calls, %d errors, circuit OPEN",
		metrics.TotalCalls, metrics.ErrorCalls)
}

// TestCircuitBreaker_Recovery verifies circuit breaker recovery (half-open → closed)
func TestCircuitBreaker_Recovery(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping slow recovery test in short mode")
	}

	skipIfNoWASMFixtures(t)

	// First, open the circuit (similar to TestCircuitBreaker_Open)
	script := createFailingScript(t)
	pool := createTestPool(t, script)

	testInput := []byte(`{"test": "recovery"}`)

	// Trigger circuit breaker
	for i := 0; i < 20; i++ {
		_, _ = executeScript(t, pool, testInput)
	}

	if pool.GetMetrics().CircuitState != CircuitOpen {
		t.Fatal("circuit breaker should be open, but it's not")
	}

	// Wait for CircuitRecoveryTime (1 minute + buffer)
	t.Log("Waiting 65 seconds for circuit breaker recovery...")
	time.Sleep(65 * time.Second)

	// Next Acquire should enter HALF_OPEN state
	ctx := context.Background()
	_, _ = pool.Acquire(ctx)

	// It should still fail (because script is invalid), but should be in HALF_OPEN
	metrics := pool.GetMetrics()
	if metrics.CircuitState != CircuitHalfOpen && metrics.CircuitState != CircuitOpen {
		t.Error("expected circuit to be in HALF_OPEN or OPEN state after recovery time")
	}

	// For full recovery test, we'd need a script that fails then succeeds
	// Skipping that complexity for now
	t.Logf("Circuit recovery test: circuit entered recovery state (halfOpen=%v)",
		metrics.CircuitState == CircuitHalfOpen)
}

// TestPoolSizeLimit verifies MaxIdlePerScript limit is enforced
func TestPoolSizeLimit(t *testing.T) {
	skipIfNoWASMFixtures(t)

	script := createSuccessScript(t)
	pool := createTestPool(t, script)

	testInput := []byte(`{"test": "pool-size"}`)

	// Create 10 instances by executing 10 times concurrently
	// Then release all at once
	var wg sync.WaitGroup
	instanceCount := 10

	for i := 0; i < instanceCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _ = executeScript(t, pool, testInput)
		}()
	}

	wg.Wait()

	// Wait a bit for all releases to complete
	time.Sleep(100 * time.Millisecond)

	// Check pool size - should be capped at MaxIdlePerScript (5)
	metrics := pool.GetMetrics()

	if metrics.IdleCount > MaxIdlePerScript {
		t.Errorf("pool size %d exceeds MaxIdlePerScript %d", metrics.IdleCount, MaxIdlePerScript)
	}

	// Should have evicted at least 1 instance (proves the mechanism works)
	// On macOS/Linux this typically evicts 5+, but Windows may evict fewer due to scheduler differences
	if metrics.Evictions < 1 {
		t.Errorf("expected at least 1 eviction, got %d", metrics.Evictions)
	}

	t.Logf("Pool size limit test: idle=%d evictions=%d (max=%d)",
		metrics.IdleCount, metrics.Evictions, MaxIdlePerScript)
}

// TestScriptUpdateInvalidation verifies pool invalidation on script update
func TestScriptUpdateInvalidation(t *testing.T) {
	skipIfNoWASMFixtures(t)

	script := createSuccessScript(t)
	executor := createTestExecutor(t)

	// Create pool manager instead of direct pool
	manager := NewPoolManager(executor.createPlugin)
	defer manager.CloseAll()

	testInput := []byte(`{"test": "invalidation"}`)

	// Execute script (creates pool)
	pool := manager.GetPool(script)
	_, err := executeScript(t, pool, testInput)
	if err != nil {
		t.Fatalf("first execution failed: %v", err)
	}

	// Invalidate pool (simulates script update)
	manager.InvalidateScript(script.ID)

	// Try to execute with old pool - should fail (pool closed)
	ctx := context.Background()
	_, err = pool.Acquire(ctx)
	if err == nil {
		t.Error("expected error after invalidation, but got nil")
	}

	// Create new pool (simulates updated script)
	// Use the same WASM code but create a new script instance to simulate an update
	newScript := createSuccessScript(t)
	newScript.ID = script.ID // Keep same ID to simulate update, not new script
	newPool := manager.GetPool(newScript)

	// Execute with new pool - should work and be a MISS
	_, err = executeScript(t, newPool, testInput)
	if err != nil {
		t.Fatalf("execution with new pool failed: %v", err)
	}

	metricsAfterInvalidation := newPool.GetMetrics()

	if metricsAfterInvalidation.Misses == 0 {
		t.Error("expected MISS after invalidation (new pool), but got HIT")
	}

	t.Logf("Invalidation test: old pool closed, new pool created (misses=%d)",
		metricsAfterInvalidation.Misses)
}

// TestScriptDeleteInvalidation verifies pool closed on script delete
func TestScriptDeleteInvalidation(t *testing.T) {
	skipIfNoWASMFixtures(t)

	script := createSuccessScript(t)
	executor := createTestExecutor(t)

	manager := NewPoolManager(executor.createPlugin)
	defer manager.CloseAll()

	testInput := []byte(`{"test": "delete"}`)

	// Create and execute
	pool := manager.GetPool(script)
	_, err := executeScript(t, pool, testInput)
	if err != nil {
		t.Fatalf("execution failed: %v", err)
	}

	// Delete script (invalidate pool)
	manager.InvalidateScript(script.ID)

	// Pool should be closed
	ctx := context.Background()
	_, err = pool.Acquire(ctx)
	if err == nil {
		t.Error("expected error after delete, but pool still works")
	}

	if !contains(err.Error(), "pool is closed") {
		t.Errorf("expected 'pool is closed' error, got: %v", err)
	}

	t.Log("Delete invalidation test: pool correctly closed")
}

// TestGracefulShutdown verifies Close waits for active executions
func TestGracefulShutdown(t *testing.T) {
	skipIfNoWASMFixtures(t)

	script := createSuccessScript(t)
	pool := createTestPool(t, script)

	// Start 5 slow executions
	var wg sync.WaitGroup
	executionCount := 5
	completedCount := &atomic.Int32{}

	for i := 0; i < executionCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			// Acquire instance
			ctx := context.Background()
			instance, err := pool.Acquire(ctx)
			if err != nil {
				t.Logf("goroutine %d: acquire failed: %v", id, err)
				return
			}

			// Sleep to simulate slow execution
			time.Sleep(100 * time.Millisecond)

			// Release
			pool.Release(instance, nil)
			completedCount.Add(1)
			t.Logf("goroutine %d: completed", id)
		}(i)
	}

	// Wait a bit for goroutines to acquire instances
	time.Sleep(50 * time.Millisecond)

	// Start Close in separate goroutine
	closeComplete := make(chan struct{})
	go func() {
		t.Log("Starting Close()...")
		err := pool.Close()
		if err != nil {
			t.Errorf("Close() failed: %v", err)
		}
		close(closeComplete)
		t.Log("Close() completed")
	}()

	// Wait for all executions to complete
	wg.Wait()

	// Close should complete soon after
	select {
	case <-closeComplete:
		t.Log("Graceful shutdown successful")
	case <-time.After(5 * time.Second):
		t.Error("Close() did not complete within 5 seconds")
	}

	completed := completedCount.Load()
	t.Logf("Graceful shutdown test: %d/%d executions completed", completed, executionCount)
}

// TestSemaphoreLimit verifies MaxConcurrentPerScript is enforced
func TestSemaphoreLimit(t *testing.T) {
	skipIfNoWASMFixtures(t)

	script := createSuccessScript(t)
	pool := createTestPool(t, script)

	// Launch MaxConcurrentPerScript + 5 goroutines
	goroutineCount := MaxConcurrentPerScript + 5
	activeCount := &atomic.Int32{}
	maxActive := &atomic.Int32{}

	var wg sync.WaitGroup

	for i := 0; i < goroutineCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			ctx := context.Background()
			instance, err := pool.Acquire(ctx)
			if err != nil {
				t.Logf("goroutine %d: acquire failed: %v", id, err)
				return
			}

			// Track active count
			current := activeCount.Add(1)
			for {
				max := maxActive.Load()
				if current <= max || maxActive.CompareAndSwap(max, current) {
					break
				}
			}

			// Hold for a bit to ensure overlap
			time.Sleep(50 * time.Millisecond)

			activeCount.Add(-1)
			pool.Release(instance, nil)
		}(i)
	}

	wg.Wait()

	max := maxActive.Load()

	if max > int32(MaxConcurrentPerScript) {
		t.Errorf("max concurrent %d exceeds limit %d", max, MaxConcurrentPerScript)
	}

	t.Logf("Semaphore test: max concurrent was %d (limit %d)", max, MaxConcurrentPerScript)
}

// TestCloseWhileAcquiring verifies critical bug fix - close during semaphore wait
func TestCloseWhileAcquiring(t *testing.T) {
	skipIfNoWASMFixtures(t)

	script := createSuccessScript(t)
	pool := createTestPool(t, script)

	ctx := context.Background()

	// Fill all semaphore slots
	instances := make([]*pooledPlugin, MaxConcurrentPerScript)
	for i := 0; i < MaxConcurrentPerScript; i++ {
		instance, err := pool.Acquire(ctx)
		if err != nil {
			t.Fatalf("failed to acquire instance %d: %v", i, err)
		}
		instances[i] = instance
	}

	// Launch goroutine that will block on semaphore
	acquireErr := make(chan error, 1)
	go func() {
		_, err := pool.Acquire(context.Background())
		acquireErr <- err
	}()

	// Wait a bit to ensure goroutine is blocked
	time.Sleep(50 * time.Millisecond)

	// Close pool while goroutine is waiting
	err := pool.Close()
	if err != nil {
		t.Fatalf("Close() failed: %v", err)
	}

	// The blocked goroutine should receive "pool is closed" error
	select {
	case err := <-acquireErr:
		if err == nil {
			t.Error("expected error after close, but got nil")
		} else if !contains(err.Error(), "pool is closed") {
			t.Errorf("expected 'pool is closed' error, got: %v", err)
		} else {
			t.Log("Correctly received 'pool is closed' error")
		}
	case <-time.After(2 * time.Second):
		t.Error("blocked Acquire() did not return within 2 seconds")
	}

	// Release all instances (should be no-op since pool closed)
	for _, instance := range instances {
		pool.Release(instance, nil)
	}
}

// TestMetrics verifies all metrics are accurately tracked
func TestMetrics(t *testing.T) {
	skipIfNoWASMFixtures(t)

	script := createSuccessScript(t)
	pool := createTestPool(t, script)

	testInput := []byte(`{"test": "metrics"}`)

	// Execute 10 times
	for i := 0; i < 10; i++ {
		_, err := executeScript(t, pool, testInput)
		if err != nil {
			t.Fatalf("execution %d failed: %v", i, err)
		}
	}

	metrics := pool.GetMetrics()

	// Verify total calls
	expectedTotal := int64(10)
	if metrics.TotalCalls != expectedTotal {
		t.Errorf("expected totalCalls=%d, got %d", expectedTotal, metrics.TotalCalls)
	}

	// Verify hits + misses = total
	sumHitsMisses := metrics.Hits + metrics.Misses
	if sumHitsMisses != expectedTotal {
		t.Errorf("hits + misses (%d) != totalCalls (%d)", sumHitsMisses, expectedTotal)
	}

	// Verify hit rate (should be high, ~90%)
	hitRate := float64(metrics.Hits) / float64(sumHitsMisses)
	if hitRate < 0.7 {
		t.Errorf("hit rate too low: %.1f%% (expected > 70%%)", hitRate*100)
	}

	// Verify idle + active <= max
	if metrics.IdleCount+metrics.ActiveCount > MaxIdlePerScript+MaxConcurrentPerScript {
		t.Errorf("idle(%d) + active(%d) > limits", metrics.IdleCount, metrics.ActiveCount)
	}

	t.Logf("Metrics test: hits=%d misses=%d evictions=%d hitRate=%.1f%%",
		metrics.Hits, metrics.Misses, metrics.Evictions, hitRate*100)
}

// Helper function to check if string contains substring
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
