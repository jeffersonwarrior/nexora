# Nexora v0.29.1-RC1 Roadmap

> **Created:** 2025-12-25
> **Status:** Active
> **Current Phase:** Phase 0 (Pre-flight)
>
> **Note:** No time estimates. AI swarm execution makes timing unpredictable.

---

## RC1 Exit Criteria

RC1 is complete when ALL of the following are true:

- [ ] All phases (0-5) completed and verified
- [ ] `go build ./...` succeeds with no errors
- [ ] `go test ./...` passes with 50%+ coverage
- [ ] TUI launches and all features function correctly
- [ ] All fixes manually verified in running application
- [ ] All changes committed to git

---

## Phase Overview

| Phase | Focus | Status |
|-------|-------|--------|
| **Phase 0** | Pre-flight Verification | 🔴 In Progress |
| **Phase 1** | Critical Fixes | ⏳ Pending |
| **Phase 2** | High Priority Fixes | ⏳ Pending |
| **Phase 3** | Test Coverage (30% → 50%) | ⏳ Pending |
| **Phase 4** | Tool Consolidation | ⏳ Pending |
| **Phase 5** | TUI Enhancements | ⏳ Pending |

**Note:** Multi-Agent Orchestration moved to v0.29.2 (see `NEXORA.0.29.2.12.26.md`)

---

# Phase 0: Pre-flight Verification

**Status:** 🔴 In Progress

Verify clean build state before making changes.

```bash
# Must all pass before proceeding
go build ./...
go vet ./...
go test ./... 2>&1 | head -50
```

- [ ] Build succeeds
- [ ] Vet passes
- [ ] Note any existing test failures (fix in Phase 3)

---

# Phase 1: Critical Fixes

**Status:** ⏳ Pending (2 remaining)

## Fix 1: Thinking Animation Display Bug

**File:** `internal/tui/components/chat/messages/messages.go:303-307`
**Issue:** Animation shows random characters (`F)@%5.!bc@£=EbA`) when thinking content exists

**Current Code:**
```go
} else if finishReason != nil && finishReason.Reason == message.FinishReasonCanceled {
    footer = t.S().Base.PaddingLeft(1).Render(m.toMarkdown("*Canceled*"))
} else {
    footer = m.anim.View()  // BUG: Shows animation even with content
}
```

**Fixed Code:**
```go
} else if finishReason != nil && finishReason.Reason == message.FinishReasonCanceled {
    footer = t.S().Base.PaddingLeft(1).Render(m.toMarkdown("*Canceled*"))
} else if strings.TrimSpace(m.message.ReasoningContent().Thinking) != "" {
    footer = ""  // Don't show animation when we have thinking content
} else {
    footer = m.anim.View()
}
```

**Verification:** Ensure `strings` package is imported.

---

## Fix 2: Observations Not Capturing

**Issue:** Observations dialog empty - ProjectID not passed to ToolExecution

### 2.1: Add ProjectID to SessionAgentOptions
**File:** `internal/agent/agent.go:175-181`
```go
type SessionAgentOptions struct {
    // ... existing fields ...
    Observer             *memory.Observer
    ProjectID            int  // ADD THIS LINE
    ResourceMonitor      *resources.Monitor
}
```

### 2.2: Add projectID to sessionAgent struct
**File:** `internal/agent/agent.go:140`
```go
observer  *memory.Observer
projectID int  // ADD THIS LINE
```

### 2.3: Initialize projectID in NewSessionAgent
**File:** `internal/agent/agent.go:195+`
```go
return &sessionAgent{
    // ... existing fields ...
    observer:             opts.Observer,
    projectID:            opts.ProjectID,  // ADD THIS LINE
    retryQueue:           csync.NewMap[string, *RetryRequest](),
}
```

### 2.4: Pass projectID from Coordinator
**File:** `internal/agent/coordinator.go:942-954`
```go
result := NewSessionAgent(SessionAgentOptions{
    LargeModel:         large,
    SmallModel:         small,
    SystemPromptPrefix: largeProviderCfg.SystemPromptPrefix,
    SystemPrompt:       systemPrompt,
    Sessions:           c.sessions,
    Messages:           c.messages,
    Tools:              nil,
    AIOPS:              c.aiops,
    Observer:           c.observer,
    ProjectID:          c.projectID,  // ADD THIS LINE
    ResourceMonitor:    c.resourceMonitor,
})
```

