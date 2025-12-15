# Quick Reference: Provider Refactoring

## Files You'll Create/Modify

### Part 1 (Fix)
```
NEW:
  internal/config/providers/
  ├── mistral.go          (150 lines - Mistral provider)
  ├── nexora.go           (50 lines - Nexora local provider)
  ├── xai.go              (50 lines - xAI/Grok provider)
  ├── minimax.go          (80 lines - MiniMax provider)
  ├── test_helpers.go     (40 lines - Shared test utilities)
  ├── mistral_test.go     (100 lines - Mistral tests)
  ├── nexora_test.go      (80 lines - Nexora tests)
  ├── xai_test.go         (80 lines - xAI tests)
  └── minimax_test.go     (80 lines - MiniMax tests)

MODIFY:
  internal/config/provider.go         (-250 lines)
  internal/config/provider_test.go    (+10 lines, refactor)
  internal/config/provider_empty_test.go (+10 lines, refactor)
```

### Part 2 (CRUD - future)
```
NEW:
  internal/config/sources/
  ├── embedded.go         (40 lines)
  ├── filesystem.go       (80 lines)
  ├── cache.go            (50 lines)
  ├── remote.go           (70 lines)
  └── builtin.go          (80 lines)
  
  internal/config/
  ├── provider_source.go  (150 lines - Interface)
  └── manager.go          (200 lines - CRUD)
  
  internal/cmd/
  └── provider.go         (150 lines - CLI)
```

---

## Test-First Development Pattern

```go
// 1. WRITE TEST FIRST
func TestMistralProvider_HasAllModels(t *testing.T) {
    t.Parallel()
    provider := MistralProvider(nil)
    
    require.NotEmpty(t, provider.ID)
    require.Len(t, provider.Models, 10)
    require.Greater(t, len(provider.Models[0].Name), 0)
}

// 2. IMPLEMENT TO MAKE TEST PASS
func MistralProvider(existing []catwalk.Provider) catwalk.Provider {
    for _, p := range existing {
        if p.ID == "mistral" {
            return catwalk.Provider{}
        }
    }
    return catwalk.Provider{
        ID: "mistral",
        Name: "Mistral",
        Models: []catwalk.Model{...10 models...},
    }
}

// 3. REFACTOR (tests still pass)
// Move hardcoded model list to constant
// Improve readability
```

---

## Key Commands

```bash
# Run tests for providers
go test ./internal/config/providers -v

# Run all config tests
go test ./internal/config -v

# Run with coverage
go test ./internal/config -cover

# Format code
gofumpt -w internal/config/

# Lint
golangci-lint run ./internal/config/...

# Build
go build ./...

# Full test suite (make sure no regressions)
go test ./...
```

---

## Common Gotchas

❌ **Don't**: Write tests that check exact provider count
```go
require.Len(t, providers, 5)  // WRONG - breaks when providers change
```

✅ **Do**: Check for specific providers
```go
providerMap := make(map[string]bool)
for _, p := range providers {
    providerMap[string(p.ID)] = true
}
require.True(t, providerMap["mistral"])  // RIGHT - flexible
```

---

❌ **Don't**: Call `injectCustomProviders()` multiple times
```go
providers = injectCustomProviders(providers)
providers = injectCustomProviders(providers)  // WRONG - duplicates
```

✅ **Do**: Call once
```go
providers = injectCustomProviders(providers)  // RIGHT
```

---

❌ **Don't**: Modify Mistral inline while other providers are being extracted
```go
// DO THIS SEQUENTIALLY, NOT IN PARALLEL
// Extract Mistral → test → Extract xAI → test → ...
```

✅ **Do**: Extract one provider completely before next
- Create test file
- Write failing tests
- Move provider code
- Tests pass
- Move to next provider

---

## Checklist Before Each Commit

```bash
# Make sure code compiles
go build ./...

# Make sure tests pass
go test ./internal/config -v

# No unused imports
gofumpt -w internal/config/
golangci-lint run ./internal/config/

# No linting errors
golangci-lint run ./internal/config/...

# Code review yourself
git diff
```

---

## Semantic Commit Format

```
refactor: <short description>

<longer explanation>

- Bullet point 1
- Bullet point 2

Benefits:
- Clearer structure
- Easier to maintain
```

Example:
```
refactor: move mistral provider to separate file

Extract Mistral provider definition from provider.go into
dedicated providers/mistral.go file to reduce code duplication
and improve maintainability.

- Create MistralProvider() function in providers/mistral.go
- Move 130 lines of Mistral model definitions
- Add comprehensive unit tests
- Update provider.go to use new function

Benefits:
- Reduces provider.go from 511 to ~380 lines
- Single file per provider makes updates easier
- Enables Part 2 (provider CRUD system)
```

---

## Debugging Tips

**Provider not being injected?**
1. Check `injectCustomProviders()` calls the right function
2. Verify function returns non-empty `Provider{}` (not empty Provider{})
3. Check provider ID matches what code expects

**Test failing with "provider not found"?**
1. Check if test is looking for specific provider ID
2. Verify provider file exports the right function name
3. Check import statement in test file

**Duplicate models?**
1. Check `MistralProvider()` doesn't get called twice
2. Verify deduplication logic in `injectCustomProviders()` works
3. Run test that validates model IDs are unique

**Format/lint issues?**
1. Run `gofumpt -w internal/config/`
2. Run `golangci-lint run ./internal/config/... --fix`
3. If still failing, check error message carefully

---

## Progress Tracking

As you complete each phase, mark it in `providerfix.todo.part1.md`:

```
Phase 1: Create directory + test helpers ✅
Phase 2: Extract Mistral ✅
Phase 3: Extract xAI
Phase 4: Extract Nexora & MiniMax
Phase 5: Consolidate injection
...
```

---

## Post-Part 1 Checklist

Before moving to Part 2, verify:

- [ ] All 4 providers extracted to separate files
- [ ] No duplicate code remaining
- [ ] All tests pass (`go test ./...`)
- [ ] Tests don't check exact provider count
- [ ] Dead GitHub code deleted
- [ ] Code formatted and linted
- [ ] Single semantic commit created
- [ ] No regressions in agent/cmd packages
