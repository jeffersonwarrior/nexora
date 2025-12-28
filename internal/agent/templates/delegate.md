Delegate a task to a specialized sub-agent for parallel execution with resource-aware pooling.

The delegate tool manages a pool of sub-agents with:
- Dynamic spawning based on available CPU and memory
- Queue with 30-minute timeout for resource-constrained situations
- Automatic resource exhaustion prevention

Use this tool when you need to:
- Break down complex tasks into smaller, independent pieces
- Execute multiple tasks in parallel (background mode)
- Offload research or exploration to a focused sub-agent
- Run long-running analysis without blocking the main conversation

The sub-agent will have access to: view, glob, grep, and bash tools.

Parameters:
- task: The specific task to delegate (required). Be clear and specific.
- context: Additional background information the sub-agent needs (optional).
- working_dir: Directory for the sub-agent to work in (optional, defaults to project root).
- max_tokens: Maximum response length (optional, default 4096).
- background: Set to true to queue the task and return immediately with a task ID (optional).

Example usage:
- Synchronous delegation: {"task": "Analyze error handling in auth module", "context": "Looking for consistency issues"}
- Background delegation: {"task": "Find all usages of deprecated API", "background": true}
- Parallel tasks: Submit multiple background delegations, then check results
