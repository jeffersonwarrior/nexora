package prompt

import (
	"context"
	"testing"
	"time"
)

func TestEnvironmentCache(t *testing.T) {
	ctx := context.Background()

	t.Run("cache initialization", func(t *testing.T) {
		cache := NewEnvironmentCache(1 * time.Minute)
		if cache == nil {
			t.Fatal("NewEnvironmentCache returned nil")
		}
		if cache.ttl != 1*time.Minute {
			t.Errorf("Expected TTL of 1 minute, got %v", cache.ttl)
		}
	})

	t.Run("cache refresh and retrieval", func(t *testing.T) {
		cache := NewEnvironmentCache(1 * time.Minute)

		// First call should populate cache
		data1, err := cache.Get(ctx, "/tmp", false)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		// Verify some data was populated
		if data1.Architecture == "" {
			t.Error("Expected Architecture to be populated")
		}

		// Second call should use cache (same data)
		data2, err := cache.Get(ctx, "/tmp", false)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		// Should be same data (from cache)
		if data1.Architecture != data2.Architecture {
			t.Error("Expected cached data to match")
		}
	})

	t.Run("cache expiration", func(t *testing.T) {
		cache := NewEnvironmentCache(100 * time.Millisecond)

		// First call
		_ , err := cache.Get(ctx, "/tmp", false)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		// Wait for cache to expire
		time.Sleep(150 * time.Millisecond)

		// Should refresh cache
		data2, err := cache.Get(ctx, "/tmp", false)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		// Data might be same or different, but should have valid values
		if data2.Architecture == "" {
			t.Error("Expected Architecture to be populated after refresh")
		}
	})

	t.Run("cache invalidation", func(t *testing.T) {
		cache := NewEnvironmentCache(1 * time.Minute)

		// Populate cache
		_, err := cache.Get(ctx, "/tmp", false)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		// Invalidate
		cache.Invalidate()

		// Next call should refresh (we can't easily verify this without
		// inspecting internals, but at least test it doesn't error)
		_, err = cache.Get(ctx, "/tmp", false)
		if err != nil {
			t.Fatalf("Get after invalidation failed: %v", err)
		}
	})

	t.Run("full env vs normal mode", func(t *testing.T) {
		cache := NewEnvironmentCache(1 * time.Minute)

		// Normal mode (fullEnv=false)
		data1, err := cache.Get(ctx, "/tmp", false)
		if err != nil {
			t.Fatalf("Get failed: %v", err)
		}

		// Should use defaults
		if data1.NetworkStatus != "online" {
			t.Errorf("Expected default network status 'online', got '%s'", data1.NetworkStatus)
		}
		if data1.ActiveServices != "" {
			t.Errorf("Expected empty active services in normal mode, got '%s'", data1.ActiveServices)
		}

		// Invalidate to force refresh with fullEnv
		cache.Invalidate()

		// Full env mode (fullEnv=true) - may take longer but should work
		data2, err := cache.Get(ctx, "/tmp", true)
		if err != nil {
			t.Fatalf("Get with fullEnv failed: %v", err)
		}

		// Network status should still be populated (may be "online" or actual check result)
		if data2.NetworkStatus == "" {
			t.Error("Expected NetworkStatus to be populated in full env mode")
		}
	})

	t.Run("parallel access", func(t *testing.T) {
		cache := NewEnvironmentCache(1 * time.Minute)

		// Simulate multiple concurrent accesses
		done := make(chan bool, 10)
		for i := 0; i < 10; i++ {
			go func() {
				_, err := cache.Get(ctx, "/tmp", false)
				if err != nil {
					t.Errorf("Concurrent Get failed: %v", err)
				}
				done <- true
			}()
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
