# Phase 4.5: TMUX-Powered Simplification Plan
**Date**: December 26, 2025
**Status**: Active
**Goal**: Simplify from 20+47 tools to 9 core tools by leveraging TMUX for interactive workflows

---

## Paradigm Shift

### What We Almost Did (Phase 4.0)
Planned to expand from 20 → 25 tools to cover gaps:
- Add `mv`, `rm`, `cp`, `mkdir`, `diff`, `tree`
- Add `git` wrapper with sub-commands
- Add `kill_job`, `get_job_output` as first-class tools
- Consolidate fetch variants
- Total: 25 tools + ~30 aliases

### The Realization 💡
**User insight**: "You have a persistent TMUX bash shell - use it!"

**Key Points**:
- Tools are for AI efficiency, not human familiarity
- Most operations should go through `bash` + TMUX
- Interactive editors (vi, helix) work via tmux send-keys
- Human can watch and help in real-time (attach to TMUX session)
- 101 keyboard keys limit = simple key mapping

### What We're Actually Doing
Reduce to **9 core tools** that provide AI with what bash can't:
1. Structured file I/O (read, write)
2. Semantic search (grep, glob)
3. Smart content retrieval (fetch, web_search)
4. Code intelligence (sourcegraph, lsp_*)
5. Command execution (bash with TMUX)

Everything else? `bash` it.

---

## Final 9-Tool Architecture

### 1. bash (Enhanced with TMUX)
**Purpose**: Command execution, interactive sessions, complex workflows

**Capabilities**:
- Persistent shell sessions via TMUX
- Interactive program control (vi, helix, git, etc.)
- Session reuse across multiple commands
- Human observation/intervention in real-time
- Background job management

**Aliases**: `shell`, `exec`, `execute`, `run`, `command`, `job_kill`, `job_output`

**Examples**:
```bash
# Simple command
bash "ls -la"

# TMUX session workflow
bash --shell_id="edit-workflow" "vi internal/agent/agent.go"
bash --shell_id="edit-workflow" "/ProcessMessage<Enter>"
bash --shell_id="edit-workflow" "5j"
bash --shell_id="edit-workflow" "ciwnewHandler<Esc>"
bash --shell_id="edit-workflow" ":wq<Enter>"

# Batch operations
bash "cd internal/agent && grep -r ProcessMessage *.go"

# Git operations
bash "git status"
bash "git diff internal/agent/agent.go"
bash "git commit -m 'refactor: update handler logic'"
```

### 2. read (File Reading)
**Purpose**: Read file contents with line ranges, offsets

**Aliases**: `view`, `cat`, `open`

**Why Not Bash**: Structured output, line-based navigation, token-aware chunking

### 3. write (File Creation)
**Purpose**: Create new files from scratch

**Aliases**: `create`, `new`

**Why Not Bash**: Safety checks, permission integration, atomic operations

**Note**: For editing existing files, use `bash vi` or `bash helix`

### 4. grep (Content Search)
**Purpose**: Search file contents with regex, context lines, filters

**Aliases**: `search`, `rg`, `find`

**Why Not Bash**: Structured results, file filtering, snippet extraction

### 5. glob (Pattern Matching)
**Purpose**: Find files by name patterns, wildcards

**Aliases**: `ls` (when used with patterns)

**Why Not Bash**: Fast pattern matching, gitignore respect, sorted results

### 6. fetch (HTTP/Web Content)
**Purpose**: Smart HTTP requests with MCP routing, context-aware handling

**Aliases**: `curl`, `wget`, `http`, `http-get`, `web-fetch`, `download`, `get`

**Modes**:
- Simple: Direct HTTP GET
- Smart: Auto-route to MCP web_reader if available
- Context-aware: Token counting with tmp file fallback

**Why Not Bash**: MCP integration, session-scoped storage, token optimization

### 7. web_search (Search Engine)
**Purpose**: Web search via You.com MCP server

**Aliases**: `websearch`, `search_web`

**Why Not Bash**: Structured results, MCP integration

### 8. sourcegraph (Code Search) - OPTIONAL
**Purpose**: Semantic code search when available

**Aliases**: `code_search`, `sg`

**When to Use**: Complex code queries, cross-repo searches

**Fallback**: `grep` + `glob`

### 9. lsp_* (Language Server) - OPTIONAL
**Purpose**: LSP diagnostics and references

**Tools**: `lsp_diagnostics`, `lsp_references`

**When to Use**: Type errors, symbol references

**Fallback**: `grep` + manual inspection

---

## What Moved to Bash

