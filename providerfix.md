# Provider System Refactoring Plan

## Executive Summary

The Nexora provider system needs two major improvements:

1. **Part 1: Fix provider.go** - Clean up hardcoded provider definitions and consolidate injection logic
2. **Part 2: Enable Provider CRUD** - Build a data-driven system for managing provider definitions without code changes

Current state: ~511 lines of repetitive Go code managing 4 custom providers (Mistral, Nexora, xAI, MiniMax) with inline model definitions and unused GitHub provider code.

Target state: Providers defined in external files (JSON/YAML), injectable at runtime via CLI/API, testable without modifying source code.

---

## Part 1: Fix provider.go (~2-3 hours)

### Current Problems

1. **Provider Functions Duplication** (Lines 206-460)
   - Each provider (Mistral, Nexora, xAI, MiniMax) has its own `inject*Provider()` function
   - Each function follows identical pattern: check for existing → return empty Provider, else build and return
   - Duplicated check logic in 4 places (lines 206-210, 345-349, 382-386, 418-422)

2. **Inline Model Definitions**
   - **Mistral**: 10 models × ~13 lines each = 130 lines total
   - **xAI**: 1 model × ~12 lines = 12 lines
   - **Nexora**: 1 model × ~10 lines = 10 lines
   - **MiniMax**: 2 models × ~12 lines each = 24 lines
   - **Total**: 176 lines of repetitive struct literals

3. **Unused Dead Code** (Lines 463-511)
   - `injectCustomGitHubProviders()` - Defined but NEVER called
   - `fetchProvidersFromGitHub()` - Supporting function for GitHub feature
   - These represent unfinished/abandoned feature that clutters codebase

4. **Configuration Inconsistency**
   - API endpoints hardcoded with env var fallbacks (line 217, 392, 428)
   - Provider types scattered: `"openai-compat"`, `"anthropic"`
   - DefaultHeaders always `map[string]string{}`
   - No single place to see all provider metadata

5. **Test Coupling**
   - `provider_test.go` line 35: `require.Len(t, providers, 5) // comment about expected count`
   - Test expectations are hardcoded and break when providers change
   - Adding a new provider requires updating multiple test files

### Solution: Create Provider Registry Pattern

**New file structure:**
```
internal/config/
├── provider.go (refactored - keep cache/load logic)
├── provider_registry.go (NEW - manages provider catalog)
├── providers/ (NEW directory)
│   ├── mistral.go
│   ├── nexora.go
│   ├── xai.go
│   └── minimax.go
└── (keep existing test files)
```

### Step 1.1: Create `provider_registry.go` (100 lines)

Define a structured provider catalog as Go code (will be data-driven in Part 2):

```go
package config

import "github.com/charmbracelet/catwalk/pkg/catwalk"

// ProviderDefinition describes a provider and its models.
type ProviderDefinition struct {
	ID                  string
	Name                string
	APIKey              string                  // e.g., "$MISTRAL_API_KEY" or ""
	APIEndpointTemplate string                  // e.g., "https://api.mistral.ai/v1"
	APIEndpointEnvVar   string                  // e.g., "MISTRAL_API_ENDPOINT" for override
	Type                string                  // "openai-compat", "anthropic", etc.
	DefaultLargeModelID string
	DefaultSmallModelID string
	Models              []ModelDefinition
}

// ModelDefinition describes a single model.
type ModelDefinition struct {
	ID               string
	Name             string
	CostPer1MIn      float64
	CostPer1MOut     float64
	ContextWindow    int
	DefaultMaxTokens int
	CanReason        bool
	SupportsImages   bool
}

// GetCustomProviders returns all custom provider definitions that should be injected.
func GetCustomProviders() []ProviderDefinition {
	return []ProviderDefinition{
		mistralProviderDef,
		nexoraProviderDef,
		xaiProviderDef,
		minimaxProviderDef,
	}
}
```

### Step 1.2: Create provider definition files (will consolidate duplication)

**File: `providers/mistral.go`** (150 lines - consolidates ~200 lines currently spread across provider.go)

