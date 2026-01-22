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
	go build $(LDFLAGS) -o bin/$(BINARY_NAME) ./cmd/slurm-exporter

# Run the application
run:
	@echo "Running $(BINARY_NAME)..."
	go run $(LDFLAGS) ./cmd/slurm-exporter

# Run tests
test:
	@echo "Running tests..."
	go test -v -race -coverprofile=coverage.out ./...

# Run tests with short flag (skip integration tests)
test-short:
	@echo "Running short tests..."
	go test -v -race -short ./...

# Run unit tests only
test-unit:
	@echo "Running unit tests..."
	go test -v -race ./internal/collector/ ./internal/config/ ./internal/testutil/

# Run integration tests only
test-integration:
	@echo "Running integration tests..."
	go test -v -race ./internal/integration/

# Run tests with coverage report
test-coverage: test
	@echo "Generating coverage report..."
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run comprehensive test coverage
coverage:
	@echo "Running comprehensive test coverage..."
	./scripts/test-coverage.sh

# Run benchmarks
benchmark:
	@echo "Running benchmarks..."
	go test -bench=. -benchmem ./...

# Run performance benchmarks
bench-performance:
	@echo "Running performance benchmarks..."
	go test -bench=. -benchmem ./internal/performance/

# Run load tests
load-test:
	@echo "Running load tests..."
	./scripts/load-test.sh

# Run CPU profiling
profile-cpu:
	@echo "Running CPU profiling..."
	./scripts/profile.sh -t cpu -d 30s

# Run memory profiling
profile-memory:
	@echo "Running memory profiling..."
	./scripts/profile.sh -t memory -d 30s

# Run all profiling
profile-all:
	@echo "Running all profiling..."
	./scripts/profile.sh -a

# Lint the code with golangci-lint
lint: install-linter
	@echo "Running golangci-lint..."
	golangci-lint run --timeout=10m

# Run golangci-lint with auto-fix
lint-fix: install-linter
	@echo "Running golangci-lint with auto-fix..."
	golangci-lint run --fix --timeout=10m

# Run only fast linters (for pre-commit)
lint-fast: install-linter
	@echo "Running fast linters..."
	golangci-lint run --fast --timeout=5m

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

# Install golangci-lint
install-linter:
	@echo "Checking for golangci-lint..."
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | \
		sh -s -- -b $$(go env GOPATH)/bin latest; \
	}

# Install development tools
install-tools: install-linter
	@echo "Installing development tools..."
	@command -v mockgen >/dev/null 2>&1 || { \
		echo "Installing mockgen..."; \
		go install go.uber.org/mock/mockgen@latest; \
	}
	@command -v gosec >/dev/null 2>&1 || { \
		echo "Installing gosec..."; \
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
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
		go install github.com/securego/gosec/v2/cmd/gosec@latest; \
	}
	gosec ./...

# Security scan (alias for security)
security-scan: security

# Run all checks (lint, vet, test)
check: fmt vet lint test

# Docker Compose commands for development
dev-up:
	@echo "Starting development environment..."
	docker-compose up -d

dev-down:
	@echo "Stopping development environment..."
	docker-compose down

dev-logs:
	@echo "Showing development logs..."
	docker-compose logs -f slurm-exporter

dev-restart:
	@echo "Restarting SLURM exporter..."
	docker-compose restart slurm-exporter

dev-build:
	@echo "Building development images..."
	docker-compose build

dev-shell:
	@echo "Starting development shell..."
	docker-compose run --rm slurm-exporter-dev bash

dev-clean:
	@echo "Cleaning development environment..."
	docker-compose down -v
	docker-compose rm -f

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

# Release build (optimized)
release: clean
	@echo "Building release version..."
	CGO_ENABLED=0 $(GOBUILD) $(LDFLAGS) -a -installsuffix cgo -o $(BINARY_NAME) ./cmd/slurm-exporter

# Build release artifacts
build-release:
	@echo "Building release artifacts..."
	./scripts/build-release.sh

# Version management
version-current:
	@./scripts/version.sh current

version-bump-patch:
	@./scripts/version.sh bump patch

version-bump-minor:
	@./scripts/version.sh bump minor

version-bump-major:
	@./scripts/version.sh bump major

version-set:
	@echo "Usage: make version-set VERSION=v1.2.3"
	@if [ -z "$(VERSION)" ]; then echo "Error: VERSION required"; exit 1; fi
	@./scripts/version.sh set $(VERSION)

# Release preparation
prepare-release-patch:
	@./scripts/version.sh release patch

