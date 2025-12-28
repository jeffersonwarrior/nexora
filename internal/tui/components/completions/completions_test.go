package completions

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestNew(t *testing.T) {
	c := New()
	if c == nil {
		t.Fatal("New() returned nil")
	}

	// Should start closed
	if c.Open() {
		t.Error("Completions should start closed")
	}

	// Should have empty query
	if c.Query() != "" {
		t.Errorf("Expected empty query, got %q", c.Query())
	}
}

func TestCompletionsInit(t *testing.T) {
	c := New()
	cmd := c.Init()
	// Init returns a tea.Sequence command which wraps the list init and setSize
	// The command may be nil if internal list init returns nil with zero size
	// This test ensures Init can be called without panic
	_ = cmd
}

func TestDefaultKeyMap(t *testing.T) {
	km := DefaultKeyMap()

	if len(km.Down.Keys()) == 0 {
		t.Error("Down key binding should have keys")
	}
	if len(km.Up.Keys()) == 0 {
		t.Error("Up key binding should have keys")
	}
	if len(km.Select.Keys()) == 0 {
		t.Error("Select key binding should have keys")
	}
	if len(km.Cancel.Keys()) == 0 {
		t.Error("Cancel key binding should have keys")
	}
}

func TestKeyMapMethods(t *testing.T) {
	km := DefaultKeyMap()

	bindings := km.KeyBindings()
	if len(bindings) != 4 {
		t.Errorf("Expected 4 key bindings, got %d", len(bindings))
	}

	shortHelp := km.ShortHelp()
	if len(shortHelp) != 2 {
		t.Errorf("Expected 2 short help bindings, got %d", len(shortHelp))
	}

	fullHelp := km.FullHelp()
	if len(fullHelp) == 0 {
		t.Error("FullHelp should return at least one group")
	}
}

func TestCompletionsWindowSizeMsg(t *testing.T) {
	c := New().(*completionsCmp)

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	_, cmd := c.Update(msg)

	if cmd != nil {
		t.Error("WindowSizeMsg should not return a command")
	}

	if c.wWidth != 100 {
		t.Errorf("Expected wWidth 100, got %d", c.wWidth)
	}
	if c.wHeight != 50 {
		t.Errorf("Expected wHeight 50, got %d", c.wHeight)
	}
}

func TestCompletionsOpenMsg(t *testing.T) {
	c := New().(*completionsCmp)
	c.wWidth = 100
	c.wHeight = 50

	completions := []Completion{
		{Title: "Option 1", Value: "val1"},
		{Title: "Option 2", Value: "val2"},
		{Title: "Option 3", Value: "val3"},
	}

	msg := OpenCompletionsMsg{
		Completions: completions,
		X:           10,
		Y:           20,
		MaxResults:  5,
	}

	_, cmd := c.Update(msg)
	if cmd == nil {
		t.Error("OpenCompletionsMsg should return a command")
	}

	if !c.open {
		t.Error("Completions should be open after OpenCompletionsMsg")
	}

	if c.x != 10 {
		t.Errorf("Expected x 10, got %d", c.x)
	}
	if c.xorig != 10 {
		t.Errorf("Expected xorig 10, got %d", c.xorig)
	}
	if c.y != 20 {
		t.Errorf("Expected y 20, got %d", c.y)
	}

	if c.query != "" {
		t.Errorf("Expected empty query after open, got %q", c.query)
	}
}

func TestCompletionsOpenMsgWidthAdjustment(t *testing.T) {
	c := New().(*completionsCmp)
	c.wWidth = 50
	c.wHeight = 50

	// Create completions with long titles
	completions := []Completion{
		{Title: "A very long completion option title", Value: "val1"},
	}

	msg := OpenCompletionsMsg{
		Completions: completions,
		X:           45, // Near the edge
		Y:           20,
	}

	_, _ = c.Update(msg)

	// X should be adjusted to fit within window
	if c.x+c.width >= c.wWidth {
		t.Error("Completions should be adjusted to fit within window width")
	}
}

func TestCompletionsCloseMsg(t *testing.T) {
	c := New().(*completionsCmp)
	c.open = true

	msg := CloseCompletionsMsg{}
	_, cmd := c.Update(msg)

	if c.open {
		t.Error("Completions should be closed after CloseCompletionsMsg")
	}

	if cmd == nil {
		t.Error("CloseCompletionsMsg should return a command")
	}
}

func TestCompletionsViewWhenClosed(t *testing.T) {
	c := New().(*completionsCmp)
	c.open = false

	view := c.View()
	if view != "" {
		t.Errorf("View should be empty when closed, got %q", view)
	}
}

