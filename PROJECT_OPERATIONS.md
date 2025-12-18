# Nexora Project Operations Guide

This document provides operational guidance for working with the Nexora codebase - how to perform day-to-day development tasks, version management, quality assurance, and project maintenance.

## 🔄 Version Management

### Version Numbering
- Semantic versioning: `MAJOR.MINOR.PATCH` (e.g., `0.22.3`)
- Version stored in: `/internal/version/version.go`
- Update before release: Edit `var Version = "x.x.x"` directly

### Release Process
1. **Update version** in `internal/version/version.go`
2. **Run quality checks** with `./scripts/lint.sh`
3. **Build and install**: `go build . && go install .`
4. **Commit changes** with semantic commit message
5. **Tag release** (if applicable): `git tag v0.22.3`

### Development Versions
- Append `+dev` to version for development builds
- Example: `v0.22.3+dev` indicates unreleased changes

## 🏗️ Build Operations

### Local Development
```bash
# Quick build
go build .

# Build with verbose output
go build -v .

# Run directly
go run .

# Install to PATH
go install .
```

### Build Verification
```bash
# Build check
go build .

# Verify installation
nexora --version
```

## 🧪 Testing Operations

### Test Suite Management
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/agent -v

# Run single test
go test ./internal/agent -run TestCoderAgent

# Run tests with coverage
go test -cover ./...

# Run tests with coverage to file
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Golden File Management
```bash
# Update all golden files
go test ./... -update

# Update specific package golden files
go test ./internal/tui/components/core -update

# Handle test failures due to prompt changes
# Golden files need updating when system prompt changes
```

### Test Types
- **Unit Tests**: Package-level functionality
- **Integration Tests**: Cross-package functionality (`internal/indexer/p6_simple_test.go`)
- **Golden File Tests**: Expected output validation
- **Benchmark Tests**: Performance measurement
- **Mock Provider Tests**: Avoid external API calls

## 🔍 Code Quality Operations

### Automated Linting
```bash
# Run comprehensive quality checks
./scripts/lint.sh

# Individual quality checks
gofumpt -w .                    # Format code
go vet ./...                     # Static analysis
go mod tidy                      # Module cleanup
golangci-lint run               # Advanced linting (if available)
```

### Quality Categories
- **Code Formatting**: gofumpt (stricter than gofmt)
- **Static Analysis**: go vet, golangci-lint
- **Module Integrity**: go mod verify, go mod tidy
- **Build Verification**: Compilability check
- **Test Coverage**: Ensure all critical paths tested

### Common Quality Issues
- `interface{}` → use `any` instead
- `context.TODO()` → use proper context
- Missing SQLite imports in test files
- Unused variables and functions
- Inconsistent error handling

## 📝 Git Operations

### Commit Workflow
```bash
# Check current status
git status

# View changes
git diff

# Check recent commits
git log -n 3 --oneline

# Stage changes (selective)
git add internal/indexer/

# Stage all relevant changes
git add go.mod internal/agent/templates/

# Commit with semantic message
git commit -m "feat: add Mistral AI integration"
```

### Semantic Commit Messages
- `feat:` New features
- `fix:` Bug fixes  
- `docs:` Documentation changes
- `refactor:` Code refactoring
- `test:` Test additions/changes
- `chore:` Maintenance tasks
- `sec:` Security fixes

### Branch Management
```bash
# Create feature branch
git checkout -b feature/mistral-integration

# Switch branches
git checkout main

# Merge changes
git checkout main
git merge feature/mistral-integration

# Delete merged branch
git branch -d feature/mistral-integration
```

## 🔧 Development Workflow Operations

### Daily Development Checklist
1. **Sync**: `git pull origin main`
2. **Build**: `go build .`
3. **Lint**: `./scripts/lint.sh`
4. **Test**: `go test ./...`
5. **Make changes**
6. **Build**: `go build .` (verify)
7. **Lint**: `./scripts/lint.sh` (verify)
8. **Test**: `go test ./...` (verify)
9. **Commit**: with semantic message

### Pre-Submit Checklist
- [ ] Code builds without errors
- [ ] All tests pass
- [ ] Lint script completes successfully
- [ ] No TODO/FIXME comments in new code
- [ ] Proper error handling implemented
- [ ] Documentation updated as needed
- [ ] Golden files updated if required