### 2.5: Use projectID in ToolExecution
**File:** `internal/agent/agent.go:1125-1132`
```go
exec := memory.ToolExecution{
    SessionID: currentAssistant.SessionID,
    ProjectID: a.projectID,  // ADD THIS LINE
    ToolName:  result.ToolName,
    Input:     map[string]interface{}{},
    Output:    output,
    Success:   success,
    Timestamp: time.Now(),
}
```

---

# Phase 2: High Priority Fixes

**Status:** ⏳ Pending

---

## Fix 3: "/" Command Trigger

**File:** `internal/tui/tui.go:544-557`
**Issue:** Typing `/` anywhere triggers commands menu

### 3.1: Add EditorValue method to ChatPage
**File:** `internal/tui/page/chat/chat.go`
```go
func (p *chatPage) EditorValue() string {
    return p.editor.Value()
}
```

### 3.2: Check editor before triggering commands
**File:** `internal/tui/tui.go:544`
```go
case key.Matches(msg, a.keyMap.Commands):
    if !a.isConfigured {
        return nil
    }
    // Only trigger "/" menu when editor is empty
    if chatPage, ok := a.pages[chat.ChatPageID].(interface{ EditorValue() string }); ok {
        if chatPage.EditorValue() != "" {
            return nil  // Let "/" pass through to editor
        }
    }
    if a.dialog.ActiveDialogID() == commands.CommandsDialogID {
        return util.CmdHandler(dialogs.CloseDialogMsg{})
    }
    // ... rest unchanged
```

---

## Fix 4: Keyboard Shortcut Conflicts

**Issue:** ctrl+l/ctrl+m, ctrl+p, ctrl+n have conflicts

### 4.1: Change Models shortcut to ctrl+e
**File:** `internal/tui/keys.go:38-41`
```go
Models: key.NewBinding(
    key.WithKeys("ctrl+e"),  // Changed from "ctrl+l", "ctrl+m"
    key.WithHelp("ctrl+e", "models"),
),
```

**File:** `internal/tui/components/dialogs/commands/commands.go`
Find Models command, change `Shortcut: "ctrl+e"`

### 4.2: Remove ctrl+n from dialog navigation
**Files:**
- `internal/tui/components/dialogs/commands/keys.go:22`
- `internal/tui/components/dialogs/sessions/keys.go:21`
- `internal/tui/components/dialogs/models/keys.go:32`
- `internal/tui/components/dialogs/reasoning/reasoning.go:60`

**Change:** `key.WithKeys("down", "ctrl+n")` → `key.WithKeys("down", "j")`

### 4.3: Remove ctrl+p from dialog navigation
**Files:**
- `internal/tui/components/dialogs/commands/keys.go:26`
- `internal/tui/components/dialogs/sessions/keys.go:25`
- `internal/tui/components/dialogs/models/keys.go:36`
- `internal/tui/components/dialogs/reasoning/reasoning.go:64`

**Change:** `key.WithKeys("up", "ctrl+p")` → `key.WithKeys("up", "k")`

---

## ~~Fix 5: Deprecated ShouldContinue Method~~

**Status:** FIXED ✓

Already using `StateMachine.ShouldContinue()` instead of deprecated `ConversationManager.ShouldContinue()`. Verified in `agent.go:1289, 1439, 1503, 1514, 1551`.

---

## Previously Completed Fixes

### Agent Continuation Bug
**Status:** FIXED ✓
Lines 1503 and 1514 in `internal/agent/agent.go` now use `shouldContinue || sm.ShouldContinue()`.

### Duplicate Commands in Help Output
**Status:** FIXED ✓
Removed duplicate registration from `agent.go:276`.

### TUI Status Indicators
**Status:** WORKING ✓ (Verified in header.go:104-118)
- `StateThinking` → "Thinking"
- `StateStreamingResponse` → "Streaming"
- `StateExecutingTool` → "Executing"

---

# Phase 3: Test Coverage (30% → 50%)

**Status:** ⏳ Pending

---

## Current State

| Metric | Value |
|--------|-------|
| Overall Coverage | ~30.1% |
| Target Coverage | 50% |
| Packages with Tests | 65 |
| Packages with 0% Coverage | 39+ |

---

## Immediate Blockers

### 1. Build Failures
```
internal/message_test.go: MockQuerier missing AddFavorite, RemoveFavorite
internal/session_test.go: MockQuerier missing AddFavorite, RemoveFavorite
```
**Fix**: Add stub methods to MockQuerier implementations.

