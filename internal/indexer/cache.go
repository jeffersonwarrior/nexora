package indexer

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"
)

// CacheConfig represents cache configuration
type CacheConfig struct {
	MaxSize       int           `json:"max_size"`       // Max number of items to cache
	TTL           time.Duration `json:"ttl"`            // Time to live for cache entries
	CleanupPeriod time.Duration `json:"cleanup_period"` // How often to clean expired entries
	EnableMetrics bool          `json:"enable_metrics"` // Whether to track cache metrics
}

// DefaultCacheConfig returns a sensible default configuration
func DefaultCacheConfig() CacheConfig {
	return CacheConfig{
		MaxSize:       10000,
		TTL:           10 * time.Minute,
		CleanupPeriod: 5 * time.Minute,
		EnableMetrics: true,
	}
}

// CacheEntry represents a cached item
type CacheEntry struct {
	Key       string    `json:"key"`
	Value     any       `json:"value"`
	Created   time.Time `json:"created"`
	Accessed  time.Time `json:"accessed"`
	ExpiresAt time.Time `json:"expires_at"`
	HitCount  int64     `json:"hit_count"`
}

// IsExpired checks if the cache entry has expired
func (ce *CacheEntry) IsExpired() bool {
	return time.Now().After(ce.ExpiresAt)
}

// CacheMetrics tracks cache performance
type CacheMetrics struct {
	Hits      int64   `json:"hits"`
	Misses    int64   `json:"misses"`
	Evictions int64   `json:"evictions"`
	Size      int     `json:"size"`
	MaxSize   int     `json:"max_size"`
	HitRate   float64 `json:"hit_rate"`
	mu        sync.RWMutex
}

// recordHit records a cache hit
func (cm *CacheMetrics) recordHit() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.Hits++
	cm.calculateHitRate()
}

// recordMiss records a cache miss
func (cm *CacheMetrics) recordMiss() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.Misses++
	cm.calculateHitRate()
}

// recordEviction records a cache eviction
func (cm *CacheMetrics) recordEviction() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.Evictions++
}

// setSize updates the current cache size
func (cm *CacheMetrics) setSize(size int) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.Size = size
}

// calculateHitRate calculates the hit rate
func (cm *CacheMetrics) calculateHitRate() {
	total := cm.Hits + cm.Misses
	if total > 0 {
		cm.HitRate = float64(cm.Hits) / float64(total)
	}
}

// GetSnapshot returns a snapshot of current metrics
func (cm *CacheMetrics) GetSnapshot() CacheMetrics {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	// Return a copy without the mutex to avoid copying the lock
	return CacheMetrics{
		Hits:      cm.Hits,
		Misses:    cm.Misses,
		Evictions: cm.Evictions,
		Size:      cm.Size,
		MaxSize:   cm.MaxSize,
		HitRate:   cm.HitRate,
	}
}

// MemoryCache provides an in-memory caching layer
type MemoryCache struct {
	config  CacheConfig
	cache   map[string]*CacheEntry
	queue   []string // LRU queue for eviction
	metrics CacheMetrics
	mu      sync.RWMutex
	stop    chan struct{}
}

// NewMemoryCache creates a new memory cache
func NewMemoryCache(config CacheConfig) *MemoryCache {
	mc := &MemoryCache{
		config: config,
		cache:  make(map[string]*CacheEntry),
		queue:  make([]string, 0),
		metrics: CacheMetrics{
			MaxSize: config.MaxSize,
		},
		stop: make(chan struct{}),
	}

	// Start cleanup routine if enabled
	if config.CleanupPeriod > 0 {
		go mc.cleanupRoutine()
	}

	return mc
}

