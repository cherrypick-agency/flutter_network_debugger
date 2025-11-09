package persistence

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	pdomain "network-debugger/internal/features/proxy/domain"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repo struct{ db *gorm.DB }

func NewRepo(db *gorm.DB) *Repo { return &Repo{db: db} }

func (r *Repo) Load(ctx context.Context) (pdomain.ProxyConfig, error) {
	var m ProxyConfigModel
	tx := r.db.WithContext(ctx).First(&m, 1)
	if tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			now := time.Now().UTC()
			// По умолчанию оставляем forward‑proxy на 9091, UI живёт на отдельном порту (Addr).
			// Можно переопределить через env variable FORWARD_PROXY_DEFAULT_PORT
			forwardAddr := ":9091"
			if port := os.Getenv("FORWARD_PROXY_DEFAULT_PORT"); port != "" {
				// Валидация порта
				if p, err := strconv.Atoi(port); err == nil && p > 0 && p <= 65535 {
					forwardAddr = fmt.Sprintf(":%d", p)
				}
				// Если порт невалидный, используем дефолтный 9091
			}
			def := ProxyConfigModel{ID: 1, ForwardEnabled: true, ForwardAddr: forwardAddr, SocksEnabled: false, SocksAddr: ":8889", SocksAuthMode: "none", UpdatedAt: now}
			if err := r.db.WithContext(ctx).Create(&def).Error; err != nil {
				return toDomain(&def), nil
			}
			return toDomain(&def), nil
		}
		return pdomain.ProxyConfig{}, tx.Error
	}
	return toDomain(&m), nil
}

func (r *Repo) Save(ctx context.Context, c pdomain.ProxyConfig) error {
	m := toModel(c)
	if m.UpdatedAt.IsZero() {
		m.UpdatedAt = time.Now().UTC()
	}
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(m).Error
}

var _ pdomain.ProxyConfigRepository = (*Repo)(nil)
