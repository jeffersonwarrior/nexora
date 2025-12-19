//go:build ignore
// +build ignore

// TODO: This file depends on MultiSessionCoordinator which is incomplete
// Re-enable when multi-session support is ready

package chat

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/bubbles/list"
	"github.com/nexora/nexora/internal/agent"
	"github.com/nexora/nexora/internal/tui/exp/list"
)

type WindowsSwitcher struct {
	list        list.List[agent.SessionWindow]
	showing     bool
	coordinator *agent.MultiSessionCoordinator
	onSwitch    func(string)
}

func NewWindowsSwitcher(coordinator *agent.MultiSessionCoordinator, onSwitch func(string)) *WindowsSwitcher {
	return &WindowsSwitcher{
		coordinator: coordinator,
		onSwitch:    onSwitch,
	}
}

func (w *WindowsSwitcher) Show() {
	w.showing = true
	w.updateList()
}

func (w *WindowsSwitcher) Hide() {
	w.showing = false
}

func (w *WindowsSwitcher) updateList() {
	windows := w.coordinator.ListWindows()
	items := make([]list.Item, len(windows))
	for i, win := range windows {
		items[i] = listItemFromWindow(win)
	}
	w.list.SetItems(items)
}

func (w *WindowsSwitcher) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	w.list, cmd = w.list.Update(msg)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if item := w.list.SelectedItem(); item != nil {
				win := item.(listItem)
				w.onSwitch(win.ID)
				w.Hide()
				return w, nil
			}
		case "esc", "ctrl+c":
			w.Hide()
		case "ctrl+tab":
			w.list.Next()
		case "ctrl+shift+tab":
			w.list.Prev()
		}
	}
	return w, cmd
}

type listItem listItemImpl

type listItemImpl struct {
	agent.SessionWindow
}

func (i listItemImpl) FilterValue() string {
	return i.Title
}

func listItemFromWindow(w agent.SessionWindow) list.Item {
	return listItemImpl{w}
}

func (i listItemImpl) Title() string {
	return i.Title
}

func (i listItemImpl) Description() string {
	status := "idle"
	if i.IsActive {
		status = "ACTIVE"
	}
	return fmt.Sprintf("%s (%d msgs)", status, len(i.History))
}
