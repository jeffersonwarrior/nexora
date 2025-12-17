// Package agent is the core orchestration layer for Nexora AI agents.
//
// It provides session-based AI agent functionality for managing
// conversations, tool execution, and message handling. It coordinates
// interactions between language models, messages, sessions, and tools while
// handling features like automatic summarization, queuing, and token
// management.
package agent

import (
	"cmp"
	"context"
	_ "embed"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"charm.land/fantasy"
	"charm.land/fantasy/providers/anthropic"
	"charm.land/fantasy/providers/bedrock"
	"charm.land/fantasy/providers/google"
	"charm.land/fantasy/providers/openai"
	"charm.land/fantasy/providers/openrouter"
	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/nexora/cli/internal/agent/tools"
	"github.com/nexora/cli/internal/aiops"
	"github.com/nexora/cli/internal/config"
	"github.com/nexora/cli/internal/csync"
	"github.com/nexora/cli/internal/message"
	"github.com/nexora/cli/internal/permission"
	"github.com/nexora/cli/internal/session"
	"github.com/nexora/cli/internal/stringext"
)

//go:embed templates/title.md
var titlePrompt []byte

//go:embed templates/summary.md
var summaryPrompt []byte

type SessionAgentCall struct {
	SessionID        string
	Prompt           string
	ProviderOptions  fantasy.ProviderOptions
	Attachments      []message.Attachment
	MaxOutputTokens  int64
	Temperature      *float64
	TopP             *float64
	TopK             *int64
	FrequencyPenalty *float64
	PresencePenalty  *float64
}

type SessionAgent interface {
	Run(context.Context, SessionAgentCall) (*fantasy.AgentResult, error)
	SetModels(large Model, small Model)
	SetTools(tools []fantasy.AgentTool)
	Cancel(sessionID string)
	CancelAll()
	IsSessionBusy(sessionID string) bool
	IsBusy() bool
	QueuedPrompts(sessionID string) int
	ClearQueue(sessionID string)
	Summarize(context.Context, string, fantasy.ProviderOptions) error
	Model() Model
}

type Model struct {
	Model      fantasy.LanguageModel
	CatwalkCfg catwalk.Model
	ModelCfg   config.SelectedModel
}

type sessionAgent struct {
	largeModel           Model
	smallModel           Model
	systemPromptPrefix   string
	systemPrompt         string
	tools                []fantasy.AgentTool
	sessions             session.Service
	messages             message.Service
	disableAutoSummarize bool
	isYolo               bool

	messageQueue   *csync.Map[string, []SessionAgentCall]
	activeRequests *csync.Map[string, context.CancelFunc]
	aiops          aiops.Ops // AIOPS client for operational support

	// State for loop and drift detection
	recentCalls       []aiops.ToolCall
	callCount         int
	recentActions     []aiops.Action
	actionCount       int
	recentCallsLock   sync.Mutex
	recentActionsLock sync.Mutex
}

type SessionAgentOptions struct {
	LargeModel           Model
	SmallModel           Model
	SystemPromptPrefix   string
	SystemPrompt         string
	DisableAutoSummarize bool
	IsYolo               bool
	Sessions             session.Service
	Messages             message.Service
	Tools                []fantasy.AgentTool
	AIOPS                aiops.Ops // AIOPS client
}

func NewSessionAgent(
	opts SessionAgentOptions,
) SessionAgent {
	return &sessionAgent{
		largeModel:           opts.LargeModel,
		smallModel:           opts.SmallModel,
		systemPromptPrefix:   opts.SystemPromptPrefix,
		systemPrompt:         opts.SystemPrompt,
		sessions:             opts.Sessions,
		messages:             opts.Messages,
		disableAutoSummarize: opts.DisableAutoSummarize,
		tools:                opts.Tools,
		isYolo:               opts.IsYolo,
		messageQueue:         csync.NewMap[string, []SessionAgentCall](),
		activeRequests:       csync.NewMap[string, context.CancelFunc](),
		aiops:                opts.AIOPS,
	}
}

func (a *sessionAgent) getOriginalPrompt(sessionID string) string {
	// Get the first user message from the session
	messages, err := a.messages.List(context.Background(), sessionID)
	if err != nil {
		return ""
	}
	
	for _, msg := range messages {
		if msg.Role == message.User {
			return msg.Content().Text
		}
	}
	return ""
}

