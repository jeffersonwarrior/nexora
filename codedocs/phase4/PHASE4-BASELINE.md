# Phase 4 Tool Consolidation - Baseline State

**Date:** 2025-12-26T07:52:00Z
**Status:** Pre-Consolidation (Clean Baseline)

## Current State

### Tool Files Count
- Total implementation files: 47
- Total test files: 31
- Backup files moved to archive: 3

### Files to Consolidate

#### Bash Tools (2 files → 1 file)
- `internal/agent/tools/bash.go` (343 lines) - standard bash execution
- `internal/agent/tools/bash_monitored.go` (90 lines) - AI-monitored execution
- **Target:** Merge into single `bash.go` with mode detection

#### Fetch Tools (3 files → 1 file)
- `internal/agent/tools/fetch.go` (187 lines) - basic fetch
- `internal/agent/tools/web_fetch.go` (72 lines) - web-specific fetch
- `internal/agent/agentic_fetch_tool.go` (237 lines) - agentic fetch
- **Target:** Merge into single `fetch.go` with format/mode options

#### Agent Tools (4 files → 1 file)
- `internal/agent/tools/agents.go` - agent management
- `internal/agent/tools/agent_list.go` - list agents
- `internal/agent/tools/agent_status.go` - get agent status
- `internal/agent/tools/agent_run.go` - run agent
- `internal/agent/tools/delegate.go` (5266 lines) - delegation (KEEP)
- **Target:** Merge functionality into `delegate.go` with action parameter

#### Analytics Tools (0 files - already removed)
- ✅ No analytics tools found (already cleaned up)

#### Aliasing System
- ✅ `internal/agent/tools/aliases.go` (exists, needs testing)
- ✅ `internal/agent/tools/aliases_test.go` (exists)

### Archived Files
Moved to `archives/phase4-incomplete-rollback/`:
- `bash_consolidated.go.backup`
- `fetch_consolidated.go.backup`
- `delegate_enhanced.go`

### Test Status
Current tool tests: **PASSING**
- Bash tests: ✅ PASS
- Fetch tests: Need verification
- Delegate tests: ✅ PASS
- Aliases tests: Need verification

## Consolidation Plan

### Step 1: Bash Consolidation
1. Create tests for both modes
2. Merge `bash_monitored.go` into `bash.go`
3. Add mode detection (purpose + completion_criteria → monitored mode)
4. Delete `bash_monitored.go`
5. Verify all tests pass

### Step 2: Fetch Consolidation
1. Create tests for all formats/modes
2. Merge all three fetch files into `fetch.go`
3. Add format parameter (text, markdown, html)
4. Add mode parameter (web_reader, raw, agentic)
5. Delete old files
6. Verify all tests pass

### Step 3: Agent Tools Consolidation
1. Create tests for all actions
2. Add action parameter to `delegate.go` (spawn, list, status, stop, run)
3. Migrate functionality from agent_*.go files
4. Delete old files
5. Verify all tests pass

### Step 4: Aliasing Integration
1. Test alias mapping
2. Integrate into tool dispatch
3. Verify backward compatibility

## Success Criteria

- Total tool files reduced from 47 to ~40 (7 files removed)
- All tests pass
- Backward compatibility via aliases
- No functionality regression
