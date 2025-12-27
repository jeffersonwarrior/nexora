package session

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/nexora/internal/db"
)

// Checkpoint represents a saved session state
type Checkpoint struct {
	ID           string
	SessionID    string
	Timestamp    time.Time
	TokenCount   int64
	MessageCount int64
	ContextHash  string
	State        []byte
	Compressed   bool
}

// CheckpointConfig configures checkpoint behavior
type CheckpointConfig struct {
	Enabled          bool
	TokenThreshold   int64
	IntervalSeconds  int
	MaxCheckpoints   int
	CompressionLevel int
}

// CheckpointService manages session checkpoints
type CheckpointService interface {
	Create(ctx context.Context, session *Session) (Checkpoint, error)
	GetLatest(ctx context.Context, sessionID string) (Checkpoint, error)
	Get(ctx context.Context, id string) (Checkpoint, error)
	List(ctx context.Context, sessionID string) ([]Checkpoint, error)
	Restore(ctx context.Context, checkpointID string) (*Session, error)
	Delete(ctx context.Context, id string) error
	Cleanup(ctx context.Context, sessionID string) error
	ShouldCheckpoint(session *Session, config CheckpointConfig) bool
	SetConfig(config CheckpointConfig)
}

type checkpointService struct {
	q      db.Querier
	config CheckpointConfig
}

// NewCheckpointService creates a new checkpoint service
func NewCheckpointService(q db.Querier) CheckpointService {
	return &checkpointService{
		q: q,
		config: CheckpointConfig{
			Enabled:          true,
			TokenThreshold:   50000,
			IntervalSeconds:  300,
			MaxCheckpoints:   10,
			CompressionLevel: 6,
		},
	}
}

// SetConfig updates the checkpoint configuration
func (s *checkpointService) SetConfig(config CheckpointConfig) {
	s.config = config
}

// Create creates a new checkpoint from a session
func (s *checkpointService) Create(ctx context.Context, session *Session) (Checkpoint, error) {
	// Serialize session state
	stateData, err := s.serializeSession(session)
	if err != nil {
		return Checkpoint{}, fmt.Errorf("failed to serialize session: %w", err)
	}

	// Optionally compress
	compressed := false
	if s.config.CompressionLevel > 0 {
		compressedData, err := s.compressData(stateData, s.config.CompressionLevel)
		if err != nil {
			return Checkpoint{}, fmt.Errorf("failed to compress state: %w", err)
		}
		stateData = compressedData
		compressed = true
	}

	// Generate hash for deduplication
	hash := s.hashData(stateData)

	// Create checkpoint
	id := uuid.New().String()
	tokenCount := session.PromptTokens + session.CompletionTokens

	dbCheckpoint, err := s.q.CreateCheckpoint(ctx, db.CreateCheckpointParams{
		ID:           id,
		SessionID:    session.ID,
		Timestamp:    time.Now(),
		TokenCount:   tokenCount,
		MessageCount: session.MessageCount,
		ContextHash:  hash,
		State:        stateData,
		Compressed:   compressed,
	})
	if err != nil {
		return Checkpoint{}, fmt.Errorf("failed to create checkpoint: %w", err)
	}

	return s.fromDBCheckpoint(dbCheckpoint), nil
}

// GetLatest retrieves the most recent checkpoint for a session
func (s *checkpointService) GetLatest(ctx context.Context, sessionID string) (Checkpoint, error) {
	dbCheckpoint, err := s.q.GetLatestCheckpoint(ctx, sessionID)
	if err != nil {
		return Checkpoint{}, fmt.Errorf("failed to get latest checkpoint: %w", err)
	}
	return s.fromDBCheckpoint(dbCheckpoint), nil
}

// Get retrieves a specific checkpoint by ID
func (s *checkpointService) Get(ctx context.Context, id string) (Checkpoint, error) {
	dbCheckpoint, err := s.q.GetCheckpoint(ctx, id)
	if err != nil {
		return Checkpoint{}, fmt.Errorf("failed to get checkpoint: %w", err)
	}
	return s.fromDBCheckpoint(dbCheckpoint), nil
}

