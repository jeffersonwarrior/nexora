package shell

import (
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionStatus_String(t *testing.T) {
	tests := []struct {
		status   SessionStatus
		expected string
	}{
		{SessionAvailable, "available"},
		{SessionBusy, "busy"},
		{SessionDraining, "draining"},
		{SessionStatus(99), "unknown"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.status.String())
	}
}

func TestDefaultPoolConfig(t *testing.T) {
	config := DefaultPoolConfig()
	assert.Equal(t, 10, config.MaxSize)
	assert.Equal(t, 5*time.Minute, config.IdleTimeout)
}

func TestPoolMetrics_CalculateReuseRate(t *testing.T) {
	tests := []struct {
		name     string
		metrics  PoolMetrics
		expected float64
	}{
		{
			name:     "no activity",
			metrics:  PoolMetrics{},
			expected: 0,
		},
		{
			name:     "all created, none reused",
			metrics:  PoolMetrics{Created: 10, Reused: 0},
			expected: 0,
		},
		{
			name:     "50% reuse rate",
			metrics:  PoolMetrics{Created: 5, Reused: 5},
			expected: 50,
		},
		{
			name:     "100% reuse rate",
			metrics:  PoolMetrics{Created: 0, Reused: 10},
			expected: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateReuseRate(tt.metrics)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPooledSession(t *testing.T) {
	session := &TmuxSession{
		ID:          "test-session",
		SessionName: "nexora-test-session",
		WorkingDir:  "/tmp",
	}

	pooled := &PooledSession{
		TmuxSession: session,
		Status:      SessionAvailable,
		LastUsedAt:  time.Now(),
	}

	assert.Equal(t, "test-session", pooled.ID)
	assert.Equal(t, SessionAvailable, pooled.Status)
}

// countNexoraTmuxSessions counts active nexora-prefixed tmux sessions
func countNexoraTmuxSessions() int {
	cmd := exec.Command("tmux", "ls")
	output, err := cmd.Output()
	if err != nil {
		return 0 // tmux not running or no sessions
	}

	count := 0
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "nexora-") {
			count++
		}
	}
	return count
}

// TestSessionPoolPreventsLeaks verifies that session pooling prevents unbounded
// session accumulation. This is a regression test for issue #18.
func TestSessionPoolPreventsLeaks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TMUX-dependent test in short mode")
	}

	if !IsTmuxAvailable() {
		t.Skip("tmux not available")
	}

	// Count sessions before test
	initialCount := countNexoraTmuxSessions()

	// Get the manager and configure small pool for testing
	manager := GetTmuxManager()
	manager.poolMu.Lock()
	originalConfig := manager.poolConfig
	manager.poolConfig = PoolConfig{
		MaxSize:     3,
		IdleTimeout: 1 * time.Minute,
	}
	manager.poolMu.Unlock()

	// Restore config after test
	defer func() {
		manager.poolMu.Lock()
		manager.poolConfig = originalConfig
		manager.poolMu.Unlock()
	}()

	// Simulate 10 "conversations" requesting sessions
	// Without pooling, this would create 10 sessions
	// With pooling (max 3), should create at most 3
	conversationIDs := []string{
		"conv-leak-test-1", "conv-leak-test-2", "conv-leak-test-3",
		"conv-leak-test-4", "conv-leak-test-5", "conv-leak-test-6",
		"conv-leak-test-7", "conv-leak-test-8", "conv-leak-test-9",
		"conv-leak-test-10",
	}

	var sessions []*TmuxSession
	for _, convID := range conversationIDs {
		session, err := manager.GetOrCreateDefaultSession(convID, "/tmp")
		if err != nil {
			// Pool exhausted is expected after max size reached
			t.Logf("Expected pool exhaustion for %s: %v", convID, err)
			continue
		}
		sessions = append(sessions, session)

		// Release immediately to simulate conversation ending
		manager.ReleaseDefaultSession(convID)
	}

	// Count sessions after test
	finalCount := countNexoraTmuxSessions()
	newSessions := finalCount - initialCount

	// CRITICAL: With pooling, we should NOT have created 10 new sessions
	// At most we should have created MaxSize (3) sessions
	maxAllowed := 3
	require.LessOrEqual(t, newSessions, maxAllowed,
		"Session leak detected! Created %d sessions but max pool size is %d. "+
			"This indicates sessions are not being reused.", newSessions, maxAllowed)

	// Verify reuse happened
	metrics := manager.GetMetrics()
	t.Logf("Pool metrics: created=%d, reused=%d, released=%d",
		metrics.Created, metrics.Reused, metrics.Released)

	// If we ran 10 conversations with max 3 pool size, we should have reused sessions
	if len(sessions) > maxAllowed {
		require.Greater(t, metrics.Reused, int64(0),
			"Sessions were not reused despite pool being full")
	}

	// Cleanup test sessions
	for _, convID := range conversationIDs {
		manager.ReleaseDefaultSession(convID)
	}
}

// TestOrphanedSessionDetection is a canary test that fails if too many
// nexora tmux sessions exist, indicating a potential leak in production.
func TestOrphanedSessionDetection(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TMUX-dependent test in short mode")
	}

	if !IsTmuxAvailable() {
		t.Skip("tmux not available")
	}

	count := countNexoraTmuxSessions()

	// Threshold: more than 20 sessions indicates a likely leak
	// Normal operation with pooling should never exceed MaxSize (10 default)
	const leakThreshold = 20

	if count > leakThreshold {
		t.Errorf("POTENTIAL SESSION LEAK DETECTED: Found %d nexora tmux sessions. "+
			"Threshold is %d. Run 'tmux ls | grep nexora' to inspect. "+
			"Clean up with: tmux ls | grep nexora | cut -d: -f1 | xargs -I {} tmux kill-session -t {}",
			count, leakThreshold)
	} else {
		t.Logf("Session count OK: %d (threshold: %d)", count, leakThreshold)
	}
}
