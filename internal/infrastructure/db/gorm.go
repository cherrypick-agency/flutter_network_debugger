package db

import (
	"log"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

// PathFromEnv returns the path to the SQLite database.
// If DB_PATH is not set, use data/network_debugger.db
func PathFromEnv() string {
	if p := os.Getenv("DB_PATH"); p != "" {
		return p
	}
	return filepath.Join("data", "network_debugger.db")
}

// NewSQLite opens a SQLite database at the specified path and returns *gorm.DB.
// Creates the directory if necessary.
func NewSQLite(path string) (*gorm.DB, error) {
	if dir := filepath.Dir(path); dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
	// GORM logger: suppress 'record not found' noise, keep WARN level
	lg := glogger.New(log.New(os.Stdout, "", log.LstdFlags), glogger.Config{
		IgnoreRecordNotFoundError: true,
		LogLevel:                  glogger.Warn,
	})
	gdb, err := gorm.Open(sqlite.Open(path), &gorm.Config{
		Logger:      lg,
		PrepareStmt: true,
	})
	if err != nil {
		return nil, err
	}

	// SQLite optimization for concurrent access
	gdb.Exec(`PRAGMA journal_mode = WAL`)
	gdb.Exec(`PRAGMA busy_timeout = 5000`)
	gdb.Exec(`PRAGMA synchronous = NORMAL`)

	return gdb, nil
}
