package metrics

import (
	"context"
	"runtime"
	"time"
)

// Hook provides a hook interface for recording metrics during tool execution.
type Hook struct {
	collector *Collector
}

// NewHook creates a new metrics hook with the given collector.
func NewHook(collector *Collector) *Hook {
	return &Hook{collector: collector}
}

// StartTimer returns a function that should be called when a tool execution completes.
// It captures the start memory and time for later recording.
func (h *Hook) StartTimer() func(toolName string, success bool, err error) {
	startTime := time.Now()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	startMemory := m.Alloc

	return func(toolName string, success bool, err error) {
		endTime := time.Now()

		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		endMemory := m.Alloc

		errMsg := ""
		if err != nil {
			errMsg = err.Error()
		}

		h.collector.RecordMetric(Metric{
			ToolName:      toolName,
			ExecutionTime: endTime.Sub(startTime),
			Success:       success,
			Error:         errMsg,
			MemoryUsed:    int64(endMemory - startMemory),
		})
	}
}

// StartTimerWithContext is like StartTimer but takes a context for additional flexibility.
func (h *Hook) StartTimerWithContext(ctx context.Context) func(toolName string, success bool, err error) {
	return h.StartTimer()
}

// GetCollector returns the underlying metrics collector.
func (h *Hook) GetCollector() *Collector {
	return h.collector
}
