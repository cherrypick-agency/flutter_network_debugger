package persistence

import (
	"time"

	pdomain "network-debugger/internal/features/proxy/domain"
)

// ProxyConfigModel - singleton with ID=1
type ProxyConfigModel struct {
	ID int64 `gorm:"primaryKey"`

	ForwardEnabled bool
	ForwardAddr    string `gorm:"type:text"`

	SocksEnabled bool
	SocksAddr    string `gorm:"type:text"`

	SocksAuthMode string `gorm:"type:text"`
	SocksUser     string `gorm:"type:text"`
	SocksPass     string `gorm:"type:text"`

	UpdatedAt time.Time
}

func (ProxyConfigModel) TableName() string { return "proxy_config" }

func toDomain(m *ProxyConfigModel) pdomain.ProxyConfig {
	if m == nil {
		return pdomain.ProxyConfig{ID: 1}
	}
	return pdomain.ProxyConfig{
		ID:             m.ID,
		ForwardEnabled: m.ForwardEnabled,
		ForwardAddr:    m.ForwardAddr,
		SocksEnabled:   m.SocksEnabled,
		SocksAddr:      m.SocksAddr,
		SocksAuthMode:  m.SocksAuthMode,
		SocksUser:      m.SocksUser,
		SocksPass:      m.SocksPass,
		UpdatedAt:      m.UpdatedAt,
	}
}

func toModel(c pdomain.ProxyConfig) *ProxyConfigModel {
	return &ProxyConfigModel{
		ID:             1,
		ForwardEnabled: c.ForwardEnabled,
		ForwardAddr:    c.ForwardAddr,
		SocksEnabled:   c.SocksEnabled,
		SocksAddr:      c.SocksAddr,
		SocksAuthMode:  c.SocksAuthMode,
		SocksUser:      c.SocksUser,
		SocksPass:      c.SocksPass,
		UpdatedAt:      c.UpdatedAt,
	}
}
