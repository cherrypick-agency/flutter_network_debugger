-- +goose Up
-- +goose StatementBegin

-- Конфигурация детекции процессов (singleton с ID=1)
CREATE TABLE IF NOT EXISTS process_detection_config (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    enabled BOOLEAN NOT NULL DEFAULT 1,
    use_helper_tool BOOLEAN NOT NULL DEFAULT 0,
    helper_installed BOOLEAN NOT NULL DEFAULT 0,
    cache_enabled BOOLEAN NOT NULL DEFAULT 1,
    cache_ttl_seconds INTEGER NOT NULL DEFAULT 300,
    fallback_enabled BOOLEAN NOT NULL DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Кеш иконок приложений
CREATE TABLE IF NOT EXISTS icon_cache (
    cache_key TEXT PRIMARY KEY,
    icon_format TEXT NOT NULL,
    icon_data BLOB NOT NULL,
    icon_path TEXT,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_icon_cache_expires ON icon_cache(expires_at);

-- Вставить дефолтную конфигурацию (enabled по умолчанию)
INSERT INTO process_detection_config (id, enabled) VALUES (1, 1)
    ON CONFLICT(id) DO NOTHING;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP INDEX IF EXISTS idx_icon_cache_expires;
DROP TABLE IF EXISTS icon_cache;
DROP TABLE IF EXISTS process_detection_config;

-- +goose StatementEnd
