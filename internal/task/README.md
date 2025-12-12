# Task Management System

This package provides a task management system to keep AI agents focused and prevent drift during development tasks.

## Features

- **Task Tracking**: Create and manage focused tasks with clear descriptions
- **Drift Detection**: Automatically analyzes AI responses for task drift
- **Context Injection**: Adds task context to prompts to maintain focus
- **Correction Prompts**: Generates guidance to redirect drifted AI responses
- **Completion Tracking**: Marks tasks as completed and removes from active tracking

## Core Components

### SimpleService
The main interface for task management:

```go
type SimpleService interface {
    CreateTask(ctx context.Context, sessionID, title, description string) *SimpleTask
    GetActiveTask(ctx context.Context, sessionID string) (*SimpleTask, bool)
    UpdateTask(ctx context.Context, sessionID, description string) *SimpleTask
    CompleteTask(ctx context.Context, sessionID string) bool
    CheckFocus(ctx context.Context, sessionID, response string) (bool, string)
    GetFocusPrompt(ctx context.Context, sessionID string) string
}
```

### SimpleTask
A basic task structure:

```go
type SimpleTask struct {
    ID          string    `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Status      string    `json:"status"`
    SessionID   string    `json:"session_id"`
    Created     time.Time `json:"created"`
    Updated     time.Time `json:"updated"`
}
```

## Usage

### Basic Task Management

```go
// Create task service
taskService := task.NewSimpleService()
ctx := context.Background()

// Start a new task
task := taskService.CreateTask(ctx, "session-123", "Fix authentication bug", "Update the JWT validation logic in auth middleware")

// Check if responses are focused
isFocused, reason := taskService.CheckFocus(ctx, "session-123", "Here's the fix for the JWT validation:")
if !isFocused {
    prompt := taskService.GetFocusPrompt(ctx, "session-123")
    // Use correction prompt to refocus the AI
}

// Complete the task
taskService.CompleteTask(ctx, "session-123")
```

### Integration with Agents

The task system can be easily integrated into AI agent coordinators:

1. **Context Injection**: Add task context to all prompts
2. **Response Monitoring**: Check each response for drift
3. **Auto-Correction**: Automatically apply corrections when drift is detected

## Drift Detection

The system detects drift using multiple indicators:

### Relevance Scoring
- Calculates word overlap between response and task context
- Requires minimum 10% relevance to maintain focus

### Drift Indicators
The system watches for patterns that indicate focus loss:
- "weather", "news", "sports", "celebrity" - Unrelated topics
- "could also", "while we're at it", "by the way", "also" - Scope creep
- "random", "joke", "fun fact" - Non-productive content

### Correction Strategy
When drift is detected, the system:
1. Identifies the drift type
2. Provides a refocusing prompt
3. Reinforces the current task requirements
4. Maintains a helpful but focused tone

## Demo

Run the built-in demo to see the system in action:

```bash
go run . task-demo
```

This demonstrates:
- Task creation
- Drift detection on various response types
- Automatic correction generation
- Task completion

## Configuration

The system is designed to be simple and require minimal configuration:

- **No external dependencies**: Pure Go implementation
- **Memory-based storage**: Tasks stored in map (can be extended to use database)
- **Context-driven**: Each session maintains independent task state

## Extensions

The basic system can be extended with:

- **Persistence**: Database backing for tasks
- **Milestones**: Breaking tasks into trackable milestones
- **Priority System**: Task prioritization across multiple tasks
- **Analytics**: Task completion metrics and drift patterns
- **Multi-agent Support**: Task coordination across multiple AI agents

## Integration Points

### Agent Integration
```go
// Wrap existing coordinator
taskCoordinator := task.NewCoordinatorMiddleware(baseCoordinator, taskService)

// Enhanced prompts with task context
result := taskCoordinator.Run(ctx, sessionID, userInput)
```

### Tool Integration
```go
// Expose task management as tools to the AI
tools := map[string]interface{}{
    "start_task":    StartTaskFunction,
    "get_status":    GetStatusFunction,
    "update_progress": UpdateProgressFunction,
    "complete_task": CompleteTaskFunction,
    "check_drift":   CheckDriftFunction,
}
```

The task management system provides a solid foundation for keeping AI agents focused and on-task during complex development work.