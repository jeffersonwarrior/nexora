# Systematic Debugging Framework

Four-phase approach to bug investigation and resolution.

## Phase 1: REPRODUCE
Before any code changes:

1. Identify exact steps to trigger the bug
2. Document expected vs actual behavior
3. Create minimal reproduction case if complex

```
Expected: X
Actual: Y
Steps: 1, 2, 3
```

## Phase 2: ISOLATE
Narrow down the source:

1. Binary search through commits if regression
2. Add strategic logging/breakpoints
3. Check recent changes to affected area
4. Verify assumptions about inputs/state

Questions to answer:
- When did it last work?
- What changed since then?
- Is it data-dependent or always reproducible?

## Phase 3: FIX
Surgical correction:

1. Understand root cause, not just symptoms
2. Fix the actual bug, not work around it
3. Avoid collateral changes
4. Consider edge cases introduced by fix

## Phase 4: VERIFY
Prove the fix works:

1. Run original reproduction steps - should pass
2. Add regression test if none exists
3. Run full test suite - no new failures
4. Document what caused it and why fix works

## Anti-patterns to Avoid

- Fixing symptoms without understanding cause
- Making changes hoping they'll work
- Skipping verification step
- Not adding regression test
- Introducing new bugs while fixing

## Output Format

When debugging is complete:
```
Bug: [description]
Root cause: [explanation]
Fix: [what changed and why]
Verification: [test output proving fix]
Regression test: [added/existing test name]
```
