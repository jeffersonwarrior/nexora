package layout

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

type Help interface {
	Bindings() []key.Binding
}

type Positional interface {
	SetPosition(x, y int) tea.Cmd
}
