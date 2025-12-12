package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strings"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run [prompt...]",
	Short: "Run a single non-interactive prompt",
	Long: `Run a single prompt in non-interactive mode and exit.
The prompt can be provided as arguments or piped from stdin.`,
	Example: `
# Run a simple prompt
nexora run Explain the use of context in Go

# Pipe input from stdin
curl https://charm.land | nexora run "Summarize this website"

# Read from a file
nexora run "What is this code doing?" <<< prrr.go

# Run in quiet mode (hide the spinner)
nexora run --quiet "Generate a README for this project"
  `,
	RunE: func(cmd *cobra.Command, args []string) error {
		quiet, _ := cmd.Flags().GetBool("quiet")

		app, err := setupApp(cmd)
		if err != nil {
			return err
		}
		defer app.Shutdown()

		if !app.Config().IsConfigured() {
			return fmt.Errorf("no providers configured - please run 'nexora' to set up a provider interactively")
		}

		prompt := strings.Join(args, " ")

		prompt, err = MaybePrependStdin(prompt)
		if err != nil {
			slog.Error("Failed to read from stdin", "error", err)
			return err
		}

		if prompt == "" {
			return fmt.Errorf("no prompt provided")
		}

		// Create a context that cancels on SIGINT (Ctrl+C), handling both interactive
		// and redirected output scenarios. The spinner in RunNonInteractive will detect
		// TTY availability and adjust rendering accordingly.
		ctx, cancel := signal.NotifyContext(cmd.Context(), os.Interrupt)
		defer cancel()

		return app.RunNonInteractive(ctx, os.Stdout, prompt, quiet)
	},
}

func init() {
	runCmd.Flags().BoolP("quiet", "q", false, "Hide spinner")
}
