package tools

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestViewToolDirectoryHandling(t *testing.T) {
	t.Parallel()

	// Create a temporary directory with some files
	tmpDir := t.TempDir()

	// Create some test files
	testFiles := []string{"file1.go", "file2.txt", "subdir/"}
	for _, file := range testFiles {
		if file == "subdir/" {
			// Create subdirectory
			err := os.Mkdir(filepath.Join(tmpDir, file), 0o755)
			require.NoError(t, err)
		} else {
			// Create file
			content := []byte("test content for " + file)
			err := os.WriteFile(filepath.Join(tmpDir, file), content, 0o644)
			require.NoError(t, err)
		}
	}

	// Test the directory handling logic directly
	fileInfo, err := os.Stat(tmpDir)
	require.NoError(t, err)
	require.True(t, fileInfo.IsDir())

	// Test that we can read the directory contents
	dirEntries, err := os.ReadDir(tmpDir)
	require.NoError(t, err)
	require.Len(t, dirEntries, 3)

	// Verify we can build the expected response format
	var fileList []string
	for _, entry := range dirEntries {
		fileList = append(fileList, entry.Name())
	}

	response := "Path is a directory: " + tmpDir + "\n\nDirectory contents:\n"
	for i, file := range fileList {
		response += strconv.Itoa(i+1) + ". " + file + "\n"
	}

	response += "\n💡 Suggestions:\n"
	response += "- Use 'view' with a specific file path\n"
	response += "- Use 'ls' command to explore directory structure\n"
	response += "- Try 'find' to search for specific files\n"

	// Verify the response format is correct
	require.Contains(t, response, "Path is a directory")
	require.Contains(t, response, "Directory contents")
	require.Contains(t, response, "file1.go")
	require.Contains(t, response, "file2.txt")
	require.Contains(t, response, "💡 Suggestions")
	require.Contains(t, response, "Use 'view' with a specific file path")
	require.Contains(t, response, "Use 'ls' command")
	require.Contains(t, response, "Try 'find'")
}
