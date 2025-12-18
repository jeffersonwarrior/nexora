package indexer

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"
)

// FileChangeEvent represents a file system change event
type FileChangeEvent struct {
	Path      string    `json:"path"`
	Type      EventType `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Size      int64     `json:"size,omitempty"`
	Checksum  string    `json:"checksum,omitempty"`
}

// EventType represents the type of file change
type EventType string

const (
	EventAdded    EventType = "added"
	EventModified EventType = "modified"
	EventRemoved  EventType = "removed"
)

// String returns the string representation of EventType
func (e EventType) String() string {
	return string(e)
}

// EventHandler defines the interface for processing file change events
type EventHandler interface {
	HandleEvent(ctx context.Context, event FileChangeEvent) error
	Name() string
	Priority() int
}

// EventBus manages event distribution to handlers
type EventBus struct {
	handlers  map[string][]EventHandler
	queuing   bool
	queue     []FileChangeEvent
	ctx       context.Context
	cancel    context.CancelFunc
	batchSize int
}

// NewEventBus creates a new event bus
func NewEventBus(ctx context.Context) *EventBus {
	busCtx, cancel := context.WithCancel(ctx)
	return &EventBus{
		handlers:  make(map[string][]EventHandler),
		queuing:   false,
		queue:     make([]FileChangeEvent, 0),
		ctx:       busCtx,
		cancel:    cancel,
		batchSize: 10,
	}
}

// RegisterHandler registers an event handler for specific event types
func (eb *EventBus) RegisterHandler(eventTypes []EventType, handler EventHandler) {
	for _, eventType := range eventTypes {
		eventTypeStr := string(eventType)
		eb.handlers[eventTypeStr] = append(eb.handlers[eventTypeStr], handler)
	}

	slog.Info("Event handler registered",
		"handler", handler.Name(),
		"types", eventTypes,
		"priority", handler.Priority())
}

// PublishEvent publishes an event to all registered handlers
func (eb *EventBus) PublishEvent(event FileChangeEvent) error {
	if eb.queuing {
		eb.queue = append(eb.queue, event)
		return nil
	}

	return eb.processEvent(event)
}

// PublishEvents publishes multiple events
func (eb *EventBus) PublishEvents(events []FileChangeEvent) error {
	for _, event := range events {
		if err := eb.PublishEvent(event); err != nil {
			slog.Warn("Failed to process event", "event", event, "error", err)
		}
	}
	return nil
}

// StartQueuing enables event queuing
func (eb *EventBus) StartQueuing() {
	eb.queuing = true
	eb.queue = make([]FileChangeEvent, 0)
	slog.Debug("Event queuing started")
}

// StopQueuing stops queuing and processes all queued events
func (eb *EventBus) StopQueuing() error {
	eb.queuing = false

	events := make([]FileChangeEvent, len(eb.queue))
	copy(events, eb.queue)
	eb.queue = make([]FileChangeEvent, 0)

	slog.Info("Event queuing stopped", "queued_events", len(events))
	return eb.PublishEvents(events)
}

// processEvent processes a single event
func (eb *EventBus) processEvent(event FileChangeEvent) error {
	eventTypeStr := string(event.Type)
	handlers, exists := eb.handlers[eventTypeStr]
	if !exists {
		slog.Debug("No handlers for event type", "type", event.Type)
		return nil
	}

	// Sort handlers by priority (higher first)
	sortedHandlers := make([]EventHandler, len(handlers))
	copy(sortedHandlers, handlers)

	// Simple sort by priority
	for i := 0; i < len(sortedHandlers)-1; i++ {
		for j := i + 1; j < len(sortedHandlers); j++ {
			if sortedHandlers[i].Priority() < sortedHandlers[j].Priority() {
				sortedHandlers[i], sortedHandlers[j] = sortedHandlers[j], sortedHandlers[i]
			}
		}
	}

	// Execute handlers
	var lastError error
	for _, handler := range sortedHandlers {
		select {
		case <-eb.ctx.Done():
			return eb.ctx.Err()
		default:
		}

		start := time.Now()
		err := handler.HandleEvent(eb.ctx, event)
		duration := time.Since(start)

		if err != nil {
			slog.Warn("Event handler failed",
				"handler", handler.Name(),
				"event", event.Path,
				"error", err,
				"duration", duration)
			lastError = err
		} else {
			slog.Debug("Event handler succeeded",
				"handler", handler.Name(),
				"event", event.Path,
				"duration", duration)
		}
	}

	return lastError
}

// GetStats returns event bus statistics
func (eb *EventBus) GetStats() EventBusStats {
	return EventBusStats{
		TotalHandlers: len(eb.handlers),
		QueuedEvents:  len(eb.queue),
		IsQueuing:     eb.queuing,
	}
}

// EventBusStats represents event bus statistics
type EventBusStats struct {
	TotalHandlers int  `json:"total_handlers"`
	QueuedEvents  int  `json:"queued_events"`
	IsQueuing     bool `json:"is_queuing"`
}

// Concrete event handlers

// SymbolUpdateHandler updates symbols when files change
type SymbolUpdateHandler struct {
	indexer  SymbolStore
	parser   CodeParser
	engine   EmbeddingGenerator
	priority int
}

// NewSymbolUpdateHandler creates a symbol update handler
func NewSymbolUpdateHandler(indexer SymbolStore, parser CodeParser, engine EmbeddingGenerator) *SymbolUpdateHandler {
	return &SymbolUpdateHandler{
		indexer:  indexer,
		parser:   parser,
		engine:   engine,
		priority: 100, // High priority
	}
}

func (h *SymbolUpdateHandler) HandleEvent(ctx context.Context, event FileChangeEvent) error {
	switch event.Type {
	case EventAdded, EventModified:
		return h.handleAddOrUpdate(ctx, event)
	case EventRemoved:
		return h.handleRemove(ctx, event)
	default:
		return nil
	}
}

func (h *SymbolUpdateHandler) handleAddOrUpdate(ctx context.Context, event FileChangeEvent) error {
	// Validate file is supported
	if !h.shouldProcessFile(event.Path) {
		return nil
	}

	// Parse symbols from file
	symbols, err := h.parser.ParseFile(ctx, event.Path)
	if err != nil {
		return fmt.Errorf("failed to parse file %s: %w", event.Path, err)
	}

	// Store symbols
	if err := h.indexer.StoreSymbols(ctx, symbols); err != nil {
		return fmt.Errorf("failed to store symbols for %s: %w", event.Path, err)
	}

	// Generate embeddings if available
	if h.engine != nil {
		embeddings, err := h.engine.GenerateSymbolEmbeddings(ctx, symbols)
		if err != nil {
			slog.Warn("Failed to generate embeddings", "file", event.Path, "error", err)
		} else if storage, ok := h.engine.(EmbeddingStore); ok {
			if err := storage.StoreEmbeddings(ctx, embeddings); err != nil {
				slog.Warn("Failed to store embeddings", "file", event.Path, "error", err)
			}
		}
	}

	slog.Info("File indexed successfully",
		"file", event.Path,
		"symbols", len(symbols),
		"event", event.Type)

	return nil
}

func (h *SymbolUpdateHandler) handleRemove(ctx context.Context, event FileChangeEvent) error {
	// Remove symbols for deleted file
	if err := h.indexer.DeleteSymbolsByFile(ctx, event.Path); err != nil {
		return fmt.Errorf("failed to delete symbols for %s: %w", event.Path, err)
	}

	slog.Info("File removed from index", "file", event.Path)
	return nil
}

func (h *SymbolUpdateHandler) shouldProcessFile(path string) bool {
	// Only process Go files
	if !strings.HasSuffix(path, ".go") {
		return false
	}

	// Skip test files for now
	if strings.HasSuffix(path, "_test.go") {
		return false
	}

	return true
}

func (h *SymbolUpdateHandler) Name() string {
	return "SymbolUpdateHandler"
}

func (h *SymbolUpdateHandler) Priority() int {
	return h.priority
}

// CacheInvalidationHandler invalidates cache entries when files change
type CacheInvalidationHandler struct {
	cache    CacheManager
	priority int
}

// NewCacheInvalidationHandler creates a cache invalidation handler
func NewCacheInvalidationHandler(cache CacheManager) *CacheInvalidationHandler {
	return &CacheInvalidationHandler{
		cache:    cache,
		priority: 50, // Medium priority
	}
}

func (h *CacheInvalidationHandler) HandleEvent(ctx context.Context, event FileChangeEvent) error {
	// For now, clear all cache entries on any file change
	// In a more sophisticated implementation, we could track which cache keys
	// are related to which files and invalidate only those
	h.cache.Clear(ctx)

	slog.Debug("Cache cleared due to file change", "file", event.Path, "event", event.Type)
	return nil
}

func (h *CacheInvalidationHandler) Name() string {
	return "CacheInvalidationHandler"
}

func (h *CacheInvalidationHandler) Priority() int {
	return h.priority
}

// LoggingHandler logs all file change events
type LoggingHandler struct {
	priority int
}

// NewLoggingHandler creates a logging handler
func NewLoggingHandler() *LoggingHandler {
	return &LoggingHandler{
		priority: 10, // Low priority
	}
}

func (h *LoggingHandler) HandleEvent(ctx context.Context, event FileChangeEvent) error {
	slog.Info("File system event",
		"event", event.Type,
		"path", event.Path,
		"timestamp", event.Timestamp,
		"size", event.Size)

	return nil
}

func (h *LoggingHandler) Name() string {
	return "LoggingHandler"
}

func (h *LoggingHandler) Priority() int {
	return h.priority
}

// Helper functions for creating FileChangeEvents

// NewFileChangeEvent creates a new file change event
func NewFileChangeEvent(path string, eventType EventType) FileChangeEvent {
	return FileChangeEvent{
		Path:      path,
		Type:      eventType,
		Timestamp: time.Now(),
	}
}

// CreateBatchFromFSWatch creates a batch of file change events from file system changes
func CreateBatchFromFSWatch(added, modified, removed []string) []FileChangeEvent {
	events := make([]FileChangeEvent, 0)

	for _, path := range added {
		events = append(events, NewFileChangeEvent(path, EventAdded))
	}

	for _, path := range modified {
		events = append(events, NewFileChangeEvent(path, EventModified))
	}

	for _, path := range removed {
		events = append(events, NewFileChangeEvent(path, EventRemoved))
	}

	return events
}

// IsGoFile checks if a file is a Go source file
func IsGoFile(path string) bool {
	return strings.HasSuffix(path, ".go")
}

// IsTestFile checks if a file is a test file
func IsTestFile(path string) bool {
	base := filepath.Base(path)
	return strings.HasSuffix(base, "_test.go") || strings.HasPrefix(base, "example_")
}