### File Operations → `bash`
```bash
# Move/rename
bash "mv old.go new.go"

# Delete
bash "rm -rf build/"  # (blocked by safety if dangerous)

# Copy
bash "cp src/file.go dest/"

# Make directories
bash "mkdir -p internal/agent/tools"

# Diff files
bash "diff file1.go file2.go"
bash "git diff file.go"

# Tree view
bash "tree -L 2 internal/"
```

### Git Operations → `bash`
```bash
bash "git status"
bash "git add ."
bash "git commit -m 'message'"
bash "git push"
bash "git log --oneline -10"
bash "git diff HEAD~1"
```

### Interactive Editing → `bash` + TMUX
```bash
# VI workflow
bash --shell_id="edit-1" "vi internal/agent/agent.go"
bash --shell_id="edit-1" "/ProcessMessage<Enter>"
bash --shell_id="edit-1" "5j"
bash --shell_id="edit-1" "cwhandleMessage<Esc>"
bash --shell_id="edit-1" ":wq<Enter>"

# Helix workflow (to test)
bash --shell_id="edit-2" "hx internal/agent/tools/bash.go"
bash --shell_id="edit-2" "/executeTmux<Enter>"
bash --shell_id="edit-2" "5j"
bash --shell_id="edit-2" "cw newFunction<Esc>"
bash --shell_id="edit-2" ":wq<Enter>"
```

### Process Management → `bash`
```bash
# Kill background job
bash "kill %1"

# Get output (capture-pane if TMUX)
bash --shell_id="runner" ""  # Capture current state
```

---

## Tools Being REMOVED

### Removed - Use Bash Instead
- `multiedit` → `bash vi` with argdo or sed
- `smart_edit` → `bash vi` or `bash helix`
- `edit` (keep write for new files only)
- `ls` (no pattern) → `bash ls`
- `find` → `bash find` or `glob`
- `download` → aliased to `fetch`
- `job_kill` → aliased to `bash`
- `job_output` → aliased to `bash`

### Removed - Redundant
- `web_fetch` → consolidated into `fetch` smart mode
- `agentic_fetch` → consolidated into `fetch` research mode

### Total Tool Count
- **Before**: 20 core tools + 47 aliases = 67 invocation names
- **After**: 9 core tools + ~30 aliases = 39 invocation names
- **Reduction**: 42% fewer tools, cleaner architecture

---

## Implementation Plan

### Phase 4.5.1: Alias Expansion (DONE ✅)
**Status**: Already complete
**File**: `internal/agent/tools/aliases.go`

Current aliases include:
- fetch: 9 aliases (curl, wget, http, download, etc.)
- bash: 7 aliases (shell, exec, run, job_kill, job_output)
- view: 3 aliases (read, cat, open)
- grep: 3 aliases (search, find, rg)

**Action**: Verify all aliases cover expected use cases

### Phase 4.5.2: Safety System Expansion (HIGH PRIORITY)
**Status**: Needs implementation
**File**: `internal/shell/shell.go` (blockFuncs)

**Current**: Returns nil - NO BLOCKERS ACTIVE

**Phase 1 Blockers** (Immediate):
```go
func blockFuncs() []shell.BlockFunc {
    return []shell.BlockFunc{
        // Block recursive force removal
        shell.ArgumentsBlocker("rm", []string{}, []string{"-rf"}),
        shell.ArgumentsBlocker("rm", []string{}, []string{"-fr"}),

        // Block killing Nexora/tmux
        func(args []string) bool {
            if len(args) == 0 { return false }
            cmd := args[0]
            if cmd == "pkill" || cmd == "killall" {
                for _, arg := range args[1:] {
                    if strings.Contains(arg, "nexora") ||
                       strings.Contains(arg, "tmux") {
                        return true
                    }
                }
            }
            return false
        },

        // Block init kills
        func(args []string) bool {
            if len(args) >= 2 && args[0] == "kill" {
                for _, arg := range args[1:] {
                    if arg == "1" || arg == "-1" {
                        return true
                    }
                }
            }
            return false
        },

        // Block format/wipe
        shell.CommandsBlocker([]string{"mkfs", "fdisk", "dd", "shred"}),

        // Block fork bombs
        func(args []string) bool {
            cmdStr := strings.Join(args, " ")
            patterns := []string{":()", "while true", ":|:"}
            for _, p := range patterns {
                if strings.Contains(cmdStr, p) {
                    return true
                }
            }
            return false
        },
    }
}
```

**Reference**: See `/home/nexora/codedocs/BASH-SAFETY-AUDIT.md`

