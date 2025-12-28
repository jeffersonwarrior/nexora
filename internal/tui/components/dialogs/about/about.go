package about

import (
	"runtime"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/nexora/nexora/internal/tui/components/dialogs"
	"github.com/nexora/nexora/internal/tui/styles"
	"github.com/nexora/nexora/internal/tui/util"
	"github.com/nexora/nexora/internal/version"
)

const (
	AboutDialogID dialogs.DialogID = "about"
)

// AboutDialog represents the about information dialog.
type AboutDialog interface {
	dialogs.DialogModel
}

type aboutDialogCmp struct {
	wWidth  int
	wHeight int
	keymap  KeyMap
}

// NewAboutDialog creates a new about dialog.
func NewAboutDialog() AboutDialog {
	return &aboutDialogCmp{
		keymap: DefaultKeymap(),
	}
}

func (a *aboutDialogCmp) Init() tea.Cmd {
	return nil
}

func (a *aboutDialogCmp) Update(msg tea.Msg) (util.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.wWidth = msg.Width
		a.wHeight = msg.Height
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, a.keymap.Close):
			return a, util.CmdHandler(dialogs.CloseDialogMsg{})
		}
	}
	return a, nil
}

func (a *aboutDialogCmp) View() string {
	t := styles.CurrentTheme()
	baseStyle := t.S().Base

	title := baseStyle.Bold(true).Foreground(t.Primary).Render("NEXORA")
	ver := baseStyle.Foreground(t.FgMuted).Render(version.Display())
	titleLine := lipgloss.JoinHorizontal(lipgloss.Center, title, " ", ver)

	desc := baseStyle.Foreground(t.FgBase).Render("AI-native terminal application")
	platform := baseStyle.Foreground(t.FgHalfMuted).Render(runtime.GOOS + "/" + runtime.GOARCH)
	goVer := baseStyle.Foreground(t.FgHalfMuted).Render("Go " + runtime.Version()[2:])

	hint := baseStyle.Foreground(t.FgMuted).Italic(true).Render("Press any key to close")

	content := baseStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Center,
			titleLine,
			"",
			desc,
			"",
			lipgloss.JoinHorizontal(lipgloss.Center, platform, " • ", goVer),
			"",
			hint,
		),
	)

	aboutStyle := baseStyle.
		Padding(1, 3).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(t.BorderFocus)

	return aboutStyle.Render(content)
}

func (a *aboutDialogCmp) Position() (int, int) {
	contentWidth := 40
	contentHeight := 9 // approximate height with padding
	row := (a.wHeight - contentHeight) / 2
	col := (a.wWidth - contentWidth) / 2
	return row, col
}

func (a *aboutDialogCmp) ID() dialogs.DialogID {
	return AboutDialogID
}