func (a *sessionAgent) Run(ctx context.Context, call SessionAgentCall) (*fantasy.AgentResult, error) {
	// Handle special continuation prompts
	isContinuation := call.Prompt == "CONTINUE_AFTER_TOOL_EXECUTION"
	
	if call.Prompt == "" && !isContinuation {
		return nil, ErrEmptyPrompt
	}
	if isContinuation {
		// For continuation, use the original user prompt
		call.Prompt = a.getOriginalPrompt(call.SessionID)
		if call.Prompt == "" {
			// Fallback if we can't get the original prompt
			call.Prompt = "Please continue with the previous task"
		}
	}
	if call.SessionID == "" {
		return nil, ErrSessionMissing
	}

	// Queue the message if busy
	if a.IsSessionBusy(call.SessionID) {
		existing, ok := a.messageQueue.Get(call.SessionID)
		if !ok {
			existing = []SessionAgentCall{}
		}
		// Limit queue size to prevent memory issues
		if len(existing) >= 50 {
			return nil, fmt.Errorf("session %s: queue is full (max 50 queued requests)", call.SessionID)
		}
		existing = append(existing, call)
		a.messageQueue.Set(call.SessionID, existing)
		// Return a specific error to indicate message was queued, not processed
		return nil, fmt.Errorf("session %s: message queued (position %d in queue)", call.SessionID, len(existing))
	}

	if len(a.tools) > 0 {
		// Add Anthropic caching to the last tool.
		a.tools[len(a.tools)-1].SetProviderOptions(a.getCacheControlOptions())
	}

	agent := fantasy.NewAgent(
		a.largeModel.Model,
		fantasy.WithSystemPrompt(a.systemPrompt),
		fantasy.WithTools(a.tools...),
	)

	sessionLock := sync.Mutex{}
	currentSession, err := a.sessions.Get(ctx, call.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	msgs, err := a.getSessionMessages(ctx, currentSession)
	if err != nil {
		return nil, fmt.Errorf("failed to get session messages: %w", err)
	}

	var wg sync.WaitGroup
	// Generate title if first message.
	if len(msgs) == 0 {
		wg.Go(func() {
			sessionLock.Lock()
			a.generateTitle(ctx, &currentSession, call.Prompt)
			sessionLock.Unlock()
		})
	}

	// Add the user message to the session.
	_, err = a.createUserMessage(ctx, call)
	if err != nil {
		return nil, err
	}

	// Add the session to the context.
	ctx = context.WithValue(ctx, tools.SessionIDContextKey, call.SessionID)

	genCtx, cancel := context.WithCancel(ctx)
	a.activeRequests.Set(call.SessionID, cancel)

	defer cancel()
	defer a.activeRequests.Del(call.SessionID)

	history, files := a.preparePrompt(msgs, call.Attachments...)

	startTime := time.Now()
	a.eventPromptSent(call.SessionID)

	var currentAssistant *message.Message
	var shouldSummarize bool
	result, err := agent.Stream(genCtx, fantasy.AgentStreamCall{
		Prompt:           call.Prompt,
		Files:            files,
		Messages:         history,
		ProviderOptions:  call.ProviderOptions,
		MaxOutputTokens:  &call.MaxOutputTokens,
		TopP:             call.TopP,
		Temperature:      call.Temperature,
		PresencePenalty:  call.PresencePenalty,
		TopK:             call.TopK,
		FrequencyPenalty: call.FrequencyPenalty,
		PrepareStep: func(callContext context.Context, options fantasy.PrepareStepFunctionOptions) (_ context.Context, prepared fantasy.PrepareStepResult, err error) {
			prepared.Messages = options.Messages
			for i := range prepared.Messages {
				prepared.Messages[i].ProviderOptions = nil
			}

			queuedCalls, _ := a.messageQueue.Get(call.SessionID)
			a.messageQueue.Del(call.SessionID)
			for _, queued := range queuedCalls {
				userMessage, createErr := a.createUserMessage(callContext, queued)
				if createErr != nil {
					return callContext, prepared, createErr
				}
				prepared.Messages = append(prepared.Messages, userMessage.ToAIMessage()...)
			}

			prepared.Messages = a.workaroundProviderMediaLimitations(prepared.Messages)

			lastSystemRoleInx := 0
			systemMessageUpdated := false
			for i, msg := range prepared.Messages {
				// Only add cache control to the last message.
				if msg.Role == fantasy.MessageRoleSystem {
					lastSystemRoleInx = i
				} else if !systemMessageUpdated {
					prepared.Messages[lastSystemRoleInx].ProviderOptions = a.getCacheControlOptions()
					systemMessageUpdated = true
				}
				// Than add cache control to the last 2 messages.
				if i > len(prepared.Messages)-3 {
					prepared.Messages[i].ProviderOptions = a.getCacheControlOptions()
				}
			}

			if promptPrefix := a.promptPrefix(); promptPrefix != "" {
				prepared.Messages = append([]fantasy.Message{fantasy.NewSystemMessage(promptPrefix)}, prepared.Messages...)
			}

			var assistantMsg message.Message
			assistantMsg, err = a.messages.Create(callContext, call.SessionID, message.CreateMessageParams{
				Role:     message.Assistant,
				Parts:    []message.ContentPart{},
				Model:    a.largeModel.ModelCfg.Model,
				Provider: a.largeModel.ModelCfg.Provider,
			})
			if err != nil {
				return callContext, prepared, err
			}
			callContext = context.WithValue(callContext, tools.MessageIDContextKey, assistantMsg.ID)
			callContext = context.WithValue(callContext, tools.SupportsImagesContextKey, a.largeModel.CatwalkCfg.SupportsImages)
			callContext = context.WithValue(callContext, tools.ModelNameContextKey, a.largeModel.CatwalkCfg.Name)
			currentAssistant = &assistantMsg
			return callContext, prepared, err
		},
		OnReasoningStart: func(id string, reasoning fantasy.ReasoningContent) error {
			currentAssistant.AppendReasoningContent(reasoning.Text)
			return a.messages.Update(genCtx, *currentAssistant)
		},
		OnReasoningDelta: func(id string, text string) error {
			currentAssistant.AppendReasoningContent(text)
			return a.messages.Update(genCtx, *currentAssistant)
		},
		OnReasoningEnd: func(id string, reasoning fantasy.ReasoningContent) error {
			// handle anthropic signature
			if anthropicData, ok := reasoning.ProviderMetadata[anthropic.Name]; ok {
				if reasoning, ok := anthropicData.(*anthropic.ReasoningOptionMetadata); ok {
					currentAssistant.AppendReasoningSignature(reasoning.Signature)
				}
			}
			if googleData, ok := reasoning.ProviderMetadata[google.Name]; ok {
				if reasoning, ok := googleData.(*google.ReasoningMetadata); ok {
					currentAssistant.AppendThoughtSignature(reasoning.Signature, reasoning.ToolID)
				}
			}
			if openaiData, ok := reasoning.ProviderMetadata[openai.Name]; ok {
				if reasoning, ok := openaiData.(*openai.ResponsesReasoningMetadata); ok {
					currentAssistant.SetReasoningResponsesData(reasoning)
				}
			}
			currentAssistant.FinishThinking()
			return a.messages.Update(genCtx, *currentAssistant)
		},
		OnTextDelta: func(id string, text string) error {
			// Strip leading newline from initial text content. This is is
			// particularly important in non-interactive mode where leading
			// newlines are very visible.
			if len(currentAssistant.Parts) == 0 {
				text = strings.TrimPrefix(text, "\n")
			}

			currentAssistant.AppendContent(text)
			return a.messages.Update(genCtx, *currentAssistant)
		},
		OnToolInputStart: func(id string, toolName string) error {
			toolCall := message.ToolCall{
				ID:               id,
				Name:             toolName,
				ProviderExecuted: false,
				Finished:         false,
			}
			currentAssistant.AddToolCall(toolCall)
			return a.messages.Update(genCtx, *currentAssistant)
		},
		OnRetry: func(err *fantasy.ProviderError, delay time.Duration) {
			// Log the retry attempt
			slog.Warn("Provider request failed, retrying",
				"error", err.Error(),
				"delay", delay,
				"session_id", currentSession.ID,
			)

			// Show user-facing message for rate limit errors
			if err.StatusCode == 429 {
				_, createErr := a.messages.Create(genCtx, currentAssistant.SessionID, message.CreateMessageParams{
					Role: message.System,
					Parts: []message.ContentPart{
						message.TextContent{Text: "Model is rate limited. Retrying request..."},
					},
				})
				if createErr != nil {
					slog.Error("Failed to create rate limit message", "error", createErr, "session_id", currentSession.ID)
				}
			}
		},
		OnToolCall: func(tc fantasy.ToolCallContent) error {
			toolCall := message.ToolCall{
				ID:               tc.ToolCallID,
				Name:             tc.ToolName,
				Input:            tc.Input,
				ProviderExecuted: false,
				Finished:         true,
			}
			currentAssistant.AddToolCall(toolCall)
			return a.messages.Update(genCtx, *currentAssistant)
		},
		OnToolResult: func(result fantasy.ToolResultContent) error {
			// Track tool call for loop detection
			if a.aiops != nil {
				var content string
				var errorMsg string

				switch result.Result.GetType() {
				case fantasy.ToolResultContentTypeText:
					if r, ok := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentText](result.Result); ok {
						content = r.Text
					}
				case fantasy.ToolResultContentTypeError:
					if r, ok := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentError](result.Result); ok {
						errorMsg = r.Error.Error()
					}
				case fantasy.ToolResultContentTypeMedia:
					if r, ok := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentMedia](result.Result); ok {
						content = fmt.Sprintf("Loaded %s content", r.MediaType)
					}
				}

				if errorMsg != "" {
					content = errorMsg // Track error messages for loop detection
				}

				// Convert to AIOPS ToolCall for tracking
				aiopsCall := aiops.ToolCall{
					ID:        result.ToolCallID,
					Name:      result.ToolName,
					Result:    content,
					Error:     errorMsg,
					Timestamp: time.Now(),
				}
				if len(aiopsCall.Result) > 10000 {
					aiopsCall.Result = aiopsCall.Result[:10000] + "..."
				}

				// Add to recent calls
				a.recentCallsLock.Lock()
				if len(a.recentCalls) >= 10 {
					// Keep only last 10 calls
					copy(a.recentCalls, a.recentCalls[1:])
					a.recentCalls[len(a.recentCalls)-1] = aiopsCall
				} else {
					a.recentCalls = append(a.recentCalls, aiopsCall)
				}

				// Increment call count
				a.callCount++

				// Check for loops every 5 tool calls
				if a.callCount%5 == 0 && len(a.recentCalls) >= 5 {
					go func(calls []aiops.ToolCall, sessionID string, agentCtx context.Context) {
						detectCtx, cancel := context.WithTimeout(agentCtx, 2*time.Second)
						defer cancel()
						detection, err := a.aiops.DetectLoop(detectCtx, calls)
						if err != nil {
				slog.Debug("Loop detection failed", "error", err, "session_id", sessionID)
				return
			}
			if detection.IsLooping {
							// Create a system message about the loop
							_, createErr := a.messages.Create(genCtx, currentAssistant.SessionID, message.CreateMessageParams{
								Role: message.System,
								Parts: []message.ContentPart{
									message.TextContent{
										Text: fmt.Sprintf("🔁 Loop detected: %s. Suggestion: %s", detection.Reason, detection.Suggestion),
									},
								},
							})
							if createErr != nil {
								slog.Error("Failed to create loop detection message", "error", createErr, "session_id", sessionID)
							}
						}
					}(a.recentCalls, currentAssistant.SessionID, genCtx)
				}

				// Track actions for drift detection
				a.recentActions = append(a.recentActions, aiops.Action{
					Description: fmt.Sprintf("Tool call: %s", result.ToolName),
					ToolCalls: []aiops.ToolCall{
						{
							ID:        result.ToolCallID,
							Name:      result.ToolName,
							Params:    map[string]any{}, // We don't have access to the original params in ToolResultContent
							Result:    content,
							Error:     errorMsg,
							Timestamp: time.Now(),
						},
					},
					Timestamp: time.Now(),
				})
				// Keep only the last 20 actions
				if len(a.recentActions) > 20 {
					a.recentActions = a.recentActions[len(a.recentActions)-20:]
				}
				a.actionCount++
				// Check for drift every 10 tool calls
				if a.actionCount%10 == 0 && a.aiops != nil {
					go func(actions []aiops.Action, sessionID string, agentCtx context.Context) {
						detectCtx, cancel := context.WithTimeout(agentCtx, 2*time.Second)
						defer cancel()

						// Get the original task from the first message in the session
						originalTask := ""
						if msgs, err := a.messages.List(detectCtx, sessionID); err == nil && len(msgs) > 0 {
							// Extract task from the first user message
							for _, msg := range msgs {
								if msg.Role == message.User {
									for _, part := range msg.Parts {
										if tc, ok := part.(message.TextContent); ok {
											originalTask = tc.Text
											break
										}
									}
									if originalTask != "" {
										break
									}
								}
							}
						}

						if originalTask != "" {
							drift, err := a.aiops.DetectDrift(detectCtx, originalTask, actions)
							if err != nil {
							slog.Debug("Drift detection failed", "error", err, "session_id", sessionID)
							return
						}
						if drift.IsDrifting {
								// Create a warning message about task drift
								_, createErr := a.messages.Create(agentCtx, sessionID, message.CreateMessageParams{
									Role: message.System,
									Parts: []message.ContentPart{
										message.TextContent{
											Text: fmt.Sprintf("⚠️ Task drift detected: %s. Suggestion: %s", drift.Reason, drift.Suggestion),
										},
									},
								})
								if createErr != nil {
									slog.Error("Failed to create drift detection message", "error", createErr, "session_id", sessionID)
								}
							}
						}
					}(a.recentActions, currentAssistant.SessionID, genCtx)
				}
				a.recentCallsLock.Unlock()
			}

			toolResult := a.convertToToolResult(result)
			_, createMsgErr := a.messages.Create(genCtx, currentAssistant.SessionID, message.CreateMessageParams{
				Role: message.Tool,
				Parts: []message.ContentPart{
					toolResult,
				},
			})
			return createMsgErr
		},
		OnStepFinish: func(stepResult fantasy.StepResult) error {
			finishReason := message.FinishReasonUnknown
			switch stepResult.FinishReason {
			case fantasy.FinishReasonLength:
				finishReason = message.FinishReasonMaxTokens
			case fantasy.FinishReasonStop:
				finishReason = message.FinishReasonEndTurn
			case fantasy.FinishReasonToolCalls:
				finishReason = message.FinishReasonToolUse
			}
			currentAssistant.AddFinish(finishReason, "", "")
			a.updateSessionUsage(a.largeModel, &currentSession, stepResult.Usage, a.openrouterCost(stepResult.ProviderMetadata))
			sessionLock.Lock()
			_, sessionErr := a.sessions.Save(genCtx, currentSession)
			sessionLock.Unlock()
			if sessionErr != nil {
				return sessionErr
			}
			return a.messages.Update(genCtx, *currentAssistant)
		},
		StopWhen: []fantasy.StopCondition{
			func(steps []fantasy.StepResult) bool {
				cw := int64(a.largeModel.CatwalkCfg.ContextWindow)
				tokens := currentSession.CompletionTokens + currentSession.PromptTokens
				remaining := cw - tokens
				var threshold int64
				if cw > 200_000 {
					threshold = 20_000
				} else {
					threshold = int64(float64(cw) * 0.2)
				}
				// Only stop if we're not actively streaming a response
				isCurrentlyStreaming := len(steps) > 0 && steps[len(steps)-1].FinishReason == ""
				if (remaining <= threshold) && !a.disableAutoSummarize && !isCurrentlyStreaming {
					shouldSummarize = true
					return true
				}
				return false
			},
		},
	})

	a.eventPromptResponded(call.SessionID, time.Since(startTime).Truncate(time.Second))

	if err != nil {
		isCancelErr := errors.Is(err, context.Canceled)
		isPermissionErr := errors.Is(err, permission.ErrorPermissionDenied)
		if currentAssistant == nil {
			return result, err
		}
		// Ensure we finish thinking on error to close the reasoning state.
		currentAssistant.FinishThinking()
		toolCalls := currentAssistant.ToolCalls()
		// INFO: we use the parent context here because the genCtx has been cancelled.
		msgs, createErr := a.messages.List(ctx, currentAssistant.SessionID)
		if createErr != nil {
			return nil, createErr
		}
		for _, tc := range toolCalls {
			if !tc.Finished {
				tc.Finished = true
				tc.Input = "{}"
				currentAssistant.AddToolCall(tc)
				updateErr := a.messages.Update(ctx, *currentAssistant)
				if updateErr != nil {
					return nil, updateErr
				}
			}

			found := false
			for _, msg := range msgs {
				if msg.Role == message.Tool {
					for _, tr := range msg.ToolResults() {
						if tr.ToolCallID == tc.ID {
							found = true
							break
						}
					}
				}
				if found {
					break
				}
			}
			if found {
				continue
			}
			content := "There was an error while executing the tool"
			isViewDirectoryHelp := false
			if isCancelErr {
				content = "Tool execution canceled by user"
			} else if isPermissionErr {
				content = "User denied permission"
			} else if tc.Name == "view" && strings.Contains(err.Error(), "Path is a directory") {
				// Automatic recovery for VIEW tool directory errors
				// Extract the directory path from the error message
				parts := strings.Split(err.Error(), ": ")
				if len(parts) >= 2 {
					dirPath := strings.TrimSpace(parts[1])
					// List directory contents to help AI understand what's available
					if dirEntries, dirErr := os.ReadDir(dirPath); dirErr == nil {
						var fileList []string
						for _, entry := range dirEntries {
							fileList = append(fileList, entry.Name())
						}

						// Create helpful response with directory contents
						response := fmt.Sprintf("Path is a directory: %s\n\nDirectory contents:\n", dirPath)
						for i, file := range fileList {
							response += fmt.Sprintf("%d. %s\n", i+1, file)
						}

						response += "\n💡 Suggestions:\n"
						response += "- Use 'view' with a specific file path (e.g., 'view " + dirPath + "/filename')\n"
						response += "- Use 'ls' command to explore directory structure\n"
						response += "- Try 'find' to search for specific files\n"

						content = response
						// Don't mark as error - this is a helpful response
						isViewDirectoryHelp = true
					} else {
						content = fmt.Sprintf("Path is a directory: %s. Cannot read directory contents: %v", dirPath, dirErr)
					}
				} else {
					content = "Path is a directory. Cannot extract directory path from error."
				}
			}
			toolResult := message.ToolResult{
				ToolCallID: tc.ID,
				Name:       tc.Name,
				Content:    content,
				IsError:    true && !isViewDirectoryHelp,
			}
			_, createErr = a.messages.Create(context.Background(), currentAssistant.SessionID, message.CreateMessageParams{
				Role: message.Tool,
				Parts: []message.ContentPart{
					toolResult,
				},
			})
			if createErr != nil {
				return nil, createErr
			}
		}
		var fantasyErr *fantasy.Error
		var providerErr *fantasy.ProviderError
		const defaultTitle = "Provider Error"
		if isCancelErr {
			currentAssistant.AddFinish(message.FinishReasonCanceled, "User canceled request", "")
		} else if isPermissionErr {
			currentAssistant.AddFinish(message.FinishReasonPermissionDenied, "User denied permission", "")
		} else if errors.As(err, &providerErr) {
			currentAssistant.AddFinish(message.FinishReasonError, cmp.Or(stringext.Capitalize(providerErr.Title), defaultTitle), providerErr.Message)
		} else if errors.As(err, &fantasyErr) {
			currentAssistant.AddFinish(message.FinishReasonError, cmp.Or(stringext.Capitalize(fantasyErr.Title), defaultTitle), fantasyErr.Message)
		} else {
			currentAssistant.AddFinish(message.FinishReasonError, defaultTitle, err.Error())
		}
		// Note: we use the parent context here because the genCtx has been
		// cancelled.
		updateErr := a.messages.Update(ctx, *currentAssistant)
		if updateErr != nil {
			return nil, updateErr
		}
		return nil, err
	}
	wg.Wait()

	if shouldSummarize {
		a.activeRequests.Del(call.SessionID)
		if summarizeErr := a.Summarize(genCtx, call.SessionID, call.ProviderOptions); summarizeErr != nil {
			return nil, summarizeErr
		}
		// If the agent wasn't done...
		if len(currentAssistant.ToolCalls()) > 0 {
			existing, ok := a.messageQueue.Get(call.SessionID)
			if !ok {
				existing = []SessionAgentCall{}
			}
			call.Prompt = fmt.Sprintf("The previous session was interrupted because it got too long, the initial user request was: `%s`", call.Prompt)
			existing = append(existing, call)
			a.messageQueue.Set(call.SessionID, existing)
		}
	}

	// Check for queued messages before cleaning up
	queuedMessages, ok := a.messageQueue.Get(call.SessionID)
	if !ok || len(queuedMessages) == 0 {
		// No queued messages
		a.activeRequests.Del(call.SessionID)
		
		// Check if we have tool calls that were just completed
		// If we have tool results but no more queued messages, we need to continue
		// the conversation by calling Run again with an empty prompt
		hasToolResults := false
		if messages, err := a.messages.List(ctx, currentAssistant.SessionID); err == nil {
			for _, msg := range messages {
				if msg.Role == message.Tool {
					hasToolResults = true
					break
				}
			}
		}
		
		// If we just completed tool calls, continue the conversation
		if hasToolResults && currentAssistant != nil && len(currentAssistant.ToolCalls()) > 0 {
			// Create a new context for the next run but don't cancel the parent
			nextRunCtx := context.WithValue(ctx, tools.SessionIDContextKey, call.SessionID)
			// Run with special continuation prompt to continue after tool execution
			return a.Run(nextRunCtx, SessionAgentCall{
				SessionID: call.SessionID,
				Prompt:    "CONTINUE_AFTER_TOOL_EXECUTION", // Special prompt to continue
				ProviderOptions: call.ProviderOptions,
			})
		}
		
		// Otherwise, clean up and return
		cancel()
		return result, err
	}
	
	// Has queued messages, create new context for next run
	firstQueuedMessage := queuedMessages[0]
	a.messageQueue.Set(call.SessionID, queuedMessages[1:])
	
	// Release active request for this call, but don't cancel parent context
	a.activeRequests.Del(call.SessionID)
	
	// Start new run with fresh context to avoid cancellation propagation
	nextRunCtx := context.WithValue(ctx, tools.SessionIDContextKey, call.SessionID)
	return a.Run(nextRunCtx, firstQueuedMessage)
}

