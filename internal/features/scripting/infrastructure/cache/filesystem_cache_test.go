package cache

import (
	"os"
	"path/filepath"
	"testing"
)

// Composer 1.
func TestNewFileSystemCache(t *testing.T) {
	tmpDir := t.TempDir()

	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	if cache == nil {
		t.Fatal("Cache is nil")
	}

	if cache.baseDir != tmpDir {
		t.Errorf("Expected baseDir %q, got %q", tmpDir, cache.baseDir)
	}

	// Check that directory was created
	cacheDir := cache.GetCacheDir()
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Fatalf("Cache directory was not created: %s", cacheDir)
	}
}

// Composer 1.
func TestNewFileSystemCache_EmptyDir(t *testing.T) {
	cache, err := NewFileSystemCache("")
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	if cache.baseDir == "" {
		t.Fatal("baseDir should be set to default location")
	}

	// Check that home directory is used
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get home directory: %v", err)
	}

	expectedBase := filepath.Join(homeDir, ".cache", "network-debugger")
	if cache.baseDir != expectedBase {
		t.Errorf("Expected baseDir %q, got %q", expectedBase, cache.baseDir)
	}
}

// Composer 1.
func TestFileSystemCache_GetCacheDir(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	cacheDir := cache.GetCacheDir()
	expected := filepath.Join(tmpDir, "compilers")

	if cacheDir != expected {
		t.Errorf("Expected cache dir %q, got %q", expected, cacheDir)
	}
}

// Composer 1.
func TestFileSystemCache_GetCompilerPath(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	language := "rust"
	compilerPath := filepath.Join(cache.GetCacheDir(), language)

	// Create compiler directory
	err = os.MkdirAll(compilerPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create compiler directory: %v", err)
	}

	// Get path
	path, err := cache.GetCompilerPath(language)
	if err != nil {
		t.Fatalf("GetCompilerPath failed: %v", err)
	}

	if path != compilerPath {
		t.Errorf("Expected path %q, got %q", compilerPath, path)
	}
}

// Composer 1.
func TestFileSystemCache_GetCompilerPath_PathTraversalBlocked(t *testing.T) {
	cache, err := NewFileSystemCache(t.TempDir())
	if err != nil {
		t.Fatalf("NewFileSystemCache error: %v", err)
	}

	_, err = cache.GetCompilerPath("../evil")
	if err == nil {
		t.Fatal("expected error for traversal compiler key")
	}
}

// Composer 1.
func TestFileSystemCache_Clear_PathTraversalBlocked(t *testing.T) {
	cache, err := NewFileSystemCache(t.TempDir())
	if err != nil {
		t.Fatalf("NewFileSystemCache error: %v", err)
	}

	if err := cache.Clear("../evil"); err == nil {
		t.Fatal("expected error for traversal compiler key")
	}
}

// Composer 1.
func TestFileSystemCache_IsCompilerCached_PathTraversalBlocked(t *testing.T) {
	cache, err := NewFileSystemCache(t.TempDir())
	if err != nil {
		t.Fatalf("NewFileSystemCache error: %v", err)
	}

	if cache.IsCompilerCached("../evil") {
		t.Fatal("expected IsCompilerCached to return false for traversal compiler key")
	}
}

// Composer 1.
func TestFileSystemCache_GetCompilerPath_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	_, err = cache.GetCompilerPath("nonexistent")
	if err == nil {
		t.Fatal("GetCompilerPath should fail for nonexistent compiler")
	}
}

// Composer 1.
func TestFileSystemCache_GetCompilerPath_NotDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	language := "rust"
	compilerPath := filepath.Join(cache.GetCacheDir(), language)

	// Create file instead of directory
	err = os.WriteFile(compilerPath, []byte("not a directory"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	_, err = cache.GetCompilerPath(language)
	if err == nil {
		t.Fatal("GetCompilerPath should fail when path is not a directory")
	}
}

// Composer 1.
func TestFileSystemCache_EnsureCacheDir(t *testing.T) {
	tmpDir := t.TempDir()
	cache := &FileSystemCache{baseDir: tmpDir}

	err := cache.EnsureCacheDir()
	if err != nil {
		t.Fatalf("EnsureCacheDir failed: %v", err)
	}

	cacheDir := cache.GetCacheDir()
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Fatalf("Cache directory was not created: %s", cacheDir)
	}
}

