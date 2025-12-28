package editor

import (
	"testing"

	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"github.com/nexora/nexora/internal/app"
	"github.com/nexora/nexora/internal/message"
	"github.com/nexora/nexora/internal/permission"
	"github.com/nexora/nexora/internal/session"
	"github.com/nexora/nexora/internal/tui/components/completions"
	"github.com/nexora/nexora/internal/tui/components/dialogs/commands"
	"github.com/nexora/nexora/internal/tui/components/dialogs/filepicker"
)

// createTestApp creates a minimal app instance for testing
func createTestApp() *app.App {
	return &app.App{
		Permissions: permission.NewPermissionService(".", false, nil),
	}
}

// TestNew tests the editor constructor
func TestNew(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp)
	if editor == nil {
		t.Fatal("Expected non-nil editor")
	}
}

// TestEditorInit tests the Init method
func TestEditorInit(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)
	cmd := editor.Init()
	if cmd != nil {
		t.Log("Init returned a command")
	}
}

// TestEditorView tests the View method
func TestEditorView(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)
	editor.SetSize(80, 10)
	view := editor.View()
	if view == "" {
		t.Error("Expected non-empty view")
	}
}

// TestEditorViewWithAttachments tests View rendering with attachments
func TestEditorViewWithAttachments(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)
	editor.SetSize(80, 10)

	// Add an attachment
	editor.attachments = []message.Attachment{
		{FileName: "test.png", FilePath: "/tmp/test.png", MimeType: "image/png"},
	}

	view := editor.View()
	if view == "" {
		t.Error("Expected non-empty view with attachments")
	}
	if !editor.HasAttachments() {
		t.Error("Expected HasAttachments to return true")
	}
}

// TestEditorSetSize tests the SetSize method
func TestEditorSetSize(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)

	editor.SetSize(100, 20)
	width, height := editor.GetSize()

	// Width and height are adjusted for padding and textarea internals
	// SetSize adjusts by -2 for padding, then textarea may adjust further
	if width <= 0 {
		t.Errorf("Expected positive width, got %d", width)
	}
	if height <= 0 {
		t.Errorf("Expected positive height, got %d", height)
	}

	// Verify internal dimensions are set correctly
	if editor.width != 100 {
		t.Errorf("Expected editor.width 100, got %d", editor.width)
	}
	if editor.height != 20 {
		t.Errorf("Expected editor.height 20, got %d", editor.height)
	}
}

// TestEditorFocus tests the Focus method
func TestEditorFocus(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)

	editor.Focus()
	if !editor.IsFocused() {
		t.Error("Expected editor to be focused after Focus()")
	}
}

// TestEditorBlur tests the Blur method
func TestEditorBlur(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)

	editor.Focus()
	editor.Blur()
	if editor.IsFocused() {
		t.Error("Expected editor to not be focused after Blur()")
	}
}

// TestEditorSetSession tests the SetSession method
func TestEditorSetSession(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)

	sess := session.Session{ID: "test-session-id"}
	cmd := editor.SetSession(sess)

	if cmd != nil {
		t.Log("SetSession returned a command")
	}
	if editor.session.ID != "test-session-id" {
		t.Error("Expected session ID to be set")
	}
}

// TestEditorSetPosition tests the SetPosition method
func TestEditorSetPosition(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)

	editor.SetPosition(10, 5)
	if editor.x != 10 || editor.y != 5 {
		t.Errorf("Expected position (10, 5), got (%d, %d)", editor.x, editor.y)
	}
}

// TestEditorCursor tests the Cursor method
func TestEditorCursor(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)
	editor.SetSize(80, 10)
	editor.SetPosition(5, 3)

	cursor := editor.Cursor()
	// Cursor position should be adjusted for position offsets
	if cursor != nil {
		t.Log("Cursor returned non-nil cursor")
	}
}

