package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// resetFlags resets all flags to their default values
func resetFlags(t *testing.T) {
	t.Helper()
	require.NoError(t, resetCmd.Flags().Set("force", "false"))
	require.NoError(t, resetCmd.Flags().Set("all", "false"))
	require.NoError(t, resetCmd.Flags().Set("clear-sessions", "false"))
}

// captureOutput captures stdout during a test
func captureOutput(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

// withStdin runs fn with stdin set to a reader containing the given input
func withStdin(t *testing.T, input string, fn func()) string {
	t.Helper()
	oldStdin := os.Stdin
	r, w, _ := os.Pipe()
	os.Stdin = r

	go func() {
		w.WriteString(input)
		w.Close()
	}()

	output := captureOutput(t, fn)

	os.Stdin = oldStdin
	return output
}

func TestResetCmd_Flags(t *testing.T) {
	// Verify the reset command has expected flags
	require.NotNil(t, resetCmd.Flags().Lookup("force"), "expected --force flag")
	require.NotNil(t, resetCmd.Flags().Lookup("all"), "expected --all flag")
	require.NotNil(t, resetCmd.Flags().Lookup("clear-sessions"), "expected --clear-sessions flag")

	// Verify shorthand flags
	forceFlag := resetCmd.Flags().Lookup("force")
	require.Equal(t, "f", forceFlag.Shorthand, "expected -f shorthand for --force")

	allFlag := resetCmd.Flags().Lookup("all")
	require.Equal(t, "a", allFlag.Shorthand, "expected -a shorthand for --all")
}

func TestResetCmd_CancelConfirmation(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config", "nexora")
	dataDir := filepath.Join(tempDir, "data", "nexora")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.MkdirAll(dataDir, 0755))

	configFile := filepath.Join(configDir, "nexora.json")
	require.NoError(t, os.WriteFile(configFile, []byte("{}"), 0644))

	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	oldDataHome := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, "config"))
	os.Setenv("XDG_DATA_HOME", filepath.Join(tempDir, "data"))
	defer func() {
		os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
		os.Setenv("XDG_DATA_HOME", oldDataHome)
		resetFlags(t)
	}()

	var runErr error
	output := withStdin(t, "n\n", func() {
		runErr = resetCmd.RunE(resetCmd, nil)
	})

	require.NoError(t, runErr)
	require.Contains(t, output, "Reset cancelled", "expected cancellation message")

	// Config should still exist
	_, err := os.Stat(configFile)
	require.NoError(t, err, "config file should still exist after cancellation")
}

func TestResetCmd_ForceFlag(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config", "nexora")
	dataDir := filepath.Join(tempDir, "data", "nexora")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.MkdirAll(dataDir, 0755))

	configFile := filepath.Join(configDir, "nexora.json")
	require.NoError(t, os.WriteFile(configFile, []byte("{}"), 0644))
	initFlagFile := filepath.Join(dataDir, "init")
	require.NoError(t, os.WriteFile(initFlagFile, []byte(""), 0644))

	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	oldDataHome := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, "config"))
	os.Setenv("XDG_DATA_HOME", filepath.Join(tempDir, "data"))
	defer func() {
		os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
		os.Setenv("XDG_DATA_HOME", oldDataHome)
		resetFlags(t)
	}()

	require.NoError(t, resetCmd.Flags().Set("force", "true"))

	output := captureOutput(t, func() {
		err := resetCmd.RunE(resetCmd, nil)
		require.NoError(t, err)
	})

	require.Contains(t, output, "Reset complete", "expected completion message")

	// Config directory should be deleted
	_, err := os.Stat(configDir)
	require.True(t, os.IsNotExist(err), "config dir should be deleted")

	// Init flag should be deleted
	_, err = os.Stat(initFlagFile)
	require.True(t, os.IsNotExist(err), "init flag should be deleted")

	// Data directory should still exist
	_, err = os.Stat(dataDir)
	require.NoError(t, err, "data directory should still exist")
}

