package extism

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	extism "github.com/extism/go-sdk"

	"network-debugger/internal/features/scripting/domain"
)

// PluginPool manages a pool of WASM plugin instances for a single script with safety mechanisms
// Features: concurrency limiting, instance expiry, circuit breaker, metrics
type PluginPool struct {
	scriptID string
	script   domain.Script // Store full script for recreation

	// Pool state
	idle   []*pooledPlugin
	active map[*pooledPlugin]struct{}
	mu     sync.Mutex

	// Concurrency limit (prevents data races)
	semaphore chan struct{}

	// Circuit breaker (auto-disable on errors)
	errorCount    atomic.Int64
	totalCount    atomic.Int64
	circuitState  atomic.Uint32 // 0=closed, 1=half_open, 2=open
	lastErrorTime atomic.Value  // stores time.Time

	// Metrics
	hits               atomic.Int64 // Found in pool
	misses             atomic.Int64 // Created new
	evictions          atomic.Int64 // Expired
	backgroundCloseOps atomic.Int64 // Pending background close operations

	// Lifecycle
	closed      atomic.Bool
	closedCh    chan struct{} // Closed to signal pool shutdown to waiting goroutines
	stopCleanup chan struct{}
	wg          sync.WaitGroup

	// Back-reference for plugin creation
	createFunc func(context.Context, domain.Script) (*extism.Plugin, error)
}

// pooledPlugin wraps extism.Plugin with metadata for expiry tracking
type pooledPlugin struct {
	plugin    *extism.Plugin
	createdAt time.Time
	useCount  atomic.Int64
	lastUsed  atomic.Value // stores time.Time
	closed    atomic.Bool  // Prevents double-close panics
}

// safeClose closes the plugin if not already closed (prevents double-close panics)
// CRITICAL: Always uses a 5-second timeout to prevent goroutine leaks if Close() blocks
func (p *pooledPlugin) safeClose(ctx context.Context) error {
	if p == nil || p.plugin == nil {
		return nil
	}

	// Atomic check-and-set to prevent double close
	if !p.closed.CompareAndSwap(false, true) {
		// Already closed, skip
		return nil
	}

	// Always use timeout context to prevent indefinite blocking
	closeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	return p.plugin.Close(closeCtx)
}

// Configuration constants
const (
	MaxConcurrentPerScript = 10              // Max parallel executions per script
	InstanceTTL            = 5 * time.Minute // Max age of instance
	MaxUseCount            = 100             // Max executions per instance
	MaxIdlePerScript       = 5               // Max idle instances in pool
	CircuitErrorThreshold  = 0.5             // 50% error rate opens circuit
	CircuitMinSamples      = 10              // Minimum calls before checking circuit
	CircuitRecoveryTime    = 1 * time.Minute // Wait time before half-open attempt
	CleanupInterval        = 1 * time.Minute // Background cleanup frequency

	// Circuit breaker states
	CircuitClosed   uint32 = 0 // Normal operation
	CircuitHalfOpen uint32 = 1 // Testing after recovery
	CircuitOpen     uint32 = 2 // Failing, rejecting requests
)

func clampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func clampDuration(v, min, max time.Duration) time.Duration {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

func clampFloat64(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

var (
	effectiveMaxConcurrentPerScript = clampInt(MaxConcurrentPerScript, 1, 1000)
	effectiveInstanceTTL            = clampDuration(InstanceTTL, 10*time.Second, time.Hour)
	effectiveMaxUseCount            = int64(clampInt(MaxUseCount, 1, 10000))
	effectiveMaxIdlePerScript       = clampInt(MaxIdlePerScript, 0, 100)
	effectiveCircuitErrorThreshold  = clampFloat64(CircuitErrorThreshold, 0.000001, 1)
	effectiveCircuitMinSamples      = int64(clampInt(CircuitMinSamples, 1, 1_000_000))
	effectiveCircuitRecoveryTime    = clampDuration(CircuitRecoveryTime, time.Second, 30*time.Minute)
	effectiveCleanupInterval        = clampDuration(CleanupInterval, 10*time.Second, 30*time.Minute)
)

// init validates configuration constants at startup (без паники: используем безопасные значения)
func init() {
	if effectiveMaxConcurrentPerScript != MaxConcurrentPerScript ||
		effectiveInstanceTTL != InstanceTTL ||
		effectiveMaxUseCount != MaxUseCount ||
		effectiveMaxIdlePerScript != MaxIdlePerScript ||
		effectiveCircuitErrorThreshold != CircuitErrorThreshold ||
		effectiveCircuitMinSamples != CircuitMinSamples ||
		effectiveCircuitRecoveryTime != CircuitRecoveryTime ||
		effectiveCleanupInterval != CleanupInterval {
		log.Printf(
			"[PluginPool] Configuration adjusted to safe values: MaxConcurrent=%d, TTL=%v, MaxUse=%d, MaxIdle=%d, CircuitThreshold=%.3f, CircuitMinSamples=%d, CircuitRecovery=%v, CleanupInterval=%v",
			effectiveMaxConcurrentPerScript,
			effectiveInstanceTTL,
			effectiveMaxUseCount,
			effectiveMaxIdlePerScript,
			effectiveCircuitErrorThreshold,
			effectiveCircuitMinSamples,
			effectiveCircuitRecoveryTime,
			effectiveCleanupInterval,
		)
	} else {
		log.Printf(
			"[PluginPool] Configuration validated: MaxConcurrent=%d, TTL=%v, MaxUse=%d, MaxIdle=%d",
			effectiveMaxConcurrentPerScript,
			effectiveInstanceTTL,
			effectiveMaxUseCount,
			effectiveMaxIdlePerScript,
		)
	}
}

// NewPluginPool creates a new plugin pool for a script
func NewPluginPool(script domain.Script, createFunc func(context.Context, domain.Script) (*extism.Plugin, error)) *PluginPool {
	if createFunc == nil {
		createFunc = func(context.Context, domain.Script) (*extism.Plugin, error) {
			return nil, errors.New("plugin createFunc is nil")
		}
	}

	pool := &PluginPool{
		scriptID:    script.ID,
		script:      script,
		active:      make(map[*pooledPlugin]struct{}),
		semaphore:   make(chan struct{}, effectiveMaxConcurrentPerScript),
		closedCh:    make(chan struct{}),
		stopCleanup: make(chan struct{}),
		createFunc:  createFunc,
	}

	// Initialize lastErrorTime to avoid nil panic
	pool.lastErrorTime.Store(time.Time{})

	// Start cleanup goroutine
	pool.wg.Add(1)
	go pool.cleanupLoop()

	log.Printf("[PluginPool:%s] Created pool for script '%s'", script.ID, script.Name)

	return pool
}

// Acquire gets a plugin instance from the pool (blocks if at concurrency limit)
func (p *PluginPool) Acquire(ctx context.Context) (*pooledPlugin, error) {
	if p.closed.Load() {
		return nil, errors.New("pool is closed")
	}

	// Check circuit breaker
	state := p.circuitState.Load()
	if state == CircuitOpen {
		lastErrorVal := p.lastErrorTime.Load()
		lastError, ok := lastErrorVal.(time.Time)
		if !ok || lastError.IsZero() {
			// Not initialized yet - treat as very old error, allow recovery
			lastError = time.Time{}
		}

		if time.Since(lastError) > effectiveCircuitRecoveryTime {
			// Atomically transition OPEN → HALF_OPEN
			// CompareAndSwap ensures only one goroutine does the transition
			if p.circuitState.CompareAndSwap(CircuitOpen, CircuitHalfOpen) {
				log.Printf("[PluginPool:%s] Circuit breaker entering HALF-OPEN state", p.scriptID)
			}
			// Continue execution in half-open state (allow test request)
		} else {
			return nil, fmt.Errorf("circuit breaker open (wait %v)", effectiveCircuitRecoveryTime-time.Since(lastError))
		}
	}

	// Acquire semaphore (concurrency limit) - CRITICAL: must release on ALL error paths
	select {
	case p.semaphore <- struct{}{}:
		// Got slot
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-p.closedCh:
		return nil, errors.New("pool is closed")
	}

	// Check if pool was closed while waiting for semaphore
	// This prevents wasting 2-5 sec on WASM compilation for a closed pool
	if p.closed.Load() {
		<-p.semaphore // CRITICAL: Release semaphore to avoid leak!
		return nil, errors.New("pool is closed")
	}

	var instance *pooledPlugin
	var err error

	// Defer to ensure semaphore is released on error
	defer func() {
		if err != nil {
			<-p.semaphore
		}
	}()

	// Try to get from idle pool
	p.mu.Lock()
	if len(p.idle) > 0 {
		// Pop from end (LIFO for better cache locality)
		instance = p.idle[len(p.idle)-1]
		p.idle = p.idle[:len(p.idle)-1]

		// Check expiry
		age := time.Since(instance.createdAt)
		uses := instance.useCount.Load()

		if age > effectiveInstanceTTL || uses >= effectiveMaxUseCount {
			// Expired, close and force recreation
			p.mu.Unlock()
			_ = instance.safeClose(context.Background())
			p.evictions.Add(1)
			instance = nil
			p.mu.Lock()
		} else {
			// Valid instance - hit!
			p.hits.Add(1)
			recordPoolCounters(p.scriptID, p.script.Name, true, false, false)
		}
	}
	p.mu.Unlock()

	// Create new instance if needed
	if instance == nil {
		// Check context AND pool closure before expensive WASM compilation (2-5 seconds)
		// CRITICAL: Prevent wasting 2-5s on compilation if pool is closing
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return nil, err
		case <-p.closedCh:
			err = errors.New("pool is closed")
			return nil, err
		default:
		}

		plugin, createErr := p.createFunc(ctx, p.script)
		if createErr != nil {
			// Учитываем ошибку создания в circuit breaker, иначе можно бесконечно жечь CPU на компиляции.
			total := p.totalCount.Add(1)
			errors := p.errorCount.Add(1)
			p.lastErrorTime.Store(time.Now())

			errorRate := float64(errors) / float64(total)
			if total >= effectiveCircuitMinSamples {
				for {
					currentState := p.circuitState.Load()
					if errorRate > effectiveCircuitErrorThreshold {
						if currentState == CircuitClosed || currentState == CircuitHalfOpen {
							if p.circuitState.CompareAndSwap(currentState, CircuitOpen) {
								log.Printf("[PluginPool:%s] Circuit breaker OPEN (%.1f%% errors, %d/%d failures)",
									p.scriptID, errorRate*100, errors, total)
								break
							}
							continue
						}
					}
					break
				}
			}

			// Log detailed error with circuit breaker state
			currentState := p.circuitState.Load()
			stateStr := "CLOSED"
			if currentState == CircuitHalfOpen {
				stateStr = "HALF-OPEN"
			} else if currentState == CircuitOpen {
				stateStr = "OPEN"
			}
			log.Printf("[PluginPool:%s] Failed to create plugin: %v (circuit: %s, errors: %d/%d = %.1f%%)",
				p.scriptID, createErr, stateStr, errors, total, errorRate*100)
			err = fmt.Errorf("failed to create plugin: %w", createErr)
			return nil, err
		}

		// Check context again after compilation - may have been cancelled
		if ctx.Err() != nil {
			// Context cancelled during compilation - close plugin and return
			tempInstance := &pooledPlugin{plugin: plugin}
			_ = tempInstance.safeClose(context.Background())
			err = ctx.Err()
			return nil, err
		}

		now := time.Now()
		instance = &pooledPlugin{
			plugin:    plugin,
			createdAt: now,
		}
		// Initialize lastUsed immediately to prevent nil panic
		instance.lastUsed.Store(now)

		p.misses.Add(1)
		recordPoolCounters(p.scriptID, p.script.Name, false, true, false)
	}

	// Track as active
	p.mu.Lock()
	// Check again if pool was closed while we were creating the plugin
	if p.active == nil {
		p.mu.Unlock()
		// Close the just-created plugin if pool was closed
		// CRITICAL: Use Background() not ctx (might be cancelled)
		if instance != nil {
			_ = instance.safeClose(context.Background())
		}
		// CRITICAL: Set err to trigger defer semaphore release
		err = fmt.Errorf("pool closed during plugin creation")
		return nil, err
	}
	p.active[instance] = struct{}{}
	p.mu.Unlock()

	return instance, nil
}

// Release returns a plugin instance to the pool
func (p *PluginPool) Release(instance *pooledPlugin, execErr error) {
	if instance == nil {
		return
	}

	// CRITICAL: Defer semaphore release at start to prevent leaks on panic
	// This ensures semaphore is always released, even if function panics
	defer func() { <-p.semaphore }()

	// Update usage metadata (atomic operations - no lock needed)
	instance.useCount.Add(1)
	instance.lastUsed.Store(time.Now())

	// === CIRCUIT BREAKER LOGIC (Pure Atomic Strategy - NO MUTEX) ===
	// This matches Acquire() strategy (lines 188-207) for consistency
	total := p.totalCount.Add(1)
	if execErr != nil {
		errors := p.errorCount.Add(1)
		p.lastErrorTime.Store(time.Now())

		// Log error details for diagnostics
		errorRate := float64(errors) / float64(total)
		log.Printf("[PluginPool:%s] Plugin execution error: %v (total: %d, errors: %d, rate: %.1f%%)",
			p.scriptID, execErr, total, errors, errorRate*100)

		// Check if should open circuit
		if total >= effectiveCircuitMinSamples {
			// CAS retry loop to handle concurrent state changes correctly
			// Prevents TOCTOU race where state changes between Load() and CAS
			for {
				currentState := p.circuitState.Load()

				// Check if error rate exceeds threshold
				if errorRate > effectiveCircuitErrorThreshold {
					// Can open from CLOSED or HALF_OPEN (not if already OPEN)
					if currentState == CircuitClosed || currentState == CircuitHalfOpen {
						// Try atomic transition to OPEN
						if p.circuitState.CompareAndSwap(currentState, CircuitOpen) {
							log.Printf("[PluginPool:%s] Circuit breaker OPEN (%.1f%% errors, %d/%d failures)",
								p.scriptID, errorRate*100, errors, total)
							break // Success - exit retry loop
						}
						// CAS failed - state changed by another goroutine, retry
						continue
					}
				}
				// State already OPEN or error rate acceptable - no change needed
				break
			}
		}
	} else {
		// Success - close circuit if in half-open state
		// CAS retry loop for correct state transition under concurrency
		for {
			currentState := p.circuitState.Load()
			if currentState == CircuitHalfOpen {
				// Try atomic transition to CLOSED
				if p.circuitState.CompareAndSwap(CircuitHalfOpen, CircuitClosed) {
					// Accept eventual consistency: reset counters AFTER state transition
					// Multiple goroutines may read stale counter values briefly, but this is
					// acceptable for circuit breaker logic (eventual consistency is sufficient)
					p.errorCount.Store(0)
					p.totalCount.Store(0)
					log.Printf("[PluginPool:%s] Circuit breaker CLOSED (recovered)", p.scriptID)
					break // Success - exit retry loop
				}
				// CAS failed - state changed by another goroutine, retry
				continue
			}
			// Not in HALF_OPEN state - no action needed
			break
		}
	}

	// === POOL OPERATIONS (Mutex Protected) ===
	// Mutex protects ONLY pool data structures (idle, active maps)
	// Circuit breaker uses ONLY atomics (above) - clear separation of concerns
	p.mu.Lock()

	// Check if pool was closed
	if p.active == nil {
		p.mu.Unlock()
		_ = instance.safeClose(context.Background())
		p.evictions.Add(1)
		// Semaphore already released by defer
		log.Printf("[PluginPool:%s] Released plugin after force close", p.scriptID)
		return
	}

	// Remove from active tracking
	delete(p.active, instance)

	// Decide whether to keep or discard instance
	age := time.Since(instance.createdAt)
	uses := instance.useCount.Load()
	poolClosed := p.closed.Load()

	ctxErr := execErr != nil && (errors.Is(execErr, context.DeadlineExceeded) || errors.Is(execErr, context.Canceled))
	shouldDiscard := age > effectiveInstanceTTL || uses >= effectiveMaxUseCount || poolClosed || ctxErr

	if shouldDiscard {
		p.mu.Unlock()
		// Discard - close in background with panic recovery and timeout
		p.backgroundCloseOps.Add(1)
		go func(inst *pooledPlugin) {
			// IMPORTANT: defer order matters! Panic recovery must run LAST (declared FIRST)
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[PluginPool:%s] Panic closing plugin: %v", p.scriptID, r)
				}
			}()
			defer p.backgroundCloseOps.Add(-1)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := inst.safeClose(ctx); err != nil {
				log.Printf("[PluginPool:%s] Failed to close plugin: %v", p.scriptID, err)
			}
		}(instance)
		p.evictions.Add(1)
		// Semaphore already released by defer
		return
	}

	// Check pool size limit
	if len(p.idle) >= effectiveMaxIdlePerScript {
		// Evict oldest (first in slice)
		oldest := p.idle[0]
		p.idle = p.idle[1:]
		p.backgroundCloseOps.Add(1)
		go func(inst *pooledPlugin) {
			// IMPORTANT: defer order matters! Panic recovery must run LAST (declared FIRST)
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[PluginPool:%s] Panic closing evicted plugin: %v", p.scriptID, r)
				}
			}()
			defer p.backgroundCloseOps.Add(-1)
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := inst.safeClose(ctx); err != nil {
				log.Printf("[PluginPool:%s] Failed to close evicted plugin: %v", p.scriptID, err)
			}
		}(oldest)
		p.evictions.Add(1)
	}

	// Return to pool
	p.idle = append(p.idle, instance)

	p.mu.Unlock()
	// Semaphore already released by defer
}

