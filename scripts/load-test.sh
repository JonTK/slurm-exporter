#!/bin/bash

# Load testing script for SLURM exporter
set -e

# Configuration
EXPORTER_URL="${EXPORTER_URL:-http://localhost:10341}"
METRICS_PATH="${METRICS_PATH:-/metrics}"
CONCURRENT_REQUESTS="${CONCURRENT_REQUESTS:-10}"
REQUEST_DURATION="${REQUEST_DURATION:-60s}"
RAMP_UP_TIME="${RAMP_UP_TIME:-10s}"
OUTPUT_DIR="${OUTPUT_DIR:-load-test-results}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}SLURM Exporter Load Testing${NC}"
echo "=================================="
echo -e "Target URL: ${BLUE}${EXPORTER_URL}${METRICS_PATH}${NC}"
echo -e "Concurrent requests: ${BLUE}${CONCURRENT_REQUESTS}${NC}"
echo -e "Duration: ${BLUE}${REQUEST_DURATION}${NC}"
echo -e "Ramp-up time: ${BLUE}${RAMP_UP_TIME}${NC}"
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Check if required tools are available
check_tool() {
    if ! command -v "$1" > /dev/null 2>&1; then
        echo -e "${RED}Error: $1 is not installed${NC}"
        echo "Please install $1 to run load tests"
        exit 1
    fi
}

# Check for load testing tools
echo -e "${YELLOW}Checking for load testing tools...${NC}"

# Check for curl (basic)
check_tool curl

# Check for Apache Bench (ab)
AB_AVAILABLE=false
if command -v ab > /dev/null 2>&1; then
    AB_AVAILABLE=true
    echo -e "${GREEN}✓ Apache Bench (ab) found${NC}"
fi

# Check for wrk
WRK_AVAILABLE=false
if command -v wrk > /dev/null 2>&1; then
    WRK_AVAILABLE=true
    echo -e "${GREEN}✓ wrk found${NC}"
fi

# Check for hey
HEY_AVAILABLE=false
if command -v hey > /dev/null 2>&1; then
    HEY_AVAILABLE=true
    echo -e "${GREEN}✓ hey found${NC}"
fi

# If no advanced tools available, suggest installation
if [ "$AB_AVAILABLE" = false ] && [ "$WRK_AVAILABLE" = false ] && [ "$HEY_AVAILABLE" = false ]; then
    echo -e "${YELLOW}No advanced load testing tools found. Installing hey...${NC}"

    # Try to install hey
    if command -v go > /dev/null 2>&1; then
        go install github.com/rakyll/hey@latest
        if [ $? -eq 0 ]; then
            HEY_AVAILABLE=true
            echo -e "${GREEN}✓ hey installed successfully${NC}"
        fi
    else
        echo -e "${YELLOW}Go not found. Will use basic curl-based testing${NC}"
    fi
fi

# Function to test endpoint availability
test_endpoint() {
    echo -e "${YELLOW}Testing endpoint availability...${NC}"

    if curl -s -f "${EXPORTER_URL}${METRICS_PATH}" > /dev/null; then
        echo -e "${GREEN}✓ Endpoint is responding${NC}"
        return 0
    else
        echo -e "${RED}✗ Endpoint is not responding${NC}"
        echo "Please ensure the SLURM exporter is running at ${EXPORTER_URL}"
        return 1
    fi
}

# Function to get baseline metrics
get_baseline() {
    echo -e "${YELLOW}Collecting baseline metrics...${NC}"

    local baseline_file="$OUTPUT_DIR/baseline.txt"
    curl -s "${EXPORTER_URL}${METRICS_PATH}" > "$baseline_file"

    if [ -s "$baseline_file" ]; then
        local metric_count=$(grep -c "^slurm_" "$baseline_file" || echo "0")
        local response_size=$(wc -c < "$baseline_file")

        echo -e "${GREEN}✓ Baseline collected${NC}"
        echo "  - Metrics count: $metric_count"
        echo "  - Response size: ${response_size} bytes"

        # Store baseline info
        echo "metric_count=${metric_count}" > "$OUTPUT_DIR/baseline_info.txt"
        echo "response_size=${response_size}" >> "$OUTPUT_DIR/baseline_info.txt"
    else
        echo -e "${RED}✗ Failed to collect baseline${NC}"
        return 1
    fi
}

