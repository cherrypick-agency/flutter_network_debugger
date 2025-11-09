-- Scripting API feature
-- Multi-language script execution system with support for Extism (WASM) and Dart runtimes

CREATE TABLE IF NOT EXISTS scripts (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,

    -- Runtime configuration
    runtime TEXT NOT NULL CHECK(runtime IN ('extism', 'dart')),
    code BLOB NOT NULL,
    language TEXT NOT NULL, -- 'rust', 'javascript', 'go', 'dart', 'python', etc.

    -- Trigger configuration
    trigger_type TEXT NOT NULL CHECK(trigger_type IN ('request', 'response', 'both')),
    priority INTEGER DEFAULT 0,

    -- Match rules (JSON)
    match_methods TEXT, -- JSON array: ["GET", "POST"]
    match_host_pattern TEXT,
    match_path_pattern TEXT,
    pattern_type TEXT CHECK(pattern_type IN ('exact', 'prefix', 'regex')),

    -- Resource limits
    timeout_ms INTEGER DEFAULT 5000,
    memory_limit_mb INTEGER DEFAULT 10,
    allowed_hosts TEXT, -- JSON array of allowed hosts for http_fetch

    -- State
    enabled BOOLEAN DEFAULT TRUE,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_scripts_enabled ON scripts(enabled);
CREATE INDEX IF NOT EXISTS idx_scripts_runtime ON scripts(runtime);
CREATE INDEX IF NOT EXISTS idx_scripts_priority ON scripts(priority DESC);
CREATE INDEX IF NOT EXISTS idx_scripts_trigger_type ON scripts(trigger_type);

-- Trigger to update updated_at timestamp
CREATE TRIGGER IF NOT EXISTS update_scripts_updated_at
AFTER UPDATE ON scripts
FOR EACH ROW
BEGIN
    UPDATE scripts SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
