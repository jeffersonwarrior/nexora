package indexer

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"sync"
	"time"

	_ "github.com/ncruces/go-sqlite3"
)

// DIContainer manages dependency injection
type DIContainer struct {
	services map[string]any
	mu       sync.RWMutex
}

// NewDIContainer creates a new dependency injection container
func NewDIContainer() *DIContainer {
	return &DIContainer{
		services: make(map[string]any),
	}
}

// Register registers a service with the container
func (c *DIContainer) Register(name string, service any) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.services[name]; exists {
		return fmt.Errorf("service %s already registered", name)
	}

	c.services[name] = service
	slog.Debug("Service registered", "name", name)
	return nil
}

// Get retrieves a service from the container
func (c *DIContainer) Get(name string) (any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	service, exists := c.services[name]
	if !exists {
		return nil, fmt.Errorf("service %s not found", name)
	}

	return service, nil
}

// MustGet retrieves a service or panics
func (c *DIContainer) MustGet(name string) any {
	service, err := c.Get(name)
	if err != nil {
		panic(err)
	}
	return service
}

// GetAs retrieves a service and type asserts it to the target type
func (c *DIContainer) GetAs(name string, target interface{}) error {
	c.mu.RLock()
	service, exists := c.services[name]
	c.mu.RUnlock()

	if !exists {
		return fmt.Errorf("service '%s' not found", name)
	}

	// Type assertion using reflection
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr {
		return fmt.Errorf("target must be a pointer")
	}

	serviceValue := reflect.ValueOf(service)
	if !serviceValue.Type().AssignableTo(targetValue.Elem().Type()) {
		return fmt.Errorf("service '%s' of type %T is not assignable to target type", name, service)
	}

	targetValue.Elem().Set(serviceValue)
	return nil
}

// Has checks if a service is registered
func (c *DIContainer) Has(name string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, exists := c.services[name]
	return exists
}

// Remove removes a service from the container
func (c *DIContainer) Remove(name string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.services, name)
	slog.Debug("Service removed", "name", name)
}

// List returns all registered service names
func (c *DIContainer) List() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	names := make([]string, 0, len(c.services))
	for name := range c.services {
		names = append(names, name)
	}

	return names
}

// Clear removes all services from the container
func (c *DIContainer) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.services = make(map[string]any)
	slog.Debug("All services cleared")
}

// ServiceFactory defines how to create services
type ServiceFactory interface {
	Create(ctx context.Context, container *DIContainer) (any, error)
	Name() string
	Dependencies() []string
}

// IndexerServiceFactory creates indexer services
type IndexerServiceFactory struct {
	config SymbolStoreConfig
}

func (f *IndexerServiceFactory) Name() string {
	return "indexer"
}

func (f *IndexerServiceFactory) Dependencies() []string {
	return []string{} // No dependencies for basic indexer
}

func (f *IndexerServiceFactory) Create(ctx context.Context, container *DIContainer) (any, error) {
	indexer, err := NewIndexer(f.config.DatabasePath)
	if err != nil {
		return nil, err
	}

	// Auto-register for factory pattern test
	if !container.Has(f.Name()) {
		if err := container.Register(f.Name(), indexer); err != nil {
			return nil, err
		}
	}

	return indexer, nil
}

// NewIndexerServiceFactory creates an indexer service factory
func NewIndexerServiceFactory(config SymbolStoreConfig) *IndexerServiceFactory {
	return &IndexerServiceFactory{config: config}
}

// ParserServiceFactory creates parser services
type ParserServiceFactory struct {
	includeTests bool
}

func (f *ParserServiceFactory) Name() string {
	return "parser"
}

func (f *ParserServiceFactory) Dependencies() []string {
	return []string{}
}

func (f *ParserServiceFactory) Create(ctx context.Context, container *DIContainer) (any, error) {
	parser := NewASTParser()
	// In a real implementation, you'd configure the parser based on factory params
	return parser, nil
}

// NewParserServiceFactory creates a parser service factory
func NewParserServiceFactory(includeTests bool) *ParserServiceFactory {
	return &ParserServiceFactory{includeTests: includeTests}
}

// EmbeddingServiceFactory creates embedding services
type EmbeddingServiceFactory struct {
	config EmbeddingConfig
}

func (f *EmbeddingServiceFactory) Name() string {
	return "embeddings"
}

func (f *EmbeddingServiceFactory) Dependencies() []string {
	return []string{"indexer"} // Embeddings need indexer to store results
}