```go
package providers

import (
	"cmp"
	"os"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/nexora/cli/internal/config"
)

// MistralProvider returns the Mistral provider definition.
func MistralProvider(existingProviders []catwalk.Provider) catwalk.Provider {
	// Check if mistral already exists
	for _, provider := range existingProviders {
		if provider.ID == "mistral" {
			return catwalk.Provider{}
		}
	}

	return catwalk.Provider{
		Name:                "Mistral",
		ID:                  "mistral",
		APIKey:              "$MISTRAL_API_KEY",
		APIEndpoint:         cmp.Or(os.Getenv("MISTRAL_API_ENDPOINT"), "https://api.mistral.ai/v1"),
		Type:                "openai-compat",
		DefaultLargeModelID: "mistral-large-3-25-12",
		DefaultSmallModelID: "ministral-3-8b-25-12",
		Models: []catwalk.Model{
			// Large reasoning models
			{
				ID:               "mistral-large-3-25-12",
				Name:             "Mistral Large 3",
				CostPer1MIn:      2.0,
				CostPer1MOut:     6.0,
				ContextWindow:    131072,
				DefaultMaxTokens: 32000,
				CanReason:        true,
				SupportsImages:   true,
				Options:          catwalk.ModelOptions{},
			},
			// ... [other models]
		},
		DefaultHeaders: map[string]string{},
	}
}
```

**Similar files for:**
- `providers/nexora.go` (50 lines)
- `providers/xai.go` (50 lines)
- `providers/minimax.go` (80 lines)

### Step 1.3: Refactor `provider.go` - Consolidate injection

Replace:
```go
func injectCustomProviders(providers []catwalk.Provider) []catwalk.Provider {
	// ... 4 separate if blocks
}

func injectMistralProvider(providers []catwalk.Provider) catwalk.Provider { ... }
func injectNexoraProvider(providers []catwalk.Provider) catwalk.Provider { ... }
func injectXAIProvider(providers []catwalk.Provider) catwalk.Provider { ... }
func injectMiniMaxProvider(providers []catwalk.Provider) catwalk.Provider { ... }
```

With:
```go
func injectCustomProviders(providers []catwalk.Provider) []catwalk.Provider {
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

	if injectedCount > 0 {
		slog.Info("Injected custom providers", "count", injectedCount, "total", len(providers))
	}
	return providers
}
```

### Step 1.4: Delete dead code

Remove lines 463-511:
- `injectCustomGitHubProviders()` function
- `fetchProvidersFromGitHub()` function

**Note**: These will return in Part 2 as a proper feature (GitHub-based provider registry).

### Step 1.5: Update tests

**In `provider_test.go`:**

Replace hardcoded provider count checks:
```go
// OLD (brittle)
require.Len(t, providers, 5) // mock + mistral + nexora + xai + minimax

// NEW (dynamic)
require.GreaterOrEqual(t, len(providers), 1)
mockFound := false
mistralFound := false
nexoraFound := false
for _, p := range providers {
	if p.ID == "mock" { mockFound = true }
	if p.ID == "mistral" { mistralFound = true }
	if p.ID == "nexora" { nexoraFound = true }
}
require.True(t, mockFound && mistralFound && nexoraFound, "Expected core providers")
```

This way tests pass regardless of how many providers exist.

### Part 1 Deliverables

- [ ] Delete lines 463-511 (dead GitHub code)
- [ ] Create `internal/config/providers/mistral.go` (150 lines)
- [ ] Create `internal/config/providers/nexora.go` (50 lines)
- [ ] Create `internal/config/providers/xai.go` (50 lines)
- [ ] Create `internal/config/providers/minimax.go` (80 lines)
- [ ] Refactor `injectCustomProviders()` in provider.go (use loop instead of 4 if statements)
- [ ] Delete individual `injectXxxProvider()` functions
- [ ] Update tests to be provider-count-agnostic
- [ ] Run `go test ./... -v` - all tests pass
- [ ] Code review: Check formatting with `gofumpt -w .`

**Lines saved**: ~250 lines removed, ~330 lines added = net change ~80 lines, but MASSIVE improvement in maintainability

---

## Part 2: Enable Provider CRUD (~4-6 hours)

### Goals

1. Define providers in **external JSON/YAML files** instead of Go code
2. Load providers from **multiple sources**: embedded, filesystem, remote URL, database
3. **CRUD operations**: Add, update, delete, list providers without code changes
4. **CLI commands**: `nexora provider add`, `nexora provider edit`, `nexora provider rm`
5. **Runtime injection**: Specify providers via `~/.config/nexora/providers.yaml` or env vars