### 2. Schema Mismatch
```
internal/agent/prompt: "table prompt_library has no column named content_hash"
```
**Fix**: Test DB setup needs migration `20251225000004_add_content_hash.sql`.

### 3. SQL Logic Bug
```
internal/db: TestCategoryAndTagFiltering - expected 2, got 3
```
**Fix**: Review `SearchPromptsWithFilters` SQL query logic.

---

## Tier 1: Critical Packages (<25% → 50%)

| Package | Current | Target |
|---------|---------|--------|
| `internal/agent` | 21.3% | 50% |
| `internal/agent/tools` | 15.0% | 50% |
| `internal/cmd` | 19.2% | 40% |
| `internal/tui/page/chat` | 4.9% | 30% |
| `internal/tui/components/chat/*` | 3-7% | 30% |

---

## Tier 2: Important Packages (25-50% → 55%)

| Package | Current | Target |
|---------|---------|--------|
| `internal/agent/delegation` | 38.2% | 55% |
| `internal/agent/memory` | 44.8% | 55% |
| `internal/db` | 32.8% | 55% |
| `internal/indexer` | 42.3% | 55% |
| `internal/config` | 58.6% | 65% |

---

## CI/CD Integration

```yaml
# .github/workflows/test.yml
- name: Check coverage
  run: |
    go test -coverprofile=coverage.out ./...
    COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}')
    if (( $(echo "$COVERAGE < 50" | bc -l) )); then
      echo "Coverage $COVERAGE is below 50%"
      exit 1
    fi
```

---

# Phase 4: Tool Consolidation

**Status:** ⏳ Pending
**Savings:** ~45K tokens/call (35%)

---

## Transparent Aliasing Strategy

Old tool names remain functional but are transparently remapped:

```go
// internal/agent/tools/aliases.go
var ToolAliases = map[string]string{
    "bash_monitored":   "bash",
    "agentic_fetch":    "fetch",
    "web_fetch":        "fetch",
    "agent_list":       "delegate",
    "agent_status":     "delegate",
    "agent_run":        "delegate",
    "track_prompt_usage": "",  // removed, log warning
    "prompt_analytics":   "",  // removed, log warning
}

func ResolveToolName(name string) (resolved string, wasAlias bool) {
    if alias, ok := ToolAliases[name]; ok {
        log.Debug("Tool alias used", "old", name, "new", alias)
        return alias, true
    }
    return name, false
}
```

**Behavior:**
- AI uses old name → silently remapped to new name
- TUI displays new/correct tool name
- Logs note the alias usage for future prompt updates
- No old names injected into prompts

---

## 4.1: Bash Tools → Single `bash`

**Remove:** `bash.go` (old implementation)
**Rename:** `bash_monitored.go` → `bash.go`

```go
type BashParams struct {
    Command            string `json:"command"`
    Description        string `json:"description,omitempty"`
    Purpose            string `json:"purpose,omitempty"`
    CompletionCriteria string `json:"completion_criteria,omitempty"`
    AutoTerminate      bool   `json:"auto_terminate,omitempty"`
    TimeoutMinutes     int    `json:"timeout_minutes,omitempty"`
}
```

**Detection Logic:**
- If `purpose` AND `completion_criteria` provided → AI monitoring
- Otherwise → Standard execution

---

## 4.2: Fetch Tools → Single `fetch`

**Remove:** `fetch.go`, `agentic_fetch_tool.go`
**Rename:** `web_fetch.go` → `fetch.go`

```go
type FetchParams struct {
    URL     string `json:"url"`
    Format  string `json:"format,omitempty"`  // text, markdown, html
    Mode    string `json:"mode,omitempty"`    // auto, web_reader, raw
    Timeout int    `json:"timeout,omitempty"`
}
```

**Auto-fallback:** web_reader → raw on failure

---

## 4.3: Agent Tools → Single `delegate`

**Remove:** `agents.go`, `agent_list.go`, `agent_status.go`, `agent_run.go`
**Enhance:** `delegate.go` with action parameter

```go
type DelegateParams struct {
    Action          string `json:"action,omitempty"`  // spawn, list, status, stop, run
    TaskDescription string `json:"task_description,omitempty"`
    AgentType       string `json:"agent_type,omitempty"`
    SessionID       string `json:"session_id,omitempty"`
    Prompt          string `json:"prompt,omitempty"`
    Blocking        bool   `json:"blocking,omitempty"`
}
```

---

