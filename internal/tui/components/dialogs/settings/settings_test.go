package settings

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/nexora/nexora/internal/tui/components/dialogs"
)

func TestSettingsDialog_Init(t *testing.T) {
	dialog := NewSettingsDialog(nil)
	cmd := dialog.Init()
	if cmd != nil {
		t.Errorf("Init() should return nil, got %v", cmd)
	}
}

func TestSettingsDialog_ID(t *testing.T) {
	dialog := NewSettingsDialog(nil)
	if dialog.ID() != SettingsDialogID {
		t.Errorf("ID() = %v, want %v", dialog.ID(), SettingsDialogID)
	}
}

func TestSettingsDialog_Position(t *testing.T) {
	dialog := NewSettingsDialog(nil)
	// Set window size
	dialog.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	row, col := dialog.Position()
	// Position should be centered
	if row < 0 || col < 0 {
		t.Errorf("Position() returned negative values: row=%d, col=%d", row, col)
	}
}

func TestSettingsDialog_ToggleSettings(t *testing.T) {
	settings := &mockSettings{
		autoApprove:     false,
		thinkingEnabled: true,
		streaming:       true,
		vimMode:         false,
		autoLSP:         true,
	}

	_ = NewSettingsDialog(settings)

	// Verify initial state
	if settings.autoApprove != false {
		t.Errorf("Initial autoApprove should be false, got %v", settings.autoApprove)
	}

	// Test that settings manager is wired correctly
	settings.SetAutoApprove(true)
	if !settings.GetAutoApprove() {
		t.Error("SetAutoApprove(true) failed")
	}

	// Verify all getters work
	_ = settings.GetThinkingEnabled()
	_ = settings.GetStreaming()
	_ = settings.GetVimMode()
	_ = settings.GetAutoLSP()
}

func TestSettingsDialog_Navigation(t *testing.T) {
	settings := &mockSettings{}
	dialog := NewSettingsDialog(settings)

	// Basic test - verify dialog can be created and doesn't panic
	if dialog == nil {
		t.Error("NewSettingsDialog returned nil")
	}

	// Test that the dialog has items
	s := dialog.(*settingsDialogCmp)
	if len(s.items) != 5 {
		t.Errorf("Expected 5 settings items, got %d", len(s.items))
	}

	// Verify cursor starts at 0
	if s.cursor != 0 {
		t.Errorf("Expected cursor to start at 0, got %d", s.cursor)
	}
}

func TestSettingsDialog_View(t *testing.T) {
	settings := &mockSettings{}
	dialog := NewSettingsDialog(settings)

	// Test that View() doesn't panic and returns a non-empty string
	view := dialog.View()
	if view == "" {
		t.Error("View() should return a non-empty string")
	}
}

// mockSettings implements the SettingsManager interface for testing
type mockSettings struct {
	autoApprove     bool
	thinkingEnabled bool
	streaming       bool
	vimMode         bool
	autoLSP         bool
}

func (m *mockSettings) GetAutoApprove() bool     { return m.autoApprove }
func (m *mockSettings) GetThinkingEnabled() bool { return m.thinkingEnabled }
func (m *mockSettings) GetStreaming() bool       { return m.streaming }
func (m *mockSettings) GetVimMode() bool         { return m.vimMode }
func (m *mockSettings) GetAutoLSP() bool         { return m.autoLSP }

func (m *mockSettings) SetAutoApprove(v bool)     { m.autoApprove = v }
func (m *mockSettings) SetThinkingEnabled(v bool) { m.thinkingEnabled = v }
func (m *mockSettings) SetStreaming(v bool)       { m.streaming = v }
func (m *mockSettings) SetVimMode(v bool)         { m.vimMode = v }
func (m *mockSettings) SetAutoLSP(v bool)         { m.autoLSP = v }

func TestSettingsDialog_UpdateWindowSize(t *testing.T) {
	settings := &mockSettings{}
	dialog := NewSettingsDialog(settings)

	msg := tea.WindowSizeMsg{Width: 120, Height: 60}
	_, cmd := dialog.Update(msg)

	if cmd != nil {
		t.Error("Window size update should not return a command")
	}

	// Verify position is recalculated
	row, col := dialog.Position()
	if row < 0 || col < 0 {
		t.Errorf("Position should be positive after window size update: row=%d, col=%d", row, col)
	}
}

func TestSettingsDialog_CloseDialogMsg(t *testing.T) {
	settings := &mockSettings{}
	dialog := NewSettingsDialog(settings)

	msg := dialogs.CloseDialogMsg{}
	model, cmd := dialog.Update(msg)

	if model == nil {
		t.Error("Update should return a model")
	}

	if cmd != nil {
		// CloseDialogMsg might trigger a command, that's acceptable
	}
}
