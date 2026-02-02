package compilers

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
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

	// Проверяем что путь лежит внутри os.TempDir()
	rel, err := filepath.Rel(os.TempDir(), ws.Path)
	if err != nil {
		t.Fatalf("Rel failed: %v", err)
	}
	if rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		t.Errorf("Workspace path should be within temp dir: %s", ws.Path)
	}

	// Проверяем что директория похожа на наш префикс
	if !strings.HasPrefix(filepath.Base(ws.Path), "go-proxy-compile-") {
		t.Errorf("Workspace base should start with go-proxy-compile-: %s", ws.Path)
	}
}

// Composer 1.
func TestWorkspace_WriteFile_PathTraversal(t *testing.T) {
	ws, err := NewWorkspace("test-traversal")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	if err := ws.WriteFile("../evil.txt", []byte("nope")); err == nil {
		t.Fatal("WriteFile should fail for path traversal")
	}
	if ws.FileExists("../evil.txt") {
		t.Fatal("FileExists should be false for path traversal")
	}
}

// Composer 1.
func TestWorkspace_ReadFile_PathTraversal(t *testing.T) {
	ws, err := NewWorkspace("test-traversal-read")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	if _, err := ws.ReadFile("../evil.txt"); err == nil {
		t.Fatal("ReadFile should fail for path traversal")
	}
}

