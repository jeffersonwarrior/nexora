package resources

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	cfg := DefaultConfig()

	require.Equal(t, 80.0, cfg.CPUThreshold)
	require.Equal(t, 85.0, cfg.MemThreshold)
	require.Equal(t, uint64(5*1024*1024*1024), cfg.DiskMinFree)
	require.Equal(t, 5*time.Second, cfg.CheckInterval)
	require.True(t, cfg.EnableAutoPause)
	require.Equal(t, 3, cfg.MaxViolations)
}

func TestNewMonitor(t *testing.T) {
	t.Parallel()

	t.Run("with default config", func(t *testing.T) {
		m := NewMonitor(DefaultConfig())
		require.NotNil(t, m)
		require.False(t, m.IsRunning())
	})

	t.Run("with custom config", func(t *testing.T) {
		cfg := Config{
			CPUThreshold:  90.0,
			MemThreshold:  95.0,
			CheckInterval: 10 * time.Second,
		}
		m := NewMonitor(cfg)
		require.Equal(t, 90.0, m.config.CPUThreshold)
		require.Equal(t, 95.0, m.config.MemThreshold)
	})

	t.Run("applies defaults for zero values", func(t *testing.T) {
		m := NewMonitor(Config{})
		require.Equal(t, 80.0, m.config.CPUThreshold)
		require.Equal(t, 85.0, m.config.MemThreshold)
		require.Equal(t, uint64(5*1024*1024*1024), m.config.DiskMinFree)
		require.Equal(t, 5*time.Second, m.config.CheckInterval)
	})
}

func TestMonitorStartStop(t *testing.T) {
	t.Parallel()

	m := NewMonitor(DefaultConfig())
	ctx := context.Background()

	// Start
	err := m.Start(ctx)
	require.NoError(t, err)
	require.True(t, m.IsRunning())

	// Start again (should be no-op)
	err = m.Start(ctx)
	require.NoError(t, err)
	require.True(t, m.IsRunning())

	// Stop
	m.Stop()
	require.False(t, m.IsRunning())

	// Stop again (should be no-op)
	m.Stop()
	require.False(t, m.IsRunning())
}

func TestMonitorCollectSnapshot(t *testing.T) {
	t.Parallel()

	m := NewMonitor(DefaultConfig())
	snapshot := m.collectSnapshot()

	// Should have collected something
	require.NotZero(t, snapshot.Timestamp)

	// CPU should be 0-100
	require.GreaterOrEqual(t, snapshot.CPUPercent, 0.0)
	require.LessOrEqual(t, snapshot.CPUPercent, 100.0)

	// Memory should be reasonable
	require.Greater(t, snapshot.MemTotal, uint64(0))
	require.LessOrEqual(t, snapshot.MemUsed, snapshot.MemTotal)
	require.GreaterOrEqual(t, snapshot.MemPercent, 0.0)
	require.LessOrEqual(t, snapshot.MemPercent, 100.0)

	// Disk should be reasonable
	require.Greater(t, snapshot.DiskTotal, uint64(0))
	require.Greater(t, snapshot.DiskFree, uint64(0))
	require.LessOrEqual(t, snapshot.DiskUsed, snapshot.DiskTotal)

	t.Logf("Snapshot: %s", snapshot.Summary())
}

func TestMonitorCallbacks(t *testing.T) {
	t.Parallel()

	cpuCalled := false
	memCalled := false
	diskCalled := false
	violationCalled := false

	m := NewMonitor(Config{
		CPUThreshold:  0.1, // Very low to trigger
		MemThreshold:  0.1,
		DiskMinFree:   1024 * 1024 * 1024 * 1024 * 1024, // 1PB to trigger
		CheckInterval: 100 * time.Millisecond,
	})

	m.SetCallbacks(
		func(usage float64) {
			cpuCalled = true
			t.Logf("CPU callback: %.2f%%", usage)
		},
		func(usage uint64) {
			memCalled = true
			t.Logf("Memory callback: %d bytes", usage)
		},
		func(free uint64) {
			diskCalled = true
			t.Logf("Disk callback: %d bytes free", free)
		},
		func(v Violation) {
			violationCalled = true
			t.Logf("Violation: %s", v.String())
		},
	)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := m.Start(ctx)
	require.NoError(t, err)

	// Wait for checks
	time.Sleep(300 * time.Millisecond)

	m.Stop()

	// At least one callback should have been triggered
	// (CPU and memory are very likely to exceed 0.1%, disk might not)
	require.True(t, cpuCalled || memCalled || diskCalled, "at least one callback should be called")
	require.True(t, violationCalled, "violation callback should be called")
}

func TestMonitorViolations(t *testing.T) {
	t.Parallel()

	m := NewMonitor(DefaultConfig())

	// Manually trigger violations
	m.handleViolation(Violation{
		Type:      ViolationTypeCPU,
		Timestamp: time.Now(),
		Value:     90.0,
		Threshold: 80.0,
		Message:   "CPU high",
	})

	violations := m.GetViolations(10)
	require.Len(t, violations, 1)
	require.Equal(t, ViolationTypeCPU, violations[0].Type)
	require.Equal(t, 90.0, violations[0].Value)
}

func TestResourceSnapshotSummary(t *testing.T) {
	t.Parallel()

	snapshot := ResourceSnapshot{
		Timestamp:   time.Now(),
		CPUPercent:  45.5,
		MemUsed:     8 * 1024 * 1024 * 1024,  // 8GB
		MemTotal:    16 * 1024 * 1024 * 1024, // 16GB
		MemPercent:  50.0,
		DiskUsed:    100 * 1024 * 1024 * 1024, // 100GB
		DiskTotal:   500 * 1024 * 1024 * 1024, // 500GB
		DiskFree:    400 * 1024 * 1024 * 1024, // 400GB
		DiskPercent: 20.0,
	}

	summary := snapshot.Summary()
	require.Contains(t, summary, "45.5%")    // CPU
	require.Contains(t, summary, "50.0%")    // Memory
	require.Contains(t, summary, "400.00GB") // Disk free

	t.Logf("Summary: %s", summary)
}

func TestViolationString(t *testing.T) {
	t.Parallel()

	v := Violation{
		Type:      ViolationTypeCPU,
		Timestamp: time.Now(),
		Value:     95.5,
		Threshold: 80.0,
		Message:   "CPU usage too high",
	}

	str := v.String()
	require.Contains(t, str, "cpu")
	require.Contains(t, str, "95.50")
	require.Contains(t, str, "80.00")
	require.Contains(t, str, "CPU usage too high")

	t.Logf("Violation: %s", str)
}

func TestMonitorIntegration(t *testing.T) {
	t.Parallel()

	// Create monitor with reasonable thresholds
	m := NewMonitor(Config{
		CPUThreshold:    80.0,
		MemThreshold:    85.0,
		DiskMinFree:     5 * 1024 * 1024 * 1024,
		CheckInterval:   100 * time.Millisecond,
		EnableAutoPause: false, // Disable auto-pause for test
		MaxViolations:   3,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := m.Start(ctx)
	require.NoError(t, err)

	// Wait for several checks
	time.Sleep(300 * time.Millisecond)

	// Get current snapshot
	snapshot := m.CurrentSnapshot()
	require.NotZero(t, snapshot.Timestamp)
	require.GreaterOrEqual(t, snapshot.CPUPercent, 0.0)

	t.Logf("Final snapshot: %s", snapshot.Summary())

	m.Stop()
}
