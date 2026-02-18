#!/bin/bash

# SLURM Exporter Configuration Wizard
# Interactive configuration generator for SLURM Exporter

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Global variables
CONFIG_FILE="slurm-exporter-config.yaml"
DEPLOYMENT_TYPE=""
ENVIRONMENT=""
SLURM_URL=""
AUTH_TYPE=""
AUTH_TOKEN=""
AUTH_USER=""
AUTH_PASS=""
COLLECTORS=()
ADVANCED_FEATURES=()

# Utility functions
print_header() {
    echo -e "${BLUE}╔══════════════════════════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║                    🚀 SLURM Exporter Configuration Wizard                    ║${NC}"
    echo -e "${BLUE}║                                                                              ║${NC}"
    echo -e "${BLUE}║         Generate production-ready configuration in 2 minutes                ║${NC}"
    echo -e "${BLUE}╚══════════════════════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

print_step() {
    echo -e "${CYAN}📋 $1${NC}"
    echo ""
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${PURPLE}ℹ️  $1${NC}"
}

ask_question() {
    local question="$1"
    local default="$2"
    local answer

    if [[ -n "$default" ]]; then
        echo -ne "${YELLOW}❓ $question [${default}]: ${NC}"
    else
        echo -ne "${YELLOW}❓ $question: ${NC}"
    fi
    
    read -r answer
    echo "${answer:-$default}"
}

ask_yes_no() {
    local question="$1"
    local default="$2"
    local answer

    while true; do
        if [[ "$default" == "y" ]]; then
            echo -ne "${YELLOW}❓ $question [Y/n]: ${NC}"
        else
            echo -ne "${YELLOW}❓ $question [y/N]: ${NC}"
        fi
        
        read -r answer
        answer="${answer:-$default}"
        
        case "${answer,,}" in
            y|yes) return 0 ;;
            n|no) return 1 ;;
            *) echo -e "${RED}Please answer yes or no.${NC}" ;;
        esac
    done
}

select_option() {
    local prompt="$1"
    shift
    local options=("$@")
    local choice

    echo -e "${YELLOW}❓ $prompt${NC}"
    for i in "${!options[@]}"; do
        echo "  $((i+1)). ${options[i]}"
    done
    
    while true; do
        echo -ne "${YELLOW}Enter choice [1-${#options[@]}]: ${NC}"
        read -r choice
        
        if [[ "$choice" =~ ^[0-9]+$ ]] && [[ "$choice" -ge 1 ]] && [[ "$choice" -le "${#options[@]}" ]]; then
            echo "${options[$((choice-1))]}"
            return
        else
            echo -e "${RED}Invalid choice. Please enter a number between 1 and ${#options[@]}.${NC}"
        fi
    done
}

