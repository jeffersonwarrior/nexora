// Package resources provides resource monitoring and protection for Nexora agents.
package resources

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/nexora/nexora/internal/agent/state"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

// Monitor provides resource monitoring with configurable thresholds.
type Monitor struct {
	mu sync.RWMutex

	// Configuration
	config Config

	// State
	running      bool
	cancel       context.CancelFunc
	violations   []Violation
	lastSnapshot ResourceSnapshot

	// Callbacks
	onCPUHigh   func(usage float64)
	onMemHigh   func(usage uint64)
	onDiskLow   func(free uint64)
	onViolation func(v Violation)

	// State machine integration (optional)
	stateMachine *state.StateMachine
}

// Config defines resource monitoring configuration.
type Config struct {
	// Thresholds
	CPUThreshold  float64       // CPU usage percentage (0-100), default 80
	MemThreshold  float64       // Memory usage percentage (0-100), default 85
	DiskMinFree   uint64        // Minimum free disk space in bytes, default 5GB
	CheckInterval time.Duration // Check interval, default 5s

	// Actions
	EnableAutoPause bool // Pause agent on resource violation
	MaxViolations   int  // Max violations before halting, default 3
}

// DefaultConfig returns default monitoring configuration.
func DefaultConfig() Config {
	return Config{
		CPUThreshold:    80.0,
		MemThreshold:    85.0,
		DiskMinFree:     5 * 1024 * 1024 * 1024, // 5GB
		CheckInterval:   5 * time.Second,
		EnableAutoPause: true,
		MaxViolations:   3,
	}
}

// NewMonitor creates a new resource monitor.
func NewMonitor(config Config) *Monitor {
	if config.CheckInterval == 0 {
		config.CheckInterval = 5 * time.Second
	}
	if config.CPUThreshold == 0 {
		config.CPUThreshold = 80.0
	}
	if config.MemThreshold == 0 {
		config.MemThreshold = 85.0
	}
	if config.DiskMinFree == 0 {
		config.DiskMinFree = 5 * 1024 * 1024 * 1024
	}
	if config.MaxViolations == 0 {
		config.MaxViolations = 3
	}

	return &Monitor{
		config:     config,
		violations: make([]Violation, 0),
	}
}

// SetStateMachine sets the state machine for integration.
func (m *Monitor) SetStateMachine(sm *state.StateMachine) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stateMachine = sm
}

// SetCallbacks sets resource violation callbacks.
func (m *Monitor) SetCallbacks(
	onCPUHigh func(float64),
	onMemHigh func(uint64),
	onDiskLow func(uint64),
	onViolation func(Violation),
) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.onCPUHigh = onCPUHigh
	m.onMemHigh = onMemHigh
	m.onDiskLow = onDiskLow
	m.onViolation = onViolation
}

// Start begins resource monitoring.
func (m *Monitor) Start(ctx context.Context) error {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return nil
	}

	monitorCtx, cancel := context.WithCancel(ctx)
	m.cancel = cancel
	m.running = true
	m.mu.Unlock()

	go m.monitorLoop(monitorCtx)

	slog.Info("resource monitor started",
		"cpu_threshold", m.config.CPUThreshold,
		"mem_threshold", m.config.MemThreshold,
		"disk_min_free_gb", float64(m.config.DiskMinFree)/(1024*1024*1024),
		"check_interval", m.config.CheckInterval,
	)

	return nil
}

// Stop stops resource monitoring.
func (m *Monitor) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running {
		return
	}

	if m.cancel != nil {
		m.cancel()
	}

	m.running = false
	slog.Info("resource monitor stopped")
}

// monitorLoop runs the monitoring loop.
func (m *Monitor) monitorLoop(ctx context.Context) {
	ticker := time.NewTicker(m.config.CheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.check()
		}
	}
}

