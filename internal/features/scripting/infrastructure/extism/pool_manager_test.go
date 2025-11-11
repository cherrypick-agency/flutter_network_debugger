package extism

import (
	"context"
	"testing"

	extism "github.com/extism/go-sdk"

	"network-debugger/internal/features/scripting/domain"
)

// TestNewPoolManager tests PoolManager creation
func TestNewPoolManager(t *testing.T) {
	createFunc := func(ctx context.Context, script domain.Script) (*extism.Plugin, error) {
		return nil, nil
	}

	pm := NewPoolManager(createFunc)
	if pm == nil {
		t.Fatal("Expected non-nil PoolManager")
	}

	if pm.pools == nil {
		t.Error("Expected pools map to be initialized")
	}

	if pm.createFunc == nil {
		t.Error("Expected createFunc to be set")
	}
}

// TestPoolManager_GetPool_CreatesNewPool tests GetPool creates new pool
func TestPoolManager_GetPool_CreatesNewPool(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)
	pm := executor.poolManager

	script := createSuccessScript(t)

	pool := pm.GetPool(script)
	if pool == nil {
		t.Fatal("Expected non-nil pool")
	}

	// Get same pool again should return same instance
	pool2 := pm.GetPool(script)
	if pool != pool2 {
		t.Error("Expected same pool instance for same script")
	}
}

// TestPoolManager_GetPool_DifferentScripts tests GetPool with different scripts
func TestPoolManager_GetPool_DifferentScripts(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)
	pm := executor.poolManager

	script1 := createSuccessScript(t)
	script2 := createSuccessScript(t)
	script2.ID = "different-script-id"

	pool1 := pm.GetPool(script1)
	pool2 := pm.GetPool(script2)

	if pool1 == pool2 {
		t.Error("Expected different pools for different scripts")
	}
}

// TestPoolManager_InvalidateScript tests InvalidateScript
func TestPoolManager_InvalidateScript(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)
	pm := executor.poolManager

	script := createSuccessScript(t)
	pool := pm.GetPool(script)

	// Invalidate script
	pm.InvalidateScript(script.ID)

	// Get pool again should create new pool
	pool2 := pm.GetPool(script)
	if pool == pool2 {
		t.Error("Expected new pool after invalidation")
	}
}

// TestPoolManager_CloseAll tests CloseAll
func TestPoolManager_CloseAll(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)
	pm := executor.poolManager

	scripts := []domain.Script{
		createSuccessScript(t),
		createSuccessScript(t),
	}

	for i, script := range scripts {
		script.ID = "script-" + string(rune(i))
		pm.GetPool(script)
	}

	pm.CloseAll()

	// After CloseAll, pools should be empty
	metrics := pm.GetAllMetrics()
	if len(metrics) != 0 {
		t.Errorf("Expected 0 metrics after CloseAll, got %d", len(metrics))
	}
}

// TestPoolManager_GetAllMetrics tests GetAllMetrics
func TestPoolManager_GetAllMetrics(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)
	pm := executor.poolManager

	script := createSuccessScript(t)
	pm.GetPool(script)

	metrics := pm.GetAllMetrics()
	if len(metrics) != 1 {
		t.Errorf("Expected 1 metric, got %d", len(metrics))
	}
}

// TestPoolManager_GetAllMetrics_Empty tests GetAllMetrics with no pools
func TestPoolManager_GetAllMetrics_Empty(t *testing.T) {
	executor, err := NewExtismExecutor("")
	if err != nil {
		t.Fatalf("NewExtismExecutor() error = %v", err)
	}
	defer executor.Close()

	pm := executor.poolManager

	metrics := pm.GetAllMetrics()
	if len(metrics) != 0 {
		t.Errorf("Expected 0 metrics, got %d", len(metrics))
	}
}

// TestPoolManager_GetPool_MultipleScripts_Concurrent tests GetPool with multiple scripts concurrently
func TestPoolManager_GetPool_MultipleScripts_Concurrent(t *testing.T) {
	skipIfNoWASMFixtures(t)

	executor := createTestExecutor(t)
	pm := executor.poolManager

	script1 := createSuccessScript(t)
	script2 := createSuccessScript(t)
	script2.ID = "script-2"

	// Concurrent access to different scripts
	done := make(chan bool, 20)
	for i := 0; i < 10; i++ {
		go func() {
			pool := pm.GetPool(script1)
			if pool == nil {
				t.Error("Expected non-nil pool")
			}
			done <- true
		}()
		go func() {
			pool := pm.GetPool(script2)
			if pool == nil {
				t.Error("Expected non-nil pool")
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 20; i++ {
		<-done
	}

	// Should have two pools
	metrics := pm.GetAllMetrics()
	if len(metrics) != 2 {
		t.Errorf("Expected 2 pools, got %d", len(metrics))
	}
}
