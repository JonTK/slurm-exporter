# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Release strategy with semantic versioning support
- GoReleaser v2 configuration for multi-platform builds
- Comprehensive release documentation and procedures
- Version injection at build time via ldflags
- Release GitHub Actions workflow with validation
- Pre-release support (alpha, beta, rc)

## [0.1.0] - 2026-01-21

### Added
- Initial SLURM exporter implementation with comprehensive features
- Concurrent collection system for efficient metric gathering
- Collection orchestrator with scheduled collection cycles
- Performance monitoring and observability features
- Comprehensive test suite with 75+ test files
- Support for all SLURM resource types (jobs, nodes, partitions, clusters)
- GitHub Actions CI/CD pipeline with multi-platform support
- Multi-platform binary builds (Linux, macOS, Windows)
- Container images for multiple distributions (Alpine, Distroless, standard)
- RPM and DEB package builds for various distributions
- Helm chart for Kubernetes deployment
- Prometheus alerting rules and Grafana dashboards
- Configuration hot-reload capability
- Caching strategies for improved performance

### Changed
- Integrated upstream SLURM client library for API communication
- Improved Windows platform compatibility
- Enhanced test timing and race condition handling
- Optimized concurrent collection scheduling

### Fixed
- Race conditions in concurrent collection scheduling
- Windows platform-specific test failures
- MkDocs documentation build failures
- Linting and static analysis violations
- Performance metrics timer implementation
- CI configuration for development phase

### Security
- Implemented comprehensive security scanning in CI/CD pipeline
- Added input validation and sanitization throughout codebase
- Configured secure defaults for authentication and API communication
- Multi-level security testing (gosec, govulncheck, trivy)