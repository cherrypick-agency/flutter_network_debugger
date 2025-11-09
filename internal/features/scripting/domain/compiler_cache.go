package domain

// CacheManager is the PORT interface for managing compiler cache
// Default implementation uses filesystem cache at ~/.cache/network-debugger
type CacheManager interface {
	// GetCacheDir returns the base cache directory
	GetCacheDir() string

	// GetCompilerPath returns the path to a specific compiler in cache
	// Returns error if compiler not in cache
	GetCompilerPath(language string) (string, error)

	// EnsureCacheDir ensures the cache directory structure exists
	EnsureCacheDir() error

	// Clear removes a specific compiler from cache
	Clear(language string) error

	// ClearAll removes all compilers from cache
	ClearAll() error

	// GetCacheSize returns total size of cache in bytes
	GetCacheSize() (int64, error)

	// GetCompilerSize returns size of specific compiler in bytes
	// Returns 0 if not in cache
	GetCompilerSize(language string) (int64, error)

	// IsCompilerCached checks if a specific compiler is in cache
	IsCompilerCached(language string) bool
}
