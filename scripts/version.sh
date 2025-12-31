#!/bin/bash

# Version management script for SLURM exporter
set -e

# Configuration
VERSION_FILE="VERSION"
CHANGELOG_FILE="CHANGELOG.md"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to show usage
show_usage() {
    echo "Version Management Script for SLURM Exporter"
    echo ""
    echo "Usage: $0 [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  current                    Show current version"
    echo "  bump <major|minor|patch>   Bump version"
    echo "  set <version>             Set specific version"
    echo "  tag                       Create git tag for current version"
    echo "  changelog                 Generate changelog"
    echo "  release                   Prepare release (bump, tag, changelog)"
    echo ""
    echo "Options:"
    echo "  --dry-run                 Show what would be done without making changes"
    echo "  --prerelease <suffix>     Add prerelease suffix (alpha, beta, rc)"
    echo "  --help                    Show this help"
    echo ""
    echo "Examples:"
    echo "  $0 current"
    echo "  $0 bump minor"
    echo "  $0 bump patch --prerelease alpha"
    echo "  $0 set v1.2.3"
    echo "  $0 release --dry-run"
}

# Function to get current version
get_current_version() {
    if [ -f "$VERSION_FILE" ]; then
        cat "$VERSION_FILE"
    elif git describe --tags --exact-match 2>/dev/null; then
        git describe --tags --exact-match
    elif git describe --tags 2>/dev/null; then
        git describe --tags
    else
        echo "v0.0.0"
    fi
}

# Function to parse version
parse_version() {
    local version="$1"
    # Remove 'v' prefix if present
    version="${version#v}"
    
    # Split version into parts
    IFS='.' read -ra PARTS <<< "$version"
    MAJOR="${PARTS[0]:-0}"
    MINOR="${PARTS[1]:-0}"
    PATCH="${PARTS[2]:-0}"
    
    # Handle prerelease suffix
    if [[ "$PATCH" =~ ^([0-9]+)(.*)$ ]]; then
        PATCH="${BASH_REMATCH[1]}"
        PRERELEASE="${BASH_REMATCH[2]}"
    else
        PRERELEASE=""
    fi
}

# Function to format version
format_version() {
    local major="$1"
    local minor="$2"
    local patch="$3"
    local prerelease="$4"
    
    echo "v${major}.${minor}.${patch}${prerelease}"
}

# Function to bump version
bump_version() {
    local bump_type="$1"
    local prerelease_suffix="$2"
    
    local current_version=$(get_current_version)
    parse_version "$current_version"
    
    case "$bump_type" in
        major)
            MAJOR=$((MAJOR + 1))
            MINOR=0
            PATCH=0
            PRERELEASE=""
            ;;
        minor)
            MINOR=$((MINOR + 1))
            PATCH=0
            PRERELEASE=""
            ;;
        patch)
            PATCH=$((PATCH + 1))
            PRERELEASE=""
            ;;
        *)
            echo -e "${RED}Error: Invalid bump type. Use major, minor, or patch.${NC}"
            exit 1
            ;;
    esac
    
    # Add prerelease suffix if specified
    if [ -n "$prerelease_suffix" ]; then
        PRERELEASE="-${prerelease_suffix}.1"
    fi
    
    local new_version=$(format_version "$MAJOR" "$MINOR" "$PATCH" "$PRERELEASE")
    echo "$new_version"
}

# Function to set version
set_version() {
    local version="$1"
    local dry_run="$2"
    
    # Validate version format
    if ! [[ "$version" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.-]+)?$ ]]; then
        echo -e "${RED}Error: Invalid version format. Use semantic versioning (e.g., v1.2.3).${NC}"
        exit 1
    fi
    
    # Ensure version starts with 'v'
    if [[ ! "$version" =~ ^v ]]; then
        version="v$version"
    fi
    
    if [ "$dry_run" = "true" ]; then
        echo -e "${YELLOW}Would set version to: ${version}${NC}"
        return
    fi
    
    echo "$version" > "$VERSION_FILE"
    echo -e "${GREEN}Version set to: ${version}${NC}"
    
    # Update version in files
    update_version_files "$version"
}

