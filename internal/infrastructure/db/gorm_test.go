package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPathFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		envValue string
		want     string
	}{
		{
			name:     "env var set",
			envValue: "/custom/path/db.sqlite",
			want:     "/custom/path/db.sqlite",
		},
		{
			name:     "env var not set",
			envValue: "",
			want:     filepath.Join("data", "network_debugger.db"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save and restore original env
			original := os.Getenv("DB_PATH")
			defer func() {
				if original != "" {
					os.Setenv("DB_PATH", original)
				} else {
					os.Unsetenv("DB_PATH")
				}
			}()

			// Set test env value
			if tt.envValue != "" {
				os.Setenv("DB_PATH", tt.envValue)
			} else {
				os.Unsetenv("DB_PATH")
			}

			got := PathFromEnv()
			if got != tt.want {
				t.Errorf("PathFromEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSQLite(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{
			name:    "create db in temp directory",
			path:    filepath.Join(t.TempDir(), "test.db"),
			wantErr: false,
		},
		{
			name:    "create db with nested directory",
			path:    filepath.Join(t.TempDir(), "nested", "dir", "test.db"),
			wantErr: false,
		},
		{
			name:    "create db in current directory",
			path:    filepath.Join(t.TempDir(), "current.db"),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db, err := NewSQLite(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewSQLite() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				if db == nil {
					t.Error("NewSQLite() returned nil db without error")
				}
				// Verify directory was created
				dir := filepath.Dir(tt.path)
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					t.Errorf("NewSQLite() did not create directory %s", dir)
				}
				// Close db
				sqlDB, _ := db.DB()
				if sqlDB != nil {
					sqlDB.Close()
				}
			}
		})
	}
}

func TestNewSQLite_Integration(t *testing.T) {
	// Integration test to verify db actually works
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "integration_test.db")

	db, err := NewSQLite(dbPath)
	if err != nil {
		t.Fatalf("NewSQLite() failed: %v", err)
	}
	defer func() {
		sqlDB, _ := db.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}()

	// Test basic operations
	type TestModel struct {
		ID   uint   `gorm:"primarykey"`
		Name string `gorm:"size:100"`
	}

	// Auto migrate
	if err := db.AutoMigrate(&TestModel{}); err != nil {
		t.Fatalf("AutoMigrate() failed: %v", err)
	}

	// Create record
	record := TestModel{Name: "test"}
	if err := db.Create(&record).Error; err != nil {
		t.Fatalf("Create() failed: %v", err)
	}

	// Read record
	var found TestModel
	if err := db.First(&found, record.ID).Error; err != nil {
		t.Fatalf("First() failed: %v", err)
	}

	if found.Name != "test" {
		t.Errorf("Expected name 'test', got '%s'", found.Name)
	}
}
