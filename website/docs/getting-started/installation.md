# Installation

This guide covers all installation methods for SLURM Exporter. Choose the method that best fits your environment and requirements.

## System Requirements

### Minimum Requirements
- **CPU**: 1 core
- **Memory**: 512MB RAM
- **Storage**: 100MB disk space
- **Network**: HTTP/HTTPS access to SLURM REST API

### Recommended for Production
- **CPU**: 2+ cores
- **Memory**: 2GB+ RAM
- **Storage**: 1GB+ disk space (for logs and caching)
- **Network**: Low-latency connection to SLURM controller

### SLURM Compatibility
- **SLURM REST API**: Enabled with authentication (slurmrestd)
- **Supported API Versions**: v0.0.40, v0.0.41, v0.0.42, v0.0.43, v0.0.44

## Docker {#docker}

Docker deployment is perfect for development, testing, and simple production setups.

### Quick Start

```bash
docker run -d \
  --name slurm-exporter \
  --restart unless-stopped \
  -p 10341:10341 \
  -v $(pwd)/config.yaml:/etc/slurm-exporter/config.yaml:ro \
  ghcr.io/jontk/slurm-exporter:latest
```

### Docker Compose

Create a `docker-compose.yml` file:

```yaml title="docker-compose.yml"
version: '3.8'

services:
  slurm-exporter:
    image: ghcr.io/jontk/slurm-exporter:latest
    container_name: slurm-exporter
    restart: unless-stopped
    ports:
      - "10341:10341"
    volumes:
      - ./config.yaml:/etc/slurm-exporter/config.yaml:ro
    healthcheck:
      test: ["CMD", "/usr/local/bin/slurm-exporter", "--health-check"]
      interval: 30s
      timeout: 10s
      retries: 3
      start_period: 40s

  # Optional: Include Prometheus for testing
  prometheus:
    image: prom/prometheus:latest
    container_name: prometheus
    ports:
      - "9090:9090"
    volumes:
      - ./prometheus.yml:/etc/prometheus/prometheus.yml:ro
    command:
      - '--config.file=/etc/prometheus/prometheus.yml'
      - '--storage.tsdb.path=/prometheus'
      - '--web.console.libraries=/etc/prometheus/console_libraries'
      - '--web.console.templates=/etc/prometheus/consoles'
```

Start the services:

```bash
docker-compose up -d
```

### Configuration File

Create a configuration file for the exporter:

```bash
# Create configuration file
cat > ./config.yaml << 'EOF'
server:
  address: ":10341"
  metrics_path: "/metrics"
  health_path: "/health"
  ready_path: "/ready"

slurm:
  base_url: "http://your-slurm-controller.example.com:6820"
  api_version: "v0.0.44"
  auth:
    type: "jwt"
    username: "root"
    token: "your-jwt-token"
  timeout: 30s
  retry_attempts: 3
  retry_delay: 5s

collectors:
  jobs:
    enabled: true
  nodes:
    enabled: true
  partitions:
    enabled: true

logging:
  level: "info"
  format: "json"
EOF

# Run with mounted config
docker run -d \
  --name slurm-exporter \
  --restart unless-stopped \
  -p 10341:10341 \
  -v $(pwd)/config.yaml:/etc/slurm-exporter/config.yaml:ro \
  ghcr.io/jontk/slurm-exporter:latest
```

## Package Installation {#packages}

System packages provide easy installation and automatic updates through your distribution's package manager.

### RPM Packages (RHEL/CentOS/Rocky/AlmaLinux)

```bash
# Add repository
curl -fsSL https://packages.slurm-exporter.io/rpm/slurm-exporter.repo \
  | sudo tee /etc/yum.repos.d/slurm-exporter.repo

# Install package
sudo dnf install slurm-exporter
# or for older systems: sudo yum install slurm-exporter
```

### DEB Packages (Ubuntu/Debian)

```bash
# Add repository key
curl -fsSL https://packages.slurm-exporter.io/deb/slurm-exporter.gpg \
  | sudo gpg --dearmor -o /usr/share/keyrings/slurm-exporter.gpg

# Add repository
echo "deb [signed-by=/usr/share/keyrings/slurm-exporter.gpg] https://packages.slurm-exporter.io/deb $(lsb_release -cs) main" \
  | sudo tee /etc/apt/sources.list.d/slurm-exporter.list

# Update and install
sudo apt update
sudo apt install slurm-exporter
```

### Configuration

After package installation, configure the service:

```bash
# Edit configuration
sudo nano /etc/slurm-exporter/config.yaml

# Enable and start service
sudo systemctl enable slurm-exporter
sudo systemctl start slurm-exporter

# Check status
sudo systemctl status slurm-exporter
```

### Service Management

```bash
# Start service
sudo systemctl start slurm-exporter

# Stop service
sudo systemctl stop slurm-exporter

# Restart service
sudo systemctl restart slurm-exporter

# Enable auto-start
sudo systemctl enable slurm-exporter

# View logs
sudo journalctl -u slurm-exporter -f
```