// Close shuts down the pool and cleans up all instances
func (p *PluginPool) Close() error {
	if !p.closed.CompareAndSwap(false, true) {
		return nil // Already closed
	}

	log.Printf("[PluginPool:%s] Closing pool...", p.scriptID)

	// Signal all waiting goroutines to wake up
	close(p.closedCh)

	// Stop cleanup goroutine
	close(p.stopCleanup)

	// Wait for cleanup goroutine to finish
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Cleanup goroutine stopped
	case <-time.After(2 * time.Second):
		log.Printf("[PluginPool:%s] Warning: cleanup goroutine did not stop in time", p.scriptID)
	}

	// Wait for all active executions to finish (CRITICAL: prevents use-after-close crashes)
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	gracefulShutdown := true
	for {
		p.mu.Lock()
		activeCount := len(p.active)
		p.mu.Unlock()

		if activeCount == 0 {
			// All executions finished, safe to close
			break
		}

		select {
		case <-timeout:
			// Force close after timeout (but log warning)
			p.mu.Lock()
			remaining := len(p.active)
			p.mu.Unlock()
			log.Printf("[PluginPool:%s] Warning: forcing close with %d active executions remaining - NOT closing active plugins to prevent crashes", p.scriptID, remaining)
			gracefulShutdown = false
			goto forceClose
		case <-ticker.C:
			// Continue waiting
			log.Printf("[PluginPool:%s] Waiting for %d active executions to finish...", p.scriptID, activeCount)
		}
	}

