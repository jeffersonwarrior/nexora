// Package delegation provides resource-aware agent delegation with pooling.
package delegation

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/nexora/nexora/internal/resources"
)

// PoolConfig configures the delegate pool behavior.
type PoolConfig struct {
	// MaxConcurrent is the maximum number of concurrent delegate agents.
	// If 0, calculated dynamically based on resources.
	MaxConcurrent int

	// QueueTimeout is how long a task can wait in queue before failing.
	// Default: 30 minutes.
	QueueTimeout time.Duration

	// CPUPerAgent is the expected CPU percentage per agent.
	// Used for dynamic sizing. Default: 15%.
	CPUPerAgent float64

	// MemPerAgent is the expected memory per agent in bytes.
	// Used for dynamic sizing. Default: 512MB.
	MemPerAgent uint64

	// MinFreeMemory is the minimum free memory to maintain.
	// Default: 1GB.
	MinFreeMemory uint64

	// MinFreeCPU is the minimum free CPU percentage to maintain.
	// Default: 20%.
	MinFreeCPU float64
}

// DefaultPoolConfig returns sensible defaults for the pool.
func DefaultPoolConfig() PoolConfig {
	return PoolConfig{
		MaxConcurrent: 0, // Dynamic
		QueueTimeout:  30 * time.Minute,
		CPUPerAgent:   10.0,              // Reduced from 15% - delegates are lightweight
		MemPerAgent:   256 * 1024 * 1024, // Reduced from 512MB - delegates share model
		MinFreeMemory: 512 * 1024 * 1024, // Reduced from 1GB - more permissive
		MinFreeCPU:    10.0,              // Reduced from 20% - more permissive
	}
}

// TaskStatus represents the status of a delegated task.
type TaskStatus string

const (
	TaskStatusQueued    TaskStatus = "queued"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
	TaskStatusTimeout   TaskStatus = "timeout"
)

// Task represents a delegated task.
type Task struct {
	ID          string
	Description string
	Context     string
	WorkingDir  string
	MaxTokens   int64
	Status      TaskStatus
	Result      string
	Error       error
	CreatedAt   time.Time
	StartedAt   time.Time
	CompletedAt time.Time
	ParentSession string
	done        chan struct{}
}

// Pool manages concurrent delegate agents with resource awareness.
type Pool struct {
	mu sync.RWMutex

	config  PoolConfig
	monitor *resources.Monitor

	// Active tasks
	running   map[string]*Task
	queued    []*Task
	completed map[string]*Task // Completed tasks awaiting retrieval

	// Metrics
	totalSpawned   int
	totalCompleted int
	totalFailed    int
	totalTimeout   int

	// Lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// Task execution function (set by coordinator)
	executor func(ctx context.Context, task *Task) (string, error)
}

// NewPool creates a new delegate pool.
func NewPool(config PoolConfig, monitor *resources.Monitor) *Pool {
	if config.QueueTimeout == 0 {
		config.QueueTimeout = 30 * time.Minute
	}
	if config.CPUPerAgent == 0 {
		config.CPUPerAgent = 15.0
	}
	if config.MemPerAgent == 0 {
		config.MemPerAgent = 512 * 1024 * 1024
	}
	if config.MinFreeMemory == 0 {
		config.MinFreeMemory = 1024 * 1024 * 1024
	}
	if config.MinFreeCPU == 0 {
		config.MinFreeCPU = 20.0
	}

	return &Pool{
		config:    config,
		monitor:   monitor,
		running:   make(map[string]*Task),
		queued:    make([]*Task, 0),
		completed: make(map[string]*Task),
	}
}

// SetExecutor sets the task execution function.
func (p *Pool) SetExecutor(fn func(ctx context.Context, task *Task) (string, error)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.executor = fn
}

// Start begins the pool's queue processor.
func (p *Pool) Start(ctx context.Context) {
	p.mu.Lock()
	if p.ctx != nil {
		p.mu.Unlock()
		return
	}
	p.ctx, p.cancel = context.WithCancel(ctx)
	p.mu.Unlock()

	p.wg.Add(1)
	go p.processQueue()

	slog.Info("delegate pool started",
		"max_concurrent", p.maxConcurrent(),
		"queue_timeout", p.config.QueueTimeout,
	)
}

