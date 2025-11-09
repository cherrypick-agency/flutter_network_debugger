package extism

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"network-debugger/internal/features/scripting/domain"
)

// TestLoadStability tests sustained high load (1000 req/s for 10 minutes)
// This is the critical test for production readiness
func TestLoadStability(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping load test in short mode (run with: go test -v -run TestLoadStability -timeout 15m)")
	}

	skipIfNoWASMFixtures(t)

	script := createSuccessScript(t)
	script.Name = "Load Test Script"
	script.Config.TimeoutMs = 1000 // 1 second timeout

	executor := createTestExecutor(t)
	manager := NewPoolManager(executor.createPlugin)
	defer manager.CloseAll()

	pool := manager.GetPool(script)
	testInput := []byte(`{"test": "load"}`)

	// Test parameters
	targetRPS := 1000            // Requests per second
	duration := 10 * time.Minute // Test duration
	totalRequests := targetRPS * int(duration.Seconds())

	t.Logf("Starting load test: %d req/s for %v (total: %d requests)", targetRPS, duration, totalRequests)

	// Measure baseline memory
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)
	t.Logf("Memory before: Alloc=%d MB, Sys=%d MB",
		memBefore.Alloc/1024/1024, memBefore.Sys/1024/1024)

	// Metrics
	var (
		successCount atomic.Int64
		errorCount   atomic.Int64
		timeoutCount atomic.Int64
	)

	// Rate limiter using ticker
	ticker := time.NewTicker(time.Second / time.Duration(targetRPS))
	defer ticker.Stop()

	// Progress reporter
	progressTicker := time.NewTicker(30 * time.Second)
	defer progressTicker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), duration+1*time.Minute)
	defer cancel()

	start := time.Now()

	// Worker pool
	var wg sync.WaitGroup
	requestCh := make(chan struct{}, targetRPS*2) // Buffer for bursts

	// Start workers (more workers than concurrency limit to handle queueing)
	workerCount := 100
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for range requestCh {
				_, execErr := executeScript(t, pool, testInput)
				if execErr != nil {
					if execErr.Error() == "timeout" {
						timeoutCount.Add(1)
					}
					errorCount.Add(1)
				} else {
					successCount.Add(1)
				}
			}
		}(i)
	}

	// Progress reporter goroutine
	go func() {
		for {
			select {
			case <-progressTicker.C:
				elapsed := time.Since(start)
				success := successCount.Load()
				errors := errorCount.Load()
				timeouts := timeoutCount.Load()
				total := success + errors
				currentRPS := float64(total) / elapsed.Seconds()
				errorRate := float64(errors) / float64(total) * 100

				metrics := pool.GetMetrics()
				t.Logf("[Progress] Elapsed: %v | Requests: %d (%.0f req/s) | Success: %d | Errors: %d (%.1f%%) | Timeouts: %d | Pool: idle=%d active=%d hits=%d misses=%d circuit_state=%d",
					elapsed.Round(time.Second), total, currentRPS, success, errors, errorRate, timeouts,
					metrics.IdleCount, metrics.ActiveCount, metrics.Hits, metrics.Misses, metrics.CircuitState)

			case <-ctx.Done():
				return
			}
		}
	}()

	// Generate load
	requestsSent := 0
	for requestsSent < totalRequests {
		select {
		case <-ticker.C:
			select {
			case requestCh <- struct{}{}:
				requestsSent++
			default:
				// Worker pool is full, skip this tick
				errorCount.Add(1)
			}
		case <-ctx.Done():
			t.Logf("Test cancelled after %v", time.Since(start))
			goto done
		}
	}