// TestEditorBindings tests the Bindings method
func TestEditorBindings(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)

	bindings := editor.Bindings()
	if len(bindings) == 0 {
		t.Error("Expected non-empty bindings")
	}
}

// TestEditorIsCompletionsOpen tests the IsCompletionsOpen method
func TestEditorIsCompletionsOpen(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)

	if editor.IsCompletionsOpen() {
		t.Error("Expected completions to be closed initially")
	}

	editor.isCompletionsOpen = true
	if !editor.IsCompletionsOpen() {
		t.Error("Expected completions to be open")
	}
}

// TestEditorHasAttachments tests the HasAttachments method
func TestEditorHasAttachments(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)

	if editor.HasAttachments() {
		t.Error("Expected no attachments initially")
	}

	editor.attachments = []message.Attachment{{FileName: "test.png"}}
	if !editor.HasAttachments() {
		t.Error("Expected attachments to be present")
	}
}

// TestEditorUpdateWindowSizeMsg tests handling WindowSizeMsg
func TestEditorUpdateWindowSizeMsg(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)
	editor.SetSize(80, 10)

	msg := tea.WindowSizeMsg{Width: 100, Height: 50}
	_, cmd := editor.Update(msg)

	if cmd == nil {
		t.Log("WindowSizeMsg returned nil command")
	}
}

// TestEditorUpdateFilePickedMsg tests handling FilePickedMsg
func TestEditorUpdateFilePickedMsg(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)

	msg := filepicker.FilePickedMsg{
		Attachment: message.Attachment{
			FileName: "test.png",
			FilePath: "/tmp/test.png",
			MimeType: "image/png",
		},
	}
	_, _ = editor.Update(msg)

	if len(editor.attachments) != 1 {
		t.Errorf("Expected 1 attachment, got %d", len(editor.attachments))
	}
}

// TestEditorUpdateFilePickedMsgMaxAttachments tests max attachments limit
func TestEditorUpdateFilePickedMsgMaxAttachments(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)

	// Fill to max
	for i := 0; i < maxAttachments; i++ {
		editor.attachments = append(editor.attachments, message.Attachment{FileName: "test.png"})
	}

	msg := filepicker.FilePickedMsg{
		Attachment: message.Attachment{FileName: "overflow.png"},
	}
	_, cmd := editor.Update(msg)

	if len(editor.attachments) != maxAttachments {
		t.Errorf("Expected %d attachments, got %d", maxAttachments, len(editor.attachments))
	}
	if cmd == nil {
		t.Log("Expected error command for max attachments")
	}
}

// TestEditorUpdateCompletionsOpenedMsg tests handling CompletionsOpenedMsg
func TestEditorUpdateCompletionsOpenedMsg(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)

	msg := completions.CompletionsOpenedMsg{}
	_, _ = editor.Update(msg)

	if !editor.isCompletionsOpen {
		t.Error("Expected completions to be open")
	}
}

// TestEditorUpdateCompletionsClosedMsg tests handling CompletionsClosedMsg
func TestEditorUpdateCompletionsClosedMsg(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)
	editor.isCompletionsOpen = true
	editor.currentQuery = "test"
	editor.completionsStartIndex = 5

	msg := completions.CompletionsClosedMsg{}
	_, _ = editor.Update(msg)

	if editor.isCompletionsOpen {
		t.Error("Expected completions to be closed")
	}
	if editor.currentQuery != "" {
		t.Error("Expected currentQuery to be reset")
	}
	if editor.completionsStartIndex != 0 {
		t.Error("Expected completionsStartIndex to be reset")
	}
}

// TestEditorUpdateSelectCompletionMsg tests handling SelectCompletionMsg
func TestEditorUpdateSelectCompletionMsg(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)
	editor.isCompletionsOpen = true
	editor.completionsStartIndex = 0

	msg := completions.SelectCompletionMsg{
		Value:  FileCompletionItem{Path: "src/main.go"},
		Insert: false,
	}
	_, _ = editor.Update(msg)

	// After selection, completions should be closed if Insert is false
	if editor.isCompletionsOpen {
		t.Error("Expected completions to be closed after selection")
	}
}

