# Nexora v0.29.2+ Roadmap: Multi-Agent Orchestration

> **Created:** 2025-12-26
> **Status:** Planned
> **Depends On:** v0.29.1-RC1 (Phase 4 Tool Consolidation)
>
> **Note:** No time estimates. AI swarm execution makes timing unpredictable.

---

## Version Roadmap

| Version | Focus | Key Deliverables |
|---------|-------|------------------|
| **0.29.2** | Agent hierarchy + capability cards + pre-flight | Basic orchestration, agent matching |
| **0.29.3** | Task graph enrichment + checkpoints | Recovery, staggered progress, semantic linking |
| **0.29.4** | Internal A2A + ACP communication | Agent-to-agent messaging, observation propagation |
| **0.29.5** | Protocol composition + conflict resolution | Behavioral constraints, protocol derivation |
| **0.30.x** | Tool calling overhaul + external network foundation | Rip up current system, prepare for external agents |

---

## Prerequisites

- [ ] v0.29.1-RC1 released
- [ ] Phase 4 (Tool Consolidation) complete - unified `delegate` tool required

---

# v0.29.2: Agent Hierarchy + Capability Cards

## Core Principles

1. **Always Plan First** - Show plan, wait for "execute"
2. **Parallel Swarm Execution** - 10x agents, max speed, tight contexts
3. **1 Agent Per Unit** - No file/table/component collisions
4. **Triple Test Verification** - Unit → Integration → System
5. **Ultrathink Final Review** - Always
6. **A2A Escalation** - Sub Agent → Senior → Overseer → User (last resort)
7. **Everything Git Tracked** - Branch per task, merge at end
8. **Pre-flight Validation** - Catch problems before they cascade
9. **Quality Always** - No cost/time/token considerations in prompts

---

## AI Model Classification

| Group | Purpose | Auto-Detection |
|-------|---------|----------------|
| **Thinking Models** | Planning, validation, oversight | Models with reasoning capability |
| **Coding Models** | Implementation, tests, changes | Models optimized for code |

- Minimum 30k context window (no small models)
- Per-project override in config
- Auto-detected from model capabilities

---

## Agent Hierarchy

```
                    ┌───────────────┐
                    │     USER      │  ← 5th escalation (complete failure only)
                    └───────┬───────┘
                            │
                    ┌───────────────┐
                    │   OVERSEER    │  ← 4th escalation, TUI planner
                    │  (Ultrathink) │
                    └───────┬───────┘
                            │
                    ┌───────────────┐
                    │ SENIOR AGENT  │  ← 3rd escalation, supervisor
                    │  (Thinking)   │
                    └───────┬───────┘
                            │
         ┌──────────────────┼──────────────────┐
         ↓                  ↓                  ↓
  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
  │ SUB AGENT 1 │    │ SUB AGENT 2 │    │ SUB AGENT N │
  │  (Coding)   │    │  (Coding)   │    │  (Coding)   │
  └─────────────┘    └─────────────┘    └─────────────┘
```

**Escalation Policy:**
| Attempt | Action |
|---------|--------|
| 1-2 | Same agent retries |
| 3 | Escalate to Senior Agent |
| 4 | Senior recommends → may switch model → Escalate to Overseer |
| 5 | User intervention |

---

## Agent Capability Cards

Before spawning, agents declare capabilities. Overseer matches tasks to agents by card.

```go
type AgentCard struct {
    AgentType    string      // "sub_agent", "senior", "overseer"
    Capabilities []string    // ["go_code", "sql_migrations", "tests", "frontend"]
    Constraints  []string    // ["no_frontend", "max_files:5", "no_destructive_sql"]
    ModelClass   string      // "thinking", "coding"
    MaxContext   int         // 128000, 200000
    Languages    []string    // ["go", "typescript", "sql"]
    Specialties  []string    // ["auth", "api", "db_schema", "testing"]
}
```

**Capability Matching:**
```
Task: "Implement JWT token generation"
Required: ["go_code", "auth"]
Preferred: ["security", "crypto"]

→ Overseer queries available agents
→ Matches agent-003 (capabilities: ["go_code", "auth", "security"])
→ Assigns task to agent-003
```

---

## Pre-flight Validation

Before any agent writes code, validate readiness. "Preoccupation with failure."

```go
type PreflightCheck struct {
    TaskValidation   bool    // Does the task make sense?
    DependencyCheck  bool    // Are upstream tasks complete?
    LockCheck        bool    // Is assigned file unlocked?
    CapabilityMatch  bool    // Does agent have required capabilities?
    ResourceCheck    bool    // Sufficient memory/context available?
}

func (p *PreflightCheck) Pass() bool {
    return p.TaskValidation && p.DependencyCheck &&
           p.LockCheck && p.CapabilityMatch && p.ResourceCheck
}
```

