package indexer

import (
	"context"
	"testing"

	_ "github.com/ncruces/go-sqlite3"
	_ "github.com/ncruces/go-sqlite3/embed"
)

func TestP5Interfaces(t *testing.T) {
	ctx := context.Background()

	t.Run("Dependency Injection", func(t *testing.T) {
		// Create a complete indexer system with DI
		config := DefaultIndexerSystemConfig()
		builder, err := CreateIndexerSystem(config)
		if err != nil {
			t.Fatalf("Failed to create indexer system: %v", err)
		}

		// Test built info
		info := builder.GetBuildInfo()
		if !info.IsBuilt {
			t.Error("System should be built")
		}
		if info.ServiceCount == 0 {
			t.Error("Should have built services")
		}
	})

	t.Run("Event System", func(t *testing.T) {
		// Test event bus
		bus := NewEventBus(ctx)

		// Test logging handler
		loggingHandler := NewLoggingHandler()
		bus.RegisterHandler([]EventType{EventAdded, EventRemoved}, loggingHandler)

		// Test event
		event := NewFileChangeEvent("/test/file.go", EventAdded)
		err := bus.PublishEvent(event)
		if err != nil {
			t.Fatalf("Failed to publish event: %v", err)
		}

		// Test stats
		stats := bus.GetStats()
		if stats.TotalHandlers == 0 {
			t.Error("Should have registered handlers")
		}
	})

	t.Run("File Type Helpers", func(t *testing.T) {
		if !IsGoFile("test.go") {
			t.Error("test.go should be identified as Go file")
		}
		if IsGoFile("test.txt") {
			t.Error("test.txt should not be identified as Go file")
		}
		if !IsTestFile("example_test.go") {
			t.Error("example_test.go should be identified as test file")
		}
	})

	t.Run("Factory Pattern", func(t *testing.T) {
		container := NewDIContainer()

		// Test indexer factory
		indexerFactory := NewIndexerServiceFactory(SymbolStoreConfig{
			DatabasePath: ":memory:",
		})

		_, err := indexerFactory.Create(ctx, container)
		if err != nil {
			t.Fatalf("Failed to create indexer: %v", err)
		}

		// Verify service exists
		if !container.Has("indexer") {
			t.Error("Indexer service should be registered")
		}
	})
}
