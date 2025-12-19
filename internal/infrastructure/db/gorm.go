package db

import (
	"log"
	"os"
	"path/filepath"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
)

// PathFromEnv возвращает путь к SQLite базе.
// Если DB_PATH не задан, используем data/network_debugger.db
func PathFromEnv() string {
	if p := os.Getenv("DB_PATH"); p != "" {
		return p
	}
	return filepath.Join("data", "network_debugger.db")
}

// NewSQLite открывает SQLite базу по указанному пути и возвращает *gorm.DB.
// Создаёт директорию при необходимости.
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

	// Оптимизация SQLite для конкурентного доступа
	gdb.Exec(`PRAGMA journal_mode = WAL`)
	gdb.Exec(`PRAGMA busy_timeout = 5000`)
	gdb.Exec(`PRAGMA synchronous = NORMAL`)

	return gdb, nil
}
