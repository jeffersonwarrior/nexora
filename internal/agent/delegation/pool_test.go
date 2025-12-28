package delegation

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/nexora/nexora/internal/resources"
)

func TestDefaultPoolConfig(t *testing.T) {
	config := DefaultPoolConfig()

	if config.MaxConcurrent != 0 {
		t.Errorf("expected MaxConcurrent=0 for dynamic sizing, got %d", config.MaxConcurrent)
	}
	if config.QueueTimeout != 30*time.Minute {
		t.Errorf("expected QueueTimeout=30min, got %v", config.QueueTimeout)
	}
	if config.CPUPerAgent != 15.0 {
		t.Errorf("expected CPUPerAgent=15.0, got %f", config.CPUPerAgent)
	}
	if config.MemPerAgent != 512*1024*1024 {
		t.Errorf("expected MemPerAgent=512MB, got %d", config.MemPerAgent)
	}
	if config.MinFreeMemory != 1024*1024*1024 {
		t.Errorf("expected MinFreeMemory=1GB, got %d", config.MinFreeMemory)
	}
	if config.MinFreeCPU != 20.0 {
		t.Errorf("expected MinFreeCPU=20.0, got %f", config.MinFreeCPU)
	}
}

func TestNewPool(t *testing.T) {
	config := DefaultPoolConfig()
	pool := NewPool(config, nil)

	if pool == nil {
		t.Fatal("NewPool returned nil")
	}

	if pool.config.MaxConcurrent != config.MaxConcurrent {
		t.Errorf("config not set correctly")
	}
}

func TestNewPool_WithZeroDefaults(t *testing.T) {
	config := PoolConfig{} // All zeros
	pool := NewPool(config, nil)

	// Should apply defaults
	if pool.config.QueueTimeout == 0 {
		t.Error("expected QueueTimeout to have default")
	}
	if pool.config.CPUPerAgent == 0 {
		t.Error("expected CPUPerAgent to have default")
	}
	if pool.config.MemPerAgent == 0 {
		t.Error("expected MemPerAgent to have default")
	}
	if pool.config.MinFreeMemory == 0 {
		t.Error("expected MinFreeMemory to have default")
	}
	if pool.config.MinFreeCPU == 0 {
		t.Error("expected MinFreeCPU to have default")
	}
}

func TestPool_SetExecutor(t *testing.T) {
	pool := NewPool(DefaultPoolConfig(), nil)

	called := false
	pool.SetExecutor(func(ctx context.Context, task *Task) (string, error) {
		called = true
		return "result", nil
	})

	if pool.executor == nil {
		t.Error("executor not set")
	}

	// Verify executor works
	task := &Task{}
	result, err := pool.executor(context.Background(), task)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if result != "result" {
		t.Errorf("expected result='result', got %q", result)
	}
	if !called {
		t.Error("executor not called")
	}
}

func TestPool_StartStop(t *testing.T) {
	pool := NewPool(DefaultPoolConfig(), nil)
	ctx := context.Background()

	pool.Start(ctx)

	// Verify pool is running
	if pool.ctx == nil {
		t.Error("pool context not set after Start")
	}

	pool.Stop()

	// Verify pool is stopped
	select {
	case <-pool.ctx.Done():
		// Expected
	default:
		t.Error("pool context not cancelled after Stop")
	}
}

