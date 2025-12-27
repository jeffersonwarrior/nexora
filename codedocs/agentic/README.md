# Nexora Swarm Orchestration Guidelines

**Philosophy**: This is math and science work, not expressionism. We measure results scientifically.

---

## Overview

The swarm orchestration process follows a rigorous, multi-agent workflow designed to ensure production-quality code through systematic review, testing, and validation.

### Core Principles

1. **Detailed Task Enumeration**: Review work thoroughly and explain it in the form of the best possible prompt for each task
2. **Interactive Design Phase**: Ping-pong interaction with agents before execution - start with "Any questions or suggestions?"
3. **Knowledge-Based Answers**: Answer agent questions using the knowledgebase; for external dependencies, read READMEs or web search
4. **Scientific Validation**: Measure results scientifically, not subjectively
5. **Agent Autonomy**: Once the task is designed to satisfaction, launch the agent to work

### Pre-Work Protocol

For each task:
1. Create the best possible prompt with all context
2. Include detailed test summary ensuring production quality
3. Make tests detailed enough to validate accuracy
4. Launch agent with "Any questions or suggestions?"
5. Read the agent's response
6. Answer questions based on knowledgebase
7. For external dependencies: READ THE README or web search
8. Iterate until design is accurate
9. Then and only then: launch agent into execution

**Critical**: When working with external dependencies or anything beyond raw math/code, **always read documentation first**.

---

## Git Workflow Requirements

### Branching Strategy

All work **MUST** be in git with proper branch management:

1. **Create branch** for each agent/task
2. **Commit to branch** as work progresses
3. **Proof of work** must exist in branch commits
4. **Merge only after validation** (see Testing Requirements)

### Branch Naming

Use descriptive names:
- `feature/task-graph-enrichment`
- `fix/delegate-banner-tui`
- `test/agent-tools-coverage`

---

## Test-Driven Development (TDD)

### Testing Hierarchy

**Every section of code must be tested.** Testing happens at multiple levels:

1. **Unit Tests**: Each function, each method
2. **Integration Tests**: Subsections working together
3. **Module Tests**: Complete modules
4. **End-to-End Tests**: Full system workflows
5. **User Journey Tests**: Complete use cases

### Testing Phases

When adding new code:
1. Test the new section
2. Test the sub-module
3. Test the entire module
4. Test the entire codebase
5. Test end-to-end workflows
6. Test user journeys

**This applies to EACH PHASE.**

### Why We Test Repeatedly

> Just because it worked once doesn't mean it will work a second time when we change test conditions.

**By definition**: When we change code, we change test conditions. Therefore:
- Test once
- Test again
- Test a third time
- Only then consider quality validated

**Never say "PRODUCTION READY" or use celebratory language.** Code is either validated or not.

---

## Philosophy: Focus and Accuracy

### No Time Limits

