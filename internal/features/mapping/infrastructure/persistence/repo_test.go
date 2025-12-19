package persistence

import (
	"context"
	"testing"

	mdomain "network-debugger/internal/features/mapping/domain"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("db open: %v", err)
	}
	if err := db.AutoMigrate(&MapRuleModel{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

func TestRepo_UpsertListReorderDelete(t *testing.T) {
	db := newTestDB(t)
	r := NewRepo(db)
	ctx := context.Background()

	// upsert two rules
	a, err := r.Upsert(ctx, mdomain.MapRule{Enabled: true, Priority: 10, Kind: mdomain.KindRemote, HostPattern: "example.com", PathPattern: "/v1/*", PatternType: mdomain.PatternGlob, TargetURLTemplate: "https://staging.example.com$0"})
	if err != nil {
		t.Fatalf("upsert a: %v", err)
	}
	b, err := r.Upsert(ctx, mdomain.MapRule{Enabled: true, Priority: 20, Kind: mdomain.KindLocal, HostPattern: "cdn.example.com", PathPattern: "/bundle.js", PatternType: mdomain.PatternGlob, StatusOverride: 200, ContentTypeOverride: "application/javascript"})
	if err != nil {
		t.Fatalf("upsert b: %v", err)
	}

	list, err := r.List(ctx)
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("want 2, got %d", len(list))
	}
	if list[0].ID != a.ID {
		t.Fatalf("order unexpected")
	}

	// reorder
	if err := r.Reorder(ctx, []string{b.ID, a.ID}); err != nil {
		t.Fatalf("reorder: %v", err)
	}
	list2, _ := r.List(ctx)
	if list2[0].ID != b.ID {
		t.Fatalf("reorder failed")
	}

	// delete
	if err := r.Delete(ctx, a.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	list3, _ := r.List(ctx)
	if len(list3) != 1 || list3[0].ID != b.ID {
		t.Fatalf("delete failed")
	}
}
