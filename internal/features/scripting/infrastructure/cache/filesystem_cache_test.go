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

	// Проверяем что директория создана
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

	// Проверяем что используется домашняя директория
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

	// Создаем директорию компилятора
	err = os.MkdirAll(compilerPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create compiler directory: %v", err)
	}

	// Получаем путь
	path, err := cache.GetCompilerPath(language)
	if err != nil {
		t.Fatalf("GetCompilerPath failed: %v", err)
	}

	if path != compilerPath {
		t.Errorf("Expected path %q, got %q", compilerPath, path)
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

	// Создаем файл вместо директории
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

	// Создаем директорию компилятора
	err = os.MkdirAll(compilerPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create compiler directory: %v", err)
	}

	// Очищаем
	err = cache.Clear(language)
	if err != nil {
		t.Fatalf("Clear failed: %v", err)
	}

	// Проверяем что директория удалена
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

	// Очистка несуществующего компилятора не должна падать
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

	// Создаем несколько компиляторов
	languages := []string{"rust", "go", "zig"}
	for _, lang := range languages {
		compilerPath := filepath.Join(cache.GetCacheDir(), lang)
		err = os.MkdirAll(compilerPath, 0755)
		if err != nil {
			t.Fatalf("Failed to create compiler directory: %v", err)
		}
	}

	// Очищаем все
	err = cache.ClearAll()
	if err != nil {
		t.Fatalf("ClearAll failed: %v", err)
	}

	// Проверяем что все директории удалены
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

	// Очистка пустого кеша не должна падать
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

	// Пустой кеш должен иметь размер 0
	size, err := cache.GetCacheSize()
	if err != nil {
		t.Fatalf("GetCacheSize failed: %v", err)
	}

	if size != 0 {
		t.Errorf("Expected cache size 0, got %d", size)
	}

	// Создаем файл в кеше
	testFile := filepath.Join(cache.GetCacheDir(), "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Проверяем размер
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

	// Несуществующий компилятор должен иметь размер 0
	size, err := cache.GetCompilerSize(language)
	if err != nil {
		t.Fatalf("GetCompilerSize failed: %v", err)
	}

	if size != 0 {
		t.Errorf("Expected compiler size 0, got %d", size)
	}

	// Создаем директорию компилятора с файлом
	err = os.MkdirAll(compilerPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create compiler directory: %v", err)
	}

	testFile := filepath.Join(compilerPath, "binary")
	err = os.WriteFile(testFile, []byte("binary content"), 0644)
	if err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	// Проверяем размер
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

	// Компилятор не в кеше
	if cache.IsCompilerCached(language) {
		t.Error("IsCompilerCached should return false for uncached compiler")
	}

	// Создаем директорию компилятора
	compilerPath := filepath.Join(cache.GetCacheDir(), language)
	err = os.MkdirAll(compilerPath, 0755)
	if err != nil {
		t.Fatalf("Failed to create compiler directory: %v", err)
	}

	// Компилятор в кеше
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
			wantErr: true, // binary не существует
		},
		{
			name:     "kotlin compiler",
			language: "kotlin",
			binary:   filepath.Join("bin", "kotlinc"),
			setup: func(path string) error {
				binaryPath := filepath.Join(path, "bin", "kotlinc")
				return os.MkdirAll(filepath.Dir(binaryPath), 0755)
			},
			wantErr: true, // binary не существует
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
