package persistence

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
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