validate_url() {
    local url="$1"
    if [[ ! "$url" =~ ^https?://[^[:space:]]+$ ]]; then
        return 1
    fi
    return 0
}

test_slurm_connection() {
    local url="$1"
    local auth_header="$2"

    print_info "Testing connection to SLURM API..."
    
    if command -v curl >/dev/null 2>&1; then
        local cmd="curl -s -f --max-time 10"
        if [[ -n "$auth_header" ]]; then
            cmd="$cmd -H 'Authorization: $auth_header'"
        fi
        cmd="$cmd '$url/slurm/v0.0.44/ping'"
        
        if eval "$cmd" >/dev/null 2>&1; then
            print_success "SLURM API connection successful!"
            return 0
        else
            print_warning "Could not connect to SLURM API. Please verify the URL and authentication."
            return 1
        fi
    else
        print_warning "curl not available. Skipping connection test."
        return 0
    fi
}

# Configuration steps
step1_welcome() {
    print_header
    print_info "This wizard will help you generate a production-ready configuration for SLURM Exporter."
    print_info "You can customize the configuration based on your deployment environment and requirements."
    echo ""
    
    if ask_yes_no "Ready to start" "y"; then
        echo ""
        return 0
    else
        echo "Configuration wizard cancelled."
        exit 0
    fi
}

step2_deployment_type() {
    print_step "Step 1: Select Deployment Type"
    
    DEPLOYMENT_TYPE=$(select_option "How will you deploy SLURM Exporter?" \
        "Kubernetes with Helm" \
        "Docker Compose" \
        "Docker (standalone)" \
        "Binary (systemd)" \
        "Development/Testing")
    
    print_success "Selected: $DEPLOYMENT_TYPE"
    echo ""
}

step3_environment() {
    print_step "Step 2: Select Environment"
    
    ENVIRONMENT=$(select_option "What type of environment is this?" \
        "Production" \
        "Staging" \
        "Development" \
        "Testing")
    
    print_success "Selected: $ENVIRONMENT"
    echo ""
}

step4_slurm_connection() {
    print_step "Step 3: SLURM API Configuration"
    
    while true; do
        SLURM_URL=$(ask_question "SLURM REST API URL (e.g., https://slurm-api.example.com:6820)" "")
        
        if validate_url "$SLURM_URL"; then
            break
        else
            print_error "Invalid URL format. Please enter a valid HTTP/HTTPS URL."
        fi
    done
    
    print_success "SLURM URL: $SLURM_URL"
    echo ""
}

step5_authentication() {
    print_step "Step 4: Authentication Configuration"
    
    AUTH_TYPE=$(select_option "Select authentication method:" \
        "JWT Token" \
        "API Key" \
        "Basic Auth (username/password)" \
        "Certificate-based" \
        "No authentication (testing only)")
    
    case "$AUTH_TYPE" in
        "JWT Token")
            AUTH_TOKEN=$(ask_question "Enter JWT token (or environment variable name)" "SLURM_TOKEN")
            ;;
        "API Key")
            AUTH_TOKEN=$(ask_question "Enter API key (or environment variable name)" "SLURM_API_KEY")
            ;;
        "Basic Auth (username/password)")
            AUTH_USER=$(ask_question "Username" "")
            AUTH_PASS=$(ask_question "Password (or environment variable name)" "SLURM_PASSWORD")
            ;;
        "Certificate-based")
            print_info "Certificate-based auth requires additional TLS configuration."
            ;;
        "No authentication (testing only)")
            if [[ "$ENVIRONMENT" == "Production" ]]; then
                print_warning "No authentication is not recommended for production!"
            fi
            ;;
    esac
    
    # Test connection
    local auth_header=""
    case "$AUTH_TYPE" in
        "JWT Token"|"API Key")
            if [[ "$AUTH_TOKEN" != *"$"* && "$AUTH_TOKEN" != *"{"* ]]; then
                auth_header="Bearer $AUTH_TOKEN"
            fi
            ;;
        "Basic Auth (username/password)")
            if [[ "$AUTH_PASS" != *"$"* && "$AUTH_PASS" != *"{"* ]]; then
                auth_header="Basic $(echo -n "$AUTH_USER:$AUTH_PASS" | base64)"
            fi
            ;;
    esac
    
    if [[ -n "$auth_header" ]]; then
        test_slurm_connection "$SLURM_URL" "$auth_header"
    fi
    
    print_success "Authentication configured: $AUTH_TYPE"
    echo ""
}

