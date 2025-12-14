# Installation Fix

## Issue
`go install .` was not installing the latest version (0.26.0) but kept showing an older version (0.25.1).

## Root Cause
1. The Go module path `github.com/nexora/cli` doesn't match the actual repository location `github.com/jeffersonwarrior/nexora`
2. The version detection logic in `internal/version/version.go` wasn't handling the `(devel)` case properly
3. There was no standardized way to build with the correct version stamp

## Solution
1. Fixed the version detection logic to properly handle development builds
2. Added an installation script that properly sets the version via ldflags
3. Added a replace directive in go.mod for local development
4. Fixed the Taskfile.yaml to use the correct module path for ldflags

## Usage
For proper installation with version stamping:

```bash
# From the repository root
./install.sh

# Or with a custom version
./install.sh v0.26.1

# Or manually
go install -ldflags="-X github.com/nexora/cli/internal/version.Version=0.26.0" .
```

This ensures the version is correctly embedded in the binary during installation.