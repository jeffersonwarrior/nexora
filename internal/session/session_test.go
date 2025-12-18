package session_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/nexora/nexora/internal/db"
	"github.com/nexora/nexora/internal/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockQuerier provides a mock database querier for testing
type MockQuerier struct {
	sessions map[string]db.Session
	calls    []string
}

func NewMockQuerier() *MockQuerier {
	return &MockQuerier{
		sessions: make(map[string]db.Session),
		calls:    make([]string, 0),
	}
}

func (m *MockQuerier) CreateSession(ctx context.Context, params db.CreateSessionParams) (db.Session, error) {
	m.calls = append(m.calls, "CreateSession")
	session := db.Session{
		ID:               params.ID,
		ParentSessionID:  params.ParentSessionID,
		Title:            params.Title,
		MessageCount:     0,
		PromptTokens:     0,
		CompletionTokens: 0,
		Cost:             0,
		CreatedAt:        time.Now().Unix(),
		UpdatedAt:        time.Now().Unix(),
	}
	m.sessions[params.ID] = session
	return session, nil
}

func (m *MockQuerier) GetSession(ctx context.Context, id string) (db.Session, error) {
	m.calls = append(m.calls, "GetSession")
	session, exists := m.sessions[id]
	if !exists {
		return db.Session{}, sql.ErrNoRows
	}
	return session, nil
}

