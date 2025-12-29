package agent

import (
	"strings"
	"testing"

	"github.com/nexora/nexora/internal/message"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompactor_EstimateTokens(t *testing.T) {
	c := NewCompactor(DefaultCompactionConfig(100000))

	tests := []struct {
		name      string
		msg       message.Message
		minTokens int
		maxTokens int
	}{
		{
			name: "text content",
			msg: message.Message{
				Role: message.User,
				Parts: []message.ContentPart{
					message.TextContent{Text: strings.Repeat("a", 100)},
				},
			},
			minTokens: 20,
			maxTokens: 40,
		},
		{
			name: "tool call",
			msg: message.Message{
				Role: message.Assistant,
				Parts: []message.ContentPart{
					message.ToolCall{
						ID:    "call_1",
						Name:  "bash",
						Input: `{"command": "ls -la"}`,
					},
				},
			},
			minTokens: 25,
			maxTokens: 50,
		},
		{
			name: "tool result",
			msg: message.Message{
				Role: message.Tool,
				Parts: []message.ContentPart{
					message.ToolResult{
						ToolCallID: "call_1",
						Name:       "bash",
						Content:    strings.Repeat("output ", 100),
					},
				},
			},
			minTokens: 150,
			maxTokens: 250,
		},
		{
			name: "empty message",
			msg: message.Message{
				Role:  message.User,
				Parts: []message.ContentPart{},
			},
			minTokens: 0,
			maxTokens: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := c.EstimateTokens(tt.msg)
			assert.GreaterOrEqual(t, tokens, tt.minTokens, "tokens should be >= min")
			assert.LessOrEqual(t, tokens, tt.maxTokens, "tokens should be <= max")
		})
	}
}

func TestCompactor_DetermineLevel(t *testing.T) {
	c := NewCompactor(DefaultCompactionConfig(100000))

	tests := []struct {
		name          string
		currentTokens int64
		expected      CompactionLevel
	}{
		{"low usage", 30000, CompactionNone},
		{"moderate usage", 55000, CompactionTruncate},
		{"high usage", 70000, CompactionDropToolResults},
		{"very high usage", 80000, CompactionKeepRecent},
		{"critical usage", 90000, CompactionAggressive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			level := c.DetermineLevel(tt.currentTokens)
			assert.Equal(t, tt.expected, level)
		})
	}
}

func TestCompactor_TruncateToolOutputs(t *testing.T) {
	config := DefaultCompactionConfig(100000)
	config.MaxToolOutputTokens = 50 // Low threshold for testing
	c := NewCompactor(config)

	msgs := []message.Message{
		{
			Role: message.User,
			Parts: []message.ContentPart{
				message.TextContent{Text: "Run a command"},
			},
		},
		{
			Role: message.Tool,
			Parts: []message.ContentPart{
				message.ToolResult{
					ToolCallID: "call_1",
					Name:       "bash",
					Content:    strings.Repeat("long output ", 100), // ~1200 chars
				},
			},
		},
	}

	result, applied := c.Compact(msgs, 60000) // 60% usage triggers truncation

	require.True(t, applied)
	require.Len(t, result, 2)

	// Check tool result was truncated
	toolMsg := result[1]
	require.Equal(t, message.Tool, toolMsg.Role)
	require.Len(t, toolMsg.Parts, 1)

	tr, ok := toolMsg.Parts[0].(message.ToolResult)
	require.True(t, ok)
	assert.Contains(t, tr.Content, "truncated for context management")
	assert.Less(t, len(tr.Content), 1200)
}

func TestCompactor_DropToolResults(t *testing.T) {
	config := DefaultCompactionConfig(100000)
	config.RecentMessageCount = 2
	c := NewCompactor(config)

	msgs := []message.Message{
		{Role: message.User, Parts: []message.ContentPart{message.TextContent{Text: "First"}}},
		{
			Role: message.Tool,
			Parts: []message.ContentPart{
				message.ToolResult{ToolCallID: "1", Content: "old output"},
			},
		},
		{Role: message.User, Parts: []message.ContentPart{message.TextContent{Text: "Second"}}},
		{
			Role: message.Tool,
			Parts: []message.ContentPart{
				message.ToolResult{ToolCallID: "2", Content: "new output"},
			},
		},
	}

	result, applied := c.Compact(msgs, 70000) // 70% triggers dropToolResults

	require.True(t, applied)
	require.Len(t, result, 4)

	// First tool result should be replaced
	tr1, ok := result[1].Parts[0].(message.ToolResult)
	require.True(t, ok)
	assert.Equal(t, "[tool output removed for context management]", tr1.Content)

	// Recent tool result should be preserved
	tr2, ok := result[3].Parts[0].(message.ToolResult)
	require.True(t, ok)
	assert.Equal(t, "new output", tr2.Content)
}

