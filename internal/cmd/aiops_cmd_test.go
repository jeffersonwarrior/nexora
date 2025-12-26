package cmd

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

func TestAIOpsCmd(t *testing.T) {
	t.Run("Command has correct metadata", func(t *testing.T) {
		assert.Equal(t, "aiops", aiopsCmd.Use)
		assert.Equal(t, "Manage AI Operations service", aiopsCmd.Short)
		assert.Contains(t, aiopsCmd.Long, "AI Operations")
	})

	t.Run("Has subcommands", func(t *testing.T) {
		subcommands := aiopsCmd.Commands()
		assert.NotEmpty(t, subcommands, "Should have subcommands")

		cmdNames := make([]string, 0)
		for _, cmd := range subcommands {
			cmdNames = append(cmdNames, cmd.Use)
		}

		// Verify expected subcommands
		assert.Contains(t, cmdNames, "status")
		assert.Contains(t, cmdNames, "test")
	})
}

func TestAIOpsStatusCmd(t *testing.T) {
	// Find the status subcommand
	var statusCmd *cobra.Command
	for _, cmd := range aiopsCmd.Commands() {
		if cmd.Use == "status" {
			statusCmd = cmd
			break
		}
	}

	assert.NotNil(t, statusCmd, "Should have status subcommand")
	if statusCmd != nil {
		assert.Equal(t, "status", statusCmd.Use)
		assert.NotNil(t, statusCmd.RunE)
	}
}

func TestAIOpsTestCmd(t *testing.T) {
	// Find the test subcommand
	var testCmd *cobra.Command
	for _, cmd := range aiopsCmd.Commands() {
		if cmd.Use == "test" {
			testCmd = cmd
			break
		}
	}

	assert.NotNil(t, testCmd, "Should have test subcommand")
	if testCmd != nil {
		assert.Equal(t, "test", testCmd.Use)
		assert.NotNil(t, testCmd.RunE)
	}
}
