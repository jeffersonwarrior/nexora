package tools

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"charm.land/fantasy"
	"github.com/nexora/cli/internal/fsext"
	"github.com/nexora/cli/internal/permission"
)

const FindToolName = "find"

// find.md contains the documentation for the Find tool
//
//go:embed find.md
var findDescription []byte

type FindParams struct {
	Pattern    string `json:"pattern,omitempty" description:"Filename pattern to search for (supports glob patterns like *.go)"`
	Path       string `json:"path,omitempty" description:"The directory to search in. Defaults to current working directory."`
	Type       string `json:"type,omitempty" description:"Type of file to find: 'f' for files, 'd' for directories, empty for both"`
	Contains   string `json:"contains,omitempty" description:"Text content to search for within files (slower search)"`
	MaxResults int    `json:"max_results,omitempty" description:"Maximum number of results to return (default: 100)"`
}

type FindPermissionsParams struct {
	Pattern    string
	Path       string
	Type       string
	Contains   string
	MaxResults int
}

type FindResult struct {
	Path     string `json:"path"`
	Type     string `json:"type"`
	Size     int64  `json:"size,omitempty"`
	Modified string `json:"modified,omitempty"`
}

type FindResponseMetadata struct {
	NumberOfResults int    `json:"number_of_results"`
	Truncated       bool   `json:"truncated"`
	ToolUsed        string `json:"tool_used"`
}

func NewFindTool(permissions permission.Service, workingDir string) fantasy.AgentTool {
	return fantasy.NewAgentTool(
		FindToolName,
		string(findDescription),
		func(ctx context.Context, params FindParams, call fantasy.ToolCall) (fantasy.ToolResponse, error) {
			// Set defaults
			if params.MaxResults <= 0 || params.MaxResults > 1000 {
				params.MaxResults = 100
			}

			searchPath := params.Path
			if searchPath == "" {
				searchPath = workingDir
			}
			absPath, err := filepath.Abs(searchPath)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("invalid path: %v", err)), nil
			}

			// Permission checking - currently handled at higher level by the agent framework
			// Future enhancement: Add explicit permission validation for security-sensitive paths
			// This could include:
			// - Checking file system permissions before search
			// - Validating path doesn't access system-critical directories
			// - Implementing user denylist patterns for sensitive locations

			// Perform the search
			results, toolUsed, err := findFiles(ctx, params, absPath)
			if err != nil {
				return fantasy.NewTextErrorResponse(fmt.Sprintf("search failed: %v", err)), nil
			}

			// Format output
			if len(results) == 0 {
				return fantasy.WithResponseMetadata(
					fantasy.NewTextResponse("No matching files found"),
					FindResponseMetadata{
						NumberOfResults: 0,
						Truncated:       false,
						ToolUsed:        toolUsed,
					},
				), nil
			}

			var output strings.Builder
			fmt.Fprintf(&output, "Found %d matching item(s):\n\n", len(results))

			for _, result := range results {
				icon := "📄"
				if result.Type == "d" {
					icon = "📁"
				}
				fmt.Fprintf(&output, "%s %s", icon, result.Path)
				if result.Type == "f" {
					fmt.Fprintf(&output, " (%d bytes)", result.Size)
				}
				if result.Modified != "" {
					fmt.Fprintf(&output, " [modified: %s]", result.Modified)
				}
				output.WriteString("\n")
			}

			truncated := len(results) >= params.MaxResults
			if truncated {
				output.WriteString(fmt.Sprintf("\n(Results truncated at %d items. Use max_results parameter to see more.)", params.MaxResults))
			}

			return fantasy.WithResponseMetadata(
				fantasy.NewTextResponse(output.String()),
				FindResponseMetadata{
					NumberOfResults: len(results),
					Truncated:       truncated,
					ToolUsed:        toolUsed,
				},
			), nil
		},
	)
}

func findFiles(ctx context.Context, params FindParams, searchPath string) ([]FindResult, string, error) {
	// Try fd first if no content search is needed
	if params.Contains == "" {
		if results, err := findWithFD(ctx, params, searchPath); err == nil {
			return results, "fd", nil
		}
	}

	// Fallback to ripgrep for pattern/content search
	if results, err := findWithRipGrep(ctx, params, searchPath); err == nil {
		return results, "ripgrep", nil
	}

	// Final fallback to simple directory listing
	if results, err := findSimple(ctx, params, searchPath); err == nil {
		return results, "simple", nil
	}

	return nil, "none", fmt.Errorf("all search methods failed")
}

