package cmd

import (
	"fmt"

	"github.com/nexora/cli/internal/agent/tools"
	"github.com/spf13/cobra"
)

var suggestToolsCmd = &cobra.Command{
	Use:   "suggest-tools",
	Short: "Suggest essential tools for enhanced AI-assisted development",
	Long: `Lists recommended command-line tools that enhance Nexora's AI capabilities.
These tools provide better performance and more AI-friendly output formats for code search, navigation, and manipulation.`,
	Run: func(cmd *cobra.Command, args []string) {
		installManager := tools.NewInstallManager()
		suggestions := installManager.GenerateSuggestions(cmd.Context())
		fmt.Println(suggestions)
	},
}

func init() {
	rootCmd.AddCommand(suggestToolsCmd)
}
