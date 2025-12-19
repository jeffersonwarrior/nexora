# AI-Enhanced Auto-Fix System for Nexora

## Overview

The AI-Enhanced Auto-Fix system automatically detects and fixes CI failures in the Nexora repository. When lint, build, or test workflows fail, the system will:

1. **Automatically trigger** when CI workflows fail
2. **Analyze the failure** to identify the root cause
3. **Apply comprehensive fixes** using the autofix script
4. **Create a pull request** with the fixes
5. **Notify the team** about the actions taken

## How It Works

### Automatic Triggers

The system automatically triggers when:
- `lint.yml` workflow completes with `failure` status
- `build.yml` workflow completes with `failure` status
- Any PR/issue is commented with `/ai-fix`
- Manually triggered from the Actions tab

### Fix Types Applied

The system runs a comprehensive suite of fixes:

1. **Dependencies**: Updates go.mod, fixes version conflicts
2. **Formatting**: Fixes code formatting with `gofmt`
3. **Imports**: Organizes imports with `goimports`
4. **Linting**: Auto-fixes golangci-lint issues
5. **Build**: Fixes compilation errors
6. **Tests**: Addresses test failures

## Manual Triggers

### Option 1: Comment Command
On any issue or PR, comment:
```
/ai-fix
```

### Option 2: Create an Issue
Use the "🤖 Auto-fix Request" template and select the issues to fix.

### Option 3: Manual Dispatch
1. Go to Actions tab in GitHub
2. Select "AI-Enhanced Auto-Fix" workflow
3. Click "Run workflow"

### Option 4: Run Locally
```bash
# Clone and run the autofix script locally
./scripts/autofix.sh all

# Or run specific types
./scripts/autofix.sh linting
./scripts/autofix.sh dependencies
```

## What Gets Fixed

### Automatically Fixed
- ✅ Go dependency issues (go.mod/go.sum)
- ✅ Code formatting (gofmt)
- ✅ Import organization (goimports)
- ✅ Linting issues (golangci-lint)
- ✅ Simple build errors
- ✅ Common test failures

### Requires Manual Review
- ❌ Complex logic errors
- ❌ Architecture changes
- ❌ Security vulnerabilities
- ❌ Performance optimizations
- ❌ New feature implementations

## Workflow Details

### 1. Failure Detection
```yaml
on:
  workflow_run:
    workflows: ["lint", "build"]
    types: [completed]
```

### 2. Analysis Phase
- Extracts failure details from the workflow
- Identifies which jobs/steps failed
- Gathers error messages and context

### 3. Auto-Fix Execution
```bash
./scripts/autofix.sh all
```
This runs all fix types in sequence:
1. Fix dependencies (go mod tidy)
2. Fix formatting (gofmt -s -w .)
3. Fix imports (goimports)
4. Fix linting (golangci-lint --fix)
5. Fix build issues
6. Fix test issues

### 4. PR Creation
If fixes are applied:
- Creates branch: `ai-autofix/<workflow>-<timestamp>`
- Commits with detailed message
- Creates PR with comprehensive description
- Adds labels: `ai-autofix, ci-failure`
- Assigns to the triggering user

### 5. Notification
- Comments on the original issue/PR
- Links to the created fix PR
- Provides summary of fixes applied

## Example Usage

### Scenario: Lint Failure
1. PR is pushed, lint workflow fails
2. AI-Enhanced Auto-Fix automatically triggers
3. Runs all fix types, finds formatting issues
4. Creates PR with fixes:
   - Title: "🤖 AI Auto-Fix: lint Failures"
   - Body: Detailed breakdown of fixes
   - Labels: ai-autofix, ci-failure
5. Comments on original PR with fix PR link

### Scenario: Manual Trigger
```bash
# On GitHub issue or PR
/ai-fix

# Response:
🤖 I've automatically created a fix PR for the CI failure: https://github.com/repo/pr/123
```

## Configuration

### Required Secrets
No additional secrets required - uses standard GITHUB_TOKEN

### Customization
To modify behavior, edit:
- `.github/workflows/ai-enhanced-autofix.yml` - Main workflow
- `scripts/autofix.sh` - Fix logic
- `.github/workflows/autofix.yml` - Original autofix

### Troubleshooting
If the AI fix system fails:
1. Check the workflow logs in Actions tab
2. Look for "AI-Enhanced Auto-Fix" workflow runs
3. Review error messages in the logs
4. Manual fixes may be required for complex issues

## Best Practices

1. **Review Auto-Fixes**: Always review the auto-generated PR before merging
2. **Monitor CI**: Keep an eye on CI results after auto-fixes are applied
3. **Test Locally**: For complex issues, test fixes locally first
4. **Custom Fix Logic**: Add custom fix patterns to `scripts/autofix.sh` as needed

## Integration with Other Workflows

The AI-Enhanced Auto-Fix integrates with:
- `lint.yml` - Triggers on lint failures
- `build.yml` - Triggers on build failures
- `autofix.yml` - Provides the core fixing logic
- `autofix-helper.yml` - Supports manual triggers

This creates a comprehensive CI/CD pipeline that can self-heal common issues, reducing manual intervention and keeping the codebase clean.