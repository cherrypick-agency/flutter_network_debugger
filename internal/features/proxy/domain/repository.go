package domain

import "context"

// ProxyConfigRepository - storage for single proxy configuration record.
type ProxyConfigRepository interface {
	Load(ctx context.Context) (ProxyConfig, error)
	Save(ctx context.Context, c ProxyConfig) error
}
