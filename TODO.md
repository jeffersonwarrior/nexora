# Nexora 3.0 Implementation Todo

**Target**: Complete visual terminal interaction with dual-mode support  
**Focus**: ModelScan integration first, then VNC implementation  
**Timeline**: 11 weeks total  

---

## 🐛 Known Issues

### Session Title Re-generation
**Issue**: Sessions that already have "New Session" as title are not retitled when first message is sent.

**Current Behavior**:
- New sessions get "New Session" as default title
- When first message is sent, `generateTitle()` runs but only updates the session object
- If the session already exists with "New Session", it's not properly detected as needing a title

**Expected Behavior**:
- First message should always generate a proper title
- "New Session" should be treated as placeholder that needs replacement

**Root Cause**:
- `generateTitle()` checks `MessageCount == 0` but doesn't check if current title is placeholder
- Race condition possible: session created 	 title set to "New Session" 	 message added 	 count becomes 1

**Possible Solutions**:
1. Check both `MessageCount == 0 OR title == "New Session"` 
2. Add `needs_title` boolean flag to session schema
3. Check if title equals any default placeholder values
4. Always regenerate title if it matches default patterns

**Priority**: Medium (UX issue, not blocking)

---

## Phase 0: ModelScan Integration Priority (Week 0-2)

### Rethinking: ModelScan First
Replace Fantasy with ModelScan BEFORE adding VNC to:
- Clean up provider architecture 
- Reduce immediate user pain
- Less moving parts when adding VNC later

### Week 0: Pre-Integration
- [ ] Remove existing provider hardcoding
- [ ] Audit Fantasy vs ModelScan API differences  
- [ ] Create ModelScan provider wrapper interface
- [ ] Update configuration system for ModelScan routing

### Week 1: Core Integration
- [ ] Replace Fantasy client with ModelScan router
- [ ] Implement ModelScan provider selection in config
- [ ] Add routing options (cheapest, fastest, fallback)
- [ ] Update error handling for ModelScan responses

### Week 2: Validation
- [ ] Test with all supported providers (OpenAI, Anthropic, Mistral, xAI)
- [ ] Verify streaming responses still work
- [ ] Add performance metrics for routing decisions
- [ ] Fix any broken tools that depend on provider specifics

---

## Phase 1: Database Foundation (Week 2-3)

### PostgreSQL Implementation
- [ ] Install PostgreSQL schemas (sessions, port_allocations, providers_auth, vnc_sessions)
- [ ] Create PostgreSQL connection pool with pgx/v5
- [ ] Migrate session tracking from SQLite
- [ ] Implement API key encryption (AES-256-GCM)
- [ ] Add database migration utilities

### SQLite Support (Lite Mode)
- [ ] Create SQLite version of same schemas  
- [ ] Implement dual database abstraction layer
- [ ] Add database type detection in config
- [ ] Test both databases with same operations

---

## Phase 2: Docker Infrastructure (Week 3-4)

### Container Build
- [ ] Create Ubuntu 24.04 Dockerfile with all dev tools
- [ ] Add ARM-specific adjustments for Mac Silicon
- [ ] Optimize image size (target under 2GB)
- [ ] Create startup scripts (X11, VNC, Chrome)
- [ ] Implement health checks and status signals

### Container Management
- [ ] Build Docker lifecycle manager
- [ ] Implement port allocation (VNC: 5900-5999, CDP: 9222-9321)
- [ ] Add workspace mounting with user permissions
- [ ] Create container startup sequence with proper error handling
- [ ] Implement orphaned container cleanup

---

## Phase 3: VNC Tools Implementation (Week 5-6)

### Screen Capture
- [ ] Implement VNC client connection
- [ ] Capture framebuffer and convert to PNG
- [ ] Add basic text extraction (ocr)
- [ ] Optimize capture frequency (100-200ms)
- [ ] Handle screen resize and resolution changes

### Keyboard Input
- [ ] Implement xdotool integration for typing
- [ ] Support special keys (Escape, Tab, Ctrl+*)
- [ ] Handle modifier keys correctly
- [ ] Add input validation and sanitization
- [ ] Implement typing rate limiting

