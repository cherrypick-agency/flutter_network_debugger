package domain

import "context"

// ProxyConfigRepository — хранилище единственной записи конфигурации прокси.
type ProxyConfigRepository interface {
	Load(ctx context.Context) (ProxyConfig, error)
	Save(ctx context.Context, c ProxyConfig) error
}
