# Context Management V2 - Implementation Plan

## Overview

Enhanced context management system with three core components:
1. **Fast Summarizer** - Auto-detect fastest provider for summarization
2. **Reference System** - Unified `nexora://` URI scheme for cross-project references
3. **Smart Pruner** - Intelligent context condensation with recall capability

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Context Manager                              │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  ┌──────────────┐    ┌──────────────┐    ┌──────────────────────┐  │
│  │    Pruner    │───▶│   Archive    │◀───│    Recall Tool       │  │
│  │              │    │              │    │                      │  │
│  │ - Classify   │    │ - Store full │    │ - Fetch by ref_id    │  │
│  │ - Condense   │    │   content    │    │ - Cross-project      │  │
│  │ - Reference  │    │ - Generate   │    │ - Token budgeting    │  │
│  └──────────────┘    │   ref IDs    │    └──────────────────────┘  │
│         │            └──────────────┘                               │
│         ▼                   │                                       │
│  ┌──────────────┐           │                                       │
│  │ Summarizer   │◀──────────┘                                       │
│  │              │                                                    │
│  │ - Auto-detect│    ┌──────────────┐                               │
│  │   fast model │───▶│  Provider    │                               │
│  │ - Stream     │    │  Detection   │                               │
│  └──────────────┘    └──────────────┘                               │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
```

---

## 1. Unified Reference ID System

### URI Format

```
nexora://<project>/<session>/<message>[:<location>]
```

| Component | Format | Description |
|-----------|--------|-------------|
| `project` | directory name | Project directory name, `_` for current |
| `session` | 8-char UUID prefix | Session ID (display), full UUID stored |
| `message` | `msg-<8-char>` | Message ID prefix |
| `location` | optional | `L<start>-<end>`, `match-<n>`, `tool-<n>` |

### Examples

```
nexora://nexora-cli/abc12345/msg-def67890              # Full reference
nexora://nexora-cli/abc12345/msg-def67890:L50-100     # Lines 50-100
nexora://nexora-cli/abc12345/msg-def67890:match-3     # 3rd grep match
nexora://_/current/msg-def67890                        # Current session shorthand
msg-def67890                                           # Shortest form (current session)
```

### Resolution Priority

1. Full URI → direct lookup
2. `msg-<id>` → search current session
3. `nexora://_/current/...` → resolve `_` to current project, `current` to active session

---

## 2. Database Schema

### New Migration: `20251217000000_add_context_archive.sql`

```sql
-- +goose Up
CREATE TABLE context_archive (
    id TEXT PRIMARY KEY,
    ref_id TEXT UNIQUE NOT NULL,
    project TEXT NOT NULL,
    session_id TEXT NOT NULL,
    message_id TEXT NOT NULL,
    location TEXT,
    
    content_type TEXT NOT NULL,
    summary TEXT NOT NULL,
    full_content TEXT NOT NULL,
    
    file_paths TEXT,
    tool_name TEXT,
    token_count INTEGER,
    
    original_created_at INTEGER,
    archived_at INTEGER DEFAULT (strftime('%s', 'now')),
    last_accessed_at INTEGER,
    access_count INTEGER DEFAULT 0
);

CREATE INDEX idx_context_archive_ref ON context_archive(ref_id);
CREATE INDEX idx_context_archive_session ON context_archive(session_id);
CREATE INDEX idx_context_archive_project ON context_archive(project);
CREATE INDEX idx_context_archive_message ON context_archive(message_id);

-- +goose Down
DROP INDEX IF EXISTS idx_context_archive_message;
DROP INDEX IF EXISTS idx_context_archive_project;
DROP INDEX IF EXISTS idx_context_archive_session;
DROP INDEX IF EXISTS idx_context_archive_ref;
DROP TABLE IF EXISTS context_archive;
```

---

## 3. Fast Summarizer

### Config Options

Add to `internal/config/config.go` Options struct:

```go
type Options struct {
    // ... existing fields
    
    // Summarization provider (auto-detected if empty)
    // Prefers fastest available: cerebras > xai > current provider
    SummarizationProvider string `json:"summarization_provider,omitempty" jsonschema:"description=Provider for summarization (auto-detected if empty)"`
    
    // Summarization model (uses provider's fast model if empty)
    SummarizationModel string `json:"summarization_model,omitempty" jsonschema:"description=Model for summarization (auto-detected if empty)"`
}
```

### Auto-Detection Logic

