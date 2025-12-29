package messages

import (
	"strings"
	"testing"

	"github.com/nexora/nexora/internal/agent"
	"github.com/nexora/nexora/internal/agent/tools"
	"github.com/nexora/nexora/internal/message"
)

// Test helper to create basic tool call component for rendering tests
func newTestToolCallCmp(toolName, input string, width int) *toolCallCmp {
	tc := &toolCallCmp{
		width: width,
		call: message.ToolCall{
			ID:       "test-call-1",
			Name:     toolName,
			Input:    input,
			Finished: true,
		},
		result: message.ToolResult{
			ToolCallID: "test-call-1",
			Content:    "Test output",
			IsError:    false,
		},
	}
	return tc
}

func TestRenderPlainContent(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		width        int
		wantNonEmpty bool
	}{
		{
			name:         "simple content",
			content:      "Hello World",
			width:        80,
			wantNonEmpty: true,
		},
		{
			name:         "empty content",
			content:      "",
			width:        80,
			wantNonEmpty: false,
		},
		{
			name:         "content with newlines",
			content:      "Line 1\nLine 2\nLine 3",
			width:        80,
			wantNonEmpty: true,
		},
		{
			name:         "content with tabs",
			content:      "Column1\tColumn2\tColumn3",
			width:        80,
			wantNonEmpty: true,
		},
		{
			name:         "content with CRLF",
			content:      "Windows\r\nLine\r\nEndings",
			width:        80,
			wantNonEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &toolCallCmp{width: tt.width}
			result := renderPlainContent(tc, tt.content)

			// Result may have styling even for empty content, just check it doesn't panic
			if result == "" && tt.wantNonEmpty {
				t.Error("Expected non-empty result")
			}

			// Verify tabs are converted to spaces
			if strings.Contains(tt.content, "\t") && strings.Contains(result, "\t") {
				t.Error("Tabs should be converted to spaces")
			}

			// Verify CRLF is normalized to LF
			if strings.Contains(tt.content, "\r\n") && strings.Contains(result, "\r\n") {
				t.Error("CRLF should be normalized to LF")
			}
		})
	}
}

func TestRenderPlainContentTruncation(t *testing.T) {
	// Create content with more than responseContextHeight lines
	lines := make([]string, 20)
	for i := range lines {
		lines[i] = "Line content"
	}
	longContent := strings.Join(lines, "\n")

	tc := &toolCallCmp{width: 80}
	result := renderPlainContent(tc, longContent)

	// Should include truncation indicator
	if !strings.Contains(result, "…") && !strings.Contains(result, "lines)") {
		t.Error("Long content should show truncation indicator")
	}
}

func TestRenderMarkdownContent(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		wantNonEmpty bool
	}{
		{
			name:         "simple markdown",
			content:      "# Heading\n\nSome **bold** text",
			wantNonEmpty: true,
		},
		{
			name:         "code block",
			content:      "```go\nfunc main() {}\n```",
			wantNonEmpty: true,
		},
		{
			name:         "empty markdown",
			content:      "",
			wantNonEmpty: false,
		},
		{
			name:         "list items",
			content:      "- Item 1\n- Item 2\n- Item 3",
			wantNonEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &toolCallCmp{width: 80}
			result := renderMarkdownContent(tc, tt.content)

			// Result may have styling even for empty content
			if result == "" && tt.wantNonEmpty {
				t.Error("Expected non-empty result")
			}
		})
	}
}

func TestRenderCodeContent(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		content string
		offset  int
	}{
		{
			name:    "go file",
			path:    "main.go",
			content: "package main\n\nfunc main() {\n\tprintln(\"hello\")\n}",
			offset:  0,
		},
		{
			name:    "with offset",
			path:    "test.go",
			content: "line 1\nline 2\nline 3",
			offset:  10,
		},
		{
			name:    "empty content",
			path:    "empty.go",
			content: "",
			offset:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &toolCallCmp{width: 80}
			result := renderCodeContent(tc, tt.path, tt.content, tt.offset)

			if tt.content == "" && result == "" {
				return // empty is ok for empty content
			}

			if result == "" {
				t.Error("Expected non-empty result for code content")
			}
		})
	}
}

