# ADR-006: Environment Detection in System Prompt

**Date**: 2025-12-18  
**Status**: Accepted  
**Deciders**: Development Team  
**Technical Story**: Agent context awareness and runtime adaptation

## Context

AI coding assistants operate in diverse runtime environments with different capabilities, tools, and configurations. The agent needs context about its environment to:

- Make intelligent tool choices
- Provide accurate suggestions
- Avoid errors from wrong assumptions
- Adapt to platform differences

### The Problem

**Before environment detection**:
- Agent assumes generic Unix-like environment
- Suggests tools that might not exist
- Unaware of container/VM constraints
- No knowledge of installed languages/tools
- Can't optimize for available resources

**Real failure examples**:
- Suggesting `grep -P` on macOS (doesn't exist)
- Recommending `systemctl` in Docker containers
- Assuming Python 3.10 when 3.8 installed
- Proposing file operations on read-only filesystems

### Information Asymmetry

The human knows:
- What OS they're on
- What tools are installed
- Container vs bare metal
- Available memory/disk

The agent knows:
- Nothing (without environment detection)

## Decision

We will **include comprehensive environment information in the system prompt** for every session.

Environment detection includes:

### 1. Core System Info
- **OS**: Linux, macOS, Windows
- **Architecture**: amd64, arm64
- **Container Type**: Docker, LXC, systemd-nspawn, or bare metal

### 2. Runtime Versions
- **Python**: Version (e.g., 3.13.7)
- **Node.js**: Version (e.g., v25.2.1)
- **Go**: Version (e.g., go1.25.5)
- **Shell**: bash, zsh, fish

### 3. Resource Availability
- **Memory**: Total and available (e.g., 31Gi total, 20Gi available)
- **Disk**: Total and free (e.g., 630G total, 110G free)

### 4. Network & Services
- **Network Status**: online/offline
- **Active Services**: PostgreSQL, Redis, Docker, etc.

### 5. Terminal Capabilities
- **Color Support**: yes/no
- **Terminal Size**: 80x24
- **Interactive**: yes/no

### 6. Git Configuration
- **User Name**: For commit attribution
- **User Email**: For commit attribution
- **Current Repo**: Yes/no

This information is gathered once at session start and included in the system prompt as a structured section:

```markdown
## Environment
Date/Time: 2025-12-18 13:30:06 UTC
OS: linux (amd64)
Current User: agent
Working Directory: /home/nexora

Git Configuration:
- Name: Jefferson Nunn
- Email: jefferson@jeffersonnunn.com
- Repo: yes

Installed Runtimes:
- Python: Python 3.13.7
- Node.js: v25.2.1
- Go: go1.25.5
- Shell: bash (mvdan/sh)

System Resources:
- Memory: 31Gi total, 20Gi available
- Disk: 630G total, 110G free

Terminal: color, 80x24, interactive
Network: online
Active Services: PostgreSQL, Redis, Docker
```

## Consequences

### Positive

- **Intelligent suggestions**: Agent knows what tools exist
- **Platform adaptation**: Different behavior on Windows vs Linux
- **Resource awareness**: Can avoid memory-intensive operations on constrained systems
- **Better error messages**: "Python not found" vs generic suggestions
- **Git awareness**: Knows when commits are possible
- **Network awareness**: Understands when web lookups are possible
- **Container detection**: Avoids systemd commands in Docker

### Negative

- **Privacy concern**: Reveals system details
- **Prompt size**: ~200-300 tokens per session
- **Startup latency**: 300-800ms to gather info (now cached, see ADR-008)
- **Maintenance burden**: Must keep detection current
- **Stale data**: Environment might change during session

### Risks

- **Information leakage**: System details in prompt logs
  - **Mitigation**: Logs are local, not sent to untrusted servers
  - **Mitigation**: No sensitive data (passwords, keys) included
  - **Mitigation**: User controls the deployment

- **Prompt bloat**: Uses token budget
  - **Mitigation**: Only ~200-300 tokens (0.1-0.15% of 200K)
  - **Mitigation**: High value-to-token ratio
  - **Mitigation**: Can be disabled via config

- **Detection failures**: Commands might not exist
  - **Mitigation**: Graceful fallback to "unknown"
  - **Mitigation**: Parallel execution with timeouts
  - **Mitigation**: Cached results (ADR-008)

- **Misleading info**: Environment changes mid-session
  - **Mitigation**: Most values stable during session
  - **Mitigation**: Can be refreshed on demand
  - **Mitigation**: Cache TTL of 5 minutes

## Alternatives Considered

### Option A: No Environment Detection

**Description**: Agent operates blind to environment

**Pros**:
- Simple implementation
- No startup latency
- No privacy concerns
- Smaller prompts

**Cons**:
- Poor user experience
- Many avoidable errors
- Generic, unhelpful suggestions
- Can't optimize behavior

**Why not chosen**: The quality improvement is worth the cost. Blind operation leads to frustrating errors.

### Option B: Lazy Detection (On Demand)

**Description**: Only detect environment when agent asks

**Pros**:
- No startup cost
- Smaller default prompts
- Only pay cost when needed

**Cons**:
- Adds tool call overhead
- Agent must know to ask
- Inconsistent behavior
- Complexity in agent logic

**Why not chosen**: Agent shouldn't need to ask for basic info. Proactive is better than reactive.

### Option C: User-Provided Environment

**Description**: User manually specifies environment in config

**Pros**:
- Accurate (user knows best)
- No detection needed
- User controls disclosure

**Cons**:
- Manual maintenance burden
- Quickly becomes stale
- Most users won't configure
- Poor default experience

**Why not chosen**: Users shouldn't have to manually configure what can be auto-detected.

### Option D: Minimal Detection (OS Only)

**Description**: Only include OS and architecture

**Pros**:
- Fast detection
- Small prompt addition
- Covers most critical info

**Cons**:
- Misses valuable context
- Can't adapt to resource constraints
- No tool awareness
- Limited improvement

**Why not chosen**: If we're adding environment info, might as well be comprehensive. Marginal cost is low.

## Implementation Notes

### Files Affected
- `internal/agent/prompt/prompt.go` - Environment detection functions
- `internal/agent/prompt/cache.go` - Caching layer (performance optimization)
- `internal/agent/templates/coder.md.tpl` - System prompt template

### Implementation Details

Detection uses:
- `runtime.GOOS`, `runtime.GOARCH` for OS/architecture
- Shell commands for version detection (cached)
- `/proc/meminfo` for memory (Linux)
- `df -h` for disk space
- Network connectivity checks
- Service detection via port scanning or ps

### Performance

**Initial implementation**: 300-800ms sequential execution
**Optimized (ADR-008)**: 
- First call: ~300-400ms (parallel)
- Cached calls: <1ms (5-minute TTL)
- **5-10x improvement**

### Configuration

```toml
[agent]
# Enable environment detection (default: true)
environment_detection = true

# Cache TTL for environment data (default: 5m)
environment_cache_ttl = "5m"

# Include network status (adds latency, default: true unless NEXORA_FULL_ENV=0)
include_network_status = true

# Include active services (adds latency, default: true unless NEXORA_FULL_ENV=0)  
include_active_services = true
```

### Migration Path

1. ✅ Implement detection functions
2. ✅ Add caching layer (performance)
3. ✅ Integrate into system prompt
4. ✅ Test across platforms (Linux, macOS, Windows)
5. ✅ Monitor token usage and detection accuracy

### Testing Strategy

- ✅ Unit tests: Each detection function
- ✅ Integration tests: Full prompt generation
- ✅ Platform tests: Linux, macOS, Windows
- ✅ Container tests: Docker, LXC
- ✅ Performance tests: Detection latency
- ✅ Failure tests: Handle missing tools gracefully

### Privacy Considerations

Information included:
- ✅ System type (OS, arch)
- ✅ Tool versions
- ✅ Resource availability
- ✅ Git user (for commits)

Information NOT included:
- ❌ Hostnames
- ❌ IP addresses (beyond connectivity check)
- ❌ API keys or credentials
- ❌ Personal file paths
- ❌ Environment variables (except relevant ones)

### Rollback Plan

If detection causes issues:
1. Set `environment_detection = false` in config
2. Agent operates without environment context
3. Investigate and fix detection issues
4. Re-enable after verification

## References

- [Prompt Generation](../../internal/agent/prompt/prompt.go)
- [Environment Caching (ADR-008)](../../internal/agent/prompt/cache.go)
- [Performance Improvements](../../docs/PERFORMANCE_IMPROVEMENTS_2025_12_18.md)
- [System Prompt Template](../../internal/agent/templates/coder.md.tpl)

## Revision History

- **2025-12-18**: Initial draft and acceptance
