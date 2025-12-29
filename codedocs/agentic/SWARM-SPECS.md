# Nexora v0.29.3 Swarm Development Specifications

**Phase:** v0.29.3 Feature Implementation
**Methodology:** Test-Driven Development (TDD)
**Target Coverage:** 40%+
**Created:** 2025-12-27

---

## Orchestration Architecture

```
OVERSEER (Claude Opus)
    |
    +---> SENIOR DEVELOPER (Orchestrator)
              |
              +---> Agent 1: Prompt Library (#3)
              |
              +---> Agent 2: About Command (#4)
              |
              +---> Agent 3: Test Coverage (#5)
              |
              +---> Agent 4: Task Graph (#6)
              |
              +---> Agent 5: Checkpoint System (#7)
```

---

## Feature 1: Prompt Library (#3)

### Branch
`feature/prompt-library`

### Files to Create
- `internal/prompts/service.go`
- `internal/prompts/service_test.go`
- `internal/prompts/models.go`
- `internal/db/queries/prompts.sql`
- `internal/tui/components/prompts/browser.go`
- `internal/tui/components/prompts/browser_test.go`
- `internal/cmd/prompts_cmd.go`
- `internal/cmd/prompts_cmd_test.go`

### Test Specification

```go
// internal/prompts/service_test.go

package prompts

import (
    "context"
    "testing"
    "github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
    t.Parallel()
    // Test creating a new prompt
    // Verify ID generation, timestamps, content_hash
    // Verify database insertion
}

func TestService_Get(t *testing.T) {
    t.Parallel()
    // Test retrieving prompt by ID
    // Test not found case
}

func TestService_List(t *testing.T) {
    t.Parallel()
    // Test listing prompts with pagination
    // Test category filtering
    // Test sorting by rating, usage, date
}

func TestService_Update(t *testing.T) {
    t.Parallel()
    // Test updating prompt content
    // Verify updated_at changes
    // Verify content_hash updates
}

func TestService_Delete(t *testing.T) {
    t.Parallel()
    // Test soft delete
    // Test hard delete
}

func TestService_Search(t *testing.T) {
    t.Parallel()
    // Test FTS5 full-text search
    // Test search by title
    // Test search by content
    // Test search by tags
    // Test ranking/relevance
}

func TestService_IncrementUsage(t *testing.T) {
    t.Parallel()
    // Test usage_count increment
    // Test last_used_at update
}

func TestService_UpdateRating(t *testing.T) {
    t.Parallel()
    // Test rating calculation
    // Test votes increment
}

func TestService_GetByCategory(t *testing.T) {
    t.Parallel()
    // Test category filtering
    // Test subcategory filtering
}
```

### Database Schema (Already Exists)
```sql
-- Migration: 20251225000007_create_prompt_library.sql
-- Table: prompt_library
-- FTS: prompt_library_fts
-- Indexes: category, rating, usage, success, content_hash, tags
```

### Implementation Notes
- Use sqlc for type-safe queries
- Content hash: SHA256 of content for deduplication
- FTS5 triggers already exist for sync
- Service should implement pubsub pattern for events

---

## Feature 2: About Command (#4)

### Branch
`feature/about-command`

### Files to Create
- `internal/cmd/about.go`
- `internal/cmd/about_test.go`

### Test Specification

```go
// internal/cmd/about_test.go

package cmd

import (
    "bytes"
    "testing"
    "github.com/stretchr/testify/require"
)

func TestAboutCommand_Output(t *testing.T) {
    t.Parallel()
    // Test command executes without error
    // Test output contains version
    // Test output contains platform info
}

func TestAboutCommand_Version(t *testing.T) {
    t.Parallel()
    // Test version string matches internal/version/version.go
}

func TestAboutCommand_Platform(t *testing.T) {
    t.Parallel()
    // Test runtime.GOOS displayed
    // Test runtime.GOARCH displayed
    // Test runtime.Version() displayed
}

func TestAboutCommand_CommunityLinks(t *testing.T) {
    t.Parallel()
    // Test Discord link present
    // Test Twitter/X link present
    // Test Reddit link present
}

func TestAboutCommand_RepositoryLinks(t *testing.T) {
    t.Parallel()
    // Test GitHub link present
    // Test Releases link present
}

func TestAboutCommand_Styling(t *testing.T) {
    t.Parallel()
    // Test lipgloss styles applied
    // Test terminal width handling
}
```

