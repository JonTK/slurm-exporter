#!/run/current-system/sw/bin/bash

# Comprehensive Integration Test Runner
# Supports both real SLURM cluster and mock server testing

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CLUSTER_HOST="${CLUSTER_HOST:-rocky9.ar.jontk.com}"
SLURM_JWT="${SLURM_JWT:-}"
TEST_MODE="${TEST_MODE:-auto}"  # auto, real, mock
TEST_PORT="10341"
JAEGER_PORT="16686"
PROMETHEUS_PORT="9090"
GRAFANA_PORT="3000"

# Directories
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
TEST_OUTPUT_DIR="${SCRIPT_DIR}/output"
LOGS_DIR="${TEST_OUTPUT_DIR}/logs"

# Create output directories
mkdir -p "${TEST_OUTPUT_DIR}" "${LOGS_DIR}"

log() {
    echo -e "${BLUE}[$(date +'%Y-%m-%d %H:%M:%S')]${NC} $*"
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $*"
}

warning() {
    echo -e "${YELLOW}[WARNING]${NC} $*"
}

error() {
    echo -e "${RED}[ERROR]${NC} $*"
}

cleanup() {
    log "Cleaning up test environment..."
    
    # Stop real cluster containers if running
    if [[ -f "${SCRIPT_DIR}/docker-compose.test.yml" ]]; then
        docker-compose -f "${SCRIPT_DIR}/docker-compose.test.yml" down -v || true
    fi
    
    # Stop mock containers if running
    if [[ -f "${SCRIPT_DIR}/docker-compose.mock.yml" ]]; then
        docker-compose -f "${SCRIPT_DIR}/docker-compose.mock.yml" down -v || true
    fi
}

# Set up signal handlers
trap cleanup EXIT INT TERM

determine_test_mode() {
    log "Determining test mode..."
    
    case "${TEST_MODE}" in
        "real")
            log "Forced to use real cluster mode"
            return 0
            ;;
        "mock")
            log "Forced to use mock server mode"
            return 1
            ;;
        "auto")
            log "Auto-detecting best test mode..."
            ;;
        *)
            error "Invalid TEST_MODE: ${TEST_MODE}. Must be 'auto', 'real', or 'mock'"
            exit 1
            ;;
    esac
    
    # Check if JWT token is provided for real cluster
    if [[ -z "${SLURM_JWT}" ]]; then
        warning "No SLURM_JWT provided, using mock server"
        return 1
    fi
    
    # Test connectivity to real cluster
    log "Testing connectivity to ${CLUSTER_HOST}..."
    if timeout 10 bash -c "</dev/tcp/${CLUSTER_HOST}/6820" 2>/dev/null; then
        log "Real cluster is reachable, testing SLURM API..."
        
        local auth_header="X-SLURM-USER-TOKEN: ${SLURM_JWT}"
        local ping_url="http://${CLUSTER_HOST}:6820/slurm/v0.0.43/ping"
        
        if curl -k -s -H "${auth_header}" "${ping_url}" >/dev/null 2>&1; then
            success "Real cluster SLURM API is accessible"
            return 0
        else
            warning "Real cluster SLURM API is not accessible"
        fi
    else
        warning "Real cluster is not reachable"
    fi
    
    log "Falling back to mock server mode"
    return 1
}

check_prerequisites() {
    log "Checking prerequisites..."
    
    # Check required tools
    local tools=("docker" "docker-compose" "go" "curl" "jq")
    for tool in "${tools[@]}"; do
        if ! command -v "${tool}" >/dev/null 2>&1; then
            error "Required tool '${tool}' is not installed"
            exit 1
        fi
    done
    
    # Check if Python is available for mock server
    if ! command -v python3 >/dev/null 2>&1; then
        warning "Python3 not available, mock server tests may not work"
    fi
    
    success "Prerequisites check completed"
}

