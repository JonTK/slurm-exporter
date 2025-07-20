# Makefile for SLURM Prometheus Exporter

.PHONY: build test lint fmt vet clean run docker docker-build docker-push help install-tools

# Variables
BINARY_NAME=slurm-exporter
PACKAGE=github.com/jontk/slurm-exporter
GO_VERSION=1.21
VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(shell date +%Y-%m-%dT%H:%M:%S%z)
LDFLAGS=-ldflags "-X $(PACKAGE)/pkg/version.Version=$(VERSION) -X $(PACKAGE)/pkg/version.BuildTime=$(BUILD_TIME)"

# Docker variables
DOCKER_REGISTRY ?= localhost
DOCKER_IMAGE ?= $(DOCKER_REGISTRY)/$(BINARY_NAME)
DOCKER_TAG ?= $(VERSION)

# Default target
all: lint test build

# Build the binary
build:
	@echo "Building $(BINARY_NAME)..."
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) .

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	go run $(LDFLAGS) .

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

# Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Lint the code
lint: install-tools
	@echo "Running linter..."
	golangci-lint run

# Format the code
fmt:
	@echo "Formatting code..."
	go fmt ./...

# Run go vet
vet:
	@echo "Running go vet..."
	go vet ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	go clean
	rm -f bin/$(BINARY_NAME)
	rm -f coverage.out coverage.html

# Install development tools
install-tools:
	@echo "Installing development tools..."
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "Installing golangci-lint..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
	}

# Tidy up dependencies
tidy:
	@echo "Tidying up dependencies..."
	go mod tidy

# Update dependencies
update:
	@echo "Updating dependencies..."
	go get -u ./...
	go mod tidy

# Security audit
security:
	@echo "Running security audit..."
	@command -v gosec >/dev/null 2>&1 || { \
		echo "Installing gosec..."; \
		go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest; \
	}
	gosec ./...

# Run all checks (lint, vet, test)
check: fmt vet lint test

# Docker build
docker-build:
	@echo "Building Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .
	docker tag $(DOCKER_IMAGE):$(DOCKER_TAG) $(DOCKER_IMAGE):latest

# Docker push
docker-push:
	@echo "Pushing Docker image $(DOCKER_IMAGE):$(DOCKER_TAG)..."
	docker push $(DOCKER_IMAGE):$(DOCKER_TAG)
	docker push $(DOCKER_IMAGE):latest

# Docker run
docker-run:
	@echo "Running Docker container..."
	docker run --rm -p 8080:8080 $(DOCKER_IMAGE):$(DOCKER_TAG)

# Docker compose up
docker-compose-up:
	@echo "Starting services with Docker Compose..."
	docker-compose up --build

# Docker compose down
docker-compose-down:
	@echo "Stopping services with Docker Compose..."
	docker-compose down

# Install the binary locally
install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(LDFLAGS) .

# Generate mocks for testing
mocks:
	@echo "Generating mocks..."
	@command -v mockgen >/dev/null 2>&1 || { \
		echo "Installing mockgen..."; \
		go install go.uber.org/mock/mockgen@latest; \
	}
	mockgen -source=internal/collector/registry.go -destination=internal/collector/mocks/registry_mock.go
	mockgen -source=internal/slurm/client.go -destination=internal/slurm/mocks/client_mock.go

# Run integration tests (requires SLURM cluster)
integration-test:
	@echo "Running integration tests..."
	@if [ -z "$$SLURM_REST_URL" ]; then \
		echo "SLURM_REST_URL environment variable is required for integration tests"; \
		exit 1; \
	fi
	go test -tags=integration -v ./tests/integration/...

# Check for Go version
check-version:
	@echo "Checking Go version..."
	@go version | grep -q "go1\\.2[1-9]" || { \
		echo "Go $(GO_VERSION) or higher is required"; \
		exit 1; \
	}

# Release preparation
release-prep: check-version clean tidy fmt vet lint test
	@echo "Release preparation complete"

# Show help
help:
	@echo "Available targets:"
	@echo "  build           - Build the binary"
	@echo "  run             - Run the application"
	@echo "  test            - Run tests"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  benchmark       - Run benchmarks"
	@echo "  lint            - Run linter"
	@echo "  fmt             - Format code"
	@echo "  vet             - Run go vet"
	@echo "  clean           - Clean build artifacts"
	@echo "  install-tools   - Install development tools"
	@echo "  tidy            - Tidy up dependencies"
	@echo "  update          - Update dependencies"
	@echo "  security        - Run security audit"
	@echo "  check           - Run all checks (fmt, vet, lint, test)"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-push     - Push Docker image"
	@echo "  docker-run      - Run Docker container"
	@echo "  docker-compose-up   - Start services with Docker Compose"
	@echo "  docker-compose-down - Stop services with Docker Compose"
	@echo "  install         - Install the binary locally"
	@echo "  mocks           - Generate mock files"
	@echo "  integration-test - Run integration tests (requires SLURM_REST_URL)"
	@echo "  check-version   - Check Go version"
	@echo "  release-prep    - Prepare for release"
	@echo "  help            - Show this help message"