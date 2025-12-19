package utils

import (
	"fmt"
	"strings"
)

// TextToolCall represents a tool call parsed from text output
type TextToolCall struct {
	ID        string
	Name      string
	Arguments string
}

// ParseTextToolCalls extracts tool calls from text that contains malformed tool call patterns.
// Some models (like MiniMax M2 via Synthetic, x.ai Grok) output tool calls as text instead of proper JSON,
// using formats like:
//   - minimax:tool_call /path 30 310 </minimax:tool_call>
//   - <tool_call>{"name": "view", "arguments": {...}}</tool_call>
//   - <function_call name="view">...</function_call>
//   - <xai:function_call name="view">...</xai:function_call>
//
// This function detects and parses these patterns to enable conversation continuation.
func ParseTextToolCalls(text string) []TextToolCall {
	var calls []TextToolCall

	// Pattern 1: minimax:tool_call format (e.g., "minimax:tool_call /path 30 310 </minimax:tool_call>")
	// This appears to be a view tool call with path, limit, offset
	if idx := strings.Index(text, "minimax:tool_call"); idx != -1 {
		endIdx := strings.Index(text[idx:], "</minimax:tool_call>")
		if endIdx != -1 {
			content := strings.TrimSpace(text[idx+len("minimax:tool_call") : idx+endIdx])
			// Parse the content - appears to be space-separated: path limit offset
			parts := strings.Fields(content)
			if len(parts) >= 1 {
				// Construct a view tool call
				call := TextToolCall{
					ID:   GenerateToolCallID("openai"),
					Name: "view",
				}
				// Build arguments JSON
				args := map[string]any{"file_path": parts[0]}
				if len(parts) >= 2 {
					if limit, err := parseInt(parts[1]); err == nil {
						args["limit"] = limit
					}
				}
				if len(parts) >= 3 {
					if offset, err := parseInt(parts[2]); err == nil {
						args["offset"] = offset
					}
				}
				call.Arguments = mapToJSON(args)
				calls = append(calls, call)
			}
		}
	}

	// Pattern 2: <tool_call> JSON format
	toolCallStart := "<tool_call>"
	toolCallEnd := "</tool_call>"
	searchText := text
	for {
		startIdx := strings.Index(searchText, toolCallStart)
		if startIdx == -1 {
			break
		}
		endIdx := strings.Index(searchText[startIdx:], toolCallEnd)
		if endIdx == -1 {
			break
		}
		content := strings.TrimSpace(searchText[startIdx+len(toolCallStart) : startIdx+endIdx])
		if call := parseJSONToolCall(content); call != nil {
			calls = append(calls, *call)
		}
		searchText = searchText[startIdx+endIdx+len(toolCallEnd):]
	}

	// Pattern 3: <function_call name="..."> format
	funcCallPattern := "<function_call"
	funcCallEnd := "</function_call>"
	searchText = text
	for {
		startIdx := strings.Index(searchText, funcCallPattern)
		if startIdx == -1 {
			break
		}
		// Find the closing >
		tagEnd := strings.Index(searchText[startIdx:], ">")
		if tagEnd == -1 {
			break
		}
		// Extract name attribute
		tagContent := searchText[startIdx : startIdx+tagEnd+1]
		name := extractAttribute(tagContent, "name")
		if name == "" {
			searchText = searchText[startIdx+tagEnd+1:]
			continue
		}

		endIdx := strings.Index(searchText[startIdx:], funcCallEnd)
		if endIdx == -1 {
			break
		}
		arguments := strings.TrimSpace(searchText[startIdx+tagEnd+1 : startIdx+endIdx])
		calls = append(calls, TextToolCall{
			ID:        GenerateToolCallID("openai"),
			Name:      name,
			Arguments: arguments,
		})
		searchText = searchText[startIdx+endIdx+len(funcCallEnd):]
	}

	// Pattern 4: <xai:function_call name="..."> format (x.ai Grok models)
	xaiCallPattern := "<xai:function_call"
	xaiCallEnd := "</xai:function_call>"
	searchText = text
	for {
		startIdx := strings.Index(searchText, xaiCallPattern)
		if startIdx == -1 {
			break
		}
		// Find the closing >
		tagEnd := strings.Index(searchText[startIdx:], ">")
		if tagEnd == -1 {
			break
		}
		// Extract name attribute
		tagContent := searchText[startIdx : startIdx+tagEnd+1]
		name := extractAttribute(tagContent, "name")
		if name == "" {
			searchText = searchText[startIdx+tagEnd+1:]
			continue
		}

		endIdx := strings.Index(searchText[startIdx:], xaiCallEnd)
		if endIdx == -1 {
			break
		}
		arguments := strings.TrimSpace(searchText[startIdx+tagEnd+1 : startIdx+endIdx])
		calls = append(calls, TextToolCall{
			ID:        GenerateToolCallID("openai"),
			Name:      name,
			Arguments: arguments,
		})
		// Fix x.ai indexing bug: use correct end position
		searchText = searchText[(startIdx+tagEnd+1)+endIdx+len(xaiCallEnd):]
	}

	return calls
}

