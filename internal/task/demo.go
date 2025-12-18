package task

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

// DemoService shows how to use the task management system
type DemoService struct {
	service SimpleService
}

// NewDemo creates a demo service with examples
func NewDemo() *DemoService {
	return &DemoService{
		service: NewSimpleService(),
	}
}

// RunDemo demonstrates the task management functionality
func (d *DemoService) RunDemo() {
	ctx := context.Background()
	sessionID := "demo-session"

	// Example 1: Start a task
	task := d.service.CreateTask(ctx, sessionID, "Implement API endpoint", "Create a new REST API endpoint for user management")
	fmt.Printf("🎯 Started task: %s\n", task.Title)

	// Example 2: Check if responses are on task
	responses := []string{
		"Here's the code for the user API endpoint with proper error handling",
		"By the way, did you know that APIs were invented in 2000?",
		"Let me also implement a weather service while we're here",
		"Here's the completed user management endpoint with all required methods",
	}

	for _, resp := range responses {
		isFocused, reason := d.service.CheckFocus(ctx, sessionID, resp)
		if isFocused {
			fmt.Printf("✅ Focused: %s\n", reason)
		} else {
			fmt.Printf("⚠️  Drift: %s\n", reason)
			fmt.Printf("📋 Correction: %s\n", d.service.GetFocusPrompt(ctx, sessionID))
		}
		fmt.Println()
	}

	// Example 3: Complete the task
	if d.service.CompleteTask(ctx, sessionID) {
		fmt.Println("🎉 Task completed successfully!")
	}
}

// DemoCommand returns a CLI command to run the task demonstration
func (d *DemoService) DemoCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task-demo",
		Short: "Demonstrate task management and drift prevention",
		Run: func(cmd *cobra.Command, args []string) {
			d.RunDemo()
		},
	}
	return cmd
}
