package agent

import (
	"bytes"
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"maps"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"charm.land/fantasy"
	"github.com/charmbracelet/catwalk/pkg/catwalk"
	"github.com/nexora/cli/internal/agent/prompt"
	"github.com/nexora/cli/internal/agent/tools"
	"github.com/nexora/cli/internal/aiops"
	"github.com/nexora/cli/internal/config"
	"github.com/nexora/cli/internal/csync"
	"github.com/nexora/cli/internal/history"
	"github.com/nexora/cli/internal/log"
	"github.com/nexora/cli/internal/lsp"
	"github.com/nexora/cli/internal/message"
	"github.com/nexora/cli/internal/permission"
	"github.com/nexora/cli/internal/session"
	"github.com/nexora/cli/internal/sessionlog"
	"golang.org/x/sync/errgroup"

	"charm.land/fantasy/providers/anthropic"
	"charm.land/fantasy/providers/azure"
	"charm.land/fantasy/providers/bedrock"
	"charm.land/fantasy/providers/google"
	"charm.land/fantasy/providers/openai"
	"charm.land/fantasy/providers/openaicompat"
	"charm.land/fantasy/providers/openrouter"
	openaisdk "github.com/openai/openai-go/v2/option"
	"github.com/qjebbs/go-jsons"
)

type Coordinator interface {
	// INFO: (kujtim) this is not used yet we will use this when we have multiple agents
	// SetMainAgent(string)
	Run(ctx context.Context, sessionID, prompt string, attachments ...message.Attachment) (*fantasy.AgentResult, error)
	Cancel(sessionID string)
	CancelAll()
	IsSessionBusy(sessionID string) bool
	IsBusy() bool
	QueuedPrompts(sessionID string) int
	ClearQueue(sessionID string)
	Summarize(context.Context, string) error
	Model() Model
	UpdateModels(ctx context.Context) error
}

type coordinator struct {
	cfg         *config.Config
	sessions    session.Service
	messages    message.Service
	permissions permission.Service
	history     history.Service
	lspClients  *csync.Map[string, *lsp.Client]
	aiops       aiops.Ops
	sessionLog  *sessionlog.Manager

	currentAgent SessionAgent
	agents       map[string]SessionAgent

	readyWg errgroup.Group
}

func NewCoordinator(
	ctx context.Context,
	cfg *config.Config,
	sessions session.Service,
	messages message.Service,
	permissions permission.Service,
	history history.Service,
	lspClients *csync.Map[string, *lsp.Client],
	aiops aiops.Ops,
	sessionLog *sessionlog.Manager,
) (Coordinator, error) {
	c := &coordinator{
		cfg:         cfg,
		sessions:    sessions,
		messages:    messages,
		permissions: permissions,
		history:     history,
		lspClients:  lspClients,
		aiops:       aiops,
		sessionLog:  sessionLog,
		agents:      make(map[string]SessionAgent),
	}

	agentCfg, ok := cfg.Agents[config.AgentCoder]
	if !ok {
		return nil, errors.New("coder agent not configured")
	}

	// TODO: make this dynamic when we support multiple agents
	prompt, err := coderPrompt(prompt.WithWorkingDir(c.cfg.WorkingDir()))
	if err != nil {
		return nil, err
	}

	agent, err := c.buildAgent(ctx, prompt, agentCfg)
	if err != nil {
		return nil, err
	}
	c.currentAgent = agent
	c.agents[config.AgentCoder] = agent
	return c, nil
}

