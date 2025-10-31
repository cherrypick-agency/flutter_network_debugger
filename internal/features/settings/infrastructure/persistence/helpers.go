package persistence

import (
	"gorm.io/gorm/clause"
	"network-debugger/pkg/shared/id"
)

func clauseOnConflictUpdateAll() clause.OnConflict {
	return clause.OnConflict{UpdateAll: true}
}

func uuidNew() string { return id.New() }