func (m *MockQuerier) ListSessions(ctx context.Context) ([]db.Session, error) {
	m.calls = append(m.calls, "ListSessions")
	sessions := make([]db.Session, 0, len(m.sessions))
	for _, s := range m.sessions {
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (m *MockQuerier) DeleteSession(ctx context.Context, id string) error {
	m.calls = append(m.calls, "DeleteSession")
	if _, exists := m.sessions[id]; !exists {
		return sql.ErrNoRows
	}
	delete(m.sessions, id)
	return nil
}

func (m *MockQuerier) UpdateSession(ctx context.Context, params db.UpdateSessionParams) (db.Session, error) {
	m.calls = append(m.calls, "UpdateSession")
	session, exists := m.sessions[params.ID]
	if !exists {
		return db.Session{}, sql.ErrNoRows
	}
	if params.Title != "" {
		session.Title = params.Title
	}
	session.PromptTokens = params.PromptTokens
	session.CompletionTokens = params.CompletionTokens
	session.Cost = params.Cost
	if params.SummaryMessageID.Valid {
		session.SummaryMessageID = params.SummaryMessageID
	}
	session.UpdatedAt = time.Now().Unix()
	m.sessions[params.ID] = session
	return session, nil
}

func (m *MockQuerier) UpdateMessage(ctx context.Context, params db.UpdateMessageParams) error {
	return nil
}
func (m *MockQuerier) CreateFile(ctx context.Context, params db.CreateFileParams) (db.File, error) {
	return db.File{}, nil
}
func (m *MockQuerier) GetFile(ctx context.Context, id string) (db.File, error) { return db.File{}, nil }
func (m *MockQuerier) DeleteFile(ctx context.Context, id string) error         { return nil }
func (m *MockQuerier) GetMessage(ctx context.Context, id string) (db.Message, error) {
	return db.Message{}, nil
}
func (m *MockQuerier) DeleteSessionFiles(ctx context.Context, sessionID string) error { return nil }
func (m *MockQuerier) GetFileByPathAndSession(ctx context.Context, params db.GetFileByPathAndSessionParams) (db.File, error) {
	return db.File{}, nil
}
func (m *MockQuerier) ListFilesByPath(ctx context.Context, path string) ([]db.File, error) {
	return []db.File{}, nil
}
func (m *MockQuerier) ListFilesBySession(ctx context.Context, sessionID string) ([]db.File, error) {
	return []db.File{}, nil
}
func (m *MockQuerier) ListLatestSessionFiles(ctx context.Context, sessionID string) ([]db.File, error) {
	return []db.File{}, nil
}
func (m *MockQuerier) ListNewFiles(ctx context.Context) ([]db.File, error) { return []db.File{}, nil }
func (m *MockQuerier) CreateMessage(ctx context.Context, params db.CreateMessageParams) (db.Message, error) {
	return db.Message{}, nil
}
func (m *MockQuerier) GetSessionByID(ctx context.Context, id string) (db.Session, error) {
	return m.GetSession(ctx, id)
}
func (m *MockQuerier) ListMessagesBySession(ctx context.Context, sessionID string) ([]db.Message, error) {
	return []db.Message{}, nil
}
func (m *MockQuerier) DeleteMessage(ctx context.Context, id string) error                { return nil }
func (m *MockQuerier) DeleteSessionMessages(ctx context.Context, sessionID string) error { return nil }

// TestDB provides an in-memory SQLite database for testing
type TestDB struct {
	*sql.DB
	q  db.Querier
	t  *testing.T
	db *db.Queries
}

// NewTestDB creates a new in-memory SQLite database with migrations
func NewTestDB(t *testing.T) *TestDB {
	// Use in-memory SQLite for speed
	dsn := ":memory:?_journal_mode=WAL&_synchronous=NORMAL&_cache_size=1000"

	sqlDB, err := sql.Open("sqlite3", dsn)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Test connection
	if err := sqlDB.Ping(); err != nil {
		t.Fatalf("Failed to ping test database: %v", err)
	}

	// Run migrations
	if err := runMigrations(sqlDB); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Create queries
	queries := db.New(sqlDB)

	return &TestDB{
		DB: sqlDB,
		q:  queries,
		t:  t,
		db: queries,
	}
}

// Querier returns the database querier interface
func (tdb *TestDB) Querier() db.Querier {
	return tdb.q
}

// Cleanup closes the database connection
func (tdb *TestDB) Cleanup() {
	if err := tdb.DB.Close(); err != nil {
		tdb.t.Logf("Warning: failed to close test database: %v", err)
	}
}

// CreateTestSession creates a test session in the database
func (tdb *TestDB) CreateTestSession(ctx context.Context, title string) db.Session {
	id := fmt.Sprintf("test-session-%s", title)
	session, err := tdb.q.CreateSession(ctx, db.CreateSessionParams{
		ID:              id,
		Title:           title,
		ParentSessionID: sql.NullString{},
	})
	if err != nil {
		tdb.t.Fatalf("Failed to create test session: %v", err)
	}
	return session
}

// Helper to run migrations
func runMigrations(db *sql.DB) error {
	// For tests, use a complete schema instead of incremental migrations
	// to avoid duplicate column issues in fresh databases
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

CREATE TRIGGER IF NOT EXISTS update_sessions_updated_at
AFTER UPDATE ON sessions
BEGIN
UPDATE sessions SET updated_at = strftime('%s', 'now')
WHERE id = new.id;
END;

-- Files
CREATE TABLE IF NOT EXISTS files (
    id TEXT PRIMARY KEY,
    session_id TEXT NOT NULL,
    path TEXT NOT NULL,
    content TEXT NOT NULL,
    version INTEGER NOT NULL DEFAULT 0,
    created_at INTEGER NOT NULL,  -- Unix timestamp in milliseconds
    updated_at INTEGER NOT NULL,  -- Unix timestamp in milliseconds
    FOREIGN KEY (session_id) REFERENCES sessions (id) ON DELETE CASCADE,
    UNIQUE(path, session_id, version)
);

CREATE INDEX IF NOT EXISTS idx_files_session_id ON files (session_id);
CREATE INDEX IF NOT EXISTS idx_files_path ON files (path);

CREATE TRIGGER IF NOT EXISTS update_files_updated_at
AFTER UPDATE ON files
BEGIN
UPDATE files SET updated_at = strftime('%s', 'now')
WHERE id = new.id;
END;

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
    FOREIGN KEY (session_id) REFERENCES sessions (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages (session_id);
CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages (created_at);
CREATE INDEX IF NOT EXISTS idx_files_created_at ON files (created_at);

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

-- Context Archive
CREATE TABLE IF NOT EXISTS context_archive (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    project TEXT NOT NULL,
    session_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    reference_uri TEXT UNIQUE NOT NULL CHECK (reference_uri LIKE 'nexora://%'),
    summary TEXT NOT NULL,
    token_count INTEGER NOT NULL CHECK (token_count >= 0),
    metadata TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    CHECK (length(summary) > 10)
);

CREATE INDEX IF NOT EXISTS idx_project_session ON context_archive(project, session_id);
CREATE INDEX IF NOT EXISTS idx_project_time ON context_archive(project, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_reference_uri ON context_archive(reference_uri);
CREATE INDEX IF NOT EXISTS idx_created_at ON context_archive(created_at DESC);
`

	if _, err := db.Exec(schema); err != nil {
		return fmt.Errorf("failed to create test schema: %w", err)
	}

	return nil
}

func TestCreateSession(t *testing.T) {
	tests := []struct {
		name     string
		title    string
		expected string
	}{
		{"with title", "My Test Session", "My Test Session"},
		{"empty title", "", "New Session"},
		{"whitespace title", "   ", "New Session"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := NewMockQuerier()
			svc := session.NewService(mock)

			ctx := context.Background()
			session, err := svc.Create(ctx, tt.title)

			require.NoError(t, err)
			assert.NotEmpty(t, session.ID)
			assert.Equal(t, tt.expected, session.Title)
			assert.Equal(t, int64(0), session.MessageCount)
			assert.Greater(t, session.CreatedAt, int64(0))
			assert.Greater(t, session.UpdatedAt, int64(0))

			// Verify mock was called
			assert.Contains(t, mock.calls, "CreateSession")
		})
	}
}

func TestCreateTaskSession(t *testing.T) {
	mock := NewMockQuerier()
	svc := session.NewService(mock)

	ctx := context.Background()
	session, err := svc.CreateTaskSession(ctx, "tool-123", "parent-456", "Build Project")

	require.NoError(t, err)
	assert.Equal(t, "tool-123", session.ID)
	assert.Equal(t, "parent-456", session.ParentSessionID)
	assert.Equal(t, "Build Project", session.Title)
}

func TestCreateTitleSession(t *testing.T) {
	mock := NewMockQuerier()
	svc := session.NewService(mock)

	ctx := context.Background()
	session, err := svc.CreateTitleSession(ctx, "session-789")

	require.NoError(t, err)
	assert.Equal(t, "title-session-789", session.ID)
	assert.Equal(t, "session-789", session.ParentSessionID)
	assert.Equal(t, "Generate a title", session.Title)
}

func TestGetSession(t *testing.T) {
	mock := NewMockQuerier()
	svc := session.NewService(mock)

	ctx := context.Background()

	// Create a session first
	created, err := mock.CreateSession(ctx, db.CreateSessionParams{
		ID:    "test-123",
		Title: "Test Session",
	})
	require.NoError(t, err)

	// Get the session
	session, err := svc.Get(ctx, "test-123")
	require.NoError(t, err)
	assert.Equal(t, created.ID, session.ID)
	assert.Equal(t, created.Title, session.Title)
}

func TestGetNonExistentSession(t *testing.T) {
	mock := NewMockQuerier()
	svc := session.NewService(mock)

	ctx := context.Background()
	_, err := svc.Get(ctx, "non-existent")

	assert.Error(t, err)
}

func TestListSessions(t *testing.T) {
	mock := NewMockQuerier()
	svc := session.NewService(mock)

	ctx := context.Background()

	// Create multiple sessions
	mock.CreateSession(ctx, db.CreateSessionParams{ID: "1", Title: "Session 1"})
	mock.CreateSession(ctx, db.CreateSessionParams{ID: "2", Title: "Session 2"})

	sessions, err := svc.List(ctx)
	require.NoError(t, err)
	assert.Len(t, sessions, 2)
}

func TestSaveSession(t *testing.T) {
	mock := NewMockQuerier()
	svc := session.NewService(mock)

	ctx := context.Background()

	// Create initial session
	_, err := mock.CreateSession(ctx, db.CreateSessionParams{ID: "update-test", Title: "Original"})
	require.NoError(t, err)

	// Update session - only allowed fields are updated
	session := session.Session{
		ID:               "update-test",
		Title:            "Updated Title",
		SummaryMessageID: "update-test",
	}

	saved, err := svc.Save(ctx, session)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", saved.Title)
	// MessageCount is not updated by Save method
	assert.Equal(t, int64(0), saved.MessageCount)
}

func TestDeleteSession(t *testing.T) {
	mock := NewMockQuerier()
	svc := session.NewService(mock)

	ctx := context.Background()

	// Create session first
	mock.CreateSession(ctx, db.CreateSessionParams{ID: "delete-test", Title: "To Delete"})

	// Delete it
	err := svc.Delete(ctx, "delete-test")
	require.NoError(t, err)

	// Verify it's gone
	_, err = mock.GetSession(ctx, "delete-test")
	assert.Error(t, err)
}

func TestDeleteNonExistentSession(t *testing.T) {
	mock := NewMockQuerier()
	svc := session.NewService(mock)

	ctx := context.Background()
	err := svc.Delete(ctx, "non-existent")

	assert.Error(t, err)
}

func TestAgentToolSessionID(t *testing.T) {
	mock := NewMockQuerier()
	svc := session.NewService(mock)

	// Test CreateAgentToolSessionID
	sessionID := svc.CreateAgentToolSessionID("msg-123", "tool-456")
	assert.Equal(t, "msg-123$$tool-456", sessionID)

	// Test ParseAgentToolSessionID
	msgID, toolID, ok := svc.ParseAgentToolSessionID(sessionID)
	assert.True(t, ok)
	assert.Equal(t, "msg-123", msgID)
	assert.Equal(t, "tool-456", toolID)

	// Test with invalid format
	_, _, ok = svc.ParseAgentToolSessionID("invalid")
	assert.False(t, ok)

	// Test IsAgentToolSession
	assert.True(t, svc.IsAgentToolSession(sessionID))
	assert.False(t, svc.IsAgentToolSession("regular-session"))
}

func TestSessionPublishEvents(t *testing.T) {
	mock := NewMockQuerier()
	svc := session.NewService(mock)

	ctx := context.Background()

	// Test basic session creation
	session, err := svc.Create(ctx, "Test Session")
	require.NoError(t, err)
	assert.NotEmpty(t, session.ID)
	assert.Equal(t, "Test Session", session.Title)

	// Note: Full pubsub testing would require access to the broker
	// which is not exposed through the public API
}

func TestWithRealDB(t *testing.T) {
	// This test uses a real in-memory SQLite database
	tdb := NewTestDB(t)
	defer tdb.Cleanup()
	svc := session.NewService(tdb.Querier())

	ctx := context.Background()

	// Test full lifecycle
	sess, err := svc.Create(ctx, "DB Test Session")
	require.NoError(t, err)
	assert.NotEmpty(t, sess.ID)

	// Get it back
	retrieved, err := svc.Get(ctx, sess.ID)
	require.NoError(t, err)
	assert.Equal(t, sess.ID, retrieved.ID)
	assert.Equal(t, sess.Title, retrieved.Title)

	// Update it
	updated := session.Session{
		ID:               sess.ID,
		Title:            "Updated Session",
		MessageCount:     5,
		SummaryMessageID: sess.SummaryMessageID, // fromDBItem sets this
	}

	saved, err := svc.Save(ctx, updated)
	require.NoError(t, err)
	assert.Equal(t, "Updated Session", saved.Title)

	// List sessions
	sessions, err := svc.List(ctx)
	require.NoError(t, err)
	assert.Len(t, sessions, 1)
	assert.Equal(t, saved.ID, sessions[0].ID)

	// Delete it
	err = svc.Delete(ctx, sess.ID)
	require.NoError(t, err)

	// Verify gone
	_, err = svc.Get(ctx, sess.ID)
	assert.Error(t, err)
}