func TestPool_SubmitBeforeStart(t *testing.T) {
	pool := NewPool(DefaultPoolConfig(), nil)

	_, _, err := pool.Submit("test task", "context", "/tmp", 1000, "session-1")
	if err == nil {
		t.Error("expected error when submitting before pool start")
	}
	if err.Error() != "pool not started" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestPool_Submit(t *testing.T) {
	pool := NewPool(DefaultPoolConfig(), nil)
	ctx := context.Background()
	pool.Start(ctx)
	defer pool.Stop()

	id, done, err := pool.Submit("test task", "context", "/tmp", 1000, "session-1")

	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	if id == "" {
		t.Error("expected non-empty task ID")
	}

	if done == nil {
		t.Error("expected done channel")
	}

	// Verify task is queued
	task, ok := pool.GetTask(id)
	if !ok {
		t.Error("task not found after submission")
	}

	if task.Status != TaskStatusQueued {
		t.Errorf("expected status=queued, got %s", task.Status)
	}

	if task.Description != "test task" {
		t.Errorf("expected description='test task', got %q", task.Description)
	}
}

func TestPool_GetTask(t *testing.T) {
	pool := NewPool(DefaultPoolConfig(), nil)
	ctx := context.Background()
	pool.Start(ctx)
	defer pool.Stop()

	// Task not found
	_, ok := pool.GetTask("nonexistent")
	if ok {
		t.Error("expected task not found")
	}

	// Submit and find task
	id, _, _ := pool.Submit("test", "ctx", "/tmp", 1000, "sess")
	task, ok := pool.GetTask(id)
	if !ok {
		t.Error("expected to find task")
	}
	if task.ID != id {
		t.Errorf("expected task ID %s, got %s", id, task.ID)
	}
}

func TestPool_Cancel(t *testing.T) {
	pool := NewPool(DefaultPoolConfig(), nil)
	ctx := context.Background()
	pool.Start(ctx)
	defer pool.Stop()

	// Cancel nonexistent task
	err := pool.Cancel("nonexistent")
	if err == nil {
		t.Error("expected error when cancelling nonexistent task")
	}

	// Submit and cancel task
	id, done, _ := pool.Submit("test", "ctx", "/tmp", 1000, "sess")

	// Get task before cancel
	task, ok := pool.GetTask(id)
	if !ok {
		t.Fatal("task not found before cancel")
	}

	err = pool.Cancel(id)
	if err != nil {
		t.Errorf("Cancel failed: %v", err)
	}

	// Verify done channel is closed
	select {
	case <-done:
		// Expected
	case <-time.After(100 * time.Millisecond):
		t.Error("done channel not closed after cancel")
	}

	// Verify task status was set to cancelled
	if task.Status != TaskStatusCancelled {
		t.Errorf("expected status=cancelled, got %s", task.Status)
	}
}

func TestPool_Stats(t *testing.T) {
	pool := NewPool(DefaultPoolConfig(), nil)

	stats := pool.Stats()

	if stats.Running != 0 {
		t.Errorf("expected Running=0, got %d", stats.Running)
	}
	if stats.Queued != 0 {
		t.Errorf("expected Queued=0, got %d", stats.Queued)
	}
	if stats.MaxConcurrent == 0 {
		t.Error("expected MaxConcurrent > 0")
	}
}

func TestPool_MaxConcurrent_Static(t *testing.T) {
	config := DefaultPoolConfig()
	config.MaxConcurrent = 5
	pool := NewPool(config, nil)

	max := pool.maxConcurrent()
	if max != 5 {
		t.Errorf("expected maxConcurrent=5, got %d", max)
	}
}

func TestPool_MaxConcurrent_Dynamic(t *testing.T) {
	config := DefaultPoolConfig()
	config.MaxConcurrent = 0 // Dynamic

	// Test without monitor (fallback)
	pool := NewPool(config, nil)
	max := pool.maxConcurrent()
	if max != 3 {
		t.Errorf("expected fallback maxConcurrent=3, got %d", max)
	}

	// Test with monitor
	monitor := resources.NewMonitor(resources.Config{CheckInterval: 100 * time.Millisecond})
	pool = NewPool(config, monitor)
	monitor.Start(context.Background())
	defer monitor.Stop()

	time.Sleep(50 * time.Millisecond) // Let monitor collect data

	max = pool.maxConcurrent()
	if max < 1 || max > 10 {
		t.Errorf("expected maxConcurrent in range [1, 10], got %d", max)
	}
}

func TestTaskStatus_Constants(t *testing.T) {
	tests := []struct {
		status   TaskStatus
		expected string
	}{
		{TaskStatusQueued, "queued"},
		{TaskStatusRunning, "running"},
		{TaskStatusCompleted, "completed"},
		{TaskStatusFailed, "failed"},
		{TaskStatusCancelled, "cancelled"},
		{TaskStatusTimeout, "timeout"},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			if string(tt.status) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(tt.status))
			}
		})
	}
}

func TestPool_CanSpawn_MaxConcurrent(t *testing.T) {
	config := DefaultPoolConfig()
	config.MaxConcurrent = 2
	pool := NewPool(config, nil)

	pool.mu.Lock()
	pool.running = make(map[string]*Task)
	pool.running["task1"] = &Task{ID: "task1"}
	pool.running["task2"] = &Task{ID: "task2"}
	pool.mu.Unlock()

	// At max, cannot spawn
	if pool.canSpawn() {
		t.Error("expected canSpawn=false when at max concurrent")
	}

	// Remove one, can spawn
	pool.mu.Lock()
	delete(pool.running, "task2")
	pool.mu.Unlock()

	if !pool.canSpawn() {
		t.Error("expected canSpawn=true when below max concurrent")
	}
}

