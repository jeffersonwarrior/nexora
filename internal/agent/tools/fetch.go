package tools

import (
	"context"
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"charm.land/fantasy"
	md "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/PuerkitoBio/goquery"
	"github.com/nexora/nexora/internal/permission"
)

const FetchToolName = "fetch"

// Context limits for different models (approximate token limits)
const (
	ContextLimitSmall  = 32000  // claude-3-5-haiku
	ContextLimitMedium = 80000  // claude-3-5-sonnet
	ContextLimitLarge  = 200000 // claude-4-opus
)

// DefaultContextLimit is used when model-specific limit is unknown
const DefaultContextLimit = ContextLimitSmall

//go:embed fetch.md
var fetchDescription []byte

//go:embed web_fetch.md
var webFetchToolDescription []byte

func NewFetchTool(permissions permission.Service, workingDir string, client *http.Client) fantasy.AgentTool {
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		}
	}

	return fantasy.NewParallelAgentTool(
		FetchToolName,
		string(fetchDescription),
		func(ctx context.Context, params FetchParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.URL == "" {
				return fantasy.NewTextErrorResponse("URL parameter is required"), nil
			}

			format := strings.ToLower(params.Format)
			if format != "text" && format != "markdown" && format != "html" {
				return fantasy.NewTextErrorResponse("Format must be one of: text, markdown, html"), nil
			}

			if !strings.HasPrefix(params.URL, "http://") && !strings.HasPrefix(params.URL, "https://") {
				return fantasy.NewTextErrorResponse("URL must start with http:// or https://"), nil
			}

			sessionID := GetSessionFromContext(ctx)
			if sessionID == "" {
				return fantasy.ToolResponse{}, fmt.Errorf("session ID is required for creating a new file")
			}

			p := permissions.Request(
				permission.CreatePermissionRequest{
					SessionID:   sessionID,
					Path:        workingDir,
					ToolCallID:  call.ID,
					ToolName:    FetchToolName,
					Action:      "fetch",
					Description: fmt.Sprintf("Fetch content from URL: %s", params.URL),
					Params:      FetchPermissionsParams(params),
				},
			)

			if !p {
				return fantasy.ToolResponse{}, permission.ErrorPermissionDenied
			}

			// Handle timeout with context
			requestCtx := ctx
			if params.Timeout > 0 {
				maxTimeout := 120 // 2 minutes
				if params.Timeout > maxTimeout {
					params.Timeout = maxTimeout
				}
				var cancel context.CancelFunc
				requestCtx, cancel = context.WithTimeout(ctx, time.Duration(params.Timeout)*time.Second)
				defer cancel()
			}

			req, err := http.NewRequestWithContext(requestCtx, "GET", params.URL, nil)
			if err != nil {
				return fantasy.ToolResponse{}, fmt.Errorf("failed to create request: %w", err)
			}

			req.Header.Set("User-Agent", "nexora/1.0")

			resp, err := client.Do(req)
			if err != nil {
				// Return friendly error message instead of propagating context errors
				if ctx.Err() != nil || strings.Contains(err.Error(), "deadline exceeded") || strings.Contains(err.Error(), "context canceled") {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("⏱️ Request to %s timed out after %ds. The server may be slow or unreachable. Try increasing timeout or check if the URL is correct.", params.URL, params.Timeout)), nil
				}
				return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to fetch URL: %s", err.Error())), nil
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("Request failed with status code: %d", resp.StatusCode)), nil
			}

			maxSize := int64(5 * 1024 * 1024) // 5MB
			body, err := io.ReadAll(io.LimitReader(resp.Body, maxSize))
			if err != nil {
				return fantasy.NewTextErrorResponse("Failed to read response body: " + err.Error()), nil
			}

			content := string(body)

			isValidUtf8 := utf8.ValidString(content)
			if !isValidUtf8 {
				return fantasy.NewTextErrorResponse("Response content is not valid UTF-8"), nil
			}
			contentType := resp.Header.Get("Content-Type")

			switch format {
			case "text":
				if strings.Contains(contentType, "text/html") {
					text, err := extractTextFromHTML(content)
					if err != nil {
						return fantasy.NewTextErrorResponse("Failed to extract text from HTML: " + err.Error()), nil
					}
					content = text
				}

			case "markdown":
				if strings.Contains(contentType, "text/html") {
					markdown, err := convertHTMLToMarkdown(content)
					if err != nil {
						return fantasy.NewTextErrorResponse("Failed to convert HTML to Markdown: " + err.Error()), nil
					}
					content = markdown
				}

				content = "```\n" + content + "\n```"

			case "html":
				if strings.Contains(contentType, "text/html") {
					doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
					if err != nil {
						return fantasy.NewTextErrorResponse("Failed to parse HTML: " + err.Error()), nil
					}
					body, err := doc.Find("body").Html()
					if err != nil {
						return fantasy.NewTextErrorResponse("Failed to extract body from HTML: " + err.Error()), nil
					}
					if body == "" {
						return fantasy.NewTextErrorResponse("No body content found in HTML"), nil
					}
					content = "<html>\n<body>\n" + body + "\n</body>\n</html>"
				}
			}

			// Context-aware content handling
			tokenCount := countTokens(content)

			// Check if content fits within context limit
			if tokenCount <= DefaultContextLimit {
				// Content fits in context - return directly
				contentSize := int64(len(content))
				if contentSize > MaxReadSize {
					content = content[:MaxReadSize]
					content += fmt.Sprintf("\n\n[Content truncated to %d bytes]", MaxReadSize)
				}
				return fantasy.NewTextResponse(content), nil
			}

			// Content too large for context - write to tmp file
			tmpDir := filepath.Join(workingDir, "nexora-fetch-"+sessionID)
			if err := os.MkdirAll(tmpDir, 0o700); err != nil {
				return fantasy.NewTextErrorResponse("Failed to create tmp directory: " + err.Error()), nil
			}

			// Generate filename
			timestamp := time.Now().Format("20060102150405.000000000")
			safeURL := sanitizeURLForFilename(params.URL)
			tmpFilename := fmt.Sprintf("%s-%s.txt", timestamp, safeURL)
			tmpPath := filepath.Join(tmpDir, tmpFilename)

			// Write content to file
			if err := os.WriteFile(tmpPath, []byte(content), 0o600); err != nil {
				return fantasy.NewTextErrorResponse("Failed to write content to file: " + err.Error()), nil
			}

			result := fmt.Sprintf("Content fetched from %s (%d tokens, %d bytes).\n\nContent saved to: %s\n\nUse the view and grep tools to analyze this file. Content was too large for context (%d tokens > %d limit).",
				params.URL,
				tokenCount,
				len(content),
				tmpPath,
				tokenCount,
				DefaultContextLimit)

			return fantasy.NewTextResponse(result), nil
		})
}

func extractTextFromHTML(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", err
	}

	text := doc.Find("body").Text()
	text = strings.Join(strings.Fields(text), " ")

	return text, nil
}

func convertHTMLToMarkdown(html string) (string, error) {
	converter := md.NewConverter("", true, nil)

	markdown, err := converter.ConvertString(html)
	if err != nil {
		return "", err
	}

	return markdown, nil
}

// countTokens estimates the number of tokens in a string
// This is a simple approximation - in production, use a proper tokenizer
func countTokens(s string) int {
	// Rough approximation: 4 characters per token on average
	// This is much faster than proper tokenization
	wordCount := 0
	inWord := false

	for _, r := range s {
		if unicode.IsSpace(r) {
			inWord = false
		} else if !inWord {
			inWord = true
			wordCount++
		}
	}

	// Add overhead for punctuation and special characters
	punctuationCount := strings.Count(s, ",") + strings.Count(s, ".") +
		strings.Count(s, "!") + strings.Count(s, "?") +
		strings.Count(s, ";") + strings.Count(s, ":")

	// Estimate: words + punctuation/2 (approximate)
	return wordCount + punctuationCount/2
}

// sanitizeURLForFilename converts a URL to a safe filename
func sanitizeURLForFilename(url string) string {
	// Remove protocol
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimPrefix(url, "http://")

	// Replace common URL characters with underscores
	replacer := strings.NewReplacer(
		"/", "_",
		":", "_",
		".", "_",
		"?", "_",
		"&", "_",
		"=", "_",
		"@", "_",
	)
	return replacer.Replace(url)
}

// generateTmpFilename generates a predictable tmp filename for testing
func generateTmpFilename(sessionID string, url string) string {
	timestamp := time.Now().Format("20060102150405")
	safeURL := sanitizeURLForFilename(url)
	return fmt.Sprintf("/tmp/nexora-fetch-%s/%s-%s.txt", sessionID, timestamp, safeURL)
}

// NewWebFetchTool creates a simple web fetch tool for sub-agents (no permissions needed).
// This is a simplified version of NewFetchTool that doesn't require permissions,
// making it suitable for use by sub-agents in delegated tasks.
func NewWebFetchTool(workingDir string, client *http.Client) fantasy.AgentTool {
	if client == nil {
		client = &http.Client{
			Timeout: 30 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		}
	}

	return fantasy.NewParallelAgentTool(
		WebFetchToolName,
		string(webFetchToolDescription),
		func(ctx context.Context, params WebFetchParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.URL == "" {
				return fantasy.NewTextErrorResponse("url is required"), nil
			}

			content, err := FetchURLAndConvert(ctx, client, params.URL)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to fetch URL: %s", err)), nil
			}

			hasLargeContent := len(content) > LargeContentThreshold
			var result strings.Builder

			if hasLargeContent {
				tempFile, err := os.CreateTemp(workingDir, "page-*.md")
				if err != nil {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to create temporary file: %s", err)), nil
				}
				tempFilePath := tempFile.Name()

				if _, err := tempFile.WriteString(content); err != nil {
					_ = tempFile.Close() // Best effort close
					return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to write content to file: %s", err)), nil
				}
				if err := tempFile.Close(); err != nil {
					return fantasy.NewTextErrorResponse(fmt.Sprintf("Failed to close temporary file: %s", err)), nil
				}

				result.WriteString(fmt.Sprintf("Fetched content from %s (large page)\n\n", params.URL))
				result.WriteString(fmt.Sprintf("Content saved to: %s\n\n", tempFilePath))
				result.WriteString("Use the view and grep tools to analyze this file.")
			} else {
				result.WriteString(fmt.Sprintf("Fetched content from %s:\n\n", params.URL))
				result.WriteString(content)
			}

			return fantasy.NewTextResponse(result.String()), nil
		})
}
