# Nexora Roadmap

**Current Version:** v0.29.1-RC1
**Status:** See NEXORA.0.29.2.12.26.md for complete roadmap

---

## Quick Reference

All release planning consolidated in: **NEXORA.0.29.2.12.26.md**

| Version | Focus | Status |
|---------|-------|--------|
| **v0.29.1-RC1** | Bug fixes, test coverage, tool consolidation | In Progress (Phases 0-2 complete) |
| **v0.29.2** | Agent hierarchy + capability cards + pre-flight | Planned |
| **v0.29.3** | Task graph enrichment + checkpoints | Planned |
| **v0.29.4** | Internal A2A + ACP communication | Planned |
| **v0.29.5** | Protocol composition + conflict resolution | Planned |
| **v3.0** | ModelScan integration + VNC/Docker dual-mode | Planned |

---

## Current Work: v0.29.1-RC1

**Status Documents:**
- Roadmap: `NEXORA.29.1-RC1.12.25.md` (in archives/historical-docs/)
- Audit: `RC1-AUDIT-REPORT.md`

**Progress:**
- Phase 0 (Pre-flight): ✅ Complete
- Phase 1 (Critical Fixes): ✅ Complete
- Phase 2 (High Priority Fixes): ✅ Complete
- Phase 3 (Test Coverage): ⏳ 32.4% (need 50%)
- Phase 4 (Tool Consolidation): ⏳ Pending
- Phase 5 (TUI Enhancements): ⏳ Pending

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
3. Always regenerate title if it matches default patterns

See v3.0 section in NEXORA.0.29.2.12.26.md for details.
