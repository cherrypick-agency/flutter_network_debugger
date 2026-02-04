package persistence

import (
	"testing"
)

// Composer 1.
func TestClauseOnConflictUpdateAll(t *testing.T) {
	clause := clauseOnConflictUpdateAll()

	if !clause.UpdateAll {
		t.Error("clauseOnConflictUpdateAll() UpdateAll = false, want true")
	}
}

// Composer 1.
func TestUuidNew(t *testing.T) {
	uuid1 := uuidNew()
	uuid2 := uuidNew()

	if uuid1 == "" {
		t.Error("uuidNew() returned empty string")
	}

	if uuid2 == "" {
		t.Error("uuidNew() returned empty string")
	}

	if uuid1 == uuid2 {
		t.Error("uuidNew() returned same UUID twice")
	}

	// Verify that this is a valid UUID via id.New()
	// id.New() should return a non-empty string
	if len(uuid1) == 0 {
		t.Error("uuidNew() returned empty UUID")
	}
}

// Composer 1.
func TestUuidNew_Uniqueness(t *testing.T) {
	// Generate many UUIDs and verify uniqueness
	uuids := make(map[string]bool)
	for i := 0; i < 100; i++ {
		uuid := uuidNew()
		if uuids[uuid] {
			t.Errorf("uuidNew() returned duplicate UUID: %s", uuid)
		}
		uuids[uuid] = true
	}
}