func TestRenderImageContent(t *testing.T) {
	tests := []struct {
		name        string
		data        string
		mediaType   string
		textContent string
		wantContain string
	}{
		{
			name:        "png image",
			data:        "base64data",
			mediaType:   "image/png",
			textContent: "",
			wantContain: "image/png",
		},
		{
			name:        "jpeg with text",
			data:        "base64data",
			mediaType:   "image/jpeg",
			textContent: "Image description",
			wantContain: "image/jpeg",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &toolCallCmp{width: 80}
			result := renderImageContent(tc, tt.data, tt.mediaType, tt.textContent)

			if !strings.Contains(result, tt.wantContain) {
				t.Errorf("Expected result to contain %q", tt.wantContain)
			}

			if tt.textContent != "" && !strings.Contains(result, tt.textContent) {
				t.Error("Expected result to include text content")
			}
		})
	}
}

func TestRenderMediaContent(t *testing.T) {
	tests := []struct {
		name        string
		mediaType   string
		textContent string
	}{
		{
			name:        "audio file",
			mediaType:   "audio/mp3",
			textContent: "",
		},
		{
			name:        "video with description",
			mediaType:   "video/mp4",
			textContent: "Video description",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := &toolCallCmp{width: 80}
			result := renderMediaContent(tc, tt.mediaType, tt.textContent)

			if !strings.Contains(result, tt.mediaType) {
				t.Errorf("Expected result to contain media type %q", tt.mediaType)
			}
		})
	}
}

func TestBashRenderer(t *testing.T) {
	renderer := bashRenderer{}

	tests := []struct {
		name      string
		input     string
		wantError bool
	}{
		{
			name:      "valid bash command",
			input:     `{"command":"ls -la","timeout":30000}`,
			wantError: false,
		},
		{
			name:      "invalid json",
			input:     `{invalid}`,
			wantError: true,
		},
		{
			name:      "command with newlines",
			input:     `{"command":"echo 'line1'\necho 'line2'"}`,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestToolCallCmp(tools.BashToolName, tt.input, 80)
			result := renderer.Render(tc)

			if tt.wantError {
				if !strings.Contains(result, "ERROR") {
					t.Error("Expected error rendering for invalid input")
				}
			} else {
				if result == "" {
					t.Error("Expected non-empty rendering")
				}
			}
		})
	}
}