step6_collectors() {
    print_step "Step 5: Select Collectors"
    
    local available_collectors=(
        "jobs:Job information and states"
        "nodes:Node status and resources"
        "partitions:Partition information"
        "qos:Quality of Service limits"
        "reservations:Resource reservations"
        "fairshare:Fairshare data and trends"
        "accounts:Account hierarchy and quotas"
        "cluster:Cluster-wide statistics"
        "job_performance:Job performance metrics"
        "job_analytics:Advanced job analytics"
        "energy_monitoring:Power and energy usage"
        "live_job_monitoring:Real-time job tracking"
    )
    
    print_info "Select collectors to enable (recommended: jobs, nodes, partitions for basic monitoring):"
    echo ""
    
    for collector in "${available_collectors[@]}"; do
        local name="${collector%%:*}"
        local desc="${collector##*:}"
        
        if ask_yes_no "Enable $name collector ($desc)" "n"; then
            COLLECTORS+=("$name")
        fi
    done
    
    if [[ ${#COLLECTORS[@]} -eq 0 ]]; then
        print_warning "No collectors selected. Adding basic collectors."
        COLLECTORS=("jobs" "nodes" "partitions")
    fi
    
    print_success "Selected collectors: ${COLLECTORS[*]}"
    echo ""
}

step7_advanced_features() {
    print_step "Step 6: Advanced Features"
    
    print_info "Configure advanced features for enhanced monitoring and performance:"
    echo ""
    
    # Performance features
    if ask_yes_no "Enable intelligent caching for better performance" "y"; then
        ADVANCED_FEATURES+=("caching")
    fi
    
    if ask_yes_no "Enable batch processing for high-throughput clusters" "y"; then
        ADVANCED_FEATURES+=("batch_processing")
    fi
    
    # Observability features
    if ask_yes_no "Enable circuit breaker for API protection" "y"; then
        ADVANCED_FEATURES+=("circuit_breaker")
    fi
    
    if ask_yes_no "Enable health monitoring endpoints" "y"; then
        ADVANCED_FEATURES+=("health_monitoring")
    fi
    
    if ask_yes_no "Enable OpenTelemetry tracing" "n"; then
        ADVANCED_FEATURES+=("opentelemetry")
    fi
    
    # Debug features
    if [[ "$ENVIRONMENT" != "Production" ]]; then
        if ask_yes_no "Enable debug endpoints for troubleshooting" "y"; then
            ADVANCED_FEATURES+=("debug_endpoints")
        fi
        
        if ask_yes_no "Enable collection profiling" "n"; then
            ADVANCED_FEATURES+=("profiling")
        fi
    fi
    
    # Security features
    if [[ "$ENVIRONMENT" == "Production" ]]; then
        if ask_yes_no "Enable TLS/SSL" "y"; then
            ADVANCED_FEATURES+=("tls")
        fi
        
        if ask_yes_no "Enable authentication for debug endpoints" "y"; then
            ADVANCED_FEATURES+=("auth_debug")
        fi
    fi
    
    print_success "Advanced features: ${ADVANCED_FEATURES[*]}"
    echo ""
}

step8_generate_config() {
    print_step "Step 7: Generate Configuration"
    
    print_info "Generating configuration file: $CONFIG_FILE"
    
    # Generate the configuration
    generate_config_file
    
    print_success "Configuration generated successfully!"
    echo ""
    
    # Show summary
    print_step "Configuration Summary"
    echo -e "${CYAN}📝 Configuration Details:${NC}"
    echo "  • Deployment: $DEPLOYMENT_TYPE"
    echo "  • Environment: $ENVIRONMENT"
    echo "  • SLURM URL: $SLURM_URL"
    echo "  • Authentication: $AUTH_TYPE"
    echo "  • Collectors: ${COLLECTORS[*]}"
    echo "  • Advanced Features: ${ADVANCED_FEATURES[*]}"
    echo "  • Config File: $CONFIG_FILE"
    echo ""
}

step9_deployment_instructions() {
    print_step "Step 8: Deployment Instructions"
    
    case "$DEPLOYMENT_TYPE" in
        "Kubernetes with Helm")
            generate_kubernetes_instructions
            ;;
        "Docker Compose")
            generate_docker_compose_instructions
            ;;
        "Docker (standalone)")
            generate_docker_instructions
            ;;
        "Binary (systemd)")
            generate_binary_instructions
            ;;
        "Development/Testing")
            generate_dev_instructions
            ;;
    esac
}

# Configuration generation functions
generate_config_file() {
    cat > "$CONFIG_FILE" << EOF
# SLURM Exporter Configuration
# Generated by Configuration Wizard
# Environment: $ENVIRONMENT
# Generated: $(date)

# SLURM Connection
slurm:
  url: "$SLURM_URL"
EOF

    # Add authentication
    case "$AUTH_TYPE" in
        "JWT Token")
            cat >> "$CONFIG_FILE" << EOF
  auth:
    type: "token"
    token: "\${$AUTH_TOKEN}"
EOF
            ;;
        "API Key")
            cat >> "$CONFIG_FILE" << EOF
  auth:
    type: "api_key"
    api_key: "\${$AUTH_TOKEN}"