build_exporter() {
    log "Building SLURM Exporter..."
    
    cd "${PROJECT_ROOT}"
    
    # Build the binary
    if ! make build; then
        error "Failed to build SLURM Exporter"
        exit 1
    fi
    
    # Build Docker image for testing
    if ! docker build -t slurm-exporter:test .; then
        error "Failed to build Docker image"
        exit 1
    fi
    
    success "SLURM Exporter built successfully"
}

start_real_cluster_tests() {
    log "Starting real cluster test environment..."
    
    export SLURM_JWT="${SLURM_JWT}"
    
    # Use the existing rocky cluster test script
    cd "${SCRIPT_DIR}"
    
    # Start the test environment
    if ! bash ./run-rocky-tests.sh; then
        error "Real cluster tests failed"
        return 1
    fi
    
    success "Real cluster tests completed"
}

start_mock_server_tests() {
    log "Starting mock server test environment..."
    
    cd "${SCRIPT_DIR}"
    
    # Build mock server image if needed
    if ! docker build -f Dockerfile.mock-slurm -t mock-slurm:test .; then
        error "Failed to build mock SLURM server image"
        return 1
    fi
    
    # Start mock environment
    if ! docker-compose -f docker-compose.mock.yml up -d; then
        error "Failed to start mock test environment"
        return 1
    fi
    
    # Wait for services to be healthy
    log "Waiting for mock services to become healthy..."
    
    local services=("mock-slurm" "slurm-exporter" "jaeger" "prometheus" "grafana")
    for service in "${services[@]}"; do
        log "Waiting for ${service} to be healthy..."
        
        local retries=30
        while [[ $retries -gt 0 ]]; do
            if docker-compose -f docker-compose.mock.yml ps "${service}" | grep -q "healthy"; then
                success "${service} is healthy"
                break
            fi
            
            retries=$((retries - 1))
            if [[ $retries -eq 0 ]]; then
                warning "${service} did not become healthy in time"
                docker-compose -f docker-compose.mock.yml logs "${service}"
                return 1
            else
                sleep 5
            fi
        done
    done
    
    success "Mock test environment started successfully"
    
    # Display service URLs
    log "Service URLs:"
    log "  Mock SLURM Server: https://localhost:6820"
    log "  SLURM Exporter: http://localhost:${TEST_PORT}"
    log "  Jaeger UI: http://localhost:${JAEGER_PORT}"
    log "  Prometheus: http://localhost:${PROMETHEUS_PORT}"
    log "  Grafana: http://localhost:${GRAFANA_PORT} (admin/admin)"
}

run_integration_tests() {
    log "Running Go integration tests..."
    
    cd "${PROJECT_ROOT}"
    
    # Set environment variables for tests
    export SLURM_JWT="${SLURM_JWT}"
    export TEST_OUTPUT_DIR="${TEST_OUTPUT_DIR}"
    
    # Run the integration tests
    local test_output="${TEST_OUTPUT_DIR}/integration-test-results.txt"
    
    if go test -v -timeout 15m ./test/integration/... > "${test_output}" 2>&1; then
        success "Integration tests passed"
    else
        error "Integration tests failed"
        echo "Test output:"
        cat "${test_output}"
        return 1
    fi
    
    # Copy any generated test reports
    find . -name "test-report-*.json" -exec cp {} "${TEST_OUTPUT_DIR}/" \; 2>/dev/null || true
}

run_performance_tests() {
    log "Running performance benchmarks..."
    
    cd "${PROJECT_ROOT}"
    
    # Run benchmark tests
    local bench_output="${TEST_OUTPUT_DIR}/benchmark-results.txt"
    
    if go test -v -bench=. -benchmem ./test/integration/... > "${bench_output}" 2>&1; then
        success "Performance benchmarks completed"
        
        # Extract key performance metrics
        grep -E "(Benchmark|ns/op|B/op|allocs/op)" "${bench_output}" || true
    else
        warning "Performance benchmarks had issues"
        cat "${bench_output}"
    fi
}

