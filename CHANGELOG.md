# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2026-02-08

### Changed
- Upgraded slurm-client dependency from development version (v0.2.5-dev) to official release v0.3.0 (#39)
  - Validated against live SLURM cluster running API version v0.0.44
  - All integration tests pass with real cluster data (nodes, jobs, partitions)
  - Confirmed rate limiting and adapter pattern functionality
- Re-enabled all disabled collectors after slurm-client API migration
  - All 16 collectors now active and operational

### Fixed
- **Critical**: Fixed duplicate collection per scrape causing 2x API load (#35)
- **Critical**: Fixed user job metrics key mismatch (UserID vs UserName) causing lookup failures (#35)
- Fixed partition metrics using placeholder data instead of real API queries (#35)
  - Idle nodes, down nodes, allocated CPUs/jobs now use actual data
  - Time limits read from partition Maximums/Defaults
- Fixed system collector using simulated values instead of real data (#35)
  - Load averages now read from `/proc/loadavg`
  - Disk usage uses `syscall.Statfs()` for accurate reporting
  - Active controllers queried from SLURM Diagnostics API
  - Config timestamps read from file modification time
- Fixed QoS resource limits extraction to properly parse TRES limits (#35)
- Fixed job memory unit conversion (was off by 1048576x - MB vs bytes) (#35)
- Fixed job NodeCount using actual field instead of hardcoded value (#35)
- Fixed job Account and QoS fields using actual data instead of defaults (#35)
- Fixed node allocation data using real AllocCPUs/AllocMemory instead of estimates (#35)
- Fixed performance collector configuration alignment with registry (#35)
- Fixed health check URL construction for non-default addresses (#35)
- Fixed integration test compilation errors (#37)
  - Corrected slurm-client interface types and API usage
  - Removed invalid t.Parallel() calls in testify suite methods
  - Fixed collector constructor signatures
  - Added nil safety guards and comprehensive API documentation
- Fixed release workflow to respect branch protection rules (#33)
  - Now creates PRs for documentation updates instead of direct commits
  - Resolved credential conflicts between actions/checkout and create-pull-request
  - Added required pull-requests and id-token permissions
- Removed inaccurate profiling configuration documentation (#40)
  - Profiling feature exists in code but is not integrated into application
  - Removed non-existent config paths from documentation

### Security
- Fixed GO-2026-4337 vulnerability: Unexpected session resumption in crypto/tls (#37)
  - Upgraded Go from 1.24.4 to 1.24.13
  - Updated all GitHub Actions workflows to use Go 1.24.13
  - Affects TLS server connections, HTTP client, and profile storage

### Chore
- Updated .gitignore to exclude test scripts and artifacts with hardcoded credentials (#36)
  - Prevents accidental commits of plaintext JWT tokens
  - Test scripts moved to external directory for security

## [0.2.0] - 2026-01-27

### Added
- Release strategy with semantic versioning support
- GoReleaser v2 configuration for multi-platform builds
- Comprehensive release documentation and procedures
- Version injection at build time via ldflags
- Release GitHub Actions workflow with validation
- Pre-release support (alpha, beta, rc)
- Full support for SLURM API v0.0.44 with adapter pattern (#32)
  - Auto-detection of latest API version
  - All 16 collectors work correctly with v0.0.44 endpoints
  - Successfully tested with rocky9 SLURM server (v25.11.1)

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