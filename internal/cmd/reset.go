package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nexora/nexora/internal/config"
	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset",
	Short: "Reset Nexora to initial state",
	Long: `Reset Nexora configuration and data to trigger first-launch setup.

This command removes configuration files and optionally data files,
allowing Nexora to run the initial setup wizard on next launch.

By default, it preserves session history and database. Use flags to
control what gets deleted.`,
	Example: `
# Reset config only (keeps sessions and data)
nexora reset

# Full reset (removes everything)
nexora reset --all

# Reset without confirmation prompt
nexora reset --force

# Reset and also clear sessions
nexora reset --clear-sessions
  `,
	RunE: func(cmd *cobra.Command, args []string) error {
		force, _ := cmd.Flags().GetBool("force")
		all, _ := cmd.Flags().GetBool("all")
		clearSessions, _ := cmd.Flags().GetBool("clear-sessions")

		configDir := filepath.Dir(config.GlobalConfig())
		dataDir := config.GlobalDataDir()

		// Show what will be deleted
		fmt.Println("This will reset Nexora to its initial state.")
		fmt.Println()
		fmt.Println("The following will be deleted:")
		fmt.Printf("  • Config directory: %s\n", configDir)
		if all {
			fmt.Printf("  • Data directory:   %s (includes database, sessions, logs)\n", dataDir)
		} else if clearSessions {
			fmt.Printf("  • Sessions and init flag in: %s\n", dataDir)
		} else {
			fmt.Printf("  • Init flag in: %s\n", dataDir)
		}
		fmt.Println()

		if !force {
			fmt.Print("Are you sure? [y/N]: ")
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}
			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Reset cancelled.")
				return nil
			}
		}

		// Delete config directory
		if err := os.RemoveAll(configDir); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("failed to remove config directory: %w", err)
		}
		fmt.Printf("✓ Removed config directory: %s\n", configDir)

		if all {
			// Delete entire data directory
			if err := os.RemoveAll(dataDir); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove data directory: %w", err)
			}
			fmt.Printf("✓ Removed data directory: %s\n", dataDir)
		} else {
			// Just remove init flag to trigger re-initialization
			initFlagPath := filepath.Join(dataDir, config.InitFlagFilename)
			if err := os.Remove(initFlagPath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to remove init flag: %w", err)
			}
			fmt.Printf("✓ Removed init flag: %s\n", initFlagPath)

			if clearSessions {
				// Remove database file
				dbPath := filepath.Join(dataDir, "nexora.db")
				if err := os.Remove(dbPath); err != nil && !os.IsNotExist(err) {
					return fmt.Errorf("failed to remove database: %w", err)
				}
				fmt.Printf("✓ Removed database: %s\n", dbPath)

				// Remove WAL and SHM files if they exist
				for _, suffix := range []string{"-wal", "-shm"} {
					walPath := dbPath + suffix
					_ = os.Remove(walPath) // Ignore errors for WAL files
				}
			}
		}

		fmt.Println()
		fmt.Println("Reset complete. Run 'nexora' to start fresh setup.")
		return nil
	},
}

func init() {
	resetCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
	resetCmd.Flags().BoolP("all", "a", false, "Remove all data including sessions and database")
	resetCmd.Flags().Bool("clear-sessions", false, "Also clear session history and database")
}
