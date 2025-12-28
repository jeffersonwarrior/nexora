package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/nexora/nexora/internal/config"
	"github.com/nexora/nexora/internal/db"
	"github.com/nexora/nexora/internal/session"
	"github.com/spf13/cobra"
)

func init() {
	checkpointCmd.AddCommand(checkpointListCmd)
	checkpointCmd.AddCommand(checkpointSaveCmd)
	checkpointCmd.AddCommand(checkpointRestoreCmd)
	checkpointCmd.AddCommand(checkpointDeleteCmd)
}

var checkpointCmd = &cobra.Command{
	Use:   "checkpoint",
	Short: "Manage session checkpoints",
	Long:  `Save, restore, and manage session checkpoints for recovery and debugging.`,
}

var checkpointListCmd = &cobra.Command{
	Use:   "list [session-id]",
	Short: "List available checkpoints",
	Long:  `List all checkpoints for a session, or all sessions if no ID provided.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		dataDir := filepath.Dir(config.GlobalConfigData())
		conn, err := db.Connect(ctx, dataDir)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer conn.Close()

		svc := session.NewCheckpointService(db.New(conn))

		var sessionID string
		if len(args) > 0 {
			sessionID = args[0]
		}

		checkpoints, err := svc.List(ctx, sessionID)
		if err != nil {
			return fmt.Errorf("failed to list checkpoints: %w", err)
		}

		if len(checkpoints) == 0 {
			cmd.Println("No checkpoints found.")
			cmd.Println()
			cmd.Println("Checkpoints are created automatically during sessions or manually with:")
			cmd.Println("  nexora checkpoint save <name>")
			return nil
		}

		cmd.Printf("%-36s  %-20s  %8s  %8s  %-20s\n", "ID", "Session", "Tokens", "Messages", "Created")
		cmd.Printf("%-36s  %-20s  %8s  %8s  %-20s\n", "------------------------------------", "--------------------", "--------", "--------", "--------------------")

		for _, cp := range checkpoints {
			age := formatAge(cp.Timestamp)
			cmd.Printf("%-36s  %-20s  %8d  %8d  %-20s\n",
				truncate(cp.ID, 36),
				truncate(cp.SessionID, 20),
				cp.TokenCount,
				cp.MessageCount,
				age,
			)
		}

		return nil
	},
}

var checkpointSaveCmd = &cobra.Command{
	Use:   "save <name>",
	Short: "Create a named checkpoint",
	Long:  `Save the current session state as a named checkpoint.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Note: This requires an active session context
		// In practice, this would be called from within the TUI
		cmd.Println("Checkpoint save is available during active sessions.")
		cmd.Println()
		cmd.Println("To save a checkpoint:")
		cmd.Println("  1. Start a nexora session")
		cmd.Println("  2. Type: /checkpoint save " + args[0])
		cmd.Println()
		cmd.Println("Checkpoints are also created automatically:")
		cmd.Println("  - Before dangerous operations (rm, git push)")
		cmd.Println("  - Every 50 messages")
		cmd.Println("  - On significant task completions")
		return nil
	},
}

var checkpointRestoreCmd = &cobra.Command{
	Use:   "restore <checkpoint-id>",
	Short: "Restore a session from checkpoint",
	Long:  `Restore a previous session state from a checkpoint.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		dataDir := filepath.Dir(config.GlobalConfigData())
		conn, err := db.Connect(ctx, dataDir)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer conn.Close()

		svc := session.NewCheckpointService(db.New(conn))

		// Get the checkpoint first
		cp, err := svc.Get(ctx, args[0])
		if err != nil {
			return fmt.Errorf("checkpoint not found: %w", err)
		}

		cmd.Printf("Checkpoint found:\n")
		cmd.Printf("  ID:       %s\n", cp.ID)
		cmd.Printf("  Session:  %s\n", cp.SessionID)
		cmd.Printf("  Tokens:   %d\n", cp.TokenCount)
		cmd.Printf("  Messages: %d\n", cp.MessageCount)
		cmd.Printf("  Created:  %s\n", formatAge(cp.Timestamp))
		cmd.Println()
		cmd.Println("To restore this checkpoint, start nexora and use:")
		cmd.Printf("  /checkpoint restore %s\n", args[0])

		return nil
	},
}

var checkpointDeleteCmd = &cobra.Command{
	Use:   "delete <checkpoint-id>",
	Short: "Delete a checkpoint",
	Long:  `Permanently delete a checkpoint.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		dataDir := filepath.Dir(config.GlobalConfigData())
		conn, err := db.Connect(ctx, dataDir)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer conn.Close()

		svc := session.NewCheckpointService(db.New(conn))

		if err := svc.Delete(ctx, args[0]); err != nil {
			return fmt.Errorf("failed to delete checkpoint: %w", err)
		}

		cmd.Printf("Checkpoint %s deleted.\n", args[0])
		return nil
	},
}

func formatAge(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "just now"
	case d < time.Hour:
		return fmt.Sprintf("%d min ago", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d hours ago", int(d.Hours()))
	default:
		return fmt.Sprintf("%d days ago", int(d.Hours()/24))
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