```go
// internal/agent/summarizer.go

type SummarizerConfig struct {
    Provider string
    Model    string
}

// Fast models ranked by speed (tokens/sec approximate)
var fastModels = []struct {
    Provider string
    Model    string
    Speed    int // tok/s approximate
}{
    {"cerebras", "llama3.1-8b", 2000},
    {"cerebras", "llama-3.3-70b", 1500},
    {"xai", "grok-3-mini", 1200},
    {"xai", "grok-4-1-fast", 1000},
    {"groq", "llama-3.1-8b-instant", 800},
}

func DetectFastestSummarizer(cfg *config.Config) SummarizerConfig {
    // 1. Check explicit config
    if cfg.Options.SummarizationProvider != "" {
        return SummarizerConfig{
            Provider: cfg.Options.SummarizationProvider,
            Model:    cfg.Options.SummarizationModel,
        }
    }
    
    // 2. Find fastest available provider
    for _, fm := range fastModels {
        if provider, ok := cfg.Providers.Get(fm.Provider); ok {
            if hasValidAPIKey(provider) {
                return SummarizerConfig{
                    Provider: fm.Provider,
                    Model:    fm.Model,
                }
            }
        }
    }
    
    // 3. Fallback to current provider's small model
    return SummarizerConfig{} // Empty means use current behavior
}
```

---

## 4. Recall Tool

### Tool Definition

```go
// internal/agent/tools/recall.go

type RecallTool struct {
    archive *ContextArchive
    db      *sql.DB
}

type RecallParams struct {
    RefID     string `json:"ref_id" jsonschema:"required,description=Reference ID (nexora:// URI or short msg-xxx form)"`
    Lines     string `json:"lines,omitempty" jsonschema:"description=Line range to extract (e.g. 10-50)"`
    Search    string `json:"search,omitempty" jsonschema:"description=Search pattern within content"`
    MaxTokens int    `json:"max_tokens,omitempty" jsonschema:"description=Max tokens to return (default 2000)"`
}

func (t *RecallTool) Name() string { return "recall" }

func (t *RecallTool) Description() string {
    return `Retrieve previously viewed content from context archive.
Use when you see references like [recall:msg-xxx] or nexora:// URIs in condensed context.
Supports cross-session and cross-project recall.`
}

func (t *RecallTool) Run(ctx context.Context, params RecallParams) (*RecallResult, error) {
    // 1. Resolve ref_id to full URI
    fullRef, err := t.archive.ResolveRef(ctx, params.RefID)
    if err != nil {
        return nil, fmt.Errorf("reference not found: %s", params.RefID)
    }
    
    // 2. Fetch from archive
    entry, err := t.archive.Get(ctx, fullRef)
    if err != nil {
        return nil, err
    }
    
    // 3. Apply filters (lines, search)
    content := entry.FullContent
    if params.Lines != "" {
        content = extractLines(content, params.Lines)
    }
    if params.Search != "" {
        content = filterBySearch(content, params.Search)
    }
    
    // 4. Token budget
    maxTokens := cmp.Or(params.MaxTokens, 2000)
    content, truncated := truncateToTokens(content, maxTokens)
    
    // 5. Update access stats
    t.archive.RecordAccess(ctx, fullRef)
    
    return &RecallResult{
        RefID:       fullRef,
        ContentType: entry.ContentType,
        Summary:     entry.Summary,
        Content:     content,
        FilePaths:   entry.FilePaths,
        Truncated:   truncated,
        TokenCount:  countTokens(content),
    }, nil
}
```

---

## 5. Smart Pruner

### Classification Rules

```go
// internal/agent/context_pruner.go

type MessageClass int

const (
    ClassKeep     MessageClass = iota // Keep in full
    ClassCondense                      // Condense to reference
    ClassRemove                        // Remove entirely
)

type PruneConfig struct {
    KeepLastN        int  // Keep last N messages intact (default 10)
    CondenseAfterN   int  // Start condensing after N messages (default 5)
    RemoveErrorsAfter int // Remove error results after N messages (default 5)
}

func (p *Pruner) Classify(msg *message.Message, position int, total int) MessageClass {
    age := total - position // 0 = newest
    
    // Always keep recent messages
    if age < p.config.KeepLastN {
        return ClassKeep
    }
    
    // Remove old errors
    if msg.IsErrorResult() && age > p.config.RemoveErrorsAfter {
        return ClassRemove
    }
    
    // Remove superseded content (file viewed then edited)
    if p.isSuperseded(msg) {
        return ClassRemove
    }
    
    // Condense tool results
    if msg.Role == message.Tool {
        return ClassCondense
    }
    
    // Condense old assistant messages with large content
    if msg.Role == message.Assistant && msg.TokenCount() > 500 && age > p.config.CondenseAfterN {
        return ClassCondense
    }
    
    return ClassKeep
}
```

### Condensation Format