**Pre-flight Failures:**
- Task invalid → reject with explanation
- Dependency incomplete → queue until ready
- File locked → wait or reassign
- Capability mismatch → find different agent
- Resource exhausted → queue with backpressure

---

## Workflow: Plan → Pre-flight → Execute → Validate

### Planning (Always Wait for Approval)
```
USER: "Add user authentication"
              ↓
┌─────────────────────────────────────────────────────────────────┐
│ OVERSEER presents:                                               │
│                                                                  │
│ PLAN: User Authentication Implementation                         │
│ ─────────────────────────────────────                            │
│ Approach: JWT with refresh tokens                                │
│                                                                  │
│ Team: 6 agents (matched by capability)                           │
│   • agent-001: auth/jwt.go        [go_code, auth, security]      │
│   • agent-002: auth/middleware.go [go_code, auth]                │
│   • agent-003: db/users_table.sql [sql, db_schema]               │
│   • agent-004: api/auth_handlers  [go_code, api]                 │
│   • agent-005: tests/auth_test.go [go_code, testing]             │
│   • agent-006: tests/integration  [go_code, testing, integration]│
│                                                                  │
│ Dependencies:                                                    │
│   agent-003 (db) ──→ agent-001 (jwt) ──→ agent-002 (middleware)  │
│                                      └──→ agent-004 (handlers)   │
│   agent-005, agent-006 run after all implementation complete     │
│                                                                  │
│ Git: Branch 'feature/auth' → work → test → merge to main         │
│                                                                  │
│ [Waiting for: "execute"]                                         │
└─────────────────────────────────────────────────────────────────┘
```

### Pre-flight (Automatic)
```
Pre-flight for agent-001 (auth/jwt.go):
  ✓ Task validation: valid
  ✓ Dependencies: none (first in chain)
  ✓ Lock check: auth/jwt.go available
  ✓ Capability match: [go_code, auth, security] ⊇ [go_code, auth]
  ✓ Resources: 128k context available
  → CLEARED FOR EXECUTION
```

### Parallel Execution
- Agents spawn respecting dependency order
- Each agent works on assigned unit ONLY
- 1 agent per file/table/component (no collisions)
- Tight context: right prompt + right tools + right scope

### Triple Verification (Automatic)
1. **Unit Tests** - Each agent runs tests for its changes
2. **Integration Tests** - Senior Agent runs cross-component tests
3. **System Tests** - Overseer runs full system validation

### Final Review
- Ultrathink AI reviews all changes against Overview
- If issues found → generate revised TODOs → retry
- If approved → merge to main (when user confirms)

---

## TUI Agent Status Panel

**Location:** Bottom Right (~23×12 chars)
```
┌─────────────────────┐
│ AGENTS              │
│ ● ● ● ○ ○ ◐        │
│ 001 002 003 004 005 │
│ ✓   ✓   ✓   ...  Q  │
│                     │
│ Run: gentle-fox-42  │
└─────────────────────┘

Legend:
● = Running (green)
○ = Queued (gray)
◐ = In Progress (yellow)
✓ = Complete (green, then fade)
✗ = Failed (red)
```

---

## Delegation Modes (Tab to Cycle)

| Mode | Behavior |
|------|----------|
| `auto` | Delegate everything automatically |
| `ask` | Show plan, wait for approval (default) |
| `manual` | Never delegate, execute in chat |

---

## Collision Avoidance

```
UNIT ASSIGNMENT RULES:
• 1 agent per FILE (maximum)
• 1 agent per TABLE (maximum)
• 1 agent per API endpoint (maximum)
• 1 agent per component (maximum)

If layered system needed:
1. Ultrathink creates SCHEMA MAP first (ASCII art)
2. DB layer completes first
3. Backend layer second
4. API layer third
5. Frontend layer last
```

---

## Database Schema: v0.29.2

