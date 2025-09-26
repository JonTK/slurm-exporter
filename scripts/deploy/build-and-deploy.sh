#!/bin/bash

# SLURM Exporter Build and Deploy Script
# This script builds the Docker image and deploys it to Kubernetes

set -euo pipefail

# Default values
ENVIRONMENT=${ENVIRONMENT:-"development"}
DOCKER_REGISTRY=${DOCKER_REGISTRY:-""}
DOCKER_REPOSITORY=${DOCKER_REPOSITORY:-"slurm-exporter"}
IMAGE_TAG=${IMAGE_TAG:-"$(git rev-parse --short HEAD)"}
PUSH_IMAGE=${PUSH_IMAGE:-"true"}
SKIP_BUILD=${SKIP_BUILD:-"false"}
SKIP_DEPLOY=${SKIP_DEPLOY:-"false"}
BUILD_ARGS=""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to show usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Build Docker image and deploy SLURM Exporter to Kubernetes

OPTIONS:
    -e, --environment      Environment to deploy to [default: development]
    -r, --registry         Docker registry [default: none]
    -p, --repository       Docker repository [default: slurm-exporter]
    -t, --tag              Image tag [default: git commit hash]
    --build-args           Additional Docker build arguments
    --skip-build           Skip Docker build step
    --skip-deploy          Skip deployment step
    --no-push              Don't push image to registry
    -h, --help             Show this help message

EXAMPLES:
    # Build and deploy to development
    $0 -e development
    
    # Build with custom registry and deploy to production
    $0 -e production -r myregistry.com -t v1.0.0
    
    # Only build, don't deploy
    $0 --skip-deploy -t latest
    
    # Only deploy with existing image
    $0 --skip-build -t v1.0.0

EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -e|--environment)
                ENVIRONMENT="$2"
                shift 2
                ;;
            -r|--registry)
                DOCKER_REGISTRY="$2"
                shift 2
                ;;
            -p|--repository)
                DOCKER_REPOSITORY="$2"
                shift 2
                ;;
            -t|--tag)
                IMAGE_TAG="$2"
                shift 2
                ;;
            --build-args)
                BUILD_ARGS="$2"
                shift 2
                ;;
            --skip-build)
                SKIP_BUILD="true"
                shift
                ;;
            --skip-deploy)
                SKIP_DEPLOY="true"
                shift
                ;;
            --no-push)
                PUSH_IMAGE="false"
                shift
                ;;
            -h|--help)
                usage
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done
}

# Check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check if docker is installed
    if ! command -v docker &> /dev/null; then
        print_error "docker is not installed or not in PATH"
        exit 1
    fi
    
    # Check if git is installed (for getting commit hash)
    if ! command -v git &> /dev/null; then
        print_warning "git is not installed - using 'latest' as default tag"
        IMAGE_TAG="latest"
    fi
    
    # Check if we're in a git repository
    if [[ "$IMAGE_TAG" == "$(git rev-parse --short HEAD 2>/dev/null || echo 'latest')" ]]; then
        if ! git rev-parse --git-dir &> /dev/null; then
            print_warning "Not in a git repository - using 'latest' as tag"
            IMAGE_TAG="latest"
        fi
    fi
    
    print_success "Prerequisites check passed"
}

# Build Docker image
build_image() {
    if [[ "$SKIP_BUILD" == "true" ]]; then
        print_info "Skipping Docker build"
        return 0
    fi
    
    local image_name="$DOCKER_REPOSITORY:$IMAGE_TAG"
    if [[ -n "$DOCKER_REGISTRY" ]]; then
        image_name="$DOCKER_REGISTRY/$image_name"
    fi
    
    print_info "Building Docker image: $image_name"
    
    # Build the image
    local build_cmd="docker build -t $image_name ."
    if [[ -n "$BUILD_ARGS" ]]; then
        build_cmd="$build_cmd $BUILD_ARGS"
    fi
    
    print_info "Build command: $build_cmd"
    
    if eval "$build_cmd"; then
        print_success "Docker image built successfully: $image_name"
    else
        print_error "Docker build failed"
        exit 1
    fi
    
    # Tag as latest for local development
    if [[ "$ENVIRONMENT" == "development" && "$IMAGE_TAG" != "latest" ]]; then
        local latest_tag="$DOCKER_REPOSITORY:latest"
        if [[ -n "$DOCKER_REGISTRY" ]]; then
            latest_tag="$DOCKER_REGISTRY/$latest_tag"
        fi
        docker tag "$image_name" "$latest_tag"
        print_info "Tagged image as: $latest_tag"
    fi
}

