#!/bin/bash

# SLURM Exporter Undeployment Script
# This script removes the SLURM exporter from Kubernetes environments

set -euo pipefail

# Default values
ENVIRONMENT=${ENVIRONMENT:-"development"}
NAMESPACE=${NAMESPACE:-"monitoring"}
HELM_RELEASE_NAME=${HELM_RELEASE_NAME:-"slurm-exporter"}
KUBECTL_CONTEXT=""
DRY_RUN=${DRY_RUN:-"false"}
DELETE_NAMESPACE=${DELETE_NAMESPACE:-"false"}
FORCE=${FORCE:-"false"}

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

Remove SLURM Exporter from Kubernetes using Helm

OPTIONS:
    -e, --environment      Environment to undeploy from [default: development]
    -n, --namespace        Kubernetes namespace [default: monitoring]
    -r, --release-name     Helm release name [default: slurm-exporter]
    -k, --context          Kubectl context to use
    --delete-namespace     Delete the namespace after undeploying
    --dry-run              Perform a dry run without making changes
    --force                Force deletion without confirmation
    -h, --help             Show this help message

EXAMPLES:
    # Remove from development environment
    $0 -e development
    
    # Remove and delete namespace
    $0 -e staging --delete-namespace
    
    # Force removal without confirmation
    $0 -e production --force
    
    # Dry run removal
    $0 --dry-run

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
            --delete-namespace)
                DELETE_NAMESPACE="true"
                shift
                ;;
            --dry-run)
                DRY_RUN="true"
                shift
                ;;
            --force)
                FORCE="true"
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

# Confirm deletion
confirm_deletion() {
    if [[ "$FORCE" == "true" || "$DRY_RUN" == "true" ]]; then
        return 0
    fi
    
    print_warning "This will remove the SLURM Exporter deployment:"
    echo "  Release: $HELM_RELEASE_NAME"
    echo "  Namespace: $NAMESPACE"
    echo "  Environment: $ENVIRONMENT"
    
    if [[ "$DELETE_NAMESPACE" == "true" ]]; then
        print_warning "The namespace '$NAMESPACE' will also be deleted!"
    fi
    
    read -p "Are you sure you want to continue? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Operation cancelled"
        exit 0
    fi
}

# Undeploy using Helm
undeploy() {
    if ! check_release_exists; then
        print_warning "Helm release '$HELM_RELEASE_NAME' not found in namespace '$NAMESPACE'"
        return 0
    fi
    
    print_info "Removing SLURM Exporter deployment..."
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "[DRY RUN] Would execute: helm uninstall $HELM_RELEASE_NAME -n $NAMESPACE"
    else
        if helm uninstall "$HELM_RELEASE_NAME" -n "$NAMESPACE"; then
            print_success "Helm release uninstalled successfully"
        else
            print_error "Failed to uninstall Helm release"
            exit 1
        fi
    fi
}

# Delete namespace if requested
delete_namespace() {
    if [[ "$DELETE_NAMESPACE" != "true" ]]; then
        return 0
    fi
    
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        print_warning "Namespace '$NAMESPACE' does not exist"
        return 0
    fi
    
    print_info "Deleting namespace: $NAMESPACE"
    
    if [[ "$DRY_RUN" == "true" ]]; then
        print_info "[DRY RUN] Would execute: kubectl delete namespace $NAMESPACE"
    else
        if kubectl delete namespace "$NAMESPACE"; then
            print_success "Namespace deleted successfully"
        else
            print_error "Failed to delete namespace"
            exit 1
        fi
    fi
}

# Clean up persistent resources
cleanup_resources() {
    print_info "Checking for persistent resources..."
    
    # Check for PVCs that might not be automatically deleted
    local pvcs
    pvcs=$(kubectl get pvc -n "$NAMESPACE" 2>/dev/null | grep -v "^NAME" | wc -l || echo "0")
    
    if [[ "$pvcs" -gt 0 ]]; then
        print_warning "Found $pvcs persistent volume claims in namespace $NAMESPACE"
        print_info "You may need to manually delete them if they're no longer needed"
    fi
    
    # Check for secrets that might be manually created
    local secrets
    secrets=$(kubectl get secrets -n "$NAMESPACE" | grep -E "(slurm-exporter|tls)" | wc -l || echo "0")
    
    if [[ "$secrets" -gt 0 ]]; then
        print_warning "Found $secrets SLURM exporter related secrets in namespace $NAMESPACE"
        print_info "Manual secrets are not automatically deleted"
    fi
}

# Show post-undeploy information
show_post_undeploy_info() {
    if [[ "$DRY_RUN" == "true" ]]; then
        print_success "Dry run completed successfully!"
        return
    fi
    
    print_success "SLURM Exporter has been removed!"
    
    # Verify removal
    if check_release_exists; then
        print_warning "Helm release still exists - removal may have failed"
    else
        print_success "Helm release successfully removed"
    fi
    
    # Check if any pods are still running
    local running_pods
    running_pods=$(kubectl get pods -n "$NAMESPACE" -l "app.kubernetes.io/instance=$HELM_RELEASE_NAME" 2>/dev/null | grep -v "^NAME" | wc -l || echo "0")
    
    if [[ "$running_pods" -gt 0 ]]; then
        print_warning "$running_pods pods are still running - they should terminate shortly"
    fi
}

# Main function
main() {
    parse_args "$@"
    check_prerequisites
    set_kubectl_context
    confirm_deletion
    undeploy
    cleanup_resources
    delete_namespace
    show_post_undeploy_info
}

# Run main function
main "$@"