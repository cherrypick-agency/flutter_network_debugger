package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	if err := db.AutoMigrate(&ProxyConfigModel{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

func TestRepo_SaveLoad(t *testing.T) {
	db := newTestDB(t)
	r := NewRepo(db)
	ctx := context.Background()
	pc, err := r.Load(ctx)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if pc.ID != 1 {
		t.Fatalf("id=%d", pc.ID)
	}
	pc.ForwardEnabled = true
	pc.ForwardAddr = ":9999"
	pc.SocksEnabled = true
	pc.SocksAddr = ":9998"
	if err := r.Save(ctx, pc); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err := r.Load(ctx)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got.ForwardAddr != ":9999" || !got.UpdatedAt.After(time.Time{}) {
		t.Fatalf("unexpected %+v", got)
	}
}

// TestRepo_Load_DefaultConfig tests Load with default configuration
func TestRepo_Load_DefaultConfig(t *testing.T) {
	db := newTestDB(t)
	r := NewRepo(db)
	ctx := context.Background()

	// Удаляем возможную существующую запись, чтобы проверить создание по умолчанию
	db.WithContext(ctx).Exec("DELETE FROM proxy_config WHERE id = 1")

	pc, err := r.Load(ctx)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if pc.ID != 1 {
		t.Errorf("Load() pc.ID = %d, want 1", pc.ID)
	}

	if !pc.ForwardEnabled {
		t.Error("Load() pc.ForwardEnabled = false, want true")
	}

	if pc.ForwardAddr != ":9091" {
		t.Errorf("Load() pc.ForwardAddr = %q, want \":9091\"", pc.ForwardAddr)
	}

	if pc.SocksEnabled {
		t.Error("Load() pc.SocksEnabled = true, want false")
	}

	if pc.UpdatedAt.IsZero() {
		t.Error("Load() pc.UpdatedAt should not be zero")
	}
}

// TestRepo_Load_MultipleTimes tests Load called multiple times
func TestRepo_Load_MultipleTimes(t *testing.T) {
	db := newTestDB(t)
	r := NewRepo(db)
	ctx := context.Background()

	pc1, err := r.Load(ctx)
	if err != nil {
		t.Fatalf("First Load() error = %v, want nil", err)
	}

	pc2, err := r.Load(ctx)
	if err != nil {
		t.Fatalf("Second Load() error = %v, want nil", err)
	}

	if pc1.ID != pc2.ID {
		t.Errorf("Load() pc1.ID = %d, pc2.ID = %d, want same", pc1.ID, pc2.ID)
	}

	if pc1.ForwardAddr != pc2.ForwardAddr {
		t.Errorf("Load() pc1.ForwardAddr = %q, pc2.ForwardAddr = %q, want same", pc1.ForwardAddr, pc2.ForwardAddr)
	}
}

// TestRepo_Load_WithContextCancellation tests Load with cancelled context
func TestRepo_Load_WithContextCancellation(t *testing.T) {
	db := newTestDB(t)
	r := NewRepo(db)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := r.Load(ctx)
	if err == nil {
		t.Error("Load() with cancelled context should return error")
	}
}

// TestRepo_Load_AfterSave tests Load after Save
func TestRepo_Load_AfterSave(t *testing.T) {
	db := newTestDB(t)
	r := NewRepo(db)
	ctx := context.Background()

	// Загружаем конфигурацию
	pc, err := r.Load(ctx)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	// Изменяем конфигурацию
	pc.ForwardAddr = ":8888"
	pc.SocksEnabled = true
	pc.SocksAddr = ":7777"

	// Сохраняем
	err = r.Save(ctx, pc)
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}

	// Загружаем снова
	pc2, err := r.Load(ctx)
	if err != nil {
		t.Fatalf("Second Load() error = %v, want nil", err)
	}

	if pc2.ForwardAddr != ":8888" {
		t.Errorf("Load() after Save() pc2.ForwardAddr = %q, want \":8888\"", pc2.ForwardAddr)
	}

	if !pc2.SocksEnabled {
		t.Error("Load() after Save() pc2.SocksEnabled = false, want true")
	}

	if pc2.SocksAddr != ":7777" {
		t.Errorf("Load() after Save() pc2.SocksAddr = %q, want \":7777\"", pc2.SocksAddr)
	}
}

// TestRepo_Save tests Save method
func TestRepo_Save(t *testing.T) {
	db := newTestDB(t)
	r := NewRepo(db)
	ctx := context.Background()

	pc, err := r.Load(ctx)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	pc.ForwardAddr = ":7777"
	pc.SocksEnabled = true

	err = r.Save(ctx, pc)
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}

	pc2, err := r.Load(ctx)
	if err != nil {
		t.Fatalf("Second Load() error = %v, want nil", err)
	}

	if pc2.ForwardAddr != ":7777" {
		t.Errorf("Save() pc2.ForwardAddr = %q, want \":7777\"", pc2.ForwardAddr)
	}

	if !pc2.SocksEnabled {
		t.Error("Save() pc2.SocksEnabled = false, want true")
	}
}