forceClose:
	// Close instances based on shutdown type
	p.mu.Lock()
	ctx := context.Background()

	// Always close idle instances
	for _, instance := range p.idle {
		_ = instance.safeClose(ctx)
	}
	p.idle = nil

	// Only close active instances if graceful shutdown completed
	// If forced (timeout), leave active instances for Release() to handle
	if gracefulShutdown {
		// CRITICAL: Check for nil to prevent panic if map already cleared
		if p.active != nil {
			for instance := range p.active {
				_ = instance.safeClose(ctx)
			}
		}
	} else {
		// Mark active map as nil to signal pool is closed, but don't close plugins
		// Release() will handle closing when goroutines complete
		activeCount := 0
		if p.active != nil {
			activeCount = len(p.active)
		}
		log.Printf("[PluginPool:%s] Leaving %d active plugins for Release() to clean up", p.scriptID, activeCount)
	}
	p.active = nil
	p.mu.Unlock()

	// Wait for background close operations to complete
	// CRITICAL: Background closes have 5s timeout, so we wait up to 6s
	// This prevents returning from Close() while background goroutines still running
	{
		bgTimeout := time.After(6 * time.Second)
		bgTicker := time.NewTicker(100 * time.Millisecond)
		defer bgTicker.Stop()

		for {
			pending := p.backgroundCloseOps.Load()
			if pending == 0 {
				break
			}

			select {
			case <-bgTimeout:
				log.Printf("[PluginPool:%s] Warning: %d background close operations still pending after 6s timeout", p.scriptID, pending)
				goto skipBackgroundWait
			case <-bgTicker.C:
				// Continue waiting
			}
		}
	}

