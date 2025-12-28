package about

import "charm.land/bubbles/v2/key"

// KeyMap defines the key bindings for the about dialog.
type KeyMap struct {
	Close key.Binding
}

// DefaultKeymap returns the default keybindings for the about dialog.
func DefaultKeymap() KeyMap {
	return KeyMap{
		Close: key.NewBinding(
			key.WithKeys("esc", "enter", "q", " "),
			key.WithHelp("esc", "close"),
		),
	}
}