// Run implements Coordinator.
func (c *coordinator) Run(ctx context.Context, sessionID string, prompt string, attachments ...message.Attachment) (*fantasy.AgentResult, error) {
	if err := c.readyWg.Wait(); err != nil {
		return nil, err
	}

	model := c.currentAgent.Model()
	maxTokens := model.CatwalkCfg.DefaultMaxTokens
	if model.ModelCfg.MaxTokens != 0 {
		maxTokens = model.ModelCfg.MaxTokens
	}

	// Ensure maxTokens is at least 1 to satisfy API requirements
	// Anthropic and other providers require max_tokens >= 1
	if maxTokens == 0 {
		// Fallback to a reasonable default based on context window
		if model.CatwalkCfg.ContextWindow > 100000 {
			maxTokens = 8000 // For large context models
		} else {
			maxTokens = 4096 // Standard default
		}
		slog.Debug("MaxTokens was 0, using fallback",
			"model", model.ModelCfg.Model,
			"fallback_max_tokens", maxTokens)
	}

	if !model.CatwalkCfg.SupportsImages && attachments != nil {
		attachments = nil
	}

	providerCfg, ok := c.cfg.Providers.Get(model.ModelCfg.Provider)
	if !ok {
		return nil, errors.New("model provider not configured")
	}

	mergedOptions, temp, topP, topK, freqPenalty, presPenalty := mergeCallOptions(model, providerCfg)

	if providerCfg.OAuthToken != nil && providerCfg.OAuthToken.IsExpired() {
		slog.Info("Detected expired OAuth token, attempting refresh", "provider", providerCfg.ID)
		if refreshErr := c.cfg.RefreshOAuthToken(ctx, providerCfg.ID); refreshErr != nil {
			slog.Error("Failed to refresh OAuth token", "provider", providerCfg.ID, "error", refreshErr)
			return nil, refreshErr
		}

		// Rebuild models with refreshed token
		if updateErr := c.UpdateModels(ctx); updateErr != nil {
			slog.Error("Failed to update models after token refresh", "error", updateErr)
			return nil, updateErr
		}
	}
	result, err := c.currentAgent.Run(ctx, SessionAgentCall{
		SessionID:        sessionID,
		Prompt:           prompt,
		Attachments:      attachments,
		MaxOutputTokens:  maxTokens,
		ProviderOptions:  mergedOptions,
		Temperature:      temp,
		TopP:             topP,
		TopK:             topK,
		FrequencyPenalty: freqPenalty,
		PresencePenalty:  presPenalty,
	})
	return result, err
}

func getProviderOptions(model Model, providerCfg config.ProviderConfig) fantasy.ProviderOptions {
	options := fantasy.ProviderOptions{}

	cfgOpts := []byte("{}")
	providerCfgOpts := []byte("{}")
	catwalkOpts := []byte("{}")

	if model.ModelCfg.ProviderOptions != nil {
		data, err := json.Marshal(model.ModelCfg.ProviderOptions)
		if err == nil {
			cfgOpts = data
		}
	}

	if providerCfg.ProviderOptions != nil {
		data, err := json.Marshal(providerCfg.ProviderOptions)
		if err == nil {
			providerCfgOpts = data
		}
	}

	if model.CatwalkCfg.Options.ProviderOptions != nil {
		data, err := json.Marshal(model.CatwalkCfg.Options.ProviderOptions)
		if err == nil {
			catwalkOpts = data
		}
	}

	readers := []io.Reader{
		bytes.NewReader(catwalkOpts),
		bytes.NewReader(providerCfgOpts),
		bytes.NewReader(cfgOpts),
	}

	got, err := jsons.Merge(readers)
	if err != nil {
		slog.Error("Could not merge call config", "err", err)
		return options
	}

	mergedOptions := make(map[string]any)

	err = json.Unmarshal([]byte(got), &mergedOptions)
	if err != nil {
		slog.Error("Could not create config for call", "err", err)
		return options
	}

	switch providerCfg.Type {
	case openai.Name, azure.Name:
		_, hasReasoningEffort := mergedOptions["reasoning_effort"]
		if !hasReasoningEffort && model.ModelCfg.ReasoningEffort != "" {
			mergedOptions["reasoning_effort"] = model.ModelCfg.ReasoningEffort
		}
		if openai.IsResponsesModel(model.CatwalkCfg.ID) {
			if openai.IsResponsesReasoningModel(model.CatwalkCfg.ID) {
				mergedOptions["reasoning_summary"] = "auto"
				mergedOptions["include"] = []openai.IncludeType{openai.IncludeReasoningEncryptedContent}
			}
			parsed, err := openai.ParseResponsesOptions(mergedOptions)
			if err == nil {
				options[openai.Name] = parsed
			}
		} else {
			parsed, err := openai.ParseOptions(mergedOptions)
			if err == nil {
				options[openai.Name] = parsed
			}
		}
	case anthropic.Name:
		_, hasThink := mergedOptions["thinking"]
		if !hasThink && model.ModelCfg.Think {
			mergedOptions["thinking"] = map[string]any{
				// TODO: kujtim see if we need to make this dynamic
				"budget_tokens": 2000,
			}
		}
		parsed, err := anthropic.ParseOptions(mergedOptions)
		if err == nil {
			options[anthropic.Name] = parsed
		}

	case openrouter.Name:
		_, hasReasoning := mergedOptions["reasoning"]
		if !hasReasoning && model.ModelCfg.ReasoningEffort != "" {
			mergedOptions["reasoning"] = map[string]any{
				"enabled": true,
				"effort":  model.ModelCfg.ReasoningEffort,
			}
		}
		parsed, err := openrouter.ParseOptions(mergedOptions)
		if err == nil {
			options[openrouter.Name] = parsed
		}
	case google.Name:
		_, hasReasoning := mergedOptions["thinking_config"]
		if !hasReasoning {
			mergedOptions["thinking_config"] = map[string]any{
				"thinking_budget":  2000,
				"include_thoughts": true,
			}
		}
		parsed, err := google.ParseOptions(mergedOptions)
		if err == nil {
			options[google.Name] = parsed
		}
	case openaicompat.Name:
		_, hasReasoningEffort := mergedOptions["reasoning_effort"]
		if !hasReasoningEffort && model.ModelCfg.ReasoningEffort != "" {
			mergedOptions["reasoning_effort"] = model.ModelCfg.ReasoningEffort
		}
		parsed, err := openaicompat.ParseOptions(mergedOptions)
		if err == nil {
			options[openaicompat.Name] = parsed
		}
	}

	return options
}