### Expected Output Format
```
Nexora v0.29.3
AI-Powered CLI Agent

Platform: linux/amd64
Go Version: go1.23.4
License: MIT

Community:
  Discord:   https://discord.gg/GCyC6qT79M
  Twitter/X: https://x.com/i/communities/2004598673062216166/
  Reddit:    r/Zackor

Repository:
  GitHub:    https://github.com/jeffersonwarrior/nexora
  Releases:  https://github.com/jeffersonwarrior/nexora/releases
```

### Implementation Notes
- Use `internal/version/version.go` for version info
- Use lipgloss for styling
- Register in rootCmd.AddCommand()
- Match README badge information

---

## Feature 3: Test Coverage Audit (#5)

### Branch
`feature/test-coverage`

### Priority Areas (By Coverage Gap)
1. `internal/agent/` - Core agent logic (LOW coverage)
2. `internal/db/` - Database operations (32.0%)
3. `internal/session/` - Session management (88.3%)
4. `internal/task/` - Task management (needs tests)
5. `internal/tui/` - UI components

### Test Specification

```go
// Add tests to these packages:

// internal/agent/agent_test.go
func TestAgent_ProcessMessage(t *testing.T) {}
func TestAgent_HandleToolCall(t *testing.T) {}
func TestAgent_Summarize(t *testing.T) {}

// internal/db/db_test.go
func TestDB_Migrations(t *testing.T) {}
func TestDB_SessionCRUD(t *testing.T) {}
func TestDB_MessageCRUD(t *testing.T) {}

// internal/task/manager_test.go
func TestManager_CreateTask(t *testing.T) {}
func TestManager_AnalyzeDrift(t *testing.T) {}
func TestManager_MilestoneProgress(t *testing.T) {}
```

### Coverage Targets
| Package | Current | Target |
|---------|---------|--------|
| internal/agent | ~10% | 30%+ |
| internal/db | 32.0% | 45%+ |
| internal/session | 88.3% | 90%+ |
| internal/task | ~20% | 50%+ |
| **Overall** | **29%** | **40%+** |

### Implementation Notes
- Use testify/require for assertions
- All tests must call t.Parallel() unless modifying global state
- Use table-driven tests where appropriate
- Mock external dependencies

---

## Feature 4: Task Graph Enrichment (#6)

### Branch
`feature/task-graph`

### Files to Modify
- `internal/task/manager.go` - Add dependency tracking
- `internal/task/service.go` - Add dependency methods

### Files to Create
- `internal/task/graph.go` - Graph algorithms
- `internal/task/graph_test.go`
- `internal/task/visualize.go` - ASCII rendering
- `internal/task/visualize_test.go`

### Test Specification

```go
// internal/task/graph_test.go

package task

import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestGraph_AddDependency(t *testing.T) {
    t.Parallel()
    // Test adding dependency between tasks
    // Test circular dependency detection
    // Test self-reference prevention
}

func TestGraph_GetDependencies(t *testing.T) {
    t.Parallel()
    // Test getting direct dependencies
    // Test getting transitive dependencies
}

func TestGraph_TopologicalSort(t *testing.T) {
    t.Parallel()
    // Test correct ordering
    // Test with diamond dependencies
    // Test with independent branches
}

func TestGraph_CalculateProgress(t *testing.T) {
    t.Parallel()
    // Test progress rollup through dependencies
    // Test weighted progress calculation
    // Test with mixed completion states
}

func TestGraph_DetectCycle(t *testing.T) {
    t.Parallel()
    // Test cycle detection algorithm
    // Test complex cycle patterns
}

// internal/task/visualize_test.go

func TestVisualize_ASCII(t *testing.T) {
    t.Parallel()
    // Test ASCII tree rendering
    // Test with single task
    // Test with linear chain
    // Test with branching dependencies
}

func TestVisualize_Progress(t *testing.T) {
    t.Parallel()
    // Test progress bars in visualization
    // Test completion indicators
}
```

### Data Model Changes
```go
// Add to Task struct
type Task struct {
    // ... existing fields
    Dependencies []string `json:"dependencies"` // Task IDs this depends on
    Dependents   []string `json:"dependents"`   // Tasks that depend on this
}

// New types
type TaskGraph struct {
    tasks map[string]*Task
    edges map[string][]string // adjacency list
}
```

### Implementation Notes
- Use Kahn's algorithm for topological sort
- DFS for cycle detection
- Box-drawing characters for ASCII visualization