## 4.4: Remove Analytics Tools

**Remove:**
- `internal/agent/tools/track_prompt_usage.go`
- `internal/agent/tools/prompt_analytics.go`

**Note:** Query database directly if needed.

---

## Files to Delete (9)

- `internal/agent/tools/bash.go`
- `internal/agent/tools/fetch.go`
- `internal/agent/agentic_fetch_tool.go`
- `internal/agent/tools/agents.go`
- `internal/agent/tools/agent_list.go`
- `internal/agent/tools/agent_status.go`
- `internal/agent/tools/agent_run.go`
- `internal/agent/tools/track_prompt_usage.go`
- `internal/agent/tools/prompt_analytics.go`

---

# Phase 5: TUI Enhancements

**Status:** ⏳ Pending

---

## 5.1: Auto-LSP Detection and Installation

**Behavior:**
- Auto-enable silently (no prompts)
- Auto-install LSP executables if missing
- Per-project config (`.nexora/nexora.json`)

**Detection Matrix:**
```
Root Marker         → LSP                       → Install Command
───────────────────────────────────────────────────────────────────
go.mod              → gopls                     → go install golang.org/x/tools/gopls@latest
Cargo.toml          → rust-analyzer             → rustup component add rust-analyzer
package.json        → typescript-language-server → npm i -g typescript-language-server
pyproject.toml      → pyright                   → pip install pyright
requirements.txt    → pyright                   → pip install pyright
```

**New file:** `internal/lsp/autodetect.go`

---

## 5.2: TUI Settings Panel

**Settings:**
| Setting | Type | Default |
|---------|------|---------|
| Yolo Mode | toggle | off |
| Sidebar | toggle | on |
| Help | toggle | on |
| LSP Auto-detect | toggle | on |
| LSP Auto-install | toggle | on |

**New file:** `internal/tui/components/dialogs/settings/settings.go`

---

## 5.3: Unified Delegate Command

### Dynamic Resource-Based Pool
```go
type ResourceConfig struct {
    CPUPerAgent     float64       // ~5%
    MemPerAgentMB   uint64        // ~512MB
    MaxCPUUsage     float64       // 70%
    MaxMemUsage     float64       // 75%
    QueueTimeout    time.Duration // 30min
}
```

### Completion Banner Notifications

**New file:** `internal/tui/components/banner/banner.go`
- Banner component with auto-dismiss (10s)
- Success (green) / Error (red) styling

---

## 5.4: Configurable Prompt Repository Import

**CLI Enhancement:**
```bash
nexora import-prompts                              # Default repo
nexora import-prompts -r https://github.com/...   # Custom repo
nexora import-prompts -u                          # Update/sync
```

---

# Testing Checklist

## Phase 0 Verification
- [ ] `go build ./...` succeeds
- [ ] `go vet ./...` passes
- [ ] Note existing test failures

## Phase 1 Verification
- [ ] Fix 1: No random characters in thinking display
- [ ] Fix 2: Observations dialog shows captured tools

## Phase 2 Verification
- [ ] Fix 3: Can type `/home/path` without triggering menu
- [ ] Fix 4: ctrl+e opens models, ctrl+p opens prompts, ctrl+n creates new session
- [ ] All dialogs navigate with j/k or arrows only

## Phase 3 Verification
- [ ] `go test ./...` passes
- [ ] Coverage ≥ 50%

## Phase 4 Verification
- [ ] Old tool names work via aliasing
- [ ] TUI shows new tool names
- [ ] Logs show alias usage

## Phase 5 Verification
- [ ] LSP auto-detected for Go project
- [ ] Settings panel accessible
- [ ] Agent completion banners display

## Full Validation
```bash
go build ./...
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out | grep total
./nexora  # Manual TUI verification
```

---

# Summary

| Phase | Focus | Fixes | Status |
|-------|-------|-------|--------|
| **Phase 0** | Pre-flight | - | 🔴 In Progress |
| **Phase 1** | Critical | Fix 1-2 | ⏳ Pending |
| **Phase 2** | High Priority | Fix 3-5 | ⏳ Pending |
| **Phase 3** | Test Coverage | 30%→50% | ⏳ Pending |
| **Phase 4** | Tool Consolidation | 27→19 tools | ⏳ Pending |
| **Phase 5** | TUI Enhancements | LSP, Settings | ⏳ Pending |

**v0.29.2:** Multi-Agent Orchestration (see `NEXORA.0.29.2.12.26.md`)