// TestCompactor_DropToolResults_PreservesToolCallPairing tests the MiniMax bug fix.
// Ensures that tool calls are included when their corresponding tool results are in
// the recent message window, preventing "tool call result does not follow tool call" errors.
func TestCompactor_DropToolResults_PreservesToolCallPairing(t *testing.T) {
	config := DefaultCompactionConfig(100000)
	config.RecentMessageCount = 1 // Only keep last 1 message
	c := NewCompactor(config)

	// Create a conversation where:
	// 1. Assistant makes tool call (old, outside recent window)
	// 2. Tool result for that call (recent, inside window)
	// Without the fix, the tool call would be dropped but result kept -> API error
	msgs := []message.Message{
		{
			Role: message.User,
			Parts: []message.ContentPart{
				message.TextContent{Text: "Run ls command"},
			},
		},
		{
			Role: message.Assistant,
			Parts: []message.ContentPart{
				message.ToolCall{
					ID:    "call_abc123",
					Name:  "bash",
					Input: `{"command":"ls"}`,
				},
			},
		},
		{
			Role: message.Tool,
			Parts: []message.ContentPart{
				message.ToolResult{
					ToolCallID: "call_abc123",
					Name:       "bash",
					Content:    "file1.txt\nfile2.txt",
				},
			},
		},
		{
			Role: message.User,
			Parts: []message.ContentPart{
				message.TextContent{Text: "What files are there?"},
			},
		},
	}

	result, applied := c.Compact(msgs, 70000) // 70% triggers dropToolResults

	require.True(t, applied)

	// Find the tool call and tool result in compacted messages
	var foundToolCall bool
	var foundToolResult bool
	var toolCallIdx, toolResultIdx int

	for i, msg := range result {
		for _, part := range msg.Parts {
			if tc, ok := part.(message.ToolCall); ok && tc.ID == "call_abc123" {
				foundToolCall = true
				toolCallIdx = i
			}
			if tr, ok := part.(message.ToolResult); ok && tr.ToolCallID == "call_abc123" {
				foundToolResult = true
				toolResultIdx = i
			}
		}
	}

	// Both tool call and result must be present
	assert.True(t, foundToolCall, "tool call should be preserved to match tool result")
	assert.True(t, foundToolResult, "tool result should be preserved")

	// Tool call must come before tool result
	assert.Less(t, toolCallIdx, toolResultIdx, "tool call must come before its result")
}

// TestCompactor_DropToolResults_MultipleToolPairs tests handling multiple tool call/result pairs.
func TestCompactor_DropToolResults_MultipleToolPairs(t *testing.T) {
	config := DefaultCompactionConfig(100000)
	config.RecentMessageCount = 2
	c := NewCompactor(config)

	msgs := []message.Message{
		{Role: message.User, Parts: []message.ContentPart{message.TextContent{Text: "First request"}}},
		{
			Role: message.Assistant,
			Parts: []message.ContentPart{
				message.ToolCall{ID: "call_1", Name: "bash", Input: `{"command":"echo 1"}`},
			},
		},
		{
			Role: message.Tool,
			Parts: []message.ContentPart{
				message.ToolResult{ToolCallID: "call_1", Content: "1"},
			},
		},
		{Role: message.User, Parts: []message.ContentPart{message.TextContent{Text: "Second request"}}},
		{
			Role: message.Assistant,
			Parts: []message.ContentPart{
				message.ToolCall{ID: "call_2", Name: "bash", Input: `{"command":"echo 2"}`},
			},
		},
		{
			Role: message.Tool,
			Parts: []message.ContentPart{
				message.ToolResult{ToolCallID: "call_2", Content: "2"},
			},
		},
	}

	result, applied := c.Compact(msgs, 70000) // 70% triggers dropToolResults

	require.True(t, applied)

	// Collect all tool call IDs and tool result IDs
	toolCalls := make(map[string]bool)
	toolResults := make(map[string]bool)

	for _, msg := range result {
		for _, part := range msg.Parts {
			if tc, ok := part.(message.ToolCall); ok {
				toolCalls[tc.ID] = true
			}
			if tr, ok := part.(message.ToolResult); ok {
				toolResults[tr.ToolCallID] = true
			}
		}
	}

	// Every tool result must have its corresponding tool call
	for resultID := range toolResults {
		assert.True(t, toolCalls[resultID],
			"tool result %s must have corresponding tool call", resultID)
	}
}