```sql
-- Core orchestration tables
CREATE TABLE delegation_templates (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    template_yaml TEXT NOT NULL,
    conflict_policy TEXT DEFAULT 'escalate',  -- first_writer_wins, merge_required, escalate
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE prompt_plans (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    description TEXT,
    plan_yaml TEXT NOT NULL,
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

CREATE TABLE agent_runs (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,  -- gentle-fox-42
    task_description TEXT,
    template_id TEXT REFERENCES delegation_templates(id),
    status TEXT NOT NULL,  -- planning, preflight, executing, validating, complete, failed
    git_branch TEXT,
    created_at INTEGER NOT NULL,
    completed_at INTEGER
);

CREATE TABLE agent_instances (
    id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL,
    agent_name TEXT NOT NULL,  -- agent-001-test-writer
    model TEXT NOT NULL,
    agent_type TEXT NOT NULL,  -- overseer, senior, sub_agent
    assigned_unit TEXT,  -- file/table/component
    status TEXT NOT NULL,  -- queued, preflight, running, complete, failed
    attempts INTEGER DEFAULT 0,
    -- Capability card fields
    capabilities TEXT,  -- JSON array: ["go_code", "auth", "security"]
    constraints TEXT,   -- JSON array: ["no_frontend", "max_files:5"]
    model_class TEXT,   -- "thinking", "coding"
    max_context INTEGER,
    created_at INTEGER NOT NULL,
    completed_at INTEGER,
    FOREIGN KEY (run_id) REFERENCES agent_runs(id)
);

-- Pre-flight validation log
CREATE TABLE preflight_checks (
    id TEXT PRIMARY KEY,
    agent_instance_id TEXT NOT NULL,
    task_valid BOOLEAN NOT NULL,
    deps_ready BOOLEAN NOT NULL,
    lock_available BOOLEAN NOT NULL,
    capability_match BOOLEAN NOT NULL,
    resource_available BOOLEAN NOT NULL,
    passed BOOLEAN NOT NULL,
    failure_reason TEXT,
    checked_at INTEGER NOT NULL,
    FOREIGN KEY (agent_instance_id) REFERENCES agent_instances(id)
);
```

---

## Files to Create: v0.29.2

| File | Purpose |
|------|---------|
| `internal/agent/orchestrator/planner.go` | Plan generation |
| `internal/agent/orchestrator/executor.go` | Parallel execution |
| `internal/agent/orchestrator/preflight.go` | Pre-flight validation |
| `internal/agent/orchestrator/capability.go` | Agent capability matching |
| `internal/agent/orchestrator/escalation.go` | A2A escalation |
| `internal/tui/components/agent_status/` | Agent status panel |
| `internal/db/migrations/xxx_orchestration_v1.sql` | New tables |

---

## Testing Checklist: v0.29.2

- [ ] Delegation mode cycles with Tab key
- [ ] Agent status panel displays correctly
- [ ] Plan shown and waits for "execute"
- [ ] Capability matching works (agent with wrong caps rejected)
- [ ] Pre-flight validation blocks invalid tasks
- [ ] Agents spawn in parallel respecting dependencies
- [ ] 1 agent per file enforced (collision error if violated)
- [ ] A2A escalation triggers on 3rd failure
- [ ] Git branch created per task
- [ ] Triple verification runs automatically
- [ ] `nexora runs list` shows active/completed runs

---

# v0.29.3: Task Graph Enrichment + Checkpoints

## Staggered Progress Checkpoints

Agents get credit for partial progress. Helps with debugging failures.

```go
type TaskCheckpoint struct {
    Milestone string   // "interface_defined", "core_logic", "tests_written", "tests_pass"
    Weight    float64  // 0.2, 0.3, 0.2, 0.3
    Completed bool
}

type StaggeredProgress struct {
    TaskID      string
    Checkpoints []TaskCheckpoint
    Progress    float64  // Computed: sum of completed weights
}
```

**Example:**
```yaml
task: "Implement auth middleware"
checkpoints:
  - milestone: "Interface defined"
    weight: 0.2
  - milestone: "Core logic implemented"
    weight: 0.3
  - milestone: "Tests written"
    weight: 0.2
  - milestone: "Tests pass"
    weight: 0.3
```

Agent at 50% (interface + core logic) can be resumed by replacement agent.

---

## Recovery Checkpoints

Agents dump state periodically. On failure, can resume mid-task.

```go
type AgentCheckpoint struct {
    ID           string
    AgentID      string
    TaskID       string
    Progress     float64           // 0.0-1.0
    FilesWritten []string          // Files created/modified so far
    Decisions    []Decision        // Why they chose X over Y
    NextStep     string            // What they were about to do
    Context      string            // Serialized conversation state
    Timestamp    time.Time
}
```

**Recovery Flow:**
```
agent-003 fails at 60% progress
  → Last checkpoint at 55%
  → Senior spawns replacement agent-003b
  → agent-003b loads checkpoint
  → Resumes from 55% with full context
  → Completes remaining 45%
```

---

## Semantic Task Linking

Tasks know their upstream/downstream dependencies explicitly.

```go
type TaskLink struct {
    TaskID      string
    DependsOn   []Dependency  // Upstream tasks
    Publishes   []Interface   // What this task provides
    Consumers   []string      // Downstream task IDs
}

type Dependency struct {
    TaskID   string
    Requires []string  // Specific interfaces needed: ["GenerateToken", "ValidateToken"]
}

type Interface struct {
    Name      string
    Signature string  // "func GenerateToken(userID string) (string, error)"
}
```

**Automatic Notifications:**
```
feature-3 changes GenerateToken signature
  → System identifies consumers: [feature-4, feature-5]
  → Notifies agent-004, agent-005 of interface change
  → Agents adapt their implementation
```

---

## Database Schema: v0.29.3 Additions

