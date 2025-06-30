#!/bin/bash

set -e
CDN="https://cdn.dployr.io"
DATE=$(date +"%Y%m%d-%H%M%S")

# 0. Check if running as root or with sudo
check_sudo() {
    if [ "$EUID" -ne 0 ]; then
        echo "This script must be run as root or with sudo privileges"
        echo "Please run: sudo $0"
        exit 1
    fi
}

# 1. Install requirements with modern Docker Compose
install_requirements() {
    echo "🔧 Installing requirements..."
    
    OS_TYPE=$(grep -w "ID" /etc/os-release | cut -d "=" -f 2 | tr -d '"')
    
    case "$OS_TYPE" in
        ubuntu|debian)
            apt-get update -qq
            apt-get install -y curl wget git jq nginx ca-certificates gnupg ufw
            
            # Add Docker's official GPG key
            install -m 0755 -d /etc/apt/keyrings
            curl -fsSL https://download.docker.com/linux/ubuntu/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
            chmod a+r /etc/apt/keyrings/docker.gpg
            
            # Add Docker repository
            echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
            
            # Install Docker and modern Compose v2
            apt-get update -qq
            apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
            ;;
        centos|rhel|rocky|alma)
            yum install -y curl wget git jq nginx ufw
            yum install -y yum-utils
            yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
            yum install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
            ;;
    esac
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
    
    systemctl enable --now docker
    usermod -aG docker $SUDO_USER

    ufw allow 7879
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
}

# 4. Modern Docker Compose configuration (no version field)
setup_config() {
    echo "Setting up Docker compose..."

    cat > /data/dployr/docker-compose.yml << EOF
services:
  dployr-web:
    image: dployr.io/dployr:latest
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
}

# Main installation flow
main() {
    echo "Installing dployr.io with modern Docker Compose"
    
    check_sudo
    install_requirements
    setup_docker
    setup_directories  
    setup_config
    
    echo "Installation completed successfully!"
    echo "You may need to log out and back in for Docker group permissions to take effect."
}

main "$@"