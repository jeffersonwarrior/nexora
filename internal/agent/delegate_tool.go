package agent

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"

	"charm.land/fantasy"

	"github.com/nexora/nexora/internal/agent/delegation"
	"github.com/nexora/nexora/internal/agent/prompt"
	"github.com/nexora/nexora/internal/agent/tools"
	"github.com/nexora/nexora/internal/permission"
)

//go:embed templates/delegate.md
var delegateToolDescription []byte

//go:embed templates/delegate_prompt.md.tpl
var delegatePromptTmpl []byte

const DelegateToolName = "delegate"

// DelegateParams represents the parameters for the delegate tool
type DelegateParams struct {
	Task       string `json:"task" description:"The task to delegate to a sub-agent. Be specific and include all necessary context."`
	Context    string `json:"context,omitempty" description:"Additional context or background information for the sub-agent."`
	WorkingDir string `json:"working_dir,omitempty" description:"Working directory for the sub-agent (defaults to current project root)."`
	MaxTokens  int    `json:"max_tokens,omitempty" description:"Maximum tokens for the sub-agent response (default: 4096)."`
	Background bool   `json:"background,omitempty" description:"Run the delegated task in the background and return immediately with a task ID."`
}

// DelegatePermissionsParams is the permission-safe version of DelegateParams
type DelegatePermissionsParams struct {
	Task       string `json:"task"`
	Context    string `json:"context"`
	WorkingDir string `json:"working_dir"`
	Background bool   `json:"background"`
}

// delegateValidationResult holds validated parameters from tool call context
type delegateValidationResult struct {
	SessionID      string
	AgentMessageID string
}

// validateDelegateParams validates the tool call parameters
func validateDelegateParams(ctx context.Context, params DelegateParams) (delegateValidationResult, error) {
	if params.Task == "" {
		return delegateValidationResult{}, fmt.Errorf("task description is required")
	}

	sessionID := tools.GetSessionFromContext(ctx)
	if sessionID == "" {
		return delegateValidationResult{}, fmt.Errorf("session id missing from context")
	}

	agentMessageID := tools.GetMessageFromContext(ctx)
	if agentMessageID == "" {
		return delegateValidationResult{}, fmt.Errorf("agent message id missing from context")
	}

	return delegateValidationResult{
		SessionID:      sessionID,
		AgentMessageID: agentMessageID,
	}, nil
}

// delegateTool creates a tool that can spawn sub-agents for parallel task execution.
// It uses the delegation pool for resource-aware spawning with queue timeout.
func (c *coordinator) delegateTool(ctx context.Context) (fantasy.AgentTool, error) {
	// Initialize pool if not already done
	if c.delegatePool == nil {
		c.delegatePool = delegation.NewPool(delegation.DefaultPoolConfig(), c.resourceMonitor)
		c.delegatePool.SetExecutor(c.executeDelegatedTask)
		c.delegatePool.Start(ctx)
	}

	return fantasy.NewParallelAgentTool(
		DelegateToolName,
		string(delegateToolDescription),
		func(ctx context.Context, params DelegateParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			validationResult, err := validateDelegateParams(ctx, params)
			if err != nil {
				return fantasy.NewTextErrorResponse(err.Error()), nil
			}

			// Request permission for delegation
			description := fmt.Sprintf("Delegate task: %s", params.Task)
			if len(description) > 100 {
				description = description[:97] + "..."
			}

			p := c.permissions.Request(
				permission.CreatePermissionRequest{
					SessionID:   validationResult.SessionID,
					Path:        c.cfg.WorkingDir(),
					ToolCallID:  call.ID,
					ToolName:    DelegateToolName,
					Action:      "delegate",
					Description: description,
					Params: DelegatePermissionsParams{
						Task:       params.Task,
						Context:    params.Context,
						WorkingDir: params.WorkingDir,
						Background: params.Background,
					},
				},
			)

			if !p {
				return fantasy.NewTextErrorResponse("Permission denied to delegate task"), nil
			}

			// Determine working directory
			workingDir := params.WorkingDir
			if workingDir == "" {
				workingDir = c.cfg.WorkingDir()
			}

			// Determine max tokens
			maxTokens := int64(params.MaxTokens)
			if maxTokens == 0 {
				maxTokens = 4096
			}

			// Submit to pool
			taskID, done, err := c.delegatePool.Submit(
				params.Task,
				params.Context,
				workingDir,
				maxTokens,
				validationResult.SessionID,
			)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to submit task: %s", err)), nil
			}

			// If background mode, return task ID immediately
			if params.Background {
				stats := c.delegatePool.Stats()
				return fantasy.NewTextResponse(fmt.Sprintf(
					"Task queued with ID: %s\n\nPool status: %d running, %d queued (max %d concurrent)\n\nUse delegate with action='status' and task_id='%s' to check progress.",
					taskID,
					stats.Running,
					stats.Queued,
					stats.MaxConcurrent,
					taskID,
				)), nil
			}

			// Wait for task completion
			<-done

			result, err := c.delegatePool.Wait(taskID)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("delegate task failed: %s", err)), nil
			}

			return fantasy.NewTextResponse(result), nil
		}), nil
}

