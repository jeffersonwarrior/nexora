# Claude-Swarm MCP Server Improvements

**Created:** 2025-12-26
**Updated:** 2025-12-26 16:10
**Priority:** CRITICAL - Fix validation enforcement + Comprehensive test coverage
**Context:** Session had 22/22 "success" but coverage stayed at 36% (target 50%)

---

## Test Coverage Implementation (In Progress)

**Goal:** Achieve comprehensive test coverage for all swarm modules (currently 10.2%)

### ✅ Completed Modules (Overall Coverage: 18.03%)

#### T1.1: utils/security.ts - 60 tests, 100% coverage
- **Bugs Found & Fixed:**
  1. Session name regex too permissive - allowed `cc-worker-feature-1` (missing hash suffix)
     - **Fix:** Enforce 6+ character hash: `/^cc-(worker|planner)-[a-zA-Z0-9_-]+-[a-z0-9]{6,}$/`
  2. **SECURITY:** validateCommand() trimmed before checking dangerous patterns
     - **Vulnerability:** Could bypass newline check with `'\nnpm test\n'`
     - **Fix:** Check dangerous patterns BEFORE trimming
- **Test Coverage:**
  - Path validation & traversal prevention
  - Feature ID validation
  - Session name validation
  - Shell quoting (injection prevention)
  - Command allowlist enforcement
  - Output sanitization

#### T1.2: utils/format.ts - 59 tests, 100% coverage
- **Bug Fixed:** Progress bar RangeError when current > total
  - **Fix:** Cap percent at 100%, prevent negative empty count
- **Test Coverage:**
  - Duration formatting (Date, milliseconds)
  - String truncation with ellipsis
  - ASCII progress bars
  - Percentage formatting
  - Average calculation

### 🔄 In Progress

#### T2.1: state/manager.ts
- State persistence, atomic writes
- Validation schema enforcement

### ⏳ Pending (Priority Order)

**Tier 2 (Depends on Tier 1):**
- T2.1: state/manager.ts - State persistence, atomic writes, validation
- T2.2: utils/feature-generator.ts - Feature list generation
- T2.3: utils/complexity-detector.ts - Feature complexity analysis
- T2.4: utils/plan-evaluator.ts - Plan quality evaluation
- T2.5: utils/prompt-templates.ts - Structured prompt generation

**Tier 3 (Depends on Tier 2):**
- T3.1: workers/confidence.ts - Confidence monitoring
- T3.2: workers/manager.ts - Worker orchestration, tmux integration

---

## Root Cause Analysis

### What Went Wrong

Workers claimed "complete" without achieving measurable targets:

| Feature | Target | Actual | Worker Claimed |
|---------|--------|--------|----------------|
| TUI components coverage | 30% | 3.2% | ✅ Complete |
| TUI pages coverage | 30% | 4.9% | ✅ Complete |
| Agent tools coverage | 50% | 17.2% → 24.2% | ✅ Complete |
| Agent core coverage | 50% | 21.2% | ✅ Complete |
| Overall coverage | 50% | 36.1% | ✅ Complete |

**Result:** 65+ TODO items still remain despite "100% success rate"

### Why It Happened

1. **No blocking validation** - Workers can mark complete without proof
2. **Trust-based completion** - Orchestrator believes worker reports
3. **Vague success criteria** - "Improve coverage" has no enforcement
4. **No incremental verification** - Coverage only checked at end
5. **Generic prompts** - Workers don't know HOW to achieve targets

---

## Priority 1: Blocking Validation System

### P1.1: Add ValidationRules to Features

**File:** `internal/server/mcp/claude-swarm/orchestrator.go`

**Current:**
```go
type Feature struct {
    ID          string
    Description string
    Status      string
    // No validation!
}
```

**Required:**
```go
type Feature struct {
    ID          string
    Description string
    Status      string
    Validation  ValidationConfig  // NEW
}

type ValidationConfig struct {
    Enabled         bool
    CoverageTarget  float64           // e.g., 50.0
    TestPassRequired bool
    CustomChecks    []ValidationCheck
    EnforceBlocking bool              // Fail if not met
}

type ValidationCheck struct {
    Name        string
    Command     string  // e.g., "go test -cover ./internal/agent/tools/..."
    Expectation string  // e.g., "coverage: 50.0%"
    Parser      func(output string) (float64, error)
}
```

