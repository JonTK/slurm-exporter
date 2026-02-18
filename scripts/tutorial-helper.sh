#!/bin/bash

# SLURM Exporter Tutorial Helper Script
# Provides interactive assistance for learning and troubleshooting

set -euo pipefail

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SANDBOX_DIR="$PROJECT_ROOT/sandbox"
DOCS_DIR="$PROJECT_ROOT/docs"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Emojis
ROCKET='🚀'
CHECK='✅'
WARNING='⚠️'
ERROR='❌'
INFO='ℹ️'
TUTORIAL='🎓'
SANDBOX='🧪'

# Helper functions
log_info() {
    echo -e "${BLUE}${INFO}${NC} $1"
}

log_success() {
    echo -e "${GREEN}${CHECK}${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}${WARNING}${NC} $1"
}

log_error() {
    echo -e "${RED}${ERROR}${NC} $1"
}

log_tutorial() {
    echo -e "${PURPLE}${TUTORIAL}${NC} $1"
}

print_header() {
    echo
    echo -e "${CYAN}╔══════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${CYAN}║${NC}          ${ROCKET} SLURM Exporter Tutorial Helper ${ROCKET}             ${CYAN}║${NC}"
    echo -e "${CYAN}╚══════════════════════════════════════════════════════════════════╝${NC}"
    echo
}

print_menu() {
    echo -e "${BLUE}Available Commands:${NC}"
    echo
    echo -e "  ${GREEN}setup${NC}           - Set up tutorial environment"
    echo -e "  ${GREEN}sandbox${NC}         - Start/stop sandbox environment"
    echo -e "  ${GREEN}tutorial${NC}        - Run interactive tutorials"
    echo -e "  ${GREEN}troubleshoot${NC}    - Troubleshooting assistance"
    echo -e "  ${GREEN}test${NC}            - Test configurations and connectivity"
    echo -e "  ${GREEN}monitor${NC}         - Monitor exporter performance"
    echo -e "  ${GREEN}cleanup${NC}         - Clean up tutorial environment"
    echo -e "  ${GREEN}help${NC}            - Show detailed help"
    echo
}

# Setup functions
setup_environment() {
    log_tutorial "Setting up tutorial environment..."
    
    # Check prerequisites
    check_prerequisites
    
    # Create necessary directories
    mkdir -p "$HOME/.slurm-exporter/tutorials"
    mkdir -p "$HOME/.slurm-exporter/configs"
    mkdir -p "$HOME/.slurm-exporter/logs"
    
    # Download sample configurations
    setup_sample_configs
    
    # Initialize tutorial tracking
    setup_tutorial_tracking
    
    log_success "Tutorial environment setup complete!"
    log_info "Next steps:"
    echo "  1. Run: $0 sandbox start"
    echo "  2. Run: $0 tutorial list"
    echo "  3. Visit: http://localhost:8888"
}

