Line-number based editing tool for 100% reliability - no whitespace matching required.

<why_use_this>
- **Zero whitespace issues**: No need to match tabs, spaces, or invisible characters
- **100% reliability**: Specify exact line numbers, never fails
- **Faster**: No fuzzy matching or retries needed
- **Clearer intent**: Shows exactly what lines you're modifying
</why_use_this>

<when_to_use>
Use smart_edit when:
- You know the exact line numbers to modify
- You want guaranteed success without whitespace headaches
- You're making surgical changes to specific lines
- The regular edit tool has failed multiple times

Use regular edit tool when:
- You need to match content patterns across files
- You don't know line numbers
- You're doing search-and-replace operations
</when_to_use>

<usage>
1. Use `grep -n "pattern" file` to find exact line numbers
2. Or use `view` tool and note the line numbers shown
3. Specify start_line and end_line (1-indexed, inclusive)
4. Provide new_string as replacement
</usage>

<parameters>
- file_path: Absolute path to file (required)
- start_line: First line to replace (1-indexed, required)
- end_line: Last line to replace (1-indexed, inclusive, required)
- new_string: Replacement content (can be multi-line)
</parameters>

<examples>
Example 1: Replace single line
```json
{
  "file_path": "/home/nexora/main.go",
  "start_line": 42,
  "end_line": 42,
  "new_string": "\tfmt.Println(\"Hello, World!\")"
}
```

Example 2: Replace multiple lines
```json
{
  "file_path": "/home/nexora/config.go",
  "start_line": 10,
  "end_line": 15,
  "new_string": "\t// New implementation\n\treturn &Config{\n\t\tDebug: true,\n\t}"
}
```

Example 3: Delete lines (empty new_string)
```json
{
  "file_path": "/home/nexora/test.go",
  "start_line": 20,
  "end_line": 25,
  "new_string": ""
}
```
</examples>

<workflow>
1. Find line numbers: `grep -n "target" file.go`
2. View context: `sed -n '10,20p' file.go`
3. Apply edit with exact line numbers
4. No whitespace matching = no failures
</workflow>

<tips>
- Line numbers are 1-indexed (first line is 1, not 0)
- end_line is inclusive (lines start_line through end_line are replaced)
- new_string can contain \n for multi-line replacements
- Empty new_string deletes the lines
- Always check line numbers with grep -n first
</tips>