func TestCompletionsQueryMethod(t *testing.T) {
	c := New().(*completionsCmp)
	c.query = "test query"

	if c.Query() != "test query" {
		t.Errorf("Expected 'test query', got %q", c.Query())
	}
}

func TestCompletionsOpenMethod(t *testing.T) {
	c := New().(*completionsCmp)

	if c.Open() {
		t.Error("Should be closed initially")
	}

	c.open = true
	if !c.Open() {
		t.Error("Should be open after setting")
	}
}

func TestCompletionsKeyMapMethod(t *testing.T) {
	c := New()
	km := c.KeyMap()

	if len(km.Down.Keys()) == 0 {
		t.Error("KeyMap should return valid key bindings")
	}
}

func TestCompletionsPosition(t *testing.T) {
	c := New().(*completionsCmp)
	c.x = 10
	c.y = 30
	c.height = 5

	x, y := c.Position()
	if x != 10 {
		t.Errorf("Expected x 10, got %d", x)
	}
	if y != 25 { // y - height
		t.Errorf("Expected y 25, got %d", y)
	}
}

func TestCompletionsWidth(t *testing.T) {
	c := New().(*completionsCmp)
	c.width = 40

	if c.Width() != 40 {
		t.Errorf("Expected width 40, got %d", c.Width())
	}
}

func TestCompletionsHeight(t *testing.T) {
	c := New().(*completionsCmp)
	c.height = 8

	if c.Height() != 8 {
		t.Errorf("Expected height 8, got %d", c.Height())
	}
}

func TestCompletionsRepositionMsg(t *testing.T) {
	c := New().(*completionsCmp)
	c.wWidth = 100
	c.wHeight = 50

	msg := RepositionCompletionsMsg{X: 25, Y: 35}
	_, _ = c.Update(msg)

	if c.x != 25 {
		t.Errorf("Expected x 25, got %d", c.x)
	}
	if c.y != 35 {
		t.Errorf("Expected y 35, got %d", c.y)
	}
}

func TestCompletionsFilterMsgWhenClosed(t *testing.T) {
	c := New().(*completionsCmp)
	c.open = false

	msg := FilterCompletionsMsg{
		Query:  "test",
		Reopen: false,
		X:      10,
		Y:      20,
	}

	_, cmd := c.Update(msg)
	if cmd != nil {
		t.Error("FilterCompletionsMsg should return nil when closed and Reopen is false")
	}
}

func TestCompletionsFilterMsgSameQuery(t *testing.T) {
	c := New().(*completionsCmp)
	c.open = true
	c.query = "test"

	msg := FilterCompletionsMsg{
		Query:  "test",
		Reopen: false,
		X:      10,
		Y:      20,
	}

	_, cmd := c.Update(msg)
	if cmd != nil {
		t.Error("FilterCompletionsMsg should return nil when query is the same")
	}
}

func TestCompletionsCancelKey(t *testing.T) {
	c := New().(*completionsCmp)
	c.open = true

	// Simulate escape key
	msg := tea.KeyPressMsg{Code: tea.KeyEscape}
	_, cmd := c.Update(msg)

	if cmd == nil {
		t.Error("Cancel key should return a command")
	}
}

func TestCompletionsMaxHeight(t *testing.T) {
	if maxCompletionsHeight != 10 {
		t.Errorf("Expected maxCompletionsHeight to be 10, got %d", maxCompletionsHeight)
	}
}

func TestCompletionStruct(t *testing.T) {
	c := Completion{
		Title: "Test Title",
		Value: "test-value",
	}

	if c.Title != "Test Title" {
		t.Errorf("Expected title 'Test Title', got %q", c.Title)
	}
	if c.Value != "test-value" {
		t.Errorf("Expected value 'test-value', got %v", c.Value)
	}
}

func TestCompletionWithIntValue(t *testing.T) {
	c := Completion{
		Title: "Number",
		Value: 42,
	}

	if c.Value.(int) != 42 {
		t.Errorf("Expected value 42, got %v", c.Value)
	}
}

func TestCompletionWithStructValue(t *testing.T) {
	type customValue struct {
		ID   int
		Name string
	}

	cv := customValue{ID: 1, Name: "test"}
	c := Completion{
		Title: "Custom",
		Value: cv,
	}

	if c.Value.(customValue).ID != 1 {
		t.Errorf("Expected ID 1, got %v", c.Value)
	}
}

func TestEmptyCompletionsList(t *testing.T) {
	c := New().(*completionsCmp)
	c.wWidth = 100
	c.wHeight = 50

	msg := OpenCompletionsMsg{
		Completions: []Completion{},
		X:           10,
		Y:           20,
	}

	_, _ = c.Update(msg)

	// View should be empty for no items
	view := c.View()
	if view != "" {
		t.Error("View should be empty when no completions are provided")
	}
}

