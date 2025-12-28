// Package agent provides background compaction during user idle time.
package agent

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/nexora/nexora/internal/message"
	"github.com/nexora/nexora/internal/session"
)

// BackgroundCompactor performs silent compaction during user idle time.
// It watches for idle signals and compacts session history in the background.
type BackgroundCompactor struct {
	mu sync.RWMutex

	// Dependencies
	sessions session.Service
	messages message.Service

	// Configuration
	idleThreshold   time.Duration // How long before triggering compaction
	compactInterval time.Duration // Minimum time between compactions
	maxMessages     int           // Max messages before compaction is useful

	// State
	lastCompaction map[string]time.Time // sessionID -> last compaction time
	compacting     map[string]bool      // sessionID -> currently compacting

	// Compacted message cache (session -> compacted messages)
	cache map[string]*CompactedHistory

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
}

// CompactedHistory holds pre-compacted message history for a session.
type CompactedHistory struct {
	Messages    []message.Message
	CompactedAt time.Time
	TokensSaved int
	OriginalLen int
}

// BackgroundCompactorConfig configures the background compactor.
type BackgroundCompactorConfig struct {
	// IdleThreshold is how long user must be idle before compaction.
	// Default: 5 seconds.
	IdleThreshold time.Duration

	// CompactInterval is minimum time between compactions for a session.
	// Default: 30 seconds.
	CompactInterval time.Duration

	// MaxMessages is the message count threshold for useful compaction.
	// Default: 20.
	MaxMessages int
}

// DefaultBackgroundCompactorConfig returns sensible defaults.
func DefaultBackgroundCompactorConfig() BackgroundCompactorConfig {
	return BackgroundCompactorConfig{
		IdleThreshold:   5 * time.Second,
		CompactInterval: 30 * time.Second,
		MaxMessages:     20,
	}
}

// NewBackgroundCompactor creates a new background compactor.
func NewBackgroundCompactor(
	sessions session.Service,
	messages message.Service,
	cfg BackgroundCompactorConfig,
) *BackgroundCompactor {
	if cfg.IdleThreshold == 0 {
		cfg.IdleThreshold = 5 * time.Second
	}
	if cfg.CompactInterval == 0 {
		cfg.CompactInterval = 30 * time.Second
	}
	if cfg.MaxMessages == 0 {
		cfg.MaxMessages = 20
	}

	return &BackgroundCompactor{
		sessions:        sessions,
		messages:        messages,
		idleThreshold:   cfg.IdleThreshold,
		compactInterval: cfg.CompactInterval,
		maxMessages:     cfg.MaxMessages,
		lastCompaction:  make(map[string]time.Time),
		compacting:      make(map[string]bool),
		cache:           make(map[string]*CompactedHistory),
	}
}

// Start begins the background compactor.
func (bc *BackgroundCompactor) Start(ctx context.Context) {
	bc.mu.Lock()
	if bc.ctx != nil {
		bc.mu.Unlock()
		return
	}
	bc.ctx, bc.cancel = context.WithCancel(ctx)
	bc.mu.Unlock()

	slog.Debug("background compactor started")
}

// Stop stops the background compactor.
func (bc *BackgroundCompactor) Stop() {
	bc.mu.Lock()
	if bc.cancel != nil {
		bc.cancel()
	}
	bc.mu.Unlock()
	slog.Debug("background compactor stopped")
}

// OnIdle is called when the user has been idle for the threshold duration.
// This triggers background compaction for the given session.
func (bc *BackgroundCompactor) OnIdle(sessionID string) {
	bc.mu.Lock()
	ctx := bc.ctx
	bc.mu.Unlock()

	if ctx == nil {
		return
	}

	go bc.tryCompact(ctx, sessionID)
}

// GetCompactedHistory returns cached compacted history if available.
// Returns nil if no compaction has been performed or cache is stale.
func (bc *BackgroundCompactor) GetCompactedHistory(sessionID string) *CompactedHistory {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.cache[sessionID]
}

// InvalidateCache clears the compaction cache for a session.
// Call this when new messages are added.
func (bc *BackgroundCompactor) InvalidateCache(sessionID string) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	delete(bc.cache, sessionID)
}

// tryCompact attempts to compact a session's history.
func (bc *BackgroundCompactor) tryCompact(ctx context.Context, sessionID string) {
	bc.mu.Lock()

	// Check if already compacting
	if bc.compacting[sessionID] {
		bc.mu.Unlock()
		return
	}

	// Check if compacted recently
	if last, ok := bc.lastCompaction[sessionID]; ok {
		if time.Since(last) < bc.compactInterval {
			bc.mu.Unlock()
			return
		}
	}

	bc.compacting[sessionID] = true
	bc.mu.Unlock()

	defer func() {
		bc.mu.Lock()
		delete(bc.compacting, sessionID)
		bc.mu.Unlock()
	}()

	bc.performCompaction(ctx, sessionID)
}

