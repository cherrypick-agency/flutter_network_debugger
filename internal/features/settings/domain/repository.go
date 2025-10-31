package domain

import "context"

type SettingsRepository interface {
	Load(ctx context.Context) (RuntimeSettings, error)
	Save(ctx context.Context, s RuntimeSettings) error
}

type ThrottleProfilesRepository interface {
	List(ctx context.Context) ([]ThrottleProfile, error)
	Upsert(ctx context.Context, p ThrottleProfile) (ThrottleProfile, error)
	Delete(ctx context.Context, id string) error
}
