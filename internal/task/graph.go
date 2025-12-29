package task

import (
	"errors"
	"fmt"
	"strings"
)

// TaskGraph manages task dependencies
type TaskGraph struct {
	nodes map[string]*Task
	edges map[string][]string // adjacency list: task -> dependencies
}

// NewTaskGraph creates a new graph
func NewTaskGraph() *TaskGraph {
	return &TaskGraph{
		nodes: make(map[string]*Task),
		edges: make(map[string][]string),
	}
}

// AddTask adds a task to the graph
func (g *TaskGraph) AddTask(task *Task) error {
	if task == nil {
		return errors.New("cannot add nil task")
	}

	if _, exists := g.nodes[task.ID]; exists {
		return fmt.Errorf("task %s already exists in graph", task.ID)
	}

	g.nodes[task.ID] = task
	g.edges[task.ID] = []string{}
	return nil
}

// AddDependency adds a dependency (from depends on to)
func (g *TaskGraph) AddDependency(from, to string) error {
	if from == to {
		return errors.New("self-dependency not allowed")
	}

	if _, exists := g.nodes[from]; !exists {
		return fmt.Errorf("task %s not found", from)
	}

	if _, exists := g.nodes[to]; !exists {
		return fmt.Errorf("task %s not found", to)
	}

	g.edges[from] = append(g.edges[from], to)
	return nil
}

// DetectCycle checks for circular dependencies using DFS
func (g *TaskGraph) DetectCycle() (bool, []string) {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	var cyclePath []string

	var dfs func(string) bool
	dfs = func(node string) bool {
		visited[node] = true
		recStack[node] = true

		for _, neighbor := range g.edges[node] {
			if !visited[neighbor] {
				cyclePath = append(cyclePath, neighbor)
				if dfs(neighbor) {
					return true
				}
				cyclePath = cyclePath[:len(cyclePath)-1]
			} else if recStack[neighbor] {
				cyclePath = append(cyclePath, neighbor)
				return true
			}
		}

		recStack[node] = false
		return false
	}

	for node := range g.nodes {
		if !visited[node] {
			cyclePath = []string{node}
			if dfs(node) {
				return true, cyclePath
			}
		}
	}

	return false, nil
}

// TopologicalSort returns tasks in execution order using Kahn's algorithm
func (g *TaskGraph) TopologicalSort() ([]*Task, error) {
	// First check for cycles
	if hasCycle, _ := g.DetectCycle(); hasCycle {
		return nil, errors.New("cannot sort graph with cycle")
	}

	// Calculate in-degree for each node (how many tasks point TO this node)
	// In our graph, edges[A] = [B] means "A depends on B"
	// So we need to count how many tasks depend on each node
	inDegree := make(map[string]int)
	for node := range g.nodes {
		inDegree[node] = 0
	}
	for from := range g.edges {
		for range g.edges[from] {
			// "from" depends on something, so increment its in-degree
			inDegree[from]++
		}
	}

	// Queue nodes with in-degree 0 (tasks that don't depend on anything)
	var queue []string
	for node, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, node)
		}
	}

	var sorted []*Task
	for len(queue) > 0 {
		// Dequeue
		node := queue[0]
		queue = queue[1:]
		sorted = append(sorted, g.nodes[node])

		// For each task that this node is a dependency of, reduce their in-degree
		// We need to find all tasks that depend on this node
		for candidate := range g.nodes {
			for _, dep := range g.edges[candidate] {
				if dep == node {
					inDegree[candidate]--
					if inDegree[candidate] == 0 {
						queue = append(queue, candidate)
					}
				}
			}
		}
	}

	return sorted, nil
}

// GetDependencies returns direct dependencies of a task
func (g *TaskGraph) GetDependencies(taskID string) []*Task {
	deps, exists := g.edges[taskID]
	if !exists {
		return []*Task{}
	}

	result := make([]*Task, 0, len(deps))
	for _, depID := range deps {
		if task, ok := g.nodes[depID]; ok {
			result = append(result, task)
		}
	}
	return result
}

