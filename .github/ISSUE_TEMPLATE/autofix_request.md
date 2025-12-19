---
name: 🤖 Auto-fix Request
about: Request automatic fixes for common issues
title: '[AUTO-FIX] '
labels: [autofix]
assignees: []
---

## 🤖 Auto-fix Request

Use this template to request automatic fixes for common issues. The bot will analyze and attempt to fix the issue automatically.

### What needs to be auto-fixed?

- [ ] **Dependencies** - Update go.mod, fix version conflicts
- [ ] **Formatting** - Fix code formatting with gofmt
- [ ] **Linting** - Fix golangci-lint issues
- [ ] **Imports** - Fix import issues and unused imports
- [ ] **Security** - Fix vulnerable dependencies and security issues
- [ ] **Performance** - Apply performance optimizations
- [ ] **Build Issues** - Fix compilation and build errors

### Issue Description

Describe the issue you're experiencing:

``[e.g: Getting build errors after dependency updates, code formatting issues, security vulnerabilities detected, performance bottlenecks, etc.]``

### Error Logs (if applicable)

```bash
# Paste error logs here
```

### Commands that trigger the autofix

You can also trigger auto-fixes by commenting on any issue with:

- `/fix-deps` - Fix dependency issues
- `/fix-format` - Fix formatting issues  
- `/fix-lint` - Fix linting issues
- `/fix-imports` - Fix import issues
- `/fix-security` - Fix security vulnerabilities
- `/fix-performance` - Apply performance optimizations
- `/fix-build` - Fix build compilation issues

### Additional Context

Add any other context about the problem here.

---

🤖 **Note**: The autofix bot will analyze the issue and create pull requests with suggested fixes. All changes will be reviewed before merging.