// HasTextToolCalls returns true if the text contains any recognizable tool call patterns
func HasTextToolCalls(text string) bool {
	patterns := []string{
		"minimax:tool_call",
		"</minimax:tool_call>",
		"<tool_call>",
		"</tool_call>",
		"<function_call",
		"</function_call>",
		"<xai:function_call",
		"</xai:function_call>",
	}
	for _, p := range patterns {
		if strings.Contains(text, p) {
			return true
		}
	}
	return false
}

// parseInt parses a string to int, used for parsing tool call arguments
func parseInt(s string) (int, error) {
	var n int
	_, err := fmt.Sscanf(s, "%d", &n)
	return n, err
}

// mapToJSON converts a map to a simple JSON string
func mapToJSON(m map[string]any) string {
	var parts []string
	for k, v := range m {
		switch val := v.(type) {
		case string:
			parts = append(parts, fmt.Sprintf(`"%s":"%s"`, k, val))
		case int:
			parts = append(parts, fmt.Sprintf(`"%s":%d`, k, val))
		default:
			parts = append(parts, fmt.Sprintf(`"%s":"%v"`, k, val))
		}
	}
	return "{" + strings.Join(parts, ",") + "}"
}

// extractAttribute extracts an attribute value from an XML-like tag
func extractAttribute(tag, attr string) string {
	// Look for attr="value" or attr='value'
	patterns := []string{
		fmt.Sprintf(`%s="`, attr),
		fmt.Sprintf(`%s='`, attr),
	}
	for _, p := range patterns {
		idx := strings.Index(tag, p)
		if idx == -1 {
			continue
		}
		start := idx + len(p)
		quote := tag[start-1]
		end := strings.IndexByte(tag[start:], quote)
		if end != -1 {
			return tag[start : start+end]
		}
	}
	return ""
}

// parseJSONToolCall attempts to parse a JSON tool call from content
func parseJSONToolCall(content string) *TextToolCall {
	// Simple JSON parsing - look for "name" and "arguments" fields
	// This is intentionally simple to avoid heavy dependencies
	name := extractJSONString(content, "name")
	if name == "" {
		return nil
	}

	// Try to extract arguments as a nested object or string
	args := extractJSONValue(content, "arguments")
	if args == "" {
		args = "{}"
	}

	return &TextToolCall{
		ID:        GenerateToolCallID("openai"),
		Name:      name,
		Arguments: args,
	}
}

// extractJSONString extracts a string value from JSON content
func extractJSONString(json, key string) string {
	// Look for "key": "value" or "key":"value"
	patterns := []string{
		fmt.Sprintf(`"%s": "`, key),
		fmt.Sprintf(`"%s":"`, key),
	}
	for _, p := range patterns {
		idx := strings.Index(json, p)
		if idx == -1 {
			continue
		}
		start := idx + len(p)
		end := strings.IndexByte(json[start:], '"')
		if end != -1 {
			return json[start : start+end]
		}
	}
	return ""
}

// extractJSONValue extracts a JSON value (object or string) from content
func extractJSONValue(json, key string) string {
	// Look for "key": { or "key":{
	patterns := []string{
		fmt.Sprintf(`"%s": {`, key),
		fmt.Sprintf(`"%s":{`, key),
	}
	for _, p := range patterns {
		idx := strings.Index(json, p)
		if idx == -1 {
			continue
		}
		start := idx + len(p) - 1 // Include the opening brace
		// Find matching closing brace
		depth := 1
		for i := start + 1; i < len(json); i++ {
			switch json[i] {
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					return json[start : i+1]
				}
			}
		}
	}
	// Try string value
	return extractJSONString(json, key)
}