```sql
-- Recovery checkpoints
CREATE TABLE agent_checkpoints (
    id TEXT PRIMARY KEY,
    agent_instance_id TEXT NOT NULL,
    task_id TEXT NOT NULL,
    progress REAL NOT NULL,  -- 0.0-1.0
    files_written TEXT,      -- JSON array
    decisions TEXT,          -- JSON array of decision objects
    next_step TEXT,
    context_snapshot TEXT,   -- Serialized conversation
    created_at INTEGER NOT NULL,
    FOREIGN KEY (agent_instance_id) REFERENCES agent_instances(id)
);

-- Staggered progress milestones
CREATE TABLE task_milestones (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    milestone TEXT NOT NULL,
    weight REAL NOT NULL,
    completed BOOLEAN DEFAULT FALSE,
    completed_at INTEGER,
    agent_instance_id TEXT,
    FOREIGN KEY (agent_instance_id) REFERENCES agent_instances(id)
);

-- Semantic task linking
CREATE TABLE task_links (
    id TEXT PRIMARY KEY,
    task_id TEXT NOT NULL,
    link_type TEXT NOT NULL,  -- 'depends_on', 'publishes', 'consumes'
    target_task_id TEXT,      -- For depends_on/consumes
    interface_name TEXT,      -- For publishes
    interface_signature TEXT, -- For publishes
    required_interfaces TEXT, -- JSON array for depends_on
    created_at INTEGER NOT NULL
);

CREATE INDEX idx_task_links_task ON task_links(task_id);
CREATE INDEX idx_task_links_target ON task_links(target_task_id);
```

---

## Files to Create: v0.29.3

| File | Purpose |
|------|---------|
| `internal/agent/orchestrator/checkpoint.go` | Checkpoint save/restore |
| `internal/agent/orchestrator/recovery.go` | Agent recovery from checkpoints |
| `internal/agent/orchestrator/progress.go` | Staggered progress tracking |
| `internal/agent/orchestrator/tasklink.go` | Semantic dependency management |
| `internal/db/migrations/xxx_checkpoints.sql` | Checkpoint tables |

---

## Testing Checklist: v0.29.3

- [ ] Checkpoints saved at configurable intervals
- [ ] Failed agent can be recovered from last checkpoint
- [ ] Staggered progress shows partial completion
- [ ] Task links correctly track dependencies
- [ ] Interface changes propagate to dependent tasks
- [ ] Progress persists across TUI restart

---

# v0.29.4: Internal A2A + ACP Communication

## Agent-to-Agent Protocol (A2A) - Internal

Agents communicate directly for coordination. Based on A2A spec principles.

```go
type A2AMessage struct {
    ID          string
    FromAgent   string    // agent-001
    ToAgent     string    // agent-002 or "senior" or "broadcast"
    MessageType string    // "query", "notify", "request", "response"
    Payload     any
    Timestamp   time.Time
}

// Message types
type QueryMessage struct {
    Question string
    Context  string
}

type NotifyMessage struct {
    Event   string  // "interface_changed", "task_complete", "conflict_detected"
    Details any
}

type RequestMessage struct {
    Action string
    Params map[string]any
}

type ResponseMessage struct {
    RequestID string
    Success   bool
    Result    any
    Error     string
}
```

---

## Sibling Observation Propagation

Agents learn from each other's successes/failures during a run.

```go
type Observation struct {
    AgentID     string
    TaskID      string
    Pattern     string              // "jwt_signing", "error_handling", "test_structure"
    Approach    string              // "RS256", "table-driven tests"
    FilesTouched []string
    Success     bool
    Timestamp   time.Time
}

// Broadcast to siblings
func (a *Agent) BroadcastObservation(obs Observation) {
    msg := A2AMessage{
        FromAgent:   a.ID,
        ToAgent:     "broadcast",
        MessageType: "notify",
        Payload: NotifyMessage{
            Event:   "observation",
            Details: obs,
        },
    }
    a.orchestrator.Broadcast(msg)
}

// Receive and adapt
func (a *Agent) HandleObservation(obs Observation) {
    if a.IsRelevant(obs) {
        a.AdaptApproach(obs)
    }
}
```

**Example Flow:**
```
agent-001 completes auth/jwt.go
  → broadcasts: {pattern: "jwt_signing", approach: "RS256", files: [...]}

agent-002 (working on middleware.go) receives observation
  → recognizes relevance: "I need to validate tokens"
  → adapts approach to match agent-001's RS256 format
  → avoids incompatibility
```

---

## Agent Communication Protocol (ACP) - Internal

REST-native messaging for structured agent interactions.

