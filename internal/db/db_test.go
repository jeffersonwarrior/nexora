package db

import (
	"context"
	"database/sql"
	"strings"
	"testing"

	_ "github.com/ncruces/go-sqlite3/embed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// createTestSchema creates the complete database schema for testing
func createTestSchema(db *sql.DB) error {
	schema := `
-- Sessions
CREATE TABLE IF NOT EXISTS sessions (
    id TEXT PRIMARY KEY,
    parent_session_id TEXT,
    title TEXT NOT NULL,
    message_count INTEGER NOT NULL DEFAULT 0 CHECK (message_count >= 0),
    prompt_tokens  INTEGER NOT NULL DEFAULT 0 CHECK (prompt_tokens >= 0),
    completion_tokens  INTEGER NOT NULL DEFAULT 0 CHECK (completion_tokens>= 0),
    cost REAL NOT NULL DEFAULT 0.0 CHECK (cost >= 0.0),
    summary_message_id TEXT,
    updated_at INTEGER NOT NULL,  -- Unix timestamp in milliseconds
    created_at INTEGER NOT NULL   -- Unix timestamp in milliseconds
);

-- Messages
CREATE TABLE IF NOT EXISTS messages (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    role TEXT NOT NULL,
    parts TEXT NOT NULL default '[]',
    model TEXT,
    provider TEXT,
    is_summary_message INTEGER DEFAULT 0 NOT NULL,
    created_at INTEGER NOT NULL,  -- Unix timestamp in milliseconds
    updated_at INTEGER NOT NULL,  -- Unix timestamp in milliseconds
    finished_at INTEGER,  -- Unix timestamp in milliseconds
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- Files
CREATE TABLE IF NOT EXISTS files (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    path TEXT NOT NULL,
    content TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 0,
    is_new INTEGER DEFAULT 0 NOT NULL,
    created_at INTEGER NOT NULL,  -- Unix timestamp in milliseconds
    updated_at INTEGER NOT NULL,  -- Unix timestamp in milliseconds
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE,
    UNIQUE(path, session_id, version)
);

-- Triggers
CREATE TRIGGER IF NOT EXISTS update_sessions_updated_at
AFTER UPDATE ON sessions
BEGIN
UPDATE sessions SET updated_at = strftime('%s', 'now')
WHERE id = new.id;
END;

CREATE TRIGGER IF NOT EXISTS update_messages_updated_at
AFTER UPDATE ON messages
BEGIN
UPDATE messages SET updated_at = strftime('%s', 'now')
WHERE id = new.id;
END;

CREATE TRIGGER IF NOT EXISTS update_session_message_count_on_insert
AFTER INSERT ON messages
BEGIN
UPDATE sessions SET
    message_count = message_count + 1
WHERE id = new.session_id;
END;

CREATE TRIGGER IF NOT EXISTS update_session_message_count_on_delete
AFTER DELETE ON messages
BEGIN
UPDATE sessions SET
    message_count = message_count - 1
WHERE id = old.session_id;
END;
`

	if _, err := db.Exec(schema); err != nil {
		return err
	}

	return nil
}

// TestCreateSchema tests creating the complete database schema for tests
func TestCreateSchema(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = createTestSchema(db)
	require.NoError(t, err)

	// Verify tables exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='sessions'").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)

	err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='messages'").Scan(&count)
	require.NoError(t, err)
	require.Equal(t, 1, count)
}

// TestNewCreateClose tests basic operations with queries
func TestNewCreateClose(t *testing.T) {
	ctx := context.Background()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = createTestSchema(db)
	require.NoError(t, err)

	// Prepare queries
	q, err := Prepare(ctx, db)
	require.NoError(t, err)

	// Test close
	err = q.Close()
	require.NoError(t, err)
}

func TestDatabaseOperations(t *testing.T) {
	ctx := context.Background()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = createTestSchema(db)
	require.NoError(t, err)

	q := New(db)

	// Create a session
	session, err := q.CreateSession(ctx, CreateSessionParams{
		ID:    "test-session-1",
		Title: "Test Session 1",
	})
	require.NoError(t, err)
	require.Equal(t, "test-session-1", session.ID)
	require.Equal(t, "Test Session 1", session.Title)

	// Get the session
	retrieved, err := q.GetSessionByID(ctx, session.ID)
	require.NoError(t, err)
	require.Equal(t, session.ID, retrieved.ID)
	require.Equal(t, session.Title, retrieved.Title)

	// Create a message
	message, err := q.CreateMessage(ctx, CreateMessageParams{
		ID:        "test-message-1",
		SessionID: session.ID,
		Role:      "user",
		Parts:     "Hello world",
	})
	require.NoError(t, err)
	require.Equal(t, "test-message-1", message.ID)
	require.Equal(t, session.ID, message.SessionID)
	require.Equal(t, "user", message.Role)
	require.Equal(t, "Hello world", message.Parts)

	// Get message
	retrievedMessage, err := q.GetMessage(ctx, message.ID)
	require.NoError(t, err)
	require.Equal(t, message.ID, retrievedMessage.ID)
}

func TestQueries_PrepareWithNilDB(t *testing.T) {
	ctx := context.Background()

	// Test prepare with nil DB causes a panic, which is expected behavior
	assert.Panics(t, func() {
		var nilDB *sql.DB
		_, _ = Prepare(ctx, nilDB)
	})
}

func TestSessionOperations_EdgeCases(t *testing.T) {
	ctx := context.Background()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = createTestSchema(db)
	require.NoError(t, err)

	q := New(db)

	// Test session with special characters
	sessionParams := CreateSessionParams{
		ID:    "session-with-special-chars-123",
		Title: "Session with 'quotes' and \n newlines",
	}

	session, err := q.CreateSession(ctx, sessionParams)
	require.NoError(t, err)
	require.Equal(t, sessionParams.ID, session.ID)

	// Test update with special characters
	updateParams := UpdateSessionParams{
		ID:    session.ID,
		Title: "Updated title with 特殊 characters",
	}

	_, err = q.UpdateSession(ctx, updateParams)
	require.NoError(t, err)
}

func TestMessageOperations_EdgeCases(t *testing.T) {
	ctx := context.Background()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = createTestSchema(db)
	require.NoError(t, err)

	q := New(db)

	// Create a session first
	session, err := q.CreateSession(ctx, CreateSessionParams{
		ID:    "message-test-session",
		Title: "Message Test Session",
	})
	require.NoError(t, err)

	// Test message with invalid JSON parts (should still store as string)
	messageParams := CreateMessageParams{
		ID:        "invalid-json-message",
		SessionID: session.ID,
		Role:      "user",
		Parts:     "Message with content",
	}

	_, err = q.CreateMessage(ctx, messageParams)
	require.NoError(t, err)

	// Test message with reasoning
	messageParamsReasoning := CreateMessageParams{
		ID:        "reasoning-message",
		SessionID: session.ID,
		Role:      "assistant",
		Parts:     "Final answer",
	}

	_, err = q.CreateMessage(ctx, messageParamsReasoning)
	require.NoError(t, err)
}

func TestLongContent(t *testing.T) {
	ctx := context.Background()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = createTestSchema(db)
	require.NoError(t, err)

	q := New(db)

	// Create a session
	session, err := q.CreateSession(ctx, CreateSessionParams{
		ID:    "long-content-session",
		Title: "Long Content Session",
	})
	require.NoError(t, err)

	// Test with very long content
	longContent := strings.Repeat("This is a very long message content. ", 100)

	messageParams := CreateMessageParams{
		ID:        "long-content-message",
		SessionID: session.ID,
		Role:      "user",
		Parts:     longContent,
	}

	message, err := q.CreateMessage(ctx, messageParams)
	require.NoError(t, err)
	require.Equal(t, longContent, message.Parts)

	// Verify we can retrieve it
	retrieved, err := q.GetMessage(ctx, message.ID)
	require.NoError(t, err)
	require.Equal(t, longContent, retrieved.Parts)
}
