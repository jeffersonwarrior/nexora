package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIndexCmd(t *testing.T) {
	// Save current WD
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "nexora-index-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test Go file
	testFile := filepath.Join(tempDir, "test.go")
	testContent := `
package main

import "fmt"

// TestFunction is a test function
func TestFunction(name string) string {
	return "Hello, " + name
}

// TestStruct is a test struct
type TestStruct struct {
	Name string
	Age  int
}
`
	err = os.WriteFile(testFile, []byte(testContent), 0o644)
	require.NoError(t, err)

	// Create temporary database path
	dbPath := filepath.Join(tempDir, "test.db")

	// Execute index command via rootCmd
	rootCmd.SetArgs([]string{"index", "--cwd", tempDir, "--output", dbPath, "--embeddings"})
	err = rootCmd.Execute()
	require.NoError(t, err)

	// Verify database was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Logf("Database not found at %s", dbPath)
		// List files in tempDir
		files, _ := os.ReadDir(tempDir)
		for _, f := range files {
			t.Logf("Found file: %s", f.Name())
		}
		cwd, _ := os.Getwd()
		t.Logf("Current WD: %s", cwd)
	}
	_, err = os.Stat(dbPath)
	require.NoError(t, err)

	// Verify database has content
	dbSize, err := os.Stat(dbPath)
	require.NoError(t, err)
	require.Greater(t, dbSize.Size(), int64(0), "Database should not be empty")
}

func TestQueryCmd(t *testing.T) {
	// Save current WD
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	// Create temporary directory for testing
	tempDir, err := os.MkdirTemp("", "nexora-query-test")
	require.NoError(t, err)
	defer os.RemoveAll(tempDir)

	// Create a test Go file
	testFile := filepath.Join(tempDir, "test.go")
	testContent := `
package main

import "fmt"

// HelloFunction says hello
func HelloFunction(name string) string {
	return fmt.Sprintf("Hello, %s!", name)
}
`
	err = os.WriteFile(testFile, []byte(testContent), 0o644)
	require.NoError(t, err)

	// Create temporary database path
	dbPath := filepath.Join(tempDir, "test.db")

	// First index the file via rootCmd
	rootCmd.SetArgs([]string{"index", "--cwd", tempDir, "--output", dbPath})
	err = rootCmd.Execute()
	require.NoError(t, err)

	// Test query command via rootCmd
	// Note: query command might be 'query' or just default args?
	// root.go adds queryCmd. Check query_cmd.go usage.
	// Assuming "query" subcommand exists.
	rootCmd.SetArgs([]string{"query", "HelloFunction", "--cwd", tempDir, "--database", dbPath, "--limit", "1"})
	err = rootCmd.Execute()
	require.NoError(t, err)
}

func TestIndexCmdWithInvalidPath(t *testing.T) {
	// Save current WD
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	// Test with non-existent directory
	// Use rootCmd to execute index command to ensure proper flag/arg parsing
	rootCmd.SetArgs([]string{"index", "/non/existent/path", "--cwd", "."})

	err = rootCmd.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "path does not exist")
}

func TestQueryCmdWithoutDatabase(t *testing.T) {
	// Save current WD
	originalWd, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalWd)

	// Test query without database
	rootCmd.SetArgs([]string{"query", "test", "--database", "/non/existent.db", "--cwd", "."})

	err = rootCmd.Execute()
	require.Error(t, err)
	// Update expectation to match actual error
	require.Contains(t, err.Error(), "index database not found")
}
