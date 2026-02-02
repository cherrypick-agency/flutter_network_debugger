-- +goose Up
-- Persist mapping feature config in runtime_settings (singleton row).
ALTER TABLE runtime_settings ADD COLUMN IF NOT EXISTS mapping_enabled BOOLEAN NOT NULL DEFAULT 1;
ALTER TABLE runtime_settings ADD COLUMN IF NOT EXISTS mapping_upload_max_mb INTEGER NOT NULL DEFAULT 20;

-- +goose Down
-- SQLite: dropping columns is not supported in older versions; keep as no-op.