// Get retrieves a value from the cache
func (mc *MemoryCache) Get(ctx context.Context, key string) (any, bool) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	entry, exists := mc.cache[key]
	if !exists {
		if mc.config.EnableMetrics {
			mc.metrics.recordMiss()
		}
		return nil, false
	}

	if entry.IsExpired() {
		delete(mc.cache, key)
		mc.removeFromQueue(key)
		if mc.config.EnableMetrics {
			mc.metrics.recordMiss()
			mc.metrics.setSize(len(mc.cache))
		}
		return nil, false
	}

	// Update access time and hit count
	entry.Accessed = time.Now()
	entry.HitCount++

	// Move to end of queue (LRU)
	mc.moveToEnd(key)

	if mc.config.EnableMetrics {
		mc.metrics.recordHit()
	}

	return entry.Value, true
}

// Set stores a value in the cache
func (mc *MemoryCache) Set(ctx context.Context, key string, value any) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	now := time.Now()
	expiresAt := now.Add(mc.config.TTL)

	entry := &CacheEntry{
		Key:       key,
		Value:     value,
		Created:   now,
		Accessed:  now,
		ExpiresAt: expiresAt,
		HitCount:  0,
	}

	// Check if updating existing entry
	if _, exists := mc.cache[key]; !exists {
		// Check if we need to evict
		if len(mc.cache) >= mc.config.MaxSize {
			mc.evictLRU()
		}
		mc.queue = append(mc.queue, key)
	} else {
		// Update existing
		mc.moveToEnd(key)
	}

	mc.cache[key] = entry

	if mc.config.EnableMetrics {
		mc.metrics.setSize(len(mc.cache))
	}
}

// Delete removes a value from the cache
func (mc *MemoryCache) Delete(ctx context.Context, key string) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if _, exists := mc.cache[key]; exists {
		delete(mc.cache, key)
		mc.removeFromQueue(key)

		if mc.config.EnableMetrics {
			mc.metrics.setSize(len(mc.cache))
		}
	}
}

// Clear removes all entries from the cache
func (mc *MemoryCache) Clear(ctx context.Context) {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	mc.cache = make(map[string]*CacheEntry)
	mc.queue = make([]string, 0)

	if mc.config.EnableMetrics {
		mc.metrics.setSize(0)
	}
}

// Close stops the cache cleanup routine
func (mc *MemoryCache) Close() {
	close(mc.stop)
}

// GetMetrics returns the current cache metrics
func (mc *MemoryCache) GetMetrics() CacheMetrics {
	if mc.config.EnableMetrics {
		return mc.metrics.GetSnapshot()
	}
	return CacheMetrics{}
}

// evictLRU evicts the least recently used entry
func (mc *MemoryCache) evictLRU() {
	if len(mc.queue) == 0 {
		return
	}

	// Remove first item (LRU)
	lruKey := mc.queue[0]
	delete(mc.cache, lruKey)
	mc.queue = mc.queue[1:]

	if mc.config.EnableMetrics {
		mc.metrics.recordEviction()
		mc.metrics.setSize(len(mc.cache))
	}
}

// moveToEnd moves a key to the end of the LRU queue
func (mc *MemoryCache) moveToEnd(key string) {
	mc.removeFromQueue(key)
	mc.queue = append(mc.queue, key)
}

// removeFromQueue removes a key from the queue
func (mc *MemoryCache) removeFromQueue(key string) {
	for i, qKey := range mc.queue {
		if qKey == key {
			mc.queue = append(mc.queue[:i], mc.queue[i+1:]...)
			break
		}
	}
}

// cleanupRoutine periodically removes expired entries
func (mc *MemoryCache) cleanupRoutine() {
	ticker := time.NewTicker(mc.config.CleanupPeriod)
	defer ticker.Stop()

	removedCount := 0

	for {
		select {
		case <-ticker.C:
			mc.mu.Lock()
			now := time.Now()
			keysToRemove := make([]string, 0)

			for key, entry := range mc.cache {
				if now.After(entry.ExpiresAt) {
					keysToRemove = append(keysToRemove, key)
				}
			}

			for _, key := range keysToRemove {
				delete(mc.cache, key)
				mc.removeFromQueue(key)
				removedCount++
			}

			if mc.config.EnableMetrics {
				mc.metrics.setSize(len(mc.cache))
			}

			mc.mu.Unlock()

			if removedCount > 0 {
				slog.Debug("Cache cleanup completed", "removed", removedCount)
				removedCount = 0
			}

		case <-mc.stop:
			return
		}
	}
}

