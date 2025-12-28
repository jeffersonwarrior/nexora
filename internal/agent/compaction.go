// Package agent provides inline compaction for managing context window usage.
package agent

import (
	"log/slog"
	"strings"

	"github.com/nexora/nexora/internal/message"
)

// CompactionLevel defines the aggressiveness of message compaction.
type CompactionLevel int

const (
	// CompactionNone applies no compaction.
	CompactionNone CompactionLevel = iota
	// CompactionTruncate truncates large tool outputs while preserving structure.
	CompactionTruncate
	// CompactionDropToolResults removes tool result content, keeps tool calls.
	CompactionDropToolResults
	// CompactionKeepRecent keeps only recent messages plus summary.
	CompactionKeepRecent
	// CompactionAggressive keeps only summary and last few messages.
	CompactionAggressive
)

// CompactionConfig configures inline compaction behavior.
type CompactionConfig struct {
	// ContextWindow is the total context window size in tokens.
	ContextWindow int64
	// TargetUsage is the target context usage percentage (0.0-1.0).
	// Default: 0.7 (70% of context window).
	TargetUsage float64
	// MaxToolOutputTokens is the maximum tokens per tool output before truncation.
	// Default: 2000.
	MaxToolOutputTokens int
	// RecentMessageCount is the number of recent messages to always keep.
	// Default: 10.
	RecentMessageCount int
	// PreserveSystemMessages keeps all system messages regardless of level.
	PreserveSystemMessages bool
}

// DefaultCompactionConfig returns sensible defaults.
func DefaultCompactionConfig(contextWindow int64) CompactionConfig {
	return CompactionConfig{
		ContextWindow:          contextWindow,
		TargetUsage:            0.7,
		MaxToolOutputTokens:    2000,
		RecentMessageCount:     10,
		PreserveSystemMessages: true,
	}
}

// Compactor performs inline message compaction.
type Compactor struct {
	config CompactionConfig
}

// NewCompactor creates a new compactor with the given config.
func NewCompactor(config CompactionConfig) *Compactor {
	if config.TargetUsage == 0 {
		config.TargetUsage = 0.7
	}
	if config.MaxToolOutputTokens == 0 {
		config.MaxToolOutputTokens = 2000
	}
	if config.RecentMessageCount == 0 {
		config.RecentMessageCount = 10
	}
	return &Compactor{config: config}
}

// EstimateTokens estimates token count for a message.
// Uses ~4 characters per token approximation with adjustments for structure.
func (c *Compactor) EstimateTokens(msg message.Message) int {
	var tokens int

	for _, part := range msg.Parts {
		switch p := part.(type) {
		case message.TextContent:
			tokens += estimateTextTokens(p.Text)
		case message.ReasoningContent:
			tokens += estimateTextTokens(p.Thinking)
		case message.ToolCall:
			tokens += 20 // Overhead for tool call structure
			tokens += estimateTextTokens(p.Name)
			tokens += estimateTextTokens(p.Input)
		case message.ToolResult:
			tokens += 20 // Overhead for tool result structure
			tokens += estimateTextTokens(p.Content)
			tokens += estimateTextTokens(p.Data)
		case message.BinaryContent:
			// Binary content is typically base64 encoded
			tokens += len(p.Data) / 3 // Base64 expansion factor
		case message.ImageURLContent:
			tokens += 100 // URL overhead
		case message.Finish:
			tokens += 10
		}
	}

	return tokens
}

// EstimateTotalTokens estimates total tokens for a message slice.
func (c *Compactor) EstimateTotalTokens(msgs []message.Message) int {
	var total int
	for _, msg := range msgs {
		total += c.EstimateTokens(msg)
	}
	return total
}

// DetermineLevel determines the appropriate compaction level based on usage.
func (c *Compactor) DetermineLevel(currentTokens int64) CompactionLevel {
	if c.config.ContextWindow == 0 {
		return CompactionNone
	}

	usage := float64(currentTokens) / float64(c.config.ContextWindow)

	switch {
	case usage < 0.5:
		return CompactionNone
	case usage < 0.65:
		return CompactionTruncate
	case usage < 0.75:
		return CompactionDropToolResults
	case usage < 0.85:
		return CompactionKeepRecent
	default:
		return CompactionAggressive
	}
}

// Compact applies inline compaction to messages based on current token usage.
// Returns compacted messages and whether compaction was applied.
func (c *Compactor) Compact(msgs []message.Message, currentTokens int64) ([]message.Message, bool) {
	level := c.DetermineLevel(currentTokens)
	if level == CompactionNone {
		return msgs, false
	}

	slog.Debug("applying inline compaction",
		"level", level,
		"current_tokens", currentTokens,
		"context_window", c.config.ContextWindow,
		"message_count", len(msgs),
	)

	var result []message.Message
	switch level {
	case CompactionTruncate:
		result = c.truncateToolOutputs(msgs)
	case CompactionDropToolResults:
		result = c.dropToolResults(msgs)
	case CompactionKeepRecent:
		result = c.keepRecent(msgs)
	case CompactionAggressive:
		result = c.aggressive(msgs)
	default:
		return msgs, false
	}

	slog.Info("inline compaction applied",
		"level", level,
		"original_count", len(msgs),
		"compacted_count", len(result),
		"estimated_reduction", c.EstimateTotalTokens(msgs)-c.EstimateTotalTokens(result),
	)

	return result, true
}

