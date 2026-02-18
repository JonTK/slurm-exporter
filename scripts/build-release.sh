#!/bin/bash

# Build release script for SLURM exporter
set -e

# Configuration
VERSION="${VERSION:-dev}"
OUTPUT_DIR="${OUTPUT_DIR:-dist}"
BUILD_PLATFORMS="${BUILD_PLATFORMS:-linux/amd64,linux/arm64,darwin/amd64,darwin/arm64,windows/amd64}"
DOCKER_REGISTRY="${DOCKER_REGISTRY:-ghcr.io}"
DOCKER_IMAGE="${DOCKER_IMAGE:-slurm-exporter}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}SLURM Exporter Release Builder${NC}"
echo "=================================="
echo -e "Version: ${BLUE}${VERSION}${NC}"
echo -e "Output directory: ${BLUE}${OUTPUT_DIR}${NC}"
echo -e "Platforms: ${BLUE}${BUILD_PLATFORMS}${NC}"
echo ""

# Build information
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

echo -e "${YELLOW}Build information:${NC}"
echo -e "  Build time: ${BUILD_TIME}"
echo -e "  Commit: ${COMMIT}"
echo ""

# Clean and create output directory
rm -rf "$OUTPUT_DIR"
mkdir -p "$OUTPUT_DIR"

# Function to build binary for a platform
build_binary() {
    local goos="$1"
    local goarch="$2"
    local goarm="$3"
    
    echo -e "${YELLOW}Building for ${goos}/${goarch}${goarm:+v$goarm}...${NC}"
    
    local binary_name="slurm-exporter"
    if [ "$goos" = "windows" ]; then
        binary_name="${binary_name}.exe"
    fi
    
    local ldflags="-s -w -X main.version=${VERSION} -X main.commit=${COMMIT} -X main.buildTime=${BUILD_TIME}"
    
    env CGO_ENABLED=0 GOOS="$goos" GOARCH="$goarch" ${goarm:+GOARM=$goarm} \
        go build -ldflags="$ldflags" -trimpath -o "${OUTPUT_DIR}/${binary_name}" ./cmd/slurm-exporter
    
    # Create archive
    local archive_name
    local platform_dir="${OUTPUT_DIR}/${goos}-${goarch}${goarm:+v$goarm}"
    mkdir -p "$platform_dir"
    
    # Copy binary and additional files
    cp "${OUTPUT_DIR}/${binary_name}" "$platform_dir/"
    cp README.md "$platform_dir/" 2>/dev/null || echo "# SLURM Exporter ${VERSION}" > "$platform_dir/README.md"
    cp LICENSE* "$platform_dir/" 2>/dev/null || true
    cp configs/config.yaml "$platform_dir/config-example.yaml" 2>/dev/null || true
    
    # Create installation script
    cat > "$platform_dir/install.sh" <<'EOF'
#!/bin/bash
# SLURM Exporter Installation Script

set -e

INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
CONFIG_DIR="${CONFIG_DIR:-/etc/slurm-exporter}"
USER="${SLURM_EXPORTER_USER:-slurm-exporter}"

echo "Installing SLURM Exporter..."

# Create user if it doesn't exist
if ! id "$USER" &>/dev/null; then
    echo "Creating user $USER..."
    sudo useradd --system --no-create-home --shell /bin/false "$USER"
fi

# Create directories
sudo mkdir -p "$INSTALL_DIR" "$CONFIG_DIR"

# Install binary
sudo cp slurm-exporter "$INSTALL_DIR/"
sudo chmod +x "$INSTALL_DIR/slurm-exporter"
sudo chown root:root "$INSTALL_DIR/slurm-exporter"

# Install config if it doesn't exist
if [ ! -f "$CONFIG_DIR/config.yaml" ] && [ -f "config-example.yaml" ]; then
    sudo cp config-example.yaml "$CONFIG_DIR/config.yaml"
    sudo chown "$USER:$USER" "$CONFIG_DIR/config.yaml"
    sudo chmod 640 "$CONFIG_DIR/config.yaml"
fi

echo "SLURM Exporter installed successfully!"
echo "Binary: $INSTALL_DIR/slurm-exporter"
echo "Config: $CONFIG_DIR/config.yaml"
echo ""
echo "To start the exporter:"
echo "  sudo -u $USER $INSTALL_DIR/slurm-exporter --config $CONFIG_DIR/config.yaml"
EOF
    chmod +x "$platform_dir/install.sh"
    
    # Create archive
    cd "$OUTPUT_DIR"
    if [ "$goos" = "windows" ]; then
        archive_name="slurm-exporter-${VERSION}-${goos}-${goarch}${goarm:+v$goarm}.zip"
        zip -r "$archive_name" "${goos}-${goarch}${goarm:+v$goarm}/" > /dev/null
    else
        archive_name="slurm-exporter-${VERSION}-${goos}-${goarch}${goarm:+v$goarm}.tar.gz"
        tar -czf "$archive_name" "${goos}-${goarch}${goarm:+v$goarm}/"
    fi
    
    # Generate checksum
    if command -v sha256sum >/dev/null; then
        sha256sum "$archive_name" > "${archive_name}.sha256"
    elif command -v shasum >/dev/null; then
        shasum -a 256 "$archive_name" > "${archive_name}.sha256"
    fi
    
    cd - > /dev/null
    rm -rf "$platform_dir"
    rm "${OUTPUT_DIR}/${binary_name}"
    
    echo -e "${GREEN}✓ Created ${archive_name}${NC}"
}