```go
type ACPEndpoint struct {
    AgentID  string
    Endpoint string  // "/status", "/query", "/task"
    Methods  []string // ["GET", "POST"]
}

type ACPRequest struct {
    Method  string
    Path    string
    Headers map[string]string
    Body    any
}

type ACPResponse struct {
    Status  int
    Headers map[string]string
    Body    any
}

// Agent exposes ACP endpoints
func (a *Agent) RegisterEndpoints() []ACPEndpoint {
    return []ACPEndpoint{
        {a.ID, "/status", []string{"GET"}},
        {a.ID, "/query", []string{"POST"}},
        {a.ID, "/interface", []string{"GET"}},  // What interfaces I provide
    }
}
```

---

## A2A Communication Display

```
[A2A] agent-001 → broadcast: Observation{jwt_signing, RS256}
[A2A] agent-002 ← received observation, adapting approach
[A2A] agent-003 → Senior: "Schema conflict with existing users table"
[A2A] Senior → agent-003: "Use migration, preserve existing columns"
[A2A] agent-004 → agent-001: Query{token format for validation?}
[A2A] agent-001 → agent-004: Response{JWT with RS256, exp claim}
```

---

## Database Schema: v0.29.4 Additions

```sql
-- A2A message log
CREATE TABLE a2a_messages (
    id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL,
    from_agent TEXT NOT NULL,
    to_agent TEXT NOT NULL,  -- agent ID or "broadcast" or "senior"
    message_type TEXT NOT NULL,  -- query, notify, request, response
    payload TEXT NOT NULL,  -- JSON
    created_at INTEGER NOT NULL,
    FOREIGN KEY (run_id) REFERENCES agent_runs(id)
);

CREATE INDEX idx_a2a_run ON a2a_messages(run_id);
CREATE INDEX idx_a2a_from ON a2a_messages(from_agent);
CREATE INDEX idx_a2a_to ON a2a_messages(to_agent);

-- Observations for propagation
CREATE TABLE agent_observations (
    id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL,
    agent_id TEXT NOT NULL,
    task_id TEXT NOT NULL,
    pattern TEXT NOT NULL,
    approach TEXT,
    files_touched TEXT,  -- JSON array
    success BOOLEAN NOT NULL,
    propagated_to TEXT,  -- JSON array of agent IDs that received this
    created_at INTEGER NOT NULL,
    FOREIGN KEY (run_id) REFERENCES agent_runs(id)
);
```

---

## Files to Create: v0.29.4

| File | Purpose |
|------|---------|
| `internal/agent/a2a/protocol.go` | A2A message definitions |
| `internal/agent/a2a/router.go` | Message routing between agents |
| `internal/agent/a2a/broadcast.go` | Broadcast handling |
| `internal/agent/acp/server.go` | ACP endpoint server |
| `internal/agent/acp/client.go` | ACP client for queries |
| `internal/agent/observation/propagate.go` | Observation broadcasting |
| `internal/db/migrations/xxx_a2a_tables.sql` | Communication tables |

---

## Testing Checklist: v0.29.4

- [ ] A2A messages route correctly between agents
- [ ] Broadcast reaches all active agents in run
- [ ] Observations propagate to relevant siblings
- [ ] Agents adapt approach based on observations
- [ ] ACP endpoints respond correctly
- [ ] Message log persists for debugging
- [ ] TUI displays A2A communication

---

# v0.29.5: Protocol Composition + Conflict Resolution

## Behavioral Protocols

Protocols define enforceable behavioral constraints.

```go
type Protocol struct {
    ID          string
    Name        string
    Version     string
    DependsOn   []string        // Protocol inheritance
    Constraints []Constraint    // Behavioral bounds
    Triggers    []Trigger       // State-based triggers
    Enforcement string          // "soft" (log warning), "hard" (block action)
}

type Constraint struct {
    Name      string
    Condition string  // Expression: "files_modified <= 5"
    Message   string  // "Agent cannot modify more than 5 files per task"
}

type Trigger struct {
    Event   string  // "file_conflict", "test_failure", "timeout"
    Action  string  // "escalate", "retry", "notify", "abort"
    Handler string  // Function name or protocol reference
}
```

**Protocol Inheritance:**
```yaml
# Base protocol
protocol: code_quality
version: "1.0"
constraints:
  - name: "test_coverage"
    condition: "coverage >= 0.5"
    message: "Must maintain 50% test coverage"

# Derived protocol
protocol: strict_code_quality
version: "1.0"
depends_on: ["code_quality"]
constraints:
  - name: "no_todos"
    condition: "todo_count == 0"
    message: "No TODO comments in production code"
```

---

## Conflict Resolution Protocol

When agents have overlapping concerns.

```go
type ConflictPolicy struct {
    Policy      string    // "first_writer_wins", "merge_required", "escalate"
    LockTimeout int       // Seconds to hold lock
    MergeAgent  string    // Optional: dedicated agent for merges
    Arbitrator  string    // "senior" or "overseer" for escalate policy
}

type Conflict struct {
    ID         string
    RunID      string
    Agent1     string
    Agent2     string
    Resource   string    // File, table, interface
    Type       string    // "write_conflict", "interface_mismatch", "import_conflict"
    Resolution string    // How it was resolved
    ResolvedBy string    // Agent or policy
    Timestamp  time.Time
}
```

