-- +goose Up
ALTER TABLE runtime_settings ADD COLUMN IF NOT EXISTS font_scale REAL DEFAULT 1.0;

-- +goose Down
-- SQLite не поддерживает удаление столбцов без пересоздания таблицы; оставим как есть