// Stop stops the pool and cancels all running tasks.
func (p *Pool) Stop() {
	p.mu.Lock()
	if p.cancel != nil {
		p.cancel()
	}
	p.mu.Unlock()

	p.wg.Wait()
	slog.Info("delegate pool stopped")
}

// Submit adds a task to the pool.
// Returns the task ID and a channel that closes when the task completes.
func (p *Pool) Submit(description, taskContext, workingDir string, maxTokens int64, parentSession string) (string, <-chan struct{}, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.ctx == nil {
		return "", nil, fmt.Errorf("pool not started")
	}

	task := &Task{
		ID:            uuid.New().String(),
		Description:   description,
		Context:       taskContext,
		WorkingDir:    workingDir,
		MaxTokens:     maxTokens,
		Status:        TaskStatusQueued,
		CreatedAt:     time.Now(),
		ParentSession: parentSession,
		done:          make(chan struct{}),
	}

	p.queued = append(p.queued, task)
	slog.Info("task queued", "task_id", task.ID, "queue_length", len(p.queued))

	return task.ID, task.done, nil
}

// GetTask returns a task by ID.
func (p *Pool) GetTask(id string) (*Task, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if task, ok := p.running[id]; ok {
		return task, true
	}

	if task, ok := p.completed[id]; ok {
		return task, true
	}

	for _, task := range p.queued {
		if task.ID == id {
			return task, true
		}
	}

	return nil, false
}

// Wait waits for a task to complete and returns its result.
// After Wait returns, the task is removed from the pool.
func (p *Pool) Wait(id string) (string, error) {
	task, ok := p.GetTask(id)
	if !ok {
		return "", fmt.Errorf("task not found: %s", id)
	}

	<-task.done

	p.mu.Lock()
	defer p.mu.Unlock()

	// Clean up from completed map after retrieval
	delete(p.completed, id)

	if task.Error != nil {
		return "", task.Error
	}
	return task.Result, nil
}

// Cancel cancels a queued or running task.
func (p *Pool) Cancel(id string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Check queued tasks
	for i, task := range p.queued {
		if task.ID == id {
			task.Status = TaskStatusCancelled
			task.CompletedAt = time.Now()
			close(task.done)
			p.queued = append(p.queued[:i], p.queued[i+1:]...)
			return nil
		}
	}

	// Running tasks can't be cancelled directly (context cancellation)
	if _, ok := p.running[id]; ok {
		return fmt.Errorf("cannot cancel running task, use pool.Stop() to stop all")
	}

	return fmt.Errorf("task not found: %s", id)
}

// Stats returns pool statistics.
func (p *Pool) Stats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return PoolStats{
		Running:        len(p.running),
		Queued:         len(p.queued),
		MaxConcurrent:  p.maxConcurrent(),
		TotalSpawned:   p.totalSpawned,
		TotalCompleted: p.totalCompleted,
		TotalFailed:    p.totalFailed,
		TotalTimeout:   p.totalTimeout,
	}
}

// PoolStats contains pool statistics.
type PoolStats struct {
	Running        int
	Queued         int
	MaxConcurrent  int
	TotalSpawned   int
	TotalCompleted int
	TotalFailed    int
	TotalTimeout   int
}

// maxConcurrent returns the maximum concurrent agents.
func (p *Pool) maxConcurrent() int {
	if p.config.MaxConcurrent > 0 {
		return p.config.MaxConcurrent
	}

	// Dynamic calculation based on resources
	if p.monitor == nil {
		return 3 // Fallback
	}

	snapshot := p.monitor.CurrentSnapshot()

	// Calculate based on CPU
	availableCPU := 100.0 - snapshot.CPUPercent - p.config.MinFreeCPU
	cpuAgents := int(availableCPU / p.config.CPUPerAgent)

	// Calculate based on memory (available = total - used)
	memAvailable := snapshot.MemTotal - snapshot.MemUsed
	if memAvailable < p.config.MinFreeMemory {
		memAvailable = 0
	} else {
		memAvailable -= p.config.MinFreeMemory
	}
	memAgents := int(memAvailable / p.config.MemPerAgent)

	// Use minimum of both
	max := min(cpuAgents, memAgents)
	if max < 1 {
		max = 1 // Always allow at least 1
	}
	if max > 10 {
		max = 10 // Cap at 10
	}

	return max
}