func TestResetCmd_AllFlag(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config", "nexora")
	dataDir := filepath.Join(tempDir, "data", "nexora")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.MkdirAll(dataDir, 0755))

	configFile := filepath.Join(configDir, "nexora.json")
	require.NoError(t, os.WriteFile(configFile, []byte("{}"), 0644))
	dbFile := filepath.Join(dataDir, "nexora.db")
	require.NoError(t, os.WriteFile(dbFile, []byte("sqlite"), 0644))

	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	oldDataHome := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, "config"))
	os.Setenv("XDG_DATA_HOME", filepath.Join(tempDir, "data"))
	defer func() {
		os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
		os.Setenv("XDG_DATA_HOME", oldDataHome)
		resetFlags(t)
	}()

	require.NoError(t, resetCmd.Flags().Set("force", "true"))
	require.NoError(t, resetCmd.Flags().Set("all", "true"))

	output := captureOutput(t, func() {
		err := resetCmd.RunE(resetCmd, nil)
		require.NoError(t, err)
	})

	require.Contains(t, output, "Removed data directory", "expected data dir removal message")

	// Both directories should be deleted
	_, err := os.Stat(configDir)
	require.True(t, os.IsNotExist(err), "config dir should be deleted")
	_, err = os.Stat(dataDir)
	require.True(t, os.IsNotExist(err), "data dir should be deleted")
}

func TestResetCmd_ClearSessionsFlag(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config", "nexora")
	dataDir := filepath.Join(tempDir, "data", "nexora")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.MkdirAll(dataDir, 0755))

	configFile := filepath.Join(configDir, "nexora.json")
	require.NoError(t, os.WriteFile(configFile, []byte("{}"), 0644))
	dbFile := filepath.Join(dataDir, "nexora.db")
	require.NoError(t, os.WriteFile(dbFile, []byte("sqlite"), 0644))
	walFile := filepath.Join(dataDir, "nexora.db-wal")
	require.NoError(t, os.WriteFile(walFile, []byte("wal"), 0644))
	shmFile := filepath.Join(dataDir, "nexora.db-shm")
	require.NoError(t, os.WriteFile(shmFile, []byte("shm"), 0644))
	initFlagFile := filepath.Join(dataDir, "init")
	require.NoError(t, os.WriteFile(initFlagFile, []byte(""), 0644))

	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	oldDataHome := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, "config"))
	os.Setenv("XDG_DATA_HOME", filepath.Join(tempDir, "data"))
	defer func() {
		os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
		os.Setenv("XDG_DATA_HOME", oldDataHome)
		resetFlags(t)
	}()

	require.NoError(t, resetCmd.Flags().Set("force", "true"))
	require.NoError(t, resetCmd.Flags().Set("clear-sessions", "true"))

	output := captureOutput(t, func() {
		err := resetCmd.RunE(resetCmd, nil)
		require.NoError(t, err)
	})

	require.Contains(t, output, "Removed database", "expected database removal message")

	// Database files should be deleted
	_, err := os.Stat(dbFile)
	require.True(t, os.IsNotExist(err), "database file should be deleted")
	_, err = os.Stat(walFile)
	require.True(t, os.IsNotExist(err), "wal file should be deleted")
	_, err = os.Stat(shmFile)
	require.True(t, os.IsNotExist(err), "shm file should be deleted")

	// Data directory should still exist
	_, err = os.Stat(dataDir)
	require.NoError(t, err, "data directory should still exist")
}

func TestResetCmd_YesConfirmation(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config", "nexora")
	dataDir := filepath.Join(tempDir, "data", "nexora")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.MkdirAll(dataDir, 0755))

	configFile := filepath.Join(configDir, "nexora.json")
	require.NoError(t, os.WriteFile(configFile, []byte("{}"), 0644))
	initFlagFile := filepath.Join(dataDir, "init")
	require.NoError(t, os.WriteFile(initFlagFile, []byte(""), 0644))

	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	oldDataHome := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, "config"))
	os.Setenv("XDG_DATA_HOME", filepath.Join(tempDir, "data"))
	defer func() {
		os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
		os.Setenv("XDG_DATA_HOME", oldDataHome)
		resetFlags(t)
	}()

	var runErr error
	output := withStdin(t, "yes\n", func() {
		runErr = resetCmd.RunE(resetCmd, nil)
	})

	require.NoError(t, runErr)
	require.Contains(t, output, "Reset complete", "expected completion message")

	// Config directory should be deleted
	_, err := os.Stat(configDir)
	require.True(t, os.IsNotExist(err), "config dir should be deleted after yes confirmation")
}