**Implementation:**
- [ ] Add ValidationConfig struct
- [ ] Add validation rules to feature schema
- [ ] Implement validation execution
- [ ] Block mark_complete if validation fails

---

### P1.2: Implement ValidateFeature Function

**File:** `internal/server/mcp/claude-swarm/validation.go` (NEW)

```go
package swarm

import (
    "fmt"
    "os/exec"
    "regexp"
    "strconv"
)

func (o *Orchestrator) ValidateFeature(featureID string) (*ValidationResult, error) {
    feature := o.getFeature(featureID)
    if !feature.Validation.Enabled {
        return &ValidationResult{Passed: true}, nil
    }

    result := &ValidationResult{
        FeatureID: featureID,
        Checks:    make([]CheckResult, 0),
    }

    // Check coverage if target set
    if feature.Validation.CoverageTarget > 0 {
        actual, err := o.measureCoverage(feature.Package)
        if err != nil {
            return nil, err
        }

        passed := actual >= feature.Validation.CoverageTarget
        result.Checks = append(result.Checks, CheckResult{
            Name:   "Coverage",
            Passed: passed,
            Expected: feature.Validation.CoverageTarget,
            Actual:   actual,
        })

        if !passed && feature.Validation.EnforceBlocking {
            result.Passed = false
            result.Error = fmt.Sprintf(
                "Coverage %.1f%% < target %.1f%%",
                actual, feature.Validation.CoverageTarget,
            )
            return result, nil
        }
    }

    // Check tests pass
    if feature.Validation.TestPassRequired {
        passed, err := o.runTests(feature.Package)
        if err != nil {
            return nil, err
        }

        result.Checks = append(result.Checks, CheckResult{
            Name:   "Tests Pass",
            Passed: passed,
        })

        if !passed && feature.Validation.EnforceBlocking {
            result.Passed = false
            result.Error = "Tests failing"
            return result, nil
        }
    }

    // Run custom checks
    for _, check := range feature.Validation.CustomChecks {
        passed, actual, err := o.runCustomCheck(check)
        if err != nil {
            return nil, err
        }

        result.Checks = append(result.Checks, CheckResult{
            Name:   check.Name,
            Passed: passed,
            Actual: actual,
        })

        if !passed && feature.Validation.EnforceBlocking {
            result.Passed = false
            result.Error = fmt.Sprintf("Check '%s' failed", check.Name)
            return result, nil
        }
    }

    result.Passed = true
    return result, nil
}

func (o *Orchestrator) measureCoverage(pkg string) (float64, error) {
    cmd := exec.Command("go", "test", "-coverprofile=coverage.out", pkg)
    cmd.Dir = o.ProjectDir
    output, err := cmd.CombinedOutput()
    if err != nil {
        return 0, fmt.Errorf("coverage command failed: %w", err)
    }

    // Parse: coverage: 24.2% of statements
    re := regexp.MustCompile(`coverage:\s+(\d+\.?\d*)%`)
    matches := re.FindStringSubmatch(string(output))
    if len(matches) < 2 {
        return 0, fmt.Errorf("could not parse coverage from output")
    }

    return strconv.ParseFloat(matches[1], 64)
}
```

**Tasks:**
- [ ] Create validation.go file
- [ ] Implement ValidateFeature()
- [ ] Implement measureCoverage()
- [ ] Implement runTests()
- [ ] Implement runCustomCheck()
- [ ] Add ValidationResult types

---

### P1.3: Enforce Validation in mark_complete

**File:** `internal/server/mcp/claude-swarm/tools.go`

**Current:**
```go
func (s *Server) markComplete(featureID string, success bool, notes string) error {
    // Just marks complete, no validation!
    feature.Status = "completed"
    feature.Success = success
    return s.saveState()
}
```