# Build for all platforms
echo -e "${YELLOW}Building binaries...${NC}"
IFS=',' read -ra PLATFORMS <<< "$BUILD_PLATFORMS"
for platform in "${PLATFORMS[@]}"; do
    IFS='/' read -ra PARTS <<< "$platform"
    goos="${PARTS[0]}"
    goarch="${PARTS[1]}"
    
    # Handle ARM variants
    if [ "$goarch" = "arm" ]; then
        # Build for ARMv7 by default
        build_binary "$goos" "$goarch" "7"
    else
        build_binary "$goos" "$goarch"
    fi
done

# Build Docker images if Docker is available
if command -v docker >/dev/null 2>&1; then
    echo -e "${YELLOW}Building Docker images...${NC}"
    
    # Build multi-platform image
    if docker buildx version >/dev/null 2>&1; then
        echo "Building multi-platform Docker image..."
        docker buildx build \
            --platform linux/amd64,linux/arm64 \
            --build-arg VERSION="$VERSION" \
            --build-arg BUILD_TIME="$BUILD_TIME" \
            --build-arg COMMIT="$COMMIT" \
            -t "${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${VERSION}" \
            -t "${DOCKER_REGISTRY}/${DOCKER_IMAGE}:latest" \
            --load \
            .
        
        echo -e "${GREEN}✓ Docker images built${NC}"
    else
        echo -e "${YELLOW}Docker buildx not available, building single-platform image...${NC}"
        docker build \
            --build-arg VERSION="$VERSION" \
            --build-arg BUILD_TIME="$BUILD_TIME" \
            --build-arg COMMIT="$COMMIT" \
            -t "${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${VERSION}" \
            -t "${DOCKER_REGISTRY}/${DOCKER_IMAGE}:latest" \
            .
        
        echo -e "${GREEN}✓ Docker image built for current platform${NC}"
    fi
else
    echo -e "${YELLOW}Docker not available, skipping container build${NC}"
fi

