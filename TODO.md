# Nexora Roadmap

**Current Version:** v0.29.1-RC1
**Status:** See NEXORA.0.29.2.12.26.md for complete roadmap

---

## Quick Reference

All release planning consolidated in: **NEXORA.0.29.2.12.26.md**

| Version | Focus | Status |
|---------|-------|--------|
| **v0.29.1-RC1** | Bug fixes, test coverage, tool consolidation | In Progress (Phases 0-2 complete) |
| **v0.29.2** | MiniMax M2.1, Synthetic provider, TMUX fixes, Windows support | ✅ Released |
| **v0.29.3** | About command, task graph enrichment + checkpoints | Planned |
| **v0.29.4** | sqlite3vfshttp, A2A + ACP communication | Planned |
| **v0.29.5** | Protocol composition + conflict resolution | Planned |
| **v3.0** | ModelScan integration + VNC/Docker dual-mode | Planned |

---

## Current Work: v0.29.1-RC1

**Status Documents:**
- Roadmap: `NEXORA.29.1-RC1.12.25.md` (in archives/historical-docs/)
- Audit: `RC1-AUDIT-REPORT.md`
- VCR Strategy: `VCR_TEST_STRATEGY.md`

**Progress:**
- Phase 0 (Pre-flight): ✅ Complete
- Phase 1 (Critical Fixes): ✅ Complete
- Phase 2 (High Priority Fixes): ✅ Complete
- Phase 3 (Test Coverage): ⚠️ **Functionally Complete with Limitations**
- Phase 4 (Tool Consolidation): ⏳ Pending
- Phase 5 (TUI Enhancements): ⏳ Pending

**Phase 3 Assessment:**

✅ **Achieved:**
- Session title generation fix implemented (agent.go:540)
- New test added: `TestTitleGenerationForNewSession`
- VCR cassettes recorded for all 4 providers
- Non-VCR test packages stable (session: 88.3%, config: 49.7%, db: 32.0%)
- Overall coverage: 29.4%

⚠️ **Limitations (Documented):**
- VCR tests have inherent flakiness due to non-deterministic request matching
- Anthropic: 100% reliable
- OpenAI: 100% reliable
- OpenRouter: 93% reliable (1 flaky test)
- ZAI: 50% reliable (7 flaky tests)
- Total reliable agent tests: 28/56 = 50%

📋 **Decision:**
Accept current state as "Phase 3 complete" with documented VCR limitations. Focus CI on stable tests, proceed to Phase 4.

**See Also:**
- `VCR_TEST_STRATEGY.md` - Known flaky tests and mitigation
- `TESTING_PROGRESS_REPORT.md` - Detailed progress tracking

---

## Future Versions

See **NEXORA.0.29.2.12.26.md** for detailed planning:
- v0.29.2-0.29.5: Multi-agent orchestration system
- v3.0: ModelScan integration + visual terminal (VNC/Docker)

---

## Known Issues

### Session Title Re-generation
**Priority:** Medium

Sessions with "New Session" as title don't get retitled on first message.

**Root Cause:** `generateTitle()` checks `MessageCount == 0` but doesn't check if current title is placeholder.

**Fix Options:**
1. Check `MessageCount == 0 OR title == "New Session"`
2. Add `needs_title` boolean flag to session schema
See v3.0 section in NEXORA.0.29.2.12.26.md for details.

---

## Phase 4: Tool Consolidation (v0.29.1-RC1)

**Status:** ⏳ Pending
**Goal:** Reduce tool files from 47 	 ~40 (7 files removed via consolidation)
**Approach:** Test-first - create new test, make changes, verify both old and new tests pass

### Phase 4 Features

#### Feature 4.1: Bash Tool Consolidation
**Goal:** Merge `bash.go` + `bash_monitored.go` 	 single `bash.go` with mode detection

**Files:**
- `internal/agent/tools/bash.go` (343 lines) - standard execution
- `internal/agent/tools/bash_monitored.go` (90 lines) - AI-monitored execution