**Required:**
```go
func (s *Server) markComplete(featureID string, success bool, notes string) error {
    // Validate BEFORE marking complete
    result, err := s.orchestrator.ValidateFeature(featureID)
    if err != nil {
        return fmt.Errorf("validation error: %w", err)
    }

    if !result.Passed {
        // Increment attempt counter
        feature.Attempts++

        // Auto-retry if under maxRetries
        if feature.Attempts < feature.MaxRetries {
            return s.retryFeatureWithGuidance(featureID, result)
        }

        // Max retries exceeded
        feature.Status = "failed"
        feature.FailureReason = result.Error
        s.saveState()

        return fmt.Errorf(
            "validation failed after %d attempts: %s",
            feature.Attempts,
            result.Error,
        )
    }

    // Validation passed
    feature.Status = "completed"
    feature.Success = true
    feature.ValidationResult = result
    return s.saveState()
}

func (s *Server) retryFeatureWithGuidance(featureID string, result *ValidationResult) error {
    feature := s.orchestrator.getFeature(featureID)

    // Generate guidance based on validation failure
    guidance := fmt.Sprintf(`
Validation Failed (Attempt %d/%d):
%s

Required Actions:
`, feature.Attempts, feature.MaxRetries, result.Error)

    for _, check := range result.Checks {
        if !check.Passed {
            guidance += fmt.Sprintf("- %s: Expected %.1f%%, got %.1f%%\n",
                check.Name, check.Expected, check.Actual)
        }
    }

    guidance += `
Run this command to verify:
` + feature.Validation.VerifyCommand + `

Do NOT mark complete until validation passes.
`

    // Restart worker with guidance
    return s.restartWorkerWithGuidance(featureID, guidance)
}
```

**Tasks:**
- [ ] Modify markComplete to validate first
- [ ] Add auto-retry logic with guidance
- [ ] Implement retryFeatureWithGuidance()
- [ ] Add ValidationResult to feature state

---

## Priority 2: Git-Based Verification (+ File Monitoring)

### P2.1: Git-Based Change Tracking

**File:** `internal/server/mcp/claude-swarm/git_verify.go` (NEW)

**Concept:** Use git to track worker changes instead of relying on tmux/files

```go
package swarm

import (
    "crypto/sha256"
    "fmt"
    "os/exec"
    "strings"
)

type GitVerification struct {
    BeforeHash string   // Git tree hash before worker started
    AfterHash  string   // Git tree hash after worker finished
    FilesChanged []string
    LinesAdded int
    LinesDeleted int
    Diff       string
}

func (w *Worker) CaptureBaselineState() error {
    // Get current git tree hash
    cmd := exec.Command("git", "rev-parse", "HEAD:.")
    cmd.Dir = w.ProjectDir
    output, err := cmd.Output()
    if err != nil {
        return fmt.Errorf("failed to get git hash: %w", err)
    }

    w.BaselineHash = strings.TrimSpace(string(output))

    // Also capture git diff stat for uncommitted changes
    cmd = exec.Command("git", "diff", "--stat")
    cmd.Dir = w.ProjectDir
    w.BaselineDiffStat, _ = cmd.Output()

    return nil
}

func (w *Worker) VerifyChanges() (*GitVerification, error) {
    verify := &GitVerification{
        BeforeHash: w.BaselineHash,
    }

    // Get files changed (including unstaged)
    cmd := exec.Command("git", "diff", "--name-only")
    cmd.Dir = w.ProjectDir
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    verify.FilesChanged = strings.Split(strings.TrimSpace(string(output)), "\n")

    // Get line counts
    cmd = exec.Command("git", "diff", "--numstat")
    cmd.Dir = w.ProjectDir
    output, err = cmd.Output()
    if err != nil {
        return nil, err
    }

    for _, line := range strings.Split(string(output), "\n") {
        parts := strings.Fields(line)
        if len(parts) >= 2 {
            added, _ := strconv.Atoi(parts[0])
            deleted, _ := strconv.Atoi(parts[1])
            verify.LinesAdded += added
            verify.LinesDeleted += deleted
        }
    }

    // Get full diff for review
    cmd = exec.Command("git", "diff")
    cmd.Dir = w.ProjectDir
    output, _ = cmd.Output()
    verify.Diff = string(output)

    // Calculate diff checksum
    hash := sha256.Sum256(output)
    verify.AfterHash = fmt.Sprintf("%x", hash)

    return verify, nil
}

func (w *Worker) ValidateExpectedChanges(verify *GitVerification) error {
    // Verify files match expected packages
    expectedPkgs := w.Feature.ExpectedPackages
    for _, file := range verify.FilesChanged {
        matched := false
        for _, pkg := range expectedPkgs {
            if strings.HasPrefix(file, pkg) {
                matched = true
                break
            }
        }
        if !matched {
            return fmt.Errorf(
                "unexpected file change: %s (expected packages: %v)",
                file, expectedPkgs,
            )
        }
    }

    // Verify minimum changes made
    if len(verify.FilesChanged) == 0 {
        return fmt.Errorf("no files changed - worker may not have done work")
    }

    return nil
}
```

