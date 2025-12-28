package tools

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestManageOutput_SmallContent(t *testing.T) {
	output := "Hello world"
	result := ManageOutput(output, "bash", "/tmp", "test-session")

	assert.Equal(t, "Hello world", result.Content)
	assert.False(t, result.WasTruncated)
	assert.False(t, result.WasWrittenToFile)
	assert.Equal(t, "returned", result.ActionTaken)
}

func TestManageOutput_EmptyContent(t *testing.T) {
	result := ManageOutput("", "bash", "/tmp", "test-session")

	assert.Equal(t, "", result.Content)
	assert.False(t, result.WasTruncated)
	assert.False(t, result.WasWrittenToFile)
	assert.Equal(t, "returned_empty", result.ActionTaken)
}

func TestManageOutput_Truncation(t *testing.T) {
	// Create large content that will exceed SmallOutputLimit (4000 tokens)
	largeContent := strings.Repeat("This is test content. ", 2000)

	result := ManageOutput(largeContent, "grep", "/tmp", "test-session")

	assert.True(t, result.WasTruncated)
	assert.NotEqual(t, largeContent, result.Content)
	assert.Less(t, len(result.Content), len(largeContent))
	assert.Equal(t, "truncated", result.ActionTaken)
}

func TestManageOutput_WriteToFile(t *testing.T) {
	// Create extremely large content
	largeContent := strings.Repeat("This is test content. ", 15000)

	tmpDir, err := os.MkdirTemp("", "nexora-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	result := ManageOutput(largeContent, "bash", tmpDir, "test-session-123")

	assert.True(t, result.WasWrittenToFile)
	assert.Contains(t, result.Content, result.FilePath)
	assert.Equal(t, "written_to_file", result.ActionTaken)

	// Verify file exists
	_, err = os.Stat(result.FilePath)
	assert.NoError(t, err)
}

func TestManageOutput_ToolSpecificLimits(t *testing.T) {
	// Create content that exceeds SmallOutputLimit but not MediumOutputLimit
	content := strings.Repeat("test ", 5000) // ~5000 tokens

	// bash uses MediumOutputLimit (12000)
	bashResult := ManageOutput(content, "bash", "/tmp", "test")
	assert.Equal(t, "returned", bashResult.ActionTaken)

	// grep uses SmallOutputLimit (4000)
	grepResult := ManageOutput(content, "grep", "/tmp", "test")
	assert.Equal(t, "truncated", grepResult.ActionTaken)
}

func TestTruncateToTokenLimit(t *testing.T) {
	content := strings.Repeat("word ", 1000)

	result := truncateToTokenLimit(content, 100)

	assert.True(t, countTokens(result) <= 100)
	assert.True(t, len(result) > 0)
}

func TestCountTokens_Accuracy(t *testing.T) {
	testCases := []struct {
		name     string
		content  string
		expected int
	}{
		{"Short text", "Hello world", 2},
		{"Paragraph", "This is a test paragraph with several words.", 10},
		{"With punctuation", "Hello, world! How are you?", 7},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := countTokens(tc.content)
			// Allow 50% variance for approximation
			assert.GreaterOrEqual(t, result, tc.expected-tc.expected/2)
			assert.LessOrEqual(t, result, tc.expected+tc.expected/2+5)
		})
	}
}

func TestWriteToTmpFile(t *testing.T) {
	content := "Test content for tmp file"
	tmpDir, err := os.MkdirTemp("", "nexora-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	filePath, err := writeToTmpFile(content, "test-tool", tmpDir, "test-session")
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(filePath)
	assert.NoError(t, err)

	// Verify content
	written, err := os.ReadFile(filePath)
	assert.NoError(t, err)
	assert.Equal(t, content, string(written))
}

func TestCleanupSessionOutputFiles(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "nexora-cleanup-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir) // Cleanup on failure

	// Create session tmp directory
	sessionTmpDir := filepath.Join(tmpDir, "nexora-output-test-session")
	err = os.MkdirAll(sessionTmpDir, 0755)
	require.NoError(t, err)

	// Create a file in it
	testFile := filepath.Join(sessionTmpDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test"), 0644)
	require.NoError(t, err)

	// Verify exists
	_, err = os.Stat(testFile)
	assert.NoError(t, err)

	// Cleanup
	err = CleanupSessionOutputFiles(tmpDir, "test-session")
	assert.NoError(t, err)

	// Verify deleted
	_, err = os.Stat(testFile)
	assert.True(t, os.IsNotExist(err))
	_, err = os.Stat(sessionTmpDir)
	assert.True(t, os.IsNotExist(err))
}

func TestFormatOutputForModel(t *testing.T) {
	result := OutputResult{
		WasTruncated: true,
		OriginalSize: 10000,
		TokenCount:   5000,
		ActionTaken:  "truncated",
	}

	msg := FormatOutputForModel(result, "bash")
	assert.Contains(t, msg, "truncated")
	assert.Contains(t, msg, "10000 bytes")

	result = OutputResult{
		WasWrittenToFile: true,
		FilePath:         "/tmp/test.txt",
		TokenCount:       60000,
		OriginalSize:     240000,
		ActionTaken:      "written_to_file",
	}

	msg = FormatOutputForModel(result, "bash")
	assert.Contains(t, msg, "written to file")
	assert.Contains(t, msg, "/tmp/test.txt")
}

func TestManageOutput_MediumLimit(t *testing.T) {
	// Content that fits in MediumOutputLimit (12000 tokens) but not Small (4000)
	// "test content " = ~2 tokens, so 5000 repetitions = ~10000 tokens
	content := strings.Repeat("test content ", 5000)

	// With MediumOutputLimit tool (bash)
	result := ManageOutput(content, "bash", "/tmp", "test")
	assert.Equal(t, "returned", result.ActionTaken)
}

func TestManageOutput_FilePathNaming(t *testing.T) {
	content := strings.Repeat("x ", 20000)
	tmpDir, err := os.MkdirTemp("", "nexora-filename-test-*")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	result := ManageOutput(content, "grep", tmpDir, "my-session")

	if result.WasWrittenToFile {
		assert.Contains(t, result.FilePath, "my-session")
		assert.Contains(t, result.FilePath, "grep")
		assert.True(t, strings.HasSuffix(result.FilePath, ".txt"))
	}
}

func TestGetOutputLimitForTool(t *testing.T) {
	testCases := []struct {
		tool     string
		expected int
	}{
		{"bash", MediumOutputLimit},
		{"shell", MediumOutputLimit},
		{"grep", SmallOutputLimit},
		{"search", SmallOutputLimit},
		{"view", SmallOutputLimit},
		{"ls", SmallOutputLimit},
		{"fetch", MediumOutputLimit},
		{"glob", SmallOutputLimit},
		{"unknown", SmallOutputLimit},
	}

	for _, tc := range testCases {
		t.Run(tc.tool, func(t *testing.T) {
			result := getOutputLimitForTool(tc.tool)
			assert.Equal(t, tc.expected, result)
		})
	}
}