func (a *sessionAgent) Summarize(ctx context.Context, sessionID string, opts fantasy.ProviderOptions) error {
	if a.IsSessionBusy(sessionID) {
		return ErrSessionBusy
	}

	currentSession, err := a.sessions.Get(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to get session: %w", err)
	}
	msgs, err := a.getSessionMessages(ctx, currentSession)
	if err != nil {
		return err
	}
	if len(msgs) == 0 {
		// Nothing to summarize.
		return nil
	}

	aiMsgs, _ := a.preparePrompt(msgs)

	genCtx, cancel := context.WithCancel(ctx)
	a.activeRequests.Set(sessionID, cancel)
	defer a.activeRequests.Del(sessionID)
	defer cancel()

	agent := fantasy.NewAgent(a.largeModel.Model,
		fantasy.WithSystemPrompt(string(summaryPrompt)),
	)
	summaryMessage, err := a.messages.Create(ctx, sessionID, message.CreateMessageParams{
		Role:             message.Assistant,
		Model:            a.largeModel.Model.Model(),
		Provider:         a.largeModel.Model.Provider(),
		IsSummaryMessage: true,
	})
	if err != nil {
		return err
	}

	resp, err := agent.Stream(genCtx, fantasy.AgentStreamCall{
		Prompt:          "Provide a detailed summary of our conversation above.",
		Messages:        aiMsgs,
		ProviderOptions: opts,
		PrepareStep: func(callContext context.Context, options fantasy.PrepareStepFunctionOptions) (_ context.Context, prepared fantasy.PrepareStepResult, err error) {
			prepared.Messages = options.Messages
			if a.systemPromptPrefix != "" {
				prepared.Messages = append([]fantasy.Message{fantasy.NewSystemMessage(a.systemPromptPrefix)}, prepared.Messages...)
			}
			return callContext, prepared, nil
		},
		OnReasoningDelta: func(id string, text string) error {
			summaryMessage.AppendReasoningContent(text)
			return a.messages.Update(genCtx, summaryMessage)
		},
		OnReasoningEnd: func(id string, reasoning fantasy.ReasoningContent) error {
			// Handle anthropic signature.
			if anthropicData, ok := reasoning.ProviderMetadata["anthropic"]; ok {
				if signature, ok := anthropicData.(*anthropic.ReasoningOptionMetadata); ok && signature.Signature != "" {
					summaryMessage.AppendReasoningSignature(signature.Signature)
				}
			}
			summaryMessage.FinishThinking()
			return a.messages.Update(genCtx, summaryMessage)
		},
		OnTextDelta: func(id, text string) error {
			summaryMessage.AppendContent(text)
			return a.messages.Update(genCtx, summaryMessage)
		},
	})
	if err != nil {
		isCancelErr := errors.Is(err, context.Canceled)
		if isCancelErr {
			// User cancelled summarize we need to remove the summary message.
			deleteErr := a.messages.Delete(ctx, summaryMessage.ID)
			return deleteErr
		}
		return err
	}

	summaryMessage.AddFinish(message.FinishReasonEndTurn, "", "")
	err = a.messages.Update(genCtx, summaryMessage)
	if err != nil {
		return err
	}

	var openrouterCost *float64
	for _, step := range resp.Steps {
		stepCost := a.openrouterCost(step.ProviderMetadata)
		if stepCost != nil {
			newCost := *stepCost
			if openrouterCost != nil {
				newCost += *openrouterCost
			}
			openrouterCost = &newCost
		}
	}

	a.updateSessionUsage(a.largeModel, &currentSession, resp.TotalUsage, openrouterCost)

	// Just in case, get just the last usage info.
	usage := resp.Response.Usage
	currentSession.SummaryMessageID = summaryMessage.ID
	currentSession.CompletionTokens = usage.OutputTokens
	currentSession.PromptTokens = 0
	_, err = a.sessions.Save(genCtx, currentSession)
	return err
}