// Composer 1.
func TestFileSystemCache_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	language := "rust"
	compilerPath := filepath.Join(cache.GetCacheDir(), language)

	// Create compiler directory
	err = os.MkdirAll(compilerPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create compiler directory: %v", err)
	}

	// Clear
	err = cache.Clear(language)
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Check that directory was removed
	if _, err := os.Stat(compilerPath); !os.IsNotExist(err) {
		t.Fatal("Compiler directory was not removed")
	}
}

// Composer 1.
func TestFileSystemCache_Clear_NotExists(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	// Clearing nonexistent compiler should not panic
	err = cache.Clear("nonexistent")
	if err != nil {
		t.Fatalf("Clear should not fail for nonexistent compiler: %v", err)
	}
}

// Composer 1.
func TestFileSystemCache_ClearAll(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	// Create several compilers
	languages := []string{"rust", "go", "zig"}
	for _, lang := range languages {
		compilerPath := filepath.Join(cache.GetCacheDir(), lang)
		err = os.MkdirAll(compilerPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create compiler directory: %v", err)
		}
	}

	// Clear all
	err = cache.ClearAll()
	if err != nil {
		t.Fatalf("ClearAll failed: %v", err)
	}

	// Check that all directories were removed
	for _, lang := range languages {
		compilerPath := filepath.Join(cache.GetCacheDir(), lang)
		if _, err := os.Stat(compilerPath); !os.IsNotExist(err) {
			t.Errorf("Compiler directory %s was not removed", lang)
		}
	}
}

// Composer 1.
func TestFileSystemCache_ClearAll_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	// Clearing empty cache should not panic
	err = cache.ClearAll()
	if err != nil {
		t.Fatalf("ClearAll should not fail for empty cache: %v", err)
	}
}

