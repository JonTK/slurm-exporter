# Release Process

This document describes the release process for slurm-exporter and provides guidelines for maintaining release quality and consistency.

## Semantic Versioning

slurm-exporter follows [Semantic Versioning](https://semver.org/) in the format `vMAJOR.MINOR.PATCH[-prerelease]`.

- **MAJOR**: Incompatible API changes or significant feature additions
- **MINOR**: New features that are backward compatible
- **PATCH**: Bug fixes and performance improvements that are backward compatible
- **Pre-release**: `-alpha.1`, `-beta.1`, `-rc.1` for testing releases

### Version Support Policy

- **Latest Release**: Full support with bug fixes, security updates, and performance improvements
- **Previous Minor Version**: Security updates only (3 months support window)
- **Older Versions**: No support; users encouraged to upgrade

## Pre-Release Checklist

Before starting the release process, ensure:

- [ ] All tests passing: `make test`
- [ ] Code quality checks pass: `make lint`
- [ ] Security scans complete: `make security-scan`
- [ ] No open critical bugs or high-priority issues
- [ ] Feature branch PRs merged and deployed
- [ ] Team agreement on version number and release scope
- [ ] Release notes prepared in CHANGELOG.md

## Release Process

### 1. Prepare for Release

Update the CHANGELOG.md file with all changes from the `[Unreleased]` section:

```bash
# Create release branch
git checkout -b release/v0.2.0

# Update CHANGELOG.md
# - Move [Unreleased] items to new [0.2.0] section
# - Update date to today
# - Create new [Unreleased] section

git add CHANGELOG.md
git commit -m "chore: prepare v0.2.0 release"
```

### 2. Validate Release

Run comprehensive tests and checks:

```bash
make clean
make test
make lint
make security-scan
```

All checks must pass before proceeding.

### 3. Create Release PR

Push release branch and create a PR for final review:

```bash
git push -u origin release/v0.2.0

# Create PR with template mentioning:
# - Release version
# - Key features and fixes
# - Backward compatibility notes
# - Testing performed
```

### 4. Merge Release PR

After team review and approval:

```bash
git checkout main
git pull origin main
git merge --squash release/v0.2.0
git commit -m "chore: prepare v0.2.0 release"
git push origin main
```

### 5. Tag Release

Create an annotated git tag:

```bash
git tag -a v0.2.0 -m "Release v0.2.0"
git push origin v0.2.0
```

The tag push automatically triggers the GitHub Actions release workflow.

### 6. Verify Release

The GitHub Actions workflow will:

1. **Validate**: Check version format matches `vX.Y.Z` pattern
2. **Test**: Run full test suite
3. **Build**: Create platform-specific binaries
4. **Package**: Generate archives, checksums, DEB/RPM packages
5. **Release**: Create GitHub Release with artifacts
6. **Container**: Build and push Docker images
7. **Documentation**: Auto-update version references (if not pre-release)

Monitor the Actions tab for completion and any errors.

### 7. Verify Artifacts

After workflow completes:

```bash
# Download and test binary on each platform
# Verify checksums
# Test basic functionality: ./slurm-exporter --version

# Verify package installation
apt install ./slurm-exporter_0.2.0_amd64.deb  # Ubuntu/Debian
rpm -i slurm-exporter-0.2.0-1.x86_64.rpm     # CentOS/RHEL

# Verify Go module availability
go get github.com/jontk/slurm-exporter@v0.2.0
```

### 8. Publish Release Announcement

Share release announcement:

```bash
# Tweet/social media
# Send to mailing lists
# Update website documentation
# Blog post (for major releases)
```

## Pre-Release Process

For alpha, beta, and release candidate releases:

### 1. Create Pre-Release

```bash
git tag -a v0.2.0-rc.1 -m "Release Candidate v0.2.0-rc.1"
git push origin v0.2.0-rc.1
```

### 2. Test Pre-Release

- Trigger extended testing on multiple environments
- Gather community feedback
- Monitor for issues in real-world scenarios

### 3. Iterate

```bash
# Fix issues found during pre-release testing
git tag -a v0.2.0-rc.2 -m "Release Candidate v0.2.0-rc.2"
git push origin v0.2.0-rc.2
```

### 4. Release Final Version

Once pre-release is stable, release final version following standard process.

## Hotfix Process

For critical bugs in released versions:

```bash
# Create hotfix branch from tag
git checkout -b hotfix/v0.1.1 v0.1.0

# Fix the issue and test
git commit -m "fix: critical security issue in collector"

# Review and merge to main
git checkout main
git merge hotfix/v0.1.1

# Tag hotfix release
git tag -a v0.1.1 -m "Hotfix v0.1.1: Critical security issue"
git push origin main
git push origin v0.1.1
```

## GoReleaser Configuration

The release process uses [GoReleaser v2](https://goreleaser.com/). See `.goreleaser.yml` for:

- Binary build configurations
- Multi-platform and multi-architecture support
- Archive and packaging formats
- Changelog generation from commits
- Release notes templates
- Docker image building

### Key Build Settings

- **Static Binaries**: `CGO_ENABLED=0` for maximum compatibility
- **Platforms**: Linux, macOS, Windows
- **Architectures**: amd64, arm64, arm (except Windows)
- **Archives**: `.tar.gz` (Unix), `.zip` (Windows)
- **Packages**: DEB and RPM for multiple distributions
- **Images**: Docker images for Alpine, Distroless, and standard

## Changelog Generation

GoReleaser automatically generates changelogs from git commits using [Conventional Commits](https://www.conventionalcommits.org/):

- `feat:` → New Features
- `fix:` → Bug Fixes
- `perf:` → Performance Improvements
- `security:` or `sec:` → Security Updates
- Excluded: `docs:`, `test:`, `chore:`, `ci:`, `refactor:`

Ensure commit messages follow conventional format for proper categorization.

## Release Artifacts

Each release produces:

### Binary Artifacts
- `slurm-exporter-v0.2.0-linux-amd64.tar.gz`
- `slurm-exporter-v0.2.0-linux-arm64.tar.gz`
- `slurm-exporter-v0.2.0-darwin-amd64.tar.gz`
- `slurm-exporter-v0.2.0-darwin-arm64.tar.gz`
- `slurm-exporter-v0.2.0-windows-amd64.zip`

### Packages
- `slurm-exporter-0.2.0-1.x86_64.rpm` (CentOS, RHEL, Fedora, AlmaLinux, Rocky)
- `slurm-exporter_0.2.0_amd64.deb` (Ubuntu, Debian)

### Containers
- `jontk/slurm-exporter:v0.2.0` (standard)
- `jontk/slurm-exporter:v0.2.0-alpine` (Alpine)
- `jontk/slurm-exporter:v0.2.0-distroless` (Distroless)
- `jontk/slurm-exporter:latest` (latest standard)

### Checksums & Source
- `slurm-exporter-v0.2.0-checksums.txt` (SHA256)
- Source archives (`.tar.gz`, `.zip`)

## Release Validation Checklist

After release is published:

- [ ] GitHub Release created with correct version and artifacts
- [ ] All binary artifacts downloaded and checksums verified
- [ ] Binaries execute correctly on each platform
- [ ] Package installation works (DEB/RPM)
- [ ] `go get` resolves new version
- [ ] Docker images exist in registry with correct tags
- [ ] CHANGELOG.md is up to date
- [ ] Documentation reflects new version
- [ ] Release announcement shared with community

## Troubleshooting

### Release workflow fails at validation

Check that git tag follows pattern: `v0.0.0` (with 'v' prefix)

```bash
# Correct tags
git tag v0.2.0      # ✓
git tag v0.2.0-rc.1 # ✓

# Incorrect tags
git tag 0.2.0       # ✗ (no 'v' prefix)
git tag release-0.2.0 # ✗ (wrong prefix)
```

### GoReleaser build fails

Ensure:
- All tests pass locally: `make test`
- `go mod tidy` is clean
- Version injection variables exist in `internal/version/version.go`
- ldflags match configuration in `.goreleaser.yml`

### Docker image push fails

Verify credentials:

```bash
# Check Docker credentials
echo $DOCKER_PASSWORD | docker login -u $DOCKER_USERNAME --password-stdin

# Manually test push (if needed for debugging)
docker push jontk/slurm-exporter:v0.2.0
```

### Release artifacts incomplete

Check GitHub Actions logs:

```bash
# Visit https://github.com/jontk/slurm-exporter/actions
# Find release workflow run for your tag
# Review step logs for specific failures
```

## See Also

- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [Conventional Commits](https://www.conventionalcommits.org/)
- [GoReleaser Documentation](https://goreleaser.com/)
- [GitHub Releases](https://docs.github.com/en/repositories/releasing-projects-on-github/about-releases)
