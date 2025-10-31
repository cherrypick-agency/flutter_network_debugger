package persistence

import (
	"time"
)

// ComposeLibraryModel — хранит снимок библиотеки в JSON (одна строка с id=1).
type ComposeLibraryModel struct {
	ID        int64  `gorm:"primaryKey"`
	JSON      string `gorm:"type:text"`
	UpdatedAt time.Time
}

func (ComposeLibraryModel) TableName() string { return "compose_library" }

// ComposeHistoryEntryModel — одна запись истории отправок.
type ComposeHistoryEntryModel struct {
	ID       string `gorm:"type:text;primaryKey"`
	Template string `gorm:"type:text"`
	Method   string `gorm:"type:text"`
	URL      string `gorm:"type:text"`
	Status   int
	When     time.Time `gorm:"column:when_ts;index"`
}

func (ComposeHistoryEntryModel) TableName() string { return "compose_history" }
