package indexer

import (
	"context"
	"time"
)

// QueryStats represents query execution statistics
type QueryStats struct {
	TotalQueries      int64         `json:"total_queries"`
	SuccessfulQueries int64         `json:"successful_queries"`
	FailedQueries     int64         `json:"failed_queries"`
	AverageLatency    time.Duration `json:"average_latency"`
	LastQueryTime     time.Time     `json:"last_query_time"`
}

// CacheStats represents cache statistics
type CacheStats struct {
	TotalItems   int64   `json:"total_items"`
	HitRate      float64 `json:"hit_rate"`
	MissRate     float64 `json:"miss_rate"`
	EvictionRate float64 `json:"eviction_rate"`
	MemoryUsage  int64   `json:"memory_usage_bytes"`
}

// SymbolStore defines operations for storing and retrieving code symbols
type SymbolStore interface {
	// Storage operations
	StoreSymbols(ctx context.Context, symbols []Symbol) error
	DeleteSymbolsByFile(ctx context.Context, filePath string) error

	// Retrieval operations
	SearchSymbols(ctx context.Context, query string, limit int) ([]Symbol, error)
	GetSymbol(ctx context.Context, id string) (*Symbol, error)
	GetCalledSymbols(ctx context.Context, symbolName string) ([]Symbol, error)
	GetAllSymbols(ctx context.Context) ([]Symbol, error)

	// Metadata operations
	Close() error
}

// EmbeddingStore defines operations for storing and retrieving embeddings
type EmbeddingStore interface {
	// Storage operations
	StoreEmbeddings(ctx context.Context, embeddings []Embedding) error
	DeleteEmbeddingsByFile(ctx context.Context, filePath string) error

	// Retrieval operations
	GetAllEmbeddings(ctx context.Context) ([]Embedding, error)
	SearchSimilar(ctx context.Context, query string, limit int) ([]Embedding, error)

	// Metadata operations
	Close() error
}

// CodeParser defines operations for parsing Go source code
type CodeParser interface {
	// Directory parsing
	ParseDirectory(ctx context.Context, dir string) ([]Symbol, error)

	// Single file parsing
	ParseFile(ctx context.Context, filename string) ([]Symbol, error)

	// Configuration
	SetIncludeTests(include bool)
	SetIgnoredDirs(dirs []string)
}

// EmbeddingGenerator defines operations for generating embeddings
type EmbeddingGenerator interface {
	// Single embedding generation
	GenerateEmbedding(ctx context.Context, text string) ([]float32, error)

	// Batch embedding generation
	GenerateBatchEmbeddings(ctx context.Context, texts []string) ([][]float32, error)

	// Symbol-specific operations
	GenerateSymbolEmbeddings(ctx context.Context, symbols []Symbol) ([]Embedding, error)

	// Metadata
	Name() string
	ValidateAPIKey(ctx context.Context) bool
}

// GraphBuilder defines operations for building and managing code graphs
type CodeGraphBuilder interface {
	// Graph construction
	BuildGraph(ctx context.Context, symbols []Symbol) (*Graph, error)

	// Graph analysis
	GetDependencies(ctx context.Context, symbolID string) ([]string, error)
	GetDependents(ctx context.Context, symbolID string) ([]string, error)
	GetCallGraph(ctx context.Context, symbolID string) ([]string, error)

	// Graph traversal
	GetTransitiveDependencies(ctx context.Context, symbolID string, maxDepth int) ([]string, error)
	GetTransitiveDependents(ctx context.Context, symbolID string, maxDepth int) ([]string, error)

	// Metadata
	GetAllNodes(ctx context.Context) ([]string, error)
	GetRelationships(ctx context.Context, symbolID string) ([]string, error)
}

// QueryExecutor defines operations for querying indexed code
type QueryExecutor interface {
	// Search operations
	Search(ctx context.Context, req *QueryRequest) ([]QueryResult, error)
	AdvancedQuery(ctx context.Context, query string) ([]QueryResult, error)

	// Search types
	SearchSymbols(ctx context.Context, query string, limit int) ([]Symbol, error)
	SearchSimilar(ctx context.Context, query string, limit int) ([]Embedding, error)
	SearchGraph(ctx context.Context, query string, limit int) ([]QueryResult, error)

	// Configuration
	SetSymbolStore(store SymbolStore)
	SetEmbeddingStore(store EmbeddingStore)
	SetGraphBuilder(builder GraphBuilder)

	// Metadata
	GetQueryStats() QueryStats
}

// FileManager defines operations for file system monitoring
type FileManager interface {
	// Path management
	AddPath(path string) error
	RemovePath(path string) error
	GetWatchedPaths() []string
	IsWatching() bool

	// Configuration
	SetIgnoredDirs(dirs []string)
	SetIgnoredExts(exts []string)
	SetDebounceDelay(delay time.Duration)

	// Event handling
	OnFileAdded(callback func(string))
	OnFileChanged(callback func(string))
	OnFileRemoved(callback func(string))

	// Lifecycle
	Start() error
	Stop() error
}

// CacheManager defines operations for caching system
type CacheManager interface {
	// Cache operations
	Get(ctx context.Context, key string) (interface{}, bool)
	Set(ctx context.Context, key string, value interface{})
	Delete(ctx context.Context, key string)
	Clear(ctx context.Context)

	// Batch operations
	GetBatch(ctx context.Context, keys []string) (map[string]interface{}, error)
	SetBatch(ctx context.Context, items map[string]interface{}) error

	// Metadata
	GetMetrics() CacheMetrics
	GetStats() CacheStats

	// Configuration
	Close() error
}

// DeltaManager defines operations for incremental updates
type DeltaManager interface {
	// Delta operations
	ProcessDelta(ctx context.Context, batch DeltaBatch) error
	CreateBatch(added, modified, removed []string) DeltaBatch

	// State management
	GetLastSync() time.Time
	SetLastSync(t time.Time)
	Checkpoint(ctx context.Context) error
	GetLastCheckpoint(ctx context.Context) (time.Time, error)

	// Configuration
	ConfigureIndexer(indexer SymbolStore)
	ConfigureParser(parser CodeParser)
	ConfigureGenerator(generator EmbeddingGenerator)
}

// Composed interfaces for common use cases
type IndexerService interface {
	SymbolStore
	EmbeddingStore
}

type SearchService interface {
	QueryExecutor
	EmbeddingStore
}

type MonitoringService interface {
	FileManager
	DeltaManager
}

// Factory interfaces for dependency injection
type SymbolStoreFactory interface {
	CreateSymbolStore(config SymbolStoreConfig) (SymbolStore, error)
}

type EmbeddingGeneratorFactory interface {
	CreateEmbeddingGenerator(config EmbeddingConfig) (EmbeddingGenerator, error)
}

// Configuration types
type SymbolStoreConfig struct {
	DatabasePath string
	CacheEnabled bool
	ReadOnly     bool
}

type EmbeddingConfig struct {
	Provider   string
	Model      string
	APIKey     string
	Timeout    time.Duration
	MaxRetries int
}

// Configuration types
