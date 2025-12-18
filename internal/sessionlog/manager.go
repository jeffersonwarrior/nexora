package sessionlog

import (
	"context"
	"log/slog"
	"os"
	"sync"
	"time"
)

// Config holds session logger configuration
type Config struct {
	// PostgreSQL connection string
	PostgresConnStr string
	// Instance ID for this Nexora instance
	InstanceID string
	// Enable logging (default: true)
	Enabled bool
	// Batch size for async writes (default: 50)
	BatchSize int
	// Batch timeout (default: 100ms)
	BatchTimeout time.Duration
}

// Manager manages session logging across PostgreSQL and SQLite
type Manager struct {
	config     Config
	pgLogger   *PostgreSQLLogger
	sessionID  string
	instanceID string
	mu         sync.Mutex
	editQueue  []EditOperationLog
	viewQueue  []ViewOperationLog
	closeChan  chan struct{}
	wg         sync.WaitGroup
}

// NewManager creates a new session log manager
func NewManager(config Config) (*Manager, error) {
	if !config.Enabled {
		return &Manager{config: config, closeChan: make(chan struct{})}, nil
	}

	if config.BatchSize == 0 {
		config.BatchSize = 50
	}
	if config.BatchTimeout == 0 {
		config.BatchTimeout = 100 * time.Millisecond
	}

	pgLogger, err := NewPostgreSQLLogger(config.PostgresConnStr)
	if err != nil {
		slog.Warn("Failed to initialize PostgreSQL logger", "error", err)
		// Don't fail completely if PG is unavailable
		pgLogger = nil
	}

	m := &Manager{
		config:     config,
		pgLogger:   pgLogger,
		instanceID: config.InstanceID,
		closeChan:  make(chan struct{}),
		editQueue:  make([]EditOperationLog, 0, config.BatchSize),
		viewQueue:  make([]ViewOperationLog, 0, config.BatchSize),
	}

	// Start background flusher
	if pgLogger != nil {
		m.wg.Add(1)
		go m.flusherLoop()
	}

	return m, nil
}

// StartSession starts a new session
func (m *Manager) StartSession(ctx context.Context, sessionID string, metadata map[string]interface{}) error {
	if !m.config.Enabled || m.pgLogger == nil {
		return nil
	}

	m.sessionID = sessionID
	return m.pgLogger.SessionStart(ctx, sessionID, m.instanceID, metadata)
}

// EndSession ends the current session
func (m *Manager) EndSession(ctx context.Context, status string, errorCount, toolCount int) error {
	if !m.config.Enabled || m.pgLogger == nil {
		return nil
	}

	// Flush any pending operations
	m.flush(ctx)

	return m.pgLogger.SessionEnd(ctx, m.sessionID, status, errorCount, toolCount)
}

// LogEditOperation logs an edit operation asynchronously
func (m *Manager) LogEditOperation(ctx context.Context, edit EditOperationLog) {
	if !m.config.Enabled || m.pgLogger == nil {
		return
	}

	m.mu.Lock()
	m.editQueue = append(m.editQueue, edit)
	shouldFlush := len(m.editQueue) >= m.config.BatchSize
	m.mu.Unlock()

	if shouldFlush {
		m.flush(ctx)
	}
}

// LogViewOperation logs a view operation asynchronously
func (m *Manager) LogViewOperation(ctx context.Context, view ViewOperationLog) {
	if !m.config.Enabled || m.pgLogger == nil {
		return
	}

	m.mu.Lock()
	m.viewQueue = append(m.viewQueue, view)
	shouldFlush := len(m.viewQueue) >= m.config.BatchSize
	m.mu.Unlock()

	if shouldFlush {
		m.flush(ctx)
	}
}

// flush writes queued operations to PostgreSQL
func (m *Manager) flush(ctx context.Context) {
	m.mu.Lock()
	edits := make([]EditOperationLog, len(m.editQueue))
	views := make([]ViewOperationLog, len(m.viewQueue))
	copy(edits, m.editQueue)
	copy(views, m.viewQueue)
	m.editQueue = m.editQueue[:0]
	m.viewQueue = m.viewQueue[:0]
	m.mu.Unlock()

	if len(edits) == 0 && len(views) == 0 {
		return
	}

	// Write to PostgreSQL in background
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		for _, edit := range edits {
			if err := m.pgLogger.EditOperation(ctx, edit); err != nil {
				slog.Error("Failed to log edit operation", "error", err, "file", edit.FilePath)
			}
		}

		for _, view := range views {
			if err := m.pgLogger.ViewOperation(ctx, view); err != nil {
				slog.Error("Failed to log view operation", "error", err, "file", view.FilePath)
			}
		}
	}()
}

// flusherLoop periodically flushes queued operations
func (m *Manager) flusherLoop() {
	defer m.wg.Done()

	ticker := time.NewTicker(m.config.BatchTimeout)
	defer ticker.Stop()

	for {
		select {
		case <-m.closeChan:
			return
		case <-ticker.C:
			m.flush(context.Background())
		}
	}
}

// Close closes the logger and flushes any pending operations
func (m *Manager) Close(ctx context.Context) error {
	close(m.closeChan)
	m.wg.Wait()

	// Final flush
	m.flush(ctx)

	if m.pgLogger != nil {
		return m.pgLogger.Close()
	}
	return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() Config {
	hostname, _ := os.Hostname()
	return Config{
		PostgresConnStr: "user=postgres dbname=nexora_sessions host=localhost sslmode=disable",
		InstanceID:      hostname,
		Enabled:         true,
		BatchSize:       50,
		BatchTimeout:    100 * time.Millisecond,
	}
}