EOF
            ;;
        "Basic Auth (username/password)")
            cat >> "$CONFIG_FILE" << EOF
  auth:
    type: "basic"
    username: "$AUTH_USER"
    password: "\${$AUTH_PASS}"
EOF
            ;;
        "Certificate-based")
            cat >> "$CONFIG_FILE" << EOF
  auth:
    type: "tls"
    cert_file: "/etc/ssl/certs/client.crt"
    key_file: "/etc/ssl/private/client.key"
    ca_file: "/etc/ssl/certs/ca.crt"
EOF
            ;;
    esac

    # Add connection settings based on environment
    if [[ "$ENVIRONMENT" == "Production" ]]; then
        cat >> "$CONFIG_FILE" << EOF
  timeout: 30s
  retry_attempts: 3
  retry_delay: 5s
EOF
    else
        cat >> "$CONFIG_FILE" << EOF
  timeout: 10s
  retry_attempts: 1
EOF
    fi

    # Add collectors configuration
    cat >> "$CONFIG_FILE" << EOF

# Collector Configuration
collectors:
  global:
EOF

    if [[ "$ENVIRONMENT" == "Production" ]]; then
        cat >> "$CONFIG_FILE" << EOF
    default_interval: 30s
    max_concurrency: 10
    timeout: 25s
EOF
    else
        cat >> "$CONFIG_FILE" << EOF
    default_interval: 15s
    max_concurrency: 5
    timeout: 10s
EOF
    fi

    # Add individual collectors
    for collector in "${COLLECTORS[@]}"; do
        cat >> "$CONFIG_FILE" << EOF
  $collector:
    enabled: true
EOF
        case "$collector" in
            "jobs")
                if [[ "$ENVIRONMENT" == "Production" ]]; then
                    cat >> "$CONFIG_FILE" << EOF
    interval: 15s
    filters:
      max_age: "24h"
      exclude_states: ["COMPLETED"]
EOF
                fi
                ;;
            "nodes")
                cat >> "$CONFIG_FILE" << EOF
    interval: 30s
EOF
                ;;
            "partitions")
                cat >> "$CONFIG_FILE" << EOF
    interval: 60s
EOF
                ;;
        esac
    done

    # Add server configuration
    cat >> "$CONFIG_FILE" << EOF

# Server Configuration
server:
  address: ":10341"
EOF

    if [[ "$ENVIRONMENT" == "Production" ]]; then
        cat >> "$CONFIG_FILE" << EOF
  read_timeout: 30s
  write_timeout: 30s
  idle_timeout: 120s
EOF
    fi

    # Add TLS if enabled
    if [[ " ${ADVANCED_FEATURES[*]} " =~ " tls " ]]; then
        cat >> "$CONFIG_FILE" << EOF
  tls:
    enabled: true
    cert_file: "/etc/ssl/certs/tls.crt"
    key_file: "/etc/ssl/private/tls.key"
EOF
    fi

    # Add performance features
    if [[ " ${ADVANCED_FEATURES[*]} " =~ " caching " ]]; then
        cat >> "$CONFIG_FILE" << EOF

# Performance Configuration
performance:
  caching:
    enabled: true
    default_ttl: 5m
    max_entries: 10000
EOF

        if [[ "$ENVIRONMENT" == "Production" ]]; then
            cat >> "$CONFIG_FILE" << EOF
    adaptive_ttl:
      enabled: true
      min_ttl: 30s
      max_ttl: 30m
      change_threshold: 0.1
EOF
        fi
    fi

    if [[ " ${ADVANCED_FEATURES[*]} " =~ " batch_processing " ]]; then
        cat >> "$CONFIG_FILE" << EOF
  batch_processing:
    enabled: true
    max_batch_size: 500
    flush_interval: 5s
    max_wait_time: 10s
EOF
    fi

    # Add observability features
    if [[ " ${ADVANCED_FEATURES[*]} " =~ " circuit_breaker " ]]; then
        cat >> "$CONFIG_FILE" << EOF

# Observability Configuration
observability:
  circuit_breaker:
    enabled: true
    failure_threshold: 10
    reset_timeout: 60s
    half_open_max_requests: 3