func mergeCallOptions(model Model, cfg config.ProviderConfig) (fantasy.ProviderOptions, *float64, *float64, *int64, *float64, *float64) {
	modelOptions := getProviderOptions(model, cfg)
	temp := cmp.Or(model.ModelCfg.Temperature, model.CatwalkCfg.Options.Temperature)
	topP := cmp.Or(model.ModelCfg.TopP, model.CatwalkCfg.Options.TopP)
	topK := cmp.Or(model.ModelCfg.TopK, model.CatwalkCfg.Options.TopK)
	freqPenalty := cmp.Or(model.ModelCfg.FrequencyPenalty, model.CatwalkCfg.Options.FrequencyPenalty)
	presPenalty := cmp.Or(model.ModelCfg.PresencePenalty, model.CatwalkCfg.Options.PresencePenalty)
	return modelOptions, temp, topP, topK, freqPenalty, presPenalty
}

func (c *coordinator) buildAgent(ctx context.Context, prompt *prompt.Prompt, agent config.Agent) (SessionAgent, error) {
	large, small, err := c.buildAgentModels(ctx)
	if err != nil {
		return nil, err
	}

	systemPrompt, err := prompt.Build(ctx, large.Model.Provider(), large.Model.Model(), *c.cfg)
	if err != nil {
		return nil, err
	}

	largeProviderCfg, _ := c.cfg.Providers.Get(large.ModelCfg.Provider)
	result := NewSessionAgent(SessionAgentOptions{
		LargeModel:           large,
		SmallModel:           small,
		SystemPromptPrefix:   largeProviderCfg.SystemPromptPrefix,
		SystemPrompt:         systemPrompt,
		DisableAutoSummarize: c.cfg.Options.DisableAutoSummarize,
		IsYolo:               c.permissions.SkipRequests(),
		Sessions:             c.sessions,
		Messages:             c.messages,
		Tools:                nil,
		AIOPS:                c.aiops,
	})
	c.readyWg.Go(func() error {
		tools, err := c.buildTools(ctx, agent)
		if err != nil {
			return err
		}
		result.SetTools(tools)
		return nil
	})

	return result, nil
}