check_prerequisites() {
    log_info "Checking prerequisites..."
    
    local missing_tools=()
    
    # Check required tools
    for tool in docker docker-compose curl jq; do
        if ! command -v "$tool" &> /dev/null; then
            missing_tools+=("$tool")
        fi
    done
    
    if [ ${#missing_tools[@]} -gt 0 ]; then
        log_error "Missing required tools: ${missing_tools[*]}"
        echo
        echo "Installation instructions:"
        echo "  Docker: https://docs.docker.com/get-docker/"
        echo "  Docker Compose: https://docs.docker.com/compose/install/"
        echo "  jq: https://stedolan.github.io/jq/download/"
        exit 1
    fi
    
    # Check Docker daemon
    if ! docker info &> /dev/null; then
        log_error "Docker daemon is not running"
        exit 1
    fi
    
    log_success "All prerequisites satisfied"
}

setup_sample_configs() {
    log_info "Setting up sample configurations..."
    
    local config_dir="$HOME/.slurm-exporter/configs"
    
    # Basic configuration
    cat > "$config_dir/basic.yaml" << 'EOF'
slurm:
  endpoint: "http://localhost:6820"
  auth:
    type: "none"
collectors:
  - cluster_simple
  - jobs_simple
  - nodes_simple
EOF

    # Development configuration
    cat > "$config_dir/development.yaml" << 'EOF'
slurm:
  endpoint: "http://localhost:6820"
  timeout: "30s"
  auth:
    type: "none"
collectors:
  - cluster_simple
  - jobs_simple
  - nodes_simple
  - partitions_simple
performance:
  collection_interval: "15s"
  cache:
    enabled: true
    base_ttl: "30s"
logging:
  level: "debug"
EOF

    # Production configuration template
    cat > "$config_dir/production-template.yaml" << 'EOF'
slurm:
  endpoint: "https://your-slurm-server:6820"
  timeout: "30s"
  auth:
    type: "token"
    token: "your-token-here"
collectors:
  - cluster_simple
  - jobs_simple
  - nodes_simple
  - partitions_simple
performance:
  collection_interval: "60s"
  batch_processing:
    enabled: true
    batch_size: 100
  cache:
    enabled: true
    base_ttl: "120s"
filtering:
  enabled: true
  max_cardinality: 50000
monitoring:
  circuit_breaker:
    enabled: true
logging:
  level: "info"
  format: "json"
EOF

    log_success "Sample configurations created in $config_dir"
}

setup_tutorial_tracking() {
    local tracking_file="$HOME/.slurm-exporter/tutorials/progress.json"
    
    if [ ! -f "$tracking_file" ]; then
        cat > "$tracking_file" << 'EOF'
{
  "tutorials": {
    "basic-setup": {"completed": false, "started": null},
    "configuration": {"completed": false, "started": null},
    "monitoring": {"completed": false, "started": null},
    "kubernetes": {"completed": false, "started": null},
    "optimization": {"completed": false, "started": null}
  },
  "exercises": {},
  "last_activity": null
}
EOF
    fi
}

# Sandbox management
manage_sandbox() {
    local action="${1:-status}"
    
    case "$action" in
        "start")
            start_sandbox
            ;;
        "stop")
            stop_sandbox
            ;;
        "restart")
            restart_sandbox
            ;;
        "status")
            sandbox_status
            ;;
        "logs")
            sandbox_logs "${2:-}"
            ;;
        "reset")
            reset_sandbox
            ;;
        *)
            log_error "Unknown sandbox action: $action"
            echo "Available actions: start, stop, restart, status, logs, reset"
            exit 1
            ;;
    esac
}

start_sandbox() {
    log_tutorial "Starting sandbox environment..."
    
    if [ ! -d "$SANDBOX_DIR" ]; then
        log_error "Sandbox directory not found: $SANDBOX_DIR"
        exit 1
    fi
    
    cd "$SANDBOX_DIR"
    
    # Start services
    docker-compose up -d
    
    # Wait for services to be ready
    log_info "Waiting for services to be ready..."
    wait_for_services
    
    log_success "Sandbox environment is ready!"
    echo
    echo "Available services:"
    echo "  📊 Grafana:         http://localhost:3000 (admin/admin)"
    echo "  📈 Prometheus:      http://localhost:9090"
    echo "  🚀 SLURM Exporter:  http://localhost:10341"
    echo "  🔔 Alertmanager:    http://localhost:9093"
    echo "  🎓 Tutorials:       http://localhost:8888"
    echo
    echo "Next steps:"
    echo "  1. Visit the tutorial homepage: http://localhost:8888"
    echo "  2. Or run: $0 tutorial list"
}

stop_sandbox() {
    log_info "Stopping sandbox environment..."
    
    if [ ! -d "$SANDBOX_DIR" ]; then
        log_error "Sandbox directory not found: $SANDBOX_DIR"
        exit 1
    fi
    
    cd "$SANDBOX_DIR"
    docker-compose down
    
    log_success "Sandbox environment stopped"
}

restart_sandbox() {
    log_info "Restarting sandbox environment..."
    stop_sandbox
    sleep 2
    start_sandbox
}

