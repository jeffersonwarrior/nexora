package session_test

import (
	"context"
	"testing"
	"time"

	"github.com/nexora/nexora/internal/session"
	"github.com/stretchr/testify/require"
)

func TestCheckpoint_Create(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Setup test database
	tdb := NewTestDB(t)
	defer tdb.Cleanup()

	q := tdb.Querier()
	svc := session.NewCheckpointService(q)

	// Create a test session
	sessionSvc := session.NewService(q)
	sess, err := sessionSvc.Create(ctx, "Test Session")
	require.NoError(t, err)

	// Create checkpoint
	checkpoint, err := svc.Create(ctx, &sess)
	require.NoError(t, err)
	require.NotEmpty(t, checkpoint.ID)
	require.Equal(t, sess.ID, checkpoint.SessionID)
	require.Equal(t, sess.PromptTokens+sess.CompletionTokens, checkpoint.TokenCount)
	require.Equal(t, sess.MessageCount, checkpoint.MessageCount)
	require.NotEmpty(t, checkpoint.ContextHash)
	require.NotEmpty(t, checkpoint.State)
}

func TestCheckpoint_Restore(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	// Setup test database
	tdb := NewTestDB(t)
	defer tdb.Cleanup()

	q := tdb.Querier()
	svc := session.NewCheckpointService(q)
	sessionSvc := session.NewService(q)

	// Create session and checkpoint
	originalSession, err := sessionSvc.Create(ctx, "Original Session")
	require.NoError(t, err)

	originalSession.PromptTokens = 1000
	originalSession.CompletionTokens = 500
	originalSession.MessageCount = 10
	originalSession, err = sessionSvc.Save(ctx, originalSession)
	require.NoError(t, err)

	checkpoint, err := svc.Create(ctx, &originalSession)
	require.NoError(t, err)

	// Restore from checkpoint
	restoredSession, err := svc.Restore(ctx, checkpoint.ID)
	require.NoError(t, err)
	require.Equal(t, originalSession.ID, restoredSession.ID)
	require.Equal(t, originalSession.Title, restoredSession.Title)
	require.Equal(t, originalSession.PromptTokens, restoredSession.PromptTokens)
	require.Equal(t, originalSession.CompletionTokens, restoredSession.CompletionTokens)
	require.Equal(t, originalSession.MessageCount, restoredSession.MessageCount)
}

func TestCheckpoint_Compression(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tdb := NewTestDB(t)
	defer tdb.Cleanup()

	q := tdb.Querier()
	svc := session.NewCheckpointService(q)
	sessionSvc := session.NewService(q)

	sess, err := sessionSvc.Create(ctx, "Compression Test")
	require.NoError(t, err)

	// Create checkpoint with compression
	config := session.CheckpointConfig{
		CompressionLevel: 6,
	}
	svc.SetConfig(config)

	checkpoint, err := svc.Create(ctx, &sess)
	require.NoError(t, err)
	require.True(t, checkpoint.Compressed)

	// Verify restore works with compression
	restoredSession, err := svc.Restore(ctx, checkpoint.ID)
	require.NoError(t, err)
	require.Equal(t, sess.ID, restoredSession.ID)
}

func TestCheckpoint_Cleanup(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tdb := NewTestDB(t)
	defer tdb.Cleanup()

	q := tdb.Querier()
	svc := session.NewCheckpointService(q)
	sessionSvc := session.NewService(q)

	sess, err := sessionSvc.Create(ctx, "Cleanup Test")
	require.NoError(t, err)

	// Create 5 checkpoints
	for i := 0; i < 5; i++ {
		_, err := svc.Create(ctx, &sess)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond) // Ensure different timestamps
	}

	// Set max checkpoints to 3
	config := session.CheckpointConfig{
		MaxCheckpoints: 3,
	}
	svc.SetConfig(config)

	// Cleanup should keep only 3 most recent
	err = svc.Cleanup(ctx, sess.ID)
	require.NoError(t, err)

	checkpoints, err := svc.List(ctx, sess.ID)
	require.NoError(t, err)
	require.Equal(t, 3, len(checkpoints))
}

func TestCheckpoint_ShouldCheckpoint(t *testing.T) {
	t.Parallel()

	svc := session.NewCheckpointService(nil)

	sess := &session.Session{
		PromptTokens:     5000,
		CompletionTokens: 3000,
		MessageCount:     10,
	}

	tests := []struct {
		name     string
		config   session.CheckpointConfig
		expected bool
	}{
		{
			name: "disabled",
			config: session.CheckpointConfig{
				Enabled: false,
			},
			expected: false,
		},
		{
			name: "token threshold met",
			config: session.CheckpointConfig{
				Enabled:        true,
				TokenThreshold: 7000,
			},
			expected: true,
		},
		{
			name: "token threshold not met",
			config: session.CheckpointConfig{
				Enabled:        true,
				TokenThreshold: 10000,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := svc.ShouldCheckpoint(sess, tt.config)
			require.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckpoint_GetLatest(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tdb := NewTestDB(t)
	defer tdb.Cleanup()

	q := tdb.Querier()
	svc := session.NewCheckpointService(q)
	sessionSvc := session.NewService(q)

	sess, err := sessionSvc.Create(ctx, "Latest Test")
	require.NoError(t, err)

	// Create multiple checkpoints
	var latestCheckpoint *session.Checkpoint
	for i := 0; i < 3; i++ {
		cp, err := svc.Create(ctx, &sess)
		require.NoError(t, err)
		latestCheckpoint = &cp
		time.Sleep(10 * time.Millisecond)
	}

	// Get latest should return the most recent
	latest, err := svc.GetLatest(ctx, sess.ID)
	require.NoError(t, err)
	require.Equal(t, latestCheckpoint.ID, latest.ID)
}

func TestCheckpoint_List(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tdb := NewTestDB(t)
	defer tdb.Cleanup()

	q := tdb.Querier()
	svc := session.NewCheckpointService(q)
	sessionSvc := session.NewService(q)

	sess, err := sessionSvc.Create(ctx, "List Test")
	require.NoError(t, err)

	// Create 3 checkpoints
	for i := 0; i < 3; i++ {
		_, err := svc.Create(ctx, &sess)
		require.NoError(t, err)
		time.Sleep(10 * time.Millisecond)
	}

	checkpoints, err := svc.List(ctx, sess.ID)
	require.NoError(t, err)
	require.Equal(t, 3, len(checkpoints))

	// Verify ordering (newest first)
	for i := 0; i < len(checkpoints)-1; i++ {
		require.True(t, checkpoints[i].Timestamp.After(checkpoints[i+1].Timestamp) ||
			checkpoints[i].Timestamp.Equal(checkpoints[i+1].Timestamp))
	}
}

func TestCheckpoint_Delete(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tdb := NewTestDB(t)
	defer tdb.Cleanup()

	q := tdb.Querier()
	svc := session.NewCheckpointService(q)
	sessionSvc := session.NewService(q)

	sess, err := sessionSvc.Create(ctx, "Delete Test")
	require.NoError(t, err)

	checkpoint, err := svc.Create(ctx, &sess)
	require.NoError(t, err)

	// Delete checkpoint
	err = svc.Delete(ctx, checkpoint.ID)
	require.NoError(t, err)

	// Verify it's gone
	_, err = svc.Get(ctx, checkpoint.ID)
	require.Error(t, err)
}
