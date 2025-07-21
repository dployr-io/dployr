#!/usr/bin/env bash

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="$SCRIPT_DIR/../../.env.dev"
CDN="https://cdn.dployr.io"
DATE=$(date +"%Y%m%d-%H%M%S")
INSTALL_START_TIME=$(date +%s)
DPLOYR_VERSION="latest"
RANDOM_SUBDOMAIN=$(openssl rand -hex 6)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Load environment variables if available
if [ -f "$ENV_FILE" ]; then
  export $(grep -v '^#' "$ENV_FILE" | xargs)
fi

# Progress bar function
show_progress() {
    local current=$1
    local total=$2
    local width=50
    local percentage=$((current * 100 / total))
    local completed=$((current * width / total))
    
    echo -n "["
    for ((i=0; i<completed; i++)); do echo -n "#"; done
    for ((i=completed; i<width; i++)); do echo -n "-"; done
    echo "] $percentage% ($current/$total)"
}

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running as root
check_sudo() {
    echo "Checking sudo privileges..."
    if [ "$EUID" -ne 0 ]; then
        log_error "This script must be run as root or with sudo privileges"
        echo "Please run: sudo $0"
        exit 1
    fi
    log_success "Running with sudo privileges"
}

# Create dployr user with safe admin permissions
create_dployr_user() {
    log_info "Creating dployr user..."
    
    # Create dployr user if it doesn't exist
    if ! id "dployr" &>/dev/null; then
        useradd -r -m -s /bin/bash -d /home/dployr -u 1000 dployr
        log_success "Created dployr user"
    else
        log_warning "User dployr already exists"
    fi
    
    # Add to docker group if Docker installation is selected
    if [ "$INSTALL_TYPE" = "docker" ]; then
        usermod -aG docker dployr
        log_info "Added dployr to docker group"
    fi
    
    # Create sudo rule for dployr (limited permissions)
    cat > /etc/sudoers.d/dployr << EOF
# Allow dployr to manage its own service and nginx
dployr ALL=(ALL) NOPASSWD: /bin/systemctl start dployr, /bin/systemctl stop dployr, /bin/systemctl restart dployr, /bin/systemctl reload nginx, /bin/systemctl restart nginx
EOF
    
    chmod 440 /etc/sudoers.d/dployr
    log_success "Configured safe sudo permissions for dployr"
}

# Download dployr binary
download_dployr() {
    log_info "Downloading dployr binary..."
    
    # Detect architecture
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) ARCH="amd64" ;;
        aarch64) ARCH="arm64" ;;
        armv7l) ARCH="arm" ;;
        *) log_error "Unsupported architecture: $ARCH"; exit 1 ;;
    esac
    
    # Create server directory
    mkdir -p /home/dployr/server
    
    # Download binary (assuming CDN has binaries)
    DOWNLOAD_URL="$CDN/releases/$DPLOYR_VERSION/dployr-linux-$ARCH"
    
    log_info "Downloading from: $DOWNLOAD_URL"
    if curl -fsSL "$DOWNLOAD_URL" -o /home/dployr/server/dployr; then
        chmod +x /home/dployr/server/dployr
        chown -R dployr:dployr /home/dployr
        log_success "Downloaded dployr binary"
    else
        log_error "Failed to download dployr binary"
        exit 1
    fi
}

# Installation type selection
select_install_type() {
    # Check for command line arguments first
    case "${1:-}" in
        --docker)
            INSTALL_TYPE="docker"
            log_info "Using Docker installation (from command line)"
            return
            ;;
        --standalone)
            INSTALL_TYPE="standalone" 
            log_info "Using Standalone installation (from command line)"
            return
            ;;
    esac
    
    # Check if we're in a pipe (non-interactive)
    if [ ! -t 0 ]; then
        INSTALL_TYPE="docker"
        log_warning "Non-interactive mode detected, defaulting to Docker installation"
        log_info "Use --docker or --standalone flags to specify installation type"
        return
    fi
    
    echo ""
    echo "╔══════════════════════════════════════╗"
    echo "║        DPLOYR.IO INSTALLER           ║"
    echo "╚══════════════════════════════════════╝"
    echo ""
    echo "Select installation type:"
    echo ""
    echo "  [1] Docker Installation (Recommended)"
    echo "      └─ Run dployr and apps in containers"
    echo ""
    echo "  [2] Standalone Installation"
    echo "      └─ Install dployr directly on host system"
    echo ""
    
    while true; do
        read -p "Enter your choice [1-2]: " choice
        case $choice in
            1)
                INSTALL_TYPE="docker"
                log_info "Selected: Docker Installation"
                break
                ;;
            2)
                INSTALL_TYPE="standalone"
                log_info "Selected: Standalone Installation"
                break
                ;;
            *)
                log_warning "Please select 1 or 2"
                ;;
        esac
    done
}

