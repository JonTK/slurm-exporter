# Agent Guidelines for SLURM Exporter

This document provides guidance for AI agents working on the SLURM Exporter project.

## Issue Tracking

This project uses **bd (beads)** for issue tracking.
Run `bd prime` for workflow context, or install hooks (`bd hooks install`) for auto-injection.

**Quick reference:**
- `bd ready` - Find unblocked work
- `bd create "Title" --type task --priority 2` - Create issue
- `bd close <id>` - Complete work
- `bd sync` - Sync with git (run at session end)

For full workflow details: `bd prime`

## Development Workflow

### Code Quality Standards

- **Linter Thresholds** (enforced via golangci-lint):
  - `funlen`: 120 lines max, 60 statements max
  - `gocognit`: 30 cognitive complexity max
  - `nestif`: 5 nesting depth max

- **Test Coverage**: Aim for 80%+ coverage for new code

- **Commit Messages**: Use Conventional Commits format
  - `feat:` - New features
  - `fix:` - Bug fixes
  - `refactor:` - Code refactoring
  - `test:` - Test additions/changes
  - `docs:` - Documentation changes
  - `chore:` - Maintenance tasks

### Refactoring Pattern

When refactoring complex functions, apply this proven pattern:
1. Extract context structs for normalized data
2. Create focused helper methods for repetitive operations
3. Separate data from logic (move to consts/functions)
4. Simplify main functions to orchestration only

### Before Committing

1. Run `go test ./...` to ensure all tests pass
2. Run `golangci-lint run` to check for linter violations
3. Run `gofmt -s -w .` to format code
4. Use descriptive commit messages following Conventional Commits

### Pull Requests

- All PRs require passing CI/CD checks
- Squash commits when merging
- Delete branch after merge
- Update PR description with clear summary of changes

## Project Structure

- `internal/collector/` - Prometheus metric collectors
- `internal/server/` - HTTP server and handlers
- `internal/slurm/` - SLURM client integration
- `internal/config/` - Configuration management
- `cmd/slurm-exporter/` - Main application entry point

## Key Patterns

- **Metric Collection**: Use helper methods for repetitive metric sending
- **Error Handling**: Always check and propagate errors appropriately
- **Logging**: Use structured logging with logrus
- **Context**: Pass context.Context through function chains