func (a *sessionAgent) getCacheControlOptions() fantasy.ProviderOptions {
	if t, _ := strconv.ParseBool(os.Getenv("NEXORA_DISABLE_ANTHROPIC_CACHE")); t {
		return fantasy.ProviderOptions{}
	}
	return fantasy.ProviderOptions{
		anthropic.Name: &anthropic.ProviderCacheControlOptions{
			CacheControl: anthropic.CacheControl{Type: "ephemeral"},
		},
		bedrock.Name: &anthropic.ProviderCacheControlOptions{
			CacheControl: anthropic.CacheControl{Type: "ephemeral"},
		},
	}
}

func (a *sessionAgent) createUserMessage(ctx context.Context, call SessionAgentCall) (message.Message, error) {
	var attachmentParts []message.ContentPart
	for _, attachment := range call.Attachments {
		attachmentParts = append(attachmentParts, message.BinaryContent{Path: attachment.FilePath, MIMEType: attachment.MimeType, Data: attachment.Content})
	}
	parts := []message.ContentPart{message.TextContent{Text: call.Prompt}}
	parts = append(parts, attachmentParts...)
	msg, err := a.messages.Create(ctx, call.SessionID, message.CreateMessageParams{
		Role:  message.User,
		Parts: parts,
	})
	if err != nil {
		return message.Message{}, fmt.Errorf("failed to create user message: %w", err)
	}
	return msg, nil
}

