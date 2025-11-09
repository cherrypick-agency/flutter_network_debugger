-- Migration: Add CHECK constraints and UNIQUE indexes for Tags & Annotations
-- This migration adds data validation constraints and performance indexes
-- Created: 2025-01-XX

-- ============================================================================
-- CHECK CONSTRAINTS for data length validation
-- ============================================================================

-- Add CHECK constraint for session_tags.tag_name length (max 100 characters)
-- This prevents DoS attacks via extremely long tag names
CREATE TABLE IF NOT EXISTS session_tags_new (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    tag_name TEXT NOT NULL CHECK(LENGTH(tag_name) <= 100 AND LENGTH(tag_name) > 0),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(session_id, tag_name)
);

-- Copy data from old table
INSERT INTO session_tags_new (id, session_id, tag_name, created_at)
SELECT id, session_id, tag_name, created_at FROM session_tags;

-- Drop old table and rename new one
DROP TABLE session_tags;
ALTER TABLE session_tags_new RENAME TO session_tags;

-- Add CHECK constraint for predefined_tags.name length (max 100 characters)
CREATE TABLE IF NOT EXISTS predefined_tags_new (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL CHECK(LENGTH(name) <= 100 AND LENGTH(name) > 0),
    color TEXT NOT NULL DEFAULT '#808080',
    category TEXT NOT NULL DEFAULT 'general',
    display_order INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name)
);

-- Copy data from old table
INSERT INTO predefined_tags_new (id, name, color, category, display_order, created_at)
SELECT id, name, color, category, display_order, created_at FROM predefined_tags;

-- Drop old table and rename new one
DROP TABLE predefined_tags;
ALTER TABLE predefined_tags_new RENAME TO predefined_tags;

-- Add CHECK constraints for session_annotations (key: max 255, value: max 10000)
CREATE TABLE IF NOT EXISTS session_annotations_new (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    key TEXT NOT NULL CHECK(LENGTH(key) <= 255 AND LENGTH(key) > 0),
    value TEXT CHECK(value IS NULL OR LENGTH(value) <= 10000),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(session_id, key)
);

-- Copy data from old table
INSERT INTO session_annotations_new (id, session_id, key, value, created_at, updated_at)
SELECT id, session_id, key, value, created_at, updated_at FROM session_annotations;

-- Drop old table and rename new one
DROP TABLE session_annotations;
ALTER TABLE session_annotations_new RENAME TO session_annotations;

-- ============================================================================
-- UNIQUE INDEXES for performance optimization
-- ============================================================================

-- Index for session_tags unique constraint (improves INSERT conflict detection)
CREATE UNIQUE INDEX IF NOT EXISTS idx_session_tags_unique
    ON session_tags(session_id, tag_name);

-- Index for session_annotations unique constraint
CREATE UNIQUE INDEX IF NOT EXISTS idx_session_annotations_unique
    ON session_annotations(session_id, key);

-- Index for predefined_tags name lookup
CREATE UNIQUE INDEX IF NOT EXISTS idx_predefined_tags_name
    ON predefined_tags(name);

-- ============================================================================
-- Additional indexes for query optimization
-- ============================================================================

-- Index for finding all tags for a session (used in UI)
CREATE INDEX IF NOT EXISTS idx_session_tags_session_id
    ON session_tags(session_id);

-- Index for finding all annotations for a session
CREATE INDEX IF NOT EXISTS idx_session_annotations_session_id
    ON session_annotations(session_id);

-- Index for finding sessions by tag name (used in filters)
CREATE INDEX IF NOT EXISTS idx_session_tags_tag_name
    ON session_tags(tag_name);

-- Index for predefined tags by category (used in UI grouping)
CREATE INDEX IF NOT EXISTS idx_predefined_tags_category
    ON predefined_tags(category, display_order);