func TestPoolStats_Fields(t *testing.T) {
	stats := PoolStats{
		Running:        1,
		Queued:         2,
		MaxConcurrent:  5,
		TotalSpawned:   10,
		TotalCompleted: 7,
		TotalFailed:    2,
		TotalTimeout:   1,
	}

	if stats.Running != 1 {
		t.Errorf("expected Running=1, got %d", stats.Running)
	}
	if stats.Queued != 2 {
		t.Errorf("expected Queued=2, got %d", stats.Queued)
	}
	if stats.MaxConcurrent != 5 {
		t.Errorf("expected MaxConcurrent=5, got %d", stats.MaxConcurrent)
	}
	if stats.TotalSpawned != 10 {
		t.Errorf("expected TotalSpawned=10, got %d", stats.TotalSpawned)
	}
	if stats.TotalCompleted != 7 {
		t.Errorf("expected TotalCompleted=7, got %d", stats.TotalCompleted)
	}
	if stats.TotalFailed != 2 {
		t.Errorf("expected TotalFailed=2, got %d", stats.TotalFailed)
	}
	if stats.TotalTimeout != 1 {
		t.Errorf("expected TotalTimeout=1, got %d", stats.TotalTimeout)
	}
}

func TestTask_Fields(t *testing.T) {
	now := time.Now()
	task := Task{
		ID:            "test-123",
		Description:   "test task",
		Context:       "test context",
		WorkingDir:    "/tmp",
		MaxTokens:     1000,
		Status:        TaskStatusQueued,
		Result:        "",
		Error:         nil,
		CreatedAt:     now,
		StartedAt:     time.Time{},
		CompletedAt:   time.Time{},
		ParentSession: "session-1",
	}

	if task.ID != "test-123" {
		t.Errorf("expected ID='test-123', got %q", task.ID)
	}
	if task.Description != "test task" {
		t.Errorf("expected Description='test task', got %q", task.Description)
	}
	if task.Context != "test context" {
		t.Errorf("expected Context='test context', got %q", task.Context)
	}
	if task.WorkingDir != "/tmp" {
		t.Errorf("expected WorkingDir='/tmp', got %q", task.WorkingDir)
	}
	if task.MaxTokens != 1000 {
		t.Errorf("expected MaxTokens=1000, got %d", task.MaxTokens)
	}
	if task.Status != TaskStatusQueued {
		t.Errorf("expected Status=queued, got %s", task.Status)
	}
	if task.ParentSession != "session-1" {
		t.Errorf("expected ParentSession='session-1', got %q", task.ParentSession)
	}
	if !task.CreatedAt.Equal(now) {
		t.Error("CreatedAt not set correctly")
	}
}

// TestPool_ExecuteAndWait tests the full delegate workflow:
// 1. Submit task
// 2. Executor runs and returns result
// 3. Wait() retrieves result successfully
func TestPool_ExecuteAndWait(t *testing.T) {
	config := DefaultPoolConfig()
	config.MaxConcurrent = 2
	pool := NewPool(config, nil)

	// Set up executor that returns a result
	executorCalled := make(chan struct{})
	pool.SetExecutor(func(ctx context.Context, task *Task) (string, error) {
		close(executorCalled)
		return "task completed successfully", nil
	})

	ctx := context.Background()
	pool.Start(ctx)
	defer pool.Stop()

	// Submit task
	id, done, err := pool.Submit("test task", "context", "/tmp", 1000, "session-1")
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	// Wait for task to complete via done channel
	select {
	case <-done:
		// Expected
	case <-time.After(5 * time.Second):
		t.Fatal("task did not complete in time")
	}

	// Verify executor was called
	select {
	case <-executorCalled:
		// Expected
	default:
		t.Error("executor was not called")
	}

	// Wait should return the result
	result, err := pool.Wait(id)
	if err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	if result != "task completed successfully" {
		t.Errorf("expected result='task completed successfully', got %q", result)
	}

	// After Wait, task should be cleaned up from completed map
	pool.mu.RLock()
	_, inCompleted := pool.completed[id]
	pool.mu.RUnlock()
	if inCompleted {
		t.Error("task should be removed from completed map after Wait")
	}
}

// TestPool_ExecuteWithError tests error handling in delegate workflow
func TestPool_ExecuteWithError(t *testing.T) {
	config := DefaultPoolConfig()
	config.MaxConcurrent = 2
	pool := NewPool(config, nil)

	expectedErr := fmt.Errorf("executor failed")
	pool.SetExecutor(func(ctx context.Context, task *Task) (string, error) {
		return "", expectedErr
	})

	ctx := context.Background()
	pool.Start(ctx)
	defer pool.Stop()

	id, done, err := pool.Submit("test task", "context", "/tmp", 1000, "session-1")
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	// Wait for completion
	<-done

	// Wait should return the error
	_, err = pool.Wait(id)
	if err == nil {
		t.Fatal("expected error from Wait")
	}

	if err.Error() != expectedErr.Error() {
		t.Errorf("expected error %q, got %q", expectedErr.Error(), err.Error())
	}
}