done:
	close(requestCh)
	t.Log("Waiting for workers to finish...")
	wg.Wait()

	elapsed := time.Since(start)
	success := successCount.Load()
	errors := errorCount.Load()
	timeouts := timeoutCount.Load()
	total := success + errors

	// Measure final memory
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	// Calculate metrics
	actualRPS := float64(total) / elapsed.Seconds()
	errorRate := float64(errors) / float64(total) * 100
	timeoutRate := float64(timeouts) / float64(total) * 100
	allocGrowthMB := int64(memAfter.Alloc-memBefore.Alloc) / 1024 / 1024

	// Final metrics
	poolMetrics := pool.GetMetrics()
	hitRate := float64(poolMetrics.Hits) / float64(poolMetrics.Hits+poolMetrics.Misses) * 100

	t.Logf("\n"+
		"=== LOAD TEST RESULTS ===\n"+
		"Duration: %v\n"+
		"Total Requests: %d\n"+
		"Successful: %d\n"+
		"Errors: %d (%.2f%%)\n"+
		"Timeouts: %d (%.2f%%)\n"+
		"Actual RPS: %.0f (target: %d)\n"+
		"Memory Growth: %d MB\n"+
		"Pool Hit Rate: %.1f%%\n"+
		"Pool Stats: idle=%d active=%d hits=%d misses=%d evictions=%d\n"+
		"Circuit Breaker: state=%d errors=%d/%d\n",
		elapsed, total, success, errors, errorRate, timeouts, timeoutRate,
		actualRPS, targetRPS, allocGrowthMB, hitRate,
		poolMetrics.IdleCount, poolMetrics.ActiveCount, poolMetrics.Hits, poolMetrics.Misses, poolMetrics.Evictions,
		poolMetrics.CircuitState, poolMetrics.ErrorCalls, poolMetrics.TotalCalls)

	// Assertions for production readiness
	if errorRate > 5.0 {
		t.Errorf("Error rate %.2f%% exceeds 5%% threshold", errorRate)
	}

	if actualRPS < float64(targetRPS)*0.8 {
		t.Errorf("Actual RPS %.0f is less than 80%% of target %d", actualRPS, targetRPS)
	}

	if allocGrowthMB > 100 {
		t.Errorf("Memory growth %d MB exceeds 100 MB threshold", allocGrowthMB)
	}

	if hitRate < 90.0 {
		t.Errorf("Pool hit rate %.1f%% is below 90%% threshold", hitRate)
	}

	if poolMetrics.CircuitState == CircuitOpen {
		t.Error("Circuit breaker is open at end of test")
	}

	if poolMetrics.ActiveCount > 0 {
		t.Errorf("Found %d active instances after test completion", poolMetrics.ActiveCount)
	}
}

