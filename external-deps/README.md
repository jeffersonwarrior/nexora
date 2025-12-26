# External Dependencies for Claude Code

**Status:** Active (Auto-starts on boot)

## Claude Code Best Practices (from NellInc gist)

### Hooks (Protection Scripts)
```
hooks/
├── bash-protection.cjs       # Blocks rm -rf, force push, DROP TABLE
├── antipattern-detector.cjs  # Catches stubs, CI weakening, empty handlers
└── suppression-abuse-detector.cjs  # Prevents mass noqa/eslint-disable
```

### Skills (On-Demand, Token-Efficient)
```
skills/
├── verification-before-completion/SKILL.md  # Completion protocol
└── systematic-debugging/SKILL.md            # Four-phase debugging
```

### Token Optimization Directives
```
optimizations/
├── haiku-explore.md    # Model selection guidelines
├── targeted-reads.md   # Surgical file reads
└── batched-edits.md    # Change batching strategy
```

### Memory System
```
memory/
├── diary/              # Session captures
└── REFLECTIONS.md      # Cross-session learnings
```

---

## Context-Engine - Hybrid Code Search

**Services:** Qdrant (6333) + MCP API (8000)
**Systemd Service:** `context-engine.service`

```bash
# Status
sudo systemctl status context-engine.service

# Management
sudo systemctl restart context-engine.service

# Indexing (after code changes)
cd /home/nexora/external-deps/Context-Engine
docker compose run --rm -v /home/nexora:/work indexer --root /work
```

**MCP Config:** `~/.config/claude-code/mcp_servers.json`
```json
{
  "mcpServers": {
    "context-engine": {
      "command": "npx",
      "args": ["-y", "@context-engine-bridge/context-engine-mcp-bridge"],
      "env": {"CTX_API_URL": "http://localhost:8000"}
    }
  }
}
```

## Helix-DB - Graph Vector Database

**CLI:** `/home/agent/.local/bin/helix` (v2.1.10)
**MCP:** Configured (tools: helix_query, helix_check, helix_init)

```bash
cd /home/nexora/.helix
helix init
helix check
helix push dev
```

## claude-mem - Session Memory

**Install:** Requires Claude Code plugin marketplace
```bash
/plugin marketplace add thedotmack/claude-mem
/plugin install claude-mem
```
**Ports:** 37777 (worker + viewer)
**Database:** `~/.claude-mem/claude-mem.db`
