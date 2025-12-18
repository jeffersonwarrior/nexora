package models

import (
	"fmt"
	"net/url"
	"strings"

	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/nexora/nexora/internal/config/providers"
	"github.com/nexora/nexora/internal/tui/components/dialogs"
	"github.com/nexora/nexora/internal/tui/styles"
)

// View type alias for compatibility
type View = tea.View

// cursorPosition represents cursor coordinates for focus management
type cursorPosition struct {
	x int
	y int
}

const (
	LocalModelsStepIP = iota
	LocalModelsStepKey
	LocalModelsStepModels
)

type LocalModelsDialog struct {
	width         int
	height        int
	step          int
	endpointInput textinput.Model
	apiKey        string
	serverType    string
	models        []providers.LocalModel
	selectedModel int
	detecting     bool
	error         string
	spinner       spinner.Model
}

// NewLocalModelsDialog creates a new local models dialog
func NewLocalModelsDialog() *LocalModelsDialog {
	s := spinner.New()
	s.Spinner = spinner.Dot

	t := styles.CurrentTheme()
	ti := textinput.New()
	ti.Placeholder = "Type or paste endpoint URL..."
	ti.Prompt = ""
	ti.SetVirtualCursor(false)
	ti.SetStyles(t.S().TextInput)
	ti.Focus()

	return &LocalModelsDialog{
		step:          LocalModelsStepIP,
		endpointInput: ti,
		models:        make([]providers.LocalModel, 0),
		spinner:       s,
	}
}

// ID returns the dialog ID
func (m *LocalModelsDialog) ID() dialogs.DialogID {
	return "local-models"
}

// Init initializes the dialog
func (m *LocalModelsDialog) Init() tea.Cmd {
	return m.spinner.Tick
}

// Update handles messages
func (m *LocalModelsDialog) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	// Handle spinner
	if m.detecting {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		if cmd != nil {
			return m, cmd
		}
	}

	var cmd tea.Cmd

	// Handle text input for endpoint
	if m.step == LocalModelsStepIP {
		m.endpointInput, cmd = m.endpointInput.Update(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "escape":
			return m, nil // Close dialog
		case "enter":
			return m.handleEnter()
		}
	case LocalModelsDetectComplete:
		m.detecting = false
		m.error = ""
		m.models = msg.Models

		// Move to next step
		if len(msg.Models) > 0 {
			m.step = LocalModelsStepModels
		} else {
			m.error = "No models found at this endpoint"
		}
	case LocalModelsDetectErrorMsg:
		m.detecting = false
		m.error = msg.Error
	}

	return m, cmd
}

func (m *LocalModelsDialog) handleEnter() (tea.Model, tea.Cmd) {
	switch m.step {
	case LocalModelsStepIP:
		endpoint := m.endpointInput.Value()
		if len(endpoint) > 0 {
			m.detecting = true
			m.error = ""
			return m, m.detectModels()
		}
	case LocalModelsStepKey:
		m.step = LocalModelsStepModels
	case LocalModelsStepModels:
		if len(m.models) > 0 && m.selectedModel < len(m.models) {
			// Return selected model
			return m, func() tea.Msg {
				return LocalModelsDetectComplete{
					Endpoint:   m.endpointInput.Value(),
					APIKey:     m.apiKey,
					ServerType: m.serverType,
					Models:     m.models,
				}
			}
		}
	}
	return m, nil
}

func (m *LocalModelsDialog) handleBackspace() (tea.Model, tea.Cmd) {
	switch m.step {
	case LocalModelsStepIP:
		// Handled by textinput.Update()
	case LocalModelsStepKey:
		if len(m.apiKey) > 0 {
			m.apiKey = m.apiKey[:len(m.apiKey)-1]
		}
	}
	return m, nil
}

// cleanEndpoint strips path components and normalizes the endpoint URL
// Returns cleaned endpoint and extracted API key
func cleanEndpoint(endpoint string) (string, string) {
	endpoint = strings.TrimSpace(endpoint)
	apiKey := ""

	// Extract API key from query params (support both ?key= and ?api_key=)
	if strings.Contains(endpoint, "?") {
		u, err := url.Parse(endpoint)
		if err == nil {
			// Try api_key first, then key
			if apiKeyVal := u.Query().Get("api_key"); apiKeyVal != "" {
				apiKey = apiKeyVal
			} else if keyVal := u.Query().Get("key"); keyVal != "" {
				apiKey = keyVal
			}
			// Remove query params from endpoint if we found a key
			if apiKey != "" {
				endpoint = u.Scheme + "://" + u.Host + u.Path
			}
		}
	}

	// Add https:// if not present (prefer https for external URLs)
	if !strings.HasPrefix(endpoint, "http://") && !strings.HasPrefix(endpoint, "https://") {
		// Check if it's a localhost/127.0.0.1 address
		if strings.HasPrefix(endpoint, "localhost") || strings.HasPrefix(endpoint, "127.0.0.1") || strings.HasPrefix(endpoint, "192.168.") {
			endpoint = "http://" + endpoint
		} else {
			endpoint = "https://" + endpoint
		}
	}

	// Parse URL to strip unwanted path components
	parsedURL, err := url.Parse(endpoint)
	if err != nil {
		// If parsing fails, return as-is
		return endpoint, apiKey
	}

	// Strip common API path suffixes like /v1/chat/completions, /api, etc.
	// Keep only scheme, host, and port
	cleanedURL := &url.URL{
		Scheme: parsedURL.Scheme,
		Host:   parsedURL.Host,
	}

	return cleanedURL.String(), apiKey
}

