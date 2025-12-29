package agent

import (
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"charm.land/fantasy"

	"github.com/nexora/nexora/internal/agent/delegation"
	"github.com/nexora/nexora/internal/agent/prompt"
	"github.com/nexora/nexora/internal/agent/tools"
	"github.com/nexora/nexora/internal/message"
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
}

// DelegatePermissionsParams is the permission-safe version of DelegateParams
type DelegatePermissionsParams struct {
	Task       string `json:"task"`
	Context    string `json:"context"`
	WorkingDir string `json:"working_dir"`
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

			// Submit to pool - always runs asynchronously
			taskID, _, err := c.delegatePool.Submit(
				params.Task,
				params.Context,
				workingDir,
				maxTokens,
				validationResult.SessionID,
			)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("failed to submit task: %s", err)), nil
			}

			// Always return immediately - the delegate will report back when complete
			stats := c.delegatePool.Stats()
			return fantasy.NewTextResponse(fmt.Sprintf(
				"Task delegated (ID: %s). The sub-agent is now working on this task.\n\nPool status: %d running, %d queued\n\nYou can continue with other work. The delegate will report back when complete.",
				taskID,
				stats.Running,
				stats.Queued,
			)), nil
		}), nil
}

// executeDelegatedTask runs a delegated task with a sub-agent.
// It implements a continuation loop to ensure the agent actually completes work
// rather than just outputting plans without executing them.
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

	// Use large model for delegated tasks (small model support removed)
	_, model, err := c.buildAgentModels(ctx)
	if err != nil {
		return "", fmt.Errorf("error building models: %w", err)
	}

	systemPrompt, err := promptTemplate.Build(ctx, model.Model.Provider(), model.Model.Model(), *c.cfg)
	if err != nil {
		return "", fmt.Errorf("error building system prompt: %w", err)
	}

	providerCfg, ok := c.cfg.Providers.Get(model.ModelCfg.Provider)
	if !ok {
		return "", fmt.Errorf("model provider not configured")
	}

	// Build tools for the sub-agent - give it access to core tools
	delegateTools := []fantasy.AgentTool{
		tools.NewGlobTool(task.WorkingDir),
		tools.NewGrepTool(task.WorkingDir),
		tools.NewViewTool(c.lspClients, c.permissions, task.WorkingDir),
		tools.NewBashTool(c.permissions, task.WorkingDir, c.cfg.Options.Attribution, model.Model.Model()),
	}

	// Create the sub-agent
	agent := NewSessionAgent(SessionAgentOptions{
		LargeModel:           model,
		SmallModel:           model,
		SystemPromptPrefix:   providerCfg.SystemPromptPrefix,
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

	// Run the sub-agent with continuation loop
	// The agent may need multiple turns to actually complete work
	const maxIterations = 10
	var allTextParts []string
	var totalToolCalls int

	for iteration := 0; iteration < maxIterations; iteration++ {
		prompt := fullPrompt
		if iteration > 0 {
			// For continuation, prompt the agent to continue with the task
			prompt = "Continue with the task. Use the available tools (view, glob, grep, bash) to complete the work. Do not just describe what you will do - actually do it using tools."
		}

		slog.Debug("delegate iteration",
			"task_id", task.ID,
			"iteration", iteration,
			"total_tool_calls", totalToolCalls,
		)

		result, err := agent.Run(ctx, SessionAgentCall{
			SessionID:        session.ID,
			Prompt:           prompt,
			MaxOutputTokens:  maxTokens,
			ProviderOptions:  getProviderOptions(model, providerCfg),
			Temperature:      model.ModelCfg.Temperature,
			TopP:             model.ModelCfg.TopP,
			TopK:             model.ModelCfg.TopK,
			FrequencyPenalty: model.ModelCfg.FrequencyPenalty,
			PresencePenalty:  model.ModelCfg.PresencePenalty,
		})
		if err != nil {
			// If we have some results, return them even on error
			if len(allTextParts) > 0 {
				slog.Warn("delegate iteration failed but returning partial results",
					"task_id", task.ID,
					"iteration", iteration,
					"error", err,
				)
				break
			}
			return "", fmt.Errorf("agent run failed: %w", err)
		}

		// Count tool calls by checking messages in the session
		// Tool calls are tracked on assistant messages, not on StepResult
		msgs, listErr := c.messages.List(ctx, session.ID)
		iterationToolCalls := 0
		if listErr == nil {
			for _, msg := range msgs {
				if msg.Role == message.Assistant {
					iterationToolCalls += len(msg.ToolCalls())
				}
			}
		}
		// Only count new tool calls from this iteration
		newToolCalls := iterationToolCalls - totalToolCalls
		if newToolCalls < 0 {
			newToolCalls = 0
		}
		totalToolCalls = iterationToolCalls

		// Extract text content from this iteration
		for _, content := range result.Response.Content {
			if content.GetType() == fantasy.ContentTypeText {
				if tc, ok := content.(fantasy.TextContent); ok && tc.Text != "" {
					allTextParts = append(allTextParts, tc.Text)
				}
			}
		}

		// Check if the agent is done or needs to continue
		responseText := strings.ToLower(result.Response.Content.Text())

		// Check for work-in-progress indicators that suggest we need to continue
		workInProgress := false
		continueIndicators := []string{
			"now let me", "next, i'll", "let me create", "i'll now",
			"let me check", "let me examine", "let me review",
			"let me implement", "now i'll", "moving on to",
			"let me update", "let me modify", "let me add",
		}
		for _, indicator := range continueIndicators {
			if strings.Contains(responseText, indicator) {
				workInProgress = true
				break
			}
		}

		// Check for completion indicators
		completionIndicators := []string{
			"task completed", "finished", "done", "complete",
			"successfully completed", "all set", "here are the results",
			"summary:", "## summary", "findings:",
		}
		taskComplete := false
		for _, indicator := range completionIndicators {
			if strings.Contains(responseText, indicator) {
				taskComplete = true
				break
			}
		}

		// Decision: continue or stop
		// Stop if:
		// 1. Task appears complete, OR
		// 2. We've done real work (tool calls) AND no work-in-progress indicators, OR
		// 3. This iteration had no tool calls and no work-in-progress (agent is stuck)
		if taskComplete {
			slog.Debug("delegate task complete",
				"task_id", task.ID,
				"iteration", iteration,
				"total_tool_calls", totalToolCalls,
			)
			break
		}

		if totalToolCalls > 0 && !workInProgress {
			slog.Debug("delegate finished work",
				"task_id", task.ID,
				"iteration", iteration,
				"total_tool_calls", totalToolCalls,
			)
			break
		}

		if newToolCalls == 0 && !workInProgress {
			// Agent didn't use tools in this iteration and isn't indicating more work - might be stuck
			if iteration > 0 {
				slog.Warn("delegate appears stuck, stopping",
					"task_id", task.ID,
					"iteration", iteration,
					"total_tool_calls", totalToolCalls,
				)
				break
			}
			// First iteration with no tools - continue once to give it a chance
		}

		// If we're still here, continue to next iteration
		slog.Debug("delegate continuing",
			"task_id", task.ID,
			"iteration", iteration,
			"work_in_progress", workInProgress,
			"new_tool_calls", newToolCalls,
			"total_tool_calls", totalToolCalls,
		)
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

	// If no text content found, try reasoning text as fallback
	if len(allTextParts) == 0 {
		// Check messages for any tool results that might be useful
		msgs, listErr := c.messages.List(ctx, session.ID)
		if listErr == nil {
			for _, msg := range msgs {
				if msg.Role == message.Assistant {
					if text := msg.Content().Text; text != "" {
						allTextParts = append(allTextParts, text)
					}
				}
			}
		}

		if len(allTextParts) == 0 {
			return "", fmt.Errorf("delegate produced no text output after %d tool calls", totalToolCalls)
		}
	}

	result := strings.Join(allTextParts, "\n\n")

	// Trigger the main AI with the delegate's report
	// This prompts the parent session to continue with the delegate's findings
	go func() {
		reportPrompt := fmt.Sprintf(
			"[DELEGATE REPORT - Task ID: %s]\n\nThe delegated sub-agent has completed its task.\n\n## Delegate's Findings:\n\n%s\n\n---\nPlease review the delegate's report and continue accordingly.",
			task.ID,
			result,
		)

		slog.Info("delegate reporting to parent session",
			"task_id", task.ID,
			"parent_session", task.ParentSession,
		)

		// Use a fresh context with timeout since the original might be cancelled
		reportCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		if _, runErr := c.Run(reportCtx, task.ParentSession, reportPrompt); runErr != nil {
			slog.Error("failed to report delegate results to parent session",
				"task_id", task.ID,
				"parent_session", task.ParentSession,
				"error", runErr,
			)
		}
	}()

	return result, nil
}
