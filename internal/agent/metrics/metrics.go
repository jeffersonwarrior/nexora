// Package metrics provides performance monitoring and observability for agent tools and operations.
package metrics

import (
	"sync"
	"time"
)

// Metric represents a single tool execution metric.
type Metric struct {
	ToolName      string        `json:"tool_name"`
	ExecutionTime time.Duration `json:"execution_time"`
	Success       bool          `json:"success"`
	Error         string        `json:"error,omitempty"`
	MemoryUsed    int64         `json:"memory_used"`
	Timestamp     time.Time     `json:"timestamp"`
}

// AggregatedMetrics holds aggregated statistics for a tool.
type AggregatedMetrics struct {
	ToolName        string
	ExecutionCount  int
	SuccessCount    int
	FailureCount    int
	AvgExecutionTime time.Duration
	MaxExecutionTime time.Duration
	MinExecutionTime time.Duration
	TotalMemoryUsed  int64
	AvgMemoryUsed    int64
}

// Collector tracks and stores metrics for tools.
type Collector struct {
	mu      sync.RWMutex
	metrics []Metric
	limits  CollectorConfig
}

// CollectorConfig defines the configuration for metrics collection.
type CollectorConfig struct {
	MaxMetrics int
}

// NewCollector creates a new metrics collector.
func NewCollector(cfg CollectorConfig) *Collector {
	if cfg.MaxMetrics <= 0 {
		cfg.MaxMetrics = 10000
	}
	return &Collector{
		metrics: make([]Metric, 0, cfg.MaxMetrics),
		limits:  cfg,
	}
}

// RecordMetric records a new tool execution metric.
func (c *Collector) RecordMetric(m Metric) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// If at capacity, remove oldest metric
	if len(c.metrics) >= c.limits.MaxMetrics {
		c.metrics = c.metrics[1:]
	}

	m.Timestamp = time.Now()
	c.metrics = append(c.metrics, m)
}

// GetMetrics returns all recorded metrics.
func (c *Collector) GetMetrics() []Metric {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]Metric, len(c.metrics))
	copy(result, c.metrics)
	return result
}

// GetAggregated returns aggregated metrics for a specific tool.
func (c *Collector) GetAggregated(toolName string) *AggregatedMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	agg := &AggregatedMetrics{
		ToolName:         toolName,
		MinExecutionTime: time.Duration(1<<63 - 1), // Max int64
	}

	var totalTime time.Duration
	for _, m := range c.metrics {
		if m.ToolName != toolName {
			continue
		}

		agg.ExecutionCount++
		agg.TotalMemoryUsed += m.MemoryUsed

		if m.Success {
			agg.SuccessCount++
		} else {
			agg.FailureCount++
		}

		totalTime += m.ExecutionTime

		if m.ExecutionTime > agg.MaxExecutionTime {
			agg.MaxExecutionTime = m.ExecutionTime
		}
		if m.ExecutionTime < agg.MinExecutionTime {
			agg.MinExecutionTime = m.ExecutionTime
		}
	}

	if agg.ExecutionCount > 0 {
		agg.AvgExecutionTime = totalTime / time.Duration(agg.ExecutionCount)
		agg.AvgMemoryUsed = agg.TotalMemoryUsed / int64(agg.ExecutionCount)
	}

	if agg.MinExecutionTime == time.Duration(1<<63-1) {
		agg.MinExecutionTime = 0
	}

	return agg
}

// GetAllAggregated returns aggregated metrics for all tools.
func (c *Collector) GetAllAggregated() map[string]*AggregatedMetrics {
	c.mu.RLock()
	defer c.mu.RUnlock()

	toolNames := make(map[string]bool)
	for _, m := range c.metrics {
		toolNames[m.ToolName] = true
	}

	result := make(map[string]*AggregatedMetrics)
	for toolName := range toolNames {
		result[toolName] = c.GetAggregated(toolName)
	}

	return result
}

// Clear removes all recorded metrics.
func (c *Collector) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.metrics = c.metrics[:0]
}

// GetSuccessRate returns the success rate as a percentage (0-100).
func (c *Collector) GetSuccessRate(toolName string) float64 {
	agg := c.GetAggregated(toolName)
	if agg.ExecutionCount == 0 {
		return 0
	}
	return float64(agg.SuccessCount) / float64(agg.ExecutionCount) * 100
}
