# Session Title Generation Fix

## Problem Statement
Sessions with "New Session" as title don't get retitled on first message.

## Root Cause
In `/home/nexora/internal/agent/agent.go` at line 540, the code checks:
```go
if currentSession.MessageCount == 0 {
    wg.Go(func() {
        sessionLock.Lock()
        a.generateTitle(ctx, &currentSession, call.Prompt)
        sessionLock.Unlock()
    })
}
```

The issue is that `MessageCount` might not be 0 even though the title is a default placeholder. This happens when:
1. A session is created with a specific title (e.g., "Project Review")
2. The title generation happens correctly
3. But if `MessageCount` is incremented before title generation completes, it won't regenerate

Wait, that's not right. Let me re-read the TODO.md more carefully.

From TODO.md line 54:
> **Root Cause:** `generateTitle()` checks `MessageCount == 0` but doesn't check if current title is placeholder.

So the issue is simpler: If a session is created with title "New Session" (the default), and MessageCount is 0, then `generateTitle()` should be called. But the current logic only checks MessageCount == 0, which should work.

Let me look at the session creation flow more carefully.

## Session Creation Flow
1. Session created with title "New Session" (default in `session.Create()`)
2. MessageCount starts at 0
3. First user message triggers `agent.Run()`
4. At line 540: checks `currentSession.MessageCount == 0`
5. If true, calls `generateTitle()`

This should work... unless:
- MessageCount is not being set correctly to 0
- The title check is missing entirely

## Proposed Fix
Based on TODO.md fix options:
1. Check `MessageCount == 0 OR title == "New Session"`
2. Add `needs_title` boolean flag to session schema
3. Always regenerate title if it matches default patterns

Option 1 is simplest and matches the TODO.md suggestion.

## Implementation
Update line 540 in `/home/nexora/internal/agent/agent.go` from:
```go
if currentSession.MessageCount == 0 {
```

To:
```go
if currentSession.MessageCount == 0 || currentSession.Title == "New Session" {
```

This ensures that even if MessageCount is somehow not 0, we still regenerate titles for placeholder-named sessions.

## Test Plan
1. Create session with default title "New Session"
2. Send first message
3. Verify title gets regenerated to something meaningful
4. Check session in database has updated title
