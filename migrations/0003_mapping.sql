-- map_rules table for Map Local/Remote feature
CREATE TABLE IF NOT EXISTS map_rules (
  id TEXT PRIMARY KEY,
  enabled BOOLEAN NOT NULL DEFAULT 1,
  priority INTEGER NOT NULL DEFAULT 100,
  kind TEXT NOT NULL,
  stop_processing BOOLEAN NOT NULL DEFAULT 1,

  methods_json TEXT,
  host_pattern TEXT,
  path_pattern TEXT,
  pattern_type TEXT,

  file_path TEXT,
  blob_path TEXT,
  status_override INTEGER NOT NULL DEFAULT 200,
  content_type_override TEXT,

  target_url_template TEXT,
  preserve_host BOOLEAN NOT NULL DEFAULT 0,

  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_map_rules_enabled ON map_rules(enabled);
CREATE INDEX IF NOT EXISTS idx_map_rules_priority ON map_rules(priority);



