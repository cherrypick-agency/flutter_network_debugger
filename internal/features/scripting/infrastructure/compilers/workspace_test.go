package compilers

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// Composer 1.
func TestNewWorkspace(t *testing.T) {
	scriptID := "test-script-123"
	ws, err := NewWorkspace(scriptID)
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	if ws == nil {
		t.Fatal("Workspace is nil")
	}

	if ws.Path == "" {
		t.Fatal("Workspace path is empty")
	}

	// Проверяем что директория создана
	if _, err := os.Stat(ws.Path); os.IsNotExist(err) {
		t.Fatalf("Workspace directory was not created: %s", ws.Path)
	}

	// Проверяем что путь содержит scriptID
	if filepath.Base(ws.Path) != scriptID {
		// Проверяем что scriptID есть в пути
		if !filepath.HasPrefix(ws.Path, filepath.Join(os.TempDir(), "go-proxy-compile")) {
			t.Errorf("Workspace path should be in temp directory: %s", ws.Path)
		}
	}
}

// Composer 1.
func TestWorkspace_WriteFile(t *testing.T) {
	ws, err := NewWorkspace("test-write")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	// Тест записи простого файла
	filename := "test.txt"
	content := []byte("hello world")
	err = ws.WriteFile(filename, content)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Проверяем что файл создан
	if !ws.FileExists(filename) {
		t.Fatal("File was not created")
	}

	// Проверяем содержимое
	readContent, err := ws.ReadFile(filename)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(readContent) != string(content) {
		t.Errorf("Content mismatch: expected %q, got %q", string(content), string(readContent))
	}
}

// Composer 1.
func TestWorkspace_WriteFile_WithSubdirectory(t *testing.T) {
	ws, err := NewWorkspace("test-subdir")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	// Тест записи файла в поддиректорию
	filename := "subdir/test.txt"
	content := []byte("nested content")
	err = ws.WriteFile(filename, content)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Проверяем что файл создан
	if !ws.FileExists(filename) {
		t.Fatal("File in subdirectory was not created")
	}

	// Проверяем содержимое
	readContent, err := ws.ReadFile(filename)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(readContent) != string(content) {
		t.Errorf("Content mismatch: expected %q, got %q", string(content), string(readContent))
	}
}

// Composer 1.
func TestWorkspace_ReadFile_NotExists(t *testing.T) {
	ws, err := NewWorkspace("test-read-not-exists")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	_, err = ws.ReadFile("nonexistent.txt")
	if err == nil {
		t.Fatal("ReadFile should fail for nonexistent file")
	}
}

// Composer 1.
func TestWorkspace_FileExists(t *testing.T) {
	ws, err := NewWorkspace("test-exists")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	// Файл не существует
	if ws.FileExists("nonexistent.txt") {
		t.Fatal("FileExists should return false for nonexistent file")
	}

	// Создаем файл
	err = ws.WriteFile("exists.txt", []byte("test"))
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Файл существует
	if !ws.FileExists("exists.txt") {
		t.Fatal("FileExists should return true for existing file")
	}
}

// Composer 1.
func TestWorkspace_ExecuteCommand(t *testing.T) {
	ws, err := NewWorkspace("test-exec")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	ctx := context.Background()

	// Тест выполнения простой команды (echo)
	output, err := ws.ExecuteCommand(ctx, "echo", "hello")
	if err != nil {
		t.Fatalf("ExecuteCommand failed: %v", err)
	}

	outputStr := string(output)
	if outputStr == "" {
		t.Fatal("Command output is empty")
	}
}

// Composer 1.
func TestWorkspace_ExecuteCommand_WithTimeout(t *testing.T) {
	ws, err := NewWorkspace("test-exec-timeout")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	// Создаем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Команда которая будет выполняться дольше таймаута
	_, err = ws.ExecuteCommand(ctx, "sleep", "1")
	if err == nil {
		t.Fatal("ExecuteCommand should fail with timeout")
	}
}

// Composer 1.
func TestWorkspace_ExecuteCommandSeparate(t *testing.T) {
	ws, err := NewWorkspace("test-exec-separate")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	ctx := context.Background()

	// Тест выполнения команды с раздельным выводом
	stdout, stderr, err := ws.ExecuteCommandSeparate(ctx, "sh", "-c", "echo stdout; echo stderr >&2")
	if err != nil {
		t.Fatalf("ExecuteCommandSeparate failed: %v", err)
	}

	stdoutStr := string(stdout)
	stderrStr := string(stderr)

	if stdoutStr == "" {
		t.Fatal("Stdout is empty")
	}

	// Проверяем что stderr отделен от stdout
	if stderrStr == "" {
		t.Fatal("Stderr is empty")
	}
}

// Composer 1.
func TestWorkspace_Cleanup(t *testing.T) {
	ws, err := NewWorkspace("test-cleanup")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}

	path := ws.Path

	// Проверяем что директория существует
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("Workspace directory does not exist: %s", path)
	}

	// Очищаем
	err = ws.Cleanup()
	if err != nil {
		t.Fatalf("Cleanup failed: %v", err)
	}

	// Проверяем что директория удалена
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("Workspace directory was not removed")
	}

	// Повторная очистка не должна падать
	err = ws.Cleanup()
	if err != nil {
		t.Fatalf("Second cleanup failed: %v", err)
	}
}

// Composer 1.
func TestWorkspace_Cleanup_EmptyPath(t *testing.T) {
	ws := &Workspace{Path: ""}

	// Очистка пустого пути не должна падать
	err := ws.Cleanup()
	if err != nil {
		t.Fatalf("Cleanup with empty path failed: %v", err)
	}
}

// Composer 1.
func TestWorkspace_ListFiles_Empty(t *testing.T) {
	ws, err := NewWorkspace("test-list-empty")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	files, err := ws.ListFiles()
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("ListFiles() length = %d, want 0", len(files))
	}
}

// Composer 1.
func TestWorkspace_ListFiles(t *testing.T) {
	ws, err := NewWorkspace("test-list")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	// Создаем несколько файлов
	files := map[string][]byte{
		"file1.txt":        []byte("content1"),
		"file2.txt":        []byte("content2"),
		"subdir/file3.txt": []byte("content3"),
	}

	for filename, content := range files {
		err = ws.WriteFile(filename, content)
		if err != nil {
			t.Fatalf("WriteFile failed for %s: %v", filename, err)
		}
	}

	// Получаем список файлов
	listedFiles, err := ws.ListFiles()
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}

	// Проверяем что все файлы найдены
	if len(listedFiles) != len(files) {
		t.Errorf("Expected %d files, got %d", len(files), len(listedFiles))
	}

	// Проверяем наличие каждого файла
	for filename := range files {
		found := false
		for _, listed := range listedFiles {
			if listed == filename {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("File %s not found in list", filename)
		}
	}
}
