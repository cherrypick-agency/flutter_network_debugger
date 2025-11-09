-- Migration 0006: Add compilation support for scripts
-- This migration adds fields needed for in-app WASM compilation

-- Add compilation-related fields to scripts table
ALTER TABLE scripts ADD COLUMN source_code TEXT;
ALTER TABLE scripts ADD COLUMN compilation_status TEXT DEFAULT 'not_compiled'
    CHECK(compilation_status IN ('not_compiled', 'pending', 'compiling', 'success', 'error'));
ALTER TABLE scripts ADD COLUMN compilation_error TEXT;
ALTER TABLE scripts ADD COLUMN dependencies TEXT; -- JSON: {"filename": "content"}
ALTER TABLE scripts ADD COLUMN last_compiled_at TIMESTAMP;

-- Create index for filtering by compilation status
CREATE INDEX IF NOT EXISTS idx_scripts_compilation_status ON scripts(compilation_status);

-- Create trigger to update updated_at timestamp
-- (already exists from previous migration, but ensure it covers new fields)
DROP TRIGGER IF EXISTS update_scripts_updated_at;
CREATE TRIGGER update_scripts_updated_at
AFTER UPDATE ON scripts
FOR EACH ROW
BEGIN
    UPDATE scripts SET updated_at = CURRENT_TIMESTAMP WHERE id = NEW.id;
END;