skipBackgroundWait:

	log.Printf("[PluginPool:%s] Closed (hits=%d, misses=%d, evictions=%d)",
		p.scriptID, p.hits.Load(), p.misses.Load(), p.evictions.Load())

	return nil
}

// cleanupLoop runs periodically to evict expired instances from idle pool
func (p *PluginPool) cleanupLoop() {
	defer p.wg.Done()

	ticker := time.NewTicker(effectiveCleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[PluginPool:%s] Panic during cleanup: %v", p.scriptID, r)
					}
				}()
				p.cleanup()
			}()
		case <-p.stopCleanup:
			return
		}
	}
}

// cleanup evicts expired instances from idle pool
func (p *PluginPool) cleanup() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if len(p.idle) == 0 {
		return
	}

	var kept []*pooledPlugin
	now := time.Now()

	for _, instance := range p.idle {
		age := now.Sub(instance.createdAt)
		uses := instance.useCount.Load()

		if age > effectiveInstanceTTL || uses >= effectiveMaxUseCount {
			// Expired - close it
			p.backgroundCloseOps.Add(1)
			go func(inst *pooledPlugin) {
				// IMPORTANT: defer order matters! Panic recovery must run LAST (declared FIRST)
				defer func() {
					if r := recover(); r != nil {
						log.Printf("[PluginPool:%s] Panic closing expired plugin: %v", p.scriptID, r)
					}
				}()
				defer p.backgroundCloseOps.Add(-1)
				_ = inst.safeClose(context.Background())
			}(instance)
			p.evictions.Add(1)
		} else {
			kept = append(kept, instance)
		}
	}

	if len(kept) < len(p.idle) {
		log.Printf("[PluginPool:%s] Cleanup evicted %d instances (%d remaining)",
			p.scriptID, len(p.idle)-len(kept), len(kept))
		p.idle = kept
	}
}