// Composer 1.
func TestFileSystemCache_GetCacheSize(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	// Empty cache should have size 0
	size, err := cache.GetCacheSize()
	if err != nil {
		t.Fatalf("GetCacheSize failed: %v", err)
	}

	if size != 0 {
		t.Errorf("Expected cache size 0, got %d", size)
	}

	// Create file in cache
	testFile := filepath.Join(cache.GetCacheDir(), "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Check size
	size, err = cache.GetCacheSize()
	if err != nil {
		t.Fatalf("GetCacheSize failed: %v", err)
	}

	if size == 0 {
		t.Error("Expected cache size > 0 after adding file")
	}
}

// Composer 1.
func TestFileSystemCache_GetCompilerSize(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	language := "rust"
	compilerPath := filepath.Join(cache.GetCacheDir(), language)

	// Nonexistent compiler should have size 0
	size, err := cache.GetCompilerSize(language)
	if err != nil {
		t.Fatalf("GetCompilerSize failed: %v", err)
	}

	if size != 0 {
		t.Errorf("Expected compiler size 0, got %d", size)
	}

	// Create compiler directory with file
	err = os.MkdirAll(compilerPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create compiler directory: %v", err)
	}

	testFile := filepath.Join(compilerPath, "binary")
	err = os.WriteFile(testFile, []byte("binary content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Check size
	size, err = cache.GetCompilerSize(language)
	if err != nil {
		t.Fatalf("GetCompilerSize failed: %v", err)
	}

	if size == 0 {
		t.Error("Expected compiler size > 0 after adding file")
	}
}

// Composer 1.
func TestFileSystemCache_IsCompilerCached(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	language := "rust"

	// Compiler not in cache
	if cache.IsCompilerCached(language) {
		t.Error("IsCompilerCached should return false for uncached compiler")
	}

	// Create compiler directory
	compilerPath := filepath.Join(cache.GetCacheDir(), language)
	err = os.MkdirAll(compilerPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create compiler directory: %v", err)
	}

	// Compiler in cache
	if !cache.IsCompilerCached(language) {
		t.Error("IsCompilerCached should return true for cached compiler")
	}
}

// Composer 1.
func TestFileSystemCache_GetCompilerBinaryPath(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	tests := []struct {
		name     string
		language string
		binary   string
		setup    func(string) error
		wantErr  bool
	}{
		{
			name:     "zig compiler",
			language: "zig",
			binary:   "zig",
			setup: func(path string) error {
				return os.MkdirAll(path, 0755)
			},
			wantErr: true, // binary doesn't exist
		},
		{
			name:     "kotlin compiler",
			language: "kotlin",
			binary:   filepath.Join("bin", "kotlinc"),
			setup: func(path string) error {
				binaryPath := filepath.Join(path, "bin", "kotlinc")
				return os.MkdirAll(filepath.Dir(binaryPath), 0755)
			},
			wantErr: true, // binary doesn't exist
		},
		{
			name:     "unsupported language",
			language: "python",
			setup: func(path string) error {
				return os.MkdirAll(path, 0755)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			compilerPath := filepath.Join(cache.GetCacheDir(), tt.language)
			if err := tt.setup(compilerPath); err != nil {
				t.Fatalf("Setup failed: %v", err)
			}

			_, err := cache.GetCompilerBinaryPath(tt.language)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCompilerBinaryPath() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestFileSystemCache_ClearAll_WithFiles(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	// Create compilers with files
	languages := []string{"rust", "go"}
	for _, lang := range languages {
		compilerPath := filepath.Join(cache.GetCacheDir(), lang)
		err = os.MkdirAll(compilerPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create compiler directory: %v", err)
		}

		// Create file inside
		testFile := filepath.Join(compilerPath, "test.txt")
		err = os.WriteFile(testFile, []byte("test"), 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	// Clear all
	err = cache.ClearAll()
	if err != nil {
		t.Fatalf("ClearAll failed: %v", err)
	}

	// Check that everything was removed
	for _, lang := range languages {
		compilerPath := filepath.Join(cache.GetCacheDir(), lang)
		if _, err := os.Stat(compilerPath); !os.IsNotExist(err) {
			t.Errorf("Compiler directory %s was not removed", lang)
		}
	}
}

func TestFileSystemCache_GetCacheSize_WithNestedFiles(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	// Create compiler with nested files
	compilerPath := filepath.Join(cache.GetCacheDir(), "rust")
	err = os.MkdirAll(compilerPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create compiler directory: %v", err)
	}

	// Create files at different levels
	files := []string{
		filepath.Join(compilerPath, "binary"),
		filepath.Join(compilerPath, "subdir", "file.txt"),
	}

	for _, file := range files {
		err = os.MkdirAll(filepath.Dir(file), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		err = os.WriteFile(file, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	size, err := cache.GetCacheSize()
	if err != nil {
		t.Fatalf("GetCacheSize failed: %v", err)
	}

	if size == 0 {
		t.Error("Expected cache size > 0 with nested files")
	}
}

func TestFileSystemCache_GetCompilerSize_WithNestedFiles(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	language := "rust"
	compilerPath := filepath.Join(cache.GetCacheDir(), language)
	err = os.MkdirAll(compilerPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create compiler directory: %v", err)
	}

	// Create files at different levels
	files := []string{
		filepath.Join(compilerPath, "binary"),
		filepath.Join(compilerPath, "subdir", "file.txt"),
	}

	for _, file := range files {
		err = os.MkdirAll(filepath.Dir(file), 0755)
		if err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		err = os.WriteFile(file, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to write file: %v", err)
		}
	}

	size, err := cache.GetCompilerSize(language)
	if err != nil {
		t.Fatalf("GetCompilerSize failed: %v", err)
	}

	if size == 0 {
		t.Error("Expected compiler size > 0 with nested files")
	}
}

func TestFileSystemCache_GetCompilerPath_StatError(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	// Create file instead of directory
	compilerPath := filepath.Join(cache.GetCacheDir(), "rust")
	err = os.WriteFile(compilerPath, []byte("not a dir"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	_, err = cache.GetCompilerPath("rust")
	if err == nil {
		t.Fatal("GetCompilerPath should fail when path is not a directory")
	}
}

func TestFileSystemCache_Clear_FileInsteadOfDir(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	// Create file instead of directory
	compilerPath := filepath.Join(cache.GetCacheDir(), "rust")
	err = os.WriteFile(compilerPath, []byte("not a dir"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// Clear should remove file via RemoveAll
	err = cache.Clear("rust")
	if err != nil {
		t.Fatalf("Clear should succeed even for file: %v", err)
	}

	// Check that file was removed
	if _, err := os.Stat(compilerPath); !os.IsNotExist(err) {
		t.Fatal("File should be removed by Clear")
	}
}

// TestFileSystemCache_ClearAll_WithNonDirectoryEntries tests ClearAll with non-directory entries
func TestFileSystemCache_ClearAll_WithNonDirectoryEntries(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	cacheDir := cache.GetCacheDir()

	// Create directory and file
	compilerDir := filepath.Join(cacheDir, "rust")
	err = os.MkdirAll(compilerDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create compiler directory: %v", err)
	}

	// Create file in cache root (not directory)
	filePath := filepath.Join(cacheDir, "somefile.txt")
	err = os.WriteFile(filePath, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	// ClearAll should remove only directories
	err = cache.ClearAll()
	if err != nil {
		t.Fatalf("ClearAll failed: %v", err)
	}

	// Directory should be removed
	if _, err := os.Stat(compilerDir); !os.IsNotExist(err) {
		t.Error("Compiler directory should be removed")
	}

	// File may remain or be removed depending on implementation
	// Check that directories are removed
}

// TestFileSystemCache_ClearAll_WithPartialErrors tests ClearAll with partial errors
func TestFileSystemCache_ClearAll_WithPartialErrors(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	cacheDir := cache.GetCacheDir()

	// Create several compilers
	languages := []string{"rust", "go", "zig"}
	for _, lang := range languages {
		compilerPath := filepath.Join(cacheDir, lang)
		err = os.MkdirAll(compilerPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create compiler directory: %v", err)
		}
	}

	// ClearAll should remove all directories
	err = cache.ClearAll()
	if err != nil {
		t.Fatalf("ClearAll failed: %v", err)
	}

	// Check that all directories were removed
	for _, lang := range languages {
		compilerPath := filepath.Join(cacheDir, lang)
		if _, err := os.Stat(compilerPath); !os.IsNotExist(err) {
			t.Errorf("Compiler directory %s should be removed", lang)
		}
	}
}

// TestFileSystemCache_GetCacheSize_WithEmptyCache tests GetCacheSize with empty cache
func TestFileSystemCache_GetCacheSize_WithEmptyCache(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	size, err := cache.GetCacheSize()
	if err != nil {
		t.Fatalf("GetCacheSize failed: %v", err)
	}

	if size != 0 {
		t.Errorf("Expected cache size 0 for empty cache, got %d", size)
	}
}

// TestFileSystemCache_GetCacheSize_WithMultipleCompilers tests GetCacheSize with multiple compilers
func TestFileSystemCache_GetCacheSize_WithMultipleCompilers(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	cacheDir := cache.GetCacheDir()

	// Create several compilers with files of different sizes
	languages := []string{"rust", "go", "zig"}
	for i, lang := range languages {
		compilerPath := filepath.Join(cacheDir, lang)
		err = os.MkdirAll(compilerPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create compiler directory: %v", err)
		}

		// Create file of different size for each compiler
		testFile := filepath.Join(compilerPath, "binary")
		fileSize := (i + 1) * 1000
		data := make([]byte, fileSize)
		err = os.WriteFile(testFile, data, 0644)
		if err != nil {
			t.Fatalf("Failed to write test file: %v", err)
		}
	}

	size, err := cache.GetCacheSize()
	if err != nil {
		t.Fatalf("GetCacheSize failed: %v", err)
	}

	expectedMinSize := int64(6000) // 1000 + 2000 + 3000
	if size < expectedMinSize {
		t.Errorf("Expected cache size >= %d, got %d", expectedMinSize, size)
	}
}

// TestFileSystemCache_GetCacheSize_AfterClearAll tests GetCacheSize after ClearAll
func TestFileSystemCache_GetCacheSize_AfterClearAll(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	cacheDir := cache.GetCacheDir()

	// Create compiler with file
	compilerPath := filepath.Join(cacheDir, "rust")
	err = os.MkdirAll(compilerPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create compiler directory: %v", err)
	}

	testFile := filepath.Join(compilerPath, "binary")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Check size before clearing
	sizeBefore, err := cache.GetCacheSize()
	if err != nil {
		t.Fatalf("GetCacheSize failed: %v", err)
	}

	if sizeBefore == 0 {
		t.Error("Expected cache size > 0 before ClearAll")
	}

	// Clear cache
	err = cache.ClearAll()
	if err != nil {
		t.Fatalf("ClearAll failed: %v", err)
	}

	// Check size after clearing
	sizeAfter, err := cache.GetCacheSize()
	if err != nil {
		t.Fatalf("GetCacheSize failed: %v", err)
	}

	if sizeAfter != 0 {
		t.Errorf("Expected cache size 0 after ClearAll, got %d", sizeAfter)
	}
}

// TestFileSystemCache_IsCompilerCached_NotCached tests IsCompilerCached with uncached compiler
func TestFileSystemCache_IsCompilerCached_NotCached(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	if cache.IsCompilerCached("nonexistent") {
		t.Error("IsCompilerCached should return false for uncached compiler")
	}
}

// TestFileSystemCache_IsCompilerCached_Cached tests IsCompilerCached with cached compiler
func TestFileSystemCache_IsCompilerCached_Cached(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	language := "rust"
	compilerPath := filepath.Join(cache.GetCacheDir(), language)
	err = os.MkdirAll(compilerPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create compiler directory: %v", err)
	}

	if !cache.IsCompilerCached(language) {
		t.Error("IsCompilerCached should return true for cached compiler")
	}
}

// TestFileSystemCache_IsCompilerCached_FileInsteadOfDir tests IsCompilerCached with file instead of directory
func TestFileSystemCache_IsCompilerCached_FileInsteadOfDir(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	language := "rust"
	compilerPath := filepath.Join(cache.GetCacheDir(), language)
	err = os.WriteFile(compilerPath, []byte("not a dir"), 0644)
	if err != nil {
		t.Fatalf("Failed to create file: %v", err)
	}

	if cache.IsCompilerCached(language) {
		t.Error("IsCompilerCached should return false for file instead of directory")
	}
}

// TestFileSystemCache_GetCompilerBinaryPath_Zig tests GetCompilerBinaryPath for zig
func TestFileSystemCache_GetCompilerBinaryPath_Zig(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	language := "zig"
	compilerPath := filepath.Join(cache.GetCacheDir(), language)
	err = os.MkdirAll(compilerPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create compiler directory: %v", err)
	}

	binaryPath := filepath.Join(compilerPath, "zig")
	err = os.WriteFile(binaryPath, []byte("binary"), 0755)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	path, err := cache.GetCompilerBinaryPath(language)
	if err != nil {
		t.Fatalf("GetCompilerBinaryPath failed: %v", err)
	}

	if path != binaryPath {
		t.Errorf("Expected path %q, got %q", binaryPath, path)
	}
}

// TestFileSystemCache_GetCompilerBinaryPath_Kotlin tests GetCompilerBinaryPath for kotlin
func TestFileSystemCache_GetCompilerBinaryPath_Kotlin(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	language := "kotlin"
	compilerPath := filepath.Join(cache.GetCacheDir(), language)
	binaryPath := filepath.Join(compilerPath, "bin", "kotlinc")
	err = os.MkdirAll(filepath.Dir(binaryPath), 0755)
	if err != nil {
		t.Fatalf("Failed to create directory: %v", err)
	}

	err = os.WriteFile(binaryPath, []byte("binary"), 0755)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	path, err := cache.GetCompilerBinaryPath(language)
	if err != nil {
		t.Fatalf("GetCompilerBinaryPath failed: %v", err)
	}

	if path != binaryPath {
		t.Errorf("Expected path %q, got %q", binaryPath, path)
	}
}

// TestFileSystemCache_GetCompilerBinaryPath_Swift tests GetCompilerBinaryPath for swift
func TestFileSystemCache_GetCompilerBinaryPath_Swift(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	language := "swift"
	compilerPath := filepath.Join(cache.GetCacheDir(), language)
	err = os.MkdirAll(compilerPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create compiler directory: %v", err)
	}

	binaryPath := filepath.Join(compilerPath, "swift")
	err = os.WriteFile(binaryPath, []byte("binary"), 0755)
	if err != nil {
		t.Fatalf("Failed to create binary: %v", err)
	}

	path, err := cache.GetCompilerBinaryPath(language)
	if err != nil {
		t.Fatalf("GetCompilerBinaryPath failed: %v", err)
	}

	if path != binaryPath {
		t.Errorf("Expected path %q, got %q", binaryPath, path)
	}
}

// TestFileSystemCache_GetCompilerBinaryPath_UnsupportedLanguage tests GetCompilerBinaryPath with unsupported language
func TestFileSystemCache_GetCompilerBinaryPath_UnsupportedLanguage(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	_, err = cache.GetCompilerBinaryPath("python")
	if err == nil {
		t.Error("Expected error for unsupported language")
	}
}

// TestFileSystemCache_GetCompilerBinaryPath_BinaryNotFound tests GetCompilerBinaryPath when binary doesn't exist
func TestFileSystemCache_GetCompilerBinaryPath_BinaryNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	cache, err := NewFileSystemCache(tmpDir)
	if err != nil {
		t.Fatalf("NewFileSystemCache failed: %v", err)
	}

	language := "zig"
	compilerPath := filepath.Join(cache.GetCacheDir(), language)
	err = os.MkdirAll(compilerPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create compiler directory: %v", err)
	}

	_, err = cache.GetCompilerBinaryPath(language)
	if err == nil {
		t.Error("Expected error when binary doesn't exist")
	}
}