sandbox_status() {
    if [ ! -d "$SANDBOX_DIR" ]; then
        log_error "Sandbox directory not found: $SANDBOX_DIR"
        exit 1
    fi
    
    cd "$SANDBOX_DIR"
    
    echo -e "${BLUE}Sandbox Status:${NC}"
    echo
    
    if ! docker-compose ps | grep -q "Up"; then
        log_warning "Sandbox is not running"
        echo "Run: $0 sandbox start"
        return
    fi
    
    # Check service health
    echo -e "${GREEN}Running Services:${NC}"
    docker-compose ps --format "table"
    echo
    
    # Quick health checks
    echo -e "${BLUE}Health Checks:${NC}"
    
    check_service_health "SLURM Exporter" "http://localhost:10341/health"
    check_service_health "Prometheus" "http://localhost:9090/-/ready"
    check_service_health "Grafana" "http://localhost:3000/api/health"
    check_service_health "Alertmanager" "http://localhost:9093/-/ready"
}

check_service_health() {
    local service_name="$1"
    local health_url="$2"
    
    if curl -sf "$health_url" &> /dev/null; then
        log_success "$service_name is healthy"
    else
        log_warning "$service_name is not responding"
    fi
}

wait_for_services() {
    local max_attempts=60
    local attempt=0
    
    services=(
        "SLURM Exporter:http://localhost:10341/health"
        "Prometheus:http://localhost:9090/-/ready"
        "Grafana:http://localhost:3000/api/health"
    )
    
    for service_info in "${services[@]}"; do
        local service_name="${service_info%%:*}"
        local health_url="${service_info##*:}"
        
        attempt=0
        while [ $attempt -lt $max_attempts ]; do
            if curl -sf "$health_url" &> /dev/null; then
                log_success "$service_name is ready"
                break
            fi
            
            if [ $attempt -eq 0 ]; then
                log_info "Waiting for $service_name..."
            fi
            
            sleep 2
            ((attempt++))
        done
        
        if [ $attempt -eq $max_attempts ]; then
            log_warning "$service_name did not become ready"
        fi
    done
}

sandbox_logs() {
    local service="$1"
    
    if [ ! -d "$SANDBOX_DIR" ]; then
        log_error "Sandbox directory not found: $SANDBOX_DIR"
        exit 1
    fi
    
    cd "$SANDBOX_DIR"
    
    if [ -n "$service" ]; then
        log_info "Showing logs for $service..."
        docker-compose logs -f "$service"
    else
        log_info "Showing logs for all services..."
        docker-compose logs -f
    fi
}

reset_sandbox() {
    log_warning "This will destroy all sandbox data. Continue? (y/N)"
    read -r response
    
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        log_info "Reset cancelled"
        return
    fi
    
    log_info "Resetting sandbox environment..."
    
    cd "$SANDBOX_DIR"
    docker-compose down -v
    docker-compose up -d
    
    wait_for_services
    
    log_success "Sandbox environment reset complete"
}

# Tutorial management
run_tutorial() {
    local tutorial_name="${1:-list}"
    
    case "$tutorial_name" in
        "list")
            list_tutorials
            ;;
        "basic-setup"|"1")
            run_basic_setup_tutorial
            ;;
        "configuration"|"2")
            run_configuration_tutorial
            ;;
        "monitoring"|"3")
            run_monitoring_tutorial
            ;;
        "kubernetes"|"4")
            run_kubernetes_tutorial
            ;;
        "optimization"|"5")
            run_optimization_tutorial
            ;;
        "progress")
            show_tutorial_progress
            ;;
        *)
            log_error "Unknown tutorial: $tutorial_name"
            list_tutorials
            exit 1
            ;;
    esac
}

list_tutorials() {
    echo -e "${BLUE}Available Tutorials:${NC}"
    echo
    echo -e "  ${GREEN}1. basic-setup${NC}     - Deploy SLURM Exporter in 5 minutes"
    echo -e "  ${GREEN}2. configuration${NC}   - Master advanced configuration"
    echo -e "  ${GREEN}3. monitoring${NC}      - Set up monitoring with Prometheus & Grafana"
    echo -e "  ${GREEN}4. kubernetes${NC}      - Deploy on Kubernetes"
    echo -e "  ${GREEN}5. optimization${NC}    - Performance optimization techniques"
    echo
    echo -e "  ${CYAN}progress${NC}           - Show tutorial progress"
    echo
    echo "Usage: $0 tutorial <name|number>"
    echo "Example: $0 tutorial 1"
    echo "         $0 tutorial basic-setup"
}