func (a *sessionAgent) preparePrompt(msgs []message.Message, attachments ...message.Attachment) ([]fantasy.Message, []fantasy.FilePart) {
	var history []fantasy.Message
	for _, m := range msgs {
		if len(m.Parts) == 0 {
			continue
		}
		// Assistant message without content or tool calls (cancelled before it
		// returned anything).
		if m.Role == message.Assistant && len(m.ToolCalls()) == 0 && m.Content().Text == "" && m.ReasoningContent().String() == "" {
			continue
		}
		history = append(history, m.ToAIMessage()...)
	}

	var files []fantasy.FilePart
	for _, attachment := range attachments {
		files = append(files, fantasy.FilePart{
			Filename:  attachment.FileName,
			Data:      attachment.Content,
			MediaType: attachment.MimeType,
		})
	}

	return history, files
}

func (a *sessionAgent) getSessionMessages(ctx context.Context, session session.Session) ([]message.Message, error) {
	msgs, err := a.messages.List(ctx, session.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list messages: %w", err)
	}

	if session.SummaryMessageID != "" {
		summaryMsgInex := -1
		for i, msg := range msgs {
			if msg.ID == session.SummaryMessageID {
				summaryMsgInex = i
				break
			}
		}
		if summaryMsgInex != -1 {
			msgs = msgs[summaryMsgInex:]
			msgs[0].Role = message.User
		}
	}
	return msgs, nil
}