// check performs a resource check.
func (m *Monitor) check() {
	snapshot := m.collectSnapshot()

	m.mu.Lock()
	m.lastSnapshot = snapshot
	m.mu.Unlock()

	// Check for violations
	if snapshot.CPUPercent > m.config.CPUThreshold {
		m.handleViolation(Violation{
			Type:      ViolationTypeCPU,
			Timestamp: time.Now(),
			Value:     snapshot.CPUPercent,
			Threshold: m.config.CPUThreshold,
			Message:   fmt.Sprintf("CPU usage %.1f%% exceeds threshold %.1f%%", snapshot.CPUPercent, m.config.CPUThreshold),
		})
	}

	if snapshot.MemPercent > m.config.MemThreshold {
		m.handleViolation(Violation{
			Type:      ViolationTypeMem,
			Timestamp: time.Now(),
			Value:     snapshot.MemPercent,
			Threshold: m.config.MemThreshold,
			Message:   fmt.Sprintf("Memory usage %.1f%% exceeds threshold %.1f%%", snapshot.MemPercent, m.config.MemThreshold),
		})
	}

	if snapshot.DiskFree < m.config.DiskMinFree {
		m.handleViolation(Violation{
			Type:      ViolationTypeDisk,
			Timestamp: time.Now(),
			Value:     float64(snapshot.DiskFree),
			Threshold: float64(m.config.DiskMinFree),
			Message:   fmt.Sprintf("Disk free %.2fGB below threshold %.2fGB", float64(snapshot.DiskFree)/(1024*1024*1024), float64(m.config.DiskMinFree)/(1024*1024*1024)),
		})
	}
}

// collectSnapshot collects current resource usage.
func (m *Monitor) collectSnapshot() ResourceSnapshot {
	snapshot := ResourceSnapshot{
		Timestamp: time.Now(),
	}

	// CPU usage
	if cpuPercents, err := cpu.Percent(0, false); err == nil && len(cpuPercents) > 0 {
		snapshot.CPUPercent = cpuPercents[0]
	}

	// Memory usage
	if vmem, err := mem.VirtualMemory(); err == nil {
		snapshot.MemUsed = vmem.Used
		snapshot.MemTotal = vmem.Total
		snapshot.MemPercent = vmem.UsedPercent
	}

	// Disk usage (current directory)
	if usage, err := disk.Usage("."); err == nil {
		snapshot.DiskUsed = usage.Used
		snapshot.DiskTotal = usage.Total
		snapshot.DiskFree = usage.Free
		snapshot.DiskPercent = usage.UsedPercent
	}

	return snapshot
}

// handleViolation processes a resource violation.
func (m *Monitor) handleViolation(v Violation) {
	m.mu.Lock()
	m.violations = append(m.violations, v)
	if len(m.violations) > 20 {
		m.violations = m.violations[1:]
	}
	violationCount := len(m.violations)
	m.mu.Unlock()

	slog.Warn("resource violation",
		"type", v.Type,
		"value", v.Value,
		"threshold", v.Threshold,
		"message", v.Message,
		"total_violations", violationCount,
	)

	// Call type-specific callbacks
	switch v.Type {
	case ViolationTypeCPU:
		if m.onCPUHigh != nil {
			m.onCPUHigh(v.Value)
		}
	case ViolationTypeMem:
		if m.onMemHigh != nil {
			m.onMemHigh(uint64(v.Value))
		}
	case ViolationTypeDisk:
		if m.onDiskLow != nil {
			m.onDiskLow(uint64(v.Value))
		}
	}

	// Call general violation callback
	if m.onViolation != nil {
		m.onViolation(v)
	}

	// Auto-pause on max violations
	if m.config.EnableAutoPause && violationCount >= m.config.MaxViolations {
		if m.stateMachine != nil {
			slog.Error("max resource violations reached - pausing agent",
				"violations", violationCount,
				"max", m.config.MaxViolations,
			)
			// Transition to resource paused state
			_ = m.stateMachine.TransitionTo(state.StateResourcePaused)
		}
	}
}

// CurrentSnapshot returns the most recent resource snapshot.
func (m *Monitor) CurrentSnapshot() ResourceSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.lastSnapshot
}

// GetViolations returns recent violations.
func (m *Monitor) GetViolations(limit int) []Violation {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if limit <= 0 || limit > len(m.violations) {
		limit = len(m.violations)
	}

	start := len(m.violations) - limit
	result := make([]Violation, limit)
	copy(result, m.violations[start:])

	return result
}

// IsRunning returns true if the monitor is active.
func (m *Monitor) IsRunning() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.running
}