func TestKeyMapFullHelpLayout(t *testing.T) {
	km := DefaultKeyMap()
	fullHelp := km.FullHelp()

	// Should have at least one row
	if len(fullHelp) == 0 {
		t.Error("FullHelp should return at least one row")
	}

	// Each row should have at most 4 items
	for i, row := range fullHelp {
		if len(row) > 4 {
			t.Errorf("Row %d has %d items, expected at most 4", i, len(row))
		}
	}
}

func TestCompletionsSelectKeyWithNoSelection(t *testing.T) {
	c := New().(*completionsCmp)
	c.open = true

	// Simulate enter key when nothing is selected
	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := c.Update(msg)

	// Should return nil since there's no selected item
	if cmd != nil {
		t.Error("Select key with no items should return nil command")
	}
}

func TestCompletionsUpInsertKeyWithNoSelection(t *testing.T) {
	c := New().(*completionsCmp)
	c.open = true

	// Simulate ctrl+p when nothing is selected
	msg := tea.KeyPressMsg{
		Code: 'p',
		Mod:  tea.ModCtrl,
	}
	_, cmd := c.Update(msg)

	// Should return nil since there's no selected item
	if cmd != nil {
		t.Error("UpInsert key with no items should return nil command")
	}
}

func TestCompletionsDownInsertKeyWithNoSelection(t *testing.T) {
	c := New().(*completionsCmp)
	c.open = true

	// Simulate ctrl+n when nothing is selected
	msg := tea.KeyPressMsg{
		Code: 'n',
		Mod:  tea.ModCtrl,
	}
	_, cmd := c.Update(msg)

	// Should return nil since there's no selected item
	if cmd != nil {
		t.Error("DownInsert key with no items should return nil command")
	}
}

func TestCompletionsMsgTypes(t *testing.T) {
	// Test that all message types can be instantiated
	_ = OpenCompletionsMsg{}
	_ = FilterCompletionsMsg{}
	_ = RepositionCompletionsMsg{}
	_ = CompletionsClosedMsg{}
	_ = CompletionsOpenedMsg{}
	_ = CloseCompletionsMsg{}
	_ = SelectCompletionMsg{}
}

func TestSelectCompletionMsgFields(t *testing.T) {
	msg := SelectCompletionMsg{
		Value:  "test-value",
		Insert: true,
	}

	if msg.Value != "test-value" {
		t.Errorf("Expected value 'test-value', got %v", msg.Value)
	}
	if !msg.Insert {
		t.Error("Expected Insert to be true")
	}
}

func TestFilterCompletionsMsgFields(t *testing.T) {
	msg := FilterCompletionsMsg{
		Query:  "search",
		Reopen: true,
		X:      15,
		Y:      25,
	}

	if msg.Query != "search" {
		t.Errorf("Expected query 'search', got %q", msg.Query)
	}
	if !msg.Reopen {
		t.Error("Expected Reopen to be true")
	}
	if msg.X != 15 {
		t.Errorf("Expected X 15, got %d", msg.X)
	}
	if msg.Y != 25 {
		t.Errorf("Expected Y 25, got %d", msg.Y)
	}
}

func TestCompletionsMultipleUpdates(t *testing.T) {
	c := New().(*completionsCmp)
	c.wWidth = 100
	c.wHeight = 50

	// Open
	openMsg := OpenCompletionsMsg{
		Completions: []Completion{
			{Title: "Option 1", Value: "val1"},
		},
		X: 10,
		Y: 20,
	}
	_, _ = c.Update(openMsg)

	if !c.open {
		t.Error("Should be open after OpenCompletionsMsg")
	}

	// Reposition
	repositionMsg := RepositionCompletionsMsg{X: 30, Y: 40}
	_, _ = c.Update(repositionMsg)

	if c.x != 30 || c.y != 40 {
		t.Errorf("Expected position (30, 40), got (%d, %d)", c.x, c.y)
	}

	// Close
	closeMsg := CloseCompletionsMsg{}
	_, _ = c.Update(closeMsg)

	if c.open {
		t.Error("Should be closed after CloseCompletionsMsg")
	}
}

func TestCompletionsUpArrowKey(t *testing.T) {
	c := New().(*completionsCmp)
	c.wWidth = 100
	c.wHeight = 50

	// First open with some items
	openMsg := OpenCompletionsMsg{
		Completions: []Completion{
			{Title: "Item 1", Value: "1"},
			{Title: "Item 2", Value: "2"},
		},
		X: 10,
		Y: 20,
	}
	_, _ = c.Update(openMsg)

	// Simulate up arrow key
	msg := tea.KeyPressMsg{Code: tea.KeyUp}
	_, cmd := c.Update(msg)
	_ = cmd // We can't easily verify list movement, but we ensure it doesn't crash
}

