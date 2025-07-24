#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HOME_DIR=$(getent passwd "$USER" | cut -d: -f6)
STATE_DIR="$HOME_DIR/.dployr/state"
mkdir -p "$STATE_DIR"
ENV_FILE="$SCRIPT_DIR/../../.env.dev"
CDN="https://github.com/tobimadehin/dployr"
DATE=$(date +"%Y%m%d-%H%M%S")
INSTALL_START_TIME=$(date +%s)
DPLOYR_VERSION="latest"
RANDOM_SUBDOMAIN=$(openssl rand -hex 6)
SERVER_IP=$(curl -s https://api.ipify.org)

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

# Progress bar function with in-place updates
show_progress() {
    local current=$1
    local total=$2
    local message="$3"
    local width=50
    local percentage=$((current * 100 / total))
    local completed=$((current * width / total))
    
    # Clear line and move cursor to beginning
    printf "\r\033[K"
    
    # Show progress bar
    printf "["
    for ((i=0; i<completed; i++)); do printf "#"; done
    for ((i=completed; i<width; i++)); do printf "-"; done
    printf "] %3d%% %s" "$percentage" "$message"
    
    # Add newline only when complete
    if [ "$current" -eq "$total" ]; then
        printf "\n"
    fi
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

# Standard error handler for all functions
handle_function_error() {
    local function_name="$1"
    local error_msg="$2"
    
    {
        printf "\n\n"
        echo -e "${RED}[ERROR]${NC} $function_name failed"
        if [ -n "$error_msg" ]; then
            echo "Error: $error_msg"
        fi
        echo ""
        echo "Full log: $LOG_FILE"
        echo ""
    } >/dev/tty
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
    local flag_file="$STATE_DIR/create_dployr_user.flag"
    if [ -f "$flag_file" ]; then
        log_info "dployr user already created. Skipping."
        return 0
    fi

    log_info "Creating dployr user..."
    
    # Create dployr user if it doesn't exist
    if ! id "dployr" &>/dev/null; then
        useradd -r -m -s /bin/bash -d /home/dployr dployr
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
    local flag_file="$STATE_DIR/download_dployr.flag"
    if [ -f "$flag_file" ]; then
        log_info "dployr binary already downloaded. Skipping."
        return 0
    fi

    log_info "Downloading dployr binary..."
    
    # Detect architecture
    ARCH=$(uname -m)
    case $ARCH in
        x86_64) ARCH="x86_64" ;;
        aarch64) ARCH="arm64" ;;
        armv7l) ARCH="arm" ;;
        *) log_error "Unsupported architecture: $ARCH"; exit 1 ;;
    esac
    
    # Create server directory
    mkdir -p /home/dployr/server
    
    # Download binary
    DOWNLOAD_URL="$CDN/releases/$DPLOYR_VERSION/download/dployr_Linux_$ARCH.tar.gz"
    
    log_info "Downloading from: $DOWNLOAD_URL"
    if curl -fsSL "$DOWNLOAD_URL" | tar -xz -C /home/dployr/server; then
        chmod +x /home/dployr/server/dployr
        chown -R dployr:dployr /home/dployr
        log_success "Downloaded dployr binary"
    else
        handle_error "Download error" "Error occoured while downloading dployr binary"
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
    echo "║         DPLOYR.IO INSTALLER          ║"
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
    local flag_file="$STATE_DIR/install_requirements.flag"
    if [ -f "$flag_file" ]; then
        log_info "dployr binary already downloaded. Skipping."
        return 0
    fi

    log_info "Installing system requirements..."
    
    OS_TYPE=$(grep -w "ID" /etc/os-release | cut -d "=" -f 2 | tr -d '"')
    OS_VERSION=$(grep -w "VERSION_CODENAME" /etc/os-release | cut -d "=" -f 2 | tr -d '"')
    
    log_info "Detected OS: $OS_TYPE $OS_VERSION"
    
    export DEBIAN_FRONTEND=noninteractive
    export DEBCONF_NONINTERACTIVE_SEEN=true
    
    case "$OS_TYPE" in
        ubuntu|debian)
            apt-get update -qq

            # Install Caddy
            apt install -y debian-keyring debian-archive-keyring apt-transport-https
            curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
            curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
            apt-get update -qq
            
            PACKAGES="curl wget git jq caddy ca-certificates gnupg ufw openssl net-tools"
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
            # Install Caddy
            yum install -y yum-plugin-copr
            yum copr enable -y @caddy/caddy

            yum install -y curl wget git jq caddy ufw openssl
            if [ "$INSTALL_TYPE" = "docker" ]; then
                yum install -y yum-utils
                yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
                yum install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
            fi
            ;;
        *)
            handle_error "Package error" "Error occoured while installing required packages" 
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

    local flag_file="$STATE_DIR/setup_docker.flag"
    if [ -f "$flag_file" ]; then
        log_info "Docker already configured. Skipping."
        return 0
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
        handle_error "Docker setup error" "Docker service failed to start within 60 seconds"
        exit 1
    }
    
    if ! systemctl is-active --quiet docker; then
        handle_error "Docker setup error" "Docker service failed to start within 60 seconds"
        exit 1
    fi
    
    ufw allow 7879
    log_success "Docker configured successfully"
}