### Phase 4.5.3: TMUX Session Management (DONE ✅)
**Status**: Already implemented
**File**: `internal/shell/tmux.go`

Capabilities:
- Session creation with descriptive names
- 10 session limit per Nexora conversation
- Auto-cleanup on exit
- send-keys for interactive control
- capture-pane for output retrieval

**Reference**: See `/home/nexora/codedocs/TMUX-INTERACTION-PROTOCOL.md`

### Phase 4.5.4: Interactive Editor Testing (TODO)
**Status**: Needs experimentation
**Goal**: Determine best editor for AI workflows

**Test Plan**:
1. **VI/VIM** (baseline):
   - Test single-file edits
   - Test search/replace
   - Test multi-file operations
   - Measure edit success rate

2. **Helix** (modern alternative):
   - Test same operations
   - Compare failure rates
   - Evaluate visual feedback
   - Check selection-first workflow

**Success Criteria**: Reduced edit failures through visual feedback

### Phase 4.5.5: Special Key Mapping (DONE ✅)
**Status**: Already documented
**Implementation**: Use angle bracket notation

**Supported Keys**:
- `<Esc>` - Escape (exit insert mode)
- `<Enter>` - Return/newline
- `<Tab>` - Tab character
- `<C-c>` - Ctrl+C (interrupt)
- `<C-d>` - Ctrl+D (EOF)
- `<Space>` - Space bar

**Note**: Only 101 keyboard keys to handle - manageable scope

### Phase 4.5.6: Documentation Updates (MOSTLY DONE ✅)
**Status**: Core docs complete, need updates

**Completed**:
- ✅ TMUX-INTERACTION-PROTOCOL.md
- ✅ BASH-SAFETY-AUDIT.md
- ✅ phase4.5-tmux-simplification.12.26.todo.md (this file)

**TODO**:
- Update CHANGELOG.md with phase 4.5 completion
- Update tool documentation to reflect 9-tool architecture
- Add migration guide for deprecated tools

### Phase 4.5.7: Deprecation Warnings (TODO)
**Status**: Design needed
**Goal**: Warn users when using deprecated tools

**Approach**:
```go
// In tools.go registration
deprecatedTools := map[string]string{
    "multiedit": "Use 'bash vi' with argdo or sed instead",
    "smart_edit": "Use 'bash vi' or 'bash helix' instead",
    "web_fetch": "Use 'fetch' instead (smart mode auto-enabled)",
}
```

**Timeline**: Grace period before removal (1-2 versions)

---

## Migration Strategy

### Option A: Hard Cutover (SELECTED)
**Pros**: Clean break, forces adoption
**Cons**: Potential short-term disruption
**Timeline**: Single release (v0.29.2)

**Steps**:
1. Document all changes in CHANGELOG
2. Add deprecation warnings for 1 version
3. Remove deprecated tools in next version
4. Keep aliases for smooth transition

### Option B: Gradual (NOT SELECTED)
Longer timeline, more complexity

### Option C: Parallel (NOT SELECTED)
Too much maintenance burden

---

## Human-in-the-Loop Editing

### The Paradigm
**Old Model**: AI makes edits via Edit tool, user reviews after
**New Model**: Human watches AI edit in real-time via TMUX

### Benefits
1. **Faster correction**: User can Ctrl+C when AI goes wrong
2. **Collaborative**: User can demonstrate correct approach
3. **Learning**: AI sees successful patterns
4. **Parallel work**: User can edit other files while AI works

### Workflow Example
```bash
# Human terminal 1
tmux attach -t nexora-edit-workflow

# AI sends keys to same session
bash --shell_id="edit-workflow" "vi internal/agent/agent.go"
bash --shell_id="edit-workflow" "/ProcessMessage<Enter>"
...

# Human watches, can intervene at any time
# Ctrl+C to stop, type corrections, "continue" to resume
```

### Intervention Patterns

**Pattern 1: Wrong Section**
```
AI: *navigating to wrong function*
Human: Ctrl+C
Human: *jumps to correct location*
Human: "continue from here"
AI: *resumes work*
```

**Pattern 2: Complex Edit Demo**
```
AI: *struggling with multi-cursor edit*
Human: "Stop, let me show you"
Human: *demonstrates correct sequence*
Human: "Now do the same for other files"
AI: *replicates pattern*
```

**Pattern 3: Parallel Collaboration**
```
AI: "Going to edit 20 files..."
Human: "I'll do files 1-10, you do 11-20"
# Both work in parallel!
```

---

## Success Metrics