// canSpawn checks if resources allow spawning another agent.
func (p *Pool) canSpawn() bool {
	if len(p.running) >= p.maxConcurrent() {
		return false
	}

	if p.monitor == nil {
		return true
	}

	snapshot := p.monitor.CurrentSnapshot()

	// Check CPU headroom
	if snapshot.CPUPercent > (100.0 - p.config.MinFreeCPU - p.config.CPUPerAgent) {
		slog.Debug("cannot spawn: CPU constraint",
			"current", snapshot.CPUPercent,
			"threshold", 100.0-p.config.MinFreeCPU-p.config.CPUPerAgent,
		)
		return false
	}

	// Check memory headroom (available = total - used)
	memAvailable := snapshot.MemTotal - snapshot.MemUsed
	if memAvailable < p.config.MinFreeMemory+p.config.MemPerAgent {
		slog.Debug("cannot spawn: memory constraint",
			"available", memAvailable,
			"required", p.config.MinFreeMemory+p.config.MemPerAgent,
		)
		return false
	}

	return true
}

// processQueue is the main queue processing loop.
func (p *Pool) processQueue() {
	defer p.wg.Done()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-p.ctx.Done():
			p.drainQueue()
			return
		case <-ticker.C:
			p.checkTimeouts()
			p.trySpawn()
		}
	}
}

// trySpawn attempts to spawn tasks from the queue.
func (p *Pool) trySpawn() {
	p.mu.Lock()

	if len(p.queued) == 0 || !p.canSpawn() {
		p.mu.Unlock()
		return
	}

	// Pop first task from queue
	task := p.queued[0]
	p.queued = p.queued[1:]

	task.Status = TaskStatusRunning
	task.StartedAt = time.Now()
	p.running[task.ID] = task
	p.totalSpawned++

	executor := p.executor
	p.mu.Unlock()

	if executor == nil {
		p.mu.Lock()
		task.Status = TaskStatusFailed
		task.Error = fmt.Errorf("no executor configured")
		task.CompletedAt = time.Now()
		delete(p.running, task.ID)
		p.totalFailed++
		close(task.done)
		p.mu.Unlock()
		return
	}

	// Execute task in goroutine
	go func() {
		slog.Info("starting delegated task", "task_id", task.ID)

		result, err := executor(p.ctx, task)

		p.mu.Lock()
		defer p.mu.Unlock()

		task.CompletedAt = time.Now()
		delete(p.running, task.ID)

		if err != nil {
			task.Status = TaskStatusFailed
			task.Error = err
			p.totalFailed++
			slog.Warn("delegated task failed", "task_id", task.ID, "error", err)
		} else {
			task.Status = TaskStatusCompleted
			task.Result = result
			p.totalCompleted++
			slog.Info("delegated task completed", "task_id", task.ID)
		}

		// Move to completed map so Wait() can retrieve results
		p.completed[task.ID] = task
		close(task.done)
	}()
}

// checkTimeouts checks for queued tasks that have exceeded the timeout.
func (p *Pool) checkTimeouts() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	remaining := make([]*Task, 0, len(p.queued))

	for _, task := range p.queued {
		if now.Sub(task.CreatedAt) > p.config.QueueTimeout {
			task.Status = TaskStatusTimeout
			task.Error = fmt.Errorf("queue timeout after %v", p.config.QueueTimeout)
			task.CompletedAt = now
			p.totalTimeout++
			close(task.done)
			slog.Warn("task timed out in queue", "task_id", task.ID)
		} else {
			remaining = append(remaining, task)
		}
	}

	p.queued = remaining
}

// drainQueue cancels all queued tasks on shutdown.
func (p *Pool) drainQueue() {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, task := range p.queued {
		task.Status = TaskStatusCancelled
		task.CompletedAt = time.Now()
		close(task.done)
	}
	p.queued = nil

	slog.Info("queue drained on shutdown", "cancelled", len(p.queued))
}