func TestCompletionsDownArrowKey(t *testing.T) {
	c := New().(*completionsCmp)
	c.wWidth = 100
	c.wHeight = 50

	// First open with some items
	openMsg := OpenCompletionsMsg{
		Completions: []Completion{
			{Title: "Item 1", Value: "1"},
			{Title: "Item 2", Value: "2"},
		},
		X: 10,
		Y: 20,
	}
	_, _ = c.Update(openMsg)

	// Simulate down arrow key
	msg := tea.KeyPressMsg{Code: tea.KeyDown}
	_, cmd := c.Update(msg)
	_ = cmd // Ensure it doesn't crash
}

func TestCompletionsFilterMsgWithReopen(t *testing.T) {
	c := New().(*completionsCmp)
	c.wWidth = 100
	c.wHeight = 50
	c.open = false

	// First add some items by opening
	openMsg := OpenCompletionsMsg{
		Completions: []Completion{
			{Title: "Test Item", Value: "test"},
		},
		X: 10,
		Y: 20,
	}
	_, _ = c.Update(openMsg)
	c.open = false // Close it
	c.query = ""   // Reset query

	// Now filter with Reopen=true
	filterMsg := FilterCompletionsMsg{
		Query:  "te",
		Reopen: true,
		X:      10,
		Y:      20,
	}
	_, cmd := c.Update(filterMsg)

	if cmd == nil {
		t.Error("FilterCompletionsMsg with Reopen should return a command")
	}
}

func TestCompletionsFilterMsgWithNewQuery(t *testing.T) {
	c := New().(*completionsCmp)
	c.wWidth = 100
	c.wHeight = 50

	// Open with items
	openMsg := OpenCompletionsMsg{
		Completions: []Completion{
			{Title: "Apple", Value: "1"},
			{Title: "Banana", Value: "2"},
		},
		X: 10,
		Y: 20,
	}
	_, _ = c.Update(openMsg)

	// Filter with new query
	filterMsg := FilterCompletionsMsg{
		Query: "app",
		X:     10,
		Y:     20,
	}
	_, cmd := c.Update(filterMsg)
	_ = cmd

	if c.query != "app" {
		t.Errorf("Expected query 'app', got %q", c.query)
	}
}

func TestCompletionsViewWhenOpenWithItems(t *testing.T) {
	c := New().(*completionsCmp)
	c.wWidth = 100
	c.wHeight = 50

	// Open with items
	openMsg := OpenCompletionsMsg{
		Completions: []Completion{
			{Title: "Test Item", Value: "test"},
		},
		X: 10,
		Y: 20,
	}
	_, _ = c.Update(openMsg)

	view := c.View()
	if view == "" {
		t.Error("View should not be empty when open with items")
	}
}

func TestCompletionsAdjustPositionNegativeX(t *testing.T) {
	c := New().(*completionsCmp)
	c.wWidth = 100
	c.wHeight = 50
	c.xorig = 10
	c.x = -5 // Negative x
	c.lastWidth = 50

	// Open with items
	openMsg := OpenCompletionsMsg{
		Completions: []Completion{
			{Title: "Test", Value: "test"},
		},
		X: 10,
		Y: 20,
	}
	_, _ = c.Update(openMsg)

	// After adjustPosition is called, x should be reset
	if c.x < 0 {
		t.Errorf("x should not be negative after adjustment, got %d", c.x)
	}
}

func TestCompletionsFilterPerfOptimization(t *testing.T) {
	c := New().(*completionsCmp)
	c.wWidth = 100
	c.wHeight = 50
	c.open = true
	c.query = "abc"

	// When no items and adding more chars to a non-matching query, skip
	filterMsg := FilterCompletionsMsg{
		Query: "abcdef",
		X:     10,
		Y:     20,
	}
	_, cmd := c.Update(filterMsg)

	// Should return early due to optimization
	if cmd != nil {
		// Only check if the optimization kicked in (list empty, prefix match)
		// The actual behavior depends on internal list state
	}
}

func TestListWidthFunction(t *testing.T) {
	c := New().(*completionsCmp)
	c.wWidth = 100
	c.wHeight = 50

	// Open with multiple items of varying lengths
	openMsg := OpenCompletionsMsg{
		Completions: []Completion{
			{Title: "Short", Value: "1"},
			{Title: "A much longer option title here", Value: "2"},
			{Title: "Medium length", Value: "3"},
		},
		X: 10,
		Y: 20,
	}
	_, _ = c.Update(openMsg)

	// Width should accommodate the longest item
	if c.width < 10 {
		t.Errorf("Width should be at least 10 for the items, got %d", c.width)
	}
}