EOF
    fi

    if [[ " ${ADVANCED_FEATURES[*]} " =~ " health_monitoring " ]]; then
        cat >> "$CONFIG_FILE" << EOF
  health_monitoring:
    enabled: true
    check_interval: 30s
    unhealthy_threshold: 3
EOF
    fi

    if [[ " ${ADVANCED_FEATURES[*]} " =~ " opentelemetry " ]]; then
        cat >> "$CONFIG_FILE" << EOF
  tracing:
    enabled: true
    endpoint: "http://jaeger:14268/api/traces"
    service_name: "slurm-exporter"
    sample_rate: 0.1
EOF
    fi

    # Add debug features
    if [[ " ${ADVANCED_FEATURES[*]} " =~ " debug_endpoints " ]]; then
        cat >> "$CONFIG_FILE" << EOF

# Debug Configuration
debug:
  endpoints:
    enabled: true
    path_prefix: "/debug"
EOF

        if [[ " ${ADVANCED_FEATURES[*]} " =~ " auth_debug " ]]; then
            cat >> "$CONFIG_FILE" << EOF
    authentication:
      enabled: true
      token: "\${DEBUG_TOKEN}"
EOF
        fi
    fi

    if [[ " ${ADVANCED_FEATURES[*]} " =~ " profiling " ]]; then
        cat >> "$CONFIG_FILE" << EOF
  profiling:
    enabled: true
    cpu_profile_duration: 30s
    memory_profile_interval: 5m
    storage_type: "file"
    storage_path: "/tmp/profiles"
EOF
    fi

    # Add logging configuration
    cat >> "$CONFIG_FILE" << EOF

# Logging Configuration
logging:
EOF

    if [[ "$ENVIRONMENT" == "Production" ]]; then
        cat >> "$CONFIG_FILE" << EOF
  level: info
  format: json
  output: stdout
EOF
    else
        cat >> "$CONFIG_FILE" << EOF
  level: debug
  format: text
  output: stdout
EOF
    fi

    # Add metrics configuration
    cat >> "$CONFIG_FILE" << EOF

# Metrics Configuration
metrics:
  namespace: slurm
  subsystem: exporter
EOF

    if [[ "$ENVIRONMENT" == "Production" ]]; then
        cat >> "$CONFIG_FILE" << EOF
  cardinality_limit: 100000
  label_limit: 50
EOF
    fi
}

generate_kubernetes_instructions() {
    local secret_file="slurm-exporter-secret.yaml"
    local values_file="slurm-exporter-values.yaml"
    
    print_info "Generating Kubernetes deployment files..."
    
    # Generate secret
    cat > "$secret_file" << EOF
apiVersion: v1
kind: Secret
metadata:
  name: slurm-exporter-credentials
  namespace: slurm-exporter
type: Opaque
stringData:
EOF

    case "$AUTH_TYPE" in
        "JWT Token"|"API Key")
            cat >> "$secret_file" << EOF
  auth-token: "your-actual-token-here"
EOF
            ;;
        "Basic Auth (username/password)")
            cat >> "$secret_file" << EOF
  username: "$AUTH_USER"
  password: "your-actual-password-here"
EOF
            ;;
    esac

    if [[ " ${ADVANCED_FEATURES[*]} " =~ " auth_debug " ]]; then
        cat >> "$secret_file" << EOF
  debug-token: "your-debug-token-here"
EOF
    fi

    # Generate Helm values
    cat > "$values_file" << EOF
# Helm values for SLURM Exporter
# Generated by Configuration Wizard

image:
  repository: ghcr.io/jontk/slurm-exporter
  tag: latest
  pullPolicy: Always

# Load configuration from ConfigMap
configMap:
  create: true
  data:
    config.yaml: |
$(sed 's/^/      /' "$CONFIG_FILE")

# Load secrets
secret:
  create: false
  name: slurm-exporter-credentials

replicaCount: $([ "$ENVIRONMENT" = "Production" ] && echo "2" || echo "1")

resources:
EOF

    if [[ "$ENVIRONMENT" == "Production" ]]; then
        cat >> "$values_file" << EOF
  limits:
    cpu: 500m
    memory: 512Mi
  requests:
    cpu: 100m
    memory: 128Mi