run_connectivity_tests() {
    log "Running connectivity validation tests..."
    
    local base_url="http://localhost:${TEST_PORT}"
    
    # Test health endpoints
    log "Testing health endpoints..."
    
    if curl -s "${base_url}/health" | jq . > "${TEST_OUTPUT_DIR}/health-check.json"; then
        success "Health endpoint accessible"
    else
        error "Health endpoint failed"
        return 1
    fi
    
    if curl -s "${base_url}/ready" > "${TEST_OUTPUT_DIR}/ready-check.txt"; then
        success "Ready endpoint accessible"
    else
        error "Ready endpoint failed"
        return 1
    fi
    
    # Test metrics endpoint
    log "Testing metrics endpoint..."
    if curl -s "${base_url}/metrics" > "${TEST_OUTPUT_DIR}/metrics-sample.txt"; then
        local metric_count=$(grep -c "^slurm_" "${TEST_OUTPUT_DIR}/metrics-sample.txt" || echo "0")
        success "Metrics endpoint accessible, found ${metric_count} SLURM metrics"
        
        # Save sample metrics
        grep "^slurm_" "${TEST_OUTPUT_DIR}/metrics-sample.txt" | head -20 > "${TEST_OUTPUT_DIR}/sample-metrics.txt"
    else
        error "Metrics endpoint failed"
        return 1
    fi
}

