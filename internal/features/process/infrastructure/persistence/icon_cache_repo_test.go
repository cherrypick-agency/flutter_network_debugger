package persistence

import (
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"

	"network-debugger/internal/features/process/domain"
)

func setupIconCacheTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	if err := db.AutoMigrate(&IconCacheModel{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

// Composer 1.
func TestNewIconCacheRepo(t *testing.T) {
	db := setupIconCacheTestDB(t)
	repo := NewIconCacheRepo(db)

	if repo == nil {
		t.Fatal("NewIconCacheRepo returned nil")
	}

	if repo.db == nil {
		t.Fatal("repo.db is nil")
	}
}

// Composer 1.
func TestIconCacheRepo_Set_Get(t *testing.T) {
	db := setupIconCacheTestDB(t)
	repo := NewIconCacheRepo(db)

	key := "test-key"
	icon := &domain.AppIcon{
		Format: "png",
		Data:   []byte("test icon data"),
		Path:   stringPtr("/path/to/icon.png"),
	}

	ttl := 5 * time.Minute

	err := repo.Set(key, icon, ttl)
	if err != nil {
		t.Fatalf("Set() error = %v, want nil", err)
	}

	// Получаем иконку
	retrieved, err := repo.Get(key)
	if err != nil {
		t.Fatalf("Get() error = %v, want nil", err)
	}

	if retrieved == nil {
		t.Fatal("Get() returned nil icon")
	}

	if retrieved.Format != icon.Format {
		t.Errorf("Get() Format = %q, want %q", retrieved.Format, icon.Format)
	}

	if string(retrieved.Data) != string(icon.Data) {
		t.Errorf("Get() Data = %q, want %q", string(retrieved.Data), string(icon.Data))
	}

	if retrieved.Path == nil || icon.Path == nil {
		if retrieved.Path != icon.Path {
			t.Errorf("Get() Path = %v, want %v", retrieved.Path, icon.Path)
		}
	} else if *retrieved.Path != *icon.Path {
		t.Errorf("Get() Path = %q, want %q", *retrieved.Path, *icon.Path)
	}
}

// Composer 1.
func TestIconCacheRepo_Get_NotFound(t *testing.T) {
	db := setupIconCacheTestDB(t)
	repo := NewIconCacheRepo(db)

	_, err := repo.Get("nonexistent-key")
	if err == nil {
		t.Error("Get() with nonexistent key should return error")
	}
}

// Composer 1.
func TestIconCacheRepo_Get_Expired(t *testing.T) {
	db := setupIconCacheTestDB(t)
	repo := NewIconCacheRepo(db)

	key := "expired-key"
	icon := &domain.AppIcon{
		Format: "png",
		Data:   []byte("test data"),
	}

	// Устанавливаем с отрицательным TTL (уже истек)
	err := repo.Set(key, icon, -1*time.Hour)
	if err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Пытаемся получить - должно вернуть ошибку
	_, err = repo.Get(key)
	if err == nil {
		t.Error("Get() with expired key should return error")
	}
}

// Composer 1.
func TestIconCacheRepo_Delete(t *testing.T) {
	db := setupIconCacheTestDB(t)
	repo := NewIconCacheRepo(db)

	key := "delete-key"
	icon := &domain.AppIcon{
		Format: "png",
		Data:   []byte("test data"),
	}

	// Устанавливаем
	if err := repo.Set(key, icon, 5*time.Minute); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Удаляем
	err := repo.Delete(key)
	if err != nil {
		t.Fatalf("Delete() error = %v, want nil", err)
	}

	// Проверяем что удалено
	_, err = repo.Get(key)
	if err == nil {
		t.Error("Get() after Delete() should return error")
	}
}

// Composer 1.
func TestIconCacheRepo_Delete_Nonexistent(t *testing.T) {
	db := setupIconCacheTestDB(t)
	repo := NewIconCacheRepo(db)

	// Удаление несуществующего ключа не должно возвращать ошибку
	err := repo.Delete("nonexistent-key")
	if err != nil {
		t.Errorf("Delete() nonexistent key error = %v, want nil", err)
	}
}

// Composer 1.
func TestIconCacheRepo_Clear(t *testing.T) {
	db := setupIconCacheTestDB(t)
	repo := NewIconCacheRepo(db)

	// Устанавливаем несколько иконок
	icons := []struct {
		key  string
		icon *domain.AppIcon
	}{
		{"key1", &domain.AppIcon{Format: "png", Data: []byte("data1")}},
		{"key2", &domain.AppIcon{Format: "png", Data: []byte("data2")}},
		{"key3", &domain.AppIcon{Format: "png", Data: []byte("data3")}},
	}

	for _, item := range icons {
		if err := repo.Set(item.key, item.icon, 5*time.Minute); err != nil {
			t.Fatalf("Set() error = %v", err)
		}
	}

	// Очищаем
	err := repo.Clear()
	if err != nil {
		t.Fatalf("Clear() error = %v, want nil", err)
	}

	// Проверяем что все удалено
	for _, item := range icons {
		_, err := repo.Get(item.key)
		if err == nil {
			t.Errorf("Get() after Clear() for key %q should return error", item.key)
		}
	}
}

// Composer 1.
func TestIconCacheRepo_CleanupExpired(t *testing.T) {
	db := setupIconCacheTestDB(t)
	repo := NewIconCacheRepo(db)

	// Устанавливаем иконку с истекшим сроком
	expiredKey := "expired-key"
	expiredIcon := &domain.AppIcon{
		Format: "png",
		Data:   []byte("expired data"),
	}
	if err := repo.Set(expiredKey, expiredIcon, -1*time.Hour); err != nil {
		t.Fatalf("Set() expired error = %v", err)
	}

	// Устанавливаем иконку с валидным сроком
	validKey := "valid-key"
	validIcon := &domain.AppIcon{
		Format: "png",
		Data:   []byte("valid data"),
	}
	if err := repo.Set(validKey, validIcon, 1*time.Hour); err != nil {
		t.Fatalf("Set() valid error = %v", err)
	}

	// Очищаем истекшие
	err := repo.CleanupExpired()
	if err != nil {
		t.Fatalf("CleanupExpired() error = %v, want nil", err)
	}

	// Проверяем что истекшая удалена
	_, err = repo.Get(expiredKey)
	if err == nil {
		t.Error("Get() expired key after CleanupExpired() should return error")
	}

	// Проверяем что валидная осталась
	_, err = repo.Get(validKey)
	if err != nil {
		t.Errorf("Get() valid key after CleanupExpired() error = %v, want nil", err)
	}
}

// Composer 1.
func TestIconCacheRepo_Set_Update(t *testing.T) {
	db := setupIconCacheTestDB(t)
	repo := NewIconCacheRepo(db)

	key := "update-key"
	icon1 := &domain.AppIcon{
		Format: "png",
		Data:   []byte("data1"),
	}
	icon2 := &domain.AppIcon{
		Format: "ico",
		Data:   []byte("data2"),
	}

	// Устанавливаем первую иконку
	if err := repo.Set(key, icon1, 5*time.Minute); err != nil {
		t.Fatalf("Set() first error = %v", err)
	}

	// Обновляем на вторую
	if err := repo.Set(key, icon2, 5*time.Minute); err != nil {
		t.Fatalf("Set() second error = %v", err)
	}

	// Проверяем что обновилось
	retrieved, err := repo.Get(key)
	if err != nil {
		t.Fatalf("Get() after update error = %v", err)
	}

	if retrieved.Format != icon2.Format {
		t.Errorf("Get() Format = %q, want %q", retrieved.Format, icon2.Format)
	}

	if string(retrieved.Data) != string(icon2.Data) {
		t.Errorf("Get() Data = %q, want %q", string(retrieved.Data), string(icon2.Data))
	}
}

// Composer 1.
func TestIconCacheRepo_Set_NilPath(t *testing.T) {
	db := setupIconCacheTestDB(t)
	repo := NewIconCacheRepo(db)

	key := "nil-path-key"
	icon := &domain.AppIcon{
		Format: "png",
		Data:   []byte("test data"),
		Path:   nil,
	}

	err := repo.Set(key, icon, 5*time.Minute)
	if err != nil {
		t.Fatalf("Set() with nil path error = %v, want nil", err)
	}

	retrieved, err := repo.Get(key)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if retrieved.Path != nil {
		t.Error("Get() Path should be nil")
	}
}

func stringPtr(s string) *string {
	return &s
}
