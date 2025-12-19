package persistence

import (
	"context"
	"testing"
	"time"

	sdomain "network-debugger/internal/features/settings/domain"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	if err := db.AutoMigrate(&RuntimeSettingsModel{}, &ThrottleProfileModel{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

func TestSettingsRepo_SaveLoad(t *testing.T) {
	db := newTestDB(t)
	r := NewSettingsRepo(db)
	ctx := context.Background()
	// load creates default
	got, err := r.Load(ctx)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.ID != 1 {
		t.Fatalf("want id=1 got %d", got.ID)
	}
	// save
	want := sdomain.RuntimeSettings{ID: 1, ResponseDelayMs: 777, ThrottleEnabled: true, ThrottleDownKbps: 1000}
	if err := r.Save(ctx, want); err != nil {
		t.Fatalf("save: %v", err)
	}
	got, err = r.Load(ctx)
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got.ResponseDelayMs != 777 || !got.UpdatedAt.After(time.Time{}) {
		t.Fatalf("unexpected %+v", got)
	}
}

func TestThrottleProfilesRepo_CRUD(t *testing.T) {
	db := newTestDB(t)
	pr := NewThrottleProfilesRepo(db)
	ctx := context.Background()
	// upsert create
	p, err := pr.Upsert(ctx, sdomain.ThrottleProfile{Name: "3G", DownKbps: 400, UpKbps: 400})
	if err != nil {
		t.Fatalf("upsert1: %v", err)
	}
	if p.ID == "" {
		t.Fatalf("empty id")
	}
	// list
	lst, err := pr.List(ctx)
	if err != nil || len(lst) != 1 {
		t.Fatalf("list: %v len=%d", err, len(lst))
	}
	// update
	p.DownKbps = 450
	if _, err := pr.Upsert(ctx, p); err != nil {
		t.Fatalf("upsert2: %v", err)
	}
	lst, _ = pr.List(ctx)
	if lst[0].DownKbps != 450 {
		t.Fatalf("update failed: %+v", lst[0])
	}
	// delete
	if err := pr.Delete(ctx, p.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	lst, _ = pr.List(ctx)
	if len(lst) != 0 {
		t.Fatalf("expected empty, got %d", len(lst))
	}
}
