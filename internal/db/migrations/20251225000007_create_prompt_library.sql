-- +goose Up
-- Create prompt_library table for prompt management
-- +goose StatementBegin

CREATE TABLE IF NOT EXISTS prompt_library (
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

-- Create indexes for common queries
CREATE INDEX IF NOT EXISTS idx_prompts_category ON prompt_library(category);
CREATE INDEX IF NOT EXISTS idx_prompts_rating ON prompt_library(rating DESC);
CREATE INDEX IF NOT EXISTS idx_prompts_usage ON prompt_library(usage_count DESC);
CREATE INDEX IF NOT EXISTS idx_prompts_success ON prompt_library(success_rate DESC);
CREATE UNIQUE INDEX IF NOT EXISTS idx_prompts_content_hash ON prompt_library(content_hash) WHERE content_hash IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_prompts_tags ON prompt_library(tags);

-- Create FTS virtual table for full-text search
CREATE VIRTUAL TABLE IF NOT EXISTS prompt_library_fts USING fts5(
    title, description, content, tags,
    content=prompt_library
);

-- Create triggers to keep FTS in sync
CREATE TRIGGER IF NOT EXISTS prompt_library_ai AFTER INSERT ON prompt_library
BEGIN
    INSERT INTO prompt_library_fts(rowid, title, description, content, tags)
    VALUES (NEW.rowid, NEW.title, NEW.description, NEW.content, NEW.tags);
END;

CREATE TRIGGER IF NOT EXISTS prompt_library_ad AFTER DELETE ON prompt_library
BEGIN
    INSERT INTO prompt_library_fts(prompt_library_fts, rowid, title, description, content, tags)
    VALUES ('delete', OLD.rowid, OLD.title, OLD.description, OLD.content, OLD.tags);
END;

CREATE TRIGGER IF NOT EXISTS prompt_library_au AFTER UPDATE ON prompt_library
BEGIN
    INSERT INTO prompt_library_fts(prompt_library_fts, rowid, title, description, content, tags)
    VALUES ('delete', OLD.rowid, OLD.title, OLD.description, OLD.content, OLD.tags);
END;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS prompt_library_ai;
DROP TRIGGER IF EXISTS prompt_library_ad;
DROP TRIGGER IF EXISTS prompt_library_au;
DROP TABLE IF EXISTS prompt_library_fts;
DROP INDEX IF EXISTS idx_prompts_tags;
DROP INDEX IF EXISTS idx_prompts_content_hash;
DROP INDEX IF EXISTS idx_prompts_success;
DROP INDEX IF EXISTS idx_prompts_usage;
DROP INDEX IF EXISTS idx_prompts_rating;
DROP INDEX IF EXISTS idx_prompts_category;
DROP TABLE IF EXISTS prompt_library;
-- +goose StatementEnd