EOF
    else
        cat >> "$values_file" << EOF
  limits:
    cpu: 200m
    memory: 256Mi
  requests:
    cpu: 50m
    memory: 64Mi
EOF
    fi

    cat >> "$values_file" << EOF

service:
  type: ClusterIP
  port: 10341

serviceMonitor:
  enabled: true
  interval: 30s
  path: /metrics

ingress:
  enabled: false
EOF

    # Show deployment commands
    cat << EOF
${GREEN}🚢 Kubernetes Deployment Instructions:${NC}

1. Create namespace and apply secrets:
   ${CYAN}kubectl create namespace slurm-exporter
   kubectl apply -f $secret_file${NC}

2. Add Helm repository:
   ${CYAN}helm repo add slurm-exporter https://jontk.github.io/slurm-exporter
   helm repo update${NC}

3. Deploy with Helm:
   ${CYAN}helm install slurm-exporter slurm-exporter/slurm-exporter \\
     --namespace slurm-exporter \\
     --values $values_file${NC}

4. Verify deployment:
   ${CYAN}kubectl get pods -n slurm-exporter
   kubectl port-forward svc/slurm-exporter 10341:10341 -n slurm-exporter${NC}

5. Test metrics endpoint:
   ${CYAN}curl http://localhost:10341/metrics${NC}

Generated files:
  • $CONFIG_FILE - Configuration file
  • $secret_file - Kubernetes secret
  • $values_file - Helm values
EOF
}

generate_docker_compose_instructions() {
    local compose_file="docker-compose.yml"
    local env_file=".env"
    
    print_info "Generating Docker Compose files..."
    
    # Generate environment file
    cat > "$env_file" << EOF
# SLURM Exporter Environment Variables
# Generated by Configuration Wizard
EOF

    case "$AUTH_TYPE" in
        "JWT Token")
            echo "$AUTH_TOKEN=your-actual-token-here" >> "$env_file"
            ;;
        "API Key")
            echo "$AUTH_TOKEN=your-actual-api-key-here" >> "$env_file"
            ;;
        "Basic Auth (username/password)")
            echo "$AUTH_PASS=your-actual-password-here" >> "$env_file"
            ;;
    esac

    if [[ " ${ADVANCED_FEATURES[*]} " =~ " auth_debug " ]]; then
        echo "DEBUG_TOKEN=your-debug-token-here" >> "$env_file"
    fi

    # Generate docker-compose.yml
    cat > "$compose_file" << EOF
version: '3.8'

services:
  slurm-exporter:
    image: ghcr.io/jontk/slurm-exporter:latest
    container_name: slurm-exporter
    restart: unless-stopped
    ports:
      - "10341:10341"
    volumes:
      - ./config:/etc/slurm-exporter:ro
EOF

    if [[ " ${ADVANCED_FEATURES[*]} " =~ " profiling " ]]; then
        cat >> "$compose_file" << EOF
      - ./profiles:/tmp/profiles
EOF
    fi

    cat >> "$compose_file" << EOF
    env_file:
      - .env
    environment:
      - CONFIG_FILE=/etc/slurm-exporter/config.yaml
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:10341/health"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s
EOF

    if [[ "$ENVIRONMENT" == "Production" ]]; then
        cat >> "$compose_file" << EOF
    deploy:
      resources:
        limits:
          cpus: '0.5'
          memory: 512M
        reservations:
          cpus: '0.1'
          memory: 128M
      restart_policy:
        condition: on-failure
        delay: 5s
        max_attempts: 3
EOF
    fi

    # Show deployment commands
    cat << EOF
${GREEN}🐳 Docker Compose Deployment Instructions:${NC}

1. Create configuration directory:
   ${CYAN}mkdir config && cp $CONFIG_FILE config/${NC}

2. Update environment variables:
   ${CYAN}# Edit $env_file and replace placeholder values${NC}

3. Start the service:
   ${CYAN}docker-compose up -d${NC}

4. Check logs:
   ${CYAN}docker-compose logs -f slurm-exporter${NC}

5. Test metrics endpoint:
   ${CYAN}curl http://localhost:10341/metrics${NC}