func TestCompactor_KeepRecent(t *testing.T) {
	config := DefaultCompactionConfig(100000)
	config.RecentMessageCount = 2
	c := NewCompactor(config)

	msgs := make([]message.Message, 20)
	for i := 0; i < 20; i++ {
		msgs[i] = message.Message{
			Role: message.User,
			Parts: []message.ContentPart{
				message.TextContent{Text: "Message " + string(rune('A'+i))},
			},
		}
	}

	result, applied := c.Compact(msgs, 80000) // 80% triggers keepRecent

	require.True(t, applied)
	assert.Less(t, len(result), len(msgs))
	// Should keep last 4 messages (2 * 2 for pairs)
	assert.LessOrEqual(t, len(result), 5) // +1 for potential context bridge
}

func TestCompactor_KeepRecent_WithSummary(t *testing.T) {
	config := DefaultCompactionConfig(100000)
	config.RecentMessageCount = 2
	c := NewCompactor(config)

	msgs := []message.Message{
		{
			Role:             message.User,
			IsSummaryMessage: true,
			Parts:            []message.ContentPart{message.TextContent{Text: "Summary of conversation"}},
		},
	}
	for i := 0; i < 10; i++ {
		msgs = append(msgs, message.Message{
			Role:  message.User,
			Parts: []message.ContentPart{message.TextContent{Text: "Message"}},
		})
	}

	result, applied := c.Compact(msgs, 80000)

	require.True(t, applied)
	// Should contain summary + context bridge + recent messages
	assert.Contains(t, result[0].Parts[0].(message.TextContent).Text, "Summary")
}

func TestCompactor_Aggressive(t *testing.T) {
	config := DefaultCompactionConfig(100000)
	c := NewCompactor(config)

	msgs := make([]message.Message, 50)
	for i := 0; i < 50; i++ {
		msgs[i] = message.Message{
			Role:  message.User,
			Parts: []message.ContentPart{message.TextContent{Text: "Message"}},
		}
	}

	result, applied := c.Compact(msgs, 90000) // 90% triggers aggressive

	require.True(t, applied)
	// Aggressive keeps only ~5-6 messages
	assert.LessOrEqual(t, len(result), 6)
}

func TestCompactor_NoCompaction(t *testing.T) {
	c := NewCompactor(DefaultCompactionConfig(100000))

	msgs := []message.Message{
		{Role: message.User, Parts: []message.ContentPart{message.TextContent{Text: "Hello"}}},
	}

	result, applied := c.Compact(msgs, 20000) // 20% - no compaction needed

	assert.False(t, applied)
	assert.Equal(t, msgs, result)
}

func TestCompactor_ZeroContextWindow(t *testing.T) {
	c := NewCompactor(CompactionConfig{ContextWindow: 0})

	msgs := []message.Message{
		{Role: message.User, Parts: []message.ContentPart{message.TextContent{Text: "Hello"}}},
	}

	result, applied := c.Compact(msgs, 50000)

	assert.False(t, applied)
	assert.Equal(t, msgs, result)
}

func TestEstimateTextTokens(t *testing.T) {
	tests := []struct {
		text      string
		minTokens int
		maxTokens int
	}{
		{"", 0, 0},
		{"hello", 1, 3},
		{strings.Repeat("a", 400), 90, 130}, // ~100 tokens
	}

	for _, tt := range tests {
		t.Run(tt.text[:min(10, len(tt.text))], func(t *testing.T) {
			tokens := estimateTextTokens(tt.text)
			assert.GreaterOrEqual(t, tokens, tt.minTokens)
			assert.LessOrEqual(t, tokens, tt.maxTokens)
		})
	}
}

func TestTruncateToTokens(t *testing.T) {
	longText := strings.Repeat("word ", 200) // 1000 chars

	result := truncateToTokens(longText, 50) // ~200 chars

	assert.Less(t, len(result), 300)
	assert.Greater(t, len(result), 100)
}
