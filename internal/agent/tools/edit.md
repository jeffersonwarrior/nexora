Edits files by replacing text, creating new files, or deleting content. For moving/renaming use Bash 'mv'. For large edits use Write tool.

**🚨 MANDATORY: Read This Before Using Edit Tool**

**Quick Decision Tree:**
1. **Know exact line numbers?** 	 Use `smart_edit` tool (100% reliable, zero whitespace issues)
2. **Edit failed once?** 	 Use `write` tool immediately (don't retry with "fixed" whitespace)
3. **First time editing?** 	 Continue below, but enable ai_mode=true

**Quick Troubleshooting**:
- Edit failed? DON'T retry - use `write` or `smart_edit` instead
- Need exact text? Use `sed -n 'START,ENDp' file` (NOT view tool output)
- Multi-line edits? Use `smart_edit` with line numbers for guaranteed success

<prerequisites>
1. **NEVER copy from view tool output** - it has visual tab indicators (	) that cause failures
2. Use `sed -n 'START,ENDp' file` or `grep -A5 -B5 pattern file` to extract EXACT text
3. For new files: Use LS tool to verify parent directory exists
4. **CRITICAL**: If you can't extract exact text, use `smart_edit` with line numbers instead
</prerequisites>

<parameters>
1. file_path: Absolute path to file (required)
2. old_string: Text to replace (must match exactly including whitespace/indentation)
3. new_string: Replacement text
4. replace_all: Replace all occurrences (default false)
</parameters>

<special_cases>

- Create file: provide file_path + new_string, leave old_string empty
- Delete content: provide file_path + old_string, leave new_string empty
  </special_cases>

<critical_requirements>
EXACT MATCHING: The tool is extremely literal. Text must match **EXACTLY**

- Every space and tab character
- Every blank line
- Every newline character
- Indentation level (count the spaces/tabs)
- Comment spacing (`// comment` vs `//comment`)
- Brace positioning (`func() {` vs `func(){`)

Common failures:

```
Expected: "    func foo() {"     (4 spaces)
Provided: "  func foo() {"       (2 spaces) ❌ FAILS

Expected: "}\n\nfunc bar() {"    (2 newlines)
Provided: "}\nfunc bar() {"      (1 newline) ❌ FAILS

Expected: "// Comment"           (space after //)
Provided: "//Comment"            (no space) ❌ FAILS
```

UNIQUENESS (when replace_all=false): old_string MUST uniquely identify target instance

- Include 3-5 lines context BEFORE and AFTER change point
- Include exact whitespace, indentation, surrounding code
- If text appears multiple times, add more context to make it unique

SINGLE INSTANCE: Tool changes ONE instance when replace_all=false

- For multiple instances: set replace_all=true OR make separate calls with unique context
- Plan calls carefully to avoid conflicts

VERIFICATION BEFORE USING: Before every edit

1. View the file and locate exact target location
2. Check how many instances of target text exist
3. Copy the EXACT text including all whitespace
4. Verify you have enough context for unique identification
5. Double-check indentation matches (count spaces/tabs)
6. Plan separate calls or use replace_all for multiple changes
   </critical_requirements>

<warnings>
Tool fails if:
- old_string matches multiple locations and replace_all=false
- old_string doesn't match exactly (including whitespace)
- Insufficient context causes wrong instance change
- Indentation is off by even one space
- Missing or extra blank lines
- Wrong tabs vs spaces
</warnings>

<recovery_steps>
If you get "old_string not found in file":

**🛑 STOP - DON'T RETRY THE SAME APPROACH**

1. **First time?** Use `smart_edit` with line numbers (100% reliable):
   ```bash
   grep -n "pattern" file  # Find line numbers
   # Then use smart_edit tool with start_line/end_line
   ```

2. **Need to use edit tool?** Extract EXACT text with sed:
   ```bash
   sed -n '10,20p' file  # Lines 10-20
   # Use this output directly as old_string
   ```

3. **Still failing?** Use `write` tool for full file replacement:
   ```bash
   view file  # Get full content
   # Make changes manually
   # Use write tool with complete new content
   ```

**❌ DON'T DO THIS:**
- Retry edit with "adjusted" whitespace - it will fail again
- Copy from view tool output - it has display artifacts (	)
- Guess at the whitespace - you'll never get it right

**✅ DO THIS:**
- Use `smart_edit` (line numbers, zero failures)
- Use `sed -n 'X,Yp' file` to extract exact text
- Use `write` tool if edit fails once
   </recovery_steps>

<best_practices>

- Ensure edits result in correct, idiomatic code
- Don't leave code in broken state
- Use absolute file paths (starting with /)
- Use forward slashes (/) for cross-platform compatibility
- Multiple edits to same file: send all in single message with multiple tool calls
- **When in doubt, include MORE context rather than less**
- Match the existing code style exactly (spaces, tabs, blank lines)
  </best_practices>

<whitespace_checklist>
Before submitting an edit, verify:

- [ ] Viewed the file first
- [ ] Counted indentation spaces/tabs
- [ ] Included blank lines if they exist
- [ ] Matched brace/bracket positioning
- [ ] Included 3-5 lines of surrounding context
- [ ] Verified text appears exactly once (or using replace_all)
- [ ] Copied text character-for-character, not approximated
      </whitespace_checklist>

<examples>
✅ Correct: Exact match with context

```
old_string: "func ProcessData(input string) error {\n    if input == \"\" {\n        return errors.New(\"empty input\")\n    }\n    return nil\n}"

new_string: "func ProcessData(input string) error {\n    if input == \"\" {\n        return errors.New(\"empty input\")\n    }\n    // New validation\n    if len(input) > 1000 {\n        return errors.New(\"input too long\")\n    }\n    return nil\n}"
```

❌ Incorrect: Not enough context

```
old_string: "return nil"  // Appears many times!
```

❌ Incorrect: Wrong indentation

```
old_string: "  if input == \"\" {"  // 2 spaces
// But file actually has:        "    if input == \"\" {"  // 4 spaces
```

✅ Correct: Including context to make unique

```
old_string: "func ProcessData(input string) error {\n    if input == \"\" {\n        return errors.New(\"empty input\")\n    }\n    return nil"
```

</examples>

<windows_notes>

- Forward slashes work throughout (C:/path/file)
- File permissions handled automatically
- Line endings converted automatically (\n ↔ \r\n)
  </windows_notes>

<common_whitespace_issues>

**Line Ending Problems**:
- CRLF vs LF differences that aren't visible in View output
- Solution: Use `cat -A file` to see $ (LF) vs ^M$ (CRLF)

**Tab vs Space Confusion**:
- Tabs (^I) vs spaces that look identical
- Solution: Use `cat -A file` to distinguish ^I (tabs) from spaces

**Hidden Unicode Characters**:
- Non-breaking spaces (U+00A0), zero-width spaces (U+200B)
- Solution: Use `hexdump -C file` to see exact byte values

**Trailing Whitespace**:
- Invisible trailing spaces that cause exact match failures
- Solution: Use `grep -n "pattern[[:space:]]*$"` to detect

**Mixed Indentation**:
- Some lines use tabs, others use spaces in same block
- Solution: Normalize with `expand` or `unexpand` commands

**When All Else Fails**:
- Use `write` tool for complete file replacement
- Or use `multiedit` with smaller, more isolated changes
  </common_whitespace_issues>

<practical_troubleshooting_guide>

**Step-by-Step Debugging Process**:

1. **Confirm exact line content**:
   ```bash
   grep -n "pattern" file
   ```

2. **Reveal invisible characters**:
   ```bash
   cat -A file | grep -A5 -B5 "pattern"
   ```

3. **Check for Unicode issues**:
   ```bash
   hexdump -C file | grep -A2 -B2 "pattern"
   ```

4. **Test with smaller context**:
   ```bash
   # Instead of large block, try smaller unique parts
   edit file_path="file" old_string="unique_line_before\ntarget_line\nunique_line_after" new_string="replacement"
   ```

5. **Fallback to write tool**:
   ```bash
   # When exact matching fails repeatedly
   view file_path="file"
   # Copy entire content, make changes, then:
   write file_path="file" content="complete_new_content"
   ```

**Pro Tip**: For complex files, use `write` tool from the start - it's often faster than debugging exact match issues.
  </practical_troubleshooting_guide>

## AI Mode

For AI agents, use `ai_mode=true` to enable automatic fixes for common issues:

```json
{
  "file_path": "/path/to/file.go",
  "old_string": "func main() {",
  "new_string": "func main() {\n    // new code",
  "ai_mode": true
}
```

### AI Mode Features

1. **Automatic Tab Normalization**: Converts VIEW tool display tabs (`→\t`) to real tabs (`\t`)
2. **Context Expansion**: Automatically expands minimal context to improve match success
3. **Enhanced Error Messages**: Provides actionable guidance for common failure patterns
4. **Lower AIOPS Threshold**: More aggressive use of AI-powered edit resolution

### When to Use AI Mode

- When copying text from VIEW tool output
- When getting whitespace-related errors
- When working with files that have complex indentation
- When initial edit attempts fail

### AI Mode Best Practices

- Still provide as much context as possible
- Check error messages for specific guidance
- Use for complex edits where exact matching is difficult
- Combine with other parameters as needed