// Composer 1.
func TestWorkspace_WriteFile_AbsolutePath(t *testing.T) {
	ws, err := NewWorkspace("test-abs-path")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	if err := ws.WriteFile(filepath.Join(os.TempDir(), "evil.txt"), []byte("nope")); err == nil {
		t.Fatal("WriteFile should fail for absolute path")
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
	// Нормализуем пути для кроссплатформенного сравнения
	normalizedListed := make(map[string]bool)
	for _, listed := range listedFiles {
		normalizedListed[filepath.ToSlash(listed)] = true
	}

	for filename := range files {
		normalizedFilename := filepath.ToSlash(filename)
		if !normalizedListed[normalizedFilename] {
			t.Errorf("File %s not found in list", filename)
		}
	}
}

// Composer 1.
func TestWorkspace_ExecuteCommandSeparate_Error(t *testing.T) {
	ws, err := NewWorkspace("test-exec-separate-error")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	ctx := context.Background()

	// Тест с несуществующей командой
	_, _, err = ws.ExecuteCommandSeparate(ctx, "nonexistent-command-12345", "arg1")
	if err == nil {
		t.Fatal("ExecuteCommandSeparate should fail for nonexistent command")
	}
}

// Composer 1.
func TestWorkspace_ExecuteCommandSeparate_CommandFails(t *testing.T) {
	ws, err := NewWorkspace("test-exec-separate-fail")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	ctx := context.Background()

	// Команда которая завершается с ошибкой
	stdout, stderr, err := ws.ExecuteCommandSeparate(ctx, "sh", "-c", "echo stdout; echo stderr >&2; exit 1")
	if err == nil {
		t.Fatal("ExecuteCommandSeparate should return error for failing command")
	}

	// Проверяем что вывод все равно получен
	if len(stdout) == 0 {
		t.Error("Stdout should not be empty even if command fails")
	}
	if len(stderr) == 0 {
		t.Error("Stderr should not be empty even if command fails")
	}
}

// Composer 1.
func TestWorkspace_WriteFile_LongPath(t *testing.T) {
	ws, err := NewWorkspace("test-long-path")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	// Создаем очень длинный путь
	longPath := "a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z/file.txt"
	content := []byte("test content")

	err = ws.WriteFile(longPath, content)
	if err != nil {
		t.Fatalf("WriteFile failed for long path: %v", err)
	}

	// Проверяем что файл создан
	if !ws.FileExists(longPath) {
		t.Fatal("File with long path was not created")
	}

	// Проверяем содержимое
	readContent, err := ws.ReadFile(longPath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(readContent) != string(content) {
		t.Errorf("Content mismatch: expected %q, got %q", string(content), string(readContent))
	}
}

// Composer 1.
func TestWorkspace_WriteFile_EmptyContent(t *testing.T) {
	ws, err := NewWorkspace("test-empty-content")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	// Записываем пустой файл
	err = ws.WriteFile("empty.txt", []byte{})
	if err != nil {
		t.Fatalf("WriteFile failed for empty content: %v", err)
	}

	// Проверяем что файл создан
	if !ws.FileExists("empty.txt") {
		t.Fatal("Empty file was not created")
	}

	// Проверяем содержимое
	readContent, err := ws.ReadFile("empty.txt")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if len(readContent) != 0 {
		t.Errorf("Expected empty file, got %d bytes", len(readContent))
	}
}

// Composer 1.
func TestWorkspace_ReadFile_LargeFile(t *testing.T) {
	ws, err := NewWorkspace("test-large-file")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	// Создаем большой файл (1MB)
	largeContent := make([]byte, 1024*1024)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	err = ws.WriteFile("large.bin", largeContent)
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Читаем большой файл
	readContent, err := ws.ReadFile("large.bin")
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if len(readContent) != len(largeContent) {
		t.Errorf("Expected %d bytes, got %d", len(largeContent), len(readContent))
	}

	// Проверяем первые и последние байты
	if readContent[0] != largeContent[0] {
		t.Error("First byte mismatch")
	}
	if readContent[len(readContent)-1] != largeContent[len(largeContent)-1] {
		t.Error("Last byte mismatch")
	}
}

// Composer 1.
func TestWorkspace_ListFiles_WithErrors(t *testing.T) {
	ws, err := NewWorkspace("test-list-errors")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	// Создаем файлы
	err = ws.WriteFile("file1.txt", []byte("content1"))
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	err = ws.WriteFile("subdir/file2.txt", []byte("content2"))
	if err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Получаем список файлов
	files, err := ws.ListFiles()
	if err != nil {
		t.Fatalf("ListFiles failed: %v", err)
	}

	// Проверяем что файлы найдены
	if len(files) < 2 {
		t.Errorf("Expected at least 2 files, got %d", len(files))
	}
}

// Composer 1.
func TestWorkspace_ExecuteCommand_EmptyOutput(t *testing.T) {
	ws, err := NewWorkspace("test-exec-empty")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	ctx := context.Background()

	// Команда которая не выводит ничего
	output, err := ws.ExecuteCommand(ctx, "true")
	if err != nil {
		t.Fatalf("ExecuteCommand failed: %v", err)
	}

	// Вывод может быть пустым или содержать перевод строки
	_ = output
}

// Composer 1.
func TestWorkspace_ExecuteCommand_WithErrorOutput(t *testing.T) {
	ws, err := NewWorkspace("test-exec-error-output")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	ctx := context.Background()

	// Команда которая выводит в stderr
	output, err := ws.ExecuteCommand(ctx, "sh", "-c", "echo 'error message' >&2; exit 1")
	if err == nil {
		t.Fatal("ExecuteCommand should return error for failing command")
	}

	// Проверяем что вывод содержит сообщение об ошибке
	outputStr := string(output)
	if outputStr == "" {
		t.Error("Command output should not be empty")
	}
}

func TestWorkspace_ExecuteCommand_TruncatesLargeOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("uses sh/head/tr")
	}

	ws, err := NewWorkspace("test-exec-truncate")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	ctx := context.Background()

	// 3MB output, should be truncated by maxCommandOutputBytes (2MB) + note.
	output, err := ws.ExecuteCommand(ctx, "sh", "-c", "head -c 3000000 /dev/zero | tr '\\000' 'a'")
	if err != nil {
		t.Fatalf("ExecuteCommand failed: %v", err)
	}
	if len(output) <= maxCommandOutputBytes {
		t.Fatalf("expected truncated output to exceed %d bytes due to note, got %d", maxCommandOutputBytes, len(output))
	}
	if !strings.Contains(string(output), "output truncated") {
		t.Fatal("expected output to include truncation note")
	}
}

func TestWorkspace_ReadFile_BlocksSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlinks may be unavailable on Windows without special permissions")
	}

	ws, err := NewWorkspace("test-symlink-escape")
	if err != nil {
		t.Fatalf("NewWorkspace failed: %v", err)
	}
	defer ws.Cleanup()

	outsidePath := filepath.Join(t.TempDir(), "outside.txt")
	if err := os.WriteFile(outsidePath, []byte("secret"), 0644); err != nil {
		t.Fatalf("write outside file: %v", err)
	}

	linkPath := filepath.Join(ws.Path, "leak.txt")
	if err := os.Symlink(outsidePath, linkPath); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	_, err = ws.ReadFile("leak.txt")
	if err == nil {
		t.Fatal("expected ReadFile to fail for symlink escape")
	}
}