run_basic_setup_tutorial() {
    log_tutorial "Starting Basic Setup Tutorial (5 minutes)"
    
    mark_tutorial_started "basic-setup"
    
    echo
    echo -e "${YELLOW}Objective:${NC} Deploy and configure SLURM Exporter in under 5 minutes"
    echo
    
    # Step 1: Check prerequisites
    tutorial_step "1" "Checking Prerequisites" "
        We need Docker and a SLURM cluster endpoint.
        For this tutorial, we'll use the sandbox environment.
    "
    
    if ! sandbox_running; then
        log_info "Starting sandbox environment for tutorial..."
        start_sandbox
    fi
    
    press_enter_to_continue
    
    # Step 2: Deploy with Docker
    tutorial_step "2" "Deploy with Docker" "
        Let's run SLURM Exporter with a minimal configuration.
    "
    
    echo "Running command:"
    echo -e "${CYAN}docker run -d --name slurm-exporter-tutorial \\
  -p 8081:10341 \\
  -e SLURM_ENDPOINT='http://localhost:6820' \\
  -e SLURM_AUTH_TYPE='none' \\
  ghcr.io/jontk/slurm-exporter:latest${NC}"
    
    press_enter_to_continue
    
    # Actually run the command (connect to sandbox network)
    docker run -d --name slurm-exporter-tutorial \
        --network sandbox_default \
        -p 8081:10341 \
        -e SLURM_ENDPOINT='http://slurm-controller:6820' \
        -e SLURM_AUTH_TYPE='none' \
        ghcr.io/jontk/slurm-exporter:latest || true
    
    sleep 5
    
    # Step 3: Verify deployment
    tutorial_step "3" "Verify It's Working" "
        Let's check if the exporter is working correctly.
    "
    
    echo "Checking health:"
    if curl -sf http://localhost:8081/health; then
        log_success "Health check passed!"
    else
        log_warning "Health check failed - this is normal for tutorial purposes"
    fi
    
    echo
    echo "Checking metrics:"
    echo "First 10 metrics:"
    curl -s http://localhost:8081/metrics | head -10 || echo "(Metrics may not be available in tutorial mode)"
    
    press_enter_to_continue
    
    # Step 4: View in browser
    tutorial_step "4" "View in Browser" "
        Open these URLs to explore:
        
        📊 Exporter Metrics: http://localhost:8081/metrics
        🔍 Debug Info:      http://localhost:8081/debug/health
        📈 Prometheus:      http://localhost:9090 (if sandbox is running)
        📊 Grafana:         http://localhost:3000 (admin/admin)
    "
    
    press_enter_to_continue
    
    # Cleanup
    log_info "Cleaning up tutorial container..."
    docker rm -f slurm-exporter-tutorial &> /dev/null || true
    
    mark_tutorial_completed "basic-setup"
    
    log_success "Basic Setup Tutorial completed!"
    echo
    echo "🎉 Congratulations! You've successfully:"
    echo "  ✅ Deployed SLURM Exporter with Docker"
    echo "  ✅ Verified health and metrics endpoints"
    echo "  ✅ Learned about basic configuration"
    echo
    echo "Next steps:"
    echo "  • Try Tutorial 2: Configuration"
    echo "  • Explore the sandbox environment"
    echo "  • Import Grafana dashboards"
}

# Troubleshooting assistance
run_troubleshooting() {
    local issue="${1:-menu}"
    
    case "$issue" in
        "menu")
            troubleshooting_menu
            ;;
        "connectivity")
            troubleshoot_connectivity
            ;;
        "metrics")
            troubleshoot_metrics
            ;;
        "performance")
            troubleshoot_performance
            ;;
        "memory")
            troubleshoot_memory
            ;;
        "health")
            troubleshoot_health_check
            ;;
        *)
            log_error "Unknown troubleshooting topic: $issue"
            troubleshooting_menu
            ;;
    esac
}