func (c *coordinator) buildTools(ctx context.Context, agent config.Agent) ([]fantasy.AgentTool, error) {
	var allTools []fantasy.AgentTool
	if slices.Contains(agent.AllowedTools, AgentToolName) {
		agentTool, err := c.agentTool(ctx)
		if err != nil {
			return nil, err
		}
		allTools = append(allTools, agentTool)
	}

	if slices.Contains(agent.AllowedTools, tools.AgenticFetchToolName) {
		agenticFetchTool, err := c.agenticFetchTool(ctx, nil)
		if err != nil {
			return nil, err
		}
		allTools = append(allTools, agenticFetchTool)
	}

	// Get the model name for the agent
	modelName := ""
	if modelCfg, ok := c.cfg.Models[agent.Model]; ok {
		if model := c.cfg.GetModel(modelCfg.Provider, modelCfg.Model); model != nil {
			modelName = model.Name
		}
	}

	allTools = append(allTools,
		tools.NewBashTool(c.permissions, c.cfg.WorkingDir(), c.cfg.Options.Attribution, modelName),
		tools.NewJobOutputTool(),
		tools.NewJobKillTool(),
		tools.NewDownloadTool(c.permissions, c.cfg.WorkingDir(), nil),
		tools.NewEditTool(c.lspClients, c.permissions, c.history, c.cfg.WorkingDir(), c.aiops),
		tools.NewMultiEditTool(c.lspClients, c.permissions, c.history, c.cfg.WorkingDir(), c.aiops),
		tools.NewFetchTool(c.permissions, c.cfg.WorkingDir(), nil),
		tools.NewGlobTool(c.cfg.WorkingDir()),
		tools.NewGrepTool(c.cfg.WorkingDir()),
		tools.NewLsTool(c.permissions, c.cfg.WorkingDir(), c.cfg.Tools.Ls),
		tools.NewSourcegraphTool(nil),
		tools.NewViewTool(c.lspClients, c.permissions, c.cfg.WorkingDir()),
		tools.NewWriteTool(c.lspClients, c.permissions, c.history, c.cfg.WorkingDir()),
	)

	if len(c.cfg.LSP) > 0 {
		allTools = append(allTools, tools.NewDiagnosticsTool(c.lspClients), tools.NewReferencesTool(c.lspClients))
	}

	var filteredTools []fantasy.AgentTool
	for _, tool := range allTools {
		if slices.Contains(agent.AllowedTools, tool.Info().Name) {
			filteredTools = append(filteredTools, tool)
		}
	}

	for _, tool := range tools.GetMCPTools(c.permissions, c.cfg.WorkingDir()) {
		if agent.AllowedMCP == nil {
			// No MCP restrictions
			filteredTools = append(filteredTools, tool)
			continue
		}
		if len(agent.AllowedMCP) == 0 {
			// No MCPs allowed
			slog.Debug("no MCPs allowed", "tool", tool.Name(), "agent", agent.Name)
			break
		}

		for mcp, tools := range agent.AllowedMCP {
			if mcp != tool.MCP() {
				continue
			}
			if len(tools) == 0 || slices.Contains(tools, tool.MCPToolName()) {
				filteredTools = append(filteredTools, tool)
			}
		}
		slog.Debug("MCP not allowed", "tool", tool.Name(), "agent", agent.Name)
	}
	slices.SortFunc(filteredTools, func(a, b fantasy.AgentTool) int {
		return strings.Compare(a.Info().Name, b.Info().Name)
	})
	return filteredTools, nil
}

// SequentialEditSolver reorders multiedit operations by line numbers to prevent conflicts
// When multiple edits target the same file, this ensures they're applied in correct order
type SequentialEditSolver struct{}

// NewSequentialEditSolver creates a new sequential edit solver
func NewSequentialEditSolver() *SequentialEditSolver {
	return &SequentialEditSolver{}
}

// EditLocation represents line position of an edit operation
type EditLocation struct {
	StartLine int
	EndLine   int
	FilePath  string
}

// ExtractEditLocation extracts line numbers from old_string to determine edit position
func (s *SequentialEditSolver) ExtractEditLocation(oldString string) EditLocation {
	lines := strings.Split(oldString, "\n")
	startLine := 0
	endLine := 0

	// Try to extract line numbers from context strings (if available)
	// Format: "file.go:line" or "file.go:start-end"
	re := regexp.MustCompile(`:(\d+)(?:-(\d+))?`)
	matches := re.FindStringSubmatch(oldString)
	if len(matches) >= 2 {
		startLine, _ = strconv.Atoi(matches[1])
		if len(matches) >= 3 && matches[2] != "" {
			endLine, _ = strconv.Atoi(matches[2])
		} else {
			endLine = startLine
		}
	} else {
		// If no line numbers found, estimate based on content
		startLine = 1 // Default to start
		endLine = len(lines)
	}

	return EditLocation{
		StartLine: startLine,
		EndLine:   endLine,
		FilePath:  "", // Will be filled by caller
	}
}