// performCompaction does the actual compaction work.
func (bc *BackgroundCompactor) performCompaction(ctx context.Context, sessionID string) {
	// Get session
	sess, err := bc.sessions.Get(ctx, sessionID)
	if err != nil {
		slog.Debug("background compaction: failed to get session", "error", err)
		return
	}

	// Get messages
	msgs, err := bc.messages.List(ctx, sessionID)
	if err != nil {
		slog.Debug("background compaction: failed to list messages", "error", err)
		return
	}

	// Check if compaction is useful
	if len(msgs) < bc.maxMessages {
		slog.Debug("background compaction: not enough messages",
			"session", sessionID,
			"count", len(msgs),
			"threshold", bc.maxMessages,
		)
		return
	}

	// Apply local compaction (no API call - just structural optimization)
	compacted := bc.compactLocally(msgs, sess)

	// Cache the result
	tokensSaved := estimateSavings(msgs, compacted)
	bc.mu.Lock()
	bc.cache[sessionID] = &CompactedHistory{
		Messages:    compacted,
		CompactedAt: time.Now(),
		TokensSaved: tokensSaved,
		OriginalLen: len(msgs),
	}
	bc.lastCompaction[sessionID] = time.Now()
	bc.mu.Unlock()

	// Update session to reflect compacted token count
	// This provides immediate UI feedback for the context window display
	if tokensSaved > 0 && sess.PromptTokens > int64(tokensSaved) {
		sess.PromptTokens -= int64(tokensSaved)
		if _, err := bc.sessions.Save(ctx, sess); err != nil {
			slog.Warn("background compaction: failed to update session tokens", "error", err)
		} else {
			slog.Debug("background compaction: updated session tokens",
				"session", sessionID,
				"tokens_reduced_by", tokensSaved,
			)
		}
	}

	slog.Info("background compaction complete",
		"session", sessionID,
		"original", len(msgs),
		"compacted", len(compacted),
		"estimated_savings", tokensSaved,
	)
}

// compactLocally performs structural compaction without API calls.
// This is fast and can run during brief idle periods.
func (bc *BackgroundCompactor) compactLocally(msgs []message.Message, sess session.Session) []message.Message {
	if len(msgs) == 0 {
		return msgs
	}

	result := make([]message.Message, 0, len(msgs))

	// Find summary message if exists
	var summaryIdx int = -1
	for i, msg := range msgs {
		if msg.IsSummaryMessage {
			summaryIdx = i
			break
		}
	}

	// If summary exists, start from there
	startIdx := 0
	if summaryIdx >= 0 {
		result = append(result, msgs[summaryIdx])
		startIdx = summaryIdx + 1
	}

	// Calculate how many messages to keep fully vs compact
	remaining := msgs[startIdx:]
	keepFull := 10 // Keep last 10 messages fully intact

	if len(remaining) <= keepFull {
		// Not enough to compact, keep all
		return append(result, remaining...)
	}

	// Compact older messages (truncate tool outputs)
	compactCount := len(remaining) - keepFull
	for i := 0; i < compactCount; i++ {
		msg := remaining[i]
		compactedMsg := compactMessage(msg)
		result = append(result, compactedMsg)
	}

	// Keep recent messages fully
	result = append(result, remaining[compactCount:]...)

	return result
}

// compactMessage reduces a single message's size.
func compactMessage(msg message.Message) message.Message {
	newParts := make([]message.ContentPart, 0, len(msg.Parts))

	for _, part := range msg.Parts {
		switch p := part.(type) {
		case message.ToolResult:
			// Truncate large tool outputs
			if len(p.Content) > 500 {
				p.Content = p.Content[:500] + "\n... [truncated]"
			}
			p.Data = "" // Clear binary data
			newParts = append(newParts, p)
		case message.ReasoningContent:
			// Keep reasoning but truncate if very long
			if len(p.Thinking) > 1000 {
				p.Thinking = p.Thinking[:1000] + "\n... [truncated]"
			}
			newParts = append(newParts, p)
		default:
			newParts = append(newParts, part)
		}
	}

	result := msg
	result.Parts = newParts
	return result
}

// estimateSavings estimates token savings from compaction.
func estimateSavings(original, compacted []message.Message) int {
	origTokens := 0
	for _, m := range original {
		origTokens += estimateMessageTokens(m)
	}

	compTokens := 0
	for _, m := range compacted {
		compTokens += estimateMessageTokens(m)
	}

	return origTokens - compTokens
}

// estimateMessageTokens estimates tokens for a message.
func estimateMessageTokens(msg message.Message) int {
	tokens := 0
	for _, part := range msg.Parts {
		switch p := part.(type) {
		case message.TextContent:
			tokens += len(p.Text) / 4
		case message.ReasoningContent:
			tokens += len(p.Thinking) / 4
		case message.ToolCall:
			tokens += 20 + len(p.Input)/4
		case message.ToolResult:
			tokens += 20 + len(p.Content)/4 + len(p.Data)/4
		}
	}
	return tokens
}