# Install system requirements
install_requirements() {
    log_info "Installing system requirements..."
    
    OS_TYPE=$(grep -w "ID" /etc/os-release | cut -d "=" -f 2 | tr -d '"')
    OS_VERSION=$(grep -w "VERSION_CODENAME" /etc/os-release | cut -d "=" -f 2 | tr -d '"')
    
    log_info "Detected OS: $OS_TYPE $OS_VERSION"
    
    export DEBIAN_FRONTEND=noninteractive
    export DEBCONF_NONINTERACTIVE_SEEN=true
    
    case "$OS_TYPE" in
        ubuntu|debian)
            apt-get update -qq
            
            PACKAGES="curl wget git jq nginx ca-certificates gnupg ufw openssl"
            if [ "$INSTALL_TYPE" = "docker" ]; then
                PACKAGES="$PACKAGES docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin"
                
                # Add Docker's official GPG key
                install -m 0755 -d /etc/apt/keyrings
                curl -fsSL https://download.docker.com/linux/$OS_TYPE/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
                chmod a+r /etc/apt/keyrings/docker.gpg
                
                # Add Docker repository
                echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$OS_TYPE $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
                apt-get update -qq
            fi
            
            apt-get install -y $PACKAGES
            ;;
        centos|rhel|rocky|alma)
            yum install -y curl wget git jq nginx ufw openssl
            if [ "$INSTALL_TYPE" = "docker" ]; then
                yum install -y yum-utils
                yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
                yum install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
            fi
            ;;
        *)
            log_error "Unsupported operating system: $OS_TYPE"
            exit 1
            ;;
    esac
    
    log_success "System requirements installed"
}

# Docker-specific setup
setup_docker() {
    if [ "$INSTALL_TYPE" != "docker" ]; then
        return
    fi
    
    log_info "Configuring Docker..."
    
    mkdir -p /etc/docker
    cat > /etc/docker/daemon.json << EOF
{
    "log-driver": "json-file",
    "log-opts": {
        "max-size": "100m",
        "max-file": "10"
    },
    "default-address-pools": [
        {"base": "172.17.0.0/12", "size": 20}
    ]
}
EOF
    
    systemctl enable docker
    timeout 60 systemctl start docker || {
        log_error "Docker service failed to start within 60 seconds"
        exit 1
    }
    
    if ! systemctl is-active --quiet docker; then
        log_error "Docker service failed to start"
        exit 1
    fi
    
    ufw allow 7879
    log_success "Docker configured successfully"
}

# Setup directories based on installation type
setup_directories() {
    log_info "Setting up directories..."
    
    if [ "$INSTALL_TYPE" = "docker" ]; then
        mkdir -p /data/dployr/{nextjs-apps,builds,images/cache,logs/{hot,warm,cold},monitoring/{prometheus,grafana},ssl,nginx/sites,redis}
        chown -R 1000:1000 /data/dployr
    else
        mkdir -p /home/dployr/{apps,builds,logs,ssl}
        mkdir -p /var/log
        touch /var/log/dployr.log
        chown -R dployr:dployr /home/dployr
        chown dployr:dployr /var/log/dployr.log
    fi
    
    log_success "Directories created"
}

# Setup Docker Compose (for Docker installation)
setup_docker_compose() {
    if [ "$INSTALL_TYPE" != "docker" ]; then
        return
    fi
    
    log_info "Setting up Docker Compose..."
    
    cat > /data/dployr/docker-compose.yml << EOF
services:
  dployr-web:
    image: dployr:latest
    user: "1000:1000" 
    ports: 
      - "7879:7879"
    volumes:
      - /data/dployr:/data
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - NODE_ENV=production
      - NEXT_TELEMETRY_DISABLED=1
    restart: unless-stopped
EOF
    
    chown dployr:dployr /data/dployr/docker-compose.yml
    log_success "Docker Compose configuration created"
}

# Create systemd service (for standalone installation)
create_systemd_service() {
    if [ "$INSTALL_TYPE" != "standalone" ]; then
        return
    fi
    
    log_info "Creating systemd service..."
    
    cat > /etc/systemd/system/dployr.service << EOF
[Unit]
Description=Dployr
After=network.target

[Service]
Type=simple
ExecStart=/home/dployr/server/dployr
Restart=on-failure
User=dployr
WorkingDirectory=/home/dployr/server
StandardOutput=append:/var/log/dployr.log
StandardError=inherit

[Install]
WantedBy=multi-user.target
EOF
    
    systemctl daemon-reload
    systemctl enable dployr
    log_success "Systemd service created and enabled"
}