// truncateToolOutputs truncates large tool outputs while preserving structure.
func (c *Compactor) truncateToolOutputs(msgs []message.Message) []message.Message {
	result := make([]message.Message, 0, len(msgs))

	for _, msg := range msgs {
		if msg.Role != message.Tool {
			result = append(result, msg)
			continue
		}

		// Clone and truncate tool results
		newParts := make([]message.ContentPart, 0, len(msg.Parts))
		for _, part := range msg.Parts {
			if tr, ok := part.(message.ToolResult); ok {
				tokens := estimateTextTokens(tr.Content)
				if tokens > c.config.MaxToolOutputTokens {
					tr.Content = truncateToTokens(tr.Content, c.config.MaxToolOutputTokens)
					tr.Content += "\n... [output truncated for context management]"
				}
				newParts = append(newParts, tr)
			} else {
				newParts = append(newParts, part)
			}
		}

		newMsg := msg
		newMsg.Parts = newParts
		result = append(result, newMsg)
	}

	return result
}

// dropToolResults removes tool result content but keeps tool calls as context.
func (c *Compactor) dropToolResults(msgs []message.Message) []message.Message {
	// First apply truncation
	msgs = c.truncateToolOutputs(msgs)

	result := make([]message.Message, 0, len(msgs))
	recentIdx := len(msgs) - c.config.RecentMessageCount
	if recentIdx < 0 {
		recentIdx = 0
	}

	for i, msg := range msgs {
		// Keep recent messages fully intact
		if i >= recentIdx {
			result = append(result, msg)
			continue
		}

		// Keep summary messages
		if msg.IsSummaryMessage {
			result = append(result, msg)
			continue
		}

		// For older tool messages, replace with placeholder
		if msg.Role == message.Tool {
			newParts := make([]message.ContentPart, 0, len(msg.Parts))
			for _, part := range msg.Parts {
				if tr, ok := part.(message.ToolResult); ok {
					tr.Content = "[tool output removed for context management]"
					tr.Data = ""
					newParts = append(newParts, tr)
				} else {
					newParts = append(newParts, part)
				}
			}
			newMsg := msg
			newMsg.Parts = newParts
			result = append(result, newMsg)
			continue
		}

		result = append(result, msg)
	}

	return result
}

// keepRecent keeps only recent messages plus summary if present.
func (c *Compactor) keepRecent(msgs []message.Message) []message.Message {
	result := make([]message.Message, 0)

	// Find summary message if exists
	var summaryIdx int = -1
	for i, msg := range msgs {
		if msg.IsSummaryMessage {
			summaryIdx = i
			break
		}
	}

	// Calculate how many recent messages to keep
	recentCount := c.config.RecentMessageCount * 2 // Double for user/assistant pairs
	recentIdx := len(msgs) - recentCount
	if recentIdx < 0 {
		recentIdx = 0
	}

	// If summary exists and is before recentIdx, include it
	if summaryIdx >= 0 && summaryIdx < recentIdx {
		// Add summary
		result = append(result, msgs[summaryIdx])
		// Add context bridge
		result = append(result, message.Message{
			Role: message.User,
			Parts: []message.ContentPart{
				message.TextContent{Text: "[Earlier conversation compacted for context management. Continuing from summary.]"},
			},
		})
	}

	// Add recent messages
	for i := recentIdx; i < len(msgs); i++ {
		// Skip the summary if we already added it
		if i == summaryIdx {
			continue
		}
		result = append(result, msgs[i])
	}

	return result
}

// aggressive keeps only essential messages.
func (c *Compactor) aggressive(msgs []message.Message) []message.Message {
	result := make([]message.Message, 0)

	// Find summary message
	for _, msg := range msgs {
		if msg.IsSummaryMessage {
			result = append(result, msg)
			break
		}
	}

	// If no summary, create a placeholder
	if len(result) == 0 && len(msgs) > 0 {
		result = append(result, message.Message{
			Role: message.User,
			Parts: []message.ContentPart{
				message.TextContent{Text: "[Session history compacted due to context limits. Please refer to recent context below.]"},
			},
		})
	}

	// Keep only last few messages (3-5)
	keepCount := 5
	startIdx := len(msgs) - keepCount
	if startIdx < 0 {
		startIdx = 0
	}

	for i := startIdx; i < len(msgs); i++ {
		if !msgs[i].IsSummaryMessage {
			result = append(result, msgs[i])
		}
	}

	return result
}

// estimateTextTokens estimates tokens for text content.
func estimateTextTokens(text string) int {
	if text == "" {
		return 0
	}
	// Approximate: 4 characters per token for English
	// Add 10% overhead for tokenizer boundaries
	return int(float64(len(text)) / 4.0 * 1.1)
}

// truncateToTokens truncates text to approximately the given token count.
func truncateToTokens(text string, maxTokens int) string {
	// Approximate 4 characters per token
	maxChars := maxTokens * 4
	if len(text) <= maxChars {
		return text
	}

	// Try to cut at a word boundary
	text = text[:maxChars]
	if lastSpace := strings.LastIndex(text, " "); lastSpace > maxChars*3/4 {
		text = text[:lastSpace]
	}

	return text
}
