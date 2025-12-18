package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCollectorRecordAndRetrieve(t *testing.T) {
	t.Parallel()
	c := NewCollector(CollectorConfig{MaxMetrics: 100})

	m := Metric{
		ToolName:      "edit",
		ExecutionTime: 100 * time.Millisecond,
		Success:       true,
		MemoryUsed:    1024,
	}

	c.RecordMetric(m)

	metrics := c.GetMetrics()
	require.Len(t, metrics, 1)
	require.Equal(t, "edit", metrics[0].ToolName)
	require.True(t, metrics[0].Success)
}

func TestAggregatedMetrics(t *testing.T) {
	t.Parallel()
	c := NewCollector(CollectorConfig{MaxMetrics: 100})

	// Record success metrics
	for i := 0; i < 3; i++ {
		c.RecordMetric(Metric{
			ToolName:      "grep",
			ExecutionTime: 100 * time.Millisecond,
			Success:       true,
			MemoryUsed:    2048,
		})
	}

	// Record failure metrics
	for i := 0; i < 2; i++ {
		c.RecordMetric(Metric{
			ToolName:      "grep",
			ExecutionTime: 50 * time.Millisecond,
			Success:       false,
			Error:         "test error",
			MemoryUsed:    1024,
		})
	}

	agg := c.GetAggregated("grep")
	require.Equal(t, 5, agg.ExecutionCount)
	require.Equal(t, 3, agg.SuccessCount)
	require.Equal(t, 2, agg.FailureCount)
	require.Equal(t, int64(3*2048+2*1024), agg.TotalMemoryUsed)
	require.Equal(t, 60.0, c.GetSuccessRate("grep"))
}

func TestMaxMetricsCapacity(t *testing.T) {
	t.Parallel()
	c := NewCollector(CollectorConfig{MaxMetrics: 5})

	for i := 0; i < 10; i++ {
		c.RecordMetric(Metric{
			ToolName:      "test",
			ExecutionTime: time.Duration(i) * time.Millisecond,
			Success:       true,
			MemoryUsed:    int64(i),
		})
	}

	metrics := c.GetMetrics()
	require.Len(t, metrics, 5)
	// Oldest metrics should be removed
	require.Equal(t, int64(5), metrics[0].MemoryUsed)
	require.Equal(t, int64(9), metrics[4].MemoryUsed)
}

func TestGetAllAggregated(t *testing.T) {
	t.Parallel()
	c := NewCollector(CollectorConfig{MaxMetrics: 100})

	c.RecordMetric(Metric{
		ToolName:      "edit",
		ExecutionTime: 100 * time.Millisecond,
		Success:       true,
		MemoryUsed:    1024,
	})

	c.RecordMetric(Metric{
		ToolName:      "grep",
		ExecutionTime: 50 * time.Millisecond,
		Success:       true,
		MemoryUsed:    512,
	})

	agg := c.GetAllAggregated()
	require.Len(t, agg, 2)
	require.NotNil(t, agg["edit"])
	require.NotNil(t, agg["grep"])
	require.Equal(t, 1, agg["edit"].ExecutionCount)
	require.Equal(t, 1, agg["grep"].ExecutionCount)
}

func TestClear(t *testing.T) {
	t.Parallel()
	c := NewCollector(CollectorConfig{MaxMetrics: 100})

	c.RecordMetric(Metric{
		ToolName:      "test",
		ExecutionTime: 100 * time.Millisecond,
		Success:       true,
		MemoryUsed:    1024,
	})

	require.Len(t, c.GetMetrics(), 1)

	c.Clear()
	require.Len(t, c.GetMetrics(), 0)
}

func TestSuccessRate(t *testing.T) {
	t.Parallel()
	c := NewCollector(CollectorConfig{MaxMetrics: 100})

	// No metrics
	require.Equal(t, 0.0, c.GetSuccessRate("nonexistent"))

	// All success
	for i := 0; i < 4; i++ {
		c.RecordMetric(Metric{
			ToolName:      "tool1",
			ExecutionTime: 100 * time.Millisecond,
			Success:       true,
			MemoryUsed:    1024,
		})
	}

	// All failure
	for i := 0; i < 6; i++ {
		c.RecordMetric(Metric{
			ToolName:      "tool2",
			ExecutionTime: 100 * time.Millisecond,
			Success:       false,
			MemoryUsed:    1024,
		})
	}

	require.Equal(t, 100.0, c.GetSuccessRate("tool1"))
	require.Equal(t, 0.0, c.GetSuccessRate("tool2"))
}