**Advantages:**
- ✅ Immutable change tracking (git history)
- ✅ Automatic diffing and line counts
- ✅ Can verify changes match expected packages
- ✅ Checksum-based verification (sha256)
- ✅ No tmux dependency for verification

**Tasks:**
- [ ] Create git_verify.go
- [ ] Implement CaptureBaselineState()
- [ ] Implement VerifyChanges()
- [ ] Implement ValidateExpectedChanges()
- [ ] Test git-based verification

---

### P2.2: Stream Worker Output to Files (Fallback)

**File:** `internal/server/mcp/claude-swarm/worker.go`

**Current:**
```go
// Only streams to tmux, hard to monitor
func (w *Worker) Start() {
    cmd := exec.Command("tmux", "new-session", "-s", sessionName, "claude", ...)
    cmd.Start()
}
```

**Required:**
```go
func (w *Worker) Start() error {
    outputPath := filepath.Join(w.WorkersDir, fmt.Sprintf("%s.output", w.FeatureID))
    statusPath := filepath.Join(w.WorkersDir, fmt.Sprintf("%s.status", w.FeatureID))

    // Create output file
    outputFile, err := os.Create(outputPath)
    if err != nil {
        return err
    }

    // Start with tee to both tmux and file
    script := fmt.Sprintf(`
#!/bin/bash
claude %s 2>&1 | tee %s
echo "WORKER_EXITED" >> %s
`, w.Args, outputPath, statusPath)

    // Write script
    scriptPath := filepath.Join(w.WorkersDir, fmt.Sprintf("%s.sh", w.FeatureID))
    os.WriteFile(scriptPath, []byte(script), 0755)

    // Run in tmux
    cmd := exec.Command("tmux", "new-session", "-d", "-s", w.SessionName, scriptPath)
    return cmd.Start()
}

func (w *Worker) GetOutput(lines int) (string, error) {
    // Priority: .output file > tmux capture
    outputPath := filepath.Join(w.WorkersDir, fmt.Sprintf("%s.output", w.FeatureID))

    if data, err := os.ReadFile(outputPath); err == nil {
        // Return last N lines from file
        return tailLines(string(data), lines), nil
    }

    // Fallback to tmux
    return w.captureTmuxOutput(lines)
}

func (w *Worker) GetStatus() WorkerStatus {
    // Check files first
    statusPath := filepath.Join(w.WorkersDir, fmt.Sprintf("%s.status", w.FeatureID))
    donePath := filepath.Join(w.WorkersDir, fmt.Sprintf("%s.done", w.FeatureID))

    // .done file = completed
    if _, err := os.Stat(donePath); err == nil {
        return StatusCompleted
    }

    // .status file with WORKER_EXITED = completed
    if data, err := os.ReadFile(statusPath); err == nil {
        if strings.Contains(string(data), "WORKER_EXITED") {
            return StatusCompleted
        }
    }

    // Check if tmux session exists
    if w.tmuxSessionExists() {
        return StatusRunning
    }

    return StatusUnknown
}
```

**Tasks:**
- [ ] Add output file streaming
- [ ] Implement GetOutput() with file priority
- [ ] Implement GetStatus() with file checks
- [ ] Add tailLines() helper
- [ ] Test file-based monitoring

---

### P2.2: Add Real-Time Progress Tracking

**File:** `internal/server/mcp/claude-swarm/metrics.go` (NEW)