func TestViewRenderer(t *testing.T) {
	renderer := viewRenderer{}

	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "view file",
			input: `{"file_path":"/tmp/test.go","limit":100}`,
		},
		{
			name:  "view with offset",
			input: `{"file_path":"main.go","offset":10,"limit":50}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := newTestToolCallCmp(tools.ViewToolName, tt.input, 80)
			tc.result.Metadata = `{"file_path":"test.go","content":"package main"}`
			result := renderer.Render(tc)

			if result == "" {
				t.Error("Expected non-empty rendering")
			}
		})
	}
}

func TestEditRenderer(t *testing.T) {
	renderer := editRenderer{}

	input := `{"file_path":"test.go","old_string":"old","new_string":"new"}`
	tc := newTestToolCallCmp(tools.EditToolName, input, 80)
	tc.result.Metadata = `{"old_content":"old content","new_content":"new content"}`

	result := renderer.Render(tc)
	if result == "" {
		t.Error("Expected non-empty rendering")
	}
}

func TestWriteRenderer(t *testing.T) {
	renderer := writeRenderer{}

	input := `{"file_path":"new.go","content":"package main\n\nfunc main() {}"}`
	tc := newTestToolCallCmp(tools.WriteToolName, input, 80)

	result := renderer.Render(tc)
	if result == "" {
		t.Error("Expected non-empty rendering")
	}
}

func TestGrepRenderer(t *testing.T) {
	renderer := grepRenderer{}

	input := `{"pattern":"func.*main","path":"./","literal_text":false}`
	tc := newTestToolCallCmp(tools.GrepToolName, input, 80)

	result := renderer.Render(tc)
	if result == "" {
		t.Error("Expected non-empty rendering")
	}
}

func TestGlobRenderer(t *testing.T) {
	renderer := globRenderer{}

	input := `{"pattern":"**/*.go","path":"./"}`
	tc := newTestToolCallCmp(tools.GlobToolName, input, 80)

	result := renderer.Render(tc)
	if result == "" {
		t.Error("Expected non-empty rendering")
	}
}

func TestAgentRenderer(t *testing.T) {
	renderer := agentRenderer{}

	input := `{"prompt":"Do something","subagent_type":"general"}`
	tc := newTestToolCallCmp(agent.AgentToolName, input, 80)

	result := renderer.Render(tc)
	if result == "" {
		t.Error("Expected non-empty rendering")
	}
}

func TestGenericRenderer(t *testing.T) {
	renderer := genericRenderer{}

	tc := newTestToolCallCmp("unknown_tool", `{"param":"value"}`, 80)
	result := renderer.Render(tc)

	if result == "" {
		t.Error("Expected non-empty rendering for generic tool")
	}
}

func TestRenderRegistry(t *testing.T) {
	// Test that all known tools are registered
	knownTools := []string{
		tools.BashToolName,
		tools.ViewToolName,
		tools.EditToolName,
		tools.WriteToolName,
		tools.GrepToolName,
		tools.GlobToolName,
		agent.AgentToolName,
	}

	for _, toolName := range knownTools {
		t.Run(toolName, func(t *testing.T) {
			renderer := registry.lookup(toolName)
			if renderer == nil {
				t.Errorf("Expected renderer for %s", toolName)
			}

			// Should not be generic renderer for known tools
			if _, ok := renderer.(genericRenderer); ok {
				t.Errorf("Known tool %s should have specific renderer", toolName)
			}
		})
	}
}

func TestRenderRegistryFallback(t *testing.T) {
	// Unknown tool should fallback to generic renderer
	renderer := registry.lookup("unknown_tool_xyz")
	if renderer == nil {
		t.Fatal("Expected fallback renderer")
	}

	if _, ok := renderer.(genericRenderer); !ok {
		t.Error("Unknown tool should use generic renderer")
	}
}

func TestParamBuilder(t *testing.T) {
	pb := newParamBuilder()

	// Test addMain
	pb.addMain("main_value")
	args := pb.build()
	if len(args) != 1 || args[0] != "main_value" {
		t.Error("addMain failed")
	}

	// Test addKeyValue
	pb = newParamBuilder()
	pb.addKeyValue("key", "value")
	args = pb.build()
	if len(args) != 2 || args[0] != "key" || args[1] != "value" {
		t.Error("addKeyValue failed")
	}

	// Test addFlag
	pb = newParamBuilder()
	pb.addFlag("flag", true)
	args = pb.build()
	if len(args) != 2 || args[0] != "flag" || args[1] != "true" {
		t.Error("addFlag with true failed")
	}

	pb = newParamBuilder()
	pb.addFlag("flag", false)
	args = pb.build()
	if len(args) != 0 {
		t.Error("addFlag with false should not add anything")
	}

	// Test chaining
	pb = newParamBuilder()
	args = pb.addMain("main").addKeyValue("k1", "v1").addFlag("f1", true).build()
	if len(args) != 5 {
		t.Errorf("Expected 5 args from chaining, got %d", len(args))
	}
}

func TestRenderParamList(t *testing.T) {
	tests := []struct {
		name   string
		params []string
		width  int
	}{
		{
			name:   "empty params",
			params: []string{},
			width:  80,
		},
		{
			name:   "single param",
			params: []string{"main_value"},
			width:  80,
		},
		{
			name:   "main with key-value",
			params: []string{"main", "key", "value"},
			width:  80,
		},
		{
			name:   "truncation test",
			params: []string{"very_long_main_parameter_that_exceeds_width"},
			width:  20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := renderParamList(false, tt.width, tt.params...)
			// Should not panic and should return a string
			_ = result
		})
	}
}

func TestPrettifyToolName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{tools.BashToolName, "Bash"},
		{tools.ViewToolName, "View"},
		{tools.EditToolName, "Edit"},
		{tools.WriteToolName, "Write"},
		{agent.AgentToolName, "Agent"},
		{"unknown_tool", "unknown_tool"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := prettifyToolName(tt.input)
			if got != tt.want {
				t.Errorf("prettifyToolName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatNonZero(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, ""},
		{1, "1"},
		{100, "100"},
		{-5, "-5"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatNonZero(tt.input)
			if got != tt.want {
				t.Errorf("formatNonZero(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatTimeout(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, ""},
		{30, "30s"},
		{60, "1m0s"},
		{3600, "1h0m0s"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatTimeout(tt.input)
			if got != tt.want {
				t.Errorf("formatTimeout(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestFormatSize(t *testing.T) {
	tests := []struct {
		input int
		want  string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1536, "1.5 KB"},
		{1048576, "1.0 MB"},
		{2097152, "2.0 MB"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatSize(tt.input)
			if got != tt.want {
				t.Errorf("formatSize(%d) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetDigits(t *testing.T) {
	tests := []struct {
		input int
		want  int
	}{
		{0, 1},
		{5, 1},
		{10, 2},
		{99, 2},
		{100, 3},
		{-15, 2},
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.input)), func(t *testing.T) {
			got := getDigits(tt.input)
			if got != tt.want {
				t.Errorf("getDigits(%d) = %d, want %d", tt.input, got, tt.want)
			}
		})
	}
}

func TestTruncateHeight(t *testing.T) {
	content := "line1\nline2\nline3\nline4\nline5"

	tests := []struct {
		name   string
		height int
		want   int
	}{
		{"no truncation", 10, 5},
		{"truncate to 3", 3, 3},
		{"truncate to 1", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateHeight(content, tt.height)
			lines := strings.Split(result, "\n")
			if len(lines) != tt.want {
				t.Errorf("Expected %d lines, got %d", tt.want, len(lines))
			}
		})
	}
}

func TestBaseRendererRenderError(t *testing.T) {
	br := baseRenderer{}
	tc := &toolCallCmp{
		width: 80,
		call: message.ToolCall{
			Name: "test_tool",
		},
	}

	result := br.renderError(tc, "Something went wrong")

	if !strings.Contains(result, "ERROR") {
		t.Error("Expected ERROR tag in error rendering")
	}
	if !strings.Contains(result, "Something went wrong") {
		t.Error("Expected error message in rendering")
	}
}

func TestEarlyState(t *testing.T) {
	tests := []struct {
		name       string
		setupFunc  func() *toolCallCmp
		wantDone   bool
		wantInView string
	}{
		{
			name: "error state",
			setupFunc: func() *toolCallCmp {
				tc := &toolCallCmp{width: 80}
				tc.result.IsError = true
				tc.result.Content = "Error occurred"
				return tc
			},
			wantDone:   true,
			wantInView: "ERROR",
		},
		{
			name: "cancelled state",
			setupFunc: func() *toolCallCmp {
				tc := &toolCallCmp{width: 80, cancelled: true}
				return tc
			},
			wantDone:   true,
			wantInView: "Canceled",
		},
		{
			name: "pending state",
			setupFunc: func() *toolCallCmp {
				tc := &toolCallCmp{width: 80}
				tc.result.ToolCallID = "" // not finished
				return tc
			},
			wantDone:   true,
			wantInView: "Waiting",
		},
		{
			name: "normal completion",
			setupFunc: func() *toolCallCmp {
				tc := &toolCallCmp{width: 80}
				tc.result.ToolCallID = "done"
				tc.result.IsError = false
				return tc
			},
			wantDone:   false,
			wantInView: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := tt.setupFunc()
			result, done := earlyState("Header", tc)

			if done != tt.wantDone {
				t.Errorf("Expected done=%v, got %v", tt.wantDone, done)
			}

			if tt.wantInView != "" && !strings.Contains(result, tt.wantInView) {
				t.Errorf("Expected result to contain %q", tt.wantInView)
			}
		})
	}
}
