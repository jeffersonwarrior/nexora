# Batched Edits Strategy

Combine related changes efficiently while maintaining safety.

## Good Batching

Combine when changes are:
- Logically related (same feature/fix)
- Low risk (formatting, renames, imports)
- Independent of order

Examples:
- "Fix type error in auth.go, update test, fix import"
- "Rename getUserData to fetchUserProfile across 3 files"
- "Add error handling to all API handlers"

## When NOT to Batch

Split when:
- Changes are high-risk (security, data, auth)
- Step N depends on outcome of step N-1
- Different features/concerns mixed
- Rollback granularity matters

Examples requiring separate operations:
- Database migrations
- Security-related changes
- Exploratory changes where direction might shift
- Changes to critical paths

## Batch Size Limits

| Risk Level | Max Files | Notes |
|------------|-----------|-------|
| Low (formatting) | 10+ | Safe to batch many |
| Medium (refactor) | 3-5 | Review each carefully |
| High (security) | 1 | One at a time |

## Rollback Consideration

If batch fails partially:
- All changes should be revertible together
- Avoid mixing commits that need different rollback strategies