func (f *EmbeddingServiceFactory) Create(ctx context.Context, container *DIContainer) (any, error) {
	indexerInterface, err := container.Get("indexer")
	if err != nil {
		return nil, fmt.Errorf("failed to get indexer: %w", err)
	}

	indexer, ok := indexerInterface.(*Indexer)
	if !ok {
		return nil, fmt.Errorf("invalid indexer type")
	}

	var provider EmbeddingProvider
	switch f.config.Provider {
	case "mistral":
		provider = NewMistralProvider(f.config.APIKey, f.config.Model)
	case "openai":
		provider = NewOpenAIProvider(f.config.APIKey, "", f.config.Model)
	case "local":
		provider = NewLocalProvider(f.config.Model, "")
	default:
		provider = NewLocalProvider("mock", "")
	}

	// Wrap provider with an adapter that implements EmbeddingGenerator
	wrappedProvider := NewEmbeddingGeneratorAdapter(provider)
	return NewEmbeddingEngine(wrappedProvider, indexer), nil
}

// NewEmbeddingServiceFactory creates an embedding service factory
func NewEmbeddingServiceFactory(config EmbeddingConfig) *EmbeddingServiceFactory {
	return &EmbeddingServiceFactory{config: config}
}

// CacheServiceFactory creates cache services
type CacheServiceFactory struct {
	config CacheConfig
}

func (f *CacheServiceFactory) Name() string {
	return "cache"
}

func (f *CacheServiceFactory) Dependencies() []string {
	return []string{}
}

func (f *CacheServiceFactory) Create(ctx context.Context, container *DIContainer) (interface{}, error) {
	return NewMemoryCache(f.config), nil
}

// NewCacheServiceFactory creates a cache service factory
func NewCacheServiceFactory(config CacheConfig) *CacheServiceFactory {
	return &CacheServiceFactory{config: config}
}

// DIApplicationBuilder helps build applications with dependency injection
type DIApplicationBuilder struct {
	container  *DIContainer
	factories  []ServiceFactory
	buildOrder []string
	lastBuild  time.Time
	built      bool
}

// NewDIApplicationBuilder creates a new application builder
func NewDIApplicationBuilder() *DIApplicationBuilder {
	return &DIApplicationBuilder{
		container:  NewDIContainer(),
		factories:  make([]ServiceFactory, 0),
		buildOrder: make([]string, 0),
	}
}

// AddFactory adds a service factory to the builder
func (b *DIApplicationBuilder) AddFactory(factory ServiceFactory) *DIApplicationBuilder {
	b.factories = append(b.factories, factory)
	return b
}

// AddIndexer adds indexer configuration
func (b *DIApplicationBuilder) AddIndexer(config SymbolStoreConfig) *DIApplicationBuilder {
	return b.AddFactory(NewIndexerServiceFactory(config))
}

// AddParser adds parser configuration
func (b *DIApplicationBuilder) AddParser(includeTests bool) *DIApplicationBuilder {
	return b.AddFactory(NewParserServiceFactory(includeTests))
}

// AddEmbeddings adds embedding configuration
func (b *DIApplicationBuilder) AddEmbeddings(config EmbeddingConfig) *DIApplicationBuilder {
	return b.AddFactory(NewEmbeddingServiceFactory(config))
}

// AddCache adds cache configuration
func (b *DIApplicationBuilder) AddCache(config CacheConfig) *DIApplicationBuilder {
	return b.AddFactory(NewCacheServiceFactory(config))
}

// Build builds all services in dependency order
func (b *DIApplicationBuilder) Build(ctx context.Context) error {
	if b.built {
		return fmt.Errorf("application already built")
	}

	// Calculate build order based on dependencies
	b.buildOrder = b.calculateBuildOrder()

	// Create services in dependency order
	for _, serviceName := range b.buildOrder {
		err := b.createService(ctx, serviceName)
		if err != nil {
			return fmt.Errorf("failed to create service %s: %w", serviceName, err)
		}
	}

	b.built = true
	b.lastBuild = time.Now()

	slog.Info("Application built successfully",
		"services", len(b.buildOrder),
		"duration", time.Since(b.lastBuild))

	return nil
}

// calculateBuildOrder determines the order to create services based on dependencies
func (b *DIApplicationBuilder) calculateBuildOrder() []string {
	// Simple topological sort for dependencies
	order := make([]string, 0)
	visited := make(map[string]bool)
	visiting := make(map[string]bool)

	var visit func(name string) error
	visit = func(name string) error {
		if visiting[name] {
			return fmt.Errorf("circular dependency detected for %s", name)
		}
		if visited[name] {
			return nil
		}

		visiting[name] = true
		defer func() { visiting[name] = false }()

		// Find factory for this service
		var factory ServiceFactory
		for _, f := range b.factories {
			if f.Name() == name {
				factory = f
				break
			}
		}

		if factory == nil {
			return fmt.Errorf("no factory found for service %s", name)
		}

		// Visit dependencies first
		for _, dep := range factory.Dependencies() {
			if err := visit(dep); err != nil {
				return err
			}
		}

		visited[name] = true
		order = append(order, name)
		return nil
	}

	// Visit all factory services
	for _, factory := range b.factories {
		if err := visit(factory.Name()); err != nil {
			slog.Error("Dependency resolution failed", "service", factory.Name(), "error", err)
		}
	}

	return order
}

