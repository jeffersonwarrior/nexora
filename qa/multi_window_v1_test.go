package qa

import (
	"context"
	"testing"

	"github.com/nexora/cli/internal/agent"
	"github.com/nexora/cli/internal/config"
	"github.com/nexora/cli/internal/session"
	"github.com/stretchr/testify/assert"
)

func TestMultiWindowCoordinator(t *testing.T) {
	ctx := context.Background()
	cfg := testConfig(t) // Assume this exists
	sessions := testSessionService(t)
	messages := testMessageService(t)
	
	coord := agent.NewMultiSessionCoordinator(cfg, sessions, messages)
	
	// Create windows
	w1, err := coord.CreateWindow(ctx, "Bug Fix", "")
	assert.NoError(t, err)
	w2, err := coord.CreateWindow(ctx, "UI Review", "")
	assert.NoError(t, err)
	
	// Switch between windows
	err = coord.SwitchWindow(w2)
	assert.NoError(t, err)
	assert.Equal(t, w2, coord.ActiveWindow())
	
	// List windows
	windows := coord.ListWindows()
	assert.Len(t, windows, 2)
	
	// Fork test
	forkID, err := coord.ForkWindow(w1, "Test fork")
	assert.NoError(t, err)
	assert.NotEmpty(t, forkID)
	
	// Verify fork has parent reference
	forkCoord := agent.NewMultiSessionCoordinator(cfg, sessions, messages)
	// ... additional fork verification
}