Generated files:
  • $CONFIG_FILE - Configuration file
  • $compose_file - Docker Compose file
  • $env_file - Environment variables
EOF
}

generate_docker_instructions() {
    local env_file="slurm-exporter.env"
    
    # Generate environment file
    cat > "$env_file" << EOF
# SLURM Exporter Environment Variables
CONFIG_FILE=/config/config.yaml
EOF

    case "$AUTH_TYPE" in
        "JWT Token")
            echo "$AUTH_TOKEN=your-actual-token-here" >> "$env_file"
            ;;
        "API Key")
            echo "$AUTH_TOKEN=your-actual-api-key-here" >> "$env_file"
            ;;
        "Basic Auth (username/password)")
            echo "$AUTH_PASS=your-actual-password-here" >> "$env_file"
            ;;
    esac

    cat << EOF
${GREEN}🐋 Docker Deployment Instructions:${NC}

1. Create configuration directory:
   ${CYAN}mkdir -p config && cp $CONFIG_FILE config/${NC}

2. Update environment file:
   ${CYAN}# Edit $env_file and replace placeholder values${NC}

3. Run with Docker:
   ${CYAN}docker run -d \\
     --name slurm-exporter \\
     --restart unless-stopped \\
     -p 10341:10341 \\
     -v \$(pwd)/config:/config:ro \\
     --env-file $env_file \\
     ghcr.io/jontk/slurm-exporter:latest${NC}

4. Check logs:
   ${CYAN}docker logs -f slurm-exporter${NC}

5. Test metrics endpoint:
   ${CYAN}curl http://localhost:10341/metrics${NC}

Generated files:
  • $CONFIG_FILE - Configuration file
  • $env_file - Environment variables
EOF
}

generate_binary_instructions() {
    local service_file="slurm-exporter.service"
    local env_file="/etc/slurm-exporter/environment"
    
    # Generate systemd service
    cat > "$service_file" << EOF
[Unit]
Description=SLURM Exporter
Documentation=https://github.com/jontk/slurm-exporter
After=network.target

[Service]
Type=simple
User=slurm-exporter
Group=slurm-exporter
ExecStart=/usr/local/bin/slurm-exporter --config.file=/etc/slurm-exporter/config.yaml
Restart=on-failure
RestartSec=5s
StandardOutput=journal
StandardError=journal
SyslogIdentifier=slurm-exporter
KillMode=mixed
KillSignal=SIGTERM

# Security settings
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/slurm-exporter

# Environment
EnvironmentFile=-$env_file

[Install]
WantedBy=multi-user.target
EOF

    cat << EOF
${GREEN}🔧 Binary Deployment Instructions:${NC}

1. Install binary:
   ${CYAN}sudo curl -L https://github.com/jontk/slurm-exporter/releases/latest/download/slurm-exporter-linux-amd64 \\
     -o /usr/local/bin/slurm-exporter
   sudo chmod +x /usr/local/bin/slurm-exporter${NC}

2. Create user and directories:
   ${CYAN}sudo useradd --system --shell /bin/false slurm-exporter
   sudo mkdir -p /etc/slurm-exporter /var/lib/slurm-exporter
   sudo chown slurm-exporter:slurm-exporter /var/lib/slurm-exporter${NC}

3. Install configuration:
   ${CYAN}sudo cp $CONFIG_FILE /etc/slurm-exporter/config.yaml
   sudo chown root:slurm-exporter /etc/slurm-exporter/config.yaml
   sudo chmod 640 /etc/slurm-exporter/config.yaml${NC}

4. Create environment file:
   ${CYAN}sudo tee $env_file << 'EOF'
EOF

    case "$AUTH_TYPE" in
        "JWT Token")
            echo "$AUTH_TOKEN=your-actual-token-here"
            ;;
        "API Key")
            echo "$AUTH_TOKEN=your-actual-api-key-here"
            ;;
        "Basic Auth (username/password)")
            echo "$AUTH_PASS=your-actual-password-here"
            ;;
    esac

    cat << EOF
EOF
   sudo chmod 600 $env_file${NC}

