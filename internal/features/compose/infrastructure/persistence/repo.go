package persistence

import (
	"context"
	"encoding/json"
	"time"

	"network-debugger/internal/domain"
	"network-debugger/internal/usecase"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type LibraryRepo struct{ db *gorm.DB }

func NewLibraryRepo(db *gorm.DB) *LibraryRepo { return &LibraryRepo{db: db} }

func (r *LibraryRepo) LoadLibrary(ctx context.Context) (domain.ComposeLibrary, error) {
	var m ComposeLibraryModel
	tx := r.db.WithContext(ctx).First(&m, 1)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return domain.ComposeLibrary{Collections: []domain.ComposeCollection{}, RequestsByID: map[string]domain.ComposeRequestTemplate{}}, nil
		}
		return domain.ComposeLibrary{}, tx.Error
	}
	var lib domain.ComposeLibrary
	if err := json.Unmarshal([]byte(m.JSON), &lib); err != nil {
		// вернём пустую библиотеку при ошибке формата
		return domain.ComposeLibrary{Collections: []domain.ComposeCollection{}, RequestsByID: map[string]domain.ComposeRequestTemplate{}}, nil
	}
	if lib.RequestsByID == nil {
		lib.RequestsByID = map[string]domain.ComposeRequestTemplate{}
	}
	return lib, nil
}

func (r *LibraryRepo) SaveLibrary(ctx context.Context, lib domain.ComposeLibrary) error {
	if lib.RequestsByID == nil {
		lib.RequestsByID = map[string]domain.ComposeRequestTemplate{}
	}
	b, _ := json.Marshal(lib)
	m := ComposeLibraryModel{ID: 1, JSON: string(b), UpdatedAt: time.Now().UTC()}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(&m).Error
}

type HistoryRepo struct{ db *gorm.DB }

func NewHistoryRepo(db *gorm.DB) *HistoryRepo { return &HistoryRepo{db: db} }

func (r *HistoryRepo) AppendHistory(ctx context.Context, e domain.ComposeHistoryEntry) error {
	if e.ID == "" {
		e.ID = time.Now().UTC().Format("20060102150405.000000000")
	}
	if e.When.IsZero() {
		e.When = time.Now().UTC()
	}
	m := ComposeHistoryEntryModel{ID: e.ID, Template: e.Template, Method: e.Method, URL: e.URL, Status: e.Status, When: e.When}
	return r.db.WithContext(ctx).Create(&m).Error
}

func (r *HistoryRepo) ListHistory(ctx context.Context, from string, limit int) ([]domain.ComposeHistoryEntry, string, error) {
	if limit <= 0 {
		limit = 50
	}
	var rows []ComposeHistoryEntryModel
	if err := r.db.WithContext(ctx).Order("when_ts DESC").Limit(limit).Find(&rows).Error; err != nil {
		return nil, "", err
	}
	out := make([]domain.ComposeHistoryEntry, 0, len(rows))
	for i := range rows {
		m := rows[i]
		out = append(out, domain.ComposeHistoryEntry{ID: m.ID, Template: m.Template, Method: m.Method, URL: m.URL, Status: m.Status, When: m.When})
	}
	return out, "", nil
}

var _ usecase.ComposeLibraryRepository = (*LibraryRepo)(nil)
var _ usecase.ComposeHistoryRepository = (*HistoryRepo)(nil)