# Setup Nginx and SSL
setup_nginx_ssl() {
    log_info "Setting up Nginx and SSL..."
    
    DOMAIN="$RANDOM_SUBDOMAIN.dployr.io"
    log_info "Using domain: $DOMAIN"
    
    # Install certbot
    if command -v apt-get &> /dev/null; then
        apt-get install -y certbot python3-certbot-nginx
    elif command -v yum &> /dev/null; then
        yum install -y certbot python3-certbot-nginx
    fi
    
    # Create Nginx configuration
    cat > /etc/nginx/sites-available/dployr << EOF
server {
    listen 80;
    server_name $DOMAIN;

    location / {
        proxy_pass http://localhost:7879;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF
    
    # Enable site
    ln -sf /etc/nginx/sites-available/dployr /etc/nginx/sites-enabled/
    
    # Remove default site if it exists
    rm -f /etc/nginx/sites-enabled/default
    
    # Test and reload nginx
    nginx -t && systemctl reload nginx
    
    log_info "Attempting to obtain SSL certificate for $DOMAIN..."
    
    # Try to get SSL certificate (will fail if domain doesn't resolve, but that's expected)
    if certbot --nginx -d "$DOMAIN" --non-interactive --agree-tos --email "admin@dployr.io" --quiet; then
        log_success "SSL certificate obtained for $DOMAIN"
    else
        log_warning "SSL certificate could not be obtained. You may need to configure DNS first."
        log_info "Manual command: certbot --nginx -d $DOMAIN"
    fi
    
    log_success "Nginx configured for domain: $DOMAIN"
}

# Start dployr service
start_dployr() {
    log_info "Starting dployr..."
    
    if [ "$INSTALL_TYPE" = "docker" ]; then
        cd /data/dployr
        docker compose up -d
        sleep 5
        if docker compose ps | grep -q "Up"; then
            log_success "Dployr started successfully (Docker)"
        else
            log_error "Failed to start dployr (Docker)"
            exit 1
        fi
    else
        systemctl start dployr
        sleep 3
        if systemctl is-active --quiet dployr; then
            log_success "Dployr started successfully (Systemd)"
        else
            log_error "Failed to start dployr (Systemd)"
            exit 1
        fi
    fi
}

# Calculate and show installation time
show_completion() {
    INSTALL_END_TIME=$(date +%s)
    INSTALL_DURATION=$((INSTALL_END_TIME - INSTALL_START_TIME))
    MINUTES=$((INSTALL_DURATION / 60))
    SECONDS=$((INSTALL_DURATION % 60))
    
    echo ""
    echo "╔══════════════════════════════════════╗"
    echo "║          INSTALLATION COMPLETE       ║"
    echo "╚══════════════════════════════════════╝"
    echo ""
    log_success "Installation completed in ${MINUTES}m ${SECONDS}s"
    echo ""
    echo "Access your dployr installation at:"
    echo "  HTTP:  http://$RANDOM_SUBDOMAIN.dployr.io"
    echo "  HTTPS: https://$RANDOM_SUBDOMAIN.dployr.io (if SSL configured)"
    echo ""
    echo "Service management:"
    if [ "$INSTALL_TYPE" = "docker" ]; then
        echo "  Start:   cd /data/dployr && docker compose up -d"
        echo "  Stop:    cd /data/dployr && docker compose down"
        echo "  Logs:    cd /data/dployr && docker compose logs -f"
    else
        echo "  Start:   sudo systemctl start dployr"
        echo "  Stop:    sudo systemctl stop dployr"
        echo "  Status:  sudo systemctl status dployr"
        echo "  Logs:    tail -f /var/log/dployr.log"
    fi
    echo ""
}

# Main installation flow
main() {
    echo "Starting dployr installer..."
    
    TOTAL_STEPS=10
    CURRENT_STEP=0
    
    # Step 1
    echo "Step 1: Checking privileges"
    check_sudo
    
    # Step 2
    echo "Step 2: Selecting installation type"
    select_install_type "$1"
    
    # Step 3
    ((CURRENT_STEP++))
    show_progress $CURRENT_STEP $TOTAL_STEPS
    create_dployr_user
    echo ""
    
    # Step 4
    ((CURRENT_STEP++))
    show_progress $CURRENT_STEP $TOTAL_STEPS
    download_dployr
    echo ""
    
    # Step 5
    ((CURRENT_STEP++))
    show_progress $CURRENT_STEP $TOTAL_STEPS
    install_requirements
    echo ""
    
    # Step 6
    ((CURRENT_STEP++))
    show_progress $CURRENT_STEP $TOTAL_STEPS
    setup_docker
    echo ""
    
    # Step 7
    ((CURRENT_STEP++))
    show_progress $CURRENT_STEP $TOTAL_STEPS
    setup_directories
    echo ""
    
    # Step 8
    ((CURRENT_STEP++))
    show_progress $CURRENT_STEP $TOTAL_STEPS
    setup_docker_compose
    create_systemd_service
    echo ""
    
    # Step 9
    ((CURRENT_STEP++))
    show_progress $CURRENT_STEP $TOTAL_STEPS
    setup_nginx_ssl
    echo ""
    
    # Step 10
    ((CURRENT_STEP++))
    show_progress $CURRENT_STEP $TOTAL_STEPS
    start_dployr
    echo ""
    
    show_completion
}

main "$@"
