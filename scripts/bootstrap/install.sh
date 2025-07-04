#!/usr/bin/env bash

# Load environment variables from .env file
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="$SCRIPT_DIR/../../.env.dev"
if [ -f "$ENV_FILE" ]; then
  export $(grep -v '^#' "$ENV_FILE" | xargs)
fi

set -e
CDN="https://cdn.dployr.io"
DATE=$(date +"%Y%m%d-%H%M%S")
REGISTRY_URL=$HOST:$PORT

# 0. Check if running as root or with sudo
check_sudo() {
    if [ "$EUID" -ne 0 ]; then
        echo "This script must be run as root or with sudo privileges"
        echo "Please run: sudo $0"
        exit 1
    fi
}

# 1. Install requirements with Docker Compose
install_requirements() {
    echo "Installing requirements..."
    
    OS_TYPE=$(grep -w "ID" /etc/os-release | cut -d "=" -f 2 | tr -d '"')
    OS_VERSION=$(grep -w "VERSION_CODENAME" /etc/os-release | cut -d "=" -f 2 | tr -d '"')
    
    echo "Detected OS: $OS_TYPE $OS_VERSION"
    
    # Set non-interactive mode to prevent hanging
    export DEBIAN_FRONTEND=noninteractive
    export DEBCONF_NONINTERACTIVE_SEEN=true
    
    case "$OS_TYPE" in
        ubuntu|debian)
            echo "docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin" | tr ' ' '\n' | while read package; do
                echo "$package docker-ce/restart_policy string none" | debconf-set-selections
            done
            
            apt-get update -qq
            apt-get install -y curl wget git jq nginx ca-certificates gnupg ufw
            
            # Add Docker's official GPG key
            install -m 0755 -d /etc/apt/keyrings
            curl -fsSL https://download.docker.com/linux/$OS_TYPE/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
            chmod a+r /etc/apt/keyrings/docker.gpg
            
            # Add Docker repository
            echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$OS_TYPE $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
            
            apt-get update -qq

            echo "Installing Docker packages..."
            apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
            
            rm -f /usr/sbin/policy-rc.d
            ;;
        centos|rhel|rocky|alma)
            yum install -y curl wget git jq nginx ufw
            yum install -y yum-utils
            yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
            yum install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
            ;;
        *)
            echo "Unsupported operating system: $OS_TYPE"
            exit 1
            ;;
    esac
    
    # Verify Docker installation
    if ! command -v docker &> /dev/null; then
        echo "ERROR: Docker installation failed"
        exit 1
    fi
    
    if ! command -v docker compose &> /dev/null; then
        echo "ERROR: Docker Compose installation failed"
        exit 1
    fi
    
    echo "Docker and Docker Compose installed successfully"
}

# 2. Docker setup 
setup_docker() {
    echo "Configuring Docker for Next.js..."
    
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
    
    echo "Starting Docker service..."
    systemctl enable docker
    
    # Use timeout to prevent hanging on service start
    timeout 60 systemctl start docker || {
        echo "ERROR: Docker service failed to start within 60 seconds"
        systemctl status docker
        exit 1
    }
    
    # Verify Docker service is running
    if ! systemctl is-active --quiet docker; then
        echo "ERROR: Docker service failed to start"
        systemctl status docker
        exit 1
    fi
    
    usermod -aG docker $SUDO_USER

    ufw allow 7879
    
    echo "Docker configured successfully"
}

# 3. dployr.io optimized structure
setup_directories() {
    echo "Setting up dployr.io directories..."
    
    mkdir -p /data/dployr/nextjs-apps
    mkdir -p /data/dployr/builds
    mkdir -p /data/dployr/images/cache
    mkdir -p /data/dployr/logs/{hot,warm,cold}
    mkdir -p /data/dployr/monitoring/{prometheus,grafana}
    mkdir -p /data/dployr/ssl
    mkdir -p /data/dployr/nginx/sites
    mkdir -p /data/dployr/redis
    
    # Set permissions for dployr.io user
    useradd -r -s /bin/false dployr || true
    chown -R dployr:dployr /data/dployr
    
    echo "Directories created successfully"
}

# 4. Modern Docker Compose configuration (no version field)
setup_config() {
    echo "Setting up Docker compose..."

    cat > /data/dployr/docker-compose.yml << EOF
services:
  dployr-web:
    image: $REGISTRY_URL/dployr:latest
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

    echo "Docker Compose configuration created"
}

# Main installation flow
main() {
    echo "Installing dployr.io with Docker Compose"
    
    check_sudo
    install_requirements
    setup_docker
    setup_directories  
    setup_config
    
    echo "Installation completed successfully!"
    echo "You may need to log out and back in for Docker group permissions to take effect."
}

main "$@"