func (a *sessionAgent) generateTitle(ctx context.Context, session *session.Session, prompt string) {
	if prompt == "" {
		return
	}

	var maxOutput int64 = 40
	if a.smallModel.CatwalkCfg.CanReason {
		maxOutput = a.smallModel.CatwalkCfg.DefaultMaxTokens
	}

	agent := fantasy.NewAgent(a.smallModel.Model,
		fantasy.WithSystemPrompt(string(titlePrompt)+"\n /no_think"),
		fantasy.WithMaxOutputTokens(maxOutput),
	)

	resp, err := agent.Stream(ctx, fantasy.AgentStreamCall{
		Prompt: fmt.Sprintf("Generate a concise title for the following content:\n\n%s\n <think>\n\n</think>", prompt),
		PrepareStep: func(callContext context.Context, options fantasy.PrepareStepFunctionOptions) (_ context.Context, prepared fantasy.PrepareStepResult, err error) {
			prepared.Messages = options.Messages
			if a.systemPromptPrefix != "" {
				prepared.Messages = append([]fantasy.Message{fantasy.NewSystemMessage(a.systemPromptPrefix)}, prepared.Messages...)
			}
			return callContext, prepared, nil
		},
	})
	if err != nil {
		slog.Error("error generating title", "err", err)
		return
	}

	title := resp.Response.Content.Text()

	title = strings.ReplaceAll(title, "\n", " ")

	// Remove thinking tags if present.
	if idx := strings.Index(title, "</think>"); idx > 0 {
		title = title[idx+len("</think>"):]
	}

	title = strings.TrimSpace(title)
	if title == "" {
		slog.Warn("failed to generate title", "warn", "empty title")
		return
	}

	session.Title = title

	var openrouterCost *float64
	for _, step := range resp.Steps {
		stepCost := a.openrouterCost(step.ProviderMetadata)
		if stepCost != nil {
			newCost := *stepCost
			if openrouterCost != nil {
				newCost += *openrouterCost
			}
			openrouterCost = &newCost
		}
	}

	a.updateSessionUsage(a.smallModel, session, resp.TotalUsage, openrouterCost)
	_, saveErr := a.sessions.Save(ctx, *session)
	if saveErr != nil {
		slog.Error("failed to save session title & usage", "error", saveErr)
		return
	}
}

