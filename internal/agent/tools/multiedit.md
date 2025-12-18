Makes multiple edits to a single file in one operation. Built on Edit tool for efficient multiple find-and-replace. Prefer over Edit tool for multiple edits to same file.

<prerequisites>
1. Use View tool to understand file contents
2. Verify directory path is correct
3. CRITICAL: Note exact whitespace, indentation from View output
</prerequisites>

<parameters>
1. file_path: Absolute path to file (required)
2. edits: Array of edit objects (MUST be array, not string!):
   - old_string: Text to replace (exact match including whitespace)
   - new_string: Replacement text
   - replace_all: Replace all occurrences (optional, default false)
   
   ✅ CORRECT: `"edits": [{"old_string": "text1", "new_string": "replace1"}]`
   ❌ INCORRECT: `"edits": "[{\"old_string\": \"text1\"}]"` (string, not array)
</parameters>

<operation>
- Edits applied sequentially in order
- Each edit operates on result of previous edit
- PARTIAL SUCCESS: Failed edits returned, successful edits kept
- File modified if at least one edit succeeds
</operation>

<inherited_rules>
**All Edit tool rules apply to each edit:**
- See edit.md for critical requirements, warnings, recovery steps
- Exact matching for whitespace, indentation, blank lines
- Uniqueness requirements (3-5 lines context)
- Verification checklist before using
- Use ai_mode=true (default) for tab handling
</inherited_rules>

<critical_requirements>
1. Apply Edit tool rules to EACH edit (see edit.md)
2. Edits applied in order; successful edits kept if later edits fail
3. Plan sequence: earlier edits change content that later edits must match
4. Ensure each old_string unique at application time (after prior edits)
5. Check response for failed edits, retry with corrections
</critical_requirements>

<verification_before_using>
1. View file, copy exact text (including whitespace) for each target
2. Check instances of each old_string BEFORE sequence
3. Dry-run: after edit #N, will edit #N+1 still match?
4. Prefer fewer, larger context blocks over many tiny fragments
5. If edits independent, consider separate multiedit batches
</verification_before_using>

<warnings>
- Check response for failed edits
- Earlier edits invalidate later matches (added/removed spaces, lines)
- Mixed tabs/spaces, trailing spaces cause failures
- replace_all may affect unintended regions
</warnings>

<recovery_steps>
If edits fail:
1. Check response metadata for failed edits with errors
2. View file to see current state after successful edits
3. Adjust failed edits based on new content
4. Retry with corrected old_string values
5. Break complex batches into smaller operations
</recovery_steps>

<best_practices>
- Result in correct, idiomatic code
- Use absolute file paths (/)
- Use replace_all only when certain
- Match existing style exactly
- Review failed edits, retry with corrections
</best_practices>

<examples>
✅ **Correct**: Sequential edits accounting for changes
```
edits: [
  {
    old_string: "func A() {\n    doOld()\n}",
    new_string: "func A() {\n    doNew()\n}",
  },
  {
    // Context exists AFTER first replacement
    old_string: "func B() {\n    callA()\n}",
    new_string: "func B() {\n    callA()\n    logChange()\n}",
  },
]
```

❌ **Incorrect**: Second old_string no longer matches
```
edits: [
  {
    old_string: "func A() {\n    doOld()\n}",
    new_string: "func A() {\n\n    doNew()\n}", // Added blank line
  },
  {
    old_string: "func A() {\n    doNew()\n}", // Missing blank line - FAILS
    new_string: "func A() {\n    doNew()\n    logChange()\n}",
  },
]
```
</examples>
