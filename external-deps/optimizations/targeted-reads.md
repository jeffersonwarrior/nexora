# Targeted File Reads

Minimize token usage by reading only what's needed.

## Strategy

1. **Locate first**: Use Grep to find specific lines/functions
2. **Surgical read**: Use `Read file offset=X limit=Y` for files >200 lines
3. **Expand if needed**: If context insufficient, read more

## When to Use Full Reads

- Files under 200 lines
- Need to understand entire module architecture
- Targeted read proved insufficient
- File structure is unknown

## Offset/Limit Patterns

```
# Read function definition + context
offset=<line-10> limit=50

# Read class/struct definition
offset=<line-5> limit=100

# Read specific section
offset=<start> limit=<end-start+10>
```

## File Size Thresholds

| Size | Strategy |
|------|----------|
| <200 lines | Full read acceptable |
| 200-500 lines | Use offset/limit when target known |
| 500+ lines | Always use targeted reads |

## Anti-patterns

- Reading entire large files "just in case"
- Multiple full reads of same file in session
- Reading files already in context
