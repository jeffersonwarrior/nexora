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

-- Checkpoints
CREATE TABLE IF NOT EXISTS checkpoints (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    timestamp INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    token_count INTEGER NOT NULL DEFAULT 0 CHECK (token_count >= 0),
    message_count INTEGER NOT NULL DEFAULT 0 CHECK (message_count >= 0),
    context_hash TEXT NOT NULL,
    state BLOB NOT NULL,
    compressed BOOLEAN NOT NULL DEFAULT 0,
    created_at INTEGER DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (session_id) REFERENCES sessions(id) ON DELETE CASCADE
);

-- Prompt Library
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

-- FTS virtual table for prompt search
CREATE VIRTUAL TABLE IF NOT EXISTS prompt_library_fts USING fts5(
    title, description, content, tags,
    content=prompt_library
);
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

func TestSessionList(t *testing.T) {
	ctx := context.Background()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = createTestSchema(db)
	require.NoError(t, err)

	q := New(db)

	tests := []struct {
		name  string
		id    string
		title string
	}{
		{"first session", "sess-1", "First Session"},
		{"second session", "sess-2", "Second Session"},
		{"third session", "sess-3", "Third Session"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := q.CreateSession(ctx, CreateSessionParams{
				ID:    tt.id,
				Title: tt.title,
			})
			require.NoError(t, err)
		})
	}

	// List all sessions
	sessions, err := q.ListSessions(ctx)
	require.NoError(t, err)
	require.Len(t, sessions, 3)
}

func TestSessionUpdate(t *testing.T) {
	ctx := context.Background()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = createTestSchema(db)
	require.NoError(t, err)

	q := New(db)

	// Create session
	session, err := q.CreateSession(ctx, CreateSessionParams{
		ID:    "update-test-session",
		Title: "Original Title",
	})
	require.NoError(t, err)
	require.Equal(t, "Original Title", session.Title)

	// Update session
	updated, err := q.UpdateSession(ctx, UpdateSessionParams{
		ID:    session.ID,
		Title: "Updated Title",
	})
	require.NoError(t, err)
	require.Equal(t, "Updated Title", updated.Title)
	require.Equal(t, session.ID, updated.ID)
}

func TestSessionDelete(t *testing.T) {
	ctx := context.Background()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = createTestSchema(db)
	require.NoError(t, err)

	q := New(db)

	// Create session
	session, err := q.CreateSession(ctx, CreateSessionParams{
		ID:    "delete-test-session",
		Title: "To Be Deleted",
	})
	require.NoError(t, err)

	// Delete session
	err = q.DeleteSession(ctx, session.ID)
	require.NoError(t, err)

	// Verify it's gone
	_, err = q.GetSessionByID(ctx, session.ID)
	require.Error(t, err)
}

func TestMessageList(t *testing.T) {
	ctx := context.Background()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = createTestSchema(db)
	require.NoError(t, err)

	q := New(db)

	// Create session
	session, err := q.CreateSession(ctx, CreateSessionParams{
		ID:    "message-list-session",
		Title: "Message List Test",
	})
	require.NoError(t, err)

	// Create multiple messages
	for i := 1; i <= 5; i++ {
		_, err := q.CreateMessage(ctx, CreateMessageParams{
			ID:        string(rune('m')) + string(rune('0'+i)),
			SessionID: session.ID,
			Role:      "user",
			Parts:     "Message content",
		})
		require.NoError(t, err)
	}

	// List messages
	messages, err := q.ListMessagesBySession(ctx, session.ID)
	require.NoError(t, err)
	require.Len(t, messages, 5)
}

func TestQueriesInitialization(t *testing.T) {
	ctx := context.Background()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = createTestSchema(db)
	require.NoError(t, err)

	// Test that Prepare works
	q, err := Prepare(ctx, db)
	require.NoError(t, err)
	require.NotNil(t, q)

	// Test that New works
	q2 := New(db)
	require.NotNil(t, q2)

	// Close prepared statements
	err = q.Close()
	require.NoError(t, err)
}

func TestTransactionSupport(t *testing.T) {
	ctx := context.Background()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = createTestSchema(db)
	require.NoError(t, err)

	// Start a transaction
	tx, err := db.BeginTx(ctx, nil)
	require.NoError(t, err)

	q := New(tx)

	// Create session in transaction
	_, err = q.CreateSession(ctx, CreateSessionParams{
		ID:    "tx-session",
		Title: "Transaction Test",
	})
	require.NoError(t, err)

	// Rollback
	err = tx.Rollback()
	require.NoError(t, err)

	// Verify session doesn't exist
	q2 := New(db)
	_, err = q2.GetSessionByID(ctx, "tx-session")
	require.Error(t, err)
}

func TestBatchOperations(t *testing.T) {
	ctx := context.Background()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = createTestSchema(db)
	require.NoError(t, err)

	q := New(db)

	// Create session
	session, err := q.CreateSession(ctx, CreateSessionParams{
		ID:    "batch-session",
		Title: "Batch Test",
	})
	require.NoError(t, err)

	// Create multiple messages in a loop
	for i := 0; i < 10; i++ {
		msgID := "msg-" + string(rune('0'+i))
		_, err := q.CreateMessage(ctx, CreateMessageParams{
			ID:        msgID,
			SessionID: session.ID,
			Role:      "user",
			Parts:     "Message content",
		})
		require.NoError(t, err)
	}

	// Verify all messages created
	messages, err := q.ListMessagesBySession(ctx, session.ID)
	require.NoError(t, err)
	require.Len(t, messages, 10)

	// Verify session message count updated by trigger
	retrieved, err := q.GetSessionByID(ctx, session.ID)
	require.NoError(t, err)
	require.Equal(t, int64(10), retrieved.MessageCount)
}

func TestNullableFields(t *testing.T) {
	ctx := context.Background()

	db, err := sql.Open("sqlite3", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	err = createTestSchema(db)
	require.NoError(t, err)

	q := New(db)

	// Create session with null parent
	session, err := q.CreateSession(ctx, CreateSessionParams{
		ID:              "null-test-session",
		ParentSessionID: sql.NullString{Valid: false},
		Title:           "Null Test",
	})
	require.NoError(t, err)
	require.False(t, session.ParentSessionID.Valid)

	// Create session with parent
	child, err := q.CreateSession(ctx, CreateSessionParams{
		ID:              "child-session",
		ParentSessionID: sql.NullString{String: session.ID, Valid: true},
		Title:           "Child Session",
	})
	require.NoError(t, err)
	require.True(t, child.ParentSessionID.Valid)
	require.Equal(t, session.ID, child.ParentSessionID.String)
}
