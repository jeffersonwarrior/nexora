# Verification Before Completion

Before marking ANY task complete:

## 1. VALIDATION OUTPUT REQUIRED
Paste actual command output showing pass. Never claim "it works" without proof.

## 2. TEST INTEGRITY CHECK
If tests modified, state why explicitly:
- "Added test for new edge case X"
- "Fixed flaky test that had race condition"
- "Removed obsolete test for deleted feature"

NOT acceptable:
- Silently weakening assertions
- Adding skips without reason
- Lowering coverage thresholds

## 3. NO SELF-CERTIFICATION
Claims require proof, not assertions:

Bad: "I verified this works correctly"
Good: "Output: `go test ./... -v` - 47 tests passed, 0 failed"

## 4. ONE IN-PROGRESS MAX
Complete current task before starting next. Partially-done work compounds into technical debt.

## 5. VALIDATION TIERS

Quick validation (use while iterating):
```bash
go build ./... && go vet ./...
```

Full validation (before marking complete):
```bash
go test ./... -race -coverprofile=coverage.out
```

## Why This Matters
LLMs optimize for appearing helpful over being helpful. External verification creates accountability that doesn't rely on self-reporting.