## Binary Installation {#binary}

Download pre-compiled binaries for quick setup or air-gapped environments.

### Download

```bash
# Set version and architecture
VERSION="v1.0.0"
ARCH="linux-amd64"  # or linux-arm64

# Download and extract
curl -LO "https://github.com/jontk/slurm-exporter/releases/download/${VERSION}/slurm-exporter-${ARCH}.tar.gz"
tar xzf "slurm-exporter-${ARCH}.tar.gz"

# Make executable and move to PATH
chmod +x slurm-exporter
sudo mv slurm-exporter /usr/local/bin/
```

### Create System User

```bash
# Create dedicated user
sudo useradd --system --shell /bin/false --home-dir /var/lib/slurm-exporter slurm-exporter

# Create directories
sudo mkdir -p /etc/slurm-exporter /var/lib/slurm-exporter /var/log/slurm-exporter
sudo chown slurm-exporter:slurm-exporter /var/lib/slurm-exporter /var/log/slurm-exporter
```

### Create Systemd Service

```bash
sudo tee /etc/systemd/system/slurm-exporter.service << 'EOF'
[Unit]
Description=SLURM Exporter
Documentation=https://jontk.github.io/slurm-exporter
After=network.target

[Service]
Type=simple
User=slurm-exporter
Group=slurm-exporter
ExecStart=/usr/local/bin/slurm-exporter --config=/etc/slurm-exporter/config.yaml
ExecReload=/bin/kill -HUP $MAINPID
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal
SyslogIdentifier=slurm-exporter

# Security settings
NoNewPrivileges=yes
PrivateTmp=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=/var/lib/slurm-exporter
CapabilityBoundingSet=

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd and enable service
sudo systemctl daemon-reload
sudo systemctl enable slurm-exporter
```

### Configure and Start

```bash
# Create basic configuration
sudo tee /etc/slurm-exporter/config.yaml << 'EOF'
server:
  address: ":10341"
  metrics_path: "/metrics"
  health_path: "/health"
  ready_path: "/ready"

slurm:
  base_url: "http://your-slurm-controller.example.com:6820"
  api_version: "v0.0.44"
  auth:
    type: "jwt"
    username: "root"
    token: "your-jwt-token"
  timeout: 30s
  retry_attempts: 3
  retry_delay: 5s

collectors:
  jobs:
    enabled: true
  nodes:
    enabled: true
  partitions:
    enabled: true

logging:
  level: "info"
  format: "json"
EOF

# Set permissions
sudo chown root:slurm-exporter /etc/slurm-exporter/config.yaml
sudo chmod 640 /etc/slurm-exporter/config.yaml

# Start service
sudo systemctl start slurm-exporter
```

## Build from Source {#source}

Build from source for development, custom features, or unsupported platforms.

### Prerequisites

- Go 1.21+
- Git
- Make

### Build Process

```bash
# Clone repository
git clone https://github.com/jontk/slurm-exporter.git
cd slurm-exporter

# Build binary
make build

# Build for specific platform
make build GOOS=linux GOARCH=amd64

# Run tests
make test

# Generate coverage report
make coverage
```

### Development Setup

```bash
# Install development dependencies
make dev-deps

# Run with live reload
make run-dev

# Format code
make fmt

# Run linting
make lint
```

## Post-Installation

### Verify Installation

After installation with any method, verify SLURM Exporter is working:

```bash
# Check metrics endpoint
curl http://localhost:10341/metrics

# Check health endpoint
curl http://localhost:10341/health

# Check readiness endpoint
curl http://localhost:10341/ready

# Check specific metric
curl -s http://localhost:10341/metrics | grep slurm_
```

### Initial Configuration

1. **Update configuration** with your SLURM details (base_url, auth token)
2. **Test connection** to SLURM REST API
3. **Enable desired collectors** based on your needs
4. **Configure Prometheus** to scrape metrics
5. **Set up Grafana dashboards** for visualization

### Next Steps

- [-> Quick Start Guide](quickstart.md) - Get basic monitoring running
- [-> Configuration Reference](../user-guide/configuration.md) - Detailed configuration options
- [-> Prometheus Integration](../integration/prometheus.md) - Set up metric collection
- [-> Grafana Dashboards](../integration/grafana.md) - Visualize your metrics

## Troubleshooting

### Common Issues

**Connection refused**
```bash
# Check if service is running
systemctl status slurm-exporter

# Check logs
journalctl -u slurm-exporter -f
```

**Authentication failed**
```bash
# Verify token against SLURM REST API
curl -H "X-SLURM-USER-NAME: your-user" \
     -H "X-SLURM-USER-TOKEN: your-token" \
     http://your-slurm-host:6820/slurm/v0.0.44/ping
```

**No metrics**
```bash
# Check exporter logs for errors
journalctl -u slurm-exporter -n 50

# Test with debug logging
slurm-exporter --config=/etc/slurm-exporter/config.yaml --log-level=debug
```

For more troubleshooting information, visit our [Troubleshooting Guide](../user-guide/troubleshooting.md).
