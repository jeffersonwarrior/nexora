package shell

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