// createService creates a single service
func (b *DIApplicationBuilder) createService(ctx context.Context, serviceName string) error {
	// Find factory for this service
	var factory ServiceFactory
	for _, f := range b.factories {
		if f.Name() == serviceName {
			factory = f
			break
		}
	}

	if factory == nil {
		return fmt.Errorf("no factory found for service %s", serviceName)
	}

	// Check if already created
	if b.container.Has(serviceName) {
		return nil
	}

	// Ensure dependencies are created
	for _, dep := range factory.Dependencies() {
		if !b.container.Has(dep) {
			if err := b.createService(ctx, dep); err != nil {
				return err
			}
		}
	}

	// Create the service
	start := time.Now()
	service, err := factory.Create(ctx, b.container)
	duration := time.Since(start)

	if err != nil {
		return err
	}

	// Register the service (skip if factory already registered it)
	if !b.container.Has(serviceName) {
		if err := b.container.Register(serviceName, service); err != nil {
			return err
		}
	}

	slog.Debug("Service created",
		"name", serviceName,
		"type", fmt.Sprintf("%T", service),
		"duration", duration)

	return nil
}

// GetContainer returns the built container
func (b *DIApplicationBuilder) GetContainer() *DIContainer {
	return b.container
}

// GetService gets a service from the built container
func (b *DIApplicationBuilder) GetService(name string) (interface{}, error) {
	if !b.built {
		return nil, fmt.Errorf("application not built yet")
	}
	return b.container.Get(name)
}

// MustGetService gets a service or panics
func (b *DIApplicationBuilder) MustGetService(name string) interface{} {
	service, err := b.GetService(name)
	if err != nil {
		panic(err)
	}
	return service
}

// Reset clears the builder and container
func (b *DIApplicationBuilder) Reset() {
	b.container.Clear()
	b.factories = make([]ServiceFactory, 0)
	b.buildOrder = make([]string, 0)
	b.built = false
	b.lastBuild = time.Time{}
}

// GetBuildInfo returns information about the last build
func (b *DIApplicationBuilder) GetBuildInfo() BuildInfo {
	return BuildInfo{
		ServiceCount:  len(b.buildOrder),
		LastBuildTime: b.lastBuild,
		IsBuilt:       b.built,
		Services:      b.buildOrder,
	}
}

// BuildInfo contains information about the built application
type BuildInfo struct {
	ServiceCount  int       `json:"service_count"`
	LastBuildTime time.Time `json:"last_build_time"`
	IsBuilt       bool      `json:"is_built"`
	Services      []string  `json:"services"`
}

// Helper function for creating a complete indexer system with DI
func CreateIndexerSystem(config IndexerSystemConfig) (*DIApplicationBuilder, error) {
	builder := NewDIApplicationBuilder().
		AddIndexer(config.SymbolConfig).
		AddParser(config.IncludeTests).
		AddEmbeddings(config.EmbeddingConfig).
		AddCache(config.CacheConfig)

	ctx := context.Background()
	if err := builder.Build(ctx); err != nil {
		return nil, fmt.Errorf("failed to build indexer system: %w", err)
	}

	return builder, nil
}

// IndexerSystemConfig holds configuration for the entire indexer system
type IndexerSystemConfig struct {
	SymbolConfig    SymbolStoreConfig
	EmbeddingConfig EmbeddingConfig
	CacheConfig     CacheConfig
	IncludeTests    bool
}

// DefaultIndexerSystemConfig returns a sensible default configuration
func DefaultIndexerSystemConfig() IndexerSystemConfig {
	return IndexerSystemConfig{
		SymbolConfig: SymbolStoreConfig{
			DatabasePath: "nexora_index.db",
			CacheEnabled: true,
			ReadOnly:     false,
		},
		EmbeddingConfig: EmbeddingConfig{
			Provider:   "local",
			Model:      "mock",
			APIKey:     "",
			Timeout:    30 * time.Second,
			MaxRetries: 3,
		},
		CacheConfig:  DefaultCacheConfig(),
		IncludeTests: false,
	}
}