// detectModels starts the model detection process
// detectModels starts the model detection process
func (m *LocalModelsDialog) detectModels() tea.Cmd {
	return func() tea.Msg {
		// Clean and normalize endpoint, extract API key
		endpoint, extractedKey := cleanEndpoint(m.endpointInput.Value())

		// Use extracted key if we don't have one already
		apiKey := m.apiKey
		if apiKey == "" && extractedKey != "" {
			apiKey = extractedKey
		}

		detector := providers.NewLocalDetector(endpoint)
		provider, err := detector.Detect("", apiKey)
		if err != nil {
			return LocalModelsDetectErrorMsg{Error: fmt.Sprintf("Detection failed: %v\n\nMake sure a local model server (Ollama, LM-Studio, vLLM) is running at %s", err, endpoint)}
		}

		// Use models from provider
		return LocalModelsDetectComplete{
			Endpoint:   endpoint,
			APIKey:     apiKey,
			ServerType: provider.Type,
			Models:     provider.Models,
		}
	}
}

// Cursor returns the cursor position for focus management
func (m *LocalModelsDialog) Cursor() *tea.Cursor {
	return tea.NewCursor(0, 0)
}

// View renders the dialog
func (m *LocalModelsDialog) View() tea.View {
	t := styles.CurrentTheme()

	var content string
	switch m.step {
	case LocalModelsStepIP:
		content = m.renderIPInput()
	case LocalModelsStepKey:
		content = m.renderKeyInput()
	case LocalModelsStepModels:
		content = m.renderModelSelection()
	}

	if m.detecting {
		content += "\n\n" + m.spinner.View() + " Detecting models..."
	}

	if m.error != "" {
		content += "\n\n" + lipgloss.NewStyle().Foreground(t.Cherry).Render(m.error)
	}

	title := lipgloss.NewStyle().Foreground(t.Primary).Bold(true).Render("Configure Local Models")
	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.Border).
		Padding(1, 2).
		Width(m.width).
		Height(m.height)

	view := tea.NewView(border.Render(title + "\n\n" + content))
	return view
}

// renderIPInput shows endpoint input with API key hint
func (m *LocalModelsDialog) renderIPInput() string {
	t := styles.CurrentTheme()
	return lipgloss.NewStyle().Foreground(t.Primary).Render("Enter server endpoint:") + "\n\n" +
		m.endpointInput.View() + "\n\n" +
		lipgloss.NewStyle().Foreground(t.FgMuted).Render("Examples:\n  127.0.0.1:11434 (Ollama)\n  192.168.1.10:1234 (LM-Studio)\n  https://your-tunnel.ngrok-free.dev?api_key=sk-... (with API key)\n\nNow supports paste! Ctrl+V in terminal OR right-click.\nEnter to detect, Escape to cancel")
}

func (m *LocalModelsDialog) renderKeyInput() string {
	t := styles.CurrentTheme()

	return lipgloss.NewStyle().Foreground(t.Primary).Render("Enter API key (if required):") + "\n\n" +
		lipgloss.NewStyle().Foreground(t.FgBase).Render(m.apiKey+"_")
}

func (m *LocalModelsDialog) renderModelSelection() string {
	t := styles.CurrentTheme()

	if len(m.models) == 0 {
		return lipgloss.NewStyle().Foreground(t.Cherry).Render("No models found")
	}

	content := lipgloss.NewStyle().Foreground(t.Primary).Render("Select a model:") + "\n\n"
	for i, model := range m.models {
		prefix := "  "
		if i == m.selectedModel {
			prefix = lipgloss.NewStyle().Foreground(t.Accent).Render("> ")
		}
		content += prefix + lipgloss.NewStyle().Foreground(t.FgBase).Render(model.ID) + "\n"
		if model.Context > 0 {
			content += "    " + lipgloss.NewStyle().Foreground(t.FgMuted).Render(fmt.Sprintf("Context: %d", model.Context)) + "\n"
		}
	}

	return content
}

// Message types
type (
	LocalModelsDetectComplete struct {
		Endpoint   string
		APIKey     string
		ServerType string
		Models     []providers.LocalModel
	}

	LocalModelsDetectErrorMsg struct {
		Error string
	}
)