// GetTransitiveDependencies returns all dependencies (recursive)
func (g *TaskGraph) GetTransitiveDependencies(taskID string) []*Task {
	visited := make(map[string]bool)
	var result []*Task

	var dfs func(string)
	dfs = func(node string) {
		for _, depID := range g.edges[node] {
			if !visited[depID] {
				visited[depID] = true
				result = append(result, g.nodes[depID])
				dfs(depID)
			}
		}
	}

	dfs(taskID)
	return result
}

// CalculateProgress computes aggregate progress with rollup
func (g *TaskGraph) CalculateProgress(taskID string) float64 {
	task, exists := g.nodes[taskID]
	if !exists {
		return 0.0
	}

	// Calculate task's own progress
	var ownProgress float64
	if len(task.Milestones) > 0 {
		completedCount := 0
		for _, milestone := range task.Milestones {
			if milestone.Status == StatusCompleted {
				completedCount++
			}
		}
		ownProgress = float64(completedCount) / float64(len(task.Milestones))
	}

	// Get dependencies
	deps := g.GetDependencies(taskID)
	if len(deps) == 0 {
		return ownProgress
	}

	// Calculate weighted average: 50% own progress, 50% dependency progress
	var depProgress float64
	for _, dep := range deps {
		depProgress += g.CalculateProgress(dep.ID)
	}
	depProgress /= float64(len(deps))

	return 0.5*ownProgress + 0.5*depProgress
}

// CriticalPath returns the longest path through the task dependency graph using topological order.
//
// The critical path is computed using a dynamic programming approach on a topologically sorted DAG:
//
// 1. First, the graph is checked for cycles. A cyclic graph has no valid critical path.
//
//  2. Tasks are processed in reverse topological order. In our graph model, edges[A] = [B] means
//     "A depends on B", so the topological sort returns tasks with no dependencies first.
//     Processing in reverse means we process dependent tasks first.
//
// 3. For each task, we compute its critical path length (longest path from that task to any leaf):
//   - A task with no dependents (no other tasks depend on it) has critical path length = 1
//   - A task with dependents has critical path length = 1 + max(critical path length of all dependents)
//
//  4. The critical path is reconstructed by tracing forward from the task with the longest
//     critical path, always choosing the dependent that contributed to that length.
//
// Example graph:
//
//	┌─────┐
//	│  A  │ (depends on nothing)
//	└──┬──┘
//	   │
//	┌──▼──┐     ┌────┐
//	│  B  │────▶│ D  │ (B depends on A, D depends on B)
//	└──┬──┘     └────┘
//	   │
//	┌──▼──┐
//	│  C  │
//	└─────┘
//
// Edges would be: B->[A], C->[A], D->[B]
// Critical path: A -> B -> D (length 3)
//
// Returns the path as a slice of task IDs ordered from root (tasks with no dependencies)
// to leaf (dependent tasks). For an empty graph, returns an empty slice.
// For a single task with no dependencies, returns a slice containing that task ID.
func (g *TaskGraph) CriticalPath() ([]string, error) {
	// Check for cycles first
	if hasCycle, _ := g.DetectCycle(); hasCycle {
		return nil, errors.New("cannot compute critical path in graph with cycle")
	}

	// Handle empty graph
	if len(g.nodes) == 0 {
		return []string{}, nil
	}

	// Get topological sort (returns tasks with no dependencies first)
	sorted, err := g.TopologicalSort()
	if err != nil {
		return nil, err
	}

	// Map to track critical path length for each task
	// Critical path length = longest path from this task downward (to leaf)
	criticalLength := make(map[string]int)
	// Map to track the next node in the critical path (which dependent contributes to the max)
	successor := make(map[string]string)

	// Build reverse edges: for each task, find which tasks depend on it
	dependents := make(map[string][]*Task)
	for taskID := range g.nodes {
		dependents[taskID] = []*Task{}
	}
	for taskID := range g.nodes {
		deps := g.GetDependencies(taskID)
		for _, dep := range deps {
			dependents[dep.ID] = append(dependents[dep.ID], g.nodes[taskID])
		}
	}

	// Process tasks in reverse topological order (dependents before dependencies)
	for i := len(sorted) - 1; i >= 0; i-- {
		task := sorted[i]
		depts := dependents[task.ID]

		if len(depts) == 0 {
			// Base case: task with no dependents (leaf node) has critical path length 1
			criticalLength[task.ID] = 1
		} else {
			// Find the maximum critical path length among dependents
			maxLen := 0
			var maxDeptID string

			for _, dept := range depts {
				deptLen := criticalLength[dept.ID]
				if deptLen > maxLen {
					maxLen = deptLen
					maxDeptID = dept.ID
				}
			}

			// Critical path length for this task is 1 + max of dependents
			criticalLength[task.ID] = maxLen + 1
			successor[task.ID] = maxDeptID
		}
	}

	// Find a task with critical path length equal to the maximum
	// Start from a root (task with no dependencies) for consistent ordering
	maxLength := 0
	for _, length := range criticalLength {
		if length > maxLength {
			maxLength = length
		}
	}

	if maxLength == 0 {
		return []string{}, nil
	}

	// Find a root task (no dependencies) with the max critical path length
	var startTask string
	for _, task := range sorted {
		if criticalLength[task.ID] == maxLength {
			startTask = task.ID
			break
		}
	}

	if startTask == "" {
		return []string{}, nil
	}

	// Reconstruct the critical path by following successors
	var path []string
	current := startTask

	for current != "" {
		path = append(path, current)
		current = successor[current]
	}

	return path, nil
}