# Function to run basic curl test
run_curl_test() {
    echo -e "${YELLOW}Running basic curl load test...${NC}"

    local output_file="$OUTPUT_DIR/curl_test.log"
    local concurrent=${1:-10}
    local duration=${2:-60}

    echo "Running $concurrent concurrent requests for ${duration}s"

    {
        echo "# Curl Load Test Results"
        echo "# Started: $(date)"
        echo "# URL: ${EXPORTER_URL}${METRICS_PATH}"
        echo "# Concurrent: $concurrent"
        echo "# Duration: ${duration}s"
        echo ""
    } > "$output_file"

    # Run concurrent curl requests
    local pids=()
    local start_time=$(date +%s)
    local end_time=$((start_time + duration))

    for ((i=1; i<=concurrent; i++)); do
        {
            local request_count=0
            local success_count=0
            local error_count=0

            while [ $(date +%s) -lt $end_time ]; do
                local req_start=$(date +%s.%3N)

                if curl -s -f -o /dev/null "${EXPORTER_URL}${METRICS_PATH}"; then
                    success_count=$((success_count + 1))
                else
                    error_count=$((error_count + 1))
                fi

                local req_end=$(date +%s.%3N)
                local req_time=$(echo "$req_end - $req_start" | bc -l 2>/dev/null || echo "0")

                request_count=$((request_count + 1))

                echo "Worker-$i: Request $request_count, Time: ${req_time}s, Success: $success_count, Errors: $error_count"
            done

            echo "Worker-$i: Total requests: $request_count, Successes: $success_count, Errors: $error_count"
        } >> "$output_file" &

        pids+=($!)
    done

    # Wait for all workers to finish
    for pid in "${pids[@]}"; do
        wait "$pid"
    done

    echo -e "${GREEN}✓ Curl test completed${NC}"
    echo "Results saved to: $output_file"
}

# Function to run Apache Bench test
run_ab_test() {
    echo -e "${YELLOW}Running Apache Bench load test...${NC}"

    local output_file="$OUTPUT_DIR/ab_test.log"
    local concurrent=${1:-10}
    local requests=${2:-1000}

    echo "Running $requests requests with $concurrent concurrent connections"

    ab -n "$requests" -c "$concurrent" -g "$OUTPUT_DIR/ab_gnuplot.tsv" \
       "${EXPORTER_URL}${METRICS_PATH}" > "$output_file" 2>&1

    echo -e "${GREEN}✓ Apache Bench test completed${NC}"
    echo "Results saved to: $output_file"

    # Extract key metrics
    if [ -f "$output_file" ]; then
        echo ""
        echo -e "${BLUE}Key Results:${NC}"
        grep "Requests per second:" "$output_file" || echo "RPS: Not available"
        grep "Time per request:" "$output_file" | head -1 || echo "Response time: Not available"
        grep "Transfer rate:" "$output_file" || echo "Transfer rate: Not available"
    fi
}

# Function to run wrk test
run_wrk_test() {
    echo -e "${YELLOW}Running wrk load test...${NC}"

    local output_file="$OUTPUT_DIR/wrk_test.log"
    local concurrent=${1:-10}
    local duration=${2:-60s}

    echo "Running for $duration with $concurrent connections"

    wrk -t4 -c"$concurrent" -d"$duration" --latency \
        "${EXPORTER_URL}${METRICS_PATH}" > "$output_file" 2>&1

    echo -e "${GREEN}✓ wrk test completed${NC}"
    echo "Results saved to: $output_file"

    # Extract key metrics
    if [ -f "$output_file" ]; then
        echo ""
        echo -e "${BLUE}Key Results:${NC}"
        grep "Requests/sec:" "$output_file" || echo "RPS: Not available"
        grep "Latency" "$output_file" || echo "Latency: Not available"
        grep "Transfer/sec:" "$output_file" || echo "Transfer rate: Not available"
    fi
}

# Function to run hey test
run_hey_test() {
    echo -e "${YELLOW}Running hey load test...${NC}"

    local output_file="$OUTPUT_DIR/hey_test.log"
    local concurrent=${1:-10}
    local duration=${2:-60s}

    echo "Running for $duration with $concurrent workers"

    hey -z "$duration" -c "$concurrent" -o csv \
        "${EXPORTER_URL}${METRICS_PATH}" > "$OUTPUT_DIR/hey_raw.csv" 2>"$output_file"

    # Also run a summary
    hey -z "$duration" -c "$concurrent" \
        "${EXPORTER_URL}${METRICS_PATH}" >> "$output_file" 2>&1

    echo -e "${GREEN}✓ hey test completed${NC}"
    echo "Results saved to: $output_file"

    # Extract key metrics
    if [ -f "$output_file" ]; then
        echo ""
        echo -e "${BLUE}Key Results:${NC}"
        grep "Requests/sec:" "$output_file" || echo "RPS: Not available"
        grep "Average:" "$output_file" || echo "Average response time: Not available"
        grep "Slowest:" "$output_file" || echo "Slowest response: Not available"
    fi
}