# Function to update version references in files
update_version_files() {
    local version="$1"
    local dry_run="$2"
    
    local files_to_update=(
        "charts/slurm-exporter/Chart.yaml"
        "README.md"
        "docs/installation.md"
    )
    
    for file in "${files_to_update[@]}"; do
        if [ -f "$file" ]; then
            if [ "$dry_run" = "true" ]; then
                echo -e "${YELLOW}Would update version in: ${file}${NC}"
            else
                case "$file" in
                    *.yaml)
                        # Update Helm chart
                        sed -i "s/^version:.*/version: ${version#v}/" "$file"
                        sed -i "s/^appVersion:.*/appVersion: \"${version}\"/" "$file"
                        ;;
                    *.md)
                        # Update documentation
                        sed -i "s/slurm-exporter:v[0-9]\+\.[0-9]\+\.[0-9]\+[^[:space:]]*/slurm-exporter:${version}/g" "$file"
                        sed -i "s/download\/v[0-9]\+\.[0-9]\+\.[0-9]\+[^\/]*/download\/${version}/g" "$file"
                        ;;
                esac
                echo -e "${GREEN}Updated version in: ${file}${NC}"
            fi
        fi
    done
}

# Function to create git tag
create_tag() {
    local version="$1"
    local dry_run="$2"
    
    if [ "$dry_run" = "true" ]; then
        echo -e "${YELLOW}Would create git tag: ${version}${NC}"
        return
    fi
    
    # Check if tag already exists
    if git tag -l | grep -q "^${version}$"; then
        echo -e "${RED}Error: Tag ${version} already exists${NC}"
        exit 1
    fi
    
    # Create annotated tag
    git tag -a "$version" -m "Release $version"
    echo -e "${GREEN}Created git tag: ${version}${NC}"
}

# Function to generate changelog
generate_changelog() {
    local version="$1"
    local dry_run="$2"
    
    local current_version=$(get_current_version)
    local previous_tag=$(git describe --tags --abbrev=0 HEAD^ 2>/dev/null || echo "")
    
    echo -e "${YELLOW}Generating changelog...${NC}"
    
    local changelog_content=""
    
    # Generate header
    changelog_content+="# Changelog\n\n"
    changelog_content+="All notable changes to this project will be documented in this file.\n\n"
    changelog_content+="The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),\n"
    changelog_content+="and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).\n\n"
    
    # Add current version
    changelog_content+="## [${version#v}] - $(date +%Y-%m-%d)\n\n"
    
    # Get commits since last tag
    if [ -n "$previous_tag" ]; then
        echo "Analyzing commits from $previous_tag to HEAD..."
        
        # Categorize commits
        local features=$(git log --pretty=format:"- %s" ${previous_tag}..HEAD | grep -E "^- (feat|feature)" | head -20)
        local fixes=$(git log --pretty=format:"- %s" ${previous_tag}..HEAD | grep -E "^- (fix|bugfix)" | head -20)
        local docs=$(git log --pretty=format:"- %s" ${previous_tag}..HEAD | grep -E "^- (docs|doc)" | head -10)
        local other=$(git log --pretty=format:"- %s" ${previous_tag}..HEAD | grep -vE "^- (feat|feature|fix|bugfix|docs|doc)" | head -10)
        
        if [ -n "$features" ]; then
            changelog_content+="### Added\n${features}\n\n"
        fi
        
        if [ -n "$fixes" ]; then
            changelog_content+="### Fixed\n${fixes}\n\n"
        fi
        
        if [ -n "$docs" ]; then
            changelog_content+="### Documentation\n${docs}\n\n"
        fi
        
        if [ -n "$other" ]; then
            changelog_content+="### Other Changes\n${other}\n\n"
        fi
    else
        changelog_content+="### Added\n- Initial release\n\n"
    fi
    
    # Add previous versions if changelog exists
    if [ -f "$CHANGELOG_FILE" ] && [ -s "$CHANGELOG_FILE" ]; then
        # Skip the first few lines (header) and append existing content
        changelog_content+="$(tail -n +5 "$CHANGELOG_FILE")"
    fi
    
    if [ "$dry_run" = "true" ]; then
        echo -e "${YELLOW}Would generate changelog:${NC}"
        echo -e "$changelog_content" | head -30
        echo "..."
    else
        echo -e "$changelog_content" > "$CHANGELOG_FILE"
        echo -e "${GREEN}Generated changelog: ${CHANGELOG_FILE}${NC}"
    fi
}