# Setup directories based on installation type
setup_directories() {
    local flag_file="$STATE_DIR/setup_directories.flag"
    if [ -f "$flag_file" ]; then
        log_info "Directories already created. Skipping."
        return 0
    fi
    
    log_info "Setting up directories..."
    
    if [ "$INSTALL_TYPE" = "docker" ]; then
        mkdir -p /data/dployr/{nextjs-apps,builds,images/cache,logs/{hot,warm,cold},monitoring/{prometheus,grafana},ssl,nginx/sites,redis}
        chown -R dployr:dployr /data/dployr
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
    
    local flag_file="$STATE_DIR/setup_docker_compose.flag"
    if [ -f "$flag_file" ]; then
        log_info "Docker Compose already configured. Skipping."
        return 0
    fi

    log_info "Setting up Docker Compose..."
    
    cat > /data/dployr/docker-compose.yml << EOF
services:
  dployr-web:
    image: dployr:latest
    user: "dployr:dployr" 
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
    
    log_success "Docker Compose configuration created"
}

# Create systemd service (for standalone installation)
create_systemd_service() {
    if [ "$INSTALL_TYPE" != "standalone" ]; then
        return
    fi

    local flag_file="$STATE_DIR/create_systemd_service.flag"
    if [ -f "$flag_file" ]; then
        log_info "Systemd service already created. Skipping."
        return 0
    fi
    
    log_info "Creating systemd service..."
    
    cat > /etc/systemd/system/dployr.service << EOF
[Unit]
Description=dployr 
After=network.target

[Service]
Type=simple
ExecStart=/home/dployr/server/dployr
Restart=on-failure
User=dployr
Group=dployr
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

# Create Cloudflare DNS record
create_cloudflare_dns() {
    local subdomain="$RANDOM_SUBDOMAIN"
    
    log_info "Creating DNS record for $SERVER_IP"
    
    # Create A record via Cloudflare API
    local response=$(curl -s -X POST "https://dployr.dev/api/dns/create" \
        -H "Content-Type: application/json" \
        --data '{
            "subdomain": "'"$subdomain"'",
            "host": "'"$SERVER_IP"'"
        }')
    
    # Check if the request was successful
    local success=$(echo $response | jq -r '.success // false')
    
    if [ "$success" = "true" ]; then
        log_success "DNS record created successfully for $subdomain.dployr.dev"
        return 0
    else
        local error_message=$(echo $response | jq -r '.errors.message // "Unknown error"')
        return 1
    fi
}

# Setup Caddy
setup_caddy() {
    local flag_file="$STATE_DIR/setup_caddy.flag"
    if [ -f "$flag_file" ]; then
        log_info "Caddy already configured. Skipping."
        return 0
    fi

    log_info "Setting up Caddy with automatic HTTPS..."
    
    # Create subdomain DNS record
    if ! create_cloudflare_dns; then
        handle_error "DNS setup error" "Failed to create DNS record"
        exit 1
    fi
    
    DOMAIN="$RANDOM_SUBDOMAIN.dployr.dev"
    log_info "Using domain: $DOMAIN"
    
    # Create Caddyfile
    cat > /etc/caddy/Caddyfile << EOF
$DOMAIN {
    reverse_proxy localhost:7879
    
    # Enable gzip compression
    encode gzip
    
    # Security headers
    header {
        X-Content-Type-Options nosniff
        X-Frame-Options DENY
        X-XSS-Protection "1; mode=block"
    }
}
EOF
    
    # Set proper permissions
    chown caddy:caddy /etc/caddy/Caddyfile
    chmod 644 /etc/caddy/Caddyfile
    
    # Open firewall ports
    ufw allow 80
    ufw allow 443
    
    log_info "Waiting for DNS propagation (30 seconds)..."
    sleep 30
    
    # Reload Caddy configuration
    systemctl reload caddy
    
    # Wait a bit for certificate provisioning
    log_info "Waiting for automatic SSL certificate provisioning..."
    sleep 15
    
    # Check if Caddy is running properly
    if systemctl is-active --quiet caddy; then
        log_success "Caddy configured successfully for domain: $DOMAIN"
        log_info "HTTPS certificate will be automatically provisioned by Caddy"
    else
        handle_error "Caddy setup error" "Caddy service failed to start properly"
        exit 1
    fi
    
    touch "$flag_file"
}

# Start dployr service
start_dployr() {
    local flag_file="$STATE_DIR/start_dployr.flag"
    if [ -f "$flag_file" ]; then
        log_info "dployr already started. Skipping."
        return 0
    fi
    
    log_info "Starting dployr..."
    
    if [ "$INSTALL_TYPE" = "docker" ]; then
        cd /data/dployr
        docker compose up -d
        sleep 5
        if docker compose ps | grep -q "Up"; then
            log_success "Dployr started successfully (Docker)"
        else
            handle_error "Program error" "Failed to start dployr (Docker)"
            exit 1
        fi
    else
        systemctl start dployr
        sleep 3
        if systemctl is-active --quiet dployr; then
            log_success "Dployr started successfully (Systemd)"
        else
            handle_error "Program error" "Failed to start dployr (Systemd)"
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
    echo "║         INSTALLATION COMPLETE        ║"
    echo "╚══════════════════════════════════════╝"
    echo ""
    log_success "Installation completed in ${MINUTES}m ${SECONDS}s"
    echo ""
    echo "Access your dployr installation at:"
    echo "  https://$RANDOM_SUBDOMAIN.dployr.dev"
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

# Save original stdout/stderr
exec 3>&1 4>&2

# Function to show errors on console regardless of redirection
handle_error() {
    local step_name="$1"
    local error_msg="$2"
    
    {
        printf "\n\n"
        echo -e "${RED}[ERROR]${NC} Installation failed during: $step_name"
        if [ -n "$error_msg" ]; then
            echo "Error: $error_msg"
        fi
        echo ""
        echo "Last 20 lines from install log:"
        echo "================================"
        sync
        if [ -f "$LOG_FILE" ] && [ -s "$LOG_FILE" ]; then
            tail -20 "$LOG_FILE"
        else
            echo "No log content available"
        fi
        echo "================================"
        echo ""
        echo "Full log available at: $LOG_FILE"
        echo ""
    } >/dev/tty 2>&1
}

# Main installation flow
main() {
    echo "Starting dployr installer..."
    echo ""

    LOG_FILE="/tmp/dployr-install-$(date +%Y%m%d-%H%M%S).log"
    echo "Installation started at $(date)" > "$LOG_FILE"
    
    check_sudo
    select_install_type "$1"
    
    TOTAL_STEPS=8
    CURRENT_STEP=0
    
    # Step 1: Create user
    show_progress $CURRENT_STEP $TOTAL_STEPS "Creating user..."
    if ! create_dployr_user >> "$LOG_FILE" 2>&1; then
        exit 1
    fi
    
    ((CURRENT_STEP++))
    show_progress $CURRENT_STEP $TOTAL_STEPS "Downloading binary..."
    if ! download_dployr >> "$LOG_FILE" 2>&1; then
        exit 1
    fi

    ((CURRENT_STEP++))
    show_progress $CURRENT_STEP $TOTAL_STEPS "Installing requirements..."
    if ! install_requirements >> "$LOG_FILE" 2>&1; then
        exit 1
    fi
    
    ((CURRENT_STEP++))
    show_progress $CURRENT_STEP $TOTAL_STEPS "Setting up Docker..."
    if ! setup_docker >> "$LOG_FILE" 2>&1; then
        exit 1
    fi
    
    ((CURRENT_STEP++))
    show_progress $CURRENT_STEP $TOTAL_STEPS "Creating directories..."
    if ! setup_directories >> "$LOG_FILE" 2>&1; then
        exit 1
    fi
    
    ((CURRENT_STEP++))
    show_progress $CURRENT_STEP $TOTAL_STEPS "Configuring services..."
    if ! setup_docker_compose >> "$LOG_FILE" 2>&1; then
        exit 1
    fi
    if ! create_systemd_service >> "$LOG_FILE" 2>&1; then
        exit 1
    fi
    
    ((CURRENT_STEP++))
    show_progress $CURRENT_STEP $TOTAL_STEPS "Setting up Caddy with automatic HTTPS..."
    if ! setup_caddy >> "$LOG_FILE" 2>&1; then
        exit 1
    fi
    
    ((CURRENT_STEP++))
    show_progress $CURRENT_STEP $TOTAL_STEPS "Starting services..."
    if ! start_dployr >> "$LOG_FILE" 2>&1; then
        exit 1
    fi
    
    # Final update
    show_progress $TOTAL_STEPS $TOTAL_STEPS "Complete!"
    echo ""
    echo ""
    
    show_completion
}

main "$@"
