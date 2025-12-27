package cmd

import (
	"bytes"
	"runtime"
	"strings"
	"testing"

	"github.com/nexora/nexora/internal/version"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestAboutCommand_Executes(t *testing.T) {
	// Test command runs without error
	var buf bytes.Buffer
	cmd := &cobra.Command{
		Use: "test",
		Run: aboutCmd.Run,
	}
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	cmd.Run(cmd, []string{})
	// No error to check since Run doesn't return error
}

func TestAboutCommand_ContainsVersion(t *testing.T) {
	// Test output includes version string
	var buf bytes.Buffer
	cmd := &cobra.Command{
		Use: "test",
		Run: aboutCmd.Run,
	}
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	cmd.Run(cmd, []string{})

	output := buf.String()
	require.Contains(t, output, version.Version)
	require.Contains(t, output, "Nexora")
}

func TestAboutCommand_ContainsPlatform(t *testing.T) {
	// Test output includes GOOS, GOARCH, Go version
	var buf bytes.Buffer
	cmd := &cobra.Command{
		Use: "test",
		Run: aboutCmd.Run,
	}
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	cmd.Run(cmd, []string{})

	output := buf.String()
	require.Contains(t, output, runtime.GOOS)
	require.Contains(t, output, runtime.GOARCH)
	require.Contains(t, output, "Go Version")
}

func TestAboutCommand_ContainsCommunityLinks(t *testing.T) {
	// Discord: https://discord.gg/GCyC6qT79M
	// Twitter/X: https://x.com/i/communities/2004598673062216166/
	// Reddit: r/Zackor
	var buf bytes.Buffer
	cmd := &cobra.Command{
		Use: "test",
		Run: aboutCmd.Run,
	}
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	cmd.Run(cmd, []string{})

	output := buf.String()
	require.Contains(t, output, "https://discord.gg/GCyC6qT79M")
	require.Contains(t, output, "https://x.com/i/communities/2004598673062216166")
	require.Contains(t, output, "r/Zackor")
	require.Contains(t, output, "Community")
}

func TestAboutCommand_ContainsRepoLinks(t *testing.T) {
	// GitHub: https://github.com/jeffersonwarrior/nexora
	// Releases: https://github.com/jeffersonwarrior/nexora/releases
	var buf bytes.Buffer
	cmd := &cobra.Command{
		Use: "test",
		Run: aboutCmd.Run,
	}
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	cmd.Run(cmd, []string{})

	output := buf.String()
	require.Contains(t, output, "https://github.com/jeffersonwarrior/nexora")
	require.Contains(t, output, "https://github.com/jeffersonwarrior/nexora/releases")
	require.Contains(t, output, "Repository")
}

func TestAboutCommand_ContainsLicense(t *testing.T) {
	// Test output includes license information
	var buf bytes.Buffer
	cmd := &cobra.Command{
		Use: "test",
		Run: aboutCmd.Run,
	}
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	cmd.Run(cmd, []string{})

	output := buf.String()
	require.Contains(t, strings.ToUpper(output), "MIT")
}
