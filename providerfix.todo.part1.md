# Provider System Refactoring - Part 1 TODO

## Objective
Consolidate and organize hardcoded provider definitions in `internal/config/provider.go` into separate files, reducing duplication and improving maintainability.

**Effort**: ~2-3 hours  
**Target**: Zero code duplication, provider-agnostic tests, cleaner injection logic

---

## TDD Strategy

### Test-First Approach
1. **Write failing tests first** for new provider files
2. **Implement provider functions** to make tests pass
3. **Refactor injection logic** with tests passing throughout
4. **Run full test suite** after each major change

### Test Organization
- One test file per provider file (`providers/mistral_test.go`, etc.)
- Shared test helpers in `providers/test_helpers.go`
- Provider-agnostic tests in `provider_test.go` (don't check provider count)

---

## TODO Checklist

### Phase 1: Create Provider Files Structure (0.5 hours)

- [ ] **1.1** Create `internal/config/providers/` directory
  ```bash
  mkdir -p internal/config/providers
  ```

- [ ] **1.2** Create `internal/config/providers/test_helpers.go`
  - Helper to create provider IDs for testing
  - Helper to validate provider structure (has required fields)
  - Helper to count models per provider
  ```go
  package providers
  
  import (
      "testing"
      "github.com/charmbracelet/catwalk/pkg/catwalk"
      "github.com/stretchr/testify/require"
  )
  
  func requireValidProvider(t *testing.T, p catwalk.Provider) {
      t.Helper()
      require.NotEmpty(t, p.ID, "Provider must have ID")
      require.NotEmpty(t, p.Name, "Provider must have Name")
      require.NotEmpty(t, p.Type, "Provider must have Type")
  }
  ```

---

### Phase 2: Extract Mistral Provider (0.75 hours)

- [ ] **2.1** Write tests for Mistral provider in `internal/config/providers/mistral_test.go`
  ```go
  // Tests to write:
  // - TestMistralProvider_ReturnsValidProvider
  // - TestMistralProvider_HasAllModels (verify 10 models)
  // - TestMistralProvider_SkipsIfAlreadyExists
  // - TestMistralProvider_SetAPIEndpointFromEnv
  // - TestMistralProvider_ModelPricing (spot-check costs)
  ```

- [ ] **2.2** Create `internal/config/providers/mistral.go`
  - Move lines 213-338 from `provider.go` → `providers/mistral.go`
  - Create `MistralProvider(existingProviders []catwalk.Provider) catwalk.Provider` function
  - Keep duplicate check logic: `for _, p := range existingProviders { if p.ID == "mistral" { return catwalk.Provider{} } }`
  - Verify tests pass: `go test ./internal/config/providers -run TestMistral -v`

- [ ] **2.3** Update `provider.go` to import and use `MistralProvider`
  - Remove old `injectMistralProvider()` function
  - Don't call it yet (will consolidate in Phase 5)

---

### Phase 3: Extract xAI Provider (0.5 hours)

- [ ] **3.1** Write tests for xAI provider in `internal/config/providers/xai_test.go`
  ```go
  // Tests to write:
  // - TestXAIProvider_ReturnsValidProvider
  // - TestXAIProvider_HasGrokModel
  // - TestXAIProvider_SkipsIfAlreadyExists
  // - TestXAIProvider_SetAPIEndpointFromEnv
  ```

- [ ] **3.2** Create `internal/config/providers/xai.go`
  - Move lines 388-410 from `provider.go` → `providers/xai.go`
  - Create `XAIProvider(existingProviders []catwalk.Provider) catwalk.Provider` function
  - Verify tests pass: `go test ./internal/config/providers -run TestXAI -v`

---

### Phase 4: Extract Nexora & MiniMax Providers (0.5 hours)

- [ ] **4.1** Write tests for Nexora provider in `internal/config/providers/nexora_test.go`
  ```go
  // Tests to write:
  // - TestNexoraProvider_ReturnsValidProvider
  // - TestNexoraProvider_UsesLocalhostEndpoint
  // - TestNexoraProvider_SkipsIfAlreadyExists
  ```

- [ ] **4.2** Create `internal/config/providers/nexora.go`
  - Move lines 352-374 from `provider.go` → `providers/nexora.go`
  - Create `NexoraProvider(existingProviders []catwalk.Provider) catwalk.Provider` function
  - Verify tests pass: `go test ./internal/config/providers -run TestNexora -v`

- [ ] **4.3** Write tests for MiniMax provider in `internal/config/providers/minimax_test.go`
  ```go
  // Tests to write:
  // - TestMiniMaxProvider_ReturnsValidProvider
  // - TestMiniMaxProvider_HasTwoModels
  // - TestMiniMaxProvider_UsesAnthropicType
  // - TestMiniMaxProvider_SkipsIfAlreadyExists
  ```

- [ ] **4.4** Create `internal/config/providers/minimax.go`
  - Move lines 424-459 from `provider.go` → `providers/minimax.go`
  - Create `MiniMaxProvider(existingProviders []catwalk.Provider) catwalk.Provider` function
  - Verify tests pass: `go test ./internal/config/providers -run TestMiniMax -v`

---

### Phase 5: Consolidate Injection Logic (0.5 hours)

- [ ] **5.1** Write test for consolidated injection in `internal/config/provider_test.go`
  ```go
  // Add test:
  // - TestInjectCustomProviders_CallsAllInjectors
  // - TestInjectCustomProviders_CountsInjected
  // - TestInjectCustomProviders_SkipsDuplicates
  // - TestInjectCustomProviders_ReturnsAllProviders
  ```

- [ ] **5.2** Refactor `injectCustomProviders()` in `provider.go`
  ```go
  // Replace 4 if-statements with loop:
  injectors := []func([]catwalk.Provider) catwalk.Provider{
      providers.MistralProvider,
      providers.NexoraProvider,
      providers.XAIProvider,
      providers.MiniMaxProvider,
  }
  
  injectedCount := 0
  for _, injector := range injectors {
      if p := injector(providers); p.ID != "" {
          providers = append(providers, p)
          injectedCount++
      }
  }
  ```

- [ ] **5.3** Delete individual `injectMistralProvider()`, `injectNexoraProvider()`, `injectXAIProvider()`, `injectMiniMaxProvider()` function definitions
  - Lines to delete: 206-210, 213-340, 344-349, 352-376, 380-386, 388-412, 416-422, 424-459

- [ ] **5.4** Verify all tests still pass
  ```bash
  go test ./internal/config -v
  go test ./internal/config/providers -v
  ```

---

### Phase 6: Delete Dead Code (0.1 hours)

- [ ] **6.1** Delete `injectCustomGitHubProviders()` function (lines 463-490)
  - This is dead code (never called)
  - Will be re-implemented in Part 2 properly

- [ ] **6.2** Delete `fetchProvidersFromGitHub()` function (lines 492-510)
  - Supporting function for dead code above
  - Will be re-implemented in Part 2

- [ ] **6.3** Verify file ends properly (no orphaned code)
  - Check last line should be around line 460 now

---

### Phase 7: Update Tests (0.5 hours)

**Goal**: Make tests provider-count agnostic (don't hardcode `require.Len(t, providers, 5)`)

- [ ] **7.1** Update `internal/config/provider_test.go`
  - **Find**: Lines checking `require.Len(t, providers, 5)`
  - **Replace with**: Provider-presence checks instead of count checks
  ```go
  // OLD (brittle):
  require.Len(t, providers, 5) // mock + mistral + nexora + xai + minimax
  
  // NEW (flexible):
  providerMap := make(map[string]bool)
  for _, p := range providers {
      providerMap[string(p.ID)] = true
  }
  require.True(t, providerMap["mistral"], "Expected Mistral provider")
  require.True(t, providerMap["mock"], "Expected mock provider")
  // Don't assert exact count, just that required providers exist
  ```

- [ ] **7.2** Update `internal/config/provider_empty_test.go`
  - Same approach: check for provider presence, not exact count
  - Verify xAI and MiniMax are in the list (newly injected)

- [ ] **7.3** Add new tests in `internal/config/providers/providers_test.go`
  ```go
  // Test all providers together:
  // - TestAllProvidersCanCoexist (no ID conflicts)
  // - TestProviderPriorityPreserved (order matters)
  // - TestProviderModelsValid (all models have required fields)
  ```

- [ ] **7.4** Run all tests
  ```bash
  go test ./internal/config -v
  go test ./internal/config/providers -v
  ```

---

### Phase 8: Code Quality (0.5 hours)

- [ ] **8.1** Format code with gofumpt
  ```bash
  gofumpt -w internal/config/provider.go
  gofumpt -w internal/config/providers/
  ```

- [ ] **8.2** Run linters
  ```bash
  golangci-lint run ./internal/config/...
  ```

- [ ] **8.3** Verify no import errors
  ```bash
  go build ./internal/config/...
  ```

- [ ] **8.4** Verify no unused imports
  - Each provider file should import: `cmp`, `os`, `catwalk`, maybe `slog`

---

### Phase 9: Full Integration Test (0.25 hours)

- [ ] **9.1** Run full test suite
  ```bash
  go test ./...
  ```
  Expected: All tests pass

- [ ] **9.2** Check provider count at runtime (should be same or higher)
  ```bash
  go run . --help  # builds successfully
  ```

- [ ] **9.3** Verify no regressions in other packages
  - `internal/agent/...` should still work
  - `internal/cmd/...` should still work

---

### Phase 10: Commit & Documentation (0.25 hours)

- [ ] **10.1** Review git diff
  ```bash
  git diff internal/config/provider.go      # Should show removed functions + streamlined code
  git diff internal/config/providers/       # Should show new files
  git status                                 # Should show all files staged/modified
  ```

- [ ] **10.2** Create semantic commit
  ```bash
  git add -A
  git commit -m "refactor: consolidate provider definitions into separate files
  
  - Move Mistral provider definition to providers/mistral.go
  - Move Nexora provider definition to providers/nexora.go
  - Move xAI provider definition to providers/xai.go
  - Move MiniMax provider definition to providers/minimax.go
  - Consolidate injectCustomProviders() to use loop instead of 4 if-statements
  - Remove unused injectCustomGitHubProviders() and fetchProvidersFromGitHub()
  - Make provider tests count-agnostic (check presence, not exact count)
  - Add per-provider unit tests for validation
  
  Benefits:
  - Eliminates 200+ lines of code duplication
  - Single file per provider makes updates easier
  - Tests no longer break when provider count changes
  - Paves way for Part 2 (provider CRUD system)"
  ```

- [ ] **10.3** Verify commit worked
  ```bash
  git log --oneline -5
  git show HEAD                    # Review the commit
  ```

---

## Test Files to Create

### `internal/config/providers/mistral_test.go`
```
TestMistralProvider_ReturnsValidProvider      | Provider has ID, Name, Type
TestMistralProvider_HasCorrectModelCount      | Mistral has 10 models
TestMistralProvider_ModelIDsUnique            | No duplicate model IDs
TestMistralProvider_AllModelsHaveMetadata     | Each model has cost, context, etc.
TestMistralProvider_SkipsIfAlreadyExists      | Returns empty provider if mistral in list
TestMistralProvider_SetAPIEndpointFromEnv    | Respects MISTRAL_API_ENDPOINT env var
TestMistralProvider_DefaultEndpoint           | Uses https://api.mistral.ai/v1 by default
TestMistralProvider_ModelPricing              | Large model is most expensive
```

### `internal/config/providers/xai_test.go`
```
TestXAIProvider_ReturnsValidProvider
TestXAIProvider_HasGrokBetaModel
TestXAIProvider_SkipsIfAlreadyExists
TestXAIProvider_SetAPIEndpointFromEnv
TestXAIProvider_DefaultEndpoint
```

### `internal/config/providers/nexora_test.go`
```
TestNexoraProvider_ReturnsValidProvider
TestNexoraProvider_UsesLocalhostEndpoint
TestNexoraProvider_SkipsIfAlreadyExists
TestNexoraProvider_EmptyAPIKey
```

### `internal/config/providers/minimax_test.go`
```
TestMiniMaxProvider_ReturnsValidProvider
TestMiniMaxProvider_HasTwoModels
TestMiniMaxProvider_UsesAnthropicType
TestMiniMaxProvider_SkipsIfAlreadyExists
TestMiniMaxProvider_ModelIDs
```

### Updated `internal/config/provider_test.go`
```
TestInjectCustomProviders_CallsAllInjectors    | All 4 injector funcs called
TestInjectCustomProviders_SkipsDuplicates      | Doesn't inject if already present
TestInjectCustomProviders_ReturnsAllProviders  | Has original + injected
TestInjectCustomProviders_LogsCount            | Logs how many injected
TestProvider_loadProvidersHasRequiredProviders | Has mistral, nexora, xai, minimax
TestProvider_loadProvidersNoIssues_FlexibleCount| Checks for specific providers
```

---

## Commands to Verify Completion

```bash
# Build succeeds
go build ./...

# All tests pass
go test ./... -v

# No format issues
gofumpt -l internal/config/

# No linting issues
golangci-lint run ./internal/config/...

# Correct line count (should be less than original 511)
wc -l internal/config/provider.go
wc -l internal/config/providers/*.go

# No orphaned code
grep -n "func inject" internal/config/provider.go
# Should return nothing (all moved to providers/)
```

---

## Success Criteria

- [x] All provider definitions moved to `internal/config/providers/`
- [x] Zero duplicate code (no two files with same provider definition)
- [x] `injectCustomProviders()` uses loop, not 4 if-statements
- [x] Dead GitHub code removed
- [x] Tests don't hardcode provider count expectations
- [x] All tests pass
- [x] Code formatted with gofumpt
- [x] No linting errors
- [x] Commit message clear and semantic
- [x] No regressions in other packages

---

## Time Breakdown

| Phase | Task | Estimated | Actual |
|-------|------|-----------|--------|
| 1 | Setup + test helpers | 30m | |
| 2 | Mistral provider | 45m | |
| 3 | xAI provider | 30m | |
| 4 | Nexora + MiniMax | 30m | |
| 5 | Consolidate injection | 30m | |
| 6 | Delete dead code | 10m | |
| 7 | Update tests | 30m | |
| 8 | Code quality | 30m | |
| 9 | Integration test | 15m | |
| 10 | Commit + doc | 15m | |
| **Total** | **Part 1** | **~3.5 hours** | |

---

## Notes

- Work incrementally: commit after each provider file is done
- Run tests after each phase to catch issues early
- If a test fails, fix immediately rather than continuing
- Use `git stash` if you need to revert a phase
- TDD means: write test → implement → test passes → refactor → tests still pass
