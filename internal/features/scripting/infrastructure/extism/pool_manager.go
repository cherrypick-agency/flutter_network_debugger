package extism

import (
	"context"
	"log"
	"sync"

	extism "github.com/extism/go-sdk"

	"network-debugger/internal/features/scripting/domain"
)

// PoolManager coordinates multiple PluginPools (one per script)
type PoolManager struct {
	pools      map[string]*PluginPool
	mu         sync.RWMutex
	createFunc func(context.Context, domain.Script) (*extism.Plugin, error)
}

// NewPoolManager creates a new pool manager
func NewPoolManager(createFunc func(context.Context, domain.Script) (*extism.Plugin, error)) *PoolManager {
	return &PoolManager{
		pools:      make(map[string]*PluginPool),
		createFunc: createFunc,
	}
}

// GetPool gets or creates a pool for the given script
func (pm *PoolManager) GetPool(script domain.Script) *PluginPool {
	// Fast path: read lock
	pm.mu.RLock()
	pool, exists := pm.pools[script.ID]
	pm.mu.RUnlock()

	if exists {
		return pool
	}

	// Slow path: write lock
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// Double-check after acquiring write lock
	if pool, exists := pm.pools[script.ID]; exists {
		return pool
	}

	// Create new pool
	pool = NewPluginPool(script, pm.createFunc)
	pm.pools[script.ID] = pool

	return pool
}

// InvalidateScript removes and closes the pool for a script (used on script update/delete)
func (pm *PoolManager) InvalidateScript(scriptID string) {
	pm.mu.Lock()
	pool, exists := pm.pools[scriptID]
	if exists {
		delete(pm.pools, scriptID)
	}
	pm.mu.Unlock()

	if pool != nil {
		log.Printf("[PoolManager] Invalidating pool for script %s", scriptID)
		_ = pool.Close()
	}
}

// CloseAll closes all pools (used on graceful shutdown)
func (pm *PoolManager) CloseAll() {
	pm.mu.Lock()
	pools := make([]*PluginPool, 0, len(pm.pools))
	for _, pool := range pm.pools {
		pools = append(pools, pool)
	}
	pm.pools = make(map[string]*PluginPool)
	pm.mu.Unlock()

	log.Printf("[PoolManager] Closing %d pool(s)...", len(pools))

	// Close all pools concurrently
	var wg sync.WaitGroup
	for _, pool := range pools {
		wg.Add(1)
		go func(p *PluginPool) {
			defer wg.Done()
			_ = p.Close()
		}(pool)
	}

	wg.Wait()
	log.Printf("[PoolManager] All pools closed")
}

// GetAllMetrics returns metrics for all pools
func (pm *PoolManager) GetAllMetrics() []PoolMetrics {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	metrics := make([]PoolMetrics, 0, len(pm.pools))
	for _, pool := range pm.pools {
		metrics = append(metrics, pool.GetMetrics())
	}

	return metrics
}
