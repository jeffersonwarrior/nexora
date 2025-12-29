# Greptile GitHub Action - Fixed

## Problem
The Greptile GitHub Action was failing with `jq: parse error: Invalid numeric literal` because:
1. **Wrong API endpoint** - Used `/v2/reviews` which doesn't exist
2. **Wrong request format** - API expects `repositories` array with objects, not simple strings
3. **Poor error handling** - No validation of API responses before parsing with `jq`

## Solution

### Correct API Format
```json
{
  "repositories": [{
    "remote": "github",
    "repository": "owner/repo",
    "branch": "branch-name"
  }],
  "query": "Your review instructions here",
  "sessionId": "optional-session-id"
}
```

### Key Changes Made

1. **Fixed API Request Structure**
   - Changed from `"repository": "string"` to `"repositories": [{"remote": "github", "repository": "string", "branch": "string"}]`
   - Proper object format in repositories array

2. **Enhanced Error Handling**
   - Validate JSON before parsing with `jq`
   - Create safe default response file
   - Capture HTTP status codes
   - Handle empty/invalid responses gracefully

3. **Better Response Processing**
   - Check if file exists and has content
   - Validate JSON before extracting fields
   - Support multiple response formats (`.message`, `.answer`, `.error`)
   - Show raw response if JSON is invalid

## Testing

Try the workflow now - it will handle API responses gracefully even if Greptile returns unexpected data!

```bash
# Create a test PR
git checkout -b test-greptile
echo "# Test" >> README.md
git add README.md
git commit -m "Test Greptile integration"
git push origin test-greptile
gh pr create --title "Test Greptile" --body "Testing fixed Greptile workflow"
```