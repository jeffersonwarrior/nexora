package prompt

import (
	"context"
	"testing"
	"time"

	"github.com/nexora/nexora/internal/config"
)

// BenchmarkPromptBuild measures prompt template parsing and execution
func BenchmarkPromptBuild(b *testing.B) {
	tempDir := b.TempDir()
	cfg, err := config.Load(tempDir, tempDir, false)
	if err != nil {
		b.Fatalf("Failed to load config: %v", err)
	}

	fixedTime := time.Date(2025, 12, 18, 12, 0, 0, 0, time.UTC)
	p, err := NewPrompt("benchmark", "{{.DateTime}} {{.Platform}} {{.CurrentUser}}", WithTimeFunc(func() time.Time {
		return fixedTime
	}), WithWorkingDir(tempDir))
	if err != nil {
		b.Fatalf("Failed to create prompt: %v", err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.Build(ctx, "openai", "gpt-4", *cfg)
		if err != nil {
			b.Fatalf("Build failed: %v", err)
		}
	}
}

// BenchmarkEnvironmentDetection measures environment detection (uncached)
func BenchmarkEnvironmentDetection(b *testing.B) {
	ctx := context.Background()
	workingDir := b.TempDir()

	// Force cache invalidation for each iteration
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		envCache.Invalidate()
		_, err := envCache.Get(ctx, workingDir, false)
		if err != nil {
			b.Fatalf("Environment detection failed: %v", err)
		}
	}
}

// BenchmarkEnvironmentDetectionCached measures cached environment lookups
func BenchmarkEnvironmentDetectionCached(b *testing.B) {
	ctx := context.Background()
	workingDir := b.TempDir()

	// Prime the cache
	_, _ = envCache.Get(ctx, workingDir, false)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := envCache.Get(ctx, workingDir, false)
		if err != nil {
			b.Fatalf("Environment detection failed: %v", err)
		}
	}
}

// BenchmarkGetMemoryInfo measures memory info detection
func BenchmarkGetMemoryInfo(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getMemoryInfo(ctx)
	}
}

// BenchmarkGetDiskInfo measures disk info detection
func BenchmarkGetDiskInfo(b *testing.B) {
	ctx := context.Background()
	workingDir := b.TempDir()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getDiskInfo(ctx, workingDir)
	}
}

// BenchmarkDetectContainer measures container detection
func BenchmarkDetectContainer(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = detectContainer(ctx)
	}
}

// BenchmarkGetNetworkStatus measures network status detection
func BenchmarkGetNetworkStatus(b *testing.B) {
	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = getNetworkStatus(ctx)
	}
}

// BenchmarkPromptDataFull measures full prompt data collection
func BenchmarkPromptDataFull(b *testing.B) {
	tempDir := b.TempDir()
	cfg, err := config.Load(tempDir, tempDir, false)
	if err != nil {
		b.Fatalf("Failed to load config: %v", err)
	}

	fixedTime := time.Date(2025, 12, 18, 12, 0, 0, 0, time.UTC)
	p, err := NewPrompt("benchmark", "{{.DateTime}}", WithTimeFunc(func() time.Time {
		return fixedTime
	}), WithWorkingDir(tempDir))
	if err != nil {
		b.Fatalf("Failed to create prompt: %v", err)
	}

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := p.promptData(ctx, "openai", "gpt-4", *cfg)
		if err != nil {
			b.Fatalf("promptData failed: %v", err)
		}
	}
}
