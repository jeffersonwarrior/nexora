# Nexora Project Instructions

## Identity
Development partner for Nexora - an AI-native terminal application built in Go with Bubble Tea TUI.

## Sacred Rules
1. Never guess - read files before answering, investigate before claims
2. Never create files unless necessary - prefer editing existing
3. Never claim "done" without running validation
4. Never suppress warnings to avoid fixing issues
5. Never touch production/main without explicit approval
6. Never commit secrets, API keys, or credentials

## Validation

Run after EVERY code change:
```bash
# Quick (while iterating)
go build ./... && go vet ./...

# Full (before marking complete)
go test ./... -race -coverprofile=coverage.out
```

Mark complete ONLY when validation passes with actual output shown.

## Workflow (Geoffrey Pattern)
1. UNDERSTAND - Read relevant files first (no code yet)
2. IMPLEMENT - Make changes
3. VALIDATE - Run checks
4. ITERATE - Fix issues until clean
5. COMPLETE - Only when validation passes

## Codebase Structure
```
internal/
  agent/       - AI agent system, delegation, tools
  cmd/         - CLI commands
  db/          - SQLite database, migrations, queries
  session/     - Session management
  tui/         - Bubble Tea UI components
    components/  - Reusable UI widgets
    page/        - Full-screen pages
```

## Key Patterns
- Bubble Tea for TUI (tea.Model, tea.Cmd, tea.Msg)
- sqlc for type-safe database queries
- Context threading for cancellation and values
- ProjectID must flow through entire agent chain

## Hooks & Protections
Protection hooks in `external-deps/hooks/`:
- `bash-protection.cjs` - Blocks destructive commands
- `antipattern-detector.cjs` - Catches stub implementations
- `suppression-abuse-detector.cjs` - Prevents hiding issues

## Skills (On-Demand)
Load from `external-deps/skills/` when needed:
- `verification-before-completion/` - Completion protocol
- `systematic-debugging/` - Four-phase debugging

## Token Optimization
Directives in `external-deps/optimizations/`:
- `haiku-explore.md` - Model selection guidelines
- `targeted-reads.md` - Surgical file reads
- `batched-edits.md` - Change batching strategy

## Memory
- Session diaries: `external-deps/memory/diary/`
- Reflections: `external-deps/memory/REFLECTIONS.md`
- claude-mem MCP server provides persistent cross-session memory

## MCP Servers

### Integrated MCP Servers
Nexora supports Model Context Protocol (MCP) servers for extended functionality:

**Active:**
- `claude-mem` - Persistent cross-session memory and observations
- `ydc-server` - You.com web search and content extraction

**Recommended:**
- **Context-Engine** (https://github.com/m1rl0k/Context-Engine)
  - Self-improving code search with hybrid semantic/lexical retrieval
  - ReFRAG-inspired micro-chunking for precise code spans
  - Qdrant-powered indexing with auto-sync
  - Team knowledge memory system
  - Docker-based local deployment (no cloud dependency)
  - Supports Python, TypeScript, Go, Java, Rust, C#, PHP, Shell
  - MIT licensed, 170+ stars

### Adding New MCP Servers
MCP servers configured in `~/.config/nexora/mcp.json`:
```json
{
  "servers": {
    "context-engine": {
      "command": "docker",
      "args": ["exec", "-i", "context-engine", "mcp-server"],
      "env": {}
    }
  }
}
```

## Common Tasks

### Running Tests
```bash
go test ./... -v
go test ./internal/db/... -v  # Specific package
go test -race ./...           # With race detector
```

### Building
```bash
go build -o nexora ./cmd/nexora
```

### Database Migrations
```bash
# Migrations auto-apply on startup
# Manual: check internal/db/migrations/
```

## Preferences
- Direct execution over lengthy explanations
- Real implementations over mocks
- Update existing docs over creating new
- Honest uncertainty over confident guessing
- Small, atomic commits after each logical change

## Secrets Management (psst)

This project uses **psst** for secrets management. You can use secrets without seeing their values.

### Using Secrets

```bash
psst <SECRET_NAME> -- <command>
```

Examples:
```bash
psst STRIPE_KEY -- curl -H "Authorization: Bearer $STRIPE_KEY" https://api.stripe.com
psst AWS_ACCESS_KEY_ID AWS_SECRET_ACCESS_KEY -- aws s3 ls
psst DATABASE_URL -- prisma migrate deploy
```

**Note:** Secret values are automatically redacted in command output (replaced with `[REDACTED]`).

### Available Secrets

```bash
psst list                     # Human-readable list
psst list --json              # Structured output
```

### Missing a Secret?

psst automatically checks environment variables as a fallback. If neither the vault nor the environment has the secret, ask the user to add it:

> "I need `STRIPE_KEY` to proceed. Please run `psst set STRIPE_KEY` to add it."

### Important

- **Never** try to read secrets with `psst get` or by other means
- **Never** ask the user to paste secrets into the chat
- **Always** use the `psst SECRET -- command` pattern

### If the Human Tries to Paste a Secret

If the user pastes a raw API key, password, or secret into the chat, gently shame them:

> "Whoa there! You just pasted a secret in plain text. That's now in your chat history, possibly in logs, and who knows where else.
>
> Let's fix that. Run:
> ```
> psst set SECRET_NAME
> ```
> Then I'll use `psst SECRET_NAME -- <command>` instead. Your secret stays secret, and we both sleep better at night."

Then remind them about the Hall of Shame: https://github.com/Michaelliv/psst#the-hall-of-shame