// CachedIndexer wraps an Indexer with caching capabilities
type CachedIndexer struct {
	indexer *Indexer
	cache   *MemoryCache
	config  CacheConfig
	engine  *EmbeddingEngine // Add embedding engine for similar search
}

// NewCachedIndexer creates a new cached indexer
func NewCachedIndexer(indexer *Indexer, config CacheConfig, engine *EmbeddingEngine) *CachedIndexer {
	return &CachedIndexer{
		indexer: indexer,
		cache:   NewMemoryCache(config),
		config:  config,
		engine:  engine,
	}
}

// SearchSymbols searches symbols with caching
func (ci *CachedIndexer) SearchSymbols(ctx context.Context, query string, limit int) ([]Symbol, error) {
	cacheKey := fmt.Sprintf("search_symbols:%s:%d", query, limit)

	// Try cache first
	if cached, found := ci.cache.Get(ctx, cacheKey); found {
		if symbols, ok := cached.([]Symbol); ok {
			slog.Debug("Cache hit for search symbols", "query", query)
			return symbols, nil
		}
	}

	// Cache miss - perform actual search
	symbols, err := ci.indexer.SearchSymbols(ctx, query, limit)
	if err != nil {
		return nil, err
	}

	// Store in cache
	ci.cache.Set(ctx, cacheKey, symbols)

	return symbols, nil
}

// SearchSimilar searches similar symbols with caching
func (ci *CachedIndexer) SearchSimilar(ctx context.Context, query string, limit int) ([]Embedding, error) {
	cacheKey := fmt.Sprintf("search_similar:%s:%d", query, limit)

	// Try cache first
	if cached, found := ci.cache.Get(ctx, cacheKey); found {
		if embeddings, ok := cached.([]Embedding); ok {
			slog.Debug("Cache hit for search similar", "query", query)
			return embeddings, nil
		}
		// Handle potential type assertion failure safely
	}

	// Cache miss - delegate to embedding engine if available
	if ci.engine != nil {
		embeddings, err := ci.engine.SearchSimilar(ctx, query, limit)
		if err != nil {
			return nil, err
		}

		// Store in cache
		ci.cache.Set(ctx, cacheKey, embeddings)
		return embeddings, nil
	}

	// Fallback: return empty result if no embedding engine
	return []Embedding{}, nil
}

// GetSymbol retrieves a symbol with caching
func (ci *CachedIndexer) GetSymbol(ctx context.Context, id string) (*Symbol, error) {
	cacheKey := fmt.Sprintf("get_symbol:%s", id)

	// Try cache first
	if cached, found := ci.cache.Get(ctx, cacheKey); found {
		if symbol, ok := cached.(*Symbol); ok {
			slog.Debug("Cache hit for get symbol", "id", id)
			return symbol, nil
		}
		// Type assertion failed, fall through to retrieve fresh data
	}

	// Cache miss - perform actual retrieval
	symbol, err := ci.indexer.GetSymbol(ctx, id)
	if err != nil {
		return nil, err
	}

	// Store in cache
	ci.cache.Set(ctx, cacheKey, symbol)

	return symbol, nil
}

// InvalidateFile removes all cached entries for a specific file
func (ci *CachedIndexer) InvalidateFile(ctx context.Context, filePath string) {
	ci.cache.Clear(ctx) // For simplicity, clear all cache
	// In a more sophisticated implementation, we could track which cache keys
	// are related to which files and invalidate only those
}

// GetCacheMetrics returns the cache performance metrics
func (ci *CachedIndexer) GetCacheMetrics() CacheMetrics {
	return ci.cache.GetMetrics()
}

// Close closes the cached indexer
func (ci *CachedIndexer) Close() error {
	ci.cache.Close()
	return ci.indexer.Close()
}
