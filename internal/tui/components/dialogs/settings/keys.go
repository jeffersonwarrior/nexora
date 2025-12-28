package settings

import "charm.land/bubbles/v2/key"

// KeyMap defines the key bindings for the settings dialog.
type KeyMap struct {
	Close  key.Binding
	Up     key.Binding
	Down   key.Binding
	Toggle key.Binding
}

// DefaultKeymap returns the default keybindings for the settings dialog.
func DefaultKeymap() KeyMap {
	return KeyMap{
		Close: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc", "close"),
		),
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("↑/k", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("↓/j", "down"),
		),
		Toggle: key.NewBinding(
			key.WithKeys("enter", " "),
			key.WithHelp("enter/space", "toggle"),
		),
	}
}
