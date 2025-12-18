package sessionlog

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// PostgreSQLLogger handles logging to PostgreSQL
type PostgreSQLLogger struct {
	db *sql.DB
}

// NewPostgreSQLLogger creates a new PostgreSQL logger
func NewPostgreSQLLogger(connStr string) (*PostgreSQLLogger, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open PostgreSQL connection: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return &PostgreSQLLogger{db: db}, nil
}

// SessionStart logs the start of a session
func (l *PostgreSQLLogger) SessionStart(ctx context.Context, sessionID, instanceID string, metadata map[string]interface{}) error {
	metadataJSON, _ := json.Marshal(metadata)

	query := `
		INSERT INTO nexora_sessions (session_id, instance_id, status, metadata)
		VALUES ($1, $2, $3, $4)
	`
	_, err := l.db.ExecContext(ctx, query, sessionID, instanceID, "active", metadataJSON)
	return err
}

// SessionEnd logs the end of a session
func (l *PostgreSQLLogger) SessionEnd(ctx context.Context, sessionID string, status string, errorCount, toolCount int) error {
	query := `
		UPDATE nexora_sessions
		SET ended_at = CURRENT_TIMESTAMP,
		    duration_seconds = EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - started_at))::INTEGER,
		    status = $2,
		    error_count = $3,
		    tool_count = $4
		WHERE session_id = $1
	`
	_, err := l.db.ExecContext(ctx, query, sessionID, status, errorCount, toolCount)
	return err
}

// EditOperation logs an edit operation
func (l *PostgreSQLLogger) EditOperation(ctx context.Context, edit EditOperationLog) error {
	metadataJSON, _ := json.Marshal(edit.Metadata)

	query := `
		INSERT INTO nexora_edit_operations (
			session_id, instance_id, file_path, status, failure_reason,
			old_string_length, new_string_length, replacement_count, attempt_count,
			duration_ms, has_tabs, has_mixed_indent, file_line_endings, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`
	_, err := l.db.ExecContext(ctx, query,
		edit.SessionID,
		edit.InstanceID,
		edit.FilePath,
		edit.Status,
		edit.FailureReason,
		edit.OldStringLength,
		edit.NewStringLength,
		edit.ReplacementCount,
		edit.AttemptCount,
		edit.DurationMS,
		edit.HasTabs,
		edit.HasMixedIndent,
		edit.FileLineEndings,
		metadataJSON,
	)
	return err
}

// ViewOperation logs a view operation
func (l *PostgreSQLLogger) ViewOperation(ctx context.Context, view ViewOperationLog) error {
	metadataJSON, _ := json.Marshal(view.Metadata)

	query := `
		INSERT INTO nexora_view_operations (
			session_id, instance_id, file_path, offset_line, limit_lines,
			file_size_bytes, total_lines, status, error_reason, duration_ms, metadata
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	_, err := l.db.ExecContext(ctx, query,
		view.SessionID,
		view.InstanceID,
		view.FilePath,
		view.OffsetLine,
		view.LimitLines,
		view.FileSizeBytes,
		view.TotalLines,
		view.Status,
		view.ErrorReason,
		view.DurationMS,
		metadataJSON,
	)
	return err
}

// Close closes the database connection
func (l *PostgreSQLLogger) Close() error {
	return l.db.Close()
}

// ===================== Types =====================

// EditOperationLog represents an edit operation log entry
type EditOperationLog struct {
	SessionID        string
	InstanceID       string
	FilePath         string
	Status           string // 'success', 'failure'
	FailureReason    string // 'whitespace', 'not_found', etc
	OldStringLength  int
	NewStringLength  int
	ReplacementCount int
	AttemptCount     int
	DurationMS       float64
	HasTabs          bool
	HasMixedIndent   bool
	FileLineEndings  string // 'LF', 'CRLF', 'Mixed'
	Metadata         map[string]interface{}
}

// ViewOperationLog represents a view operation log entry
type ViewOperationLog struct {
	SessionID     string
	InstanceID    string
	FilePath      string
	OffsetLine    int
	LimitLines    int
	FileSizeBytes int64
	TotalLines    int
	Status        string // 'success', 'error'
	ErrorReason   string
	DurationMS    float64
	Metadata      map[string]interface{}
}
