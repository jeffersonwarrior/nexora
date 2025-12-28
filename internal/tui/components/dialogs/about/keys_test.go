package about

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

func TestDefaultKeymap(t *testing.T) {
	keymap := DefaultKeymap()

	// Verify Close key binding has the expected keys
	expectedKeys := []string{"esc", "enter", "q", " "}

	for _, expectedKey := range expectedKeys {
		found := false
		for _, binding := range keymap.Close.Keys() {
			if binding == expectedKey {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected Close binding to contain key %q", expectedKey)
		}
	}
}

func TestKeyMapStructure(t *testing.T) {
	keymap := DefaultKeymap()

	// Verify the KeyMap has the expected field
	if keymap.Close.Keys() == nil {
		t.Error("expected Close binding to be initialized")
	}

	// Verify key binding is valid
	testKey := key.NewBinding(key.WithKeys("esc"))
	testMsg := tea.KeyPressMsg(tea.Key{Code: tea.KeyEscape})
	if key.Matches(testMsg, testKey) == false {
		t.Error("key binding validation failed")
	}
}