```go
func (p *Pruner) Condense(msg *message.Message, archive *ContextArchive) (*message.Message, error) {
    // 1. Archive full content
    ref, err := archive.Store(ctx, ArchiveEntry{
        MessageID:   msg.ID,
        ContentType: p.classifyContentType(msg),
        Summary:     p.generateSummary(msg),
        FullContent: msg.FullText(),
        FilePaths:   msg.ExtractFilePaths(),
        ToolName:    msg.ToolName(),
        TokenCount:  msg.TokenCount(),
    })
    if err != nil {
        return nil, err
    }
    
    // 2. Create condensed message
    condensed := &message.Message{
        ID:        msg.ID,
        Role:      msg.Role,
        SessionID: msg.SessionID,
        CreatedAt: msg.CreatedAt,
    }
    
    // 3. Generate condensed content based on type
    var summary string
    switch p.classifyContentType(msg) {
    case "file_view":
        summary = fmt.Sprintf("📎 Viewed %s - %s [recall:%s]", 
            msg.FilePath(), p.generateSummary(msg), ref)
    case "grep":
        summary = fmt.Sprintf("🔍 Searched '%s' in %s - %d matches [recall:%s]",
            msg.GrepPattern(), msg.GrepPath(), msg.MatchCount(), ref)
    case "bash":
        summary = fmt.Sprintf("⚡ Ran: %s - %s [recall:%s]",
            truncate(msg.Command(), 50), msg.ExitStatus(), ref)
    case "edit":
        summary = fmt.Sprintf("✏️ Edited %s [recall:%s]",
            msg.FilePath(), ref)
    default:
        summary = fmt.Sprintf("📝 %s [recall:%s]",
            truncate(p.generateSummary(msg), 80), ref)
    }
    
    condensed.SetText(summary)
    return condensed, nil
}
```

---

## 6. Integration Flow

### Summarization Trigger (modified)

```go
// In agent.go StopCondition

func (a *sessionAgent) shouldPruneAndSummarize(session *Session) bool {
    cw := int64(a.largeModel.CatwalkCfg.ContextWindow)
    tokens := session.CompletionTokens + session.PromptTokens
    remaining := cw - tokens
    
    var threshold int64
    if cw > 200_000 {
        threshold = 20_000
    } else {
        threshold = int64(float64(cw) * 0.2)
    }
    
    return remaining <= threshold && !a.disableAutoSummarize
}

// New pre-summarization hook
func (a *sessionAgent) prepareForSummarization(ctx context.Context, sessionID string) error {
    // 1. Prune and condense old messages
    pruned, err := a.pruner.PruneSession(ctx, sessionID)
    if err != nil {
        return err
    }
    
    slog.Info("context pruned", 
        "session_id", sessionID,
        "condensed", pruned.Condensed,
        "removed", pruned.Removed,
        "tokens_saved", pruned.TokensSaved)
    
    // 2. Now summarize with fast model
    return a.Summarize(ctx, sessionID, a.providerOptions)
}
```

---

## 7. Implementation Phases

### Phase 1: Fast Summarizer (2-3 hours)
Files to create/modify:
- [ ] `internal/agent/summarizer.go` - new file
- [ ] `internal/config/config.go` - add Options fields
- [ ] `internal/agent/agent.go` - update Summarize() to use detected provider

### Phase 2: Reference System + Archive (3-4 hours)
Files to create/modify:
- [ ] `internal/db/migrations/20251217000000_add_context_archive.sql`
- [ ] `internal/agent/context_archive.go` - new file
- [ ] `internal/agent/tools/recall.go` - new file
- [ ] `internal/agent/coordinator.go` - register recall tool

### Phase 3: Smart Pruner (4-5 hours)
Files to create/modify:
- [ ] `internal/agent/context_pruner.go` - new file
- [ ] `internal/agent/agent.go` - integrate pruner with summarization
- [ ] `internal/message/message.go` - add helper methods (TokenCount, FilePath, etc.)

---

## 8. Testing Strategy

### Unit Tests
- `summarizer_test.go` - provider detection, fallback logic
- `context_archive_test.go` - store/retrieve, ref resolution
- `recall_test.go` - tool execution, cross-project
- `context_pruner_test.go` - classification, condensation

### Integration Tests
- Full flow: accumulate context → trigger prune → summarize → recall
- Cross-project recall with multiple databases
- Token budget enforcement

---

## 9. Configuration Example

```json
{
  "options": {
    "summarization_provider": "cerebras",
    "summarization_model": "llama3.1-8b",
    "disable_auto_summarize": false
  }
}
```

Auto-detection (no config needed):
```json
{
  "options": {}
}
```

---

## 10. Future Enhancements

1. **Semantic search over archive** - Find related past context
2. **Archive compression** - Compress old entries to save disk
3. **Archive export** - Export session with full archive for sharing
4. **Selective recall** - "Recall all file views from last hour"
5. **Archive browser** - TUI component to browse archived content
