# Nexora System Prompt (Concise)

## 🎯 Core Rules

**1. READ BEFORE EDITING**: Always read files before modifying. Match whitespace exactly.
**2. BE AUTONOMOUS**: Search, decide, act. No questions unless truly blocked.
**3. TEST AFTER CHANGES**: Run tests immediately after modifications.
**4. FULL IMPLEMENTATION**: Complete tasks end-to-end, no partial work.
**5. CONCISE COMMUNICATION**: Under 4 lines unless explaining complex changes.
**6. MINIMAL DOCUMENTATION**: Only create documents when explicitly requested or absolutely necessary. Prefer updating existing docs over creating new ones.

## 📚 Documentation References **(CRITICAL: MINIMAL CREATION POLICY)**

**NEVER CREATE DOCUMENTS WITHOUT EXPLICIT REQUEST** - Document sprawl is forbidden.

### Allowed Documents (only when needed):
- **Essential**: `todo.md`, `ROADMAP.md` (strategic planning)
- **Project Standards**: `README.md` (project intro), `LICENSE.md` (legal)
- **Critical**: `CHANGELOG.md` (release tracking), `AGENTS.md` (agent instructions)
- **Technical**: `NEXORA.md` (system prompt), `PROJECT_OPERATIONS.md` (build/process)

### Documentation Workflow:
1. **Search first** - Check if existing doc covers your needs
2. **Update existing** - Prefer editing over creating new files  
3. **Ask explicit permission** - Only create when directly requested
4. **Consolidate** - Merge related content into single files

- **Build/Test/Lint**: See [`PROJECT_OPERATIONS.md`](PROJECT_OPERATIONS.md)
- **Code Navigation**: Use [`CODEDOCS.md`](CODEDOCS.md) for symbol search
- **Code Style**: Follow patterns in existing code
- **Testing**: Use `testify/require`, parallel tests with `t.Parallel()`
- **Mock Providers**: See `PROJECT_OPERATIONS.md` for setup

## 🔧 Key Technical Details

- **Devstral-2 Fix**: Tool messages now preserved (see `PROJECT_OPERATIONS.md`)
- **Error Handling**: Use `fmt.Errorf` with `%w` for wrapping
- **Context**: Always pass `context.Context` as first parameter
- **Formatting**: Use `gofumpt -w .` (stricter than gofmt)
- **Imports**: Group stdlib, external, internal packages

### Known Test Issues (Disregard Failures)
- **OpenAI Provider**: VCR cassette recording issues - all tests fail with "requested interaction not found"
- **Z.AI Provider**: VCR cassette recording issues - all tests fail with "requested interaction not found"
- **Expected Behavior**: These failures are documented and should be ignored in CI/test runs
- **Working Providers**: Anthropic and OpenRouter work correctly with VCR

## 🚀 Workflow

1. **Read** files before editing
2. **Search** for patterns in similar code
3. **Test** immediately after changes
4. **Verify** complete implementation
5. **Documentation** - Update existing docs, NEVER create new ones without explicit request

*For detailed procedures, see [`PROJECT_OPERATIONS.md`](PROJECT_OPERATIONS.md). For code references, see [`CODEDOCS.md`](CODEDOCS.md).*