// TaskCount returns the number of tasks in the graph
func (g *TaskGraph) TaskCount() int {
	return len(g.nodes)
}

// GetTask returns a task by ID
func (g *TaskGraph) GetTask(id string) *Task {
	return g.nodes[id]
}

// GetAllTasks returns all tasks in the graph
func (g *TaskGraph) GetAllTasks() []*Task {
	tasks := make([]*Task, 0, len(g.nodes))
	for _, t := range g.nodes {
		tasks = append(tasks, t)
	}
	return tasks
}

// GetRootTasks returns tasks with no dependencies (entry points)
func (g *TaskGraph) GetRootTasks() []*Task {
	var roots []*Task
	for id, t := range g.nodes {
		if len(g.edges[id]) == 0 {
			roots = append(roots, t)
		}
	}
	return roots
}

// GetParallelTasks returns groups of tasks that can run in parallel
func (g *TaskGraph) GetParallelTasks() [][]*Task {
	// Tasks can run in parallel if they:
	// 1. Have the same dependencies satisfied
	// 2. Don't depend on each other
	sorted, err := g.TopologicalSort()
	if err != nil || len(sorted) == 0 {
		return nil
	}

	// Group tasks by their dependency set
	groups := make(map[string][]*Task)
	for _, t := range sorted {
		if t.Status == StatusCompleted {
			continue
		}
		deps := g.edges[t.ID]
		// Check if all deps are completed
		allDepsCompleted := true
		for _, depID := range deps {
			if dep := g.nodes[depID]; dep != nil && dep.Status != StatusCompleted {
				allDepsCompleted = false
				break
			}
		}
		if allDepsCompleted {
			key := fmt.Sprintf("%v", deps)
			groups[key] = append(groups[key], t)
		}
	}

	// Return groups with more than one task
	var result [][]*Task
	for _, group := range groups {
		if len(group) > 1 {
			result = append(result, group)
		}
	}
	return result
}

// ExportDOT exports the graph in Graphviz DOT format
func (g *TaskGraph) ExportDOT() string {
	var sb strings.Builder
	sb.WriteString("digraph tasks {\n")
	sb.WriteString("  rankdir=TB;\n")
	sb.WriteString("  node [shape=box, style=rounded];\n\n")

	// Add nodes
	for id, t := range g.nodes {
		color := "gray"
		switch t.Status {
		case StatusCompleted:
			color = "green"
		case StatusActive:
			color = "yellow"
		case StatusBlocked:
			color = "red"
		}
		sb.WriteString(fmt.Sprintf("  %q [label=%q, color=%s];\n",
			id, fmt.Sprintf("%s\\n[%s]", t.Title, t.Status), color))
	}

	sb.WriteString("\n")

	// Add edges
	for from, deps := range g.edges {
		for _, to := range deps {
			sb.WriteString(fmt.Sprintf("  %q -> %q;\n", from, to))
		}
	}

	sb.WriteString("}\n")
	return sb.String()
}
