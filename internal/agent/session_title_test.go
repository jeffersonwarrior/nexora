package agent

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestTitleGenerationForNewSession verifies that sessions with the default
// "New Session" title get retitled on the first message.
// This tests the fix for the issue where MessageCount == 0 check alone
// wasn't sufficient to trigger title generation.
func TestTitleGenerationForNewSession(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping on windows for now")
	}

	// Check if at least one provider has API keys
	hasAnyKey := hasAPIKey("anthropic") || hasAPIKey("openai") || hasAPIKey("openrouter") || hasAPIKey("zai")
	if !hasAnyKey {
		t.Skip("Agent tests skipped - no API keys set (set NEXORA_ANTHROPIC_API_KEY, NEXORA_OPENAI_API_KEY, etc.)")
	}

	for _, pair := range modelPairs {
		t.Run(pair.name+"/title_generation", func(t *testing.T) {
			// Skip known VCR failing providers
			if pair.name == "openai-gpt-5" || pair.name == "zai-glm4.6" {
				t.Skip("Skipping known VCR failing provider - see NEXORA.md for details")
			}
			
			agent, env := setupAgent(t, pair)

			// Create session with default "New Session" title
			session, err := env.sessions.Create(t.Context(), "New Session")
			require.NoError(t, err)
			assert.Equal(t, "New Session", session.Title, "Session should be created with 'New Session' title")

			// Send first message
			res, err := agent.Run(t.Context(), SessionAgentCall{
				Prompt:          "Hello, help me write a Go function that adds two numbers",
				SessionID:       session.ID,
				MaxOutputTokens: 10000,
			})
			require.NoError(t, err)
			assert.NotNil(t, res)

			// Fetch updated session from database
			updatedSession, err := env.sessions.Get(t.Context(), session.ID)
			require.NoError(t, err)

			// Verify title was updated (should not be "New Session")
			assert.NotEqual(t, "New Session", updatedSession.Title,
				"Session title should be updated from 'New Session' after first message")
			assert.NotEmpty(t, updatedSession.Title,
				"Session title should not be empty after generation")
			assert.LessOrEqual(t, len(updatedSession.Title), 50,
				"Generated title should be 50 characters or less")
		})
	}
}