### Execute Command
- [ ] Add direct docker exec for non-visual operations
- [ ] Implement stdout/stderr capture
- [ ] Add exit code tracking
- [ ] Handle long-running commands
- [ ] Add command timeout management

### Tool Integration
- [ ] Create fantasy tools wrapper for VNC operations
- [ ] Integrate ModelScan with VNC tools
- [ ] Update agent prompts to use screen/keyboard/execute
- [ ] Add fallback to old tools for Lite mode

---

## Phase 4: Session Management (Week 7-8)

### Session Lifecycle
- [ ] Implement session start with database tracking
- [ ] Add container binding per session
- [ ] Create session state persistence
- [ ] Implement graceful session termination
- [ ] Add session recovery after crashes

### Mode Selection
- [ ] Build installer with mode selection (Lite/Full)
- [ ] Implement mode detection in code paths
- [ ] Add configuration validation per mode
- [ ] Create mode-specific help messages
- [ ] Test switching requires reinstall (no in-place upgrade)

### Error Handling
- [ ] Add container crash detection
- [ ] Implement automatic container restart
- [ ] Create session state restoration
- [ ] Add port conflict resolution
- [ ] Handle Docker daemon failures

---

## Phase 5: Integration & Testing (Week 9-10)

### End-to-End Testing
- [ ] Test complete workflows in both modes
- [ ] Verify ModelScan routing works with VNC
- [ ] Test concurrent sessions (up to 10)
- [ ] Validate resource cleanup
- [ ] Test ARM vs x86 architecture differences

### Performance Validation
- [ ] Measure screen capture latency
- [ ] Validate VNC vs tool speed improvements
- [ ] Test ModelScan routing performance
- [ ] Monitor memory usage per session
- [ ] Verify resource limits enforcement

### Documentation
- [ ] Create unified README.md with mode descriptions
- [ ] Document VNC justification (examples of broken tools)
- [ ] Add troubleshooting guide
- [ ] Create installation guide with mode selection
- [ ] Document ModelScan configuration options

---

## Phase 6: Release Preparation (Week 11)

### Final Integration
- [ ] Remove legacy Fantasy code entirely
- [ ] Clean up unused tool abstractions
- [ ] Finalize configuration defaults
- [ ] Add startup diagnostics
- [ ] Implement version checking

### Packaging & Distribution
- [ ] Build release binaries for all platforms
- [ ] Create Docker image for easy deployment
- [ ] Add automated tests to CI
- [ ] Prepare release notes
- [ ] Update GitHub releases

---

## Critical Path & Dependencies

### Blockers (must complete first):
1. **ModelScan integration (Phase 0)** - Nothing works without it
2. **Database layer (Phase 1)** - Required for VNC session management  
3. **Docker infrastructure (Phase 2)** - No VNC without containers

### Parallelizable Work:
- SQLite compatibility (alongside PostgreSQL)
- Tool migration (works with both databases)
- Documentation (can be written incrementally)

### Timeline Summary:
- **Week 0-2**: ModelScan integration (HIGH PRIORITY)
- **Week 3-4**: Database + Docker foundations  
- **Week 5-6**: Core VNC implementation
- **Week 7-8**: Session management
- **Week 9-10**: Integration testing
- **Week 11**: Release prep

**Total: 11 weeks to 3.0 release**

---

## Architecture Notes

### ModelScan Configuration Migration
```yaml
# Old config (multiple providers)
providers:
  mistral: {api_key: "..."}
  openai: {api_key: "..."}

# New config (ModelScan routing)
modelscan:
  enabled: true
  routing: "cheapest"  # or fastest, balanced, fallback
  providers:
    mistral: {api_key: "...", priority: 1, cost: 0.15}
    openai: {api_key: "...", priority: 2, cost: 0.30}
```

### Why ModelScan First
- Reduces risk - immediate user value even if VNC delayed
- Cleaner architecture for VNC development
- Faster provider switching and cost optimization
- Foundation for 4.0 multi-agent architecture

### Mode Summary
**Lite Mode**: SQLite + ModelScan + CLI tools (servers, laptops)
**Full Mode**: PostgreSQL + ModelScan + VNC (visual pair-programming)