**STATUS: ALREADY COMPLETED ✅**
- `bash_monitored.go` was already removed (doesn't exist)
- `bash.go` is the only bash tool
- Test file `bash_monitored_test.go` is a placeholder

**Success Criteria:**
- [x] New test created: N/A (already complete)
- [x] Both standard and monitored modes work: N/A (merged)
- [x] Old bash tests still pass: bash_tool_test.go passes
- [x] `bash_monitored.go` deleted: Confirmed
- [x] No functionality regression: Confirmed

---

#### Feature 4.2: Fetch Tool Consolidation - SMART FETCH

**Goal:** Create intelligent, context-aware fetch tool with MCP auto-routing

**Design:**
```
fetch(url, format, force_builtin?) 
  	 MCP available? 	 Use MCP (always prefer MCP if installed)
  	 No MCP? 	 Built-in fetch
  	 Check token count 	 Within context limit? 	 Return in response
  	 Too large? 	 Write to ./tmp/session-id/, return path
```

**Key Features:**

1. **MCP Auto-Routing** (ALWAYS prefer MCP if installed)
   - Detect web_reader MCP availability
   - Auto-route to MCP for best experience
   - Fallback to built-in if MCP fails
   - Log which path was taken

2. **Context-Aware Content Handling**
   - Count tokens in content (simple math, not HTTP call)
   - Compare to max model context size
   - If within limit 	 return in response
   - If over limit 	 write to tmp file

3. **Session-Based Tmp Files**
   - Write to `./tmp/nexora-{session-id}/`
   - Delete on session end
   - Return file path to user

4. **Timeout**
   - Built-in fetch: 30 second timeout
   - Built-in is fallback, MCP is primary

**Files:**
- `internal/agent/tools/fetch.go` - ✅ Modified with smart routing (completed 2025-12-26)
- `internal/agent/tools/web_fetch.go` - ⏸️ KEPT for sub-agents (no permissions needed)
- `internal/agent/tools/fetch_types.go` - Keep (params definitions)
- `internal/agent/tools/fetch_smart_test.go` - ✅ Created (452 lines, all tests pass)

**Test Plan (TDD):**
```go
// TestMCPDetection - MCP available 	 uses MCP
// TestMCPFallback - MCP fails 	 falls back to built-in
// TestContextUnderLimit - Small content 	 returns in response
// TestContextOverLimit - Large content 	 writes to tmp
// TestTmpFilePath - Returns session-scoped path
// TestTimeoutBuiltin - Built-in respects 30s timeout
// TestFormatAll - text/markdown/html all work
```

**Approach:**
1. Create new test: `fetch_smart_test.go`
2. Write tests for all behaviors (RED)
3. Implement features to make tests pass (GREEN)
4. Verify old tests still pass (REFACTOR)
5. Delete `web_fetch.go` (merge complete)

**Success Criteria:**
- [x] New test created: `fetch_smart_test.go` (452 lines, 10 test functions)
- [x] MCP auto-routing works (MCP available 	 uses MCP, else fallback to built-in)
- [x] Context token counting works (simple word-based approximation)
- [x] Session-based tmp files work (./tmp/nexora-{session-id}/ naming)
- [x] All old fetch tests still pass (18 tests pass)
- [x] No functionality regression
- [x] `web_fetch.go` - DECISION: KEPT for sub-agents (agentic_fetch_tool.go uses it)

---

#### Feature 4.3: Agent Tools Consolidation
**Goal:** Merge agent_*.go files 	 `delegate.go` with action parameter

**Files:**
- `internal/agent/tools/agents.go`
- `internal/agent/tools/agent_list.go`
- `internal/agent/tools/agent_status.go`
- `internal/agent/tools/agent_run.go`
- `internal/agent/tools/delegate.go` (KEEP, add actions)

**Approach:**
1. Create new test: `delegate_consolidated_test.go`
2. Add action parameter to `delegate.go`: spawn, list, status, stop, run
3. Run new test - verify it passes
4. Run old tests - verify backward compatibility
5. Delete old agent_*.go files

**Success Criteria:**
- [ ] New test created
- [ ] All actions work: spawn, list, status, stop, run
- [ ] Old delegate tests still pass
- [ ] Old agent files deleted
- [ ] No functionality regression

---

#### Feature 4.4: Remove Analytics Tools
**Goal:** Delete deprecated analytics tools

**Files:**
- `internal/agent/tools/track_prompt_usage.go` - already removed ✅
- `internal/agent/tools/prompt_analytics.go` - already removed ✅

**Approach:**
1. Verify files are removed
2. Run tests to confirm no breakage

**Success Criteria:**
- [ ] Files confirmed removed
- [ ] No test failures from removal

---

#### Feature 4.5: Tool Aliasing System
**Goal:** Enhance and test `aliases.go` for backward compatibility

**Files:**
- `internal/agent/tools/aliases.go`
- `internal/agent/tools/aliases_test.go`

**Approach:**
1. Create new integration test: `aliasing_integration_test.go`
2. Test all alias mappings
3. Test tool dispatch with aliases
4. Verify backward compatibility

**Success Criteria:**
- [ ] New integration test created
- [ ] All aliases map correctly
- [ ] Tool dispatch works with aliases
- [ ] Backward compatibility verified

---

#### Feature 4.3: Agent Tools Consolidation
**Goal:** Merge agent_*.go files 	 `delegate.go` with action parameter

**Files:**
- `internal/agent/tools/agents.go` - agent management
- `internal/agent/tools/agent_list.go` - list agents
- `internal/agent/tools/agent_status.go` - get agent status
- `internal/agent/tools/agent_run.go` - run agent
- `internal/agent/tools/delegate.go` (5266 lines) - KEEP, add actions

**Approach:**
1. Create new test: `delegate_consolidated_test.go` - tests all actions
2. Add action parameter to `delegate.go`: spawn, list, status, stop, run
3. Run new test - verify it passes
4. Run old tests - verify backward compatibility
5. Delete old agent_*.go files

**Success Criteria:**
- [ ] New test created: `delegate_consolidated_test.go`
- [ ] All actions work: spawn, list, status, stop, run
- [ ] Old delegate tests still pass
- [ ] Old agent files deleted
- [ ] No functionality regression

**Test Plan:**
```go
// TestDelegateSpawn - action=spawn works
// TestDelegateList - action=list works
// TestDelegateStatus - action=status works
// TestDelegateStop - action=stop works
// TestDelegateRun - action=run works
```

---

#### Feature 4.4: Remove Analytics Tools
**Goal:** Delete deprecated analytics tools

**Files:**
- `internal/agent/tools/track_prompt_usage.go` - already removed ✅
- `internal/agent/tools/prompt_analytics.go` - already removed ✅

**Approach:**
1. Verify files are removed
2. Run tests to confirm no breakage

**Success Criteria:**
- [ ] Files confirmed removed
- [ ] No test failures from removal

---

#### Feature 4.5: Tool Aliasing System
**Goal:** Enhance and test `aliases.go` for backward compatibility

**Files:**
- `internal/agent/tools/aliases.go` (exists)
- `internal/agent/tools/aliases_test.go` (exists, needs verification)

**Approach:**
1. Create new integration test: `aliasing_integration_test.go`
2. Test all alias mappings
3. Test tool dispatch with aliases
4. Verify backward compatibility

**Success Criteria:**
- [ ] New integration test created
- [ ] All aliases map correctly
- [ ] Tool dispatch works with aliases
- [ ] Backward compatibility verified

**Test Plan:**
```go
// TestBashAlias - bash_monitored 	 bash works
// TestFetchAlias - web_fetch/agentic_fetch 	 fetch works
// TestAgentAlias - agent_* 	 delegate works
// TestToolDispatchWithAliases - dispatch uses aliases
```

---

### Phase 4 Execution Order

1. **Feature 4.1** - Bash Consolidation (Foundation)
2. **Feature 4.2** - Fetch Consolidation (Foundation)
3. **Feature 4.4** - Analytics Removal (Trivial)
4. **Feature 4.3** - Agent Consolidation (Depends on 4.1, 4.2)
5. **Feature 4.5** - Aliasing Integration (Depends on 4.1-4.3)

---

| 4.5 Aliasing | ✅ COMPLETE | ✅ PASS | ✅ PASS | ⬜ |
| 4.6 Bash TMUX | ✅ COMPLETE | ⬜ | ⬜ | ⬜ |

## Phase 4 Analysis Summary

### Feature 4.1: Bash - ALREADY COMPLETE ✅
- `bash_monitored.go` was already removed before this phase
- Only `bash.go` exists (341 lines)
- All bash tests pass

### Feature 4.2: Fetch - ✅ COMPLETE
**Decision: Option A (Keep Separate) ✅ CONFIRMED 2025-12-26**

**Implementation Summary:**
- ✅ `fetch.go` = full-featured with permissions, MCP auto-routing, smart context handling
- ✅ `web_fetch.go` = KEPT for sub-agents (no permissions needed, agentic_fetch_tool.go uses it)
- ✅ Both serve different purposes - documented in code
- ✅ All tests pass (18 fetch tests, 10 new smart fetch tests)

### Feature 4.3: Agent - ⏸️ N/A (Already Separate Package)
**Status: NOT APPLICABLE**
- Agent management is in separate `/internal/agent/agents/` package
- `agent_*.go` files were archived (moved to `_archived/`) because they don't belong in tools
- Tools should NOT have agent management - that's in the agents package
### Feature 4.4: Analytics - ✅ ALREADY DONE
- Analytics files moved to `_archived/` in previous phase
- `track_prompt_usage.go.txt` - archived
- `prompt_analytics.go.txt` - archived
- No analytics tools remain in active codebase
- ✅ VERIFIED: No action needed
### Feature 4.5: Aliasing - ✅ COMPLETE
**Purpose:** Allow models to use tool name variations (curl, wget, get 	 fetch)

**Implementation:**
- Created `internal/agent/tools/aliases.go` with alias map and resolution function
- Created `internal/agent/tools/aliases_test.go` with 16 comprehensive tests
- Integrated into `agent.go` line 185 to resolve aliases before tool dispatch

**Aliases Supported:**
- Fetch: curl, wget, http-get, http_get, web-fetch, webfetch, web_fetch, http
- View: read, cat, open
- List: ls, dir, directory
- Edit: modify, change, replace, update
- Write: create, make, new
- Search: search, find, rg 	 grep
- Bash: shell, exec, execute, run, command
- Web Search: web-search, websearch, search-web 	 web_search
- Sourcegraph: sg, code-search 	 sourcegraph
- Download: (separate tool, not aliased)

**Test Results:** 16/16 tests pass ✅

### Feature 4.6: Bash TMUX Integration - ✅ COMPLETE
**Purpose:** Unified bash shell management with TMUX support, eliminate job_kill/job_output tools

**Implementation:**
- Created `internal/shell/tmux.go` - TMUX session manager (400+ lines)
  - `NewTmuxSession()` - Create TMUX sessions/panes
  - `SendCommand()` - Execute commands in TMUX panes
  - `CaptureOutput()` - Get pane output
  - `KillSession()` / `KillAll()` - Session cleanup
  - `IsTmuxAvailable()` - TMUX detection
- Enhanced `internal/agent/tools/bash.go`:
  - Added `ShellID` parameter for session continuation
  - Added TMUX routing (TMUX if available, else legacy)
  - Added `executeTmuxCommand()` function
  - Added `executeJobManagement()` for job_kill/job_output aliasing
- Updated `internal/agent/tools/aliases.go`:
  - Added `job_kill` 	 `bash` alias
  - Added `job_output` 	 `bash` alias

**Key Features:**
- ✅ TMUX-first: Uses TMUX if available, falls back to legacy shell manager
- ✅ Multiple sessions: Unlimited TMUX sessions per Nexora session
- ✅ Session cleanup: All TMUX sessions killed on session end
- ✅ Unified tool: job_kill and job_output aliased to bash
- ✅ Metadata: TMUX session/pane IDs included in responses

**Aliases Added:**
- `job_kill` 	 `bash`
- `job_output` 	 `bash`

**Files:**
- Created: `internal/shell/tmux.go` (400+ lines)
- Modified: `internal/agent/tools/bash.go` (+200 lines)
- Modified: `internal/agent/tools/aliases.go` (+2 lines)
- **Kept:** `install_manager.go` (dependency management, 357 lines)
- **Removed:** N/A (aliases, no file deletion)

---

### Phase 4 Commands

```bash
# Run all new consolidation tests
go test ./internal/agent/tools -run "Consolidated|Aliasing" -v

# Run all tool tests
go test ./internal/agent/tools -v

# Check file count
find internal/agent/tools -name "*.go" -type f | wc -l
```

---

## Phase 5: TUI Enhancements (Future)

**Prerequisite:** Phase 4 complete
---

## v0.29.3 Features

**Status:** Planned  
**Release Date:** TBD

### Feature: About Command (`nexora about`)

**Goal:** Display project information, version, community links, and platform details matching README badges

**Command:** `nexora about`

**Output Format:**
```
Nexora v0.29.3
AI-Powered CLI Agent

Platform: linux/amd64
Go Version: go1.23.4
License: MIT

Production-Ready AI Terminal Assistant with intelligent state 
management, adaptive resource monitoring, and self-healing execution.

🌐 Community
  Discord:    https://discord.gg/GCyC6qT79M
  Twitter/X:  https://x.com/i/communities/2004598673062216166/
  Reddit:     r/Zackor

📦 Repository
  GitHub:     https://github.com/jeffersonwarrior/nexora
  Releases:   https://github.com/jeffersonwarrior/nexora/releases

✨ Features
  • 70+ AI models across 9+ providers
  • TMUX-based persistent shell sessions
  • MCP integration (Z.AI Vision, Web Reader/Search)
  • Cross-platform support (Linux, macOS, Windows)

For more information, visit: https://nexora.land
```

**Implementation:**
- New command file: `internal/cmd/about.go`
- Add to root command in `internal/cmd/root.go`
- Use lipgloss for styled output
- Include runtime.GOOS, runtime.GOARCH, runtime.Version()
- Pull version from `internal/version/version.go`

**Success Criteria:**
- [ ] `nexora about` displays formatted project info
- [ ] All community links included (Discord, Twitter/X, Reddit)
- [ ] Platform and Go version shown
- [ ] Styled output with lipgloss
- [ ] Matches README badge information

---

### Other v0.29.3 Features (Planned)

- Task graph enrichment
- Checkpoint system
- Additional improvements TBD

---

## v0.29.4 Features

**Status:** Planned
**Release Date:** TBD

### Feature: SQLite HTTP VFS (`sqlite3vfshttp`)

**Goal:** Replace local SQLite file reads with HTTP-based virtual filesystem for remote database access

**Library:** [psanford/sqlite3vfshttp](https://github.com/psanford/sqlite3vfshttp)

**Use Case:**
- Enable reading SQLite databases over HTTP without downloading the entire file
- Supports range requests for efficient partial reads
- Useful for distributed/remote database access scenarios

**Implementation:**
- Add `github.com/psanford/sqlite3vfshttp` dependency
- Modify `internal/db/db.go` to support HTTP-based database connections
- Add configuration option for HTTP vs local file mode
- Implement connection string parsing for HTTP URLs

**Technical Notes:**
```go
import (
    "github.com/psanford/sqlite3vfshttp"
    "database/sql"
)

// Register HTTP VFS
sqlite3vfshttp.RegisterHTTP()

// Open database over HTTP
db, err := sql.Open("sqlite3", "http://example.com/database.db?vfs=httpvfs")
```

**Success Criteria:**
- [ ] HTTP VFS registered on startup
- [ ] Database reads work over HTTP with range requests
- [ ] Fallback to local file if HTTP unavailable
- [ ] Configuration option for HTTP mode
- [ ] All existing DB tests pass

---

### Feature: Internal A2A + ACP Communication

**Goal:** Enable agent-to-agent and agent-control-plane communication

**Status:** Planned

---

### Feature: Project Management & Per-Project Database

**Goal:** Enable nexora to spawn from any folder and connect to project-scoped resources

**Problem:**
- Current DB structure ties sessions to filesystem location
- Cannot work on a project from different directories
- No support for remote project resources

**Solution:**
- Add project-scoped database schema (projects table with id, name, path, remote_url, metadata)
- CRUD operations for project management
- Project context resolution independent of CWD
- Support both local and remote project resources

**Implementation:**
- New migration: `internal/db/migrations/00X_add_projects.sql`
  - `projects` table (id, name, local_path, remote_url, metadata_json, created_at, updated_at)
  - Add `project_id` FK to sessions table
- New queries: `internal/db/queries/projects.sql`
  - CreateProject, GetProject, ListProjects, UpdateProject, DeleteProject
  - GetProjectByPath, GetProjectByName
- New commands:
  - `nexora project add <name> [--path=.] [--remote=url]`
  - `nexora project list`
  - `nexora project set <name>` (set active project)
  - `nexora project rm <name>`
- Project resolution on startup:
  - Check for active project in config
  - Fall back to CWD-based project detection
  - Create implicit project if none exists
- Remote resource support:
  - Git remote detection
  - HTTP/HTTPS resource URLs
  - Future: SSH, cloud storage

**Database Schema:**
```sql
CREATE TABLE projects (
    id INTEGER PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    local_path TEXT,
    remote_url TEXT,
    metadata_json TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE sessions ADD COLUMN project_id INTEGER REFERENCES projects(id);
```

**Success Criteria:**
- [ ] Projects table and migration created
- [ ] CRUD sqlc queries generated
- [ ] CLI commands for project management
- [ ] Project context resolution on startup
- [ ] Sessions scoped to projects, not CWD
- [ ] Backward compatibility (CWD fallback)
- [ ] Remote URL support in schema
- [ ] All existing tests pass

---

### Other v0.29.4 Features (Planned)

- A2A protocol implementation
- ACP communication layer
- Additional improvements TBD