### Working with Subsystems
- **Indexer**: Use `./scripts/lint.sh` to catch SQLite import issues
- **Agent**: Update golden files when system prompt changes
- **TUI**: Run UI-specific tests in `./internal/tui/`
- **Tools**: Test individual tools with `./internal/agent/tools/`

## 🛠️ Script Operations

### Available Scripts
```bash
# Auto-lint and quality check
./scripts/lint.sh

# GitHub PR labeling
./scripts/run-labeler.sh
```

### Custom Scripts
Create new scripts in `/scripts/` directory:
- Ensure executable: `chmod +x scripts/your-script.sh`
- Follow existing naming conventions
- Include proper error handling
- Use semantic output

## 📊 Performance Operations

### Benchmark Testing
```bash
# Run benchmarks
go test -bench=. ./internal/indexer/

# Run specific benchmark
go test -bench=BenchmarkIndexing ./internal/indexer/

# Run with memory profiling
go test -bench=. -memprofile=mem.prof ./internal/indexer/
```

### Performance Monitoring
- Use built-in performance tests in `internal/indexer/p6_simple_test.go`
- Monitor indexing speed: symbols/second
- Track query response times
- Profile memory usage patterns

## 🔄 Continuous Integration Operations

### CI/CD Pipeline
- **GitHub Actions**: `.github/workflows/`
- **Build Verification**: Automated on each push
- **Testing**: Full test suite execution
- **Quality Assurance**: Lint and static analysis
- **Release**: Automated goreleaser deployment

### Local CI Simulation
```bash
# Simulate CI checks
./scripts/lint.sh
go test ./...
go build .
```

## 🐛 Troubleshooting Operations

### Common Issues & Solutions

#### Build Failures
```bash
# Clean build
go clean -cache
go build .

# Module issues
go mod download
go mod tidy
```

#### Test Failures
```bash
# Update golden files (if legit change)
go test ./... -update

# Run specific failing test
go test ./internal/agent -v -run TestSpecificTest

# Check with mock providers
# Enable in code: config.UseMockProviders = true
```

#### SQLite Import Issues
```bash
# Check import exists
grep -r "_ \"github.com/ncruces/go-sqlite3\"" internal/indexer/

# Add to test files if missing
echo '_ "github.com/ncruces/go-sqlite3"' >> internal/indexer/test_file.go
```

#### Golden File Mismatches
```bash
# Regenerate after prompt changes
go test ./internal/agent -update

# View diff to understand changes
git diff internal/agent/testdata/
```

## 📋 Maintenance Operations

### Regular Tasks
- **Daily**: Update git status, run lint script
- **Weekly**: Check for dependency updates, clean temp files
- **Release**: Update version, comprehensive testing, tagging

### Cleanup Operations
```bash
# Remove temporary files
rm -f *.db *.prof coverage.out

# Clean build cache
go clean -testcache

# Update dependencies
go get -u ./...
go mod tidy
```

### Documentation Updates
- Update `NEXORA.md` for system changes
- Update this document for operational changes
- Update README.md for user-facing changes
- Update TODO.md for new issues identified

## 🚀 Advanced Operations

### Profiling
```bash
# CPU profiling
go test -cpuprofile=cpu.prof ./...

# Memory profiling
go test -memprofile=mem.prof ./...

# Visualize profiles
go tool pprof cpu.prof
go tool pprof mem.prof
```

### Debugging
```bash
# Debug tests
go test -debug ./...

# Race condition detection
go test -race ./...

# Build with debug symbols
go build -gcflags="all=-N -l"
```

### Cross-Platform Building
```bash
# Build for multiple platforms
GOOS=linux GOARCH=amd64 go build .
GOOS=darwin GOARCH=arm64 go build .
GOOS=windows GOARCH=amd64 go build .
```

---

## 📞 Support & Resources

### Quick Reference Commands
```bash
# Full quality check
./scripts/lint.sh

# Build and test
go build . && go test ./...

# Version check
nexora --version

# Git status check
git status && git log -n 1
```

### Key Documentation
- **System Architecture**: `NEXORA.md`
- **API Documentation**: `CODEDOCS.md` 
- **Development Standards**: This document
- **Open Issues**: `README.md` (if any)

### Getting Help
- Check existing issues in repository
- Review this guide for operational procedures
- Consult `NEXORA.md` for system understanding
- Use `git help <command>` for Git operations

---

*This document focuses on the "how" of project operations, complementing the system documentation in `NEXORA.md` which covers the "what" and "why" of the Nexora architecture.*