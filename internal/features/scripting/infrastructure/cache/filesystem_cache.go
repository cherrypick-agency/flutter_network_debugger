package cache

import (
	"fmt"
	"os"
	"path/filepath"

	"network-debugger/internal/features/scripting/domain"
)

// FileSystemCache implements CacheManager interface using the local filesystem
// Stores compilers in ~/.cache/network-debugger/compilers/
type FileSystemCache struct {
	baseDir string
}

// NewFileSystemCache creates a new filesystem-based cache manager
// baseDir is the root directory for cache (e.g., ~/.cache/network-debugger)
func NewFileSystemCache(baseDir string) (*FileSystemCache, error) {
	// Expand ~ to home directory if needed
	if baseDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		baseDir = filepath.Join(homeDir, ".cache", "network-debugger")
	}

	cache := &FileSystemCache{
		baseDir: baseDir,
	}

	// Ensure cache directory exists
	if err := cache.EnsureCacheDir(); err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	return cache, nil
}

// GetCacheDir returns the base cache directory
func (f *FileSystemCache) GetCacheDir() string {
	return filepath.Join(f.baseDir, "compilers")
}

// GetCompilerPath returns the path to a specific compiler in cache
// Returns error if compiler not in cache
func (f *FileSystemCache) GetCompilerPath(language string) (string, error) {
	compilerPath := filepath.Join(f.GetCacheDir(), language)

	// Check if compiler directory exists
	info, err := os.Stat(compilerPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("compiler %s not found in cache", language)
		}
		return "", fmt.Errorf("failed to check compiler path: %w", err)
	}

	if !info.IsDir() {
		return "", fmt.Errorf("compiler path is not a directory: %s", compilerPath)
	}

	return compilerPath, nil
}

// EnsureCacheDir ensures the cache directory structure exists
func (f *FileSystemCache) EnsureCacheDir() error {
	cacheDir := f.GetCacheDir()

	// Create directory with proper permissions
	err := os.MkdirAll(cacheDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create cache directory: %w", err)
	}

	return nil
}

// Clear removes a specific compiler from cache
func (f *FileSystemCache) Clear(language string) error {
	compilerPath := filepath.Join(f.GetCacheDir(), language)

	// Check if compiler exists
	_, err := os.Stat(compilerPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Already removed, no error
			return nil
		}
		return fmt.Errorf("failed to check compiler path: %w", err)
	}

	// Remove compiler directory and all contents
	err = os.RemoveAll(compilerPath)
	if err != nil {
		return fmt.Errorf("failed to remove compiler %s: %w", language, err)
	}

	return nil
}

// ClearAll removes all compilers from cache
func (f *FileSystemCache) ClearAll() error {
	cacheDir := f.GetCacheDir()

	// Get all compiler directories
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Cache directory doesn't exist, nothing to clear
			return nil
		}
		return fmt.Errorf("failed to read cache directory: %w", err)
	}

	// Remove each compiler directory
	var lastErr error
	for _, entry := range entries {
		if entry.IsDir() {
			compilerPath := filepath.Join(cacheDir, entry.Name())
			if err := os.RemoveAll(compilerPath); err != nil {
				lastErr = fmt.Errorf("failed to remove %s: %w", entry.Name(), err)
			}
		}
	}

	return lastErr
}

// GetCacheSize returns total size of cache in bytes
func (f *FileSystemCache) GetCacheSize() (int64, error) {
	cacheDir := f.GetCacheDir()

	// Check if cache directory exists
	_, err := os.Stat(cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("failed to check cache directory: %w", err)
	}

	// Calculate total size
	var totalSize int64
	err = filepath.Walk(cacheDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to calculate cache size: %w", err)
	}

	return totalSize, nil
}

// GetCompilerSize returns size of specific compiler in bytes
// Returns 0 if not in cache
func (f *FileSystemCache) GetCompilerSize(language string) (int64, error) {
	compilerPath := filepath.Join(f.GetCacheDir(), language)

	// Check if compiler exists
	_, err := os.Stat(compilerPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil // Not in cache
		}
		return 0, fmt.Errorf("failed to check compiler path: %w", err)
	}

	// Calculate compiler size
	var size int64
	err = filepath.Walk(compilerPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	if err != nil {
		return 0, fmt.Errorf("failed to calculate compiler size: %w", err)
	}

	return size, nil
}

// IsCompilerCached checks if a specific compiler is in cache
func (f *FileSystemCache) IsCompilerCached(language string) bool {
	compilerPath := filepath.Join(f.GetCacheDir(), language)

	info, err := os.Stat(compilerPath)
	if err != nil {
		return false
	}

	return info.IsDir()
}

// GetCompilerBinaryPath returns the path to the compiler binary executable
// This is a helper method for finding the actual compiler binary within the cached directory
func (f *FileSystemCache) GetCompilerBinaryPath(language string) (string, error) {
	compilerPath, err := f.GetCompilerPath(language)
	if err != nil {
		return "", err
	}

	// Map language to binary name
	var binaryName string
	switch language {
	case "zig":
		binaryName = "zig"
	case "kotlin":
		binaryName = filepath.Join("bin", "kotlinc")
	case "swift":
		// SwiftWasm uses system swift + SDK
		binaryName = "swift"
	default:
		return "", fmt.Errorf("unsupported language: %s", language)
	}

	binaryPath := filepath.Join(compilerPath, binaryName)

	// Check if binary exists
	_, err = os.Stat(binaryPath)
	if err != nil {
		return "", fmt.Errorf("compiler binary not found: %w", err)
	}

	return binaryPath, nil
}

// Verify interface implementation at compile time
var _ domain.CacheManager = (*FileSystemCache)(nil)
