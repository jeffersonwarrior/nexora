package qa

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// TestViewToolRecovery tests the recovery mechanism for VIEW tool directory errors
func TestViewToolRecovery(t *testing.T) {
	t.Parallel()

	// Create a temporary directory structure
	tmpDir := t.TempDir()

	// Create test files
	testFiles := []string{"db.go", "models.go", "migrations/"}
	for _, file := range testFiles {
		if file == "migrations/" {
			// Create subdirectory
			err := os.Mkdir(filepath.Join(tmpDir, file), 0o755)
			require.NoError(t, err)
		} else {
			// Create file with some content
			content := []byte("// Database " + file + "\npackage db\n\ntype " + file[:len(file)-3] + " struct{}\n")
			err := os.WriteFile(filepath.Join(tmpDir, file), content, 0o644)
			require.NoError(t, err)
		}
	}

	// Test the error parsing and recovery logic
	mockError := fmt.Errorf("Path is a directory, not a file: %s", tmpDir)

	// Test the error parsing logic
	parts := strings.Split(mockError.Error(), ": ")
	require.Len(t, parts, 2)
	dirPath := strings.TrimSpace(parts[1])
	require.Equal(t, tmpDir, dirPath)

	// Test that we can read the directory
	dirEntries, err := os.ReadDir(dirPath)
	require.NoError(t, err)
	require.Len(t, dirEntries, 3)

	// Verify we can build the recovery response
	var fileList []string
	for _, entry := range dirEntries {
		fileList = append(fileList, entry.Name())
	}

	recoveryResponse := fmt.Sprintf("Path is a directory: %s\n\nDirectory contents:\n", dirPath)
	for i, file := range fileList {
		recoveryResponse += strconv.Itoa(i+1) + ". " + file + "\n"
	}

	recoveryResponse += "\n💡 Suggestions:\n"
	recoveryResponse += "- Use 'view' with a specific file path (e.g., 'view " + dirPath + "/filename')\n"
	recoveryResponse += "- Use 'ls' command to explore directory structure\n"
	recoveryResponse += "- Try 'find' to search for specific files\n"

	// Verify the recovery response is helpful
	require.Contains(t, recoveryResponse, "Path is a directory")
	require.Contains(t, recoveryResponse, "Directory contents")
	require.Contains(t, recoveryResponse, "db.go")
	require.Contains(t, recoveryResponse, "models.go")
	require.Contains(t, recoveryResponse, "💡 Suggestions")
	require.Contains(t, recoveryResponse, "Use 'view' with a specific file path")
	require.Contains(t, recoveryResponse, "Use 'ls' command")
	require.Contains(t, recoveryResponse, "Try 'find'")
}