func TestResetCmd_YConfirmation(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config", "nexora")
	dataDir := filepath.Join(tempDir, "data", "nexora")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.MkdirAll(dataDir, 0755))

	configFile := filepath.Join(configDir, "nexora.json")
	require.NoError(t, os.WriteFile(configFile, []byte("{}"), 0644))

	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	oldDataHome := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, "config"))
	os.Setenv("XDG_DATA_HOME", filepath.Join(tempDir, "data"))
	defer func() {
		os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
		os.Setenv("XDG_DATA_HOME", oldDataHome)
		resetFlags(t)
	}()

	var runErr error
	output := withStdin(t, "y\n", func() {
		runErr = resetCmd.RunE(resetCmd, nil)
	})

	require.NoError(t, runErr)
	require.Contains(t, output, "Reset complete", "expected completion message with 'y' confirmation")

	// Config directory should be deleted
	_, err := os.Stat(configDir)
	require.True(t, os.IsNotExist(err), "config dir should be deleted after 'y' confirmation")
}

func TestResetCmd_OutputDisplay(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config", "nexora")
	dataDir := filepath.Join(tempDir, "data", "nexora")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.MkdirAll(dataDir, 0755))

	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	oldDataHome := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, "config"))
	os.Setenv("XDG_DATA_HOME", filepath.Join(tempDir, "data"))
	defer func() {
		os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
		os.Setenv("XDG_DATA_HOME", oldDataHome)
		resetFlags(t)
	}()

	output := withStdin(t, "n\n", func() {
		_ = resetCmd.RunE(resetCmd, nil)
	})

	// Check that the output shows what will be deleted
	require.Contains(t, output, "This will reset Nexora", "expected reset description")
	require.Contains(t, output, "The following will be deleted", "expected deletion list header")
	require.Contains(t, output, "Config directory", "expected config dir in output")
}

func TestResetCmd_AllFlagOutput(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config", "nexora")
	dataDir := filepath.Join(tempDir, "data", "nexora")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.MkdirAll(dataDir, 0755))

	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	oldDataHome := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, "config"))
	os.Setenv("XDG_DATA_HOME", filepath.Join(tempDir, "data"))
	defer func() {
		os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
		os.Setenv("XDG_DATA_HOME", oldDataHome)
		resetFlags(t)
	}()

	require.NoError(t, resetCmd.Flags().Set("all", "true"))

	output := withStdin(t, "n\n", func() {
		_ = resetCmd.RunE(resetCmd, nil)
	})

	// With --all flag, should show data directory will be deleted
	require.Contains(t, output, "Data directory", "expected data dir in output with --all")
	require.Contains(t, output, "includes database, sessions, logs", "expected full removal notice")
}

func TestResetCmd_ClearSessionsFlagOutput(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config", "nexora")
	dataDir := filepath.Join(tempDir, "data", "nexora")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.MkdirAll(dataDir, 0755))

	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	oldDataHome := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, "config"))
	os.Setenv("XDG_DATA_HOME", filepath.Join(tempDir, "data"))
	defer func() {
		os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
		os.Setenv("XDG_DATA_HOME", oldDataHome)
		resetFlags(t)
	}()

	require.NoError(t, resetCmd.Flags().Set("clear-sessions", "true"))

	output := withStdin(t, "n\n", func() {
		_ = resetCmd.RunE(resetCmd, nil)
	})

	// With --clear-sessions flag, should show sessions will be cleared
	require.Contains(t, output, "Sessions and init flag", "expected sessions notice with --clear-sessions")
}

