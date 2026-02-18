#!/bin/bash

# SLURM Exporter Status Script
# This script checks the status of SLURM exporter deployment

set -euo pipefail

# Default values
ENVIRONMENT=${ENVIRONMENT:-"development"}
NAMESPACE=${NAMESPACE:-"monitoring"}
HELM_RELEASE_NAME=${HELM_RELEASE_NAME:-"slurm-exporter"}
KUBECTL_CONTEXT=""
SHOW_LOGS=${SHOW_LOGS:-"false"}
TAIL_LINES=${TAIL_LINES:-"50"}

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

Check the status of SLURM Exporter deployment

OPTIONS:
    -e, --environment     Environment to check (development|staging|production) [default: development]
    -n, --namespace       Kubernetes namespace [default: monitoring]
    -r, --release-name    Helm release name [default: slurm-exporter]
    -k, --context         Kubectl context to use
    --logs                Show recent logs from the exporter
    --tail                Number of log lines to show [default: 50]
    -h, --help            Show this help message

EXAMPLES:
    # Check status in development
    $0 -e development

    # Check status with logs
    $0 -e production --logs

    # Check specific release
    $0 -r my-slurm-exporter -n monitoring

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
            --logs)
                SHOW_LOGS="true"
                shift
                ;;
            --tail)
                TAIL_LINES="$2"
                shift 2
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
}

# Set kubectl context if provided
set_kubectl_context() {
    if [[ -n "$KUBECTL_CONTEXT" ]]; then
        print_info "Setting kubectl context to: $KUBECTL_CONTEXT"
        kubectl config use-context "$KUBECTL_CONTEXT"
    fi
}

# Check if namespace exists
check_namespace() {
    if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
        print_error "Namespace '$NAMESPACE' does not exist"
        return 1
    fi
    return 0
}

# Check Helm release status
check_helm_status() {
    print_info "=== Helm Release Status ==="

    if helm list -n "$NAMESPACE" | grep -q "^$HELM_RELEASE_NAME"; then
        helm list -n "$NAMESPACE" | grep "^$HELM_RELEASE_NAME"
        print_success "Helm release is deployed"

        # Show release history
        print_info "\nRelease history (last 5):"
        helm history -n "$NAMESPACE" "$HELM_RELEASE_NAME" --max 5 || true

        return 0
    else
        print_error "Helm release '$HELM_RELEASE_NAME' not found in namespace '$NAMESPACE'"
        return 1
    fi
}

# Check deployment status
check_deployment_status() {
    print_info "\n=== Deployment Status ==="

    if kubectl get deployment -n "$NAMESPACE" "$HELM_RELEASE_NAME" &> /dev/null; then
        kubectl get deployment -n "$NAMESPACE" "$HELM_RELEASE_NAME"

        # Check rollout status
        local rollout_status
        rollout_status=$(kubectl rollout status deployment -n "$NAMESPACE" "$HELM_RELEASE_NAME" --timeout=1s 2>/dev/null || echo "not ready")

        if [[ "$rollout_status" == *"successfully rolled out"* ]]; then
            print_success "Deployment is ready"
        else
            print_warning "Deployment rollout is not complete"
        fi

        return 0
    else
        print_error "Deployment '$HELM_RELEASE_NAME' not found"
        return 1
    fi
}

# Check pod status
check_pod_status() {
    print_info "\n=== Pod Status ==="

    local pods
    pods=$(kubectl get pods -n "$NAMESPACE" -l "app.kubernetes.io/instance=$HELM_RELEASE_NAME" 2>/dev/null | grep -v "^NAME" || echo "")

    if [[ -n "$pods" ]]; then
        kubectl get pods -n "$NAMESPACE" -l "app.kubernetes.io/instance=$HELM_RELEASE_NAME"

        # Check if any pods are running
        local running_pods
        running_pods=$(echo "$pods" | grep -c "Running" || echo "0")

        if [[ "$running_pods" -gt 0 ]]; then
            print_success "$running_pods pod(s) running"
        else
            print_warning "No pods are in Running state"
        fi

        # Check for any failed pods
        local failed_pods
        failed_pods=$(echo "$pods" | grep -E "(Failed|Error|CrashLoopBackOff)" | wc -l || echo "0")

        if [[ "$failed_pods" -gt 0 ]]; then
            print_error "$failed_pods pod(s) in failed state"
        fi

        return 0
    else
        print_error "No pods found for release '$HELM_RELEASE_NAME'"
        return 1
    fi
}

# Check service status
check_service_status() {
    print_info "\n=== Service Status ==="

    if kubectl get service -n "$NAMESPACE" "$HELM_RELEASE_NAME" &> /dev/null; then
        kubectl get service -n "$NAMESPACE" "$HELM_RELEASE_NAME"

        # Get service endpoints
        local endpoints
        endpoints=$(kubectl get endpoints -n "$NAMESPACE" "$HELM_RELEASE_NAME" 2>/dev/null | grep -v "^NAME" || echo "")

        if [[ -n "$endpoints" && "$endpoints" != *"<none>"* ]]; then
            print_success "Service has endpoints"
            kubectl get endpoints -n "$NAMESPACE" "$HELM_RELEASE_NAME"
        else
            print_warning "Service has no endpoints"
        fi

        return 0
    else
        print_error "Service '$HELM_RELEASE_NAME' not found"
        return 1
    fi
}