# Push Docker image
push_image() {
    if [[ "$PUSH_IMAGE" != "true" || "$SKIP_BUILD" == "true" ]]; then
        print_info "Skipping image push"
        return 0
    fi
    
    if [[ -z "$DOCKER_REGISTRY" ]]; then
        print_warning "No registry specified - skipping push"
        return 0
    fi
    
    local image_name="$DOCKER_REGISTRY/$DOCKER_REPOSITORY:$IMAGE_TAG"
    
    print_info "Pushing Docker image: $image_name"
    
    if docker push "$image_name"; then
        print_success "Docker image pushed successfully"
    else
        print_error "Docker push failed"
        exit 1
    fi
    
    # Push latest tag for development
    if [[ "$ENVIRONMENT" == "development" && "$IMAGE_TAG" != "latest" ]]; then
        local latest_tag="$DOCKER_REGISTRY/$DOCKER_REPOSITORY:latest"
        docker push "$latest_tag"
        print_info "Pushed latest tag: $latest_tag"
    fi
}

# Deploy to Kubernetes
deploy_to_kubernetes() {
    if [[ "$SKIP_DEPLOY" == "true" ]]; then
        print_info "Skipping deployment"
        return 0
    fi
    
    print_info "Deploying to Kubernetes environment: $ENVIRONMENT"
    
    # Construct image name for deployment
    local full_image_name="$DOCKER_REPOSITORY:$IMAGE_TAG"
    if [[ -n "$DOCKER_REGISTRY" ]]; then
        full_image_name="$DOCKER_REGISTRY/$full_image_name"
    fi
    
    # Check if deploy script exists
    local deploy_script="$(dirname "$0")/deploy.sh"
    if [[ ! -f "$deploy_script" ]]; then
        print_error "Deploy script not found: $deploy_script"
        exit 1
    fi
    
    # Run the deploy script
    print_info "Running deployment with image: $full_image_name"
    
    if "$deploy_script" -e "$ENVIRONMENT" -t "$IMAGE_TAG"; then
        print_success "Deployment completed successfully"
    else
        print_error "Deployment failed"
        exit 1
    fi
}

# Run tests before deployment
run_tests() {
    print_info "Running tests..."
    
    if command -v make &> /dev/null; then
        if make test; then
            print_success "Tests passed"
        else
            print_error "Tests failed"
            exit 1
        fi
    else
        print_warning "Make not found - skipping tests"
    fi
}

# Show summary
show_summary() {
    local image_name="$DOCKER_REPOSITORY:$IMAGE_TAG"
    if [[ -n "$DOCKER_REGISTRY" ]]; then
        image_name="$DOCKER_REGISTRY/$image_name"
    fi
    
    print_success "Build and deployment summary:"
    echo "  Environment: $ENVIRONMENT"
    echo "  Image: $image_name"
    echo "  Build: $(if [[ "$SKIP_BUILD" == "true" ]]; then echo "skipped"; else echo "completed"; fi)"
    echo "  Push: $(if [[ "$PUSH_IMAGE" == "true" && "$SKIP_BUILD" != "true" && -n "$DOCKER_REGISTRY" ]]; then echo "completed"; else echo "skipped"; fi)"
    echo "  Deploy: $(if [[ "$SKIP_DEPLOY" == "true" ]]; then echo "skipped"; else echo "completed"; fi)"
}

# Main function
main() {
    parse_args "$@"
    check_prerequisites
    
    # Only run tests for production deployments
    if [[ "$ENVIRONMENT" == "production" ]]; then
        run_tests
    fi
    
    build_image
    push_image
    deploy_to_kubernetes
    show_summary
}

# Run main function
main "$@"