func findWithFD(ctx context.Context, params FindParams, searchPath string) ([]FindResult, error) {
	fdCmd := getFDCmd(ctx)
	if fdCmd == nil {
		return nil, fmt.Errorf("fd not available")
	}

	fdCmd.Dir = searchPath

	// Build fd arguments
	args := []string{"--absolute-path", "--color=never"}

	if params.Type == "f" {
		args = append(args, "--type", "file")
	} else if params.Type == "d" {
		args = append(args, "--type", "directory")
	}

	if params.Pattern != "" {
		args = append(args, params.Pattern)
	}

	// Limit results
	args = append(args, "--max-results", fmt.Sprintf("%d", params.MaxResults))

	fdCmd.Args = args

	// Execute with timeout
	output, err := fdCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("fd command failed: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	results := make([]FindResult, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		info, err := filepath.Abs(line)
		if err != nil {
			continue
		}

		fileInfo, err := os.Stat(info)
		if err != nil {
			continue
		}

		typ := "f"
		if fileInfo.IsDir() {
			typ = "d"
		}

		result := FindResult{
			Path: info,
			Type: typ,
			Size: fileInfo.Size(),
		}

		if !fileInfo.IsDir() {
			result.Modified = fileInfo.ModTime().Format("2006-01-02 15:04:05")
		}

		results = append(results, result)
	}

	// Sort by type (directories first) then by name
	sort.Slice(results, func(i, j int) bool {
		if results[i].Type != results[j].Type {
			return results[i].Type > results[j].Type // directories (d) > files (f)
		}
		return results[i].Path < results[j].Path
	})

	return results, nil
}

func findWithRipGrep(ctx context.Context, params FindParams, searchPath string) ([]FindResult, error) {
	// Build ripgrep command for pattern matching
	if params.Pattern == "" && params.Contains == "" {
		return nil, fmt.Errorf("no pattern or content specified for ripgrep search")
	}

	pattern := params.Pattern
	if params.Contains != "" {
		pattern = params.Contains
	}

	rgCmd := getRgSearchCmd(ctx, pattern, searchPath, "")
	if rgCmd == nil {
		return nil, fmt.Errorf("ripgrep not available")
	}

	// Configure ripgrep settings
	rgCmd.Args = append(rgCmd.Args,
		"--absolute-path",
		"--no-heading",
		"--with-filename",
		"--line-number",
		"--max-count", "1", // Only one line per file for speed
	)

	if params.Type == "f" {
		rgCmd.Args = append(rgCmd.Args, "--type", "file")
	} else if params.Type == "d" {
		// Ripgrep doesn't search directories, so skip
		return []FindResult{}, nil
	}

	// Execute with timeout
	output, err := rgCmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ripgrep command failed: %w", err)
	}

	// Parse output - we only need file paths
	seen := make(map[string]bool)
	results := make([]FindResult, 0, params.MaxResults)

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Extract path from ripgrep output (format: filename:line:content)
		parts := strings.SplitN(line, ":", 3)
		if len(parts) < 2 {
			continue
		}

		filePath := parts[0]
		if _, exists := seen[filePath]; exists {
			continue
		}

		seen[filePath] = true
		absPath, err := filepath.Abs(filePath)
		if err != nil {
			continue
		}

		fileInfo, err := os.Stat(absPath)
		if err != nil {
			continue
		}

		result := FindResult{
			Path:     absPath,
			Type:     "f",
			Size:     fileInfo.Size(),
			Modified: fileInfo.ModTime().Format("2006-01-02 15:04:05"),
		}

		results = append(results, result)

		if len(results) >= params.MaxResults {
			break
		}
	}

	return results, nil
}

func findSimple(ctx context.Context, params FindParams, searchPath string) ([]FindResult, error) {
	// Simple directory listing as fallback
	files, _, err := fsext.ListDirectory(searchPath, nil, 0, params.MaxResults)
	if err != nil {
		return nil, err
	}

	results := make([]FindResult, 0, len(files))

	for _, file := range files {
		absPath, err := filepath.Abs(file)
		if err != nil {
			continue
		}

		fileInfo, err := os.Stat(absPath)
		if err != nil {
			continue
		}

		typ := "f"
		if fileInfo.IsDir() {
			typ = "d"
		}

		// Apply pattern filter if specified
		if params.Pattern != "" {
			matched, err := filepath.Match(params.Pattern, filepath.Base(file))
			if err != nil || !matched {
				continue
			}
		}

		result := FindResult{
			Path: file,
			Type: typ,
			Size: fileInfo.Size(),
		}

		if !fileInfo.IsDir() {
			result.Modified = fileInfo.ModTime().Format("2006-01-02 15:04:05")
		}

		results = append(results, result)
	}

	return results, nil
}