// OrderEdits sorts multiedit operations by line number to prevent conflicts
func (s *SequentialEditSolver) OrderEdits(edits []tools.EditParams, filePath string) []tools.EditParams {
	if len(edits) <= 1 {
		return edits
	}

	// Create sortable edits with location info
	type sortableEdit struct {
		Edit     tools.EditParams
		Location EditLocation
	}

	sortableEdits := make([]sortableEdit, len(edits))
	for i, edit := range edits {
		location := s.ExtractEditLocation(edit.OldString)
		location.FilePath = filePath
		sortableEdits[i] = sortableEdit{
			Edit:     edit,
			Location: location,
		}
	}

	// Sort by line number (ascending)
	slices.SortFunc(sortableEdits, func(a, b sortableEdit) int {
		if a.Location.StartLine != b.Location.StartLine {
			return cmp.Compare(a.Location.StartLine, b.Location.StartLine)
		}
		return cmp.Compare(a.Location.EndLine, b.Location.EndLine)
	})

	// Extract ordered edits
	orderedEdits := make([]tools.EditParams, len(edits))
	for i, sortable := range sortableEdits {
		orderedEdits[i] = sortable.Edit
	}

	return orderedEdits
}

// TODO: when we support multiple agents we need to change this so that we pass in the agent specific model config
// TODO: enhance agent execution by implementing: 1) Execution-first prompting that prioritizes tool usage over explanations, 2) Incremental execution pipeline with immediate tool calls after analysis, 3) Self-correction loops that retry failed actions instead of stopping, 4) Tool-chain orchestration for multi-step operations
func (c *coordinator) buildAgentModels(ctx context.Context) (Model, Model, error) {
	largeModelCfg, ok := c.cfg.Models[config.SelectedModelTypeLarge]
	if !ok {
		return Model{}, Model{}, errors.New("large model not selected")
	}
	smallModelCfg, ok := c.cfg.Models[config.SelectedModelTypeSmall]
	if !ok {
		return Model{}, Model{}, errors.New("small model not selected")
	}

	largeProviderCfg, ok := c.cfg.Providers.Get(largeModelCfg.Provider)
	if !ok {
		return Model{}, Model{}, errors.New("large model provider not configured")
	}

	largeProvider, err := c.buildProvider(largeProviderCfg, largeModelCfg)
	if err != nil {
		return Model{}, Model{}, err
	}

	smallProviderCfg, ok := c.cfg.Providers.Get(smallModelCfg.Provider)
	if !ok {
		return Model{}, Model{}, errors.New("small model provider not configured")
	}

	smallProvider, err := c.buildProvider(smallProviderCfg, smallModelCfg)
	if err != nil {
		return Model{}, Model{}, err
	}

	var largeCatwalkModel *catwalk.Model
	var smallCatwalkModel *catwalk.Model

	for _, m := range largeProviderCfg.Models {
		if m.ID == largeModelCfg.Model {
			largeCatwalkModel = &m
		}
	}
	for _, m := range smallProviderCfg.Models {
		if m.ID == smallModelCfg.Model {
			smallCatwalkModel = &m
		}
	}

	// Fall back to catwalk known providers if not found in provider config
	knownProviders, err := config.Providers(c.cfg)
	if err == nil {
		if largeCatwalkModel == nil {
			for _, p := range knownProviders {
				if string(p.ID) == largeModelCfg.Provider {
					for i, m := range p.Models {
						if m.ID == largeModelCfg.Model {
							largeCatwalkModel = &p.Models[i]
						}
					}
					break
				}
			}
		}
		if smallCatwalkModel == nil {
			for _, p := range knownProviders {
				if string(p.ID) == smallModelCfg.Provider {
					for i, m := range p.Models {
						if m.ID == smallModelCfg.Model {
							smallCatwalkModel = &p.Models[i]
						}
					}
					break
				}
			}
		}
	}

	if largeCatwalkModel == nil {
		return Model{}, Model{}, fmt.Errorf("large model %s not found for provider %s", largeModelCfg.Model, largeModelCfg.Provider)
	}

	if smallCatwalkModel == nil {
		return Model{}, Model{}, fmt.Errorf("small model %s not found for provider %s", smallModelCfg.Model, smallModelCfg.Provider)
	}

	largeModelID := largeModelCfg.Model
	smallModelID := smallModelCfg.Model

	if largeModelCfg.Provider == openrouter.Name && isExactoSupported(largeModelID) {
		largeModelID += ":exacto"
	}

	if smallModelCfg.Provider == openrouter.Name && isExactoSupported(smallModelID) {
		smallModelID += ":exacto"
	}

	// Legacy GPT-OSS-120B mappings removed - DeepSeek Coder 2 is default
	// Special handling for Cerebras no longer needed

	largeModel, err := largeProvider.LanguageModel(ctx, largeModelID)
	if err != nil {
		return Model{}, Model{}, err
	}
	smallModel, err := smallProvider.LanguageModel(ctx, smallModelID)
	if err != nil {
		return Model{}, Model{}, err
	}

	return Model{
			Model:      largeModel,
			CatwalkCfg: *largeCatwalkModel,
			ModelCfg:   largeModelCfg,
		}, Model{
			Model:      smallModel,
			CatwalkCfg: *smallCatwalkModel,
			ModelCfg:   smallModelCfg,
		}, nil
}

