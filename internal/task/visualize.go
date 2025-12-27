package task

import (
	"fmt"
	"strings"
)

// VisualizeASCII generates an ASCII tree representation of the task graph
func (g *TaskGraph) VisualizeASCII(rootTaskID string) string {
	task, exists := g.nodes[rootTaskID]
	if !exists {
		return fmt.Sprintf("Task %s not found", rootTaskID)
	}

	var sb strings.Builder
	g.renderNode(&sb, task, "", true)
	return sb.String()
}

// renderNode recursively renders a task node and its dependencies
func (g *TaskGraph) renderNode(sb *strings.Builder, task *Task, prefix string, isLast bool) {
	progress := g.CalculateProgress(task.ID)
	status := getStatusSymbol(task.Status, progress)
	progressStr := formatProgress(progress)

	// Render current node
	if prefix == "" {
		// Root node
		sb.WriteString(fmt.Sprintf("%s (%s)\n", task.Title, progressStr))
	} else {
		// Child node
		connector := "├── "
		if isLast {
			connector = "└── "
		}
		sb.WriteString(fmt.Sprintf("%s%s[%s] %s (%s)\n", prefix, connector, status, task.Title, progressStr))
	}

	// Render dependencies
	deps := g.GetDependencies(task.ID)
	if len(deps) > 0 {
		for i, dep := range deps {
			isLastDep := i == len(deps)-1
			var newPrefix string
			if prefix == "" {
				// Root level - children get a minimal prefix to show tree structure
				newPrefix = " "
			} else if isLast {
				newPrefix = prefix + "    "
			} else {
				newPrefix = prefix + "│   "
			}
			g.renderNode(sb, dep, newPrefix, isLastDep)
		}
	}
}

// formatProgress formats progress as a percentage
func formatProgress(progress float64) string {
	return fmt.Sprintf("%d%%", int(progress*100))
}

// getStatusSymbol returns a status indicator based on task status and progress
func getStatusSymbol(status TaskStatus, progress float64) string {
	switch status {
	case StatusCompleted:
		return "done"
	case StatusBlocked:
		return "blocked"
	case StatusPaused:
		return "paused"
	case StatusCancelled:
		return "cancelled"
	case StatusActive:
		if progress > 0 {
			return "progress"
		}
		return "pending"
	default:
		return "unknown"
	}
}