```go
package swarm

import (
    "encoding/json"
    "os"
    "time"
)

type WorkerMetrics struct {
    FeatureID      string
    StartTime      time.Time
    LastUpdate     time.Time
    Coverage       float64
    TestsPassing   int
    TestsFailing   int
    LinesChanged   int
    FilesModified  []string
}

func (w *Worker) TrackMetrics() {
    ticker := time.NewTicker(30 * time.Second)
    defer ticker.Stop()

    metricsPath := filepath.Join(w.WorkersDir, fmt.Sprintf("%s.metrics.json", w.FeatureID))

    for range ticker.C {
        if w.Status != StatusRunning {
            return
        }

        metrics := WorkerMetrics{
            FeatureID:  w.FeatureID,
            LastUpdate: time.Now(),
        }

        // Measure current coverage
        if w.Feature.Validation.CoverageTarget > 0 {
            coverage, err := measureCoverage(w.ProjectDir, w.Feature.Package)
            if err == nil {
                metrics.Coverage = coverage
            }
        }

        // Save metrics
        data, _ := json.MarshalIndent(metrics, "", "  ")
        os.WriteFile(metricsPath, data, 0644)

        // Send guidance if stalled
        if w.Feature.Validation.Enabled {
            if metrics.Coverage < w.Feature.Validation.CoverageTarget {
                w.SendGuidance(fmt.Sprintf(
                    "Coverage at %.1f%%, target is %.1f%%. Add more tests to increase coverage.",
                    metrics.Coverage,
                    w.Feature.Validation.CoverageTarget,
                ))
            }
        }
    }
}
```

**Tasks:**
- [ ] Create metrics.go file
- [ ] Implement WorkerMetrics tracking
- [ ] Add incremental guidance sending
- [ ] Test real-time progress monitoring

---

## Priority 3: Structured Prompts with Validation

### P3.1: Create Prompt Templates

**File:** `internal/server/mcp/claude-swarm/prompts.go` (NEW)

```go
package swarm

import (
    "text/template"
)

type PromptData struct {
    Task            string
    SuccessCriteria []Criterion
    ValidationCmd   string
    CurrentMetrics  WorkerMetrics
    Guidance        string
}

type Criterion struct {
    Name        string
    Description string
    Metric      string  // "coverage", "test_count", etc.
    Operator    string  // ">=", "==", "contains"
    Target      interface{}
    Current     interface{}
    Passed      bool
}

const CoveragePromptTemplate = `## Your Task
{{.Task}}

## Success Criteria (MUST achieve ALL):
{{range .SuccessCriteria}}
- [ ] {{.Name}}: {{.Description}}
  - Required: {{.Metric}} {{.Operator}} {{.Target}}
  {{if .Current}}- Current: {{.Current}} {{if not .Passed}}⚠️ NOT MET{{end}}{{end}}
{{end}}

## Validation Command
After EVERY change, run:
` + "```bash\n{{.ValidationCmd}}\n```" + `

The orchestrator will run this command automatically when you mark complete.
If validation fails, you will be asked to retry with guidance.

{{if .CurrentMetrics}}
## Current Progress
- Coverage: {{.CurrentMetrics.Coverage}}%
- Tests Passing: {{.CurrentMetrics.TestsPassing}}
- Tests Failing: {{.CurrentMetrics.TestsFailing}}
- Files Modified: {{len .CurrentMetrics.FilesModified}}
{{end}}

{{if .Guidance}}
## Guidance from Previous Attempt
{{.Guidance}}
{{end}}

## Important Rules
1. Do NOT mark feature complete until ALL criteria are met
2. Run validation command frequently during development
3. If stuck, ask for help or guidance
4. Coverage is measured automatically - focus on writing good tests

Begin implementing now.
`

func (o *Orchestrator) GeneratePrompt(feature *Feature, attempt int, previousResult *ValidationResult) (string, error) {
    data := PromptData{
        Task: feature.Description,
        ValidationCmd: feature.Validation.VerifyCommand,
    }

    // Build success criteria
    if feature.Validation.CoverageTarget > 0 {
        criterion := Criterion{
            Name:        "Test Coverage",
            Description: fmt.Sprintf("Achieve %%.1f%% test coverage", feature.Validation.CoverageTarget),
            Metric:      "coverage",
            Operator:    ">=",
            Target:      feature.Validation.CoverageTarget,
        }

        if previousResult != nil {
            for _, check := range previousResult.Checks {
                if check.Name == "Coverage" {
                    criterion.Current = check.Actual
                    criterion.Passed = check.Passed
                }
            }
        }

        data.SuccessCriteria = append(data.SuccessCriteria, criterion)
    }

    // Add guidance from previous failure
    if previousResult != nil && !previousResult.Passed {
        data.Guidance = previousResult.Error + "\n\n"
        data.Guidance += "Focus on the failing criteria above."
    }

    tmpl, err := template.New("prompt").Parse(CoveragePromptTemplate)
    if err != nil {
        return "", err
    }

    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return "", err
    }

    return buf.String(), nil
}
```