prepare-release-minor:
	@./scripts/version.sh release minor

prepare-release-major:
	@./scripts/version.sh release major

# Generate changelog
changelog:
	@./scripts/version.sh changelog

# Release preparation
release-prep: check-version clean tidy fmt vet lint test security-scan
	@echo "Release preparation complete"

# GoReleaser validation
goreleaser-check:
	@echo "Checking GoReleaser configuration..."
	@command -v goreleaser >/dev/null 2>&1 || { \
		echo "Installing goreleaser..."; \
		go install github.com/goreleaser/goreleaser@latest; \
	}
	goreleaser check

# GoReleaser snapshot build (local testing)
goreleaser-snapshot: goreleaser-check
	@echo "Building GoReleaser snapshot..."
	goreleaser release --snapshot --clean

# GoReleaser release build
goreleaser-release: goreleaser-check
	@echo "Building GoReleaser release..."
	goreleaser release --clean

# Create release tag (usage: make release-tag VERSION=v0.2.0)
release-tag:
	@if [ -z "$(VERSION)" ]; then echo "Error: VERSION required (e.g., make release-tag VERSION=v0.2.0)"; exit 1; fi
	@echo "Creating release tag $(VERSION)..."
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)
	@echo "Tag created and pushed. GitHub Actions release workflow will start automatically."

# Show help
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build targets:"
	@echo "  build           - Build the binary"
	@echo "  release         - Build optimized release binary"
	@echo "  build-release   - Build all release artifacts"
	@echo "  clean           - Clean build artifacts"
	@echo ""
	@echo "Testing targets:"
	@echo "  test            - Run all tests"
	@echo "  test-short      - Run short tests (no integration)"
	@echo "  test-unit       - Run unit tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  test-coverage   - Run tests with coverage report"
	@echo "  coverage        - Run comprehensive coverage analysis"
	@echo "  benchmark       - Run benchmarks"
	@echo "  bench-performance - Run performance benchmarks"
	@echo ""
	@echo "Quality targets:"
	@echo "  lint            - Run golangci-lint (10m timeout)"
	@echo "  lint-fix        - Run golangci-lint with auto-fix"
	@echo "  lint-fast       - Run fast linters (for pre-commit, 5m timeout)"
	@echo "  fmt             - Format code"
	@echo "  vet             - Run go vet"
	@echo "  security        - Run security audit"
	@echo "  check           - Run all checks (fmt, vet, lint, test)"
	@echo ""
	@echo "Performance targets:"
	@echo "  load-test       - Run load tests"
	@echo "  profile-cpu     - Run CPU profiling"
	@echo "  profile-memory  - Run memory profiling"
	@echo "  profile-all     - Run all profiling"
	@echo ""
	@echo "Version management:"
	@echo "  version-current - Show current version"
	@echo "  version-bump-patch - Bump patch version"
	@echo "  version-bump-minor - Bump minor version"
	@echo "  version-bump-major - Bump major version"
	@echo "  version-set     - Set specific version (VERSION=v1.2.3)"
	@echo ""
	@echo "Release targets:"
	@echo "  prepare-release-patch - Prepare patch release"
	@echo "  prepare-release-minor - Prepare minor release"
	@echo "  prepare-release-major - Prepare major release"
	@echo "  changelog       - Generate changelog"
	@echo "  release-prep    - Run all release preparation checks"
	@echo "  goreleaser-check - Validate GoReleaser configuration"
	@echo "  goreleaser-snapshot - Build GoReleaser snapshot (local testing)"
	@echo "  goreleaser-release - Build GoReleaser release"
	@echo "  release-tag     - Create and push release tag (VERSION=v0.2.0)"
	@echo ""
	@echo "Docker targets:"
	@echo "  docker-build    - Build Docker image"
	@echo "  docker-push     - Push Docker image"
	@echo "  docker-run      - Run Docker container"
	@echo "  docker-compose-up   - Start services with Docker Compose"
	@echo "  docker-compose-down - Stop services with Docker Compose"
	@echo ""
	@echo "Utility targets:"
	@echo "  run             - Run the application"
	@echo "  install         - Install the binary locally"
	@echo "  install-tools   - Install all development tools"
	@echo "  install-linter  - Install golangci-lint"
	@echo "  mocks           - Generate mock files"
	@echo "  tidy            - Tidy up dependencies"
	@echo "  update          - Update dependencies"
	@echo "  check-version   - Check Go version"
	@echo "  help            - Show this help message"