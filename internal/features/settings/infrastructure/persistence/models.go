package persistence

import (
	"time"

	sdomain "network-debugger/internal/features/settings/domain"

	"gorm.io/gorm"
)

// RuntimeSettingsModel — singleton with ID=1
type RuntimeSettingsModel struct {
	ID                 int64 `gorm:"primaryKey"`
	ResponseDelayMs    int
	ResponseDelayMinMs int
	ResponseDelayMaxMs int

	ThrottleEnabled       bool
	ThrottleDownKbps      int
	ThrottleUpKbps        int
	ThrottlePacketLoss    int
	ThrottleLatencyMs     int
	ThrottleLatencyJitter int
	ThrottleOffline       bool

	FontScale      float64
	HighlightTheme string `gorm:"type:text"`

	MappingEnabled     bool
	MappingUploadMaxMB int

	UpdatedAt time.Time
}

func (RuntimeSettingsModel) TableName() string { return "runtime_settings" }

type ThrottleProfileModel struct {
	ID            string `gorm:"type:text;primaryKey"`
	Name          string `gorm:"type:text;uniqueIndex;not null"`
	DownKbps      int
	UpKbps        int
	PacketLossPct int
	LatencyMs     int
	LatencyJitter int
	Offline       bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `gorm:"index"`
}

func (ThrottleProfileModel) TableName() string { return "throttle_profiles" }

func toSettingsModel(s sdomain.RuntimeSettings) *RuntimeSettingsModel {
	return &RuntimeSettingsModel{
		ID:                    1,
		ResponseDelayMs:       s.ResponseDelayMs,
		ResponseDelayMinMs:    s.ResponseDelayMinMs,
		ResponseDelayMaxMs:    s.ResponseDelayMaxMs,
		ThrottleEnabled:       s.ThrottleEnabled,
		ThrottleDownKbps:      s.ThrottleDownKbps,
		ThrottleUpKbps:        s.ThrottleUpKbps,
		ThrottlePacketLoss:    s.ThrottlePacketLoss,
		ThrottleLatencyMs:     s.ThrottleLatencyMs,
		ThrottleLatencyJitter: s.ThrottleLatencyJitter,
		ThrottleOffline:       s.ThrottleOffline,
		FontScale:             s.FontScale,
		HighlightTheme:        s.HighlightTheme,
		MappingEnabled:        s.MappingEnabled,
		MappingUploadMaxMB:    s.MappingUploadMaxMB,
		UpdatedAt:             s.UpdatedAt,
	}
}

func toSettingsDomain(m *RuntimeSettingsModel) sdomain.RuntimeSettings {
	if m == nil {
		return sdomain.RuntimeSettings{ID: 1}
	}
	return sdomain.RuntimeSettings{
		ID:                    m.ID,
		ResponseDelayMs:       m.ResponseDelayMs,
		ResponseDelayMinMs:    m.ResponseDelayMinMs,
		ResponseDelayMaxMs:    m.ResponseDelayMaxMs,
		ThrottleEnabled:       m.ThrottleEnabled,
		ThrottleDownKbps:      m.ThrottleDownKbps,
		ThrottleUpKbps:        m.ThrottleUpKbps,
		ThrottlePacketLoss:    m.ThrottlePacketLoss,
		ThrottleLatencyMs:     m.ThrottleLatencyMs,
		ThrottleLatencyJitter: m.ThrottleLatencyJitter,
		ThrottleOffline:       m.ThrottleOffline,
		FontScale:             m.FontScale,
		HighlightTheme:        m.HighlightTheme,
		MappingEnabled:        m.MappingEnabled,
		MappingUploadMaxMB:    m.MappingUploadMaxMB,
		UpdatedAt:             m.UpdatedAt,
	}
}

func toProfileModel(p sdomain.ThrottleProfile) *ThrottleProfileModel {
	return &ThrottleProfileModel{
		ID:            p.ID,
		Name:          p.Name,
		DownKbps:      p.DownKbps,
		UpKbps:        p.UpKbps,
		PacketLossPct: p.PacketLossPct,
		LatencyMs:     p.LatencyMs,
		LatencyJitter: p.LatencyJitter,
		Offline:       p.Offline,
		CreatedAt:     p.CreatedAt,
		UpdatedAt:     p.UpdatedAt,
	}
}

func toProfileDomain(m *ThrottleProfileModel) sdomain.ThrottleProfile {
	return sdomain.ThrottleProfile{
		ID:            m.ID,
		Name:          m.Name,
		DownKbps:      m.DownKbps,
		UpKbps:        m.UpKbps,
		PacketLossPct: m.PacketLossPct,
		LatencyMs:     m.LatencyMs,
		LatencyJitter: m.LatencyJitter,
		Offline:       m.Offline,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}
