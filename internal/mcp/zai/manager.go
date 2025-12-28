package zai

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	"github.com/nexora/nexora/internal/config"
)

// Manager handles the lifecycle of Z.ai MCP clients
type Manager struct {
	client *Client
	config config.Config
	logger *slog.Logger
	mutex  sync.RWMutex
}

// NewManager creates a new Z.ai MCP manager
func NewManager(cfg config.Config) *Manager {
	return &Manager{
		config: cfg,
		logger: slog.Default().With("mcp", "zai-manager"),
	}
}

// Start initializes the Z.ai MCP client
func (m *Manager) Start(ctx context.Context) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.logger.Info("starting Z.ai MCP manager")

	// Check for Z.ai API key
	if apiKey := os.Getenv("ZAI_API_KEY"); apiKey == "" {
		m.logger.Info("ZAI_API_KEY not set, Z.ai vision tools will be unavailable")
		return fmt.Errorf("ZAI_API_KEY environment variable is required for Z.ai vision tools")
	}

	client, err := NewClient(m.config)
	if err != nil {
		return fmt.Errorf("failed to create Z.ai MCP client: %w", err)
	}

	if err := client.Connect(ctx); err != nil {
		return fmt.Errorf("failed to connect Z.ai MCP client: %w", err)
	}

	m.client = client
	m.logger.Info("Z.ai MCP manager started successfully")
	return nil
}

// GetClient returns the Z.ai MCP client
func (m *Manager) GetClient() (*Client, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.client == nil {
		return nil, fmt.Errorf("z.ai MCP client not initialized; call Start() first")
	}

	return m.client, nil
}

// IsStarted returns true if the Z.ai MCP manager is started
func (m *Manager) IsStarted() bool {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	return m.client != nil
}

// Stop closes the Z.ai MCP client
func (m *Manager) Stop() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if m.client != nil {
		m.logger.Info("stopping Z.ai MCP manager")
		if err := m.client.Close(); err != nil {
			m.logger.Error("error closing Z.ai MCP client", "error", err)
		}
		m.client = nil
		m.logger.Info("Z.ai MCP manager stopped")
	}

	return nil
}

// GetStatus returns the current status of the Z.ai MCP manager
func (m *Manager) GetStatus() Status {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	if m.client == nil {
		if os.Getenv("ZAI_API_KEY") == "" {
			return Status{State: StateNotConfigured, Message: "ZAI_API_KEY not set"}
		}
		return Status{State: StateStopped, Message: "Manager stopped"}
	}

	return Status{State: StateRunning, Message: "Z.ai vision tools available"}
}

// Status represents the status of the Z.ai MCP manager
type Status struct {
	State   State
	Message string
}

// State represents the state of the Z.ai MCP manager
type State int

const (
	StateNotConfigured State = iota
	StateStopped
	StateRunning
)

func (s State) String() string {
	switch s {
	case StateNotConfigured:
		return "not_configured"
	case StateStopped:
		return "stopped"
	case StateRunning:
		return "running"
	default:
		return "unknown"
	}
}

// ValidateConfig checks if the Z.ai configuration is valid
func ValidateConfig() error {
	apiKey := os.Getenv("ZAI_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("ZAI_API_KEY environment variable is required")
	}

	if len(apiKey) < 10 {
		return fmt.Errorf("ZAI_API_KEY appears to be too short (minimum 10 characters)")
	}

	return nil
}

// GetAvailableTools returns the list of available vision tools if the manager is running
func (m *Manager) GetAvailableTools() ([]string, error) {
	if !m.IsStarted() {
		return nil, fmt.Errorf("z.ai MCP manager not started")
	}

	client, err := m.GetClient()
	if err != nil {
		return nil, err
	}

	tools := client.GetTools()
	toolNames := make([]string, len(tools))
	for i, tool := range tools {
		toolNames[i] = tool.Name
	}

	return toolNames, nil
}
