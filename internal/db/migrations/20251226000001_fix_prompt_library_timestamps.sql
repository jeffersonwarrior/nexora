-- +goose Up
-- Fix prompt_library timestamp columns from DATETIME to INTEGER
-- SQLite doesn't support ALTER COLUMN, so we need to recreate the table

-- +goose StatementBegin

-- Drop existing triggers first (they will be recreated later by migration 20251225000007)
DROP TRIGGER IF EXISTS update_prompt_library_fts_update;
DROP TRIGGER IF EXISTS update_prompt_library_fts_delete;
DROP TRIGGER IF EXISTS update_prompt_library_fts_insert;
DROP TRIGGER IF EXISTS prompt_library_au;
DROP TRIGGER IF EXISTS prompt_library_ad;
DROP TRIGGER IF EXISTS prompt_library_ai;

-- Drop existing FTS and indexes
DROP TABLE IF EXISTS prompt_library_fts;
DROP INDEX IF EXISTS idx_prompts_tags;
DROP INDEX IF EXISTS idx_prompts_content_hash;
DROP INDEX IF EXISTS idx_prompts_success;
DROP INDEX IF EXISTS idx_prompts_usage;
DROP INDEX IF EXISTS idx_prompts_rating;
DROP INDEX IF EXISTS idx_prompts_category;

-- Create new table with INTEGER timestamps
CREATE TABLE IF NOT EXISTS prompt_library_new (
    id TEXT PRIMARY KEY,
    category TEXT NOT NULL,
    subcategory TEXT,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    content TEXT NOT NULL,
    content_hash TEXT,
    tags TEXT,
    variables TEXT,
    author TEXT,
    source TEXT,
    source_url TEXT,
    votes INTEGER DEFAULT 0,
    rating REAL DEFAULT 0.0,
    usage_count INTEGER DEFAULT 0,
    success_rate REAL DEFAULT 0.5,
    avg_tokens INTEGER DEFAULT 0,
    avg_latency_ms INTEGER DEFAULT 0,
    last_used_at INTEGER,
    favorites_count INTEGER DEFAULT 0,
    created_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

-- Copy existing data, converting timestamps
-- First check if prompt_library exists, if not create it with default values
INSERT OR IGNORE INTO prompt_library_new (
    id, category, subcategory, title, description, content, content_hash, tags, variables,
    author, source, source_url, votes, rating, usage_count, success_rate, avg_tokens,
    avg_latency_ms, last_used_at, favorites_count, created_at, updated_at
)
SELECT
    id, category, subcategory, title, description, content, content_hash, tags, variables,
    author, source, source_url, votes, rating, usage_count, success_rate, avg_tokens,
    avg_latency_ms, last_used_at, favorites_count,
    CASE
        WHEN typeof(created_at) = 'integer' THEN created_at
        WHEN created_at IS NULL THEN strftime('%s', 'now')
        ELSE strftime('%s', created_at)
    END,
    CASE
        WHEN typeof(updated_at) = 'integer' THEN updated_at
        WHEN updated_at IS NULL THEN strftime('%s', 'now')
        ELSE strftime('%s', updated_at)
    END
FROM prompt_library;

-- Drop old table
DROP TABLE IF EXISTS prompt_library;

-- Rename new table
ALTER TABLE prompt_library_new RENAME TO prompt_library;

-- Recreate FTS table
CREATE VIRTUAL TABLE IF NOT EXISTS prompt_library_fts USING fts5(
    title, description, content, tags,
    content=prompt_library
);

-- Recreate indexes
CREATE INDEX IF NOT EXISTS idx_prompts_category ON prompt_library(category);
CREATE INDEX IF NOT EXISTS idx_prompts_rating ON prompt_library(rating DESC);
CREATE INDEX IF NOT EXISTS idx_prompts_usage ON prompt_library(usage_count DESC);
CREATE INDEX IF NOT EXISTS idx_prompts_success ON prompt_library(success_rate DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_prompts_content_hash ON prompt_library(content_hash) WHERE content_hash IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_prompts_tags ON prompt_library(tags);

-- +goose StatementEnd

-- Recreate FTS triggers (goose already ran 20251225000007 before this migration)
-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS update_prompt_library_fts_insert AFTER INSERT ON prompt_library
BEGIN
    INSERT INTO prompt_library_fts(rowid, title, description, content, tags)
    VALUES (NEW.rowid, NEW.title, NEW.description, NEW.content, NEW.tags);
END;

CREATE TRIGGER IF NOT EXISTS update_prompt_library_fts_delete AFTER DELETE ON prompt_library
BEGIN
    INSERT INTO prompt_library_fts(prompt_library_fts, rowid, title, description, content, tags)
    VALUES ('delete', OLD.rowid, OLD.title, OLD.description, OLD.content, OLD.tags);
END;

CREATE TRIGGER IF NOT EXISTS update_prompt_library_fts_update AFTER UPDATE ON prompt_library
BEGIN
    INSERT INTO prompt_library_fts(prompt_library_fts, rowid, title, description, content, tags)
    VALUES ('delete', OLD.rowid, OLD.title, OLD.description, OLD.content, OLD.tags);
    INSERT INTO prompt_library_fts(rowid, title, description, content, tags)
    VALUES (NEW.rowid, NEW.title, NEW.description, NEW.content, NEW.tags);
END;
-- +goose StatementEnd

-- +goose Down
-- Reverting would require similar table recreation with DATETIME columns
-- For simplicity, this is a one-way migration
