# Makefile for patternizer

# Variables
BINARY_NAME := patternizer
GO_VERSION := 1.24
GOLANGCI_LINT_VERSION := v2.1.6
IMAGE_NAME := patternizer
LOCAL_TAG := local
SRC_DIR := src
CONTAINER_ENGINE := podman

# Go-related variables
GO_CMD := go
GO_BUILD := $(GO_CMD) build
GO_TEST := $(GO_CMD) test
GO_CLEAN := $(GO_CMD) clean
GO_VET := $(GO_CMD) vet
GO_FMT := gofmt

# Default target
.DEFAULT_GOAL := help

# Help target
.PHONY: help
help: ## Show this help message
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-20s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

# Build target
.PHONY: build
build: ## Build the patternizer binary
	@echo "Building patternizer..."
	cd $(SRC_DIR) && $(GO_BUILD) -v -o $(BINARY_NAME) .
	@echo "Build complete: $(SRC_DIR)/$(BINARY_NAME)"

# Clean target
.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	cd $(SRC_DIR) && $(GO_CLEAN)
	rm -f $(SRC_DIR)/$(BINARY_NAME)
	rm -f $(SRC_DIR)/coverage.out
	@echo "Clean complete"

# Install dependencies
.PHONY: deps
deps: ## Download and install Go dependencies
	@echo "Installing dependencies..."
	cd $(SRC_DIR) && $(GO_CMD) mod download
	cd $(SRC_DIR) && $(GO_CMD) mod tidy
	@echo "Dependencies installed"

# Unit tests
.PHONY: test-unit
test-unit: ## Run unit tests
	@echo "Running unit tests..."
	cd $(SRC_DIR) && $(GO_TEST) -v ./...

# Test with coverage
.PHONY: test-coverage
test-coverage: ## Run unit tests with coverage report
	@echo "Running unit tests with coverage..."
	cd $(SRC_DIR) && $(GO_TEST) ./... -coverprofile=coverage.out
	cd $(SRC_DIR) && $(GO_CMD) tool cover -func=coverage.out

# Integration tests
.PHONY: test-integration
test-integration: build ## Run integration tests
	@echo "Running integration tests..."
	PATTERNIZER_BINARY=./$(SRC_DIR)/$(BINARY_NAME) ./test/integration_test.sh

# All tests
.PHONY: test
test: test-unit test-integration ## Run all tests (unit + integration)

# Lint target
.PHONY: lint
lint: lint-fmt lint-vet lint-golangci ## Run all linting checks

# Format check
.PHONY: lint-fmt
lint-fmt: ## Check Go formatting
	@echo "Checking Go formatting..."
	@cd $(SRC_DIR) && if [ "$$($(GO_FMT) -s -l . | wc -l)" -gt 0 ]; then \
		echo "The following files are not formatted:"; \
		$(GO_FMT) -s -l .; \
		exit 1; \
	fi
	@echo "Go formatting check passed"

# Vet check
.PHONY: lint-vet
lint-vet: ## Run go vet
	@echo "Running go vet..."
	cd $(SRC_DIR) && $(GO_VET) ./...
	@echo "Go vet passed"

# golangci-lint check
.PHONY: lint-golangci
lint-golangci: ## Run golangci-lint
	@echo "Running golangci-lint..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
	fi
	cd $(SRC_DIR) && $$(go env GOPATH)/bin/golangci-lint run
	@echo "golangci-lint passed"

# Format code
.PHONY: fmt
fmt: ## Format Go code
	@echo "Formatting Go code..."
	cd $(SRC_DIR) && $(GO_FMT) -s -w .
	@echo "Go code formatted"

# Local container build
.PHONY: local-container-build
local-container-build: ## Build container image locally
	@echo "Building container image..."
	$(CONTAINER_ENGINE) build -t $(IMAGE_NAME):$(LOCAL_TAG) -f Containerfile .
	@echo "Container image built: $(IMAGE_NAME):$(LOCAL_TAG)"

# Full CI pipeline locally
.PHONY: ci
ci: lint build test ## Run the full CI pipeline locally

# Development setup
.PHONY: dev-setup
dev-setup: deps ## Set up development environment
	@echo "Setting up development environment..."
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b $$(go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
	fi
	@echo "Development environment ready"

# Version info
.PHONY: version
version: ## Show version information
	@echo "Go version: $$(go version)"
	@echo "Git commit: $$(git rev-parse --short HEAD 2>/dev/null || echo 'unknown')"
	@echo "Build date: $$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# Generate documentation
.PHONY: docs
docs: ## Generate Go documentation
	@echo "Generating documentation..."
	cd $(SRC_DIR) && $(GO_CMD) doc -all ./...

# Quick check (fast feedback loop)
.PHONY: check
check: lint-fmt lint-vet build test-unit ## Quick check (format, vet, build, unit tests)

.PHONY: all
all: clean deps lint build test local-container-build ## Run everything
