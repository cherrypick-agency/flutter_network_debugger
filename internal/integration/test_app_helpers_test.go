package integration

import (
	"path/filepath"
	"testing"
	"time"

	interceptp "network-debugger/internal/features/intercept/infrastructure/persistence"
	mappingp "network-debugger/internal/features/mapping/infrastructure/persistence"
	processp "network-debugger/internal/features/process/infrastructure/persistence"
	proxyp "network-debugger/internal/features/proxy/infrastructure/persistence"
	settingsp "network-debugger/internal/features/settings/infrastructure/persistence"
	dbpkg "network-debugger/internal/infrastructure/db"
	httpapi "network-debugger/internal/infrastructure/httpapi"
)

func attachTestSQLite(t *testing.T, deps *httpapi.Deps) {
	t.Helper()

	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "test.db")
	gdb, err := dbpkg.NewSQLite(dbPath)
	if err != nil {
		return
	}
	_ = gdb.AutoMigrate(
		&proxyp.ProxyConfigModel{},
		&settingsp.RuntimeSettingsModel{},
		&settingsp.ThrottleProfileModel{},
		&mappingp.MapRuleModel{},
		&processp.ProcessDetectionConfigModel{},
		&processp.IconCacheModel{},
		&interceptp.InterceptConfigModel{},
		&interceptp.InterceptRuleModel{},
	)
	_ = gdb.Create(&proxyp.ProxyConfigModel{
		ID:             1,
		ForwardEnabled: false,
		ForwardAddr:    "127.0.0.1:0",
		SocksEnabled:   false,
		SocksAddr:      "127.0.0.1:0",
		SocksAuthMode:  "none",
		UpdatedAt:      time.Now().UTC(),
	}).Error
	deps.DB = gdb
}