**Conflict Types:**
- **Write conflict**: Two agents want same file
- **Interface mismatch**: Consumer expects different signature than producer
- **Import conflict**: Incompatible dependency versions

**Resolution Flow:**
```
agent-002 tries to write auth/middleware.go
  → Lock check: agent-001 has lock
  → Policy: "escalate"
  → Senior Agent notified
  → Senior: "agent-001 has priority, agent-002 wait"
  → agent-002 queued until agent-001 releases lock
```

---

## Protocol Enforcement

```go
type ProtocolEnforcer struct {
    Protocols map[string]Protocol
    Mode      string  // "strict", "permissive"
}

func (e *ProtocolEnforcer) Check(action AgentAction) (bool, []Violation) {
    var violations []Violation
    for _, proto := range e.Protocols {
        for _, constraint := range proto.Constraints {
            if !e.Evaluate(constraint, action) {
                violations = append(violations, Violation{
                    Protocol:   proto.ID,
                    Constraint: constraint.Name,
                    Message:    constraint.Message,
                    Severity:   proto.Enforcement,
                })
            }
        }
    }

    if e.Mode == "strict" && len(violations) > 0 {
        return false, violations
    }
    return true, violations
}
```

---

## Database Schema: v0.29.5 Additions

```sql
-- Protocol definitions
CREATE TABLE protocols (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    version TEXT NOT NULL,
    depends_on TEXT,  -- JSON array of protocol IDs
    constraints TEXT NOT NULL,  -- JSON array
    triggers TEXT,  -- JSON array
    enforcement TEXT DEFAULT 'soft',  -- soft, hard
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL
);

-- Protocol violations log
CREATE TABLE protocol_violations (
    id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL,
    agent_id TEXT NOT NULL,
    protocol_id TEXT NOT NULL,
    constraint_name TEXT NOT NULL,
    message TEXT,
    severity TEXT NOT NULL,  -- soft, hard
    action_blocked BOOLEAN NOT NULL,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (run_id) REFERENCES agent_runs(id),
    FOREIGN KEY (protocol_id) REFERENCES protocols(id)
);

-- Conflict tracking
CREATE TABLE conflicts (
    id TEXT PRIMARY KEY,
    run_id TEXT NOT NULL,
    agent1_id TEXT NOT NULL,
    agent2_id TEXT NOT NULL,
    resource TEXT NOT NULL,
    conflict_type TEXT NOT NULL,
    resolution TEXT,
    resolved_by TEXT,
    resolved_at INTEGER,
    created_at INTEGER NOT NULL,
    FOREIGN KEY (run_id) REFERENCES agent_runs(id)
);
```

---

## Files to Create: v0.29.5

| File | Purpose |
|------|---------|
| `internal/agent/protocol/definition.go` | Protocol data structures |
| `internal/agent/protocol/enforcer.go` | Constraint checking |
| `internal/agent/protocol/inheritance.go` | Protocol composition |
| `internal/agent/conflict/detector.go` | Conflict detection |
| `internal/agent/conflict/resolver.go` | Resolution strategies |
| `internal/agent/conflict/lock.go` | Resource locking |
| `internal/db/migrations/xxx_protocols.sql` | Protocol tables |

---

## Testing Checklist: v0.29.5

- [ ] Protocols load and parse correctly
- [ ] Protocol inheritance resolves dependencies
- [ ] Soft constraints log warnings but don't block
- [ ] Hard constraints block violating actions
- [ ] Conflict detection catches write conflicts
- [ ] Escalation policy routes to Senior/Overseer
- [ ] Lock timeout releases stuck locks
- [ ] Violations logged for debugging

---

# Implementation Priority (All Versions)

## v0.29.2
1. Database schema (delegation_templates, agent_instances with capabilities)
2. Agent capability cards and matching
3. Pre-flight validation
4. Basic orchestrator (Plan → Execute → Validate)
5. Agent status panel
6. Delegation mode cycling

## v0.29.3
1. Checkpoint tables and save/restore
2. Staggered progress milestones
3. Task linking and dependency tracking
4. Recovery from checkpoints
5. Interface change propagation

## v0.29.4
1. A2A message protocol and routing
2. Observation propagation system
3. ACP endpoints for agents
4. TUI A2A communication display
5. Message logging

## v0.29.5
1. Protocol definition schema
2. Protocol inheritance/composition
3. Constraint enforcement
4. Conflict detection and resolution
5. Resource locking

---

# Full Testing Checklist

## v0.29.2
- [ ] Capability matching assigns correct agents
- [ ] Pre-flight blocks invalid tasks
- [ ] Agents spawn respecting dependencies
- [ ] 1 agent per file enforced
- [ ] Escalation triggers on 3rd failure

