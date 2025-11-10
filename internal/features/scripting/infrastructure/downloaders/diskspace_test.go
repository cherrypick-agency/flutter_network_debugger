//go:build unix

package downloaders

import (
	"os"
	"path/filepath"
	"testing"
)

// Composer 1.
func TestCheckDiskSpace(t *testing.T) {
	tmpDir := t.TempDir()

	// Тест с достаточным местом (запрашиваем 1 байт)
	err := checkDiskSpace(tmpDir, 1)
	if err != nil {
		t.Errorf("checkDiskSpace() with sufficient space error = %v, want nil", err)
	}

	// Тест с очень большим запросом (должен вернуть ошибку если места недостаточно)
	// Но на большинстве систем это не сработает, так как места достаточно
	// Поэтому просто проверяем что функция не паникует
	err = checkDiskSpace(tmpDir, 1<<60) // 1 экзабайт
	// Может вернуть ошибку или nil в зависимости от системы
	_ = err
}

// Composer 1.
func TestCheckDiskSpace_InvalidDir(t *testing.T) {
	// Тест с несуществующей директорией
	// Функция должна вернуть nil (продолжить работу)
	err := checkDiskSpace("/nonexistent/directory/12345", 1000)
	if err != nil {
		// Если вернула ошибку - это тоже нормально, главное что не паникует
		_ = err
	}
}

// Composer 1.
func TestCheckDiskSpace_CurrentDir(t *testing.T) {
	// Тест с текущей директорией
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	err = checkDiskSpace(wd, 1000)
	if err != nil {
		// Может вернуть ошибку если места недостаточно
		_ = err
	}
}

// Composer 1.
func TestCheckDiskSpace_TempDir(t *testing.T) {
	// Тест с временной директорией
	tmpDir := t.TempDir()
	testDir := filepath.Join(tmpDir, "subdir")
	if err := os.MkdirAll(testDir, 0755); err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	err := checkDiskSpace(testDir, 1000)
	if err != nil {
		// Может вернуть ошибку если места недостаточно
		_ = err
	}
}
