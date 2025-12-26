package cmd

import (
	"os"
	"path/filepath"
	"testing"

	uv "github.com/charmbracelet/ultraviolet"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestResolveCwd(t *testing.T) {
	// Save original directory
	originalDir, err := os.Getwd()
	require.NoError(t, err)
	defer os.Chdir(originalDir)

	t.Run("Returns current directory when no cwd flag", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("cwd", "", "")

		cwd, err := ResolveCwd(cmd)
		require.NoError(t, err)
		assert.NotEmpty(t, cwd)
	})

	t.Run("Changes to specified directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		cmd := &cobra.Command{}
		cmd.Flags().String("cwd", "", "")
		cmd.Flags().Set("cwd", tmpDir)

		cwd, err := ResolveCwd(cmd)
		require.NoError(t, err)
		assert.Equal(t, tmpDir, cwd)

		// Verify we're actually in that directory
		currentDir, err := os.Getwd()
		require.NoError(t, err)
		assert.Equal(t, tmpDir, currentDir)
	})

	t.Run("Returns error for non-existent directory", func(t *testing.T) {
		cmd := &cobra.Command{}
		cmd.Flags().String("cwd", "", "")
		cmd.Flags().Set("cwd", "/non/existent/path/12345")

		_, err := ResolveCwd(cmd)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to change directory")
	})
}

func TestCreateDotNexoraDir(t *testing.T) {
	t.Run("Creates directory successfully", func(t *testing.T) {
		tmpDir := t.TempDir()
		nexoraDir := filepath.Join(tmpDir, ".nexora")

		err := createDotNexoraDir(nexoraDir)
		require.NoError(t, err)

		// Verify directory exists
		info, err := os.Stat(nexoraDir)
		require.NoError(t, err)
		assert.True(t, info.IsDir())
	})

	t.Run("Creates .gitignore file", func(t *testing.T) {
		tmpDir := t.TempDir()
		nexoraDir := filepath.Join(tmpDir, ".nexora")

		err := createDotNexoraDir(nexoraDir)
		require.NoError(t, err)

		// Verify .gitignore exists
		gitignorePath := filepath.Join(nexoraDir, ".gitignore")
		content, err := os.ReadFile(gitignorePath)
		require.NoError(t, err)
		assert.Equal(t, "*\n", string(content))
	})

	t.Run("Does not overwrite existing .gitignore", func(t *testing.T) {
		tmpDir := t.TempDir()
		nexoraDir := filepath.Join(tmpDir, ".nexora")

		// Create directory and custom .gitignore
		err := os.MkdirAll(nexoraDir, 0o700)
		require.NoError(t, err)
		gitignorePath := filepath.Join(nexoraDir, ".gitignore")
		err = os.WriteFile(gitignorePath, []byte("custom\n"), 0o644)
		require.NoError(t, err)

		// Call createDotNexoraDir
		err = createDotNexoraDir(nexoraDir)
		require.NoError(t, err)

		// Verify .gitignore wasn't overwritten
		content, err := os.ReadFile(gitignorePath)
		require.NoError(t, err)
		assert.Equal(t, "custom\n", string(content))
	})
}

func TestShouldQueryTerminalVersion(t *testing.T) {
	tests := []struct {
		name     string
		env      []string
		expected bool
	}{
		{
			name:     "No TERM_PROGRAM and no SSH_TTY",
			env:      []string{"TERM=xterm"},
			expected: true,
		},
		{
			name:     "Apple Terminal",
			env:      []string{"TERM_PROGRAM=Apple_Terminal", "TERM=xterm"},
			expected: false,
		},
		{
			name:     "SSH session",
			env:      []string{"SSH_TTY=/dev/pts/0", "TERM=xterm"},
			expected: false,
		},
		{
			name:     "Kitty terminal",
			env:      []string{"TERM=xterm-kitty"},
			expected: true,
		},
		{
			name:     "Alacritty terminal",
			env:      []string{"TERM=alacritty"},
			expected: true,
		},
		{
			name:     "WezTerm terminal",
			env:      []string{"TERM=wezterm"},
			expected: true,
		},
		{
			name:     "Ghostty terminal",
			env:      []string{"TERM=xterm-ghostty"},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := uv.Environ(tt.env)
			result := shouldQueryTerminalVersion(env)
			assert.Equal(t, tt.expected, result)
		})
	}
}