## v0.29.3
- [ ] Checkpoints save/restore correctly
- [ ] Failed agents resume from checkpoint
- [ ] Progress shows partial completion
- [ ] Dependency changes notify consumers

## v0.29.4
- [ ] A2A messages route correctly
- [ ] Observations propagate to siblings
- [ ] Agents adapt based on observations
- [ ] TUI shows communication log

## v0.29.5
- [ ] Protocols enforce constraints
- [ ] Inheritance resolves correctly
- [ ] Conflicts detected and resolved
- [ ] Violations logged

## Integration
- [ ] Full workflow: Plan → Pre-flight → Execute → Checkpoint → Verify
- [ ] Recovery from mid-task failure
- [ ] Multi-agent coordination via A2A
- [ ] Protocol violations handled gracefully
- [ ] `nexora runs list` shows complete history

---

# v3.0: Visual Terminal Interaction

## Version Roadmap

| Version | Focus | Key Deliverables |
|---------|-------|------------------|
| **v0.29.2** | Agent hierarchy + capability cards + pre-flight | Basic orchestration, agent matching |
| **v0.29.3** | Task graph enrichment + checkpoints | Recovery, staggered progress, semantic linking |
| **v0.29.4** | Internal A2A + ACP communication | Agent-to-agent messaging, observation propagation |
| **v0.29.5** | Protocol composition + conflict resolution | Behavioral constraints, protocol derivation |
| **v3.0** | Visual terminal + multi-provider routing | ModelScan integration, VNC/Docker, dual-mode |

---

## Prerequisites

- [ ] v0.29.1-RC1 released
- [ ] v0.29.2-0.29.5 complete (optional - can proceed in parallel)

---

## Phase 0: ModelScan Integration Priority (Week 0-2)

### Rethinking: ModelScan First
Replace Fantasy with ModelScan BEFORE adding VNC to:
- Clean up provider architecture
- Reduce immediate user pain
- Less moving parts when adding VNC later

### Week 0: Pre-Integration
- [ ] Remove existing provider hardcoding
- [ ] Audit Fantasy vs ModelScan API differences
- [ ] Create ModelScan provider wrapper interface
- [ ] Update configuration system for ModelScan routing

### Week 1: Core Integration
- [ ] Replace Fantasy client with ModelScan router
- [ ] Implement ModelScan provider selection in config
- [ ] Add routing options (cheapest, fastest, fallback)
- [ ] Update error handling for ModelScan responses

### Week 2: Validation
- [ ] Test with all supported providers (OpenAI, Anthropic, Mistral, xAI)
- [ ] Verify streaming responses still work
- [ ] Add performance metrics for routing decisions
- [ ] Fix any broken tools that depend on provider specifics

---

## Phase 1: Database Foundation (Week 2-3)

### PostgreSQL Implementation
- [ ] Install PostgreSQL schemas (sessions, port_allocations, providers_auth, vnc_sessions)
- [ ] Create PostgreSQL connection pool with pgx/v5
- [ ] Migrate session tracking from SQLite
- [ ] Implement API key encryption (AES-256-GCM)
- [ ] Add database migration utilities

### SQLite Support (Lite Mode)
- [ ] Create SQLite version of same schemas
- [ ] Implement dual database abstraction layer
- [ ] Add database type detection in config
- [ ] Test both databases with same operations

---

## Phase 2: Docker Infrastructure (Week 3-4)

### Container Build
- [ ] Create Ubuntu 24.04 Dockerfile with all dev tools
- [ ] Add ARM-specific adjustments for Mac Silicon
- [ ] Optimize image size (target under 2GB)
- [ ] Create startup scripts (X11, VNC, Chrome)
- [ ] Implement health checks and status signals

### Container Management
- [ ] Build Docker lifecycle manager
- [ ] Implement port allocation (VNC: 5900-5999, CDP: 9222-9321)
- [ ] Add workspace mounting with user permissions
- [ ] Create container startup sequence with proper error handling
- [ ] Implement orphaned container cleanup

---

## Phase 3: VNC Tools Implementation (Week 5-6)

### Screen Capture
- [ ] Implement VNC client connection
- [ ] Capture framebuffer and convert to PNG
- [ ] Add basic text extraction (ocr)
- [ ] Optimize capture frequency (100-200ms)
- [ ] Handle screen resize and resolution changes

### Keyboard Input
- [ ] Implement xdotool integration for typing
- [ ] Support special keys (Escape, Tab, Ctrl+*)
- [ ] Handle modifier keys correctly
- [ ] Add input validation and sanitization
- [ ] Implement typing rate limiting

### Execute Command
- [ ] Add direct docker exec for non-visual operations
- [ ] Implement stdout/stderr capture
- [ ] Add exit code tracking
- [ ] Handle long-running commands
- [ ] Add command timeout management

