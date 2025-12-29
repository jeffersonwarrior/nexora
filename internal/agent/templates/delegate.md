Delegate a task to a specialized sub-agent for asynchronous execution.

The delegate tool spawns a sub-agent that works independently. After calling this tool:
1. The tool returns immediately - you can continue with other work or end your turn
2. The sub-agent works on the task in the background
3. When complete, the sub-agent automatically reports back to this conversation
4. You will receive the report as a new message and can respond to it

Use this tool when you need to:
- Break down complex tasks into smaller, independent pieces
- Offload research or exploration to a focused sub-agent
- Run long-running analysis without blocking the main conversation
- Execute multiple tasks in parallel (call delegate multiple times)

The sub-agent has access to: view, glob, grep, and bash tools.

Parameters:
- task: The specific task to delegate (required). Be clear and specific.
- context: Additional background information the sub-agent needs (optional).
- working_dir: Directory for the sub-agent to work in (optional, defaults to project root).
- max_tokens: Maximum response length (optional, default 4096).

IMPORTANT: After delegating, you should end your turn. The delegate will report back when complete.
