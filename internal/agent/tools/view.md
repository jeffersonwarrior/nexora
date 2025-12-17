Reads and displays file contents with line numbers for examining code, logs, or text data.

<usage>
- Provide file path to read
- Optional offset: start reading from specific line (0-based)
- Optional limit: control lines read (default 100)
- Don't use for directories (use LS tool instead)
- Supports image files (PNG, JPEG, GIF, BMP, SVG, WebP)
</usage>

<features>
- Displays contents with line numbers
- Can read from any file position using offset
- Handles large files by limiting lines read
- Auto-truncates very long lines for display
- Suggests similar filenames when file not found
- Renders image files directly in terminal
</features>

<limitations>
- Max file size: 5MB
- Default limit: 100 lines (reduced to prevent context window issues)
- Lines >2000 chars truncated
- Binary files (except images) cannot be displayed
</limitations>

<cross_platform>
- Handles Windows (CRLF) and Unix (LF) line endings
- Works with forward slashes (/) and backslashes (\)
- Auto-detects text encoding for common formats
</cross_platform>

<tips>
- Use with Glob to find files first
- For code exploration: Grep to find relevant files, then View to examine
- For large files: use offset and limit parameters for specific sections
- The tool uses smaller chunks (100 lines default) to manage context window efficiently
- For very large files, consider using 'head' or 'tail' commands via bash tool
- View tool automatically detects and renders image files
</tips>
