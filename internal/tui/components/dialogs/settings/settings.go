package settings

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/nexora/nexora/internal/tui/components/dialogs"
	"github.com/nexora/nexora/internal/tui/styles"
	"github.com/nexora/nexora/internal/tui/util"
)

const (
	SettingsDialogID dialogs.DialogID = "settings"
)

// OpenSettingsMsg is sent to open the settings dialog
type OpenSettingsMsg struct{}

// SettingsManager defines the interface for managing settings
type SettingsManager interface {
	GetAutoApprove() bool
	GetThinkingEnabled() bool
	GetStreaming() bool
	GetVimMode() bool
	GetAutoLSP() bool

	SetAutoApprove(bool)
	SetThinkingEnabled(bool)
	SetStreaming(bool)
	SetVimMode(bool)
	SetAutoLSP(bool)
}

// SettingsDialog represents the settings configuration dialog
type SettingsDialog interface {
	dialogs.DialogModel
}

type settingItem struct {
	label   string
	key     string
	getter  func() bool
	setter  func(bool)
	enabled bool
}

type settingsDialogCmp struct {
	wWidth   int
	wHeight  int
	keymap   KeyMap
	settings SettingsManager
	items    []settingItem
	cursor   int
}

// NewSettingsDialog creates a new settings dialog
func NewSettingsDialog(settings SettingsManager) SettingsDialog {
	s := &settingsDialogCmp{
		keymap:   DefaultKeymap(),
		settings: settings,
		cursor:   0,
	}

	// Initialize items
	if settings != nil {
		s.items = []settingItem{
			{
				label:   "Auto Approve",
				key:     "auto_approve",
				getter:  settings.GetAutoApprove,
				setter:  settings.SetAutoApprove,
				enabled: settings.GetAutoApprove(),
			},
			{
				label:   "Thinking Mode",
				key:     "thinking_enabled",
				getter:  settings.GetThinkingEnabled,
				setter:  settings.SetThinkingEnabled,
				enabled: settings.GetThinkingEnabled(),
			},
			{
				label:   "Streaming",
				key:     "streaming",
				getter:  settings.GetStreaming,
				setter:  settings.SetStreaming,
				enabled: settings.GetStreaming(),
			},
			{
				label:   "Vim Mode",
				key:     "vim_mode",
				getter:  settings.GetVimMode,
				setter:  settings.SetVimMode,
				enabled: settings.GetVimMode(),
			},
			{
				label:   "Auto LSP",
				key:     "auto_lsp",
				getter:  settings.GetAutoLSP,
				setter:  settings.SetAutoLSP,
				enabled: settings.GetAutoLSP(),
			},
		}
	} else {
		// For testing without settings manager
		s.items = []settingItem{
			{label: "Auto Approve", key: "auto_approve", enabled: false},
			{label: "Thinking Mode", key: "thinking_enabled", enabled: true},
			{label: "Streaming", key: "streaming", enabled: true},
			{label: "Vim Mode", key: "vim_mode", enabled: false},
			{label: "Auto LSP", key: "auto_lsp", enabled: true},
		}
	}

	return s
}

func (s *settingsDialogCmp) Init() tea.Cmd {
	return nil
}

func (s *settingsDialogCmp) Update(msg tea.Msg) (util.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		s.wWidth = msg.Width
		s.wHeight = msg.Height
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, s.keymap.Close):
			return s, util.CmdHandler(dialogs.CloseDialogMsg{})
		case key.Matches(msg, s.keymap.Up):
			if s.cursor > 0 {
				s.cursor--
			}
		case key.Matches(msg, s.keymap.Down):
			if s.cursor < len(s.items)-1 {
				s.cursor++
			}
		case key.Matches(msg, s.keymap.Toggle):
			// Toggle the current item
			if s.cursor >= 0 && s.cursor < len(s.items) {
				item := &s.items[s.cursor]
				item.enabled = !item.enabled
				if item.setter != nil {
					item.setter(item.enabled)
				}
			}
		}
	}
	return s, nil
}

func (s *settingsDialogCmp) View() string {
	t := styles.CurrentTheme()
	baseStyle := t.S().Base

	title := baseStyle.Bold(true).Foreground(t.Primary).Render("Settings")

	// Render each setting item
	var items []string
	for i, item := range s.items {
		var itemStr string
		cursor := " "
		if i == s.cursor {
			cursor = "›"
		}

		toggle := "[ ]"
		if item.enabled {
			toggle = "[✓]"
		}

		labelStyle := baseStyle.Foreground(t.FgBase)
		toggleStyle := baseStyle.Foreground(t.FgMuted)

		if i == s.cursor {
			labelStyle = labelStyle.Foreground(t.Primary).Bold(true)
			toggleStyle = toggleStyle.Foreground(t.Primary)
		}

		itemStr = lipgloss.JoinHorizontal(
			lipgloss.Left,
			baseStyle.Foreground(t.Primary).Render(cursor),
			" ",
			toggleStyle.Render(toggle),
			" ",
			labelStyle.Render(item.label),
		)
		items = append(items, itemStr)
	}

	hint := baseStyle.Foreground(t.FgMuted).Italic(true).Render("↑/↓: navigate  enter/space: toggle  esc: close")

	content := baseStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			"",
			lipgloss.JoinVertical(lipgloss.Left, items...),
			"",
			hint,
		),
	)

	settingsStyle := baseStyle.
		Padding(1, 3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderFocus)

	return settingsStyle.Render(content)
}

func (s *settingsDialogCmp) Position() (int, int) {
	contentWidth := 50
	contentHeight := 12 // title + 5 items + spacing + hint
	row := (s.wHeight - contentHeight) / 2
	col := (s.wWidth - contentWidth) / 2
	return row, col
}

func (s *settingsDialogCmp) ID() dialogs.DialogID {
	return SettingsDialogID
}