# Function to monitor system resources during test
monitor_resources() {
    local output_file="$OUTPUT_DIR/system_monitor.log"
    local duration=${1:-60}

    echo -e "${YELLOW}Starting system resource monitoring...${NC}"

    {
        echo "# System Resource Monitoring"
        echo "# Started: $(date)"
        echo "# Duration: ${duration}s"
        echo ""
        echo "timestamp,cpu_percent,memory_mb,load_1m,load_5m,load_15m"
    } > "$output_file"

    local start_time=$(date +%s)
    local end_time=$((start_time + duration))

    while [ $(date +%s) -lt $end_time ]; do
        local timestamp=$(date +%s)
        local cpu_usage=$(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1 || echo "0")
        local memory_usage=$(free -m | awk 'NR==2{printf "%.0f", $3}')
        local load_avg=$(uptime | awk -F'load average:' '{print $2}' | awk '{gsub(/,/, "", $1); gsub(/,/, "", $2); gsub(/,/, "", $3); print $1","$2","$3}')

        echo "${timestamp},${cpu_usage},${memory_usage},${load_avg}" >> "$output_file"

        sleep 5
    done &

    MONITOR_PID=$!
}

# Function to stop resource monitoring
stop_monitoring() {
    if [ ! -z "$MONITOR_PID" ]; then
        kill "$MONITOR_PID" 2>/dev/null || true
        wait "$MONITOR_PID" 2>/dev/null || true
        echo -e "${GREEN}✓ Resource monitoring stopped${NC}"
    fi
}

# Function to generate summary report
generate_report() {
    echo -e "${YELLOW}Generating load test report...${NC}"

    local report_file="$OUTPUT_DIR/load_test_report.md"

    {
        echo "# SLURM Exporter Load Test Report"
        echo ""
        echo "**Generated:** $(date)"
        echo "**Target:** ${EXPORTER_URL}${METRICS_PATH}"
        echo ""

        echo "## Test Configuration"
        echo "- Concurrent requests: $CONCURRENT_REQUESTS"
        echo "- Duration: $REQUEST_DURATION"
        echo "- Ramp-up time: $RAMP_UP_TIME"
        echo ""

        echo "## Baseline Metrics"
        if [ -f "$OUTPUT_DIR/baseline_info.txt" ]; then
            source "$OUTPUT_DIR/baseline_info.txt"
            echo "- Metric count: $metric_count"
            echo "- Response size: $response_size bytes"
        else
            echo "- Baseline data not available"
        fi
        echo ""

        echo "## Test Results"

        # Include results from different tools
        for tool in ab wrk hey curl; do
            local result_file="$OUTPUT_DIR/${tool}_test.log"
            if [ -f "$result_file" ]; then
                echo ""
                echo "### $tool Results"
                echo "\`\`\`"
                tail -20 "$result_file"
                echo "\`\`\`"
            fi
        done

        echo ""
        echo "## System Resources"
        if [ -f "$OUTPUT_DIR/system_monitor.log" ]; then
            echo "Resource monitoring data available in system_monitor.log"
        else
            echo "No system resource data collected"
        fi

    } > "$report_file"

    echo -e "${GREEN}✓ Report generated: $report_file${NC}"
}

# Main execution
main() {
    echo -e "${YELLOW}Starting load testing sequence...${NC}"

    # Test endpoint
    if ! test_endpoint; then
        exit 1
    fi

    # Get baseline
    get_baseline

    # Convert duration to seconds for monitoring
    local duration_seconds
    case "$REQUEST_DURATION" in
        *s) duration_seconds=${REQUEST_DURATION%s} ;;
        *m) duration_seconds=$((${REQUEST_DURATION%m} * 60)) ;;
        *) duration_seconds=60 ;;
    esac

    # Start resource monitoring
    monitor_resources "$duration_seconds"

    # Run tests based on available tools
    if [ "$HEY_AVAILABLE" = true ]; then
        run_hey_test "$CONCURRENT_REQUESTS" "$REQUEST_DURATION"
    elif [ "$WRK_AVAILABLE" = true ]; then
        run_wrk_test "$CONCURRENT_REQUESTS" "$REQUEST_DURATION"
    elif [ "$AB_AVAILABLE" = true ]; then
        local total_requests=$((CONCURRENT_REQUESTS * duration_seconds))
        run_ab_test "$CONCURRENT_REQUESTS" "$total_requests"
    else
        run_curl_test "$CONCURRENT_REQUESTS" "$duration_seconds"
    fi

    # Stop monitoring
    stop_monitoring

    # Generate report
    generate_report

    echo ""
    echo -e "${GREEN}Load testing completed!${NC}"
    echo -e "Results available in: ${BLUE}$OUTPUT_DIR/${NC}"
}

# Handle script interruption
trap 'echo -e "\n${YELLOW}Load test interrupted${NC}"; stop_monitoring; exit 1' INT TERM

# Run main function
main "$@"