// TestLoadStabilityMultiScript tests load with multiple concurrent scripts
func TestLoadStabilityMultiScript(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping multi-script load test in short mode")
	}

	skipIfNoWASMFixtures(t)

	// Create 10 different scripts
	scriptCount := 10
	scripts := make([]domain.Script, scriptCount)
	for i := 0; i < scriptCount; i++ {
		script := createSuccessScript(t)
		script.ID = fmt.Sprintf("script-%d", i)
		script.Name = fmt.Sprintf("Multi Script %d", i)
		script.Config.TimeoutMs = 1000
		scripts[i] = script
	}

	executor := createTestExecutor(t)
	manager := NewPoolManager(executor.createPlugin)
	defer manager.CloseAll()

	// Test parameters
	totalRPS := 1000 // Total requests per second across all scripts
	rpsPerScript := totalRPS / scriptCount
	duration := 5 * time.Minute

	t.Logf("Starting multi-script load test: %d scripts, %d total req/s (%d per script), %v duration",
		scriptCount, totalRPS, rpsPerScript, duration)

	// Measure baseline memory
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	ctx, cancel := context.WithTimeout(context.Background(), duration+1*time.Minute)
	defer cancel()

	var (
		totalSuccess atomic.Int64
		totalErrors  atomic.Int64
	)

	start := time.Now()
	var wg sync.WaitGroup

	// Launch a load generator for each script
	for scriptIdx, script := range scripts {
		wg.Add(1)
		go func(idx int, s domain.Script) {
			defer wg.Done()

			pool := manager.GetPool(s)
			testInput := []byte(fmt.Sprintf(`{"script": %d}`, idx))

			ticker := time.NewTicker(time.Second / time.Duration(rpsPerScript))
			defer ticker.Stop()

			requestsSent := 0
			targetRequests := rpsPerScript * int(duration.Seconds())

			for requestsSent < targetRequests {
				select {
				case <-ticker.C:
					go func() {
						_, err := executeScript(t, pool, testInput)
						if err != nil {
							totalErrors.Add(1)
						} else {
							totalSuccess.Add(1)
						}
					}()
					requestsSent++
				case <-ctx.Done():
					return
				}
			}
		}(scriptIdx, script)
	}

	// Progress reporter
	progressTicker := time.NewTicker(30 * time.Second)
	defer progressTicker.Stop()

	progressDone := make(chan struct{})
	go func() {
		defer close(progressDone)
		for {
			select {
			case <-progressTicker.C:
				elapsed := time.Since(start)
				success := totalSuccess.Load()
				errors := totalErrors.Load()
				total := success + errors
				currentRPS := float64(total) / elapsed.Seconds()

				t.Logf("[Multi-Script Progress] Elapsed: %v | Total: %d (%.0f req/s) | Success: %d | Errors: %d",
					elapsed.Round(time.Second), total, currentRPS, success, errors)

			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()
	<-progressDone // Wait for progress reporter to finish

	elapsed := time.Since(start)
	success := totalSuccess.Load()
	errors := totalErrors.Load()
	total := success + errors

	// Measure final memory
	runtime.GC()
	time.Sleep(100 * time.Millisecond)
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	actualRPS := float64(total) / elapsed.Seconds()
	errorRate := float64(errors) / float64(total) * 100
	allocGrowthMB := int64(memAfter.Alloc-memBefore.Alloc) / 1024 / 1024

	t.Logf("\n"+
		"=== MULTI-SCRIPT LOAD TEST RESULTS ===\n"+
		"Scripts: %d\n"+
		"Duration: %v\n"+
		"Total Requests: %d\n"+
		"Successful: %d\n"+
		"Errors: %d (%.2f%%)\n"+
		"Actual RPS: %.0f (target: %d)\n"+
		"Memory Growth: %d MB\n",
		scriptCount, elapsed, total, success, errors, errorRate, actualRPS, totalRPS, allocGrowthMB)

	// Assertions
	if errorRate > 5.0 {
		t.Errorf("Error rate %.2f%% exceeds 5%% threshold", errorRate)
	}

	if actualRPS < float64(totalRPS)*0.8 {
		t.Errorf("Actual RPS %.0f is less than 80%% of target %d", actualRPS, totalRPS)
	}

	if allocGrowthMB > 200 {
		t.Errorf("Memory growth %d MB exceeds 200 MB threshold for %d scripts", allocGrowthMB, scriptCount)
	}
}

// TestThunderingHerd tests cold start with simultaneous requests
func TestThunderingHerd(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping thundering herd test in short mode")
	}

	skipIfNoWASMFixtures(t)

	script := createSuccessScript(t)
	script.Name = "Thundering Herd Script"

	executor := createTestExecutor(t)
	manager := NewPoolManager(executor.createPlugin)
	defer manager.CloseAll()

	pool := manager.GetPool(script)
	testInput := []byte(`{"test": "thundering_herd"}`)

	// Simulate 100 simultaneous cold-start requests
	concurrency := 100
	t.Logf("Starting thundering herd test with %d concurrent requests", concurrency)

	start := time.Now()
	var wg sync.WaitGroup
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			_, err := executeScript(t, pool, testInput)
			if err != nil {
				errors <- fmt.Errorf("request %d failed: %w", id, err)
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	elapsed := time.Since(start)

	errorCount := 0
	for err := range errors {
		t.Log(err)
		errorCount++
	}

	successCount := concurrency - errorCount
	successRate := float64(successCount) / float64(concurrency) * 100

	metrics := pool.GetMetrics()

	t.Logf("\n"+
		"=== THUNDERING HERD RESULTS ===\n"+
		"Concurrent Requests: %d\n"+
		"Duration: %v\n"+
		"Successful: %d (%.1f%%)\n"+
		"Failed: %d\n"+
		"Pool: hits=%d misses=%d\n",
		concurrency, elapsed, successCount, successRate, errorCount,
		metrics.Hits, metrics.Misses)

	// All requests should succeed (no thundering herd crash)
	if errorCount > 0 {
		t.Errorf("Expected all requests to succeed, but %d/%d failed", errorCount, concurrency)
	}

	// Should not take too long (max 30 seconds for 100 concurrent compilations)
	if elapsed > 30*time.Second {
		t.Errorf("Thundering herd took %v, expected < 30s", elapsed)
	}
}