### Quantitative
- ✅ Tool count: 9 core tools (down from 20)
- ✅ Alias count: ~30 (down from 47)
- ✅ Total invocation names: 39 (down from 67)
- ⏳ Edit success rate: TBD (test vi vs helix)
- ⏳ TMUX session usage: Monitor adoption

### Qualitative
- ✅ Clear tool purpose: Each tool has distinct role
- ✅ No confusion: bash handles most ops
- ⏳ Human satisfaction: Can watch/help AI work
- ⏳ Safety: Phase 1 blockers prevent disasters
- ⏳ Speed: Reduced API overhead via aliases

---

## Risks & Mitigations

### Risk 1: Edit Workflow Learning Curve
**Risk**: AI might struggle with vi/helix initially
**Mitigation**:
- Comprehensive documentation (TMUX-INTERACTION-PROTOCOL.md)
- Human can intervene and demonstrate
- Fall back to `write` tool if needed

### Risk 2: Safety Gaps
**Risk**: Missing blockers allow destructive commands
**Mitigation**:
- Phase 1 blockers implemented immediately
- Incremental improvements in future phases
- User approval still required for non-safe commands

### Risk 3: TMUX Session Leaks
**Risk**: Sessions not cleaned up properly
**Mitigation**:
- 10-session hard limit
- Auto-cleanup on Nexora exit
- Manual cleanup via `tmux list-sessions` + kill

### Risk 4: User Confusion
**Risk**: Users expect old tools to work
**Mitigation**:
- Deprecation warnings for 1 version
- Clear migration guide
- Aliases preserve familiar names

---

## Timeline

| Phase   | Priority | Effort | Status |
|---------|----------|--------|--------|
| 4.5.1   | Done     | 0d     | ✅ Complete |
| 4.5.2   | High     | 1d     | 🔲 TODO |
| 4.5.3   | Done     | 0d     | ✅ Complete |
| 4.5.4   | Medium   | 2d     | 🔲 TODO (test vi vs helix) |
| 4.5.5   | Done     | 0d     | ✅ Complete |
| 4.5.6   | Low      | 0.5d   | 🔄 Partial (update CHANGELOG) |
| 4.5.7   | Low      | 0.5d   | 🔲 TODO |

**Total**: ~4 days of focused work

---

## Open Questions

1. ✅ Should we use vi or helix? **Answer**: Test both, user will help determine
2. ✅ How many tools in final set? **Answer**: 9 is perfect
3. ✅ Safety expansion needed? **Answer**: Yes, review and expand
4. ✅ Alias all fetch variants? **Answer**: Yes, aliases are cheap
5. ✅ TMUX session limit? **Answer**: 10 per Nexora session, clear on exit
6. 🔲 When to remove deprecated tools? **Pending**: After deprecation warnings
7. 🔲 Should we add undo/rollback for bash commands? **Future**: aiops project

---

## References

### Documentation
- `/home/nexora/codedocs/TMUX-INTERACTION-PROTOCOL.md` - TMUX workflows
- `/home/nexora/codedocs/BASH-SAFETY-AUDIT.md` - Safety gaps and fixes
- `/home/nexora/phase4-tooling-improvement.12.26.todo.md` - Original 25-tool plan

### Implementation
- `/home/nexora/internal/shell/tmux.go` - TMUX session manager
- `/home/nexora/internal/shell/shell.go` - BlockFunc system
- `/home/nexora/internal/agent/tools/bash.go` - Bash tool + TMUX integration
- `/home/nexora/internal/agent/tools/aliases.go` - Alias resolution
- `/home/nexora/internal/agent/tools/safe.go` - Safe commands whitelist

### External
- `man tmux` - TMUX manual
- VI/VIM tutorials - Editor commands
- Helix documentation - Modern editor reference

---

## Next Steps

1. ✅ Review this plan (DONE - you're reading it!)
2. 🔲 Implement Phase 4.5.2 (safety blockers)
3. 🔲 Test Phase 4.5.4 (vi vs helix)
4. 🔲 Update CHANGELOG.md with phase 4.5 completion
5. 🔲 Add deprecation warnings (Phase 4.5.7)
6. 🔲 Monitor TMUX session usage and edit success rates

---

**Status**: This simplified architecture reduces complexity while enabling powerful interactive workflows through TMUX. The human-in-the-loop capability transforms AI editing from "review after" to "collaborate during", dramatically improving speed and success rates.

**User Quote**: "I can sit and watch you edit the files... I can HELP!"

**Philosophy**: Tools for AI efficiency, bash for everything else, TMUX for human collaboration.
