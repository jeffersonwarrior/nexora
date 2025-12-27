#!/bin/bash
set -e
cd '/home/nexora'
PROMPT=$(cat '/home/nexora/.claude/orchestrator/workers/feature-4.prompt')
# Worker: allowed tools and MCP servers configured via env vars
claude --model claude-sonnet-4-5 --allowedTools "Bash,Read,Write,Edit,Glob,Grep,Task,TodoWrite" --permission-mode default --mcp-config '{"mcpServers":{}}' -p "$PROMPT" 2>&1 | tee '/home/nexora/.claude/orchestrator/workers/feature-4.log'
echo 'WORKER_EXITED' >> '/home/nexora/.claude/orchestrator/workers/feature-4.log'
