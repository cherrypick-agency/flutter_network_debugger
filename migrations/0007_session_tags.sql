-- Session Tags & Annotations feature
-- Allows tagging sessions and adding key-value annotations

-- Predefined tags with colors and categories
CREATE TABLE IF NOT EXISTS predefined_tags (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    color TEXT NOT NULL DEFAULT '#808080',
    category TEXT DEFAULT 'general',
    display_order INTEGER DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Session tags (supports both predefined and custom tags)
CREATE TABLE IF NOT EXISTS session_tags (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    tag_name TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(session_id, tag_name)
);

-- Session annotations (key-value pairs)
CREATE TABLE IF NOT EXISTS session_annotations (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    key TEXT NOT NULL,
    value TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(session_id, key)
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_session_tags_session_id ON session_tags(session_id);
CREATE INDEX IF NOT EXISTS idx_session_tags_tag_name ON session_tags(tag_name);
CREATE INDEX IF NOT EXISTS idx_session_annotations_session_id ON session_annotations(session_id);
CREATE INDEX IF NOT EXISTS idx_predefined_tags_display_order ON predefined_tags(display_order);

-- Trigger to update updated_at timestamp on annotations
CREATE TRIGGER IF NOT EXISTS update_session_annotations_updated_at
AFTER UPDATE ON session_annotations
FOR EACH ROW
BEGIN
    UPDATE session_annotations SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;

-- Insert default predefined tags
INSERT INTO predefined_tags (id, name, color, category, display_order) VALUES
    ('tag-bug', 'bug', '#EF4444', 'issue', 1),
    ('tag-important', 'important', '#F59E0B', 'priority', 2),
    ('tag-review', 'review', '#3B82F6', 'workflow', 3),
    ('tag-security', 'security', '#DC2626', 'issue', 4),
    ('tag-performance', 'performance', '#8B5CF6', 'analysis', 5),
    ('tag-error', 'error', '#EF4444', 'status', 6),
    ('tag-success', 'success', '#10B981', 'status', 7),
    ('tag-pending', 'pending', '#F59E0B', 'status', 8),
    ('tag-testing', 'testing', '#06B6D4', 'workflow', 9),
    ('tag-production', 'production', '#DC2626', 'environment', 10),
    ('tag-staging', 'staging', '#F59E0B', 'environment', 11),
    ('tag-development', 'development', '#10B981', 'environment', 12)
ON CONFLICT(name) DO NOTHING;