// GetMetrics returns current pool metrics and updates Prometheus metrics
func (p *PluginPool) GetMetrics() PoolMetrics {
	p.mu.Lock()
	idleCount := len(p.idle)
	activeCount := len(p.active)
	p.mu.Unlock()

	state := p.circuitState.Load()
	metrics := PoolMetrics{
		ScriptID:           p.scriptID,
		IdleCount:          idleCount,
		ActiveCount:        activeCount,
		Hits:               p.hits.Load(),
		Misses:             p.misses.Load(),
		Evictions:          p.evictions.Load(),
		TotalCalls:         p.totalCount.Load(),
		ErrorCalls:         p.errorCount.Load(),
		CircuitState:       state,
		BackgroundCloseOps: p.backgroundCloseOps.Load(),
	}

	// Update Prometheus metrics
	recordPoolMetrics(p.scriptID, p.script.Name, metrics)

	return metrics
}

// PoolMetrics contains pool statistics
type PoolMetrics struct {
	ScriptID           string
	IdleCount          int
	ActiveCount        int
	Hits               int64
	Misses             int64
	Evictions          int64
	TotalCalls         int64
	ErrorCalls         int64
	CircuitState       uint32 // 0=closed, 1=half_open, 2=open
	BackgroundCloseOps int64  // Number of pending background close operations
}