# Package Helm chart if Helm is available
if command -v helm >/dev/null 2>&1; then
    echo -e "${YELLOW}Packaging Helm chart...${NC}"
    
    # Update chart version
    if [ -f charts/slurm-exporter/Chart.yaml ]; then
        # Backup original
        cp charts/slurm-exporter/Chart.yaml charts/slurm-exporter/Chart.yaml.bak
        
        # Update versions
        sed -i "s/^version:.*/version: ${VERSION#v}/" charts/slurm-exporter/Chart.yaml
        sed -i "s/^appVersion:.*/appVersion: \"${VERSION}\"/" charts/slurm-exporter/Chart.yaml
        
        # Package chart
        helm package charts/slurm-exporter --destination "$OUTPUT_DIR"
        
        # Restore original
        mv charts/slurm-exporter/Chart.yaml.bak charts/slurm-exporter/Chart.yaml
        
        echo -e "${GREEN}✓ Helm chart packaged${NC}"
    else
        echo -e "${YELLOW}No Helm chart found, skipping${NC}"
    fi
else
    echo -e "${YELLOW}Helm not available, skipping chart packaging${NC}"
fi

# Generate release summary
echo -e "${YELLOW}Generating release summary...${NC}"
cat > "${OUTPUT_DIR}/RELEASE_NOTES.md" <<EOF
# SLURM Exporter ${VERSION}

## Release Information
- **Version:** ${VERSION}
- **Build Time:** ${BUILD_TIME}
- **Commit:** ${COMMIT}

## Artifacts

### Binaries
EOF

# List binary archives
for file in "$OUTPUT_DIR"/*.tar.gz "$OUTPUT_DIR"/*.zip; do
    if [ -f "$file" ]; then
        filename=$(basename "$file")
        size=$(ls -lh "$file" | awk '{print $5}')
        echo "- \`${filename}\` (${size})" >> "${OUTPUT_DIR}/RELEASE_NOTES.md"
    fi
done

cat >> "${OUTPUT_DIR}/RELEASE_NOTES.md" <<EOF

### Container Images
\`\`\`bash
docker pull ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${VERSION}
docker pull ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:latest
\`\`\`

### Helm Chart
EOF

# List Helm charts
for file in "$OUTPUT_DIR"/*.tgz; do
    if [ -f "$file" ]; then
        filename=$(basename "$file")
        echo "- \`${filename}\`" >> "${OUTPUT_DIR}/RELEASE_NOTES.md"
    fi
done

cat >> "${OUTPUT_DIR}/RELEASE_NOTES.md" <<EOF

## Installation

### Binary Installation
1. Download the appropriate binary for your platform
2. Extract the archive
3. Run the included \`install.sh\` script (Linux/macOS) or manually install

### Docker Installation
\`\`\`bash
docker run -d \\
  --name slurm-exporter \\
  -p 10341:10341 \\
  -v /path/to/config.yaml:/etc/slurm-exporter/config.yaml \\
  ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${VERSION}
\`\`\`

### Helm Installation
\`\`\`bash
helm install slurm-exporter ./${OUTPUT_DIR}/slurm-exporter-*.tgz
\`\`\`

## Verification
All binary archives include SHA256 checksums for verification:
\`\`\`bash
sha256sum -c slurm-exporter-${VERSION}-*.sha256
\`\`\`
EOF

# List all release artifacts
echo ""
echo -e "${GREEN}Release build completed!${NC}"
echo -e "${BLUE}Artifacts created in ${OUTPUT_DIR}/:${NC}"
ls -la "$OUTPUT_DIR/"

echo ""
echo -e "${GREEN}Release summary:${NC}"
echo -e "  Version: ${VERSION}"
echo -e "  Commit: ${COMMIT}"
echo -e "  Build time: ${BUILD_TIME}"
echo -e "  Artifacts: $(ls "$OUTPUT_DIR" | wc -l) files"

# Calculate total size
total_size=$(du -sh "$OUTPUT_DIR" 2>/dev/null | cut -f1 || echo "unknown")
echo -e "  Total size: ${total_size}"

echo ""
echo -e "${BLUE}Next steps:${NC}"
echo -e "  1. Test the binaries on target platforms"
echo -e "  2. Push Docker images: docker push ${DOCKER_REGISTRY}/${DOCKER_IMAGE}:${VERSION}"
echo -e "  3. Create GitHub release with artifacts from ${OUTPUT_DIR}/"
echo -e "  4. Update documentation with new version"