func (c *coordinator) buildAnthropicProvider(baseURL, apiKey string, headers map[string]string) (fantasy.Provider, error) {
	var opts []anthropic.Option

	// MiniMax uses JWT tokens
	if strings.HasPrefix(apiKey, "Bearer ") {
		// NOTE: Prevent the SDK from picking up the API key from env.
		os.Setenv("ANTHROPIC_API_KEY", "")

		headers["Authorization"] = apiKey
	} else if strings.HasPrefix(apiKey, "eyJ") { // JWT token (MiniMax)
		// NOTE: Prevent the SDK from picking up the API key from env.
		os.Setenv("ANTHROPIC_API_KEY", "")

		// Use the JWT as the API key and set Authorization
		opts = append(opts, anthropic.WithAPIKey(apiKey))
		headers["Authorization"] = "Bearer " + apiKey
	} else if apiKey != "" {
		// X-Api-Key header
		opts = append(opts, anthropic.WithAPIKey(apiKey))
	}

	if len(headers) > 0 {
		opts = append(opts, anthropic.WithHeaders(headers))
	}

	if baseURL != "" {
		opts = append(opts, anthropic.WithBaseURL(baseURL))
	}

	if c.cfg.Options.Debug {
		httpClient := log.NewHTTPClient()
		opts = append(opts, anthropic.WithHTTPClient(httpClient))
	}

	return anthropic.New(opts...)
}

func (c *coordinator) buildOpenaiProvider(baseURL, apiKey string, headers map[string]string) (fantasy.Provider, error) {
	opts := []openai.Option{
		openai.WithAPIKey(apiKey),
		openai.WithUseResponsesAPI(),
	}
	if c.cfg.Options.Debug {
		httpClient := log.NewHTTPClient()
		opts = append(opts, openai.WithHTTPClient(httpClient))
	}
	if len(headers) > 0 {
		opts = append(opts, openai.WithHeaders(headers))
	}
	if baseURL != "" {
		opts = append(opts, openai.WithBaseURL(baseURL))
	}
	return openai.New(opts...)
}

func (c *coordinator) buildOpenrouterProvider(_, apiKey string, headers map[string]string) (fantasy.Provider, error) {
	opts := []openrouter.Option{
		openrouter.WithAPIKey(apiKey),
	}
	if c.cfg.Options.Debug {
		httpClient := log.NewHTTPClient()
		opts = append(opts, openrouter.WithHTTPClient(httpClient))
	}
	if len(headers) > 0 {
		opts = append(opts, openrouter.WithHeaders(headers))
	}
	return openrouter.New(opts...)
}