// executeDelegatedTask runs a delegated task with a sub-agent.
func (c *coordinator) executeDelegatedTask(ctx context.Context, task *delegation.Task) (string, error) {
	// Build the full prompt with context
	fullPrompt := task.Description
	if task.Context != "" {
		fullPrompt = fmt.Sprintf("Context:\n%s\n\nTask:\n%s", task.Context, task.Description)
	}

	// Create prompt template for the delegate
	promptOpts := []prompt.Option{
		prompt.WithWorkingDir(task.WorkingDir),
	}

	promptTemplate, err := prompt.NewPrompt("delegate", string(delegatePromptTmpl), promptOpts...)
	if err != nil {
		return "", fmt.Errorf("error creating prompt: %w", err)
	}

	// Use small model for delegated tasks (efficient and cost-effective)
	_, small, err := c.buildAgentModels(ctx)
	if err != nil {
		return "", fmt.Errorf("error building models: %w", err)
	}

	systemPrompt, err := promptTemplate.Build(ctx, small.Model.Provider(), small.Model.Model(), *c.cfg)
	if err != nil {
		return "", fmt.Errorf("error building system prompt: %w", err)
	}

	smallProviderCfg, ok := c.cfg.Providers.Get(small.ModelCfg.Provider)
	if !ok {
		return "", fmt.Errorf("small model provider not configured")
	}

	// Build tools for the sub-agent - give it access to core tools
	delegateTools := []fantasy.AgentTool{
		tools.NewGlobTool(task.WorkingDir),
		tools.NewGrepTool(task.WorkingDir),
		tools.NewViewTool(c.lspClients, c.permissions, task.WorkingDir),
		tools.NewBashTool(c.permissions, task.WorkingDir, c.cfg.Options.Attribution, small.Model.Model()),
	}

	// Create the sub-agent
	agent := NewSessionAgent(SessionAgentOptions{
		LargeModel:           small,
		SmallModel:           small,
		SystemPromptPrefix:   smallProviderCfg.SystemPromptPrefix,
		SystemPrompt:         systemPrompt,
		DisableAutoSummarize: true, // Sub-agents don't need auto-summarize
		IsYolo:               c.permissions.SkipRequests(),
		Sessions:             c.sessions,
		Messages:             c.messages,
		Tools:                delegateTools,
	})

	// Create a task session for the delegated work
	agentToolSessionID := c.sessions.CreateAgentToolSessionID(task.ParentSession, task.ID)
	taskTitle := "Delegated: " + task.Description
	if len(taskTitle) > 50 {
		taskTitle = taskTitle[:47] + "..."
	}
	session, err := c.sessions.CreateTaskSession(ctx, agentToolSessionID, task.ParentSession, taskTitle)
	if err != nil {
		return "", fmt.Errorf("error creating delegate session: %w", err)
	}

	// Auto-approve permissions for the sub-agent session
	c.permissions.AutoApproveSession(session.ID)

	// Determine max tokens
	maxTokens := task.MaxTokens
	if maxTokens < 1 {
		maxTokens = 4096
		slog.Warn("delegate: MaxTokens < 1, using fallback", "original", maxTokens, "fallback", 4096)
	}

	// Run the sub-agent
	result, err := agent.Run(ctx, SessionAgentCall{
		SessionID:        session.ID,
		Prompt:           fullPrompt,
		MaxOutputTokens:  maxTokens,
		ProviderOptions:  getProviderOptions(small, smallProviderCfg),
		Temperature:      small.ModelCfg.Temperature,
		TopP:             small.ModelCfg.TopP,
		TopK:             small.ModelCfg.TopK,
		FrequencyPenalty: small.ModelCfg.FrequencyPenalty,
		PresencePenalty:  small.ModelCfg.PresencePenalty,
	})
	if err != nil {
		return "", fmt.Errorf("agent run failed: %w", err)
	}

	// Update parent session with costs
	updatedSession, err := c.sessions.Get(ctx, session.ID)
	if err != nil {
		return "", fmt.Errorf("failed to get updated session: %w", err)
	}

	parentSession, err := c.sessions.Get(ctx, task.ParentSession)
	if err != nil {
		return "", fmt.Errorf("failed to get parent session: %w", err)
	}

	parentSession.Cost += updatedSession.Cost
	_, err = c.sessions.Save(ctx, parentSession)
	if err != nil {
		return "", fmt.Errorf("failed to save parent session cost: %w", err)
	}

	return result.Response.Content.Text(), nil
}