// TestRepo_Save_UpdatesTimestamp tests Save updates UpdatedAt timestamp
func TestRepo_Save_UpdatesTimestamp(t *testing.T) {
	db := newTestDB(t)
	r := NewRepo(db)
	ctx := context.Background()

	pc, err := r.Load(ctx)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	originalTime := pc.UpdatedAt

	time.Sleep(10 * time.Millisecond)

	pc.ForwardAddr = ":8888"
	err = r.Save(ctx, pc)
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}

	pc2, err := r.Load(ctx)
	if err != nil {
		t.Fatalf("Second Load() error = %v, want nil", err)
	}

	if !pc2.UpdatedAt.After(originalTime) {
		t.Error("Save() should update UpdatedAt timestamp")
	}
}

// TestRepo_Save_WithZeroTimestamp tests Save with zero timestamp
func TestRepo_Save_WithZeroTimestamp(t *testing.T) {
	db := newTestDB(t)
	r := NewRepo(db)
	ctx := context.Background()

	pc, err := r.Load(ctx)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	pc.UpdatedAt = time.Time{}

	err = r.Save(ctx, pc)
	if err != nil {
		t.Fatalf("Save() error = %v, want nil", err)
	}

	pc2, err := r.Load(ctx)
	if err != nil {
		t.Fatalf("Second Load() error = %v, want nil", err)
	}

	if pc2.UpdatedAt.IsZero() {
		t.Error("Save() should set UpdatedAt if zero")
	}
}

// TestRepo_Save_WithContextCancellation tests Save with cancelled context
func TestRepo_Save_WithContextCancellation(t *testing.T) {
	db := newTestDB(t)
	r := NewRepo(db)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	pc, err := r.Load(context.Background())
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	err = r.Save(ctx, pc)
	if err == nil {
		t.Error("Save() with cancelled context should return error")
	}
}

// TestRepo_Save_MultipleTimes tests Save called multiple times
func TestRepo_Save_MultipleTimes(t *testing.T) {
	db := newTestDB(t)
	r := NewRepo(db)
	ctx := context.Background()

	pc, err := r.Load(ctx)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	pc.ForwardAddr = ":1111"
	err = r.Save(ctx, pc)
	if err != nil {
		t.Fatalf("First Save() error = %v, want nil", err)
	}

	pc.ForwardAddr = ":2222"
	err = r.Save(ctx, pc)
	if err != nil {
		t.Fatalf("Second Save() error = %v, want nil", err)
	}

	pc2, err := r.Load(ctx)
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if pc2.ForwardAddr != ":2222" {
		t.Errorf("Save() pc2.ForwardAddr = %q, want \":2222\"", pc2.ForwardAddr)
	}
}
