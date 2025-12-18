-- +goose Up
-- Migration: Add context_archive table for Context Management V2

CREATE TABLE IF NOT EXISTS context_archive (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project TEXT NOT NULL,
    session_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    reference_uri TEXT UNIQUE NOT NULL CHECK (reference_uri LIKE 'nexora://%'),
    summary TEXT NOT NULL,
    token_count INTEGER NOT NULL CHECK (token_count >= 0),
    metadata JSONH,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    CHECK (length(summary) > 10)
);

CREATE INDEX idx_project_session ON context_archive(project, session_id);
CREATE INDEX idx_project_time ON context_archive(project, created_at DESC);
CREATE INDEX idx_reference_uri ON context_archive(reference_uri);
CREATE INDEX idx_created_at ON context_archive(created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_created_at;
DROP INDEX IF EXISTS idx_reference_uri;
DROP INDEX IF EXISTS idx_project_time;
DROP INDEX IF EXISTS idx_project_session;
DROP TABLE IF EXISTS context_archive;