# Function to prepare release
prepare_release() {
    local bump_type="$1"
    local prerelease_suffix="$2"
    local dry_run="$3"
    
    if [ -z "$bump_type" ]; then
        echo -e "${RED}Error: Bump type required for release. Use major, minor, or patch.${NC}"
        exit 1
    fi
    
    echo -e "${BLUE}Preparing release...${NC}"
    
    # Check if working directory is clean
    if ! git diff-index --quiet HEAD --; then
        echo -e "${RED}Error: Working directory is not clean. Commit your changes first.${NC}"
        exit 1
    fi
    
    # Bump version
    local new_version=$(bump_version "$bump_type" "$prerelease_suffix")
    echo -e "${YELLOW}New version: ${new_version}${NC}"
    
    # Set version
    set_version "$new_version" "$dry_run"
    
    # Update version files
    update_version_files "$new_version" "$dry_run"
    
    # Generate changelog
    generate_changelog "$new_version" "$dry_run"
    
    if [ "$dry_run" != "true" ]; then
        # Commit changes
        git add "$VERSION_FILE" "$CHANGELOG_FILE"
        
        # Add other updated files
        for file in charts/slurm-exporter/Chart.yaml README.md docs/installation.md; do
            if [ -f "$file" ]; then
                git add "$file"
            fi
        done
        
        git commit -m "chore: prepare release $new_version"
        
        # Create tag
        create_tag "$new_version" "$dry_run"
        
        echo ""
        echo -e "${GREEN}Release $new_version prepared!${NC}"
        echo -e "${BLUE}Next steps:${NC}"
        echo -e "  1. Review the changes: git show"
        echo -e "  2. Push the changes: git push origin main"
        echo -e "  3. Push the tag: git push origin $new_version"
        echo -e "  4. Create GitHub release or run CI/CD pipeline"
    fi
}

# Parse command line arguments
DRY_RUN=false
PRERELEASE_SUFFIX=""

while [[ $# -gt 0 ]]; do
    case $1 in
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --prerelease)
            PRERELEASE_SUFFIX="$2"
            shift 2
            ;;
        --help|-h)
            show_usage
            exit 0
            ;;
        *)
            break
            ;;
    esac
done

# Main command handling
COMMAND="$1"
case "$COMMAND" in
    current)
        echo "Current version: $(get_current_version)"
        ;;
    bump)
        BUMP_TYPE="$2"
        if [ -z "$BUMP_TYPE" ]; then
            echo -e "${RED}Error: Bump type required (major, minor, patch)${NC}"
            exit 1
        fi
        new_version=$(bump_version "$BUMP_TYPE" "$PRERELEASE_SUFFIX")
        set_version "$new_version" "$DRY_RUN"
        ;;
    set)
        VERSION="$2"
        if [ -z "$VERSION" ]; then
            echo -e "${RED}Error: Version required${NC}"
            exit 1
        fi
        set_version "$VERSION" "$DRY_RUN"
        ;;
    tag)
        version=$(get_current_version)
        create_tag "$version" "$DRY_RUN"
        ;;
    changelog)
        version=$(get_current_version)
        generate_changelog "$version" "$DRY_RUN"
        ;;
    release)
        BUMP_TYPE="$2"
        prepare_release "$BUMP_TYPE" "$PRERELEASE_SUFFIX" "$DRY_RUN"
        ;;
    *)
        echo -e "${RED}Error: Unknown command: $COMMAND${NC}"
        echo ""
        show_usage
        exit 1
        ;;
esac