# Makefile for airule CLI tool
# This Makefile provides targets for building, testing, and installing the airule application

# Variables
BINARY_NAME := airule
GO := go
GOFLAGS :=
LDFLAGS := -s -w
SRC_DIR := ./cmd/airule

# Version information
VERSION := $(shell git describe --tags --always --dirty 2>/dev/null || echo "0.1.0")
COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
VERSION_FLAGS := -X github.com/upamune/airule/internal/cli.version=$(VERSION) \
                 -X github.com/upamune/airule/internal/cli.commit=$(COMMIT) \
                 -X github.com/upamune/airule/internal/cli.buildDate=$(BUILD_DATE)

# Build the application
.PHONY: build
build:
	@echo "Building $(BINARY_NAME)..."
	$(GO) build $(GOFLAGS) -ldflags "$(LDFLAGS) $(VERSION_FLAGS)" -o $(BINARY_NAME) $(SRC_DIR)
	@echo "Build complete: $(BINARY_NAME)"

# Run tests
.PHONY: test
test:
	@echo "Running tests..."
	$(GO) test -v ./...
	@echo "Tests complete"

# Run tests with coverage
.PHONY: test-coverage
test-coverage:
	@echo "Running tests with coverage..."
	$(GO) test -v -coverprofile=coverage.out ./...
	$(GO) tool cover -html=coverage.out
	@echo "Coverage tests complete"

# Clean build artifacts
.PHONY: clean
clean:
	@echo "Cleaning build artifacts..."
	rm -f $(BINARY_NAME)
	rm -f coverage.out
	@echo "Clean complete"

# Format code
.PHONY: fmt
fmt:
	@echo "Formatting code..."
	$(GO) fmt ./...
	@echo "Formatting complete"

# Show version information
.PHONY: version
version:
	@echo "Version: $(VERSION)"
	@echo "Commit: $(COMMIT)"
	@echo "Build Date: $(BUILD_DATE)"

# Show help information
.PHONY: help
help:
	@echo "airule Makefile"
	@echo ""
	@echo "Targets:"
	@echo "  build          - Build the application"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  clean          - Clean build artifacts"
	@echo "  fmt            - Format code"
	@echo "  version        - Show version information"
	@echo "  help           - Show this help message"
	@echo ""
	@echo "Variables:"
	@echo "  BINARY_NAME    - Name of the binary (default: $(BINARY_NAME))"
	@echo "  GO             - Go command (default: $(GO))"
	@echo "  GOFLAGS        - Additional flags for go command"

# Default target
.DEFAULT_GOAL := build