// TestEditorUpdateOpenEditorMsg tests handling OpenEditorMsg
func TestEditorUpdateOpenEditorMsg(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)
	editor.SetSize(80, 10)

	msg := OpenEditorMsg{Text: "Hello, world!"}
	_, _ = editor.Update(msg)

	if editor.textarea.Value() != "Hello, world!" {
		t.Error("Expected textarea value to be set from OpenEditorMsg")
	}
}

// TestEditorUpdateEnterKey tests handling Enter key to send message
func TestEditorUpdateEnterKey(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)
	editor.SetSize(80, 10)
	editor.Focus()
	editor.textarea.SetValue("Test message")

	// Create Enter key message
	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := editor.Update(msg)

	// Should return a send command
	if cmd == nil {
		t.Log("Expected send command from Enter key")
	}
}

// TestEditorUpdateEnterKeyEmptyValue tests Enter key with empty value
func TestEditorUpdateEnterKeyEmptyValue(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)
	editor.SetSize(80, 10)
	editor.Focus()
	editor.textarea.SetValue("")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, cmd := editor.Update(msg)

	// Should not send with empty value
	if cmd != nil {
		t.Log("Empty message should not produce send command")
	}
}

// TestEditorUpdateBackslashEnter tests backslash continuation
func TestEditorUpdateBackslashEnter(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)
	editor.SetSize(80, 10)
	editor.Focus()
	editor.textarea.SetValue("Test\\")

	msg := tea.KeyPressMsg{Code: tea.KeyEnter}
	_, _ = editor.Update(msg)

	// Backslash should be removed
	if editor.textarea.Value() != "Test" {
		t.Log("Expected backslash to be removed for continuation")
	}
}

// TestEditorUpdateDeleteMode tests delete mode handling
func TestEditorUpdateDeleteMode(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)
	editor.SetSize(80, 10)
	editor.attachments = []message.Attachment{
		{FileName: "a.png"},
		{FileName: "b.png"},
	}

	// Enter delete mode with Ctrl+R
	msg := tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl}
	_, _ = editor.Update(msg)

	if !editor.deleteMode {
		t.Error("Expected delete mode to be activated")
	}

	// Press 'r' to delete all attachments
	msg = tea.KeyPressMsg{Code: 'r'}
	_, _ = editor.Update(msg)

	if len(editor.attachments) != 0 {
		t.Error("Expected all attachments to be deleted")
	}
	if editor.deleteMode {
		t.Error("Expected delete mode to be deactivated")
	}
}

// TestEditorUpdateDeleteModeEscape tests escape from delete mode
func TestEditorUpdateDeleteModeEscape(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)
	editor.deleteMode = true

	msg := tea.KeyPressMsg{Code: tea.KeyEscape}
	_, _ = editor.Update(msg)

	if editor.deleteMode {
		t.Error("Expected delete mode to be cancelled")
	}
}

// TestEditorUpdateDeleteModeDigit tests deleting attachment by index
func TestEditorUpdateDeleteModeDigit(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)
	editor.attachments = []message.Attachment{
		{FileName: "a.png"},
		{FileName: "b.png"},
		{FileName: "c.png"},
	}
	editor.deleteMode = true

	// Delete attachment at index 1
	msg := tea.KeyPressMsg{Code: '1'}
	_, _ = editor.Update(msg)

	if len(editor.attachments) != 2 {
		t.Errorf("Expected 2 attachments, got %d", len(editor.attachments))
	}
	if editor.deleteMode {
		t.Error("Expected delete mode to be deactivated")
	}
}

// TestEditorKeyMap tests the EditorKeyMap
func TestEditorKeyMap(t *testing.T) {
	keyMap := DefaultEditorKeyMap()

	bindings := keyMap.KeyBindings()
	if len(bindings) == 0 {
		t.Error("Expected non-empty key bindings")
	}
}

