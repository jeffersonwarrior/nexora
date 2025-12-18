package tools

import (
	"bufio"
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"charm.land/fantasy"
	"github.com/nexora/nexora/internal/fsext"
)

// regexCache provides thread-safe caching of compiled regex patterns
type regexCache struct {
	cache map[string]*regexp.Regexp
	mu    sync.RWMutex
}

// newRegexCache creates a new regex cache
func newRegexCache() *regexCache {
	return &regexCache{
		cache: make(map[string]*regexp.Regexp),
	}
}

// get retrieves a compiled regex from cache or compiles and caches it
func (rc *regexCache) get(pattern string) (*regexp.Regexp, error) {
	// Try to get from cache first (read lock)
	rc.mu.RLock()
	if regex, exists := rc.cache[pattern]; exists {
		rc.mu.RUnlock()
		return regex, nil
	}
	rc.mu.RUnlock()

	// Compile the regex (write lock)
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// Double-check in case another goroutine compiled it while we waited
	if regex, exists := rc.cache[pattern]; exists {
		return regex, nil
	}

	// Compile and cache the regex
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}

	rc.cache[pattern] = regex
	return regex, nil
}

// Global regex cache instances
var (
	searchRegexCache = newRegexCache()
	globRegexCache   = newRegexCache()
	// Pre-compiled regex for glob conversion (used frequently)
	globBraceRegex = regexp.MustCompile(`\{([^}]+)\}`)
)

type GrepParams struct {
	Pattern     string `json:"pattern" description:"The regex pattern to search for in file contents"`
	Path        string `json:"path,omitempty" description:"The directory to search in. Defaults to the current working directory."`
	Include     string `json:"include,omitempty" description:"File pattern to include in the search (e.g. \"*.js\", \"*.{ts,tsx}\")"`
	LiteralText bool   `json:"literal_text,omitempty" description:"If true, the pattern will be treated as literal text with special regex characters escaped. Default is false."`
}

type grepMatch struct {
	path     string
	modTime  time.Time
	lineNum  int
	charNum  int
	lineText string
}

type GrepResponseMetadata struct {
	NumberOfMatches int  `json:"number_of_matches"`
	Truncated       bool `json:"truncated"`
}

const (
	GrepToolName        = "grep"
	maxGrepContentWidth = 500
)

//go:embed grep.md
var grepDescription []byte

// escapeRegexPattern escapes special regex characters so they're treated as literal characters
func escapeRegexPattern(pattern string) string {
	specialChars := []string{"\\", ".", "+", "*", "?", "(", ")", "[", "]", "{", "}", "^", "$", "|"}
	escaped := pattern

	for _, char := range specialChars {
		escaped = strings.ReplaceAll(escaped, char, "\\"+char)
	}

	return escaped
}

func NewGrepTool(workingDir string) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		GrepToolName,
		string(grepDescription),
		func(ctx context.Context, params GrepParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			if params.Pattern == "" {
				return fantasy.NewTextErrorResponse("pattern is required"), nil
			}

			// If literal_text is true, escape the pattern
			searchPattern := params.Pattern
			if params.LiteralText {
				searchPattern = escapeRegexPattern(params.Pattern)
			}

			searchPath := params.Path
			if searchPath == "" {
				searchPath = workingDir
			}

			// Create a child context with timeout for the entire search operation
			searchCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
			defer cancel()

			matches, truncated, err := searchFiles(searchCtx, searchPattern, searchPath, params.Include, 100)
			if err != nil {
				if searchCtx.Err() == context.DeadlineExceeded {
					return fantasy.NewTextErrorResponse("search timed out after 90 seconds - the search area may be too large or contain problematic files"), nil
				}
				if searchCtx.Err() == context.Canceled {
					return fantasy.NewTextErrorResponse("search was cancelled"), nil
				}
				return fantasy.NewTextErrorResponse(fmt.Sprintf("error searching files: %v", err)), nil
			}

			var output strings.Builder
			if len(matches) == 0 {
				output.WriteString("No files found")
			} else {
				fmt.Fprintf(&output, "Found %d matches\n", len(matches))

				currentFile := ""
				for _, match := range matches {
					if currentFile != match.path {
						if currentFile != "" {
							output.WriteString("\n")
						}
						currentFile = match.path
						fmt.Fprintf(&output, "%s:\n", filepath.ToSlash(match.path))
					}
					if match.lineNum > 0 {
						lineText := match.lineText
						if len(lineText) > maxGrepContentWidth {
							lineText = lineText[:maxGrepContentWidth] + "..."
						}
						if match.charNum > 0 {
							fmt.Fprintf(&output, "  Line %d, Char %d: %s\n", match.lineNum, match.charNum, lineText)
						} else {
							fmt.Fprintf(&output, "  Line %d: %s\n", match.lineNum, lineText)
						}
					} else {
						fmt.Fprintf(&output, "  %s\n", match.path)
					}
				}

				if truncated {
					output.WriteString("\n(Results are truncated. Consider using a more specific path or pattern.)")
				}
			}

			return fantasy.WithResponseMetadata(
				fantasy.NewTextResponse(output.String()),
				GrepResponseMetadata{
					NumberOfMatches: len(matches),
					Truncated:       truncated,
				},
			), nil
		})
}