// List retrieves all checkpoints for a session
func (s *checkpointService) List(ctx context.Context, sessionID string) ([]Checkpoint, error) {
	dbCheckpoints, err := s.q.ListCheckpoints(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to list checkpoints: %w", err)
	}

	checkpoints := make([]Checkpoint, len(dbCheckpoints))
	for i, dbcp := range dbCheckpoints {
		checkpoints[i] = s.fromDBCheckpoint(dbcp)
	}
	return checkpoints, nil
}

// Restore restores a session from a checkpoint
func (s *checkpointService) Restore(ctx context.Context, checkpointID string) (*Session, error) {
	checkpoint, err := s.Get(ctx, checkpointID)
	if err != nil {
		return nil, err
	}

	// Decompress if needed
	stateData := checkpoint.State
	if checkpoint.Compressed {
		decompressedData, err := s.decompressData(stateData)
		if err != nil {
			return nil, fmt.Errorf("failed to decompress checkpoint: %w", err)
		}
		stateData = decompressedData
	}

	// Deserialize session
	session, err := s.deserializeSession(stateData)
	if err != nil {
		return nil, fmt.Errorf("failed to deserialize session: %w", err)
	}

	return session, nil
}

// Delete deletes a checkpoint
func (s *checkpointService) Delete(ctx context.Context, id string) error {
	err := s.q.DeleteCheckpoint(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete checkpoint: %w", err)
	}
	return nil
}

// Cleanup removes old checkpoints beyond MaxCheckpoints
func (s *checkpointService) Cleanup(ctx context.Context, sessionID string) error {
	maxCheckpoints := s.config.MaxCheckpoints
	if maxCheckpoints <= 0 {
		maxCheckpoints = 10
	}

	err := s.q.DeleteOldCheckpoints(ctx, db.DeleteOldCheckpointsParams{
		SessionID:   sessionID,
		SessionID_2: sessionID,
		Limit:       int64(maxCheckpoints),
	})
	if err != nil {
		return fmt.Errorf("failed to cleanup checkpoints: %w", err)
	}
	return nil
}

// ShouldCheckpoint determines if a checkpoint should be created
func (s *checkpointService) ShouldCheckpoint(session *Session, config CheckpointConfig) bool {
	if !config.Enabled {
		return false
	}

	totalTokens := session.PromptTokens + session.CompletionTokens
	if config.TokenThreshold > 0 && totalTokens >= config.TokenThreshold {
		return true
	}

	return false
}

// serializeSession serializes a session to bytes
func (s *checkpointService) serializeSession(session *Session) ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	if err := enc.Encode(session); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// deserializeSession deserializes bytes to a session
func (s *checkpointService) deserializeSession(data []byte) (*Session, error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	var session Session
	if err := dec.Decode(&session); err != nil {
		return nil, err
	}
	return &session, nil
}

// compressData compresses data using gzip
func (s *checkpointService) compressData(data []byte, level int) ([]byte, error) {
	var buf bytes.Buffer
	w, err := gzip.NewWriterLevel(&buf, level)
	if err != nil {
		return nil, err
	}

	if _, err := w.Write(data); err != nil {
		w.Close()
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// decompressData decompresses gzip data
func (s *checkpointService) decompressData(data []byte) ([]byte, error) {
	buf := bytes.NewBuffer(data)
	r, err := gzip.NewReader(buf)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	decompressed, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return decompressed, nil
}

// hashData generates a SHA-256 hash of data
func (s *checkpointService) hashData(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// fromDBCheckpoint converts a database checkpoint to domain model
func (s *checkpointService) fromDBCheckpoint(dbcp db.Checkpoint) Checkpoint {
	return Checkpoint{
		ID:           dbcp.ID,
		SessionID:    dbcp.SessionID,
		Timestamp:    dbcp.Timestamp,
		TokenCount:   dbcp.TokenCount,
		MessageCount: dbcp.MessageCount,
		ContextHash:  dbcp.ContextHash,
		State:        dbcp.State,
		Compressed:   dbcp.Compressed,
	}
}