func (c *coordinator) buildMinimaxProvider(baseURL, apiKey string) (fantasy.Provider, error) {
	// Custom MiniMax provider using anthropic SDK with JWT as API key
	opts := []anthropic.Option{
		anthropic.WithAPIKey(apiKey),
	}
	if c.cfg.Options.Debug {
		httpClient := log.NewHTTPClient()
		opts = append(opts, anthropic.WithHTTPClient(httpClient))
	}
	opts = append(opts, anthropic.WithBaseURL(baseURL))

	return anthropic.New(opts...)
}

func (c *coordinator) buildOpenaiCompatProvider(baseURL, apiKey string, headers map[string]string, extraBody map[string]any) (fantasy.Provider, error) {
	opts := []openaicompat.Option{
		openaicompat.WithBaseURL(baseURL),
		openaicompat.WithAPIKey(apiKey),
	}
	if c.cfg.Options.Debug {
		httpClient := log.NewHTTPClient()
		opts = append(opts, openaicompat.WithHTTPClient(httpClient))
	}
	if len(headers) > 0 {
		opts = append(opts, openaicompat.WithHeaders(headers))
	}

	for extraKey, extraValue := range extraBody {
		opts = append(opts, openaicompat.WithSDKOptions(openaisdk.WithJSONSet(extraKey, extraValue)))
	}

	return openaicompat.New(opts...)
}

func (c *coordinator) buildAzureProvider(baseURL, apiKey string, headers map[string]string, options map[string]string) (fantasy.Provider, error) {
	opts := []azure.Option{
		azure.WithBaseURL(baseURL),
		azure.WithAPIKey(apiKey),
		azure.WithUseResponsesAPI(),
	}
	if c.cfg.Options.Debug {
		httpClient := log.NewHTTPClient()
		opts = append(opts, azure.WithHTTPClient(httpClient))
	}
	if options == nil {
		options = make(map[string]string)
	}
	if apiVersion, ok := options["apiVersion"]; ok {
		opts = append(opts, azure.WithAPIVersion(apiVersion))
	}
	if len(headers) > 0 {
		opts = append(opts, azure.WithHeaders(headers))
	}

	return azure.New(opts...)
}

func (c *coordinator) buildBedrockProvider(headers map[string]string) (fantasy.Provider, error) {
	var opts []bedrock.Option
	if c.cfg.Options.Debug {
		httpClient := log.NewHTTPClient()
		opts = append(opts, bedrock.WithHTTPClient(httpClient))
	}
	if len(headers) > 0 {
		opts = append(opts, bedrock.WithHeaders(headers))
	}
	bearerToken := os.Getenv("AWS_BEARER_TOKEN_BEDROCK")
	if bearerToken != "" {
		opts = append(opts, bedrock.WithAPIKey(bearerToken))
	}
	return bedrock.New(opts...)
}

func (c *coordinator) buildGoogleProvider(baseURL, apiKey string, headers map[string]string) (fantasy.Provider, error) {
	opts := []google.Option{
		google.WithBaseURL(baseURL),
		google.WithGeminiAPIKey(apiKey),
	}
	if c.cfg.Options.Debug {
		httpClient := log.NewHTTPClient()
		opts = append(opts, google.WithHTTPClient(httpClient))
	}
	if len(headers) > 0 {
		opts = append(opts, google.WithHeaders(headers))
	}
	return google.New(opts...)
}

func (c *coordinator) buildGoogleVertexProvider(headers map[string]string, options map[string]string) (fantasy.Provider, error) {
	opts := []google.Option{}
	if c.cfg.Options.Debug {
		httpClient := log.NewHTTPClient()
		opts = append(opts, google.WithHTTPClient(httpClient))
	}
	if len(headers) > 0 {
		opts = append(opts, google.WithHeaders(headers))
	}

	project := options["project"]
	location := options["location"]

	opts = append(opts, google.WithVertex(project, location))

	return google.New(opts...)
}

func (c *coordinator) isAnthropicThinking(model config.SelectedModel) bool {
	if model.Think {
		return true
	}

	if model.ProviderOptions == nil {
		return false
	}

	opts, err := anthropic.ParseOptions(model.ProviderOptions)
	if err != nil {
		return false
	}
	if opts.Thinking != nil {
		return true
	}
	return false
}

