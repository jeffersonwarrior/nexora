package chat

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/nexora/nexora/internal/app"
	"github.com/nexora/nexora/internal/message"
	"github.com/nexora/nexora/internal/session"
	"github.com/nexora/nexora/internal/tui/exp/list"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name string
		app  *app.App
	}{
		{
			name: "creates message list component",
			app:  &app.App{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmp := New(tt.app)
			if cmp == nil {
				t.Fatal("New() returned nil")
			}

			// Verify it implements the required interface
			// Type is already MessageListCmp from New(), no need to assert
		})
	}
}

func TestMessageListCmpInit(t *testing.T) {
	app := &app.App{}
	cmp := New(app)

	cmd := cmp.Init()
	// Init should return a command (from list component)
	// We just verify it doesn't panic
	_ = cmd
}

func TestMessageListCmpGetSize(t *testing.T) {
	app := &app.App{}
	cmp := New(app).(*messageListCmp)

	// Initial size should be zero
	width, height := cmp.GetSize()
	if width != 0 || height != 0 {
		t.Errorf("Initial size should be (0, 0), got (%d, %d)", width, height)
	}

	// Set size and verify
	cmp.SetSize(80, 24)
	width, height = cmp.GetSize()
	if width != 80 || height != 24 {
		t.Errorf("Expected size (80, 24), got (%d, %d)", width, height)
	}
}

func TestMessageListCmpSetSize(t *testing.T) {
	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"standard terminal", 80, 24},
		{"wide screen", 120, 40},
		{"narrow screen", 40, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &app.App{}
			cmp := New(app).(*messageListCmp)

			cmd := cmp.SetSize(tt.width, tt.height)
			_ = cmd // Just verify it doesn't panic

			width, height := cmp.GetSize()
			if width != tt.width || height != tt.height {
				t.Errorf("Expected size (%d, %d), got (%d, %d)", tt.width, tt.height, width, height)
			}
		})
	}
}

func TestMessageListCmpFocus(t *testing.T) {
	app := &app.App{}
	cmp := New(app).(*messageListCmp)

	// Focus should return a command
	cmd := cmp.Focus()
	_ = cmd

	// Verify focused state
	if !cmp.IsFocused() {
		t.Error("Component should be focused after Focus()")
	}
}

func TestMessageListCmpBlur(t *testing.T) {
	app := &app.App{}
	cmp := New(app).(*messageListCmp)

	// First focus
	cmp.Focus()
	if !cmp.IsFocused() {
		t.Fatal("Component should be focused after Focus()")
	}

	// Then blur
	cmd := cmp.Blur()
	_ = cmd

	// Verify blurred state
	if cmp.IsFocused() {
		t.Error("Component should not be focused after Blur()")
	}
}

func TestMessageListCmpIsFocused(t *testing.T) {
	app := &app.App{}
	cmp := New(app).(*messageListCmp)

	// Initially not focused
	if cmp.IsFocused() {
		t.Error("Component should not be focused initially")
	}

	// After focus
	cmp.Focus()
	if !cmp.IsFocused() {
		t.Error("Component should be focused after Focus()")
	}

	// After blur
	cmp.Blur()
	if cmp.IsFocused() {
		t.Error("Component should not be focused after Blur()")
	}
}

func TestMessageListCmpView(t *testing.T) {
	app := &app.App{}
	cmp := New(app).(*messageListCmp)

	// Set a size
	cmp.SetSize(80, 24)

	// View should return a string
	view := cmp.View()
	if view == "" {
		t.Error("View() should return a non-empty string")
	}
}

func TestMessageListCmpGetSelectedText(t *testing.T) {
	app := &app.App{}
	cmp := New(app).(*messageListCmp)

	// Initially no selection
	text := cmp.GetSelectedText()
	if text != "" {
		t.Errorf("Expected empty selected text, got '%s'", text)
	}
}