func (a *sessionAgent) openrouterCost(metadata fantasy.ProviderMetadata) *float64 {
	openrouterMetadata, ok := metadata[openrouter.Name]
	if !ok {
		return nil
	}

	opts, ok := openrouterMetadata.(*openrouter.ProviderMetadata)
	if !ok {
		return nil
	}
	return &opts.Usage.Cost
}

func (a *sessionAgent) updateSessionUsage(model Model, session *session.Session, usage fantasy.Usage, overrideCost *float64) {
	modelConfig := model.CatwalkCfg
	cost := modelConfig.CostPer1MInCached/1e6*float64(usage.CacheCreationTokens) +
		modelConfig.CostPer1MOutCached/1e6*float64(usage.CacheReadTokens) +
		modelConfig.CostPer1MIn/1e6*float64(usage.InputTokens) +
		modelConfig.CostPer1MOut/1e6*float64(usage.OutputTokens)

	if a.isClaudeCode() {
		cost = 0
	}

	a.eventTokensUsed(session.ID, model, usage, cost)

	if overrideCost != nil {
		session.Cost += *overrideCost
	} else {
		session.Cost += cost
	}

	session.CompletionTokens = usage.OutputTokens + usage.CacheReadTokens
	session.PromptTokens = usage.InputTokens + usage.CacheCreationTokens
}

func (a *sessionAgent) Cancel(sessionID string) {
	// Cancel regular requests.
	if cancel, ok := a.activeRequests.Take(sessionID); ok && cancel != nil {
		slog.Info("Request cancellation initiated", "session_id", sessionID)
		cancel()
	}

	// Also check for summarize requests.
	if cancel, ok := a.activeRequests.Take(sessionID + "-summarize"); ok && cancel != nil {
		slog.Info("Summarize cancellation initiated", "session_id", sessionID)
		cancel()
	}

	if a.QueuedPrompts(sessionID) > 0 {
		slog.Info("Clearing queued prompts", "session_id", sessionID)
		a.messageQueue.Del(sessionID)
	}
}

func (a *sessionAgent) ClearQueue(sessionID string) {
	if a.QueuedPrompts(sessionID) > 0 {
		slog.Info("Clearing queued prompts", "session_id", sessionID)
		a.messageQueue.Del(sessionID)
	}
}

func (a *sessionAgent) CancelAll() {
	if !a.IsBusy() {
		return
	}
	for key := range a.activeRequests.Seq2() {
		a.Cancel(key) // key is sessionID
	}

	timeout := time.After(5 * time.Second)
	for a.IsBusy() {
		select {
		case <-timeout:
			return
		default:
			time.Sleep(200 * time.Millisecond)
		}
	}
}