func TestResetCmd_NonexistentDirectories(t *testing.T) {
	tempDir := t.TempDir()

	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	oldDataHome := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, "nonexistent_config"))
	os.Setenv("XDG_DATA_HOME", filepath.Join(tempDir, "nonexistent_data"))
	defer func() {
		os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
		os.Setenv("XDG_DATA_HOME", oldDataHome)
		resetFlags(t)
	}()

	require.NoError(t, resetCmd.Flags().Set("force", "true"))

	output := captureOutput(t, func() {
		err := resetCmd.RunE(resetCmd, nil)
		require.NoError(t, err)
	})

	// Should not error when directories don't exist
	require.Contains(t, output, "Reset complete", "expected completion even when dirs don't exist")
}

func TestResetCmd_CommandMetadata(t *testing.T) {
	require.Equal(t, "reset", resetCmd.Use)
	require.Equal(t, "Reset Nexora to initial state", resetCmd.Short)
	require.NotEmpty(t, resetCmd.Long, "expected long description")
	require.NotEmpty(t, resetCmd.Example, "expected example usage")
	require.Contains(t, resetCmd.Example, "nexora reset")
}

func TestResetCmd_EmptyResponse(t *testing.T) {
	tempDir := t.TempDir()
	configDir := filepath.Join(tempDir, "config", "nexora")
	dataDir := filepath.Join(tempDir, "data", "nexora")
	require.NoError(t, os.MkdirAll(configDir, 0755))
	require.NoError(t, os.MkdirAll(dataDir, 0755))

	configFile := filepath.Join(configDir, "nexora.json")
	require.NoError(t, os.WriteFile(configFile, []byte("{}"), 0644))

	oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
	oldDataHome := os.Getenv("XDG_DATA_HOME")
	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, "config"))
	os.Setenv("XDG_DATA_HOME", filepath.Join(tempDir, "data"))
	defer func() {
		os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
		os.Setenv("XDG_DATA_HOME", oldDataHome)
		resetFlags(t)
	}()

	var runErr error
	output := withStdin(t, "\n", func() {
		runErr = resetCmd.RunE(resetCmd, nil)
	})

	require.NoError(t, runErr)
	// Empty response (just Enter) should cancel
	require.Contains(t, output, "Reset cancelled", "expected cancellation on empty response")

	// Config should still exist
	_, err := os.Stat(configFile)
	require.NoError(t, err, "config file should still exist after empty response")
}

func TestResetCmd_CaseInsensitiveConfirmation(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"Y\n", "Reset complete"},
		{"YES\n", "Reset complete"},
		{"Yes\n", "Reset complete"},
		{"N\n", "Reset cancelled"},
		{"NO\n", "Reset cancelled"},
		{"no\n", "Reset cancelled"},
	}

	for _, tc := range testCases {
		t.Run(strings.TrimSpace(tc.input), func(t *testing.T) {
			tempDir := t.TempDir()
			configDir := filepath.Join(tempDir, "config", "nexora")
			dataDir := filepath.Join(tempDir, "data", "nexora")
			require.NoError(t, os.MkdirAll(configDir, 0755))
			require.NoError(t, os.MkdirAll(dataDir, 0755))

			configFile := filepath.Join(configDir, "nexora.json")
			require.NoError(t, os.WriteFile(configFile, []byte("{}"), 0644))

			oldConfigHome := os.Getenv("XDG_CONFIG_HOME")
			oldDataHome := os.Getenv("XDG_DATA_HOME")
			os.Setenv("XDG_CONFIG_HOME", filepath.Join(tempDir, "config"))
			os.Setenv("XDG_DATA_HOME", filepath.Join(tempDir, "data"))
			defer func() {
				os.Setenv("XDG_CONFIG_HOME", oldConfigHome)
				os.Setenv("XDG_DATA_HOME", oldDataHome)
				resetFlags(t)
			}()

			var runErr error
			output := withStdin(t, tc.input, func() {
				runErr = resetCmd.RunE(resetCmd, nil)
			})

			require.NoError(t, runErr)
			require.Contains(t, output, tc.expected)
		})
	}
}