func TestMessageExists(t *testing.T) {
	tests := []struct {
		name      string
		items     []list.Item
		messageID string
		want      bool
	}{
		{
			name:      "empty list",
			items:     []list.Item{},
			messageID: "msg-123",
			want:      false,
		},
		{
			name:      "message not found",
			items:     []list.Item{},
			messageID: "msg-456",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &app.App{}
			cmp := New(app).(*messageListCmp)
			cmp.listCmp.SetItems(tt.items)

			got := cmp.messageExists(tt.messageID)
			if got != tt.want {
				t.Errorf("messageExists() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuildToolResultMap(t *testing.T) {
	tests := []struct {
		name     string
		messages []message.Message
		want     int // expected map size
	}{
		{
			name:     "empty messages",
			messages: []message.Message{},
			want:     0,
		},
		{
			name: "messages without tool results",
			messages: []message.Message{
				{Role: message.User, Parts: []message.ContentPart{message.TextContent{Text: "Hello"}}},
				{Role: message.Assistant, Parts: []message.ContentPart{message.TextContent{Text: "Hi there"}}},
			},
			want: 0,
		},
		{
			name: "messages with tool results",
			messages: []message.Message{
				{
					Role: message.Tool,
					Parts: []message.ContentPart{
						message.ToolResult{ToolCallID: "call-1", Content: "result-1"},
					},
				},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &app.App{}
			cmp := New(app).(*messageListCmp)

			resultMap := cmp.buildToolResultMap(tt.messages)
			if len(resultMap) != tt.want {
				t.Errorf("buildToolResultMap() map size = %d, want %d", len(resultMap), tt.want)
			}
		})
	}
}

func TestShouldShowAssistantMessage(t *testing.T) {
	tests := []struct {
		name string
		msg  message.Message
		want bool
	}{
		{
			name: "message with content",
			msg:  message.Message{Parts: []message.ContentPart{message.TextContent{Text: "Hello"}}},
			want: true,
		},
		{
			name: "message without content or tool calls",
			msg:  message.Message{},
			want: true, // No tool calls, so should show
		},
		{
			name: "message with tool calls but no content",
			msg: message.Message{
				Parts: []message.ContentPart{
					message.ToolCall{ID: "call-1", Name: "tool1"},
				},
			},
			want: false, // Has tool calls but no content
		},
		{
			name: "message with tool calls and content",
			msg: message.Message{
				Parts: []message.ContentPart{
					message.ToolCall{ID: "call-1", Name: "tool1"},
					message.TextContent{Text: "Using tool"},
				},
			},
			want: true, // Has content despite tool calls
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &app.App{}
			cmp := New(app).(*messageListCmp)

			got := cmp.shouldShowAssistantMessage(tt.msg)
			if got != tt.want {
				t.Errorf("shouldShowAssistantMessage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFindToolCallByID(t *testing.T) {
	tests := []struct {
		name       string
		items      []list.Item
		toolCallID string
		want       int
	}{
		{
			name:       "empty list",
			items:      []list.Item{},
			toolCallID: "call-1",
			want:       NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &app.App{}
			cmp := New(app).(*messageListCmp)

			got := cmp.findToolCallByID(tt.items, tt.toolCallID)
			if got != tt.want {
				t.Errorf("findToolCallByID() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestMessageListCmpUpdate(t *testing.T) {
	tests := []struct {
		name string
		msg  tea.Msg
	}{
		{
			name: "session cleared message",
			msg:  SessionClearedMsg{},
		},
		{
			name: "send message",
			msg:  SendMsg{Text: "Hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &app.App{}
			cmp := New(app)

			// Update should not panic
			_, cmd := cmp.Update(tt.msg)
			_ = cmd
		})
	}
}

func TestConvertMessagesToUI(t *testing.T) {
	tests := []struct {
		name     string
		messages []message.Message
		want     int // expected number of UI items
	}{
		{
			name:     "empty messages",
			messages: []message.Message{},
			want:     0,
		},
		{
			name: "user message",
			messages: []message.Message{
				{Role: message.User, Parts: []message.ContentPart{message.TextContent{Text: "Hello"}}},
			},
			want: 1,
		},
		{
			name: "assistant message",
			messages: []message.Message{
				{Role: message.Assistant, Parts: []message.ContentPart{message.TextContent{Text: "Hi there"}}},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := &app.App{}
			cmp := New(app).(*messageListCmp)

			toolResultMap := cmp.buildToolResultMap(tt.messages)
			items := cmp.convertMessagesToUI(tt.messages, toolResultMap)

			if len(items) != tt.want {
				t.Errorf("convertMessagesToUI() returned %d items, want %d", len(items), tt.want)
			}
		})
	}
}

func TestSetSession_SameSession(t *testing.T) {
	app := &app.App{}
	cmp := New(app).(*messageListCmp)

	// Set the session ID manually
	cmp.session = session.Session{ID: "session-1"}

	// Call SetSession with same session
	cmd := cmp.SetSession(session.Session{ID: "session-1"})

	// Should return nil when setting same session
	if cmd != nil {
		t.Error("SetSession should return nil when setting same session")
	}
}
