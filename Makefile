# Makefile for Nexora

LDFLAGS := -X github.com/nexora/nexora/internal/version.Version=$(VERSION)
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	go build -ldflags="$(LDFLAGS)" -o nexora .

# Install to system (requires sudo or write permissions)
.PHONY: install
install: build
	@if [ -w "/usr/local/bin" ]; then \
		echo "Installing to /usr/local/bin/nexora..."; \
		mv nexora /usr/local/bin/nexora; \
		chmod +x /usr/local/bin/nexora; \
		echo "Nexora installed successfully!"; \
	else \
		echo "Need sudo privileges to install to /usr/local/bin"; \
		echo "Installing with sudo..."; \
		sudo mv nexora /usr/local/bin/nexora; \
		sudo chmod +x /usr/local/bin/nexora; \
		echo "Nexora installed successfully!"; \
	fi

# Install to user directory (no sudo required)
.PHONY: install-user
install-user: build
	@mkdir -p ~/.local/bin
	mv nexora ~/.local/bin/nexora
	chmod +x ~/.local/bin/nexora
	@echo "Nexora installed to ~/.local/bin/nexora"
	@echo "Make sure ~/.local/bin is in your PATH"

# Clean build artifacts
.PHONY: clean
clean:
	rm -f nexora
	rm -f /tmp/nexora

# Run tests with timeout protection
.PHONY: test-qa
test-qa:
	go test ./qa/... -timeout=5m

# Build with tests
.PHONY: build-safe
build-safe: test-qa
	go build ./..

# Full clean build
.PHONY: build-full
build-full: clean test-qa
	go build -ldflags="$(LDFLAGS)" -o nexora .

.PHONY: test
test: test-qa
	go test ./... -timeout=10m -coverprofile=coverage.out

# Quick unit tests (no integration)
.PHONY: test-quick
test-quick:
	go test ./internal/... -short -timeout=5m -coverprofile=coverage-quick.out

# Full test suite with coverage analysis
.PHONY: test-full
test-full: 
	go test ./... -v -timeout=15m -coverprofile=coverage-full.out
	go tool cover -html=coverage-full.out -o coverage-full.html
	@echo "Full coverage report: coverage-full.html"

# Coverage check - ensure we meet target
.PHONY: test-coverage
test-coverage: test
	@echo "Current coverage:"
	@go tool cover -func=coverage.out | tail -1
	@echo "Package coverage breakdown:"
	@go tool cover -func=coverage.out | grep "%" | sort -nk3 | head -20

# Run tests with resource limits (memory, CPU, timeout)
.PHONY: test-limited
test-limited:
	@echo "Running tests with resource limits..."
	@./scripts/run-tests-with-limits.sh ./... -timeout=10m

# Run QA tests with resource limits
.PHONY: test-qa-limited
test-qa-limited:
	@echo "Running QA tests with resource limits..."
	@./scripts/run-tests-with-limits.sh ./qa/... -timeout=5m

# Updated testing infrastructure with coverage targets

# Install development tools
.PHONY: install-tools
install-tools:
	./scripts/install-tools.sh all

# Show version
.PHONY: version
version:
	@echo $(VERSION)

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  build        - Build the binary"
	@echo "  install      - Install to /usr/local/bin (may require sudo)"
	@echo "  install-user - Install to ~/.local/bin (no sudo required)"
	@echo "  clean        - Clean build artifacts"
	@echo "  test         - Run tests with coverage"
	@echo "  test-qa      - Run QA tests only"
	@echo "  test-quick   - Run quick unit tests"
	@echo "  test-full    - Run full test suite with HTML coverage"
	@echo "  test-coverage - Show current coverage statistics"
	@echo "  test-limited - Run tests with resource limits (memory/CPU)"
	@echo "  test-qa-limited - Run QA tests with resource limits"
	@echo "  install-tools- Install development tools"
	@echo "  version      - Show version"
	@echo "  help         - Show this help"