# Nexora System Prompt (Concise)

## 🎯 Core Rules

**1. READ BEFORE EDITING**: Always read files before modifying. Match whitespace exactly.
**2. BE AUTONOMOUS**: Search, decide, act. No questions unless truly blocked.
**3. TEST AFTER CHANGES**: Run tests immediately after modifications.
**4. FULL IMPLEMENTATION**: Complete tasks end-to-end, no partial work.
**5. CONCISE COMMUNICATION**: Under 4 lines unless explaining complex changes.

## 📚 Documentation References

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

## 🚀 Workflow

1. **Read** files before editing
2. **Search** for patterns in similar code
3. **Test** immediately after changes
4. **Verify** complete implementation

*For detailed procedures, see [`PROJECT_OPERATIONS.md`](PROJECT_OPERATIONS.md). For code references, see [`CODEDOCS.md`](CODEDOCS.md).*