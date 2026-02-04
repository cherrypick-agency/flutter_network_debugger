package persistence

import (
	"context"
	"time"

	sdomain "network-debugger/internal/features/settings/domain"

	"gorm.io/gorm"
)

type SettingsRepo struct{ db *gorm.DB }

func NewSettingsRepo(db *gorm.DB) *SettingsRepo { return &SettingsRepo{db: db} }

func (r *SettingsRepo) Load(ctx context.Context) (sdomain.RuntimeSettings, error) {
	var m RuntimeSettingsModel
	tx := r.db.WithContext(ctx).First(&m, 1)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			// create a default record
			now := time.Now().UTC()
			def := RuntimeSettingsModel{ID: 1, UpdatedAt: now}
			if err := r.db.WithContext(ctx).Create(&def).Error; err != nil {
				return sdomain.RuntimeSettings{ID: 1, UpdatedAt: now}, nil
			}
			return toSettingsDomain(&def), nil
		}
		return sdomain.RuntimeSettings{}, tx.Error
	}
	return toSettingsDomain(&m), nil
}

func (r *SettingsRepo) Save(ctx context.Context, s sdomain.RuntimeSettings) error {
	m := toSettingsModel(s)
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now().UTC()
	}
	return r.db.WithContext(ctx).Clauses(clauseOnConflictUpdateAll()).Create(m).Error
}

type ThrottleProfilesRepo struct{ db *gorm.DB }

func NewThrottleProfilesRepo(db *gorm.DB) *ThrottleProfilesRepo { return &ThrottleProfilesRepo{db: db} }

func (r *ThrottleProfilesRepo) List(ctx context.Context) ([]sdomain.ThrottleProfile, error) {
	var rows []ThrottleProfileModel
	if err := r.db.WithContext(ctx).Order("updated_at DESC").Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]sdomain.ThrottleProfile, 0, len(rows))
	for i := range rows {
		out = append(out, toProfileDomain(&rows[i]))
	}
	return out, nil
}

func (r *ThrottleProfilesRepo) Upsert(ctx context.Context, p sdomain.ThrottleProfile) (sdomain.ThrottleProfile, error) {
	if p.ID == "" {
		p.ID = newUUID()
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now().UTC()
	}
	p.UpdatedAt = time.Now().UTC()
	m := toProfileModel(p)
	if err := r.db.WithContext(ctx).Clauses(clauseOnConflictUpdateAll()).Create(m).Error; err != nil {
		return sdomain.ThrottleProfile{}, err
	}
	return toProfileDomain(m), nil
}

func (r *ThrottleProfilesRepo) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&ThrottleProfileModel{ID: id}).Error
}

// --- helpers ---
func newUUID() string {
	// simple dependency already exists
	return uuidNew()
}