generate_test_report() {
    log "Generating comprehensive test report..."
    
    local report_file="${TEST_OUTPUT_DIR}/integration-test-report.html"
    local test_duration="${1:-unknown}"
    local test_mode="${2:-unknown}"
    
    cat > "${report_file}" << EOF
<!DOCTYPE html>
<html>
<head>
    <title>SLURM Exporter Integration Test Report</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; }
        .header { background: #f0f0f0; padding: 20px; border-radius: 5px; }
        .section { margin: 20px 0; }
        .success { color: #28a745; }
        .warning { color: #ffc107; }
        .error { color: #dc3545; }
        .code { background: #f8f9fa; padding: 10px; border-radius: 3px; font-family: monospace; }
        table { border-collapse: collapse; width: 100%; }
        th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
        th { background-color: #f2f2f2; }
    </style>
</head>
<body>
    <div class="header">
        <h1>SLURM Exporter Integration Test Report</h1>
        <p><strong>Test Mode:</strong> ${test_mode}</p>
        <p><strong>Cluster:</strong> ${CLUSTER_HOST}</p>
        <p><strong>Test Date:</strong> $(date)</p>
        <p><strong>Duration:</strong> ${test_duration}</p>
    </div>
    
    <div class="section">
        <h2>Test Results Summary</h2>
        <ul>
            <li>Health Checks: $(test -f "${TEST_OUTPUT_DIR}/health-check.json" && echo "✓ PASS" || echo "✗ FAIL")</li>
            <li>Metrics Collection: $(test -f "${TEST_OUTPUT_DIR}/metrics-sample.txt" && echo "✓ PASS" || echo "✗ FAIL")</li>
            <li>Integration Tests: $(test -f "${TEST_OUTPUT_DIR}/integration-test-results.txt" && echo "✓ PASS" || echo "✗ FAIL")</li>
            <li>Performance Benchmarks: $(test -f "${TEST_OUTPUT_DIR}/benchmark-results.txt" && echo "✓ PASS" || echo "✗ FAIL")</li>
        </ul>
    </div>
EOF

    # Add sample metrics if available
    if [[ -f "${TEST_OUTPUT_DIR}/sample-metrics.txt" ]]; then
        echo '<div class="section"><h2>Sample SLURM Metrics</h2>' >> "${report_file}"
        echo '<div class="code"><pre>' >> "${report_file}"
        head -20 "${TEST_OUTPUT_DIR}/sample-metrics.txt" >> "${report_file}"
        echo '</pre></div></div>' >> "${report_file}"
    fi
    
    # Add health check results if available
    if [[ -f "${TEST_OUTPUT_DIR}/health-check.json" ]]; then
        echo '<div class="section"><h2>Health Check Results</h2>' >> "${report_file}"
        echo '<div class="code"><pre>' >> "${report_file}"
        jq . "${TEST_OUTPUT_DIR}/health-check.json" 2>/dev/null >> "${report_file}" || cat "${TEST_OUTPUT_DIR}/health-check.json" >> "${report_file}"
        echo '</pre></div></div>' >> "${report_file}"
    fi
    
    echo "</body></html>" >> "${report_file}"
    
    success "Test report generated: ${report_file}"
}

main() {
    local start_time=$(date +%s)
    
    log "Starting SLURM Exporter comprehensive integration tests"
    log "Output directory: ${TEST_OUTPUT_DIR}"
    
    # Check prerequisites
    check_prerequisites
    
    # Build exporter
    build_exporter
    
    # Determine test mode
    local use_real_cluster=false
    if determine_test_mode; then
        use_real_cluster=true
        test_mode="real_cluster"
    else
        use_real_cluster=false
        test_mode="mock_server"
    fi
    
    # Start appropriate test environment
    if [[ "${use_real_cluster}" == "true" ]]; then
        log "Using real SLURM cluster: ${CLUSTER_HOST}"
        start_real_cluster_tests
    else
        log "Using mock SLURM server"
        start_mock_server_tests
        
        # Allow services to stabilize
        sleep 10
        
        # Run connectivity tests
        run_connectivity_tests
        
        # Run integration tests
        run_integration_tests
        
        # Run performance tests
        run_performance_tests
    fi
    
    local end_time=$(date +%s)
    local test_duration=$((end_time - start_time))
    
    # Generate report
    generate_test_report "${test_duration}s" "${test_mode}"
    
    success "Integration tests completed successfully!"
    log "Total test duration: ${test_duration} seconds"
    log "Test mode: ${test_mode}"
    log "Results available in: ${TEST_OUTPUT_DIR}"
    
    # Optionally keep environment running
    if [[ "${KEEP_RUNNING:-false}" == "true" ]]; then
        log "Keeping test environment running (KEEP_RUNNING=true)"
        log "Press Ctrl+C to stop and cleanup"
        
        # Wait for interrupt
        while true; do
            sleep 10
        done
    fi
}

# Handle command line arguments
case "${1:-run}" in
    "run")
        main
        ;;
    "cleanup")
        cleanup
        ;;
    "logs")
        if [[ -f "${SCRIPT_DIR}/docker-compose.mock.yml" ]]; then
            docker-compose -f "${SCRIPT_DIR}/docker-compose.mock.yml" logs -f
        elif [[ -f "${SCRIPT_DIR}/docker-compose.test.yml" ]]; then
            docker-compose -f "${SCRIPT_DIR}/docker-compose.test.yml" logs -f
        else
            error "No active test environment found"
            exit 1
        fi
        ;;
    "mock")
        TEST_MODE="mock" main
        ;;
    "real")
        TEST_MODE="real" main
        ;;
    *)
        echo "Usage: $0 [run|cleanup|logs|mock|real]"
        echo "  run     - Run integration tests (auto-detect mode, default)"
        echo "  cleanup - Clean up test environment"
        echo "  logs    - Show logs from test environment"
        echo "  mock    - Force mock server mode"
        echo "  real    - Force real cluster mode"
        echo ""
        echo "Environment variables:"
        echo "  CLUSTER_HOST   - SLURM cluster hostname (default: rocky.ar.jontk.com)"
        echo "  SLURM_JWT      - JWT token for SLURM authentication"
        echo "  TEST_MODE      - Test mode: auto, real, mock (default: auto)"
        echo "  KEEP_RUNNING   - Keep test environment running after tests (default: false)"
        exit 1
        ;;
esac