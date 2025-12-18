package cmd

import (
	"context"
	"fmt"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/nexora/cli/internal/aiops"
	"github.com/nexora/cli/internal/config"
	"github.com/spf13/cobra"
)

var aiopsCmd = &cobra.Command{
	Use:   "aiops",
	Short: "Manage AI Operations service",
	Long:  "AI Operations provides intelligent assistance for edit resolution, loop detection, and task drift detection.",
}

var aiopsStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check AIOPS service status",
	RunE:  runAIOpsStatus,
}

var aiopsTestCmd = &cobra.Command{
	Use:   "test",
	Short: "Test AIOPS service connectivity",
	RunE:  runAIOpsTest,
}

func init() {
	aiopsCmd.AddCommand(aiopsStatusCmd)
	aiopsCmd.AddCommand(aiopsTestCmd)
}

func runAIOpsStatus(cmd *cobra.Command, args []string) error {
	cwd, err := cmd.Flags().GetString("cwd")
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %v", err)
	}

	dataDir, err := cmd.Flags().GetString("data-dir")
	if err != nil {
		return fmt.Errorf("failed to get data directory: %v", err)
	}

	cfg, err := config.Load(cwd, dataDir, false)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !cfg.AIOPS.Enabled {
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("AIOPS: disabled"))
		return nil
	}

	// Create AIOPS client
	aiopsCfg := aiops.Config{
		Enabled:  cfg.AIOPS.Enabled,
		Endpoint: cfg.AIOPS.Endpoint,
		Timeout:  cfg.AIOPS.Timeout,
		Fallback: cfg.AIOPS.Fallback,
	}
	client := aiops.NewClient(aiopsCfg)

	// Check availability
	if !client.Available() {
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("AIOPS: unavailable"))
		fmt.Printf("Endpoint: %s\n", cfg.AIOPS.Endpoint)
		return nil
	}

	// Get model info
	modelInfo := client.ModelInfo()
	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("10")).
		Bold(true)

	fmt.Println(style.Render("AIOPS: available"))
	fmt.Printf("Endpoint: %s\n", cfg.AIOPS.Endpoint)
	fmt.Printf("Model: %s (%s)\n", modelInfo.Name, modelInfo.Runtime)
	if modelInfo.Latency > 0 {
		fmt.Printf("Latency: %s\n", modelInfo.Latency)
	}

	return nil
}

func runAIOpsTest(cmd *cobra.Command, args []string) error {
	cwd, err := cmd.Flags().GetString("cwd")
	if err != nil {
		return fmt.Errorf("failed to get current working directory: %v", err)
	}

	dataDir, err := cmd.Flags().GetString("data-dir")
	if err != nil {
		return fmt.Errorf("failed to get data directory: %v", err)
	}

	cfg, err := config.Load(cwd, dataDir, false)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if !cfg.AIOPS.Enabled {
		return fmt.Errorf("AIOPS is disabled in config")
	}

	// Create AIOPS client
	aiopsCfg := aiops.Config{
		Enabled:  cfg.AIOPS.Enabled,
		Endpoint: cfg.AIOPS.Endpoint,
		Timeout:  cfg.AIOPS.Timeout,
		Fallback: cfg.AIOPS.Fallback,
	}
	client := aiops.NewClient(aiopsCfg)
	if !client.Available() {
		return fmt.Errorf("AIOPS service not available at %s", cfg.AIOPS.Endpoint)
	}

	// Test edit resolution
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	fmt.Println(lipgloss.NewStyle().Bold(true).Render("Testing AIOPS service..."))

	// Test edit resolution
	res, err := client.ResolveEdit(ctx,
		"func hello() {\n\tfmt.Println(\"Hello\")\n}",
		"Hello", "World")
	if err != nil {
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(fmt.Sprintf("Edit resolution test failed: %v", err)))
	} else {
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(fmt.Sprintf("✓ Edit resolution: confidence %.2f", res.Confidence)))
	}

	// Test loop detection
	testCalls := []aiops.ToolCall{
		{Name: "edit", Result: "old_string not found", Timestamp: time.Now()},
		{Name: "edit", Result: "old_string not found", Timestamp: time.Now()},
		{Name: "edit", Result: "old_string not found", Timestamp: time.Now()},
	}
	loop, err := client.DetectLoop(ctx, testCalls)
	if err != nil {
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(fmt.Sprintf("Loop detection test failed: %v", err)))
	} else {
		if loop.IsLooping {
			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(fmt.Sprintf("✓ Loop detection: %s", loop.PatternType)))
		} else {
			msg := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render("✓ Loop detection: no loop detected\n")
			fmt.Print(msg)
		}
	}

	// Test drift detection
	testActions := []aiops.Action{
		{Description: "Edit file", Timestamp: time.Now()},
		{Description: "Another edit", Timestamp: time.Now()},
	}
	drift, err := client.DetectDrift(ctx, "Add error handling", testActions)
	if err != nil {
		fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(fmt.Sprintf("Drift detection test failed: %v", err)))
	} else {
		if drift.IsDrifting {
			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Render(fmt.Sprintf("✓ Drift detection: drift detected (score: %.2f)", drift.DriftScore)))
		} else {
			fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(fmt.Sprintf("✓ Drift detection: on track (score: %.2f)", drift.DriftScore)))
		}
	}

	return nil
}