- **No time estimates** ("developer weeks", "sprints", etc.)
- **No context token limits** (use what's needed)
- **No excessive pushback** on scope

### What Matters

- **The TODO list exists**
- **Completing the TODO list**
- **Heads-down focus**
- **Accuracy in the work**

### The Nature of Work

> Complex equations can be solved in a short amount of time if we are focused.

Work is an expression of raw energy. For as long as the universe has existed and will exist, there will be work. Therefore, our best interest lies in **expressing this energy into the most accurate work ever created**.

**Distractions to avoid**:
- Handwaving timelines
- "100% PRODUCTION READY!!!" declarations
- Excessive emoji celebrations
- Premature completion claims

**This is not difficult work. It is just work.**

---

## Architecture: Multi-Level Agent Workflow

### Level 1: Overseer

**Role**: You (the orchestrating agent)

**Responsibilities**:
1. Explain and design the overarching task for this phase
2. Create the master test suite
3. Send task specifications to Senior Developer
4. Conduct Q&A with Senior Developer
5. Send Dev team to work
6. Review completed work
7. Run final validation tests
8. Return results to user

**Workflow**:
```
Overseer → Design Phase → Tests Created → Senior Developer → Q&A → Execute
```

### Level 2: Senior Developer

**Role**: Architect and monitor agent execution

**Responsibilities**:

#### A. Pre-Execution Design
1. **Architect team member prompts** with:
   - Relevant languages (Go, Python, etc.)
   - API documentation
   - Reference documents in MD format
   - Specific code/test references
   - Any helpful context

2. **Create branch** for each agent

3. **Create perfect prompt** including:
   - All relevant details
   - Context from knowledgebase
   - External documentation links
   - Error handling guidance
   - A2A escalation protocol

4. **Create test specifications** for the agent to code against

5. **Launch agent** with "Any questions or suggestions?"

6. **A2A communications**: Query, consider, answer, iterate

7. **Send agent to work** when design is validated

#### B. Monitoring Loop
1. **Monitor all running agents**
2. **60-second loop** until job complete or agent freezes
3. **If no activity**: Prompt with "What is the status?"
4. **If frozen**: Restart the agent
5. **If still blocked**: Create new prompt with different approach

#### C. Completion
1. **Merge branch** when agent completes
2. **Test again** after merge
3. **Correct any issues** to best ability
4. **Return task to Overseer**

### Level 3: Agent (Worker)

**Role**: Execute the task autonomously

**Responsibilities**:
1. **Do the work** according to prompt specifications
2. **Commit to branch** when done
3. **Handle challenges autonomously**:
   - Review documentation if available
   - Web search for error solutions
   - Research unexpected conditions
   - A2A to Senior Developer if still blocked

**Autonomous Challenge Resolution Protocol** (MUST BE IN PROMPT):
```
1. Encounter challenge
2. Review local documentation
3. If unresolved: Web search for solution
4. If still unresolved: A2A Senior Developer with:
   - Challenge description
   - What was tried
   - Current state
   - Suggested approaches
5. Wait for guidance
6. Resume work
```

**Completion Criteria**:
- All tests pass
- Code committed to branch
- Documentation updated
- No errors or warnings

---

## Agent Prompt Template

### For Senior Developer Creating Agent Prompts

```markdown
# Task: [Specific Feature/Fix Name]

## Objective
[Clear, specific goal statement]

## Context
[Background information, why this matters]

## Technical Requirements
- Language: [Go/Python/etc.]
- Package: [internal/agent/tools]
- Files to modify: [list]
- Files to create: [list]

## Reference Documentation
- [Link to relevant README]
- [Link to API docs]
- [Link to related code]
- [Link to similar tests]

## Test Specifications
[Detailed test requirements - what must pass]

### Unit Tests
- [Specific test 1]
- [Specific test 2]

### Integration Tests
- [Specific integration scenario]

### Edge Cases
- [Edge case 1]
- [Edge case 2]

## Acceptance Criteria
- [ ] All tests pass
- [ ] Code coverage ≥ [target]%
- [ ] No linting errors
- [ ] Documentation updated
- [ ] Committed to branch: [branch-name]

## Autonomous Challenge Protocol
If you encounter challenges:
1. Review documentation in: [specific paths]
2. Web search: [suggest search terms]
3. If blocked: A2A Senior Developer with detailed context

## Branch
Work on: `[branch-name]`

## Questions?
Any questions or suggestions before beginning?
```

---

## Validation Gates

### After Agent Completion
```bash
# Agent completes work
git checkout [branch-name]
go test ./... -v -race
go vet ./...
golangci-lint run
```

### After Branch Merge
```bash
# Senior Developer merges
git checkout main
git merge [branch-name]
go test ./... -v -race
go build ./...
# Run full test suite again
```

### After Phase Completion
```bash
# Overseer validates
go test ./... -v -race -timeout 10m
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
# Check coverage targets met
# Run E2E tests
# Run user journey tests
```

---

## Communication Protocols

### A2A (Agent-to-Agent)

**When to use**:
- Agent is blocked
- Unexpected behavior encountered
- Clarification needed on requirements
- Design decision required

**Format**:
```
TO: Senior Developer
FROM: Agent [name/id]
TASK: [task name]
STATUS: BLOCKED

ISSUE:
[Detailed description]

ATTEMPTED:
1. [What was tried]
2. [What was tried]

CURRENT STATE:
[Code state, test results, error messages]

SUGGESTED APPROACHES:
1. [Option A]
2. [Option B]

REQUEST: Guidance on best path forward
```

### Status Updates

**Every agent should report**:
- Current step
- Tests passing/failing
- Blockers encountered
- ETA to completion (if determinable)

**Format**:
```
AGENT: [name/id]
TASK: [task name]
PROGRESS: [X/Y steps complete]
TESTS: [passing/total]
BLOCKERS: [none/description]
NEXT: [next action]
```

---

## Quality Standards

### Code Quality
- All tests pass
- No linting errors
- No vet warnings
- Coverage targets met
- Documentation complete

### Test Quality
- Tests are deterministic
- Tests cover edge cases
- Tests validate behavior, not implementation
- Tests run in parallel where possible
- Tests clean up resources

### Documentation Quality
- Clear purpose statements
- Usage examples included
- Edge cases documented
- API changes noted
- Migration guides provided

---

## Anti-Patterns to Avoid

### ❌ Don't Do This
- Skipping tests "to save time"
- Merging without validation
- Celebrating prematurely
- Estimating completion times
- Using vague acceptance criteria
- Launching agents without Q&A
- Ignoring documentation

### ✅ Do This
- Test at every level
- Validate before merge
- Focus on accuracy
- Complete the TODO list
- Define specific, measurable criteria
- Conduct thorough Q&A before execution
- Read all relevant documentation

---

## Swarm Execution Checklist

### Pre-Execution (Overseer)
- [ ] Phase goal clearly defined
- [ ] Master test suite created
- [ ] Task breakdown complete
- [ ] Senior Developer briefed

### Design Phase (Senior Developer)
- [ ] Each agent's prompt created
- [ ] All reference docs linked
- [ ] Test specifications detailed
- [ ] Branches created
- [ ] Q&A completed with each agent
- [ ] Agent questions answered

### Execution Phase (Agents)
- [ ] Work commenced on assigned branch
- [ ] Progress commits made
- [ ] Tests written and passing
- [ ] Documentation updated
- [ ] Challenges resolved autonomously or escalated
- [ ] Final commit made

### Validation Phase (Senior Developer)
- [ ] Agent work reviewed
- [ ] Branch tests passing
- [ ] Merge to main
- [ ] Post-merge tests passing
- [ ] Issues corrected
- [ ] Task returned to Overseer

### Completion Phase (Overseer)
- [ ] All agent work integrated
- [ ] Full test suite passing
- [ ] Coverage targets met
- [ ] E2E tests passing
- [ ] User journey tests passing
- [ ] Results returned to user

---

## Example Workflow

### Overseer: Task Graph Enrichment

```markdown
PHASE: v0.29.3 Feature Implementation
TASK: Task Graph Enrichment (Issue #6)

OBJECTIVE:
Implement dependency visualization with ASCII graph rendering
and parallel execution detection.

SUBTASKS:
1. Dependency graph data structure
2. ASCII rendering engine
3. Parallel execution detector
4. Integration with existing task system

TESTS:
- Unit tests for graph construction
- Rendering tests with golden files
- Detection tests for parallel patterns
- Integration tests with real task data
```

### Senior Developer: Subtask 1

```markdown
TO: Agent Alpha
TASK: Dependency Graph Data Structure

PROMPT:
Implement a directed acyclic graph (DAG) for task dependencies.

LANGUAGE: Go
PACKAGE: internal/task
FILE: graph.go (NEW)

REQUIREMENTS:
1. Graph struct with adjacency list
2. AddNode(id, task) method
3. AddEdge(from, to) method
4. DetectCycle() method (returns error if cycle found)
5. TopologicalSort() method (returns ordered tasks)
6. GetParallelizable() method (returns tasks that can run concurrently)

REFERENCE:
- See internal/task/manager.go for existing Task type
- See internal/task/service.go for task service patterns
- Graph algorithms: https://en.wikipedia.org/wiki/Topological_sorting

TESTS:
- TestGraphConstruction
- TestCycleDetection
- TestTopologicalSort
- TestParallelDetection
- TestEdgeCases (empty graph, single node, disconnected components)

BRANCH: feature/task-graph-structure

ACCEPTANCE:
- [ ] All tests pass
- [ ] Coverage ≥ 80%
- [ ] No cycles in valid graphs
- [ ] Correct topological ordering

Any questions or suggestions?
```

### Agent Alpha Response

```markdown
Questions:
1. Should we support weighted edges (priority)?
2. What's the max expected graph size?
3. Should we cache topological sort results?

Suggestions:
1. Add String() method for debug printing
2. Add Validate() for graph integrity checks
3. Use sync.RWMutex for concurrent access
```

### Senior Developer Response

```markdown
Answers:
1. Yes - add optional Weight field to edges for future priority support
2. Max ~1000 nodes typical, 10000 max - optimize for <1000
3. Yes - cache topo sort, invalidate on AddEdge

Approved suggestions:
- String() method: Yes, helpful for debugging
- Validate(): Yes, call in AddEdge to prevent invalid state
- sync.RWMutex: Yes, good thinking on concurrency

Updated requirements in prompt. Proceed with implementation.
```

### Agent Alpha Execution

```bash
# Work begins
git checkout -b feature/task-graph-structure

# Implementation
# ... writes code ...
# ... writes tests ...

# Validation
go test ./internal/task -v -run TestGraph
# All tests pass

# Commit
git add internal/task/graph.go internal/task/graph_test.go
git commit -m "feat: implement DAG for task dependencies

- Add Graph struct with adjacency list
- Implement cycle detection
- Implement topological sort
- Add parallel task detection
- Include RWMutex for concurrency
- Tests: 15/15 passing, 87% coverage"

# Report
TO: Senior Developer
STATUS: COMPLETE
TESTS: 15/15 passing
COVERAGE: 87%
BRANCH: feature/task-graph-structure
READY FOR: Review and merge
```

---

## Summary

This swarm orchestration methodology ensures:
- **Quality**: Multi-level testing and validation
- **Accuracy**: Scientific measurement of results
- **Autonomy**: Agents work independently with escalation paths
- **Focus**: Heads-down work without distractions
- **Accountability**: Git branches track all work
- **Completeness**: TODO list drives execution

**Remember**: This is math and science. We measure. We validate. We complete the work.
