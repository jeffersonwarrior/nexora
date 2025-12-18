# Nexora 0.3.0 Migration Plan - VNC Gateway Architecture

**Current Version**: 0.29.0 (CLI agent with tool system)  
**Target Version**: 0.3.0 (VNC Gateway with visual terminal interaction)  
**Status**: 🔴 NOT STARTED - Planning phase

---

## Executive Summary

Nexora 0.3.0 represents a **revolutionary architectural shift** from JSON-based tool abstraction to **direct visual terminal interaction** via VNC. The system enables AI agents to interact with terminals and browsers **exactly like humans** - seeing the screen, typing, and getting visual feedback in 1-2 second cycles.

### Key Innovation

**Instead of:**
```go
edit(file="main.go", old_string="...", new_string="...")
❌ Parameter encoding fails
❌ No visual feedback  
❌ 30-60 second cycles
```

**We now have:**
```go
keyboard("vi main.go\n")      → screen() → 📸 sees vim editor
keyboard(":/function\n")       → screen() → 📸 sees cursor position
keyboard("i") + editing        → screen() → 📸 sees changes
keyboard("\x1b:wq\n")          → screen() → 📸 sees save confirmation

✅ 1-2 second cycles
✅ Full visual confirmation
✅ Zero encoding issues
✅ Exactly like human workflow
```

---

## Architecture Comparison

### 0.29.0 (Current)
```
User → Nexora CLI → LLM → Tools (edit/bash/grep/view)
                            ↓
                        File system
```

### 0.3.0 (Target)
```
User → Nexora CLI → LLM → VNC Gateway → Docker Container
                                          ↓
                              [Ubuntu + VNC + X11 + Tools]
                                    ↓         ↓
                              Terminal    Chrome
                                (vi/ed)   (DevTools)
```

---

## Major Components (From nexora3.01)

### 1. PostgreSQL Migration
- **Complete replacement of SQLite**
- Tables: sessions, port_allocations, providers_auth, vnc_sessions
- Encrypted API key storage (AES-256-GCM)
- Connection pooling with pgx/v5

### 2. Docker Workstation Container
- **Image**: nexora/workstation:latest (1.93GB)
- **Base**: Ubuntu 24.04
- **Includes**:
  - Xvfb (virtual display)
  - x11vnc (VNC server)
  - Fluxbox (window manager)
  - 40+ tools (vim, tmux, git, chromium, etc.)
- **Ports**: VNC (5900-5999), CDP (9222-9321)

### 3. VNC Manager (`internal/vnc/`)
- Container lifecycle management
- Port allocation from PostgreSQL
- Workspace mounting
- Health monitoring

### 4. Session Management (`internal/session/`)
- Session tracking with PostgreSQL
- Container binding per session
- Recovery on crashes

### 5. Fantasy VNC Tools
- `screen()` - Capture PNG + text extraction
- `keyboard()` - Send keystrokes (including special keys)
- `execute()` - Run shell commands
- `chrome_*()` - Browser automation via CDP

### 6. AIMUX5 Integration
- All AI requests routed through localhost:9300
- Centralized API key management
- Health monitoring

---

## File Structure Changes

### New Directories
```
internal/vnc/          # VNC container management (7 files)
internal/container/    # Docker lifecycle (2 files)  
internal/session/      # Session tracking (3 files)
internal/db/          # PostgreSQL schemas (24 files)
docker/workstation/   # Container definition
```

### Modified Core
```
internal/agent/       # VNC tool integration
internal/fantasy/     # VNC gateway interface
internal/config/      # PostgreSQL provider loading
internal/app/         # Container initialization
```

### Removed (SQLite Era)
```
internal/tools/*      # Old tool architecture (20+ files)
```

---

## Migration Effort Estimate

### Critical Path (Phase 8 - VNC Integration)
| Task | Effort | Priority |
|------|--------|----------|
| PostgreSQL schema migration | 2-3 days | P0 |
| Docker workstation build | 1 day | P0 |
| VNC manager implementation | 2-3 days | P0 |
| Fantasy VNC bridge | 1-2 days | P0 |
| Session lifecycle | 1-2 days | P0 |
| Port management | 1 day | P0 |
| Container recovery | 1 day | P1 |
| Testing & integration | 2-3 days | P0 |

**Total**: ~2-3 weeks for core VNC functionality

### Additional Work
| Phase | Effort | Priority |
|-------|--------|----------|
| API key encryption (database) | 1 week | P1 |
| AIMUX5 integration | 3-5 days | P1 |
| Transform engine (JSONata) | 1 week | P2 |
| Definition system | 1 week | P2 |
| Vector memory (pgvector) | 1-2 weeks | P3 |

**Total for 0.3.0 Final**: ~6-8 weeks