**Tasks:**
- [ ] Create prompts.go file
- [ ] Implement prompt templates
- [ ] Add criterion tracking
- [ ] Generate prompts with validation context
- [ ] Test prompt generation

---

### P3.2: Update Worker Start to Use Structured Prompts

**File:** `internal/server/mcp/claude-swarm/worker.go`

**Modify:**
```go
func (w *Worker) Start() error {
    // Generate structured prompt
    prompt, err := w.orchestrator.GeneratePrompt(w.Feature, w.Attempt, w.PreviousResult)
    if err != nil {
        return err
    }

    // Write prompt to file
    promptPath := filepath.Join(w.WorkersDir, fmt.Sprintf("%s.prompt", w.FeatureID))
    if err := os.WriteFile(promptPath, []byte(prompt), 0644); err != nil {
        return err
    }

    // Start worker with structured prompt
    // ...
}
```

**Tasks:**
- [ ] Modify worker start to generate prompts
- [ ] Save prompts to .prompt files
- [ ] Pass structured prompts to workers

---

## Priority 4: Auto-Retry with Incremental Guidance

### P4.1: Implement Smart Retry Logic

**File:** `internal/server/mcp/claude-swarm/retry.go` (NEW)

```go
package swarm

func (o *Orchestrator) retryFeatureWithGuidance(featureID string, validationResult *ValidationResult) error {
    feature := o.getFeature(featureID)

    // Generate next attempt prompt with guidance
    prompt, err := o.GeneratePrompt(feature, feature.Attempts+1, validationResult)
    if err != nil {
        return err
    }

    // Kill current worker if still running
    if err := o.killWorker(featureID); err != nil {
        log.Printf("Warning: Failed to kill worker: %v", err)
    }

    // Start new worker with updated prompt
    worker := &Worker{
        FeatureID:      featureID,
        Feature:        feature,
        Attempt:        feature.Attempts + 1,
        PreviousResult: validationResult,
        ProjectDir:     o.ProjectDir,
        WorkersDir:     o.WorkersDir,
    }

    return worker.Start()
}
```

**Tasks:**
- [ ] Create retry.go file
- [ ] Implement retryFeatureWithGuidance()
- [ ] Add worker restart logic
- [ ] Test retry with guidance

---

## Priority 5: Feature Schema Updates

### P5.1: Update Feature Definition

**File:** `internal/server/mcp/claude-swarm/schema.go`

**Current:**
```json
{
  "id": "feature-1",
  "description": "Improve test coverage...",
  "status": "pending"
}
```

**Required:**
```json
{
  "id": "feature-1",
  "description": "Improve test coverage for internal/tui/components/chat from 3.2% to 30%",
  "package": "./internal/tui/components/chat",
  "status": "pending",
  "attempts": 0,
  "maxRetries": 3,
  "validation": {
    "enabled": true,
    "coverageTarget": 30.0,
    "testPassRequired": true,
    "enforceBlocking": true,
    "verifyCommand": "go test -coverprofile=coverage.out ./internal/tui/components/chat/... && go tool cover -func=coverage.out | grep total"
  },
  "successCriteria": [
    {
      "name": "Coverage >= 30%",
      "metric": "coverage",
      "operator": ">=",
      "target": 30.0
    },
    {
      "name": "All tests pass",
      "metric": "test_pass",
      "operator": "==",
      "target": true
    }
  ]
}
```

**Tasks:**
- [ ] Update feature schema
- [ ] Add validation field
- [ ] Add package field
- [ ] Add attempts/maxRetries
- [ ] Add successCriteria
- [ ] Migrate existing features

---

## Priority 6: Test the System

### P6.1: Create Test Suite for Validation

**File:** `internal/server/mcp/claude-swarm/validation_test.go` (NEW)