5. Install systemd service:
   ${CYAN}sudo cp $service_file /etc/systemd/system/
   sudo systemctl daemon-reload
   sudo systemctl enable slurm-exporter
   sudo systemctl start slurm-exporter${NC}

6. Check status:
   ${CYAN}sudo systemctl status slurm-exporter
   journalctl -u slurm-exporter -f${NC}

7. Test metrics endpoint:
   ${CYAN}curl http://localhost:10341/metrics${NC}

Generated files:
  • $CONFIG_FILE - Configuration file
  • $service_file - systemd service unit
EOF
}

generate_dev_instructions() {
    cat << EOF
${GREEN}🛠️  Development Deployment Instructions:${NC}

1. Clone repository:
   ${CYAN}git clone https://github.com/jontk/slurm-exporter.git
   cd slurm-exporter${NC}

2. Install dependencies:
   ${CYAN}go mod download${NC}

3. Copy configuration:
   ${CYAN}cp $CONFIG_FILE configs/dev-config.yaml${NC}

4. Set environment variables:
EOF

    case "$AUTH_TYPE" in
        "JWT Token")
            echo "   ${CYAN}export $AUTH_TOKEN=your-actual-token-here${NC}"
            ;;
        "API Key")
            echo "   ${CYAN}export $AUTH_TOKEN=your-actual-api-key-here${NC}"
            ;;
        "Basic Auth (username/password)")
            echo "   ${CYAN}export $AUTH_PASS=your-actual-password-here${NC}"
            ;;
    esac

    cat << EOF

5. Run exporter:
   ${CYAN}go run ./cmd/slurm-exporter --config.file=configs/dev-config.yaml${NC}

6. Or build and run:
   ${CYAN}make build
   ./bin/slurm-exporter --config.file=configs/dev-config.yaml${NC}

7. Test metrics endpoint:
   ${CYAN}curl http://localhost:10341/metrics${NC}

Development commands:
  • ${CYAN}make test${NC} - Run tests
  • ${CYAN}make lint${NC} - Run linting
  • ${CYAN}make fmt${NC} - Format code

Generated files:
  • $CONFIG_FILE - Configuration file
EOF
}

step10_next_steps() {
    print_step "Next Steps"
    
    cat << EOF
${GREEN}🎉 Configuration Complete!${NC}

Your SLURM Exporter is ready to deploy. Here's what to do next:

${CYAN}📋 Immediate Actions:${NC}
  1. Review and test the generated configuration
  2. Replace placeholder values with actual credentials
  3. Deploy using the instructions above
  4. Verify metrics are being collected

${CYAN}📊 Monitoring Setup:${NC}
  1. Configure Prometheus to scrape metrics
  2. Import Grafana dashboards
  3. Set up alerting rules
  4. Monitor exporter health

${CYAN}🔧 Production Readiness:${NC}
  1. Enable TLS/SSL for secure communication
  2. Set up authentication for debug endpoints
  3. Configure log aggregation
  4. Implement backup and monitoring

${CYAN}📚 Resources:${NC}
  • Documentation: https://github.com/jontk/slurm-exporter/docs
  • Examples: https://github.com/jontk/slurm-exporter/examples
  • Support: https://github.com/jontk/slurm-exporter/issues

${CYAN}💡 Quick Validation:${NC}
After deployment, test these endpoints:
  • Health: http://your-host:10341/health
  • Metrics: http://your-host:10341/metrics
EOF

    if [[ " ${ADVANCED_FEATURES[*]} " =~ " debug_endpoints " ]]; then
        echo "  • Debug: http://your-host:10341/debug/status"
    fi

    echo ""
    print_success "Happy monitoring! 🚀"
}

# Main execution
main() {
    step1_welcome
    step2_deployment_type
    step3_environment
    step4_slurm_connection
    step5_authentication
    step6_collectors
    step7_advanced_features
    step8_generate_config
    step9_deployment_instructions
    step10_next_steps
}

# Trap for cleanup
trap 'echo -e "\n${YELLOW}Configuration wizard interrupted.${NC}"; exit 1' INT TERM

# Check dependencies
if ! command -v bc >/dev/null 2>&1; then
    print_error "bc (calculator) is required but not installed."
    exit 1
fi

# Run main function
main "$@"