func (c *coordinator) buildProvider(providerCfg config.ProviderConfig, model config.SelectedModel) (fantasy.Provider, error) {
	headers := maps.Clone(providerCfg.ExtraHeaders)
	if headers == nil {
		headers = make(map[string]string)
	}

	// handle special headers for anthropic
	if providerCfg.Type == anthropic.Name && c.isAnthropicThinking(model) {
		if v, ok := headers["anthropic-beta"]; ok {
			headers["anthropic-beta"] = v + ",interleaved-thinking-2025-05-14"
		} else {
			headers["anthropic-beta"] = "interleaved-thinking-2025-05-14"
		}
	}

	apiKey, _ := c.cfg.Resolve(providerCfg.APIKey)
	baseURL, _ := c.cfg.Resolve(providerCfg.BaseURL)

	switch providerCfg.Type {
	case openai.Name:
		return c.buildOpenaiProvider(baseURL, apiKey, headers)
	case anthropic.Name:
		if providerCfg.ID == "minimax" {
			return c.buildMinimaxProvider(baseURL, apiKey)
		}
		return c.buildAnthropicProvider(baseURL, apiKey, headers)
	case openrouter.Name:
		return c.buildOpenrouterProvider(baseURL, apiKey, headers)
	case azure.Name:
		return c.buildAzureProvider(baseURL, apiKey, headers, providerCfg.ExtraParams)
	case bedrock.Name:
		return c.buildBedrockProvider(headers)
	case google.Name:
		return c.buildGoogleProvider(baseURL, apiKey, headers)
	case "google-vertex":
		return c.buildGoogleVertexProvider(headers, providerCfg.ExtraParams)
	case openaicompat.Name:
		if providerCfg.ID == string(catwalk.InferenceProviderZAI) {
			if providerCfg.ExtraBody == nil {
				providerCfg.ExtraBody = map[string]any{}
			}
			providerCfg.ExtraBody["tool_stream"] = true
		}
		return c.buildOpenaiCompatProvider(baseURL, apiKey, headers, providerCfg.ExtraBody)
	default:
		return nil, fmt.Errorf("provider type not supported: %q", providerCfg.Type)
	}
}

func isExactoSupported(modelID string) bool {
	supportedModels := []string{
		"moonshotai/kimi-k2-0905",
		"deepseek/deepseek-v3.1-terminus",
		"z-ai/glm-4.6",
		"openai/gpt-oss-120b",
		"qwen/qwen3-coder",
		"deepseek-coder-2", // Cerebras DeepSeek default
	}
	return slices.Contains(supportedModels, modelID)
}

func (c *coordinator) Cancel(sessionID string) {
	c.currentAgent.Cancel(sessionID)
}

func (c *coordinator) CancelAll() {
	c.currentAgent.CancelAll()
}

func (c *coordinator) ClearQueue(sessionID string) {
	c.currentAgent.ClearQueue(sessionID)
}

func (c *coordinator) IsBusy() bool {
	return c.currentAgent.IsBusy()
}

func (c *coordinator) IsSessionBusy(sessionID string) bool {
	return c.currentAgent.IsSessionBusy(sessionID)
}

func (c *coordinator) Model() Model {
	return c.currentAgent.Model()
}

func (c *coordinator) UpdateModels(ctx context.Context) error {
	// build the models again so we make sure we get the latest config
	large, small, err := c.buildAgentModels(ctx)
	if err != nil {
		return err
	}
	c.currentAgent.SetModels(large, small)

	agentCfg, ok := c.cfg.Agents[config.AgentCoder]
	if !ok {
		return errors.New("coder agent not configured")
	}

	tools, err := c.buildTools(ctx, agentCfg)
	if err != nil {
		return err
	}
	c.currentAgent.SetTools(tools)
	return nil
}

func (c *coordinator) QueuedPrompts(sessionID string) int {
	return c.currentAgent.QueuedPrompts(sessionID)
}

func (c *coordinator) Summarize(ctx context.Context, sessionID string) error {
	providerCfg, ok := c.cfg.Providers.Get(c.currentAgent.Model().ModelCfg.Provider)
	if !ok {
		return errors.New("model provider not configured")
	}
	return c.currentAgent.Summarize(ctx, sessionID, getProviderOptions(c.currentAgent.Model(), providerCfg))
}