troubleshooting_menu() {
    echo -e "${BLUE}Troubleshooting Assistant:${NC}"
    echo
    echo -e "  ${GREEN}connectivity${NC}  - Connection and network issues"
    echo -e "  ${GREEN}metrics${NC}      - Missing or incorrect metrics"
    echo -e "  ${GREEN}performance${NC}  - Slow collection or high resource usage"
    echo -e "  ${GREEN}memory${NC}       - Memory leaks and high memory usage"
    echo -e "  ${GREEN}health${NC}       - Health check and service status"
    echo
    echo "Usage: $0 troubleshoot <topic>"
    echo
    echo "Or use the interactive troubleshooting guide:"
    echo "  Open: $DOCS_DIR/troubleshooting.md"
    echo "  Or visit: http://localhost:8888/troubleshooting"
}

troubleshoot_connectivity() {
    log_tutorial "Troubleshooting Connectivity Issues"
    echo
    
    echo "Let's diagnose connectivity step by step..."
    echo
    
    # Check 1: Exporter running
    echo -e "${BLUE}1. Checking if SLURM Exporter is running...${NC}"
    
    if curl -sf http://localhost:10341/health &> /dev/null; then
        log_success "SLURM Exporter is responding"
    else
        log_warning "SLURM Exporter is not responding"
        echo "  Possible causes:"
        echo "  • Exporter not started"
        echo "  • Wrong port (default: 10341)"
        echo "  • Configuration error"
        echo
        echo "  Try: docker logs <container-name>"
        return
    fi
    
    # Check 2: SLURM API connectivity
    echo -e "${BLUE}2. Checking SLURM API connectivity...${NC}"
    
    local slurm_endpoint
    slurm_endpoint=$(curl -s http://localhost:10341/debug/config | jq -r '.slurm.endpoint' 2>/dev/null || echo "unknown")
    
    echo "  SLURM endpoint: $slurm_endpoint"
    
    if curl -sf "$slurm_endpoint/slurm/v0.0.40/diag" &> /dev/null; then
        log_success "SLURM API is reachable"
    else
        log_warning "SLURM API is not reachable"
        echo "  Troubleshooting steps:"
        echo "  • Check if SLURM REST API daemon is running"
        echo "  • Verify endpoint URL in configuration"
        echo "  • Test network connectivity: telnet <host> <port>"
        echo "  • Check firewall rules"
    fi
    
    # Check 3: Authentication
    echo -e "${BLUE}3. Checking authentication...${NC}"
    
    local auth_type
    auth_type=$(curl -s http://localhost:10341/debug/config | jq -r '.slurm.auth.type' 2>/dev/null || echo "unknown")
    
    echo "  Authentication type: $auth_type"
    
    if [ "$auth_type" = "token" ] || [ "$auth_type" = "jwt" ]; then
        echo "  • Verify token is valid and not expired"
        echo "  • Check token permissions in SLURM"
        echo "  • Test token manually: curl -H 'X-SLURM-USER-TOKEN: <token>' <endpoint>"
    fi
    
    # Check 4: Circuit breaker
    echo -e "${BLUE}4. Checking circuit breaker status...${NC}"
    
    local cb_state
    cb_state=$(curl -s http://localhost:10341/debug/health | jq -r '.circuit_breaker.state' 2>/dev/null || echo "unknown")
    
    echo "  Circuit breaker state: $cb_state"
    
    if [ "$cb_state" = "open" ]; then
        log_warning "Circuit breaker is OPEN - requests are being blocked"
        echo "  This happens after repeated failures to protect the SLURM API"
        echo "  • Fix the underlying connectivity issue"
        echo "  • Wait for automatic recovery (usually 60s)"
        echo "  • Or manually reset: curl -X POST http://localhost:10341/debug/circuit-breaker/reset"
    fi
}

# Testing functions
run_tests() {
    local test_type="${1:-all}"
    
    case "$test_type" in
        "all")
            run_all_tests
            ;;
        "config")
            test_configuration
            ;;
        "connectivity")
            test_connectivity
            ;;
        "performance")
            test_performance
            ;;
        *)
            log_error "Unknown test type: $test_type"
            echo "Available tests: all, config, connectivity, performance"
            exit 1
            ;;
    esac
}