func searchFiles(ctx context.Context, pattern, rootPath, include string, limit int) ([]grepMatch, bool, error) {
	matches, err := searchWithRipgrep(ctx, pattern, rootPath, include)
	if err != nil {
		matches, err = searchFilesWithRegex(pattern, rootPath, include)
		if err != nil {
			return nil, false, err
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].modTime.After(matches[j].modTime)
	})

	truncated := len(matches) > limit
	if truncated {
		matches = matches[:limit]
	}

	return matches, truncated, nil
}

func searchWithRipgrep(ctx context.Context, pattern, path, include string) ([]grepMatch, error) {
	// Create timeout context specifically for ripgrep command
	ripgrepCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	cmd := getRgSearchCmd(ripgrepCtx, pattern, path, include)
	if cmd == nil {
		return nil, fmt.Errorf("ripgrep not found in $PATH")
	}

	// Set safer parameters
	cmd.Args = append(cmd.Args, "--max-filesize", "50M")
	cmd.Args = append(cmd.Args, "--max-columns", "500")

	// Only add ignore files if they exist
	for _, ignoreFile := range []string{".gitignore", ".nexoraignore"} {
		ignorePath := filepath.Join(path, ignoreFile)
		if _, err := os.Stat(ignorePath); err == nil {
			cmd.Args = append(cmd.Args, "--ignore-file", ignorePath)
		}
	}

	// Use a separate goroutine to handle the command with timeout
	done := make(chan struct{})
	var (
		output []byte
		err    error
	)

	go func() {
		defer close(done)
		output, err = cmd.Output()
	}()

	// Wait for either completion, timeout, or cancellation
	select {
	case <-done:
		// Command completed normally
	case <-ripgrepCtx.Done():
		// Context was cancelled or timed out
		if cmd.Process != nil {
			_ = cmd.Process.Kill() // Force kill the process
		}
		return nil, fmt.Errorf("ripgrep search %w", ripgrepCtx.Err())
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return []grepMatch{}, nil
		}
		return nil, err
	}

	var matches []grepMatch
	for line := range bytes.SplitSeq(bytes.TrimSpace(output), []byte{'\n'}) {
		if len(line) == 0 {
			continue
		}
		var match ripgrepMatch
		if err := json.Unmarshal(line, &match); err != nil {
			continue
		}
		if match.Type != "match" {
			continue
		}
		for _, m := range match.Data.Submatches {
			fi, err := os.Stat(match.Data.Path.Text)
			if err != nil {
				continue // Skip files we can't access
			}
			matches = append(matches, grepMatch{
				path:     match.Data.Path.Text,
				modTime:  fi.ModTime(),
				lineNum:  match.Data.LineNumber,
				charNum:  m.Start + 1, // ensure 1-based
				lineText: strings.TrimSpace(match.Data.Lines.Text),
			})
			// only get the first match of each line
			break
		}
	}
	return matches, nil
}

type ripgrepMatch struct {
	Type string `json:"type"`
	Data struct {
		Path struct {
			Text string `json:"text"`
		} `json:"path"`
		Lines struct {
			Text string `json:"text"`
		} `json:"lines"`
		LineNumber int `json:"line_number"`
		Submatches []struct {
			Start int `json:"start"`
		} `json:"submatches"`
	} `json:"data"`
}

func searchFilesWithRegex(pattern, rootPath, include string) ([]grepMatch, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	return searchFilesWithRegexContext(ctx, pattern, rootPath, include)
}

// Separate function that accepts context
func searchFilesWithRegexContext(ctx context.Context, pattern, rootPath, include string) ([]grepMatch, error) {
	matches := []grepMatch{}

	// Use cached regex compilation
	regex, err := searchRegexCache.get(pattern)
	if err != nil {
		return nil, fmt.Errorf("invalid regex pattern: %w", err)
	}

	var includePattern *regexp.Regexp
	if include != "" {
		regexPattern := globToRegex(include)
		includePattern, err = globRegexCache.get(regexPattern)
		if err != nil {
			return nil, fmt.Errorf("invalid include pattern: %w", err)
		}
	}

	// Create walker with gitignore and nexoraignore support
	walker := fsext.NewFastGlobWalker(rootPath)

	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		// Check context cancellation first
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			// Check if directory should be skipped
			if walker.ShouldSkip(path) {
				return filepath.SkipDir
			}
			return nil // Continue into directory
		}

		// Use walker's shouldSkip method for files
		if walker.ShouldSkip(path) {
			return nil
		}

		// Skip hidden files (starting with a dot) to match ripgrep's default behavior
		base := filepath.Base(path)
		if base != "." && strings.HasPrefix(base, ".") {
			return nil
		}

		// Skip very large files
		if info.Size() > 50*1024*1024 { // 50MB
			return nil
		}

		if includePattern != nil && !includePattern.MatchString(path) {
			return nil
		}

		// Check file with timeout
		fileCtx, fileCancel := context.WithTimeout(ctx, 5*time.Second)
		match, lineNum, charNum, lineText, err := fileContainsPatternWithContext(fileCtx, path, regex)
		fileCancel()

		if err != nil {
			return nil // Skip files we can't read
		}

		if match {
			matches = append(matches, grepMatch{
				path:     path,
				modTime:  info.ModTime(),
				lineNum:  lineNum,
				charNum:  charNum,
				lineText: lineText,
			})

			if len(matches) >= 200 {
				return filepath.SkipAll
			}
		}

		return nil
	})
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("file walk timed out after 60 seconds")
		}
		return nil, err
	}

	return matches, nil
}

func fileContainsPattern(filePath string, pattern *regexp.Regexp) (bool, int, int, string, error) {
	return fileContainsPatternWithContext(context.Background(), filePath, pattern)
}

// fileContainsMultiLinePattern searches for a pattern that may span multiple lines
// This is used for edit validation where the old_string might be a multi-line block
func fileContainsMultiLinePattern(filePath string, pattern *regexp.Regexp) (bool, int, int, string, error) {
	// Only search text files.
	if !isTextFile(filePath) {
		return false, 0, 0, "", nil
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return false, 0, 0, "", err
	}

	// Convert to string for regex matching
	fileContent := string(content)

	// Find the pattern in the entire content
	loc := pattern.FindStringIndex(fileContent)
	if loc == nil {
		return false, 0, 0, "", nil
	}

	// Find the line numbers where the match starts and ends
	lines := strings.Split(fileContent, "\n")
	startLine := 0
	endLine := 0
	charPos := 0

	for i, line := range lines {
		lineLength := len(line) + 1 // +1 for newline character
		if charPos <= loc[0] && loc[0] < charPos+lineLength {
			startLine = i + 1 // 1-based line number
		}
		if charPos <= loc[1] && loc[1] < charPos+lineLength {
			endLine = i + 1 // 1-based line number
		}
		charPos += lineLength
	}

	// Get the actual matched text
	matchedText := fileContent[loc[0]:loc[1]]

	return true, startLine, endLine, matchedText, nil
}

// ValidateEditString confirms that old_string exists in the file before edit
// Returns error with detailed information if the pattern is not found or found multiple times
func ValidateEditString(filePath string, oldString string, replaceAll bool) error {
	if oldString == "" {
		return nil // No validation needed for empty old_string (file creation)
	}

	// Create a literal pattern that matches the exact string
	pattern := regexp.MustCompile(regexp.QuoteMeta(oldString))

	// First try multi-line pattern matching for edit operations
	found, _, _, _, err := fileContainsMultiLinePattern(filePath, pattern)
	if err != nil {
		return fmt.Errorf("failed to validate edit string in file %s: %w", filePath, err)
	}

	if !found {
		return fmt.Errorf("old_string not found in file %s", filePath)
	}

	if !replaceAll {
		// Check for multiple occurrences using multi-line search
		// Read the entire file content to count occurrences
		content, err := os.ReadFile(filePath)
		if err != nil {
			return fmt.Errorf("failed to read file for multiple occurrence check: %w", err)
		}

		fileContent := string(content)
		occurrences := pattern.FindAllStringIndex(fileContent, -1)

		if len(occurrences) > 1 {
			// Get a sample of the second occurrence for the error message
			sampleText := fileContent[occurrences[1][0]:occurrences[1][1]]
			// Count lines to the second occurrence
			lines := strings.Split(fileContent[:occurrences[1][0]], "\n")
			return fmt.Errorf("old_string appears multiple times in file %s. Found at line %d: %s",
				filePath, len(lines), strings.TrimSpace(sampleText))
		}
	}

	return nil
}

// fileContainsPatternFromLine starts searching from a specific line number
func fileContainsPatternFromLine(filePath string, pattern *regexp.Regexp, startLine int) (bool, int, int, string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, 0, 0, "", err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		if lineNum < startLine {
			continue // Skip lines before startLine
		}

		lineContent := scanner.Text()
		if pattern.MatchString(lineContent) {
			colIndex := pattern.FindStringIndex(lineContent)
			if len(colIndex) > 0 {
				return true, lineNum, colIndex[0], lineContent, nil
			}
		}
	}

	return false, 0, 0, "", nil
}

func fileContainsPatternWithContext(ctx context.Context, filePath string, pattern *regexp.Regexp) (bool, int, int, string, error) {
	// Only search text files.
	if !isTextFile(filePath) {
		return false, 0, 0, "", nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return false, 0, 0, "", err
	}
	defer file.Close()

	// Use scanner with buffer size limits to handle very long lines
	scanner := bufio.NewScanner(file)
	buffer := make([]byte, 0, 64*1024) // 64KB initial buffer
	scanner.Buffer(buffer, 1024*1024)  // Max 1MB per line

	lineNum := 0
	for scanner.Scan() {
		// Check context cancellation frequently
		select {
		case <-ctx.Done():
			return false, 0, 0, "", ctx.Err()
		default:
		}

		lineNum++
		line := scanner.Text()
		if loc := pattern.FindStringIndex(line); loc != nil {
			charNum := loc[0] + 1
			return true, lineNum, charNum, line, nil
		}
	}

	return false, 0, 0, "", scanner.Err()
}

// isTextFile checks if a file is a text file by examining its MIME type.
func isTextFile(filePath string) bool {
	file, err := os.Open(filePath)
	if err != nil {
		return false
	}
	defer file.Close()

	// Read first 512 bytes for MIME type detection.
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return false
	}

	// Detect content type.
	contentType := http.DetectContentType(buffer[:n])

	// Check if it's a text MIME type.
	return strings.HasPrefix(contentType, "text/") ||
		contentType == "application/json" ||
		contentType == "application/xml" ||
		contentType == "application/javascript" ||
		contentType == "application/x-sh"
}

func globToRegex(glob string) string {
	regexPattern := strings.ReplaceAll(glob, ".", "\\.")
	regexPattern = strings.ReplaceAll(regexPattern, "*", ".*")
	regexPattern = strings.ReplaceAll(regexPattern, "?", ".")

	// Use pre-compiled regex instead of compiling each time
	regexPattern = globBraceRegex.ReplaceAllStringFunc(regexPattern, func(match string) string {
		inner := match[1 : len(match)-1]
		return "(" + strings.ReplaceAll(inner, ",", "|") + ")"
	})

	return regexPattern
}
