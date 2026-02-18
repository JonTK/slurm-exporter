#!/bin/bash

# SLURM Exporter Deployment Script
# This script deploys the SLURM exporter to different environments

set -euo pipefail

# Default values
ENVIRONMENT=${ENVIRONMENT:-"development"}
NAMESPACE=${NAMESPACE:-"monitoring"}
HELM_RELEASE_NAME=${HELM_RELEASE_NAME:-"slurm-exporter"}
HELM_CHART_PATH=${HELM_CHART_PATH:-"./charts/slurm-exporter"}
VALUES_FILE=""
DRY_RUN=${DRY_RUN:-"false"}
KUBECTL_CONTEXT=""
SLURM_BASE_URL=""
IMAGE_TAG="latest"
CREATE_NAMESPACE=${CREATE_NAMESPACE:-"true"}

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

Deploy SLURM Exporter to Kubernetes using Helm

OPTIONS:
    -e, --environment     Environment to deploy to (development|staging|production) [default: development]
    -n, --namespace       Kubernetes namespace [default: monitoring]
    -r, --release-name    Helm release name [default: slurm-exporter]
    -c, --chart-path      Path to Helm chart [default: ./charts/slurm-exporter]
    -f, --values-file     Custom values file (overrides environment defaults)
    -k, --context         Kubectl context to use
    -u, --slurm-url       SLURM REST API base URL
    -t, --image-tag       Docker image tag [default: latest]
    --dry-run             Perform a dry run without making changes
    --no-create-namespace Don't create namespace if it doesn't exist
    -h, --help            Show this help message

EXAMPLES:
    # Deploy to development
    $0 -e development -u "http://slurm-dev:6820"

    # Deploy to production with custom image
    $0 -e production -t "v1.0.0" -u "https://slurm-prod:6820"

    # Deploy with custom values file
    $0 -f custom-values.yaml -k prod-cluster

    # Dry run deployment
    $0 -e staging --dry-run

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
            -u|--slurm-url)
                SLURM_BASE_URL="$2"
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
            --no-create-namespace)
                CREATE_NAMESPACE="false"
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

# Validate environment
validate_environment() {
    case $ENVIRONMENT in
        development|staging|production)
            print_info "Deploying to environment: $ENVIRONMENT"
            ;;
        *)
            print_error "Invalid environment: $ENVIRONMENT"
            print_error "Valid environments: development, staging, production"
            exit 1
            ;;
    esac
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

    # Check if chart exists
    if [[ ! -d "$HELM_CHART_PATH" ]]; then
        print_error "Helm chart not found at: $HELM_CHART_PATH"
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

# Create namespace if it doesn't exist
create_namespace() {
    if [[ "$CREATE_NAMESPACE" == "true" ]]; then
        if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
            print_info "Creating namespace: $NAMESPACE"
            if [[ "$DRY_RUN" == "false" ]]; then
                kubectl create namespace "$NAMESPACE"
            else
                print_info "[DRY RUN] Would create namespace: $NAMESPACE"
            fi
        else
            print_info "Namespace $NAMESPACE already exists"
        fi
    fi
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

# Build helm command
build_helm_command() {
    local helm_cmd="helm"

    if [[ "$DRY_RUN" == "true" ]]; then
        helm_cmd="$helm_cmd upgrade --install --dry-run"
    else
        helm_cmd="$helm_cmd upgrade --install"
    fi

    helm_cmd="$helm_cmd $HELM_RELEASE_NAME $HELM_CHART_PATH"
    helm_cmd="$helm_cmd --namespace $NAMESPACE"
    helm_cmd="$helm_cmd --values $VALUES_FILE"
    helm_cmd="$helm_cmd --set image.tag=$IMAGE_TAG"

    if [[ -n "$SLURM_BASE_URL" ]]; then
        helm_cmd="$helm_cmd --set config.slurm.baseURL=$SLURM_BASE_URL"
    fi

    # Environment-specific settings
    case $ENVIRONMENT in
        production)
            helm_cmd="$helm_cmd --timeout 600s"
            helm_cmd="$helm_cmd --wait"
            ;;
        development)
            helm_cmd="$helm_cmd --timeout 300s"
            ;;
        staging)
            helm_cmd="$helm_cmd --timeout 300s"
            ;;
    esac

    echo "$helm_cmd"
}

# Deploy using Helm
deploy() {
    local helm_cmd
    helm_cmd=$(build_helm_command)

    print_info "Deploying SLURM Exporter..."
    print_info "Command: $helm_cmd"

    if eval "$helm_cmd"; then
        if [[ "$DRY_RUN" == "false" ]]; then
            print_success "Deployment completed successfully!"
            print_info "Release: $HELM_RELEASE_NAME"
            print_info "Namespace: $NAMESPACE"
            print_info "Environment: $ENVIRONMENT"
        else
            print_success "Dry run completed successfully!"
        fi
    else
        print_error "Deployment failed!"
        exit 1
    fi
}

# Show post-deployment information
show_post_deployment_info() {
    if [[ "$DRY_RUN" == "true" ]]; then
        return
    fi

    print_info "Post-deployment information:"

    # Check pod status
    print_info "Checking pod status..."
    kubectl get pods -n "$NAMESPACE" -l "app.kubernetes.io/instance=$HELM_RELEASE_NAME" || true

    # Get service information
    print_info "Service information:"
    kubectl get service -n "$NAMESPACE" "$HELM_RELEASE_NAME" || true

    # Show how to access the exporter
    print_info "To access the SLURM Exporter:"
    echo "  kubectl port-forward -n $NAMESPACE service/$HELM_RELEASE_NAME 10341:10341"
    echo "  curl http://localhost:10341/health"
    echo "  curl http://localhost:10341/metrics"
}

# Main function
main() {
    parse_args "$@"
    validate_environment
    check_prerequisites
    set_kubectl_context
    create_namespace
    set_values_file
    deploy
    show_post_deployment_info
}

# Run main function
main "$@"