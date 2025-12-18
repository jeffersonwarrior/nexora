# Find Tool Documentation

Find is an enhanced file search tool that uses modern command-line utilities when available.

## Usage

### Basic Search
```bash
nexora run "find all go files"
```

### Advanced Search
```bash
nexora run "find files containing 'main' in src/"
```

### Pattern Matching
- Accepts glob patterns: `*.go`, `test_*.rb`
- Recursive directory searches
- File content searches via ripgrep when available

## Permissions

The find tool respects system permissions and will only search accessible directories.

## Implementation Notes

- Uses `fd` for fast file finding when available
- Falls back to `find` command for compatibility
- Integrates with ripgrep for content searches
- Supports both filename and content-based searches