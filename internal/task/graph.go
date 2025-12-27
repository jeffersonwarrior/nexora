package task

import (
	"errors"
	"fmt"
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
