-- +goose Up
-- Migration: Add checkpoints table for session state checkpointing

CREATE TABLE IF NOT EXISTS checkpoints (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    token_count INTEGER NOT NULL DEFAULT 0 CHECK (token_count >= 0),
    message_count INTEGER NOT NULL DEFAULT 0 CHECK (message_count >= 0),
    context_hash TEXT NOT NULL,
    state BLOB NOT NULL,
    compressed BOOLEAN NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

CREATE INDEX idx_checkpoints_session_id ON checkpoints(session_id);
CREATE INDEX idx_checkpoints_timestamp ON checkpoints(session_id, timestamp DESC);
CREATE INDEX idx_checkpoints_created_at ON checkpoints(created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS idx_checkpoints_created_at;
DROP INDEX IF EXISTS idx_checkpoints_timestamp;
DROP INDEX IF EXISTS idx_checkpoints_session_id;
DROP TABLE IF EXISTS checkpoints;