### Current Architecture

```
User invokes nexora
    ↓
config.Providers(cfg) called
    ↓
loadProviders() function:
  1. Try cache from disk (providers.json)
  2. If miss, fetch from Catwalk URL (remote HTTP)
  3. If fail, use embedded providers
  4. In all cases: injectCustomProviders() adds Mistral, Nexora, xAI, MiniMax
    ↓
Returns []catwalk.Provider slice
    ↓
Used by agent/coordinator to select model
```

**Problem**: Injection happens automatically; no way to:
- Add new provider without modifying code
- Remove a provider
- Override provider settings
- Use different provider sets per project/session

### New Architecture

```
config/
├── provider_registry.go
├── providers/
│   ├── builtin/          (NEW) - Built-in Mistral, Nexora, etc.
│   │   ├── mistral.go
│   │   ├── nexora.go
│   │   └── ...
│   ├── file.go           (NEW) - Load from JSON/YAML
│   └── http.go           (NEW) - Fetch from URL
├── loader.go             (NEW) - Coordinates all sources
└── persister.go          (NEW) - Save/update providers

~/.config/nexora/
├── config.yaml           (existing)
├── providers.yaml        (NEW) - User-defined providers
└── providers.d/          (NEW) - Provider overrides directory
    ├── mistral-custom.yaml
    └── my-provider.yaml
```

### Step 2.1: Create Provider Interface (150 lines)

**File: `internal/config/provider_source.go`** (NEW)

Define how providers come from different places:

```go
package config

import (
	"context"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// ProviderSource defines how to fetch providers from a source.
type ProviderSource interface {
	// Name returns the human-readable name of this source (e.g., "embedded", "filesystem", "catwalk").
	Name() string

	// Providers returns the list of providers from this source.
	// Empty list + no error means source is empty but valid.
	// Error means source failed to load.
	Providers(ctx context.Context) ([]catwalk.Provider, error)

	// Priority determines load order (higher = loads first, can be overridden by lower).
	// Embedded: 10, Local FS: 50, Cache: 60, Remote: 70, Builtin Injections: 100
	Priority() int
}

// ProviderMerger handles conflicts when same provider ID exists in multiple sources.
type ProviderMerger interface {
	// Merge combines providers from multiple sources, respecting priority.
	// Higher priority sources override lower priority on conflicts.
	Merge(sources []ProviderSource) ([]catwalk.Provider, error)
}

// ProviderLoader orchestrates loading providers from all sources.
type ProviderLoader struct {
	sources []ProviderSource
	merger  ProviderMerger
	cache   []catwalk.Provider
	cacheOK bool
}

// NewProviderLoader creates a loader with default sources.
func NewProviderLoader(sources ...ProviderSource) *ProviderLoader {
	return &ProviderLoader{
		sources: sources,
		merger:  &defaultMerger{},
	}
}

// Load gets all providers, respecting priority and caching.
func (pl *ProviderLoader) Load(ctx context.Context) ([]catwalk.Provider, error) {
	if pl.cacheOK {
		return pl.cache, nil
	}

	providers, err := pl.merger.Merge(pl.sources)
	if err == nil {
		pl.cache = providers
		pl.cacheOK = true
	}
	return providers, err
}

// Reset clears the cache (useful after provider changes).
func (pl *ProviderLoader) Reset() {
	pl.cache = nil
	pl.cacheOK = false
}
```

### Step 2.2: Implement Provider Sources (300 lines)

**File: `internal/config/sources/embedded.go`** (NEW, 40 lines)

```go
package sources

import (
	"context"

	"github.com/charmbracelet/catwalk/pkg/embedded"
	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// EmbeddedSource returns providers bundled with Nexora release.
type EmbeddedSource struct{}

func (s *EmbeddedSource) Name() string    { return "embedded" }
func (s *EmbeddedSource) Priority() int   { return 10 }

func (s *EmbeddedSource) Providers(ctx context.Context) ([]catwalk.Provider, error) {
	return embedded.GetAll(), nil
}
```

**File: `internal/config/sources/filesystem.go`** (NEW, 80 lines)

