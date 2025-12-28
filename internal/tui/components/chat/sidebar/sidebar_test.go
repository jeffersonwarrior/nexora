package sidebar

import (
	"context"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/nexora/nexora/internal/csync"
	"github.com/nexora/nexora/internal/history"
	"github.com/nexora/nexora/internal/lsp"
	"github.com/nexora/nexora/internal/pubsub"
	"github.com/nexora/nexora/internal/tui/components/chat"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name        string
		compactMode bool
	}{
		{
			name:        "creates sidebar in normal mode",
			compactMode: false,
		},
		{
			name:        "creates sidebar in compact mode",
			compactMode: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHistory := &mockHistoryService{}
			lspClients := csync.NewMap[string, *lsp.Client]()

			sidebar := New(mockHistory, lspClients, tt.compactMode)
			if sidebar == nil {
				t.Fatal("New() returned nil")
			}

			// Verify it implements the Sidebar interface
			// Type is already Sidebar from New(), no need to assert
		})
	}
}

func TestSidebarInit(t *testing.T) {
	mockHistory := &mockHistoryService{}
	lspClients := csync.NewMap[string, *lsp.Client]()
	sidebar := New(mockHistory, lspClients, false)

	cmd := sidebar.Init()
	// Init should return nil
	if cmd != nil {
		t.Error("Init() should return nil")
	}
}

func TestSidebarGetSize(t *testing.T) {
	mockHistory := &mockHistoryService{}
	lspClients := csync.NewMap[string, *lsp.Client]()
	sidebar := New(mockHistory, lspClients, false).(*sidebarCmp)

	// Initial size should be zero
	width, height := sidebar.GetSize()
	if width != 0 || height != 0 {
		t.Errorf("Initial size should be (0, 0), got (%d, %d)", width, height)
	}

	// Manually set size fields
	sidebar.width = 40
	sidebar.height = 30
	width, height = sidebar.GetSize()
	if width != 40 || height != 30 {
		t.Errorf("Expected size (40, 30), got (%d, %d)", width, height)
	}
}

// Note: Skipping View test as it requires full app dependencies (config, etc.)

func TestSidebarSetCompactMode(t *testing.T) {
	tests := []struct {
		name        string
		initialMode bool
		newMode     bool
	}{
		{"change from normal to compact", false, true},
		{"change from compact to normal", true, false},
		{"keep compact mode", true, true},
		{"keep normal mode", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHistory := &mockHistoryService{}
			lspClients := csync.NewMap[string, *lsp.Client]()
			sidebar := New(mockHistory, lspClients, tt.initialMode).(*sidebarCmp)

			sidebar.SetCompactMode(tt.newMode)

			if sidebar.compactMode != tt.newMode {
				t.Errorf("Expected compact mode %v, got %v", tt.newMode, sidebar.compactMode)
			}
		})
	}
}

func TestSidebarUpdate(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.Msg
	}{
		{
			name: "session cleared message",
			msg:  chat.SessionClearedMsg{},
		},
		{
			name: "session files message",
			msg: SessionFilesMsg{
				Files: []SessionFile{
					{FilePath: "test.go", Additions: 10, Deletions: 5},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHistory := &mockHistoryService{}
			lspClients := csync.NewMap[string, *lsp.Client]()
			sidebar := New(mockHistory, lspClients, false)

			// Update should not panic
			_, cmd := sidebar.Update(tt.msg)
			_ = cmd
		})
	}
}

func TestGetMaxWidth(t *testing.T) {
	tests := []struct {
		name     string
		width    int
		expected int
	}{
		{"width 40", 40, 38},  // 40 - 2 (padding)
		{"width 60", 60, 58},  // max is 58
		{"width 30", 30, 28},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHistory := &mockHistoryService{}
			lspClients := csync.NewMap[string, *lsp.Client]()
			sidebar := New(mockHistory, lspClients, false).(*sidebarCmp)

			sidebar.width = tt.width
			maxWidth := sidebar.getMaxWidth()

			if maxWidth != tt.expected {
				t.Errorf("Expected max width %d, got %d", tt.expected, maxWidth)
			}
		})
	}
}

