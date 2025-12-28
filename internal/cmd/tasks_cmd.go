package cmd

import (
	"fmt"
	"strings"

	"github.com/nexora/nexora/internal/task"
	"github.com/spf13/cobra"
)

func init() {
	tasksCmd.AddCommand(tasksGraphCmd)
	tasksCmd.AddCommand(tasksStatusCmd)
	tasksCmd.AddCommand(tasksDotCmd)
}

var tasksCmd = &cobra.Command{
	Use:   "tasks",
	Short: "Manage and visualize task graphs",
	Long:  `Commands for viewing task dependencies, progress, and execution plans.`,
}

var tasksGraphCmd = &cobra.Command{
	Use:   "graph [task-id]",
	Short: "Display task dependency graph",
	Long:  `Show an ASCII visualization of the task dependency graph.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		graph := task.NewTaskGraph()

		// Demo tasks if no real tasks exist
		if graph.TaskCount() == 0 {
			cmd.Println("No tasks in graph. Showing demo:")
			cmd.Println()
			showDemoGraph(cmd)
			return
		}

		if len(args) > 0 {
			output := graph.VisualizeASCII(args[0])
			cmd.Println(output)
		} else {
			// Show all root tasks
			roots := graph.GetRootTasks()
			if len(roots) == 0 {
				cmd.Println("No root tasks found.")
				return
			}
			for _, root := range roots {
				output := graph.VisualizeASCII(root.ID)
				cmd.Println(output)
				cmd.Println()
			}
		}
	},
}

var tasksStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show task progress summary",
	Long:  `Display a summary of all tasks and their current status.`,
	Run: func(cmd *cobra.Command, args []string) {
		graph := task.NewTaskGraph()

		if graph.TaskCount() == 0 {
			cmd.Println("No tasks tracked.")
			cmd.Println()
			cmd.Println("Tasks are created when you use the TodoWrite tool during a session.")
			return
		}

		// Count by status
		var active, completed, blocked, paused int
		tasks := graph.GetAllTasks()
		for _, t := range tasks {
			switch t.Status {
			case task.StatusActive:
				active++
			case task.StatusCompleted:
				completed++
			case task.StatusBlocked:
				blocked++
			case task.StatusPaused:
				paused++
			}
		}

		total := len(tasks)
		progress := float64(completed) / float64(total) * 100

		cmd.Printf("Task Status Summary\n")
		cmd.Printf("==================\n\n")
		cmd.Printf("Total:     %d tasks\n", total)
		cmd.Printf("Completed: %d (%.0f%%)\n", completed, progress)
		cmd.Printf("Active:    %d\n", active)
		cmd.Printf("Blocked:   %d\n", blocked)
		cmd.Printf("Paused:    %d\n", paused)

		// Show parallel execution opportunities
		parallel := graph.GetParallelTasks()
		if len(parallel) > 0 {
			cmd.Printf("\nParallel Execution Opportunities:\n")
			for _, group := range parallel {
				ids := make([]string, len(group))
				for i, t := range group {
					ids[i] = t.ID
				}
				cmd.Printf("  Can run together: %s\n", strings.Join(ids, ", "))
			}
		}
	},
}

var tasksDotCmd = &cobra.Command{
	Use:   "dot",
	Short: "Export task graph to Graphviz DOT format",
	Long:  `Export the task dependency graph in DOT format for Graphviz visualization.`,
	Run: func(cmd *cobra.Command, args []string) {
		graph := task.NewTaskGraph()

		if graph.TaskCount() == 0 {
			showDemoDot(cmd)
			return
		}

		cmd.Println(graph.ExportDOT())
	},
}

func showDemoGraph(cmd *cobra.Command) {
	demo := `Task Graph (Demo)

[done] Read config files (100%)
 └── [progress] Parse config (50%)
     └── [pending] Validate config (0%)
 └── [pending] Load plugins (0%)

Execution plan: #1 → (#2 || #4) → #3
`
	cmd.Println(demo)
}

func showDemoDot(cmd *cobra.Command) {
	dot := `digraph tasks {
  rankdir=TB;
  node [shape=box, style=rounded];

  // Demo graph - no real tasks
  "read_config" [label="Read config files\n[completed]", color=green];
  "parse_config" [label="Parse config\n[in_progress]", color=yellow];
  "validate_config" [label="Validate config\n[pending]", color=gray];
  "load_plugins" [label="Load plugins\n[pending]", color=gray];

  "read_config" -> "parse_config";
  "read_config" -> "load_plugins";
  "parse_config" -> "validate_config";
}
`
	fmt.Fprintln(cmd.OutOrStdout(), dot)
}