func (a *sessionAgent) IsBusy() bool {
	var busy bool
	for cancelFunc := range a.activeRequests.Seq() {
		if cancelFunc != nil {
			busy = true
			break
		}
	}
	return busy
}

func (a *sessionAgent) IsSessionBusy(sessionID string) bool {
	_, busy := a.activeRequests.Get(sessionID)
	return busy
}

func (a *sessionAgent) QueuedPrompts(sessionID string) int {
	l, ok := a.messageQueue.Get(sessionID)
	if !ok {
		return 0
	}
	return len(l)
}

func (a *sessionAgent) SetModels(large Model, small Model) {
	a.largeModel = large
	a.smallModel = small
}

func (a *sessionAgent) SetTools(tools []fantasy.AgentTool) {
	a.tools = tools
}

func (a *sessionAgent) Model() Model {
	return a.largeModel
}

func (a *sessionAgent) promptPrefix() string {
	if a.isClaudeCode() {
		return "You are Claude Code, Anthropic's official CLI for Claude."
	}
	return a.systemPromptPrefix
}

func (a *sessionAgent) isClaudeCode() bool {
	cfg := config.Get()
	pc, ok := cfg.Providers.Get(a.largeModel.ModelCfg.Provider)
	return ok && pc.ID == string(catwalk.InferenceProviderAnthropic) && pc.OAuthToken != nil
}

// convertToToolResult converts a fantasy tool result to a message tool result.
func (a *sessionAgent) convertToToolResult(result fantasy.ToolResultContent) message.ToolResult {
	baseResult := message.ToolResult{
		ToolCallID: result.ToolCallID,
		Name:       result.ToolName,
		Metadata:   result.ClientMetadata,
	}

	switch result.Result.GetType() {
	case fantasy.ToolResultContentTypeText:
		if r, ok := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentText](result.Result); ok {
			baseResult.Content = r.Text
		}
	case fantasy.ToolResultContentTypeError:
		if r, ok := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentError](result.Result); ok {
			baseResult.Content = r.Error.Error()
			baseResult.IsError = true
		}
	case fantasy.ToolResultContentTypeMedia:
		if r, ok := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentMedia](result.Result); ok {
			content := r.Text
			if content == "" {
				content = fmt.Sprintf("Loaded %s content", r.MediaType)
			}
			baseResult.Content = content
			baseResult.Data = r.Data
			baseResult.MIMEType = r.MediaType
		}
	}

	return baseResult
}

// workaroundProviderMediaLimitations converts media content in tool results to
// user messages for providers that don't natively support images in tool results.
//
// Problem: OpenAI, Google, OpenRouter, and other OpenAI-compatible providers
// don't support sending images/media in tool result messages - they only accept
// text in tool results. However, they DO support images in user messages.
//
// If we send media in tool results to these providers, the API returns an error.
//
// Solution: For these providers, we:
//  1. Replace the media in the tool result with a text placeholder
//  2. Inject a user message immediately after with the image as a file attachment
//  3. This maintains the tool execution flow while working around API limitations
//
// Anthropic and Bedrock support images natively in tool results, so we skip
// this workaround for them.
//
// Example transformation:
//
//	BEFORE: [tool result: image data]
//	AFTER:  [tool result: "Image loaded - see attached"], [user: image attachment]
func (a *sessionAgent) workaroundProviderMediaLimitations(messages []fantasy.Message) []fantasy.Message {
	providerSupportsMedia := a.largeModel.ModelCfg.Provider == string(catwalk.InferenceProviderAnthropic) ||
		a.largeModel.ModelCfg.Provider == string(catwalk.InferenceProviderBedrock)

	if providerSupportsMedia {
		return messages
	}

	convertedMessages := make([]fantasy.Message, 0, len(messages))

	for _, msg := range messages {
		if msg.Role != fantasy.MessageRoleTool {
			convertedMessages = append(convertedMessages, msg)
			continue
		}

		textParts := make([]fantasy.MessagePart, 0, len(msg.Content))
		var mediaFiles []fantasy.FilePart

		for _, part := range msg.Content {
			toolResult, ok := fantasy.AsMessagePart[fantasy.ToolResultPart](part)
			if !ok {
				textParts = append(textParts, part)
				continue
			}

			if media, ok := fantasy.AsToolResultOutputType[fantasy.ToolResultOutputContentMedia](toolResult.Output); ok {
				decoded, err := base64.StdEncoding.DecodeString(media.Data)
				if err != nil {
					slog.Warn("failed to decode media data", "error", err)
					textParts = append(textParts, part)
					continue
				}

				mediaFiles = append(mediaFiles, fantasy.FilePart{
					Data:      decoded,
					MediaType: media.MediaType,
					Filename:  fmt.Sprintf("tool-result-%s", toolResult.ToolCallID),
				})

				textParts = append(textParts, fantasy.ToolResultPart{
					ToolCallID: toolResult.ToolCallID,
					Output: fantasy.ToolResultOutputContentText{
						Text: "[Image/media content loaded - see attached file]",
					},
					ProviderOptions: toolResult.ProviderOptions,
				})
			} else {
				textParts = append(textParts, part)
			}
		}

		convertedMessages = append(convertedMessages, fantasy.Message{
			Role:    fantasy.MessageRoleTool,
			Content: textParts,
		})

		if len(mediaFiles) > 0 {
			convertedMessages = append(convertedMessages, fantasy.NewUserMessage(
				"Here is the media content from the tool result:",
				mediaFiles...,
			))
		}
	}

	return convertedMessages
}