// TestPool_CompletedTasksAccessible tests that completed tasks remain
// accessible until Wait() is called (the bug that was fixed)
func TestPool_CompletedTasksAccessible(t *testing.T) {
	config := DefaultPoolConfig()
	config.MaxConcurrent = 2
	pool := NewPool(config, nil)

	pool.SetExecutor(func(ctx context.Context, task *Task) (string, error) {
		return "done", nil
	})

	ctx := context.Background()
	pool.Start(ctx)
	defer pool.Stop()

	id, done, err := pool.Submit("test", "ctx", "/tmp", 1000, "sess")
	if err != nil {
		t.Fatalf("Submit failed: %v", err)
	}

	// Wait for task to complete
	<-done

	// Task should still be accessible via GetTask after completion
	task, ok := pool.GetTask(id)
	if !ok {
		t.Fatal("completed task should be accessible via GetTask before Wait()")
	}

	if task.Status != TaskStatusCompleted {
		t.Errorf("expected status=completed, got %s", task.Status)
	}

	if task.Result != "done" {
		t.Errorf("expected result='done', got %q", task.Result)
	}

	// Now call Wait - should succeed
	result, err := pool.Wait(id)
	if err != nil {
		t.Fatalf("Wait failed: %v", err)
	}

	if result != "done" {
		t.Errorf("expected result='done', got %q", result)
	}

	// After Wait, GetTask should return false
	_, ok = pool.GetTask(id)
	if ok {
		t.Error("task should not be accessible after Wait()")
	}
}

// TestPool_MultipleTasksSequential tests multiple tasks processed sequentially
func TestPool_MultipleTasksSequential(t *testing.T) {
	config := DefaultPoolConfig()
	config.MaxConcurrent = 1 // Force sequential execution
	pool := NewPool(config, nil)

	var executionOrder []string
	var mu sync.Mutex

	pool.SetExecutor(func(ctx context.Context, task *Task) (string, error) {
		mu.Lock()
		executionOrder = append(executionOrder, task.Description)
		mu.Unlock()
		return task.Description + " result", nil
	})

	ctx := context.Background()
	pool.Start(ctx)
	defer pool.Stop()

	// Submit multiple tasks
	id1, done1, _ := pool.Submit("task1", "", "/tmp", 1000, "sess")
	id2, done2, _ := pool.Submit("task2", "", "/tmp", 1000, "sess")
	id3, done3, _ := pool.Submit("task3", "", "/tmp", 1000, "sess")

	// Wait for all to complete
	<-done1
	<-done2
	<-done3

	// All results should be retrievable
	result1, err := pool.Wait(id1)
	if err != nil {
		t.Errorf("Wait(id1) failed: %v", err)
	}
	if result1 != "task1 result" {
		t.Errorf("expected 'task1 result', got %q", result1)
	}

	result2, err := pool.Wait(id2)
	if err != nil {
		t.Errorf("Wait(id2) failed: %v", err)
	}
	if result2 != "task2 result" {
		t.Errorf("expected 'task2 result', got %q", result2)
	}

	result3, err := pool.Wait(id3)
	if err != nil {
		t.Errorf("Wait(id3) failed: %v", err)
	}
	if result3 != "task3 result" {
		t.Errorf("expected 'task3 result', got %q", result3)
	}

	// Verify execution order (should be sequential with MaxConcurrent=1)
	mu.Lock()
	if len(executionOrder) != 3 {
		t.Errorf("expected 3 executions, got %d", len(executionOrder))
	}
	mu.Unlock()
}

// TestPool_ParallelExecution tests multiple tasks running in parallel
func TestPool_ParallelExecution(t *testing.T) {
	config := DefaultPoolConfig()
	config.MaxConcurrent = 3
	pool := NewPool(config, nil)

	startedCount := int32(0)
	releaseExecution := make(chan struct{})

	pool.SetExecutor(func(ctx context.Context, task *Task) (string, error) {
		atomic.AddInt32(&startedCount, 1)
		<-releaseExecution // Wait for signal to complete
		return "done", nil
	})

	ctx := context.Background()
	pool.Start(ctx)
	defer pool.Stop()

	// Submit 3 tasks
	id1, done1, _ := pool.Submit("task1", "", "/tmp", 1000, "sess")
	id2, done2, _ := pool.Submit("task2", "", "/tmp", 1000, "sess")
	id3, done3, _ := pool.Submit("task3", "", "/tmp", 1000, "sess")

	// Give time for all tasks to start (poll until all started or timeout)
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&startedCount) >= 3 {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}

	// All 3 should be running in parallel
	count := atomic.LoadInt32(&startedCount)
	if count < 3 {
		t.Errorf("expected 3 tasks running in parallel, got %d", count)
	}

	// Release all tasks
	close(releaseExecution)

	// Wait for all to complete
	<-done1
	<-done2
	<-done3

	// All results should be retrievable
	for _, id := range []string{id1, id2, id3} {
		result, err := pool.Wait(id)
		if err != nil {
			t.Errorf("Wait(%s) failed: %v", id, err)
		}
		if result != "done" {
			t.Errorf("expected 'done', got %q", result)
		}
	}
}
