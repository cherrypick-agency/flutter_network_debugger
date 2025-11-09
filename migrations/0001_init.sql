-- +goose Up
CREATE TABLE IF NOT EXISTS runtime_settings (
  id INTEGER PRIMARY KEY,
  response_delay_ms INTEGER NOT NULL DEFAULT 0,
  response_delay_min_ms INTEGER NOT NULL DEFAULT 0,
  response_delay_max_ms INTEGER NOT NULL DEFAULT 0,
  throttle_enabled BOOLEAN NOT NULL DEFAULT 0,
  throttle_down_kbps INTEGER NOT NULL DEFAULT 0,
  throttle_up_kbps INTEGER NOT NULL DEFAULT 0,
  throttle_packet_loss INTEGER NOT NULL DEFAULT 0,
  throttle_offline BOOLEAN NOT NULL DEFAULT 0,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS throttle_profiles (
  id TEXT PRIMARY KEY,
  name TEXT NOT NULL UNIQUE,
  down_kbps INTEGER NOT NULL DEFAULT 0,
  up_kbps INTEGER NOT NULL DEFAULT 0,
  packet_loss_pct INTEGER NOT NULL DEFAULT 0,
  offline BOOLEAN NOT NULL DEFAULT 0,
  created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_throttle_profiles_deleted ON throttle_profiles(deleted_at);

CREATE TABLE IF NOT EXISTS compose_library (
  id INTEGER PRIMARY KEY,
  json TEXT NOT NULL,
  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS compose_history (
  id TEXT PRIMARY KEY,
  template TEXT,
  method TEXT,
  url TEXT,
  status INTEGER,
  "when" TIMESTAMP NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_compose_history_when ON compose_history("when");

-- seed singleton row for runtime_settings if absent
INSERT INTO runtime_settings(id, updated_at)
  SELECT 1, CURRENT_TIMESTAMP
  WHERE NOT EXISTS (SELECT 1 FROM runtime_settings WHERE id=1);

-- +goose Down
DROP TABLE IF EXISTS compose_history;
DROP TABLE IF EXISTS compose_library;
DROP TABLE IF EXISTS throttle_profiles;
DROP TABLE IF EXISTS runtime_settings;











