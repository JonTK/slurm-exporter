#!/bin/bash

# SLURM Exporter Upgrade Script
# This script upgrades an existing SLURM exporter deployment

set -euo pipefail

# Default values
ENVIRONMENT=${ENVIRONMENT:-"development"}
NAMESPACE=${NAMESPACE:-"monitoring"}
HELM_RELEASE_NAME=${HELM_RELEASE_NAME:-"slurm-exporter"}
HELM_CHART_PATH=${HELM_CHART_PATH:-"./charts/slurm-exporter"}
VALUES_FILE=""
DRY_RUN=${DRY_RUN:-"false"}
KUBECTL_CONTEXT=""
IMAGE_TAG=""
WAIT_FOR_ROLLOUT=${WAIT_FOR_ROLLOUT:-"true"}

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

Upgrade existing SLURM Exporter deployment in Kubernetes

OPTIONS:
    -e, --environment     Environment to upgrade (development|staging|production) [default: development]
    -n, --namespace       Kubernetes namespace [default: monitoring]
    -r, --release-name    Helm release name [default: slurm-exporter]
    -c, --chart-path      Path to Helm chart [default: ./charts/slurm-exporter]
    -f, --values-file     Custom values file (overrides environment defaults)
    -k, --context         Kubectl context to use
    -t, --image-tag       Docker image tag to upgrade to
    --dry-run             Perform a dry run without making changes
    --no-wait             Don't wait for rollout to complete
    -h, --help            Show this help message

EXAMPLES:
    # Upgrade to new image tag
    $0 -e production -t "v1.2.0"
    
    # Upgrade with custom values
    $0 -f custom-values.yaml -t "latest"
    
    # Dry run upgrade
    $0 -e staging --dry-run -t "v1.1.0"

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
            -n|--namespace)
                NAMESPACE="$2"
                shift 2
                ;;
            -r|--release-name)
                HELM_RELEASE_NAME="$2"
                shift 2
                ;;
            -c|--chart-path)
                HELM_CHART_PATH="$2"
                shift 2
                ;;
            -f|--values-file)
                VALUES_FILE="$2"
                shift 2
                ;;
            -k|--context)
                KUBECTL_CONTEXT="$2"
                shift 2
                ;;
            -t|--image-tag)
                IMAGE_TAG="$2"
                shift 2
                ;;
            --dry-run)
                DRY_RUN="true"
                shift
                ;;
            --no-wait)
                WAIT_FOR_ROLLOUT="false"
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
    
    # Check if kubectl is installed
    if ! command -v kubectl &> /dev/null; then
        print_error "kubectl is not installed or not in PATH"
        exit 1
    fi
    
    # Check if helm is installed
    if ! command -v helm &> /dev/null; then
        print_error "helm is not installed or not in PATH"
        exit 1
    fi
    
    print_success "Prerequisites check passed"
}

# Set kubectl context if provided
set_kubectl_context() {
    if [[ -n "$KUBECTL_CONTEXT" ]]; then
        print_info "Setting kubectl context to: $KUBECTL_CONTEXT"
        kubectl config use-context "$KUBECTL_CONTEXT"
    fi
}

# Check if release exists
check_release_exists() {
    if helm list -n "$NAMESPACE" | grep -q "^$HELM_RELEASE_NAME"; then
        return 0
    else
        return 1
    fi
}

# Get current deployment info
get_current_info() {
    if ! check_release_exists; then
        print_error "Helm release '$HELM_RELEASE_NAME' not found in namespace '$NAMESPACE'"
        print_info "Run the deploy script first to create the initial deployment"
        exit 1
    fi
    
    print_info "Current deployment information:"
    helm list -n "$NAMESPACE" | grep "^$HELM_RELEASE_NAME" || true
    
    # Get current image
    local current_image
    current_image=$(kubectl get deployment -n "$NAMESPACE" "$HELM_RELEASE_NAME" -o jsonpath='{.spec.template.spec.containers[0].image}' 2>/dev/null || echo "unknown")
    print_info "Current image: $current_image"
}

# Set default values file based on environment
set_values_file() {
    if [[ -z "$VALUES_FILE" ]]; then
        case $ENVIRONMENT in
            development)
                VALUES_FILE="$HELM_CHART_PATH/values-development.yaml"
                ;;
            production)
                VALUES_FILE="$HELM_CHART_PATH/values-production.yaml"
                ;;
            staging)
                VALUES_FILE="$HELM_CHART_PATH/values.yaml"
                ;;
        esac
    fi
    
    if [[ ! -f "$VALUES_FILE" ]]; then
        print_error "Values file not found: $VALUES_FILE"
        exit 1
    fi
    
    print_info "Using values file: $VALUES_FILE"
}

# Build helm upgrade command
build_helm_command() {
    local helm_cmd="helm upgrade"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        helm_cmd="$helm_cmd --dry-run"
    fi
    
    helm_cmd="$helm_cmd $HELM_RELEASE_NAME $HELM_CHART_PATH"
    helm_cmd="$helm_cmd --namespace $NAMESPACE"
    helm_cmd="$helm_cmd --values $VALUES_FILE"
    
    if [[ -n "$IMAGE_TAG" ]]; then
        helm_cmd="$helm_cmd --set image.tag=$IMAGE_TAG"
    fi
    
    # Environment-specific settings
    case $ENVIRONMENT in
        production)
            helm_cmd="$helm_cmd --timeout 600s"
            if [[ "$WAIT_FOR_ROLLOUT" == "true" ]]; then
                helm_cmd="$helm_cmd --wait"
            fi
            ;;
        development|staging)
            helm_cmd="$helm_cmd --timeout 300s"
            ;;
    esac
    
    echo "$helm_cmd"
}

# Upgrade deployment
upgrade() {
    local helm_cmd
    helm_cmd=$(build_helm_command)
    
    print_info "Upgrading SLURM Exporter..."
    print_info "Command: $helm_cmd"
    
    if eval "$helm_cmd"; then
        if [[ "$DRY_RUN" == "false" ]]; then
            print_success "Upgrade completed successfully!"
        else
            print_success "Dry run completed successfully!"
        fi
    else
        print_error "Upgrade failed!"
        exit 1
    fi
}

# Wait for rollout
wait_for_rollout() {
    if [[ "$DRY_RUN" == "true" || "$WAIT_FOR_ROLLOUT" == "false" ]]; then
        return 0
    fi
    
    print_info "Waiting for rollout to complete..."
    
    if kubectl rollout status deployment -n "$NAMESPACE" "$HELM_RELEASE_NAME" --timeout=300s; then
        print_success "Rollout completed successfully"
    else
        print_error "Rollout failed or timed out"
        print_info "Check pod status:"
        kubectl get pods -n "$NAMESPACE" -l "app.kubernetes.io/instance=$HELM_RELEASE_NAME" || true
        exit 1
    fi
}

# Show post-upgrade information
show_post_upgrade_info() {
    if [[ "$DRY_RUN" == "true" ]]; then
        return
    fi
    
    print_info "Post-upgrade information:"
    
    # Show updated deployment info
    print_info "Updated deployment:"
    helm list -n "$NAMESPACE" | grep "^$HELM_RELEASE_NAME" || true
    
    # Get new image
    local new_image
    new_image=$(kubectl get deployment -n "$NAMESPACE" "$HELM_RELEASE_NAME" -o jsonpath='{.spec.template.spec.containers[0].image}' 2>/dev/null || echo "unknown")
    print_info "New image: $new_image"
    
    # Check pod status
    print_info "Pod status:"
    kubectl get pods -n "$NAMESPACE" -l "app.kubernetes.io/instance=$HELM_RELEASE_NAME" || true
}

# Main function
main() {
    parse_args "$@"
    check_prerequisites
    set_kubectl_context
    get_current_info
    set_values_file
    upgrade
    wait_for_rollout
    show_post_upgrade_info
}

# Run main function
main "$@"