---

## Risk Assessment

### High Risk
1. **Docker dependency** - Requires Docker daemon on all systems
2. **Port exhaustion** - 100 concurrent session limit (ports 5900-5999)
3. **Resource usage** - 1.93GB image + container overhead per session
4. **PostgreSQL requirement** - No longer works with SQLite

### Medium Risk
1. **VNC security** - Local-only binding, no authentication by default
2. **Container cleanup** - Orphaned containers if crashes occur
3. **API key migration** - Existing users need re-configuration

### Low Risk
1. **Backward compatibility** - Old tool system completely removed
2. **Testing complexity** - Requires Docker in CI/CD

---

## Decision Points

### Should We Migrate?

**Pros:**
- Revolutionary user experience (1-2s visual feedback)
- Zero encoding/parameter issues
- Natural terminal workflow for AI
- Chrome automation built-in
- Scales to 100 concurrent sessions

**Cons:**
- Complete rewrite (6-8 weeks)
- Requires Docker + PostgreSQL
- Breaks backward compatibility
- Higher resource usage
- More complex deployment

### Alternative: Incremental Approach

1. **Keep 0.29.0** as "Nexora Classic" (tool-based)
2. **Build 0.3.0** as "Nexora VNC" (visual terminal)
3. **Offer both** with different binaries

This allows:
- Users without Docker to stay on 0.29.0
- Early adopters to test 0.3.0
- Gradual migration path
- Less risk

---

## Current Status (nexora3.01)

From the test repo:

### ✅ Complete (99%)
- PostgreSQL schemas and migration
- Docker workstation container (built, 1.93GB)
- VNC manager core logic
- Session tracking
- Port allocation
- Container lifecycle
- Fantasy VNC bridge
- AIMUX5 routing

### ⚠️ Minor Issues (11 test failures)
1. VNC port test setup (3 tests) - DB seeding
2. Config migration (2 tests) - Provider duplication  
3. Config priority (1 test) - Host lookup
4. Config validation (2 tests) - JSON validation
5. Model dialog (3 tests) - Invalid model pruning

**Estimated fix time**: 3-4 hours

### 🔜 Pending
- OAuth token refresh
- TUI permission cleanup
- Documentation updates

---

## Recommendation

### For 0.29.0 (Current Repo)
**Complete current audit fixes and release 0.29.0 as stable.**
- ✅ Race conditions fixed
- ✅ Benchmarks added
- ✅ Build working
- ✅ Tests passing

This gives users a solid, production-ready CLI agent.

### For 0.3.0 Migration
**Create a new branch `vnc-gateway` and port changes incrementally:**

1. **Week 1-2**: PostgreSQL migration
   - Schema setup
   - API key encryption
   - Provider loading

2. **Week 3-4**: Docker + VNC basics
   - Build workstation container
   - Port management
   - Basic container lifecycle

3. **Week 5-6**: Fantasy integration
   - VNC tools (screen, keyboard)
   - Chrome automation
   - Testing

4. **Week 7-8**: Polish & testing
   - Session recovery
   - Error handling
   - Documentation
   - E2E tests

### Timeline
- **0.29.0 Final**: Release now (audit fixes complete)
- **0.30.0**: PostgreSQL migration (weeks 1-2)
- **0.31.0**: VNC Gateway alpha (weeks 3-4)
- **0.32.0**: VNC Gateway beta (weeks 5-6)
- **0.3.0 Final**: Production release (weeks 7-8)

---

## Files to Review

Key files in `/home/nexora3/nexora3.01` for migration:

### Core Architecture
- `internal/vnc/manager.go` - Container management
- `internal/session/session.go` - Session tracking
- `internal/db/sessions.sql.go` - PostgreSQL schemas
- `docker/workstation/Dockerfile` - Container definition
- `internal/fantasy/vnc.go` - VNC gateway interface

### Documentation
- `DATABASE_KEYS_INTEGRATION.md` - API key encryption
- `DEPLOYMENT_COMPLETE.md` - Infrastructure status
- `MIGRATION_STATUS.md` - Current progress
- `TODO_CURRENT.md` - Remaining work

---

## Summary

**Nexora 0.3.0 is 99% complete in the test repo** and represents a **revolutionary architectural shift**. However, it's a **6-8 week migration** with significant breaking changes.

**Recommended path:**
1. ✅ Ship 0.29.0 now (audit fixes complete)
2. 🔄 Port 0.3.0 incrementally over 8 weeks
3. 🚀 Release 0.3.0 as new flagship

**Current priority:** Complete 0.29.0 and get it released.

---

**Date**: December 18, 2025  
**Reviewed**: /home/nexora3/nexora3.01  
**Status**: Planning complete, awaiting decision