### Tool Integration
- [ ] Create fantasy tools wrapper for VNC operations
- [ ] Integrate ModelScan with VNC tools
- [ ] Update agent prompts to use screen/keyboard/execute
- [ ] Add fallback to old tools for Lite mode

---

## Phase 4: Session Management (Week 7-8)

### Session Lifecycle
- [ ] Implement session start with database tracking
- [ ] Add container binding per session
- [ ] Create session state persistence
- [ ] Implement graceful session termination
- [ ] Add session recovery after crashes

### Mode Selection
- [ ] Build installer with mode selection (Lite/Full)
- [ ] Implement mode detection in code paths
- [ ] Add configuration validation per mode
- [ ] Create mode-specific help messages
- [ ] Test switching requires reinstall (no in-place upgrade)

### Error Handling
- [ ] Add container crash detection
- [ ] Implement automatic container restart
- [ ] Create session state restoration
- [ ] Add port conflict resolution
- [ ] Handle Docker daemon failures

---

## Phase 5: Integration & Testing (Week 9-10)

### End-to-End Testing
- [ ] Test complete workflows in both modes
- [ ] Verify ModelScan routing works with VNC
- [ ] Test concurrent sessions (up to 10)
- [ ] Validate resource cleanup
- [ ] Test ARM vs x86 architecture differences

### Performance Validation
- [ ] Measure screen capture latency
- [ ] Validate VNC vs tool speed improvements
- [ ] Test ModelScan routing performance
- [ ] Monitor memory usage per session
- [ ] Verify resource limits enforcement

### Documentation
- [ ] Create unified README.md with mode descriptions
- [ ] Document VNC justification (examples of broken tools)
- [ ] Add troubleshooting guide
- [ ] Create installation guide with mode selection
- [ ] Document ModelScan configuration options

---

## Phase 6: Release Preparation (Week 11)

### Final Integration
- [ ] Remove legacy Fantasy code entirely
- [ ] Clean up unused tool abstractions
- [ ] Finalize configuration defaults
- [ ] Add startup diagnostics
- [ ] Implement version checking

### Packaging & Distribution
- [ ] Build release binaries for all platforms
- [ ] Create Docker image for easy deployment
- [ ] Add automated tests to CI
- [ ] Prepare release notes
- [ ] Update GitHub releases

---

## Critical Path & Dependencies

### Blockers (must complete first):
1. **ModelScan integration (Phase 0)** - Nothing works without it
2. **Database layer (Phase 1)** - Required for VNC session management
3. **Docker infrastructure (Phase 2)** - No VNC without containers

### Parallelizable Work:
- SQLite compatibility (alongside PostgreSQL)
- Tool migration (works with both databases)
- Documentation (can be written incrementally)

### Timeline Summary:
- **Week 0-2**: ModelScan integration (HIGH PRIORITY)
- **Week 3-4**: Database + Docker foundations
- **Week 5-6**: Core VNC implementation
- **Week 7-8**: Session management
- **Week 9-10**: Integration testing
- **Week 11**: Release prep

**Total: 11 weeks to 3.0 release**

---

## Architecture Notes

### ModelScan Configuration Migration
```yaml
# Old config (multiple providers)
providers:
  mistral: {api_key: "..."}
  openai: {api_key: "..."}

# New config (ModelScan routing)
modelscan:
  enabled: true
  routing: "cheapest"  # or fastest, balanced, fallback
  providers:
    mistral: {api_key: "...", priority: 1, cost: 0.15}
    openai: {api_key: "...", priority: 2, cost: 0.30}
```

### Why ModelScan First
- Reduces risk - immediate user value even if VNC delayed
- Cleaner architecture for VNC development
- Faster provider switching and cost optimization
- Foundation for 4.0 multi-agent architecture

### Mode Summary
**Lite Mode**: SQLite + ModelScan + CLI tools (servers, laptops)
**Full Mode**: PostgreSQL + ModelScan + VNC (visual pair-programming)

---

## Known Issues to Address

### Session Title Re-generation
**Issue**: Sessions that already have "New Session" as title are not retitled when first message is sent.

**Current Behavior**:
- New sessions get "New Session" as default title
- When first message is sent, `generateTitle()` runs but only updates the session object
- If the session already exists with "New Session", it's not properly detected as needing a title

**Expected Behavior**:
- First message should always generate a proper title
- "New Session" should be treated as placeholder that needs replacement

**Root Cause**:
- `generateTitle()` checks `MessageCount == 0` but doesn't check if current title is placeholder
- Race condition possible: session created → title set to "New Session" → message added → count becomes 1

**Possible Solutions**:
1. Check both `MessageCount == 0 OR title == "New Session"`
2. Add `needs_title` boolean flag to session schema
3. Check if title equals any default placeholder values
4. Always regenerate title if it matches default patterns

**Priority**: Medium (UX issue, not blocking)
