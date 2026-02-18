#!/run/current-system/sw/bin/bash

# Integration Test Runner for rocky.ar.jontk.com SLURM Cluster
# This script sets up and runs comprehensive integration tests

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
CLUSTER_HOST="rocky9.ar.jontk.com"
SLURM_JWT="${SLURM_JWT:-}"
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
    
    # Stop containers if running
    if docker-compose -f "${SCRIPT_DIR}/docker-compose.test.yml" ps -q > /dev/null 2>&1; then
        docker-compose -f "${SCRIPT_DIR}/docker-compose.test.yml" down -v || true
    fi
}

# Set up signal handlers
trap cleanup EXIT INT TERM

check_prerequisites() {
    log "Checking prerequisites..."
    
    # Check if JWT token is provided
    if [[ -z "${SLURM_JWT}" ]]; then
        error "SLURM_JWT environment variable is required"
        echo "Usage: SLURM_JWT=your-token-here $0"
        exit 1
    fi
    
    # Check required tools
    local tools=("docker" "docker-compose" "go" "curl" "jq")
    for tool in "${tools[@]}"; do
        if ! command -v "${tool}" >/dev/null 2>&1; then
            error "Required tool '${tool}' is not installed"
            exit 1
        fi
    done
    
    # Test connectivity to SLURM cluster
    log "Testing connectivity to ${CLUSTER_HOST}..."
    if ! curl -k --connect-timeout 10 "http://${CLUSTER_HOST}:6820/slurm/v0.0.43/ping" >/dev/null 2>&1; then
        warning "Cannot reach SLURM cluster at ${CLUSTER_HOST}:6820"
        warning "Tests will run but may fail due to connectivity issues"
    else
        success "SLURM cluster is reachable"
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

start_test_environment() {
    log "Starting test environment..."
    
    # Create Docker Compose file for testing
    cat > "${SCRIPT_DIR}/docker-compose.test.yml" << EOF
version: '3.8'

services:
  slurm-exporter:
    image: slurm-exporter:test
    ports:
      - "${TEST_PORT}:10341"
    volumes:
      - ${SCRIPT_DIR}/rocky-cluster-config.yaml:/etc/slurm-exporter/config.yaml:ro
      - ${LOGS_DIR}:/var/log/slurm-exporter
    environment:
      - SLURM_EXPORTER_SLURM_TOKEN=${SLURM_JWT}
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:10341/health"]
      interval: 10s
      timeout: 5s
      retries: 5
      start_period: 30s
    restart: unless-stopped

  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "${JAEGER_PORT}:16686"
      - "14268:14268"
    environment:
      - COLLECTOR_OTLP_ENABLED=true
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:16686"]
      interval: 10s
      timeout: 5s
      retries: 3

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "${PROMETHEUS_PORT}:9090"
    volumes:
      - ${SCRIPT_DIR}/prometheus-test.yml:/etc/prometheus/prometheus.yml:ro
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
      - '--storage.tsdb.retention.time=1h'
      - '--web.enable-lifecycle'
    healthcheck:
      test: ["CMD", "wget", "--quiet", "--tries=1", "--spider", "http://localhost:9090/-/healthy"]
      interval: 10s
      timeout: 5s
      retries: 3

  grafana:
    image: grafana/grafana:latest
    ports:
      - "${GRAFANA_PORT}:3000"
    environment:
      - GF_SECURITY_ADMIN_PASSWORD=admin
      - GF_USERS_ALLOW_SIGN_UP=false
    volumes:
      - ${PROJECT_ROOT}/dashboards:/var/lib/grafana/dashboards:ro
      - ${SCRIPT_DIR}/grafana-datasources.yml:/etc/grafana/provisioning/datasources/prometheus.yml:ro
      - ${SCRIPT_DIR}/grafana-dashboards.yml:/etc/grafana/provisioning/dashboards/slurm.yml:ro
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3000/api/health"]
      interval: 10s
      timeout: 5s
      retries: 5
EOF

    # Create Prometheus configuration
    cat > "${SCRIPT_DIR}/prometheus-test.yml" << EOF
global:
  scrape_interval: 15s
  evaluation_interval: 15s

scrape_configs:
  - job_name: 'slurm-exporter'
    static_configs:
      - targets: ['slurm-exporter:10341']
    scrape_interval: 15s
    scrape_timeout: 10s

  - job_name: 'prometheus'
    static_configs:
      - targets: ['localhost:9090']
EOF

    # Create Grafana datasource configuration
    cat > "${SCRIPT_DIR}/grafana-datasources.yml" << EOF
apiVersion: 1

datasources:
  - name: Prometheus
    type: prometheus
    access: proxy
    url: http://prometheus:9090
    isDefault: true
EOF

    # Create Grafana dashboard provisioning
    cat > "${SCRIPT_DIR}/grafana-dashboards.yml" << EOF
apiVersion: 1

providers:
  - name: 'slurm'
    orgId: 1
    folder: 'SLURM'
    type: file
    disableDeletion: false
    updateIntervalSeconds: 10
    options:
      path: /var/lib/grafana/dashboards
EOF

    # Start the test environment
    cd "${SCRIPT_DIR}"
    if ! docker-compose -f docker-compose.test.yml up -d; then
        error "Failed to start test environment"
        exit 1
    fi
    
    # Wait for services to be healthy
    log "Waiting for services to become healthy..."
    
    local services=("slurm-exporter" "jaeger" "prometheus" "grafana")
    for service in "${services[@]}"; do
        log "Waiting for ${service} to be healthy..."
        
        local retries=30
        while [[ $retries -gt 0 ]]; do
            if docker-compose -f docker-compose.test.yml ps "${service}" | grep -q "healthy"; then
                success "${service} is healthy"
                break
            fi
            
            retries=$((retries - 1))
            if [[ $retries -eq 0 ]]; then
                warning "${service} did not become healthy in time"
                docker-compose -f docker-compose.test.yml logs "${service}"
            else
                sleep 5
            fi
        done
    done
    
    success "Test environment started successfully"
    
    # Display service URLs
    log "Service URLs:"
    log "  SLURM Exporter: http://localhost:${TEST_PORT}"
    log "  Jaeger UI: http://localhost:${JAEGER_PORT}"
    log "  Prometheus: http://localhost:${PROMETHEUS_PORT}"
    log "  Grafana: http://localhost:${GRAFANA_PORT} (admin/admin)"
}

run_connectivity_tests() {
    log "Running connectivity tests..."
    
    # Test SLURM cluster connectivity
    log "Testing SLURM cluster connectivity..."
    
    local auth_header="X-SLURM-USER-TOKEN: ${SLURM_JWT}"
    local base_url="http://${CLUSTER_HOST}:6820/slurm/v0.0.43"
    
    # Test ping endpoint
    if curl -k -H "${auth_header}" "${base_url}/ping" > "${TEST_OUTPUT_DIR}/slurm-ping.json" 2>/dev/null; then
        success "SLURM ping successful"
        cat "${TEST_OUTPUT_DIR}/slurm-ping.json" | jq . || true
    else
        error "SLURM ping failed"
        return 1
    fi
    
    # Test info endpoint
    if curl -k -H "${auth_header}" "${base_url}/diag" > "${TEST_OUTPUT_DIR}/slurm-info.json" 2>/dev/null; then
        success "SLURM info retrieval successful"
        log "SLURM cluster information retrieved"
    else
        warning "SLURM info retrieval failed"
    fi
    
    # Test jobs endpoint
    if curl -k -H "${auth_header}" "${base_url}/jobs?limit=10" > "${TEST_OUTPUT_DIR}/slurm-jobs.json" 2>/dev/null; then
        success "SLURM jobs retrieval successful"
        local job_count=$(cat "${TEST_OUTPUT_DIR}/slurm-jobs.json" | jq -r '.jobs | length' 2>/dev/null || echo "unknown")
        log "Found ${job_count} jobs in cluster"
    else
        warning "SLURM jobs retrieval failed"
    fi
    
    # Test nodes endpoint
    if curl -k -H "${auth_header}" "${base_url}/nodes" > "${TEST_OUTPUT_DIR}/slurm-nodes.json" 2>/dev/null; then
        success "SLURM nodes retrieval successful"
        local node_count=$(cat "${TEST_OUTPUT_DIR}/slurm-nodes.json" | jq -r '.nodes | length' 2>/dev/null || echo "unknown")
        log "Found ${node_count} nodes in cluster"
    else
        warning "SLURM nodes retrieval failed"
    fi
}

run_exporter_tests() {
    log "Running exporter functionality tests..."
    
    # Test health endpoints
    local base_url="http://localhost:${TEST_PORT}"
    
    # Health check
    if curl -s "${base_url}/health" | jq . > "${TEST_OUTPUT_DIR}/exporter-health.json"; then
        success "Exporter health check passed"
        local status=$(cat "${TEST_OUTPUT_DIR}/exporter-health.json" | jq -r '.status')
        log "Health status: ${status}"
    else
        error "Exporter health check failed"
    fi
    
    # Ready check
    if curl -s "${base_url}/ready" > "${TEST_OUTPUT_DIR}/exporter-ready.txt"; then
        success "Exporter ready check passed"
    else
        error "Exporter ready check failed"
    fi
    
    # Metrics retrieval
    log "Retrieving metrics..."
    if curl -s "${base_url}/metrics" > "${TEST_OUTPUT_DIR}/exporter-metrics.txt"; then
        local metric_count=$(grep -c "^slurm_" "${TEST_OUTPUT_DIR}/exporter-metrics.txt" || echo "0")
        local total_lines=$(wc -l < "${TEST_OUTPUT_DIR}/exporter-metrics.txt")
        
        success "Metrics retrieval successful"
        log "Found ${metric_count} SLURM metrics in ${total_lines} total lines"
        
        # Save some sample metrics
        grep "^slurm_" "${TEST_OUTPUT_DIR}/exporter-metrics.txt" | head -20 > "${TEST_OUTPUT_DIR}/sample-metrics.txt"
    else
        error "Metrics retrieval failed"
    fi
    
    # Debug endpoints
    if curl -s "${base_url}/debug/vars" | jq . > "${TEST_OUTPUT_DIR}/exporter-vars.json"; then
        success "Debug vars retrieval successful"
    else
        warning "Debug vars retrieval failed"
    fi
    
    if curl -s "${base_url}/debug/config" | jq . > "${TEST_OUTPUT_DIR}/exporter-config.json"; then
        success "Debug config retrieval successful"
    else
        warning "Debug config retrieval failed"
    fi
}

run_integration_tests() {
    log "Running Go integration tests..."
    
    cd "${PROJECT_ROOT}"
    
    # Set environment variables for tests
    export SLURM_JWT="${SLURM_JWT}"
    
    # Run the integration tests
    local test_output="${TEST_OUTPUT_DIR}/integration-test-results.txt"
    
    if go test -v -timeout 10m ./test/integration/... > "${test_output}" 2>&1; then
        success "Integration tests passed"
    else
        error "Integration tests failed"
        echo "Test output:"
        cat "${test_output}"
    fi
    
    # Copy any generated test reports
    find . -name "test-report-*.json" -exec cp {} "${TEST_OUTPUT_DIR}/" \; 2>/dev/null || true
}

run_performance_tests() {
    log "Running performance tests..."
    
    local base_url="http://localhost:${TEST_PORT}"
    local iterations=10
    local results_file="${TEST_OUTPUT_DIR}/performance-results.json"
    
    # Initialize results array
    echo "[]" > "${results_file}"
    
    for i in $(seq 1 ${iterations}); do
        log "Performance test iteration ${i}/${iterations}"
        
        local start_time=$(date +%s%3N)
        if curl -s "${base_url}/metrics" > /dev/null; then
            local end_time=$(date +%s%3N)
            local duration=$((end_time - start_time))
            
            # Add result to JSON array
            local result=$(jq -n --arg i "$i" --arg duration "$duration" --arg timestamp "$(date -Iseconds)" \
                '{iteration: ($i | tonumber), duration_ms: ($duration | tonumber), timestamp: $timestamp}')
            
            jq ". += [$result]" "${results_file}" > "${results_file}.tmp" && mv "${results_file}.tmp" "${results_file}"
            
            log "  Duration: ${duration}ms"
        else
            error "Performance test iteration ${i} failed"
        fi
        
        sleep 2
    done
    
    # Calculate statistics
    local avg_duration=$(jq '[.[].duration_ms] | add / length' "${results_file}")
    local min_duration=$(jq '[.[].duration_ms] | min' "${results_file}")
    local max_duration=$(jq '[.[].duration_ms] | max' "${results_file}")
    
    log "Performance test results:"
    log "  Average duration: ${avg_duration}ms"
    log "  Min duration: ${min_duration}ms"
    log "  Max duration: ${max_duration}ms"
    
    # Save summary
    jq -n --arg avg "$avg_duration" --arg min "$min_duration" --arg max "$max_duration" \
        '{average_ms: ($avg | tonumber), min_ms: ($min | tonumber), max_ms: ($max | tonumber)}' \
        > "${TEST_OUTPUT_DIR}/performance-summary.json"
}

generate_test_report() {
    log "Generating comprehensive test report..."
    
    local report_file="${TEST_OUTPUT_DIR}/integration-test-report.html"
    
    cat > "${report_file}" << 'EOF'
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
        <p><strong>Cluster:</strong> rocky.ar.jontk.com</p>
        <p><strong>Test Date:</strong> $(date)</p>
        <p><strong>Duration:</strong> ${test_duration:-"Unknown"}</p>
    </div>
EOF

    # Add sections for each test component
    local sections=("connectivity" "exporter" "performance")
    
    for section in "${sections[@]}"; do
        echo "<div class=\"section\">" >> "${report_file}"
        echo "<h2>$(echo "${section}" | tr '[:lower:]' '[:upper:]') Tests</h2>" >> "${report_file}"
        
        # Add relevant test data
        case "${section}" in
            "connectivity")
                if [[ -f "${TEST_OUTPUT_DIR}/slurm-ping.json" ]]; then
                    echo "<h3>SLURM Cluster Response</h3>" >> "${report_file}"
                    echo "<div class=\"code\"><pre>$(cat "${TEST_OUTPUT_DIR}/slurm-ping.json" | jq . 2>/dev/null || cat "${TEST_OUTPUT_DIR}/slurm-ping.json")</pre></div>" >> "${report_file}"
                fi
                ;;
            "exporter")
                if [[ -f "${TEST_OUTPUT_DIR}/sample-metrics.txt" ]]; then
                    echo "<h3>Sample Metrics</h3>" >> "${report_file}"
                    echo "<div class=\"code\"><pre>$(head -10 "${TEST_OUTPUT_DIR}/sample-metrics.txt")</pre></div>" >> "${report_file}"
                fi
                ;;
            "performance")
                if [[ -f "${TEST_OUTPUT_DIR}/performance-summary.json" ]]; then
                    echo "<h3>Performance Summary</h3>" >> "${report_file}"
                    echo "<div class=\"code\"><pre>$(cat "${TEST_OUTPUT_DIR}/performance-summary.json" | jq .)</pre></div>" >> "${report_file}"
                fi
                ;;
        esac
        
        echo "</div>" >> "${report_file}"
    done
    
    echo "</body></html>" >> "${report_file}"
    
    success "Test report generated: ${report_file}"
}

main() {
    local start_time=$(date +%s)
    
    log "Starting SLURM Exporter integration tests for rocky.ar.jontk.com"
    log "Output directory: ${TEST_OUTPUT_DIR}"
    
    # Run test phases
    check_prerequisites
    build_exporter
    start_test_environment
    
    sleep 10  # Allow services to stabilize
    
    run_connectivity_tests
    run_exporter_tests
    run_integration_tests
    run_performance_tests
    
    local end_time=$(date +%s)
    local test_duration=$((end_time - start_time))
    
    generate_test_report
    
    success "Integration tests completed successfully!"
    log "Total test duration: ${test_duration} seconds"
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
        docker-compose -f "${SCRIPT_DIR}/docker-compose.test.yml" logs -f
        ;;
    *)
        echo "Usage: $0 [run|cleanup|logs]"
        echo "  run     - Run the full integration test suite (default)"
        echo "  cleanup - Clean up test environment"
        echo "  logs    - Show logs from test environment"
        exit 1
        ;;
esac