---

## Feature 5: Checkpoint System (#7)

### Branch
`feature/checkpoint`

### Files to Create
- `internal/session/checkpoint.go`
- `internal/session/checkpoint_test.go`
- `internal/agent/recovery/checkpoint.go`
- `internal/agent/recovery/checkpoint_test.go`

### Test Specification

```go
// internal/session/checkpoint_test.go

package session

import (
    "testing"
    "github.com/stretchr/testify/require"
)

func TestCheckpoint_Create(t *testing.T) {
    t.Parallel()
    // Test checkpoint creation
    // Test serialization of session state
    // Test file/db storage
}

func TestCheckpoint_Restore(t *testing.T) {
    t.Parallel()
    // Test session restoration
    // Test message history recovery
    // Test context reconstruction
}

func TestCheckpoint_Auto(t *testing.T) {
    t.Parallel()
    // Test auto-checkpoint trigger on token threshold
    // Test auto-checkpoint on time interval
    // Test auto-checkpoint on message count
}

func TestCheckpoint_List(t *testing.T) {
    t.Parallel()
    // Test listing available checkpoints
    // Test checkpoint metadata
}

func TestCheckpoint_Delete(t *testing.T) {
    t.Parallel()
    // Test checkpoint deletion
    // Test cleanup of old checkpoints
}

func TestCheckpoint_Corruption(t *testing.T) {
    t.Parallel()
    // Test handling of corrupted checkpoints
    // Test fallback behavior
}

// internal/agent/recovery/checkpoint_test.go

func TestRecovery_DetectCrash(t *testing.T) {
    t.Parallel()
    // Test crash detection on startup
    // Test unclean shutdown detection
}

func TestRecovery_OfferRestore(t *testing.T) {
    t.Parallel()
    // Test restore prompt to user
    // Test decline restore option
}

func TestRecovery_ResumeSession(t *testing.T) {
    t.Parallel()
    // Test full session resume
    // Test partial context restoration
}
```

### Data Model
```go
type Checkpoint struct {
    ID           string    `json:"id"`
    SessionID    string    `json:"session_id"`
    Timestamp    time.Time `json:"timestamp"`
    TokenCount   int64     `json:"token_count"`
    MessageCount int64     `json:"message_count"`
    ContextHash  string    `json:"context_hash"`
    State        []byte    `json:"state"` // Serialized session state
}

type CheckpointConfig struct {
    Enabled           bool  `json:"enabled"`
    TokenThreshold    int64 `json:"token_threshold"`    // Auto-checkpoint at N tokens
    IntervalSeconds   int   `json:"interval_seconds"`   // Auto-checkpoint every N seconds
    MaxCheckpoints    int   `json:"max_checkpoints"`    // Keep only last N
    CompressionLevel  int   `json:"compression_level"`  // 0-9 for gzip
}
```

### Implementation Notes
- Use gob or msgpack for serialization
- Store in SQLite with BLOB column
- Compress large states
- Keep last N checkpoints only

---

## Git Workflow

### Branch Naming
```
feature/prompt-library   <- Feature #3
feature/about-command    <- Feature #4
feature/test-coverage    <- Feature #5
feature/task-graph       <- Feature #6
feature/checkpoint       <- Feature #7
```

### Commit Message Format
```
type(scope): description

[body]

🤖 Generated with [Claude Code](https://claude.com/claude-code)
Co-Authored-By: Claude Opus 4.5 <noreply@anthropic.com>
```

### Types
- `feat`: New feature
- `test`: Adding tests
- `fix`: Bug fix
- `refactor`: Code restructuring
- `docs`: Documentation

---

## Validation Commands

### After Each Change
```bash
go build ./...
go vet ./...
```

### Before Marking Complete
```bash
go test ./... -race -coverprofile=coverage.out
go tool cover -func=coverage.out | grep total
```

### Full Test Suite
```bash
go test ./... -v -race -timeout 5m
```

---

## Success Criteria

- [ ] All tests pass: `go test ./...`
- [ ] Coverage >= 40%: `go tool cover -func=coverage.out`
- [ ] No race conditions: `-race` flag passes
- [ ] Build succeeds: `go build ./...`
- [ ] Vet passes: `go vet ./...`
- [ ] Each feature merged to main
- [ ] CHANGELOG updated

---

**Last Updated:** 2025-12-27