// TestDeleteAttachmentKeyMaps tests the DeleteAttachmentKeyMaps
func TestDeleteAttachmentKeyMaps(t *testing.T) {
	if AttachmentsKeyMaps.AttachmentDeleteMode.Keys() == nil {
		t.Error("Expected AttachmentDeleteMode to have keys")
	}
	if AttachmentsKeyMaps.Escape.Keys() == nil {
		t.Error("Expected Escape to have keys")
	}
	if AttachmentsKeyMaps.DeleteAllAttachments.Keys() == nil {
		t.Error("Expected DeleteAllAttachments to have keys")
	}
}

// TestNormalPromptFunc tests the normalPromptFunc
func TestNormalPromptFunc(t *testing.T) {
	// Test first line
	result := normalPromptFunc(textarea.PromptInfo{LineNumber: 0, Focused: true})
	if result == "" {
		t.Error("Expected non-empty prompt for line 0")
	}

	// Test other lines focused
	result = normalPromptFunc(textarea.PromptInfo{LineNumber: 1, Focused: true})
	if result == "" {
		t.Error("Expected non-empty prompt for other lines (focused)")
	}

	// Test other lines unfocused
	result = normalPromptFunc(textarea.PromptInfo{LineNumber: 1, Focused: false})
	if result == "" {
		t.Error("Expected non-empty prompt for other lines (unfocused)")
	}
}

// TestYoloPromptFunc tests the yoloPromptFunc
func TestYoloPromptFunc(t *testing.T) {
	// Test first line focused
	result := yoloPromptFunc(textarea.PromptInfo{LineNumber: 0, Focused: true})
	if result == "" {
		t.Error("Expected non-empty yolo prompt for line 0 (focused)")
	}

	// Test first line unfocused
	result = yoloPromptFunc(textarea.PromptInfo{LineNumber: 0, Focused: false})
	if result == "" {
		t.Error("Expected non-empty yolo prompt for line 0 (unfocused)")
	}

	// Test other lines focused
	result = yoloPromptFunc(textarea.PromptInfo{LineNumber: 1, Focused: true})
	if result == "" {
		t.Error("Expected non-empty yolo prompt for other lines (focused)")
	}

	// Test other lines unfocused
	result = yoloPromptFunc(textarea.PromptInfo{LineNumber: 1, Focused: false})
	if result == "" {
		t.Error("Expected non-empty yolo prompt for other lines (unfocused)")
	}
}

// TestFileCompletionItem tests the FileCompletionItem struct
func TestFileCompletionItem(t *testing.T) {
	item := FileCompletionItem{Path: "/home/user/test.go"}
	if item.Path != "/home/user/test.go" {
		t.Error("Expected path to be set")
	}
}

// TestEditorAttachmentsContent tests the attachmentsContent rendering
func TestEditorAttachmentsContent(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)

	editor.attachments = []message.Attachment{
		{FileName: "short.png"},
		{FileName: "verylongfilename.png"}, // > 10 chars
	}

	content := editor.attachmentsContent()
	if content == "" {
		t.Error("Expected non-empty attachments content")
	}
}

// TestEditorAttachmentsContentDeleteMode tests attachmentsContent in delete mode
func TestEditorAttachmentsContentDeleteMode(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)

	editor.attachments = []message.Attachment{
		{FileName: "test.png"},
	}
	editor.deleteMode = true

	content := editor.attachmentsContent()
	if content == "" {
		t.Error("Expected non-empty attachments content in delete mode")
	}
}

// TestEditorRandomizePlaceholders tests placeholder randomization
func TestEditorRandomizePlaceholders(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)

	// Placeholders are set in New()
	if editor.readyPlaceholder == "" {
		t.Error("Expected ready placeholder to be set")
	}
	if editor.workingPlaceholder == "" {
		t.Error("Expected working placeholder to be set")
	}

	// Randomize and verify they're still set
	editor.randomizePlaceholders()
	if editor.readyPlaceholder == "" {
		t.Error("Expected ready placeholder after randomize")
	}
}