```go
package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// FilesystemSource loads providers from local files.
// Supports both `~/.config/nexora/providers.yaml` and `providers.d/*.yaml`.
type FilesystemSource struct {
	basePath string // e.g., ~/.config/nexora
}

func NewFilesystemSource(basePath string) *FilesystemSource {
	return &FilesystemSource{basePath: basePath}
}

func (s *FilesystemSource) Name() string    { return "filesystem" }
func (s *FilesystemSource) Priority() int   { return 50 }

func (s *FilesystemSource) Providers(ctx context.Context) ([]catwalk.Provider, error) {
	var providers []catwalk.Provider

	// Load main providers.yaml
	mainFile := filepath.Join(s.basePath, "providers.yaml")
	if data, err := os.ReadFile(mainFile); err == nil {
		var p []catwalk.Provider
		if err := json.Unmarshal(data, &p); err == nil {
			providers = append(providers, p...)
		}
	}

	// Load *.yaml from providers.d/ directory
	dirPath := filepath.Join(s.basePath, "providers.d")
	if entries, err := os.ReadDir(dirPath); err == nil {
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".yaml") {
				continue
			}
			filePath := filepath.Join(dirPath, entry.Name())
			if data, err := os.ReadFile(filePath); err == nil {
				var p []catwalk.Provider
				if err := json.Unmarshal(data, &p); err == nil {
					providers = append(providers, p...)
				}
			}
		}
	}

	return providers, nil
}
```

**File: `internal/config/sources/cache.go`** (NEW, 50 lines)

```go
package sources

// CacheSource loads providers from ~/.local/share/nexora/providers.json
// This is the auto-updated provider list from Catwalk.
type CacheSource struct {
	filePath string
}

func NewCacheSource(filePath string) *CacheSource {
	return &CacheSource{filePath: filePath}
}

func (s *CacheSource) Name() string    { return "cache" }
func (s *CacheSource) Priority() int   { return 60 }

func (s *CacheSource) Providers(ctx context.Context) ([]catwalk.Provider, error) {
	// Same as current loadProvidersFromCache()
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return nil, fmt.Errorf("cache file not found: %w", err)
	}

	var providers []catwalk.Provider
	if err := json.Unmarshal(data, &providers); err != nil {
		return nil, fmt.Errorf("invalid cache file: %w", err)
	}

	return providers, nil
}
```

**File: `internal/config/sources/remote.go`** (NEW, 70 lines)

```go
package sources

// RemoteSource fetches providers from Catwalk or custom HTTP endpoint.
type RemoteSource struct {
	url string
}

func NewRemoteSource(url string) *RemoteSource {
	return &RemoteSource{url: url}
}

func (s *RemoteSource) Name() string    { return "remote" }
func (s *RemoteSource) Priority() int   { return 70 }

func (s *RemoteSource) Providers(ctx context.Context) ([]catwalk.Provider, error) {
	// Same as current catwalk.NewWithURL(url).GetProviders()
	client := catwalk.NewWithURL(s.url)
	return client.GetProviders(ctx, "")
}
```

**File: `internal/config/sources/builtin.go`** (NEW, 80 lines)

```go
package sources

// BuiltinSource provides hardcoded Mistral, Nexora, xAI, MiniMax providers.
// These are always injected LAST so they can be overridden by other sources.
type BuiltinSource struct{}

func (s *BuiltinSource) Name() string    { return "builtin" }
func (s *BuiltinSource) Priority() int   { return 100 }

func (s *BuiltinSource) Providers(ctx context.Context) ([]catwalk.Provider, error) {
	// Call the provider constructors created in Part 1
	// providers.MistralProvider(), providers.NexoraProvider(), etc.
	var providers []catwalk.Provider

	if p := providers.MistralProvider(providers); p.ID != "" {
		providers = append(providers, p)
	}
	if p := providers.NexoraProvider(providers); p.ID != "" {
		providers = append(providers, p)
	}
	// ... etc

	return providers, nil
}
```

### Step 2.3: Create Provider Manager (200 lines)

**File: `internal/config/manager.go`** (NEW)

```go
package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/catwalk/pkg/catwalk"
)

// ProviderManager handles CRUD operations on providers.
type ProviderManager struct {
	basePath string // ~/.config/nexora
	loader   *ProviderLoader
}

// NewProviderManager creates a new manager.
func NewProviderManager(basePath string) *ProviderManager {
	return &ProviderManager{
		basePath: basePath,
		loader:   NewProviderLoader(), // uses default sources
	}
}

// List returns all available providers.
func (pm *ProviderManager) List(ctx context.Context) ([]catwalk.Provider, error) {
	return pm.loader.Load(ctx)
}

// Get returns a single provider by ID.
func (pm *ProviderManager) Get(ctx context.Context, providerID string) (*catwalk.Provider, error) {
	providers, err := pm.List(ctx)
	if err != nil {
		return nil, err
	}

	for _, p := range providers {
		if p.ID == providerID {
			return &p, nil
		}
	}

	return nil, fmt.Errorf("provider not found: %s", providerID)
}

// Add saves a new provider to ~/.config/nexora/providers.d/{id}.yaml
func (pm *ProviderManager) Add(ctx context.Context, provider catwalk.Provider) error {
	if provider.ID == "" {
		return fmt.Errorf("provider ID is required")
	}

	dir := filepath.Join(pm.basePath, "providers.d")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create providers.d directory: %w", err)
	}

	filePath := filepath.Join(dir, fmt.Sprintf("%s.yaml", provider.ID))

	data, err := json.MarshalIndent(provider, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal provider: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write provider file: %w", err)
	}

	pm.loader.Reset() // Clear cache
	return nil
}

// Update modifies an existing provider.
func (pm *ProviderManager) Update(ctx context.Context, provider catwalk.Provider) error {
	// Check if provider exists
	if _, err := pm.Get(ctx, provider.ID); err != nil {
		return fmt.Errorf("cannot update non-existent provider: %w", err)
	}

	return pm.Add(ctx, provider) // Same logic as Add, overwrites
}

// Remove deletes a provider.
func (pm *ProviderManager) Remove(ctx context.Context, providerID string) error {
	// Don't allow removing builtin providers
	builtins := []string{"mistral", "nexora", "xai", "minimax"}
	for _, b := range builtins {
		if providerID == b {
			return fmt.Errorf("cannot remove builtin provider: %s", providerID)
		}
	}

	filePath := filepath.Join(pm.basePath, "providers.d", fmt.Sprintf("%s.yaml", providerID))
	if err := os.Remove(filePath); err != nil {
		return fmt.Errorf("failed to remove provider: %w", err)
	}

	pm.loader.Reset()
	return nil
}
```

### Step 2.4: Add CLI Commands (150 lines)

**File: `internal/cmd/provider.go`** (NEW)

```go
package cmd

import (
	"fmt"

	"github.com/nexora/cli/internal/config"
	"github.com/spf13/cobra"
)

var providerCmd = &cobra.Command{
	Use:   "provider",
	Short: "Manage AI providers",
	Long:  "Add, update, remove, or list AI providers",
}

var providerListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available providers",
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := config.NewProviderManager(config.ConfigDir())
		providers, err := mgr.List(cmd.Context())
		if err != nil {
			return err
		}

		for _, p := range providers {
			fmt.Printf("%s (%s) - %d models\n", p.Name, p.ID, len(p.Models))
		}
		return nil
	},
}

var providerAddCmd = &cobra.Command{
	Use:   "add <provider-file.yaml>",
	Short: "Add a new provider from a YAML file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Read provider from file, validate, call mgr.Add()
		// Implementation details...
		return nil
	},
}

var providerRmCmd = &cobra.Command{
	Use:   "rm <provider-id>",
	Short: "Remove a provider",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		mgr := config.NewProviderManager(config.ConfigDir())
		if err := mgr.Remove(cmd.Context(), args[0]); err != nil {
			return err
		}

		fmt.Printf("Provider '%s' removed\n", args[0])
		return nil
	},
}

func init() {
	providerCmd.AddCommand(providerListCmd, providerAddCmd, providerRmCmd)
	rootCmd.AddCommand(providerCmd)
}
```

### Step 2.5: Refactor UpdateProviders command

Update `internal/cmd/update_providers.go` to use new manager:

```go
// OLD: Hard-coded to fetch and save to cache
err := config.UpdateProviders(pathOrUrl)

// NEW: Use manager with explicit source
ctx := cmd.Context()
mgr := config.NewProviderManager(config.ConfigDir())

// Fetch from source
source := sources.NewRemoteSource(pathOrUrl)
providers, err := source.Providers(ctx)
if err != nil {
	return err
}

// Save to cache (or user providers.d)
for _, p := range providers {
	if err := mgr.Add(ctx, p); err != nil {
		return err
	}
}
```

### Step 2.6: Update Config Loading

In `config.go`, update the `Providers()` function to use new loader:

```go
func Providers(cfg *Config) ([]catwalk.Provider, error) {
	providerOnce.Do(func() {
		mgr := NewProviderManager(ConfigDir())
		ctx := context.Background() // or use proper context

		providerList, providerErr = mgr.List(ctx)
	})
	return providerList, providerErr
}
```

### Part 2 Deliverables

- [ ] Create `internal/config/sources/` directory
- [ ] Create `sources/embedded.go` (40 lines)
- [ ] Create `sources/filesystem.go` (80 lines)
- [ ] Create `sources/cache.go` (50 lines)
- [ ] Create `sources/remote.go` (70 lines)
- [ ] Create `sources/builtin.go` (80 lines)
- [ ] Create `internal/config/manager.go` (200 lines)
- [ ] Create `internal/config/provider_source.go` (150 lines)
- [ ] Create `internal/cmd/provider.go` with list/add/rm commands (150 lines)
- [ ] Update `config.go` to use manager
- [ ] Update `cmd/update_providers.go` to use manager
- [ ] Create example provider files in docs: `examples/provider-anthropic.yaml`, `examples/provider-gemini.yaml`
- [ ] Update tests for manager and sources
- [ ] Run full test suite: `go test ./...`
- [ ] Test CLI: `nexora provider list`, `nexora provider add`, etc.

---

## Implementation Order & Timeline

### Week 1: Part 1 (Fix Duplication)
- **Day 1-2**: Create provider subdirectory, move Mistral/Nexora/xAI/MiniMax definitions
- **Day 2-3**: Refactor injection logic, delete dead code, update tests
- **Day 4**: Code review, formatting, test execution
- **Day 5**: Commit with message: `refactor: consolidate provider definitions into separate files`

### Week 2: Part 2 (Enable CRUD)
- **Day 1**: Design ProviderLoader and ProviderSource interface
- **Day 2-3**: Implement sources (embedded, filesystem, cache, remote, builtin)
- **Day 4**: Create ProviderManager and CLI commands
- **Day 5**: Integration testing, documentation, commit with message: `feat: enable CRUD operations for providers`

---

## Testing Strategy

### Unit Tests

```bash
# Part 1
go test ./internal/config/providers/... -v
go test ./internal/config -run TestProvider -v

# Part 2
go test ./internal/config/sources/... -v
go test ./internal/config -run TestManager -v
go test ./internal/cmd -run TestProvider -v
```

### Integration Tests

1. **Load providers from all sources** - Verify priority/override behavior
2. **Add new provider** - Via CLI and programmatically
3. **Update existing provider** - Models, costs, endpoints
4. **Remove custom provider** - But not builtins
5. **Provider override** - Builtin provider overridden by filesystem source

### Examples

```bash
# List all providers
$ nexora provider list
OpenAI (openai) - 8 models
Mistral (mistral) - 10 models
...

# Add custom provider from file
$ nexora provider add ./my-provider.yaml
Provider 'my-llm' added

# List again
$ nexora provider list | grep my-llm
My LLM (my-llm) - 3 models

# Remove provider
$ nexora provider rm my-llm
Provider 'my-llm' removed

# Update from Catwalk
$ nexora update-providers https://api.catwalk.example.com/v2/providers
Providers updated successfully
```

---

## Files Modified Summary

### Part 1
| File | Action | Lines | Notes |
|------|--------|-------|-------|
| `internal/config/provider.go` | Refactor | -200, +80 | Remove duplication, consolidate |
| `internal/config/providers/mistral.go` | Create | 150 | Move from provider.go |
| `internal/config/providers/nexora.go` | Create | 50 | Move from provider.go |
| `internal/config/providers/xai.go` | Create | 50 | Move from provider.go |
| `internal/config/providers/minimax.go` | Create | 80 | Move from provider.go |
| `internal/config/provider_test.go` | Update | -20, +30 | Make tests provider-agnostic |
| `internal/config/provider_empty_test.go` | Update | -10, +15 | Make tests provider-agnostic |

### Part 2
| File | Action | Lines | Notes |
|------|--------|-------|-------|
| `internal/config/provider_source.go` | Create | 150 | Interface & loader |
| `internal/config/sources/embedded.go` | Create | 40 | Load from embedded |
| `internal/config/sources/filesystem.go` | Create | 80 | Load from ~/.config/nexora |
| `internal/config/sources/cache.go` | Create | 50 | Load from cache |
| `internal/config/sources/remote.go` | Create | 70 | Fetch from HTTP |
| `internal/config/sources/builtin.go` | Create | 80 | Inject builtin providers |
| `internal/config/manager.go` | Create | 200 | CRUD operations |
| `internal/cmd/provider.go` | Create | 150 | CLI commands |
| `internal/config/config.go` | Update | -10, +20 | Use new manager |
| `internal/cmd/update_providers.go` | Update | -20, +30 | Use new sources |

**Total: ~100 lines removed, ~950 lines added (net +850), but with MASSIVE improvements in:**
- Code organization (separation of concerns)
- Maintainability (single provider = single file)
- Extensibility (new providers don't require code changes)
- Testability (each source independently testable)
- User experience (CLI commands for provider management)

---

## Success Criteria

### Part 1 Complete When:
1. ✅ All provider definitions moved to separate files
2. ✅ `injectCustomProviders()` uses loop instead of 4 if-statements
3. ✅ Dead GitHub code removed
4. ✅ All tests pass with provider-agnostic expectations
5. ✅ Code formatted and linted
6. ✅ No duplicate code across provider files

### Part 2 Complete When:
1. ✅ Providers can be loaded from JSON/YAML files
2. ✅ CLI commands work: `nexora provider list`, `add`, `rm`
3. ✅ Provider priority/override behavior works correctly
4. ✅ Manager CRUD operations complete
5. ✅ New providers can be added without code changes
6. ✅ Example provider files exist in docs
7. ✅ All tests pass (unit + integration)

---

## Notes & Considerations

### Backward Compatibility
- Existing `providers.json` cache format unchanged
- `nexora update-providers` continues to work
- Existing configs/models/sessions unaffected
- New API is additive; old code keeps working

### Performance
- Provider loading still uses sync.Once (single initialization)
- Filesystem sources checked once at startup
- Cache prevents repeated disk I/O
- No performance regression vs. current

### Future Enhancements
After Part 2 complete:
1. GitHub provider registry (re-implement `fetchProvidersFromGitHub`)
2. Database-backed provider persistence
3. Provider aliasing (shorthand names)
4. Provider health checks (availability monitoring)
5. Cost tracking per provider/model
6. Multi-workspace provider sets

---

## Estimated Effort

| Phase | Task | Effort | Status |
|-------|------|--------|--------|
| 1.1 | Move Mistral provider | 1h | Planned |
| 1.2 | Move Nexora/xAI/MiniMax | 1.5h | Planned |
| 1.3 | Refactor injection logic | 0.5h | Planned |
| 1.4 | Delete dead code | 0.1h | Planned |
| 1.5 | Update tests | 0.5h | Planned |
| **Part 1 Total** | **Code organization** | **~3.5 hours** | **Ready** |
| 2.1 | Create interfaces & loader | 1.5h | Planned |
| 2.2 | Implement sources | 2h | Planned |
| 2.3 | Create manager | 1h | Planned |
| 2.4 | Add CLI commands | 1h | Planned |
| 2.5 | Refactor UpdateProviders | 0.5h | Planned |
| 2.6 | Update config loading | 0.5h | Planned |
| 2.7 | Write tests & examples | 1.5h | Planned |
| **Part 2 Total** | **Provider CRUD** | **~8 hours** | **Planned** |
| **GRAND TOTAL** | **Full refactor** | **~11-12 hours** | **Doable in 2-3 work days** |

---

## Questions & Decisions Needed

1. **Provider storage format**: JSON or YAML? (YAML is more readable)
2. **Config directory**: Continue with `~/.config/nexora`? Or allow override?
3. **Provider IDs**: Keep current (mistral, nexora, xai, minimax) or change?
4. **Builtin overrides**: Should users be able to override builtin providers?
5. **CLI flags**: Do we need `--priority` when adding providers?
6. **Documentation**: Create migration guide for existing users?