```go
package swarm

func TestValidateFeature_CoverageTarget(t *testing.T) {
    // Test that validation fails when coverage < target
    feature := &Feature{
        ID: "test-1",
        Package: "./internal/test",
        Validation: ValidationConfig{
            Enabled: true,
            CoverageTarget: 50.0,
            EnforceBlocking: true,
        },
    }

    orchestrator := &Orchestrator{...}
    result, err := orchestrator.ValidateFeature(feature.ID)

    require.NoError(t, err)
    assert.False(t, result.Passed, "Should fail with low coverage")
    assert.Contains(t, result.Error, "coverage")
}

func TestMarkComplete_BlocksOnValidationFailure(t *testing.T) {
    // Test that mark_complete rejects when validation fails
    feature := &Feature{
        ID: "test-1",
        Validation: ValidationConfig{
            Enabled: true,
            CoverageTarget: 50.0,
            EnforceBlocking: true,
        },
    }

    server := &Server{...}
    err := server.markComplete("test-1", true, "Done")

    assert.Error(t, err, "Should reject completion when coverage < target")
    assert.Contains(t, err.Error(), "validation failed")
}

func TestRetryWithGuidance(t *testing.T) {
    // Test that retry generates helpful guidance
    validationResult := &ValidationResult{
        Passed: false,
        Error: "Coverage 20% < target 50%",
        Checks: []CheckResult{
            {Name: "Coverage", Passed: false, Expected: 50.0, Actual: 20.0},
        },
    }

    guidance := generateGuidance(validationResult)

    assert.Contains(t, guidance, "Coverage")
    assert.Contains(t, guidance, "20")
    assert.Contains(t, guidance, "50")
}
```

**Tasks:**
- [ ] Create validation tests
- [ ] Test coverage measurement
- [ ] Test blocking enforcement
- [ ] Test retry logic
- [ ] Test guidance generation

---

## Summary of Changes

### Files to Create (8 new files)
1. `internal/server/mcp/claude-swarm/validation.go` - Validation logic
2. `internal/server/mcp/claude-swarm/metrics.go` - Worker metrics tracking
3. `internal/server/mcp/claude-swarm/prompts.go` - Structured prompts
4. `internal/server/mcp/claude-swarm/retry.go` - Retry logic
5. `internal/server/mcp/claude-swarm/validation_test.go` - Validation tests
6. `internal/server/mcp/claude-swarm/metrics_test.go` - Metrics tests
7. `internal/server/mcp/claude-swarm/prompts_test.go` - Prompt tests
8. `internal/server/mcp/claude-swarm/retry_test.go` - Retry tests

### Files to Modify (5 files)
1. `internal/server/mcp/claude-swarm/orchestrator.go` - Add validation
2. `internal/server/mcp/claude-swarm/worker.go` - File-based monitoring
3. `internal/server/mcp/claude-swarm/tools.go` - Enforce validation in mark_complete
4. `internal/server/mcp/claude-swarm/schema.go` - Update feature schema
5. `internal/server/mcp/claude-swarm/types.go` - Add new types

### Estimated Impact
- **Development Time:** 3-5 days
- **Lines of Code:** ~1,500 new lines
- **Tests:** 20+ test cases
- **Breaking Changes:** Feature schema (need migration)

### Success Criteria
- [ ] mark_complete blocks if validation fails
- [ ] Workers auto-retry with guidance (up to maxRetries)
- [ ] Coverage enforced for all features with targets
- [ ] File-based monitoring replaces tmux-only
- [ ] Structured prompts include validation context
- [ ] Real-time metrics track progress every 30s

---

## Migration Plan

### Phase 1: Add Validation (No Breaking Changes)
1. Add validation.go (optional validation)
2. Add metrics tracking (passive)
3. Update mark_complete to validate if enabled
4. Test with one feature

### Phase 2: Enable for New Features
1. New features use validation by default
2. Old features continue without validation
3. Gradual migration

### Phase 3: Migrate All Features
1. Update existing feature definitions
2. Add validation to all coverage features
3. Set enforceBlocking = true
4. Full enforcement

### Phase 4: Make Validation Mandatory
1. Remove validation.enabled flag
2. All features must have validation
3. No completion without validation

---

## Quick Start

To implement Priority 1 (Blocking Validation):

```bash
cd /home/nexora/internal/server/mcp/claude-swarm

# Create validation.go
touch validation.go

# Add ValidationConfig to types.go
# Add ValidateFeature() to orchestrator.go
# Modify markComplete() in tools.go

# Test
go test ./... -v
```

Expected result: Workers cannot mark complete without hitting coverage targets.

---

**Next Action:** Implement P1 (Blocking Validation) first - this is the critical fix.