func TestCalculateAvailableHeight(t *testing.T) {
	tests := []struct {
		name          string
		height        int
		compactMode   bool
		minExpected   int
	}{
		{"normal mode tall", 40, false, 20},
		{"normal mode short", 20, false, 5},
		{"compact mode", 30, true, 15},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHistory := &mockHistoryService{}
			lspClients := csync.NewMap[string, *lsp.Client]()
			sidebar := New(mockHistory, lspClients, tt.compactMode).(*sidebarCmp)

			sidebar.height = tt.height
			availableHeight := sidebar.calculateAvailableHeight()

			// Should be greater than minimum expected
			if availableHeight < tt.minExpected {
				t.Errorf("Expected available height >= %d, got %d", tt.minExpected, availableHeight)
			}
		})
	}
}

func TestFormatTokensAndCost(t *testing.T) {
	tests := []struct {
		name          string
		tokens        int64
		contextWindow int64
		cost          float64
		wantContains  []string
	}{
		{
			name:          "zero values",
			tokens:        0,
			contextWindow: 0,
			cost:          0,
			wantContains:  []string{"0"},
		},
		{
			name:          "with tokens and cost",
			tokens:        1000,
			contextWindow: 200000,
			cost:          0.05,
			wantContains:  []string{"1K", "$0.05"},
		},
		{
			name:          "large tokens",
			tokens:        150000,
			contextWindow: 200000,
			cost:          1.50,
			wantContains:  []string{"150K", "$1.50"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTokensAndCost(tt.tokens, tt.contextWindow, tt.cost)

			for _, want := range tt.wantContains {
				if !contains(result, want) {
					t.Errorf("formatTokensAndCost() result should contain '%s', got '%s'", want, result)
				}
			}
		})
	}
}

func TestGetDynamicLimits(t *testing.T) {
	tests := []struct {
		name        string
		height      int
		compactMode bool
	}{
		{"normal mode with tall height", 50, false},
		{"normal mode with short height", 20, false},
		{"compact mode", 30, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHistory := &mockHistoryService{}
			lspClients := csync.NewMap[string, *lsp.Client]()
			sidebar := New(mockHistory, lspClients, tt.compactMode).(*sidebarCmp)

			sidebar.height = tt.height
			maxFiles, maxLSPs, maxMCPs := sidebar.getDynamicLimits()

			// All limits should be at least the minimum
			if maxFiles < MinItemsPerSection {
				t.Errorf("maxFiles should be >= %d, got %d", MinItemsPerSection, maxFiles)
			}
			if maxLSPs < MinItemsPerSection {
				t.Errorf("maxLSPs should be >= %d, got %d", MinItemsPerSection, maxLSPs)
			}
			if maxMCPs < MinItemsPerSection {
				t.Errorf("maxMCPs should be >= %d, got %d", MinItemsPerSection, maxMCPs)
			}
		})
	}
}

// Mock history service for testing
type mockHistoryService struct{}

func (m *mockHistoryService) Subscribe(ctx context.Context) <-chan pubsub.Event[history.File] {
	ch := make(chan pubsub.Event[history.File])
	close(ch)
	return ch
}

func (m *mockHistoryService) Create(ctx context.Context, sessionID, path, content string) (history.File, error) {
	return history.File{}, nil
}

func (m *mockHistoryService) CreateVersion(ctx context.Context, sessionID, path, content string) (history.File, error) {
	return history.File{}, nil
}

func (m *mockHistoryService) Get(ctx context.Context, id string) (history.File, error) {
	return history.File{}, nil
}

func (m *mockHistoryService) GetByPathAndSession(ctx context.Context, path, sessionID string) (history.File, error) {
	return history.File{}, nil
}

func (m *mockHistoryService) ListBySession(ctx context.Context, sessionID string) ([]history.File, error) {
	return []history.File{}, nil
}

func (m *mockHistoryService) ListLatestSessionFiles(ctx context.Context, sessionID string) ([]history.File, error) {
	return []history.File{}, nil
}

func (m *mockHistoryService) Delete(ctx context.Context, id string) error {
	return nil
}

func (m *mockHistoryService) DeleteSessionFiles(ctx context.Context, sessionID string) error {
	return nil
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return len(s) > 0 && (s == substr || len(substr) == 0 ||
		(len(s) >= len(substr) && (s[:len(substr)] == substr ||
		s[len(s)-len(substr):] == substr ||
		findSubstring(s, substr))))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
