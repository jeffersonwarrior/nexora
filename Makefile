# Makefile for Nexora

# Version information
VERSION ?= 0.28.1
LDFLAGS := -X github.com/nexora/cli/internal/version.Version=$(VERSION)

# Default target
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

# Run tests
.PHONY: test-qa
test-qa:
	go test ./qa/...

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
	go test ./...

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
	@echo "  test         - Run tests"
	@echo "  install-tools- Install development tools"
	@echo "  version      - Show version"
	@echo "  help         - Show this help"