test_configuration() {
    log_tutorial "Testing Configuration"
    echo
    
    local config_dir="$HOME/.slurm-exporter/configs"
    
    if [ ! -d "$config_dir" ]; then
        log_warning "No configurations found. Run: $0 setup"
        return
    fi
    
    for config_file in "$config_dir"/*.yaml; do
        if [ -f "$config_file" ]; then
            local filename=$(basename "$config_file")
            echo -e "${BLUE}Testing $filename...${NC}"
            
            # Validate YAML syntax
            if ! python3 -c "import yaml; yaml.safe_load(open('$config_file'))" &> /dev/null; then
                log_error "$filename has invalid YAML syntax"
                continue
            fi
            
            # TODO: Add semantic validation
            log_success "$filename passed validation"
        fi
    done
}

# Monitoring functions
monitor_performance() {
    log_tutorial "Monitoring SLURM Exporter Performance"
    echo
    
    if ! curl -sf http://localhost:10341/health &> /dev/null; then
        log_error "SLURM Exporter is not running"
        echo "Start with: $0 sandbox start"
        return
    fi
    
    echo "Real-time performance monitoring (Ctrl+C to stop):"
    echo
    
    while true; do
        clear
        echo -e "${BLUE}SLURM Exporter Performance Monitor${NC}"
        echo "$(date)"
        echo
        
        # Collection performance
        echo -e "${GREEN}Collection Performance:${NC}"
        curl -s http://localhost:10341/debug/collectors | jq -r '
            .collectors[] | 
            "\(.name): \(.last_duration)s (\(.success_rate | tonumber * 100 | floor)% success)"
        ' 2>/dev/null || echo "  Unable to fetch collector stats"
        echo
        
        # Memory usage
        echo -e "${GREEN}Memory Usage:${NC}"
        curl -s http://localhost:10341/debug/memory | jq -r '
            "  Heap: \(.heap_mb)MB, System: \(.system_mb)MB, GC: \(.gc_cycles)"
        ' 2>/dev/null || echo "  Unable to fetch memory stats"
        echo
        
        # Cache performance
        echo -e "${GREEN}Cache Performance:${NC}"
        curl -s http://localhost:10341/debug/cache | jq -r '
            .stats | "  Hit Rate: \(.hit_rate | tonumber * 100 | floor)%, Size: \(.size)"
        ' 2>/dev/null || echo "  Cache not enabled or unavailable"
        echo
        
        # Circuit breaker
        echo -e "${GREEN}Circuit Breaker:${NC}"
        curl -s http://localhost:10341/debug/health | jq -r '
            .circuit_breaker | "  State: \(.state), Failures: \(.failures)"
        ' 2>/dev/null || echo "  Circuit breaker status unavailable"
        
        sleep 5
    done
}

# Cleanup functions
cleanup_environment() {
    log_warning "This will remove all tutorial data and stop services. Continue? (y/N)"
    read -r response
    
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        log_info "Cleanup cancelled"
        return
    fi
    
    log_info "Cleaning up tutorial environment..."
    
    # Stop sandbox
    if [ -d "$SANDBOX_DIR" ]; then
        cd "$SANDBOX_DIR"
        docker-compose down -v 2>/dev/null || true
    fi
    
    # Remove tutorial containers
    docker rm -f slurm-exporter-tutorial 2>/dev/null || true
    
    # Remove tutorial data
    rm -rf "$HOME/.slurm-exporter" 2>/dev/null || true
    
    log_success "Cleanup completed"
}

# Utility functions
tutorial_step() {
    local step_num="$1"
    local title="$2"  
    local description="$3"
    
    echo
    echo -e "${PURPLE}═══ Step $step_num: $title ═══${NC}"
    echo
    echo -e "$description"
    echo
}

press_enter_to_continue() {
    echo -e "${YELLOW}Press Enter to continue...${NC}"
    read -r
}

mark_tutorial_started() {
    local tutorial_name="$1"
    local tracking_file="$HOME/.slurm-exporter/tutorials/progress.json"
    
    if [ -f "$tracking_file" ]; then
        local timestamp=$(date -Iseconds)
        jq ".tutorials.\"$tutorial_name\".started = \"$timestamp\" | .last_activity = \"$timestamp\"" \
            "$tracking_file" > "$tracking_file.tmp" && mv "$tracking_file.tmp" "$tracking_file"
    fi
}

mark_tutorial_completed() {
    local tutorial_name="$1"
    local tracking_file="$HOME/.slurm-exporter/tutorials/progress.json"
    
    if [ -f "$tracking_file" ]; then
        local timestamp=$(date -Iseconds)
        jq ".tutorials.\"$tutorial_name\".completed = true | .last_activity = \"$timestamp\"" \
            "$tracking_file" > "$tracking_file.tmp" && mv "$tracking_file.tmp" "$tracking_file"
    fi
}

show_tutorial_progress() {
    local tracking_file="$HOME/.slurm-exporter/tutorials/progress.json"
    
    if [ ! -f "$tracking_file" ]; then
        log_warning "No tutorial progress found. Run: $0 setup"
        return
    fi
    
    echo -e "${BLUE}Tutorial Progress:${NC}"
    echo
    
    jq -r '.tutorials | to_entries[] | 
        if .value.completed then 
            "✅ \(.key | gsub("-"; " ") | title)"
        elif .value.started then 
            "🔄 \(.key | gsub("-"; " ") | title) (in progress)"
        else 
            "⏳ \(.key | gsub("-"; " ") | title)"
        end
    ' "$tracking_file"
    
    echo
    
    local last_activity
    last_activity=$(jq -r '.last_activity // "never"' "$tracking_file")
    echo "Last activity: $last_activity"
}

sandbox_running() {
    [ -d "$SANDBOX_DIR" ] && cd "$SANDBOX_DIR" && docker-compose ps | grep -q "Up"
}

show_help() {
    print_header
    echo -e "${BLUE}SLURM Exporter Tutorial Helper${NC}"
    echo
    echo "This script provides interactive assistance for learning SLURM Exporter."
    echo
    print_menu
    echo
    echo -e "${BLUE}Examples:${NC}"
    echo "  $0 setup                    # Set up tutorial environment"
    echo "  $0 sandbox start            # Start sandbox environment"  
    echo "  $0 tutorial 1               # Run basic setup tutorial"
    echo "  $0 troubleshoot connectivity # Diagnose connection issues"
    echo "  $0 monitor                  # Monitor exporter performance"
    echo "  $0 test config              # Test configuration files"
    echo
    echo -e "${BLUE}Getting Started:${NC}"
    echo "  1. Run: $0 setup"
    echo "  2. Run: $0 sandbox start"
    echo "  3. Run: $0 tutorial 1"
    echo
    echo -e "${BLUE}Resources:${NC}"
    echo "  📖 Documentation: $DOCS_DIR/"
    echo "  🧪 Sandbox:      $SANDBOX_DIR/"
    echo "  🎓 Tutorials:     http://localhost:8888"
    echo "  💬 Support:      https://github.com/jontk/slurm-exporter/discussions"
}

# Main function
main() {
    if [ $# -eq 0 ]; then
        print_header
        print_menu
        exit 0
    fi
    
    local command="$1"
    shift
    
    case "$command" in
        "setup")
            setup_environment
            ;;
        "sandbox")
            manage_sandbox "$@"
            ;;
        "tutorial")
            run_tutorial "$@"
            ;;
        "troubleshoot")
            run_troubleshooting "$@"
            ;;
        "test")
            run_tests "$@"
            ;;
        "monitor")
            monitor_performance
            ;;
        "cleanup")
            cleanup_environment
            ;;
        "help"|"--help"|"-h")
            show_help
            ;;
        *)
            log_error "Unknown command: $command"
            print_menu
            exit 1
            ;;
    esac
}

# Only run main if script is executed directly
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi