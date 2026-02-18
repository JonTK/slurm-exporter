#!/bin/bash

# SLURM Exporter Rollback Script
# This script rolls back the SLURM exporter to a previous deployment version

set -euo pipefail

# Default values
ENVIRONMENT=${ENVIRONMENT:-"development"}
NAMESPACE=${NAMESPACE:-"monitoring"}
HELM_RELEASE_NAME=${HELM_RELEASE_NAME:-"slurm-exporter"}
REVISION=""
DRY_RUN=${DRY_RUN:-"false"}
KUBECTL_CONTEXT=""
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

Rollback SLURM Exporter deployment to a previous version

OPTIONS:
    -e, --environment     Environment to rollback (development|staging|production) [default: development]
    -n, --namespace       Kubernetes namespace [default: monitoring]
    -r, --release-name    Helm release name [default: slurm-exporter]
    -k, --context         Kubectl context to use
    --revision            Specific revision to rollback to (default: previous version)
    --dry-run             Perform a dry run without making changes
    --no-wait             Don't wait for rollout to complete
    -h, --help            Show this help message

EXAMPLES:
    # Rollback to previous version
    $0 -e production

    # Rollback to specific revision
    $0 -e staging --revision 3

    # Dry run rollback
    $0 -e development --dry-run

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
            -k|--context)
                KUBECTL_CONTEXT="$2"
                shift 2
                ;;
            --revision)
                REVISION="$2"
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

# Show release history
show_release_history() {
    if ! check_release_exists; then
        print_error "Helm release '$HELM_RELEASE_NAME' not found in namespace '$NAMESPACE'"
        exit 1
    fi

    print_info "Release history:"
    helm history -n "$NAMESPACE" "$HELM_RELEASE_NAME" || true

    # Get current revision
    local current_revision
    current_revision=$(helm history -n "$NAMESPACE" "$HELM_RELEASE_NAME" --output json | jq -r '.[] | select(.status=="deployed") | .revision' 2>/dev/null || echo "unknown")
    print_info "Current revision: $current_revision"

    # If no specific revision provided, use previous one
    if [[ -z "$REVISION" ]]; then
        if [[ "$current_revision" != "unknown" && "$current_revision" -gt 1 ]]; then
            REVISION=$((current_revision - 1))
            print_info "Will rollback to revision: $REVISION"
        else
            print_error "Cannot determine previous revision. Please specify --revision"
            exit 1
        fi
    fi
}

# Validate revision
validate_revision() {
    if [[ -z "$REVISION" ]]; then
        print_error "No revision specified for rollback"
        exit 1
    fi

    # Check if revision exists
    if ! helm history -n "$NAMESPACE" "$HELM_RELEASE_NAME" | grep -q "^\s*$REVISION\s"; then
        print_error "Revision $REVISION not found in release history"
        print_info "Available revisions:"
        helm history -n "$NAMESPACE" "$HELM_RELEASE_NAME" || true
        exit 1
    fi

    print_info "Rolling back to revision: $REVISION"
}

# Confirm rollback
confirm_rollback() {
    if [[ "$DRY_RUN" == "true" ]]; then
        return 0
    fi

    # For production, always require confirmation
    if [[ "$ENVIRONMENT" == "production" ]]; then
        print_warning "This will rollback the SLURM Exporter in PRODUCTION:"
        echo "  Release: $HELM_RELEASE_NAME"
        echo "  Namespace: $NAMESPACE"
        echo "  Target revision: $REVISION"

        read -p "Are you sure you want to continue? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Rollback cancelled"
            exit 0
        fi
    fi
}

# Perform rollback
rollback() {
    print_info "Rolling back SLURM Exporter..."

    local rollback_cmd="helm rollback $HELM_RELEASE_NAME $REVISION --namespace $NAMESPACE"

    # Environment-specific settings
    case $ENVIRONMENT in
        production)
            rollback_cmd="$rollback_cmd --timeout 600s"
            if [[ "$WAIT_FOR_ROLLOUT" == "true" ]]; then
                rollback_cmd="$rollback_cmd --wait"
            fi
            ;;
        development|staging)
            rollback_cmd="$rollback_cmd --timeout 300s"
            ;;
    esac

    if [[ "$DRY_RUN" == "true" ]]; then
        rollback_cmd="$rollback_cmd --dry-run"
    fi

    print_info "Command: $rollback_cmd"

    if eval "$rollback_cmd"; then
        if [[ "$DRY_RUN" == "false" ]]; then
            print_success "Rollback completed successfully!"
        else
            print_success "Dry run completed successfully!"
        fi
    else
        print_error "Rollback failed!"
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

# Show post-rollback information
show_post_rollback_info() {
    if [[ "$DRY_RUN" == "true" ]]; then
        return
    fi

    print_info "Post-rollback information:"

    # Show updated deployment info
    print_info "Current release status:"
    helm list -n "$NAMESPACE" | grep "^$HELM_RELEASE_NAME" || true

    # Get current image
    local current_image
    current_image=$(kubectl get deployment -n "$NAMESPACE" "$HELM_RELEASE_NAME" -o jsonpath='{.spec.template.spec.containers[0].image}' 2>/dev/null || echo "unknown")
    print_info "Current image: $current_image"

    # Check pod status
    print_info "Pod status:"
    kubectl get pods -n "$NAMESPACE" -l "app.kubernetes.io/instance=$HELM_RELEASE_NAME" || true

    # Show how to access the service
    print_info "Access the SLURM Exporter:"
    echo "  kubectl port-forward -n $NAMESPACE service/$HELM_RELEASE_NAME 10341:10341"
    echo "  curl http://localhost:10341/health"
}

# Main function
main() {
    parse_args "$@"
    check_prerequisites
    set_kubectl_context
    show_release_history
    validate_revision
    confirm_rollback
    rollback
    wait_for_rollout
    show_post_rollback_info
}

# Run main function
main "$@"