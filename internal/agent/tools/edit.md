Edits files by replacing text, creating new files, or deleting content. For moving/renaming use Bash 'mv'. For large edits use Write tool.

**Quick Decision Tree:**
1. **Know exact line numbers?** → Use `smart_edit` tool (100% reliable)
2. **Edit failed once?** → Use `write` tool immediately
3. **First time editing?** → Enable ai_mode=true

<prerequisites>
1. **NEVER copy from view tool output** - has visual tab indicators (→	) causing failures
2. Use `sed -n 'START,ENDp' file` or `grep -A5 -B5 pattern file` for EXACT text
3. For new files: Use LS tool to verify parent directory exists
4. If you can't extract exact text, use `smart_edit` with line numbers
</prerequisites>

<parameters>
1. file_path: Absolute path to file (required)
2. old_string: Text to replace (exact match including whitespace/indentation)
3. new_string: Replacement text
4. replace_all: Replace all occurrences (default false)
</parameters>

<special_cases>
- Create file: provide file_path + new_string, leave old_string empty
- Delete content: provide file_path + old_string, leave new_string empty
</special_cases>

<critical_requirements>
**EXACT MATCHING**: Text must match character-for-character
- Every space, tab, blank line, newline, indentation level
- Comment spacing (`// comment` vs `//comment`)
- Brace positioning (`func() {` vs `func(){`)

**UNIQUENESS**: old_string MUST uniquely identify target (when replace_all=false)
- Include 3-5 lines context before/after change point
- If text appears multiple times, add more context or use replace_all=true

**VERIFICATION BEFORE USING**:
1. View file and locate exact target
2. Check how many instances exist
3. Copy EXACT text including whitespace
4. Verify enough context for unique identification
5. Double-check indentation (count spaces/tabs)
</critical_requirements>

<warnings>
Tool fails if:
- old_string matches multiple locations and replace_all=false
- Whitespace doesn't match exactly
- Indentation off by even one space
- Missing or extra blank lines
</warnings>

<recovery_steps>
If "old_string not found":

**🛑 DON'T RETRY - Use alternatives:**
1. `smart_edit` with line numbers (100% reliable)
2. Extract exact text: `sed -n '10,20p' file`
3. Use `write` tool for full file replacement

**❌ DON'T:**
- Retry with "adjusted" whitespace
- Copy from view output (has artifacts)
- Guess at whitespace

**✅ DO:**
- Use `smart_edit` (line numbers)
- Use `sed -n 'X,Yp' file` for exact text
- Use `write` if edit fails once
</recovery_steps>

<best_practices>
- View file first, match exact formatting
- Use absolute paths with forward slashes (/)
- Multiple edits to same file: send all in single message
- Include MORE context when uncertain
- Match existing code style exactly
</best_practices>

<examples>
✅ **Correct**: Exact match with context
```
old_string: "func ProcessData(input string) error {\n    if input == \"\" {\n        return errors.New(\"empty input\")\n    }\n    return nil\n}"
new_string: "func ProcessData(input string) error {\n    if input == \"\" {\n        return errors.New(\"empty input\")\n    }\n    if len(input) > 1000 {\n        return errors.New(\"input too long\")\n    }\n    return nil\n}"
```

❌ **Incorrect**: Not enough context
```
old_string: "return nil"  // Appears many times!
```

❌ **Incorrect**: Wrong indentation
```
old_string: "  if input == \"\" {"  // 2 spaces
// File actually has: "    if input == \"\" {"  // 4 spaces
```
</examples>

## AI Mode
Use `ai_mode=true` (default) for automatic fixes:
- Converts VIEW tool display tabs (`→\t`) to real tabs (`\t`)
- Expands minimal context automatically
- Enhanced error messages
- Lower failure rate on whitespace issues
