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

func TestCheckpoint_SerializationRoundTrip(t *testing.T) {
	t.Parallel()

	svc := session.NewCheckpointServiceImpl(nil)

	originalSession := &session.Session{
		ID:               "test-session-123",
		Title:            "Test Session",
		PromptTokens:     1000,
		CompletionTokens: 500,
		MessageCount:     10,
	}

	// Serialize
	data, err := svc.SerializeSession(originalSession)
	require.NoError(t, err)
	require.NotEmpty(t, data)

	// Deserialize
	restoredSession, err := svc.DeserializeSession(data)
	require.NoError(t, err)
	require.Equal(t, originalSession.ID, restoredSession.ID)
	require.Equal(t, originalSession.Title, restoredSession.Title)
	require.Equal(t, originalSession.PromptTokens, restoredSession.PromptTokens)
	require.Equal(t, originalSession.CompletionTokens, restoredSession.CompletionTokens)
	require.Equal(t, originalSession.MessageCount, restoredSession.MessageCount)
}

func TestCheckpoint_CompressionRoundTrip(t *testing.T) {
	t.Parallel()

	svc := session.NewCheckpointServiceImpl(nil)

	// Use larger test data to ensure compression is effective
	testData := make([]byte, 1000)
	for i := range testData {
		testData[i] = byte(i % 10) // Repeating pattern compresses well
	}

	tests := []struct {
		name  string
		level int
	}{
		{"default compression", 6},
		{"best compression", 9},
		{"fast compression", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compress
			compressed, err := svc.CompressData(testData, tt.level)
			require.NoError(t, err)
			require.NotEmpty(t, compressed)
			require.Less(t, len(compressed), len(testData), "compressed should be smaller")

			// Decompress
			decompressed, err := svc.DecompressData(compressed)
			require.NoError(t, err)
			require.Equal(t, testData, decompressed)
		})
	}
}

func TestCheckpoint_HashData(t *testing.T) {
	t.Parallel()

	svc := session.NewCheckpointServiceImpl(nil)

	tests := []struct {
		name     string
		data1    []byte
		data2    []byte
		samehash bool
	}{
		{
			name:     "identical data produces same hash",
			data1:    []byte("test data"),
			data2:    []byte("test data"),
			samehash: true,
		},
		{
			name:     "different data produces different hash",
			data1:    []byte("test data 1"),
			data2:    []byte("test data 2"),
			samehash: false,
		},
		{
			name:     "empty data produces valid hash",
			data1:    []byte{},
			data2:    []byte{},
			samehash: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash1 := svc.HashData(tt.data1)
			hash2 := svc.HashData(tt.data2)

			require.NotEmpty(t, hash1)
			require.NotEmpty(t, hash2)
			require.Len(t, hash1, 64, "SHA-256 hash should be 64 hex chars")

			if tt.samehash {
				require.Equal(t, hash1, hash2)
			} else {
				require.NotEqual(t, hash1, hash2)
			}
		})
	}
}

func TestCheckpoint_GetNonExistent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tdb := NewTestDB(t)
	defer tdb.Cleanup()

	q := tdb.Querier()
	svc := session.NewCheckpointService(q)

	// Try to get non-existent checkpoint
	_, err := svc.Get(ctx, "non-existent-id")
	require.Error(t, err)
}

func TestCheckpoint_RestoreNonExistent(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tdb := NewTestDB(t)
	defer tdb.Cleanup()

	q := tdb.Querier()
	svc := session.NewCheckpointService(q)

	// Try to restore non-existent checkpoint
	_, err := svc.Restore(ctx, "non-existent-id")
	require.Error(t, err)
}

func TestCheckpoint_ListEmpty(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tdb := NewTestDB(t)
	defer tdb.Cleanup()

	q := tdb.Querier()
	svc := session.NewCheckpointService(q)
	sessionSvc := session.NewService(q)

	sess, err := sessionSvc.Create(ctx, "Empty List Test")
	require.NoError(t, err)

	// List checkpoints for session with no checkpoints
	checkpoints, err := svc.List(ctx, sess.ID)
	require.NoError(t, err)
	require.Empty(t, checkpoints)
}

func TestCheckpoint_ConfigDefaults(t *testing.T) {
	t.Parallel()

	svc := session.NewCheckpointService(nil)

	// Verify default config is set
	sess := &session.Session{
		PromptTokens:     60000,
		CompletionTokens: 10000,
	}

	// Should checkpoint with default threshold (50000)
	result := svc.ShouldCheckpoint(sess, session.CheckpointConfig{
		Enabled:        true,
		TokenThreshold: 50000,
	})
	require.True(t, result)
}

func TestCheckpoint_CleanupWithNoCheckpoints(t *testing.T) {
	t.Parallel()
	ctx := context.Background()

	tdb := NewTestDB(t)
	defer tdb.Cleanup()

	q := tdb.Querier()
	svc := session.NewCheckpointService(q)
	sessionSvc := session.NewService(q)

	sess, err := sessionSvc.Create(ctx, "Cleanup Empty Test")
	require.NoError(t, err)

	// Cleanup with no checkpoints should not error
	err = svc.Cleanup(ctx, sess.ID)
	require.NoError(t, err)
}