# Check configuration
check_configuration() {
    print_info "\n=== Configuration ==="

    # Check ConfigMap
    local configmap_name
    configmap_name=$(kubectl get configmap -n "$NAMESPACE" | grep "$HELM_RELEASE_NAME" | awk '{print $1}' | head -1 || echo "")

    if [[ -n "$configmap_name" ]]; then
        print_success "ConfigMap found: $configmap_name"

        # Show config summary
        local slurm_url
        slurm_url=$(kubectl get configmap -n "$NAMESPACE" "$configmap_name" -o yaml | grep -E "baseURL|base_url" | head -1 | awk '{print $2}' || echo "not configured")
        print_info "SLURM URL: $slurm_url"
    else
        print_warning "No ConfigMap found"
    fi

    # Check Secrets
    local secrets
    secrets=$(kubectl get secrets -n "$NAMESPACE" | grep "$HELM_RELEASE_NAME" | wc -l || echo "0")

    if [[ "$secrets" -gt 0 ]]; then
        print_info "Found $secrets secret(s)"
        kubectl get secrets -n "$NAMESPACE" | grep "$HELM_RELEASE_NAME" || true
    fi
}

# Test health endpoint
test_health_endpoint() {
    print_info "\n=== Health Check ==="

    # Get a running pod
    local pod_name
    pod_name=$(kubectl get pods -n "$NAMESPACE" -l "app.kubernetes.io/instance=$HELM_RELEASE_NAME" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")

    if [[ -n "$pod_name" ]]; then
        print_info "Testing health endpoint on pod: $pod_name"

        # Test health endpoint
        if kubectl exec -n "$NAMESPACE" "$pod_name" -- wget -q -O - http://localhost:10341/health 2>/dev/null; then
            print_success "Health endpoint is responding"
        else
            print_warning "Health endpoint is not responding"
        fi

        # Test metrics endpoint
        print_info "Testing metrics endpoint..."
        local metrics_count
        metrics_count=$(kubectl exec -n "$NAMESPACE" "$pod_name" -- wget -q -O - http://localhost:10341/metrics 2>/dev/null | grep -c "^slurm_" || echo "0")

        if [[ "$metrics_count" -gt 0 ]]; then
            print_success "Metrics endpoint is working ($metrics_count SLURM metrics found)"
        else
            print_warning "Metrics endpoint has no SLURM metrics"
        fi
    else
        print_warning "No running pods found for health check"
    fi
}

# Show recent logs
show_logs() {
    if [[ "$SHOW_LOGS" != "true" ]]; then
        return 0
    fi

    print_info "\n=== Recent Logs (last $TAIL_LINES lines) ==="

    local pod_name
    pod_name=$(kubectl get pods -n "$NAMESPACE" -l "app.kubernetes.io/instance=$HELM_RELEASE_NAME" -o jsonpath='{.items[0].metadata.name}' 2>/dev/null || echo "")

    if [[ -n "$pod_name" ]]; then
        kubectl logs -n "$NAMESPACE" "$pod_name" --tail="$TAIL_LINES" || print_warning "Could not retrieve logs"
    else
        print_warning "No pods found for log retrieval"
    fi
}

# Show access information
show_access_info() {
    print_info "\n=== Access Information ==="

    local service_type
    service_type=$(kubectl get service -n "$NAMESPACE" "$HELM_RELEASE_NAME" -o jsonpath='{.spec.type}' 2>/dev/null || echo "unknown")

    case $service_type in
        ClusterIP|"")
            print_info "Service type: ClusterIP"
            echo "To access the SLURM Exporter:"
            echo "  kubectl port-forward -n $NAMESPACE service/$HELM_RELEASE_NAME 10341:10341"
            echo "  curl http://localhost:10341/health"
            echo "  curl http://localhost:10341/metrics"
            ;;
        NodePort)
            local node_port
            node_port=$(kubectl get service -n "$NAMESPACE" "$HELM_RELEASE_NAME" -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null || echo "unknown")
            print_info "Service type: NodePort (port: $node_port)"
            echo "Access via any cluster node on port $node_port"
            ;;
        LoadBalancer)
            local external_ip
            external_ip=$(kubectl get service -n "$NAMESPACE" "$HELM_RELEASE_NAME" -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "pending")
            print_info "Service type: LoadBalancer (IP: $external_ip)"
            if [[ "$external_ip" != "pending" && -n "$external_ip" ]]; then
                echo "Access via: http://$external_ip:10341"
            fi
            ;;
    esac
}

# Generate summary
generate_summary() {
    print_info "\n=== Summary ==="

    # Overall status
    local overall_status="Unknown"

    # Check if helm release exists and deployment is ready
    if helm list -n "$NAMESPACE" | grep -q "^$HELM_RELEASE_NAME"; then
        local rollout_status
        rollout_status=$(kubectl rollout status deployment -n "$NAMESPACE" "$HELM_RELEASE_NAME" --timeout=1s 2>/dev/null || echo "not ready")

        if [[ "$rollout_status" == *"successfully rolled out"* ]]; then
            overall_status="Healthy"
        else
            overall_status="Degraded"
        fi
    else
        overall_status="Not Deployed"
    fi

    case $overall_status in
        Healthy)
            print_success "Overall Status: $overall_status"
            ;;
        Degraded)
            print_warning "Overall Status: $overall_status"
            ;;
        *)
            print_error "Overall Status: $overall_status"
            ;;
    esac

    echo "  Environment: $ENVIRONMENT"
    echo "  Namespace: $NAMESPACE"
    echo "  Release: $HELM_RELEASE_NAME"
}

# Main function
main() {
    parse_args "$@"
    check_prerequisites
    set_kubectl_context

    print_info "Checking SLURM Exporter status in environment: $ENVIRONMENT"

    # Only continue if namespace exists
    if ! check_namespace; then
        exit 1
    fi

    # Run all status checks
    check_helm_status
    check_deployment_status
    check_pod_status
    check_service_status
    check_configuration
    test_health_endpoint
    show_logs
    show_access_info
    generate_summary
}

# Run main function
main "$@"