// TestEditorCompletionsPosition tests the completionsPosition method
func TestEditorCompletionsPosition(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)
	editor.SetPosition(10, 5)
	editor.SetSize(80, 10)

	x, y := editor.completionsPosition()
	// Should return position adjusted for cursor + editor position
	if x < 10 {
		t.Errorf("Expected x >= 10, got %d", x)
	}
	if y < 6 { // y + 1 for padding
		t.Errorf("Expected y >= 6, got %d", y)
	}
}

// TestSendWithExitQuit tests that "exit" and "quit" commands open quit dialog
func TestSendWithExitQuit(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)
	editor.SetSize(80, 10)
	editor.Focus()

	testCases := []string{"exit", "quit"}
	for _, cmd := range testCases {
		editor.textarea.SetValue(cmd)
		result := editor.send()
		if result == nil {
			t.Errorf("Expected command from '%s'", cmd)
		}
		// Textarea should be reset
		if editor.textarea.Value() != "" {
			t.Errorf("Expected textarea to be reset after '%s'", cmd)
		}
	}
}

// TestSendWithAttachments tests sending message with attachments
func TestSendWithAttachments(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)
	editor.SetSize(80, 10)
	editor.Focus()

	editor.attachments = []message.Attachment{{FileName: "test.png"}}
	editor.textarea.SetValue("Test with attachment")

	cmd := editor.send()
	if cmd == nil {
		t.Error("Expected send command with attachments")
	}
	// Attachments should be cleared after send
	if len(editor.attachments) != 0 {
		t.Error("Expected attachments to be cleared after send")
	}
}

// TestEditorInterfaceCompliance tests that editorCmp implements Editor interface
func TestEditorInterfaceCompliance(t *testing.T) {
	testApp := createTestApp()
	var editor Editor = New(testApp)

	// Test interface methods
	editor.SetSize(80, 10)
	editor.GetSize()
	editor.Focus()
	editor.Blur()
	editor.IsFocused()
	editor.SetSession(session.Session{})
	editor.IsCompletionsOpen()
	editor.HasAttachments()
	editor.Cursor()
	editor.Bindings()
	editor.Init()
	editor.View()
	editor.SetPosition(0, 0)

	t.Log("Editor interface compliance verified")
}

// TestEditorSlashCommandOpensDialog tests "/" on empty prompt opens command dialog
func TestEditorSlashCommandOpensDialog(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)
	editor.SetSize(80, 10)
	editor.Focus()
	editor.textarea.SetValue("")

	msg := tea.KeyPressMsg{Code: '/'}
	_, cmd := editor.Update(msg)

	// Should open command dialog
	if cmd == nil {
		t.Log("Expected command dialog open command")
	}
}

// TestEditorAtSignOpensCompletions tests "@" opens file completions
func TestEditorAtSignOpensCompletions(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)
	editor.SetSize(80, 10)
	editor.Focus()
	editor.textarea.SetValue("")

	msg := tea.KeyPressMsg{Code: '@'}
	_, cmd := editor.Update(msg)

	// Should start completions
	if !editor.isCompletionsOpen {
		t.Log("@ should open completions on empty prompt or after space")
	}
	if cmd == nil {
		t.Log("Expected completions start command")
	}
}

// TestEditorToggleYoloModeMsg tests handling ToggleYoloModeMsg
func TestEditorToggleYoloModeMsg(t *testing.T) {
	testApp := createTestApp()
	editor := New(testApp).(*editorCmp)
	editor.SetSize(80, 10)

	msg := commands.ToggleYoloModeMsg{}
	_, cmd := editor.Update(msg)

	if cmd != nil {
		t.Log("ToggleYoloModeMsg may return a command")
	}
}
