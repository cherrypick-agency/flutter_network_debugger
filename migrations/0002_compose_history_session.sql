-- +goose Up
ALTER TABLE compose_history ADD COLUMN session TEXT;

-- +goose Down
-- SQLite не поддерживает DROP COLUMN напрямую; оставим столбец.



