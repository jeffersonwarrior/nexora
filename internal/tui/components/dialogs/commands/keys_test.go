package commands

import (
	"testing"
)

func TestDefaultCommandsDialogKeyMap(t *testing.T) {
	keymap := DefaultCommandsDialogKeyMap()

	// Verify all key bindings are initialized
	if keymap.Select.Keys() == nil {
		t.Error("Select binding not initialized")
	}
	if keymap.Next.Keys() == nil {
		t.Error("Next binding not initialized")
	}
	if keymap.Previous.Keys() == nil {
		t.Error("Previous binding not initialized")
	}
	if keymap.Tab.Keys() == nil {
		t.Error("Tab binding not initialized")
	}
	if keymap.Close.Keys() == nil {
		t.Error("Close binding not initialized")
	}
}

func TestCommandsDialogKeyMap_KeyBindings(t *testing.T) {
	keymap := DefaultCommandsDialogKeyMap()
	bindings := keymap.KeyBindings()

	expectedCount := 5 // Select, Next, Previous, Tab, Close
	if len(bindings) != expectedCount {
		t.Errorf("expected %d key bindings, got %d", expectedCount, len(bindings))
	}
}

func TestCommandsDialogKeyMap_ShortHelp(t *testing.T) {
	keymap := DefaultCommandsDialogKeyMap()
	help := keymap.ShortHelp()

	if len(help) == 0 {
		t.Error("ShortHelp should not be empty")
	}

	// Should have 4 entries: Tab, choose, Select, Close
	if len(help) != 4 {
		t.Errorf("expected 4 short help entries, got %d", len(help))
	}
}

func TestCommandsDialogKeyMap_FullHelp(t *testing.T) {
	keymap := DefaultCommandsDialogKeyMap()
	fullHelp := keymap.FullHelp()

	if len(fullHelp) == 0 {
		t.Error("FullHelp should not be empty")
	}

	// Count total bindings across all rows
	total := 0
	for _, row := range fullHelp {
		total += len(row)
	}

	if total != 5 {
		t.Errorf("expected 5 total bindings in FullHelp, got %d", total)
	}
}

func TestDefaultArgumentsDialogKeyMap(t *testing.T) {
	keymap := DefaultArgumentsDialogKeyMap()

	// Verify all key bindings are initialized
	if keymap.Confirm.Keys() == nil {
		t.Error("Confirm binding not initialized")
	}
	if keymap.Next.Keys() == nil {
		t.Error("Next binding not initialized")
	}
	if keymap.Previous.Keys() == nil {
		t.Error("Previous binding not initialized")
	}
	if keymap.Close.Keys() == nil {
		t.Error("Close binding not initialized")
	}
}

func TestArgumentsDialogKeyMap_KeyBindings(t *testing.T) {
	keymap := DefaultArgumentsDialogKeyMap()
	bindings := keymap.KeyBindings()

	expectedCount := 4 // Confirm, Next, Previous, Close
	if len(bindings) != expectedCount {
		t.Errorf("expected %d key bindings, got %d", expectedCount, len(bindings))
	}
}

func TestArgumentsDialogKeyMap_ShortHelp(t *testing.T) {
	keymap := DefaultArgumentsDialogKeyMap()
	help := keymap.ShortHelp()

	if len(help) == 0 {
		t.Error("ShortHelp should not be empty")
	}

	if len(help) != 4 {
		t.Errorf("expected 4 short help entries, got %d", len(help))
	}
}

func TestArgumentsDialogKeyMap_FullHelp(t *testing.T) {
	keymap := DefaultArgumentsDialogKeyMap()
	fullHelp := keymap.FullHelp()

	if len(fullHelp) == 0 {
		t.Error("FullHelp should not be empty")
	}

	// Count total bindings across all rows
	total := 0
	for _, row := range fullHelp {
		total += len(row)
	}

	if total != 4 {
		t.Errorf("expected 4 total bindings in FullHelp, got %d", total)
	}
}
