#!/usr/bin/env bash

HOME_DIR=$(getent passwd "$USER" | cut -d: -f6)
STATE_DIR="$HOME_DIR/.dployr/state"
mkdir -p "$STATE_DIR"
CDN="https://github.com/tobimadehin/dployr"
INSTALL_START_TIME=$(date +%s)
RANDOM_SUBDOMAIN=$(openssl rand -hex 6)
SERVER_IP=$(curl -s https://api.ipify.org)

# console color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color


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


check_sudo() {
    echo "Checking sudo privileges..."
    if [ "$EUID" -ne 0 ]; then
        log_error "This script must be run as root or with sudo privileges"
        echo "Please run: sudo $0"
        exit 1
    fi
    log_success "Running with sudo privileges"
}

create_dployr_user() {
    local flag_file="$STATE_DIR/create_dployr_user.flag"
    if [ -f "$flag_file" ]; then
        log_info "dployr user already created. Skipping."
        return 0
    fi
    log_info "Creating dployr user..."

    if ! id "dployr" &>/dev/null; then
        useradd -r -m -s /bin/bash -d /home/dployr dployr
        log_success "Created dployr user"
    else
        log_warning "User dployr already exists"
    fi
    
    if [ "$INSTALL_TYPE" = "docker" ]; then
        usermod -aG docker dployr
        log_info "Added dployr to docker group"
    fi
    
    # Create sudo rule for dployr user with safe permissions
    cat > /etc/sudoers.d/dployr << EOF
# Allow dployr to manage its own service and nginx
dployr ALL=(ALL) NOPASSWD: /bin/systemctl start dployr, /bin/systemctl stop dployr, /bin/systemctl restart dployr, /bin/systemctl reload nginx, /bin/systemctl restart nginx
EOF
    
    chmod 440 /etc/sudoers.d/dployr
    log_success "Configured safe sudo permissions for dployr"
}

get_latest_tag() {
    local headers
    local tag
    headers=$(curl -sLI "https://github.com/tobimadehin/dployr/releases/latest")
    tag=$(echo "$headers" | grep -i "location:" | grep -o 'tag/v\?[0-9]\+\.[0-9]\+\(\.[0-9]\+\)\?' | head -1 | cut -d'/' -f2)
    
    if [ -z "$tag" ]; then
        echo "Error: No version found in redirect" >&2
        return 1
    fi
    if [[ ! "$tag" =~ ^v ]]; then
        tag="v$tag"
    fi
    echo "$tag"
}

download_dployr() {
    local flag_file="$STATE_DIR/download_dployr.flag"
    if [ -f "$flag_file" ]; then
        log_info "dployr application already downloaded. Skipping."
        return 0
    fi

    log_info "Downloading dployr..."
    local server_dir="/home/dployr"

    mkdir -p $server_dir
    log_info "Getting latest release information..."

    LATEST_TAG=$(get_latest_tag)
    if [ $? -ne 0 ] || [ -z "$LATEST_TAG" ]; then
        log_error "Error occurred while obtaining release tag"
        exit 1
    fi

    DOWNLOAD_URL="$CDN/releases/download/$LATEST_TAG/dployr-$LATEST_TAG.zip"
    log_info "Downloading from: $DOWNLOAD_URL"

    if curl -fsSL "$DOWNLOAD_URL" -o /tmp/dployr.zip; then
        log_info "Extracting archive..."
        cd /tmp || exit 1

        unzip -oq dployr.zip -d $server_dir
        rm -rf dployr.zip

        chown -R dployr:dployr /home/dployr
        log_success "dployr has been downloaded successfully"
    else
        handle_error "Download error" "Error occurred while downloading dployr"
        exit 1
    fi

    touch "$flag_file"
}

configure_dployr() {
    local flag_file="$STATE_DIR/configure_dployr.flag"
    if [ -f "$flag_file" ]; then
        log_info "Environment already configured. Skipping."
        return 0
    fi

    log_info "Configuring environment..."
    
    cd /home/dployr
    
    log_info "Installing Composer dependencies..."
    sudo -u dployr composer install --no-dev --optimize-autoloader
    
    if [ ! -f .env ]; then
        sudo -u dployr cp .env.example .env
        log_info "Created .env file from example"
    fi
    
    log_info "Generating application key..."
    sudo -u dployr php artisan key:generate --force

    log_info "Setting up database..."
    sudo -u dployr touch database/database.sqlite

    log_info "Running database migrations..."
    sudo -u dployr php artisan migrate --force
    
    log_info "Setting proper permissions..."
    chown -R dployr:www-data /home/dployr
    chmod -R 755 /home/dployr
    chmod -R 775 /home/dployr/storage
    chmod -R 775 /home/dployr/bootstrap/cache
    chmod 664 /home/dployr/database/database.sqlite
    
    log_success "Environment configured successfully"
    touch "$flag_file"
}

select_install_type() {
    case "${1:-}" in
        --docker)
            INSTALL_TYPE="docker"
            log_info "Using Docker installation"
            return
            ;;
        --standalone)
            INSTALL_TYPE="standalone" 
            log_info "Using Standalone installation"
            return
            ;;
    esac
    
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
        read -r -p "Enter your choice [1-2]: " choice
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

install_requirements() {
    local flag_file="$STATE_DIR/install_requirements.flag"
    if [ -f "$flag_file" ]; then
        log_info "System requirements already installed. Skipping."
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
            apt install -y debian-keyring debian-archive-keyring apt-transport-https software-properties-common
            
            log_info "Download PHP 8.3 repository..."
            if [ "$OS_TYPE" = "ubuntu" ]; then
                add-apt-repository -y ppa:ondrej/php
            else
                curl -fsSL https://packages.sury.org/php/apt.gpg | gpg --dearmor -o /usr/share/keyrings/sury-php-keyring.gpg
                echo "deb [signed-by=/usr/share/keyrings/sury-php-keyring.gpg] https://packages.sury.org/php/ $(lsb_release -cs) main" | tee /etc/apt/sources.list.d/sury-php.list
            fi
            
            log_info "Setting up PHP 8.3 repository locally..."
            curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
            curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
            chmod o+r /usr/share/keyrings/caddy-stable-archive-keyring.gpg
            chmod o+r /etc/apt/sources.list.d/caddy-stable.list
            
            apt-get update -qq
            
            PHP_PACKAGES="php8.3-fpm php8.3-cli php8.3-common php8.3-dom php8.3-curl php8.3-mbstring php8.3-xml php8.3-zip php8.3-bcmath php8.3-intl php8.3-gd php8.3-sqlite3 php8.3-tokenizer composer"
            PACKAGES="curl wget git jq caddy ca-certificates gnupg ufw openssl net-tools unzip ansible $PHP_PACKAGES"
            
            if [ "$INSTALL_TYPE" = "docker" ]; then
                PACKAGES="$PACKAGES docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin"
                
                install -m 0755 -d /etc/apt/keyrings
                curl -fsSL https://download.docker.com/linux/$OS_TYPE/gpg | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
                chmod a+r /etc/apt/keyrings/docker.gpg
                
                echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$OS_TYPE $(lsb_release -cs) stable" | tee /etc/apt/sources.list.d/docker.list > /dev/null
                apt-get update -qq
            fi
            
            apt-get install -y $PACKAGES
            
            log_info "Starting fpm..."
            systemctl enable php8.3-fpm
            systemctl start php8.3-fpm
            ;;
        centos|rhel|rocky|alma)
            yum install -y epel-release
            
            log_info "Instally remy repository for PHP 8.3..."
            yum install -y "https://rpms.remirepo.net/enterprise/remi-release-$(rpm -E %rhel).rpm"
            yum-config-manager --enable remi-php83
            
            log_info "Installing Caddy..."
            yum install -y yum-plugin-copr
            yum copr enable -y @caddy/caddy

            PHP_PACKAGES="php php-fpm php-cli php-common php-dom php-curl php-mbstring php-xml php-zip php-bcmath php-intl php-gd php-sqlite3 composer"
            
            yum install -y curl wget git jq caddy ufw openssl ansible $PHP_PACKAGES
            
            if [ "$INSTALL_TYPE" = "docker" ]; then
                yum install -y yum-utils
                yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
                yum install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
            fi
            
            log_info "Starting fpm..."
            systemctl enable php-fpm
            systemctl start php-fpm
            ;;
        *)
            handle_error "Package error" "Error occurred while installing required packages" 
            exit 1
            ;;
    esac
    
    log_success "System requirements installed"
    touch "$flag_file"
}

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
    
    # Copy to Docker data directory
    cp -r /home/dployr/* /data/dployr/
    chown -R dployr:dployr /data/dployr
    
    cat > /data/dployr/docker-compose.yml << EOF
services:
  dployr-web:
    image: php:8.3-fpm-alpine
    user: "dployr:dployr"
    ports: 
      - "7879:8000"
    volumes:
      - /data/dployr:/var/www/html
      - /var/run/docker.sock:/var/run/docker.sock
    working_dir: /var/www/html
    command: php artisan serve --host=0.0.0.0 --port=8000
    environment:
      - APP_ENV=production
      - APP_DEBUG=false
    restart: unless-stopped
    depends_on:
      - dployr-queue

  dployr-queue:
    image: php:8.3-fpm-alpine
    user: "dployr:dployr"
    volumes:
      - /data/dployr:/var/www/html
      - /var/run/docker.sock:/var/run/docker.sock
    working_dir: /var/www/html
    command: php artisan queue:work --sleep=3 --tries=3 --max-time=3600
    environment:
      - APP_ENV=production
      - APP_DEBUG=false
    restart: unless-stopped
EOF
    
    log_success "Docker Compose configuration created"
    touch "$flag_file"
}

determine_ip_addr() {
    local public_ip=""
    local private_ip=""
    
    if ! public_ip=$(curl -s "https://api.ipify.org"); then
        echo "Error: failed to get public IP address" >&2
        return 1
    fi
    
    # Trim whitespace
    public_ip=$(echo "$public_ip" | tr -d '[:space:]')
    
    # Get private IP - find first non-loopback IPv4 address on active interface
    private_ip=$(ip route get 8.8.8.8 2>/dev/null | grep -oP 'src \K\S+' | head -1)
    
    # Fallback
    if [[ -z "$private_ip" ]]; then
        private_ip=$(hostname -I 2>/dev/null | awk '{print $1}')
    fi
    
    if [[ -z "$private_ip" ]]; then
        echo "Error: no private IP found" >&2
        return 1
    fi

    export PUBLIC_IP="$public_ip"
    export PRIVATE_IP="$private_ip"
    
    echo "Public IP: $public_ip"
    echo "Private IP: $private_ip"
    return 0
}

create_systemd_service() {
    if [ "$INSTALL_TYPE" != "standalone" ]; then
        return
    fi
    
    log_info "Creating systemd service..."

# main service
    cat > /etc/systemd/system/dployr.service << EOF
[Unit]
Description=dployr 
After=network.target php8.3-fpm.service

[Service]
Type=simple
ExecStart=/usr/bin/php /home/dployr/artisan serve --host=0.0.0.0 
Restart=on-failure
User=dployr
Group=dployr
WorkingDirectory=/home/dployr
StandardOutput=append:/var/log/dployr.log
StandardError=inherit
Environment=APP_ENV=production

[Install]
WantedBy=multi-user.target
EOF

# worker service
    cat > /etc/systemd/system/dployr-worker.service << EOF
[Unit]
Description=dployr - queue worker
After=network.target php8.3-fpm.service

[Service]
Type=simple
ExecStart=/usr/bin/php /home/dployr/artisan queue:work --sleep=3 --tries=3 --max-time=3600
Restart=on-failure
User=dployr
Group=dployr
WorkingDirectory=/home/dployr
StandardOutput=append:/var/log/dployr.log
StandardError=inherit
Environment=APP_ENV=production

[Install]
WantedBy=multi-user.target
EOF
    
    systemctl daemon-reload
    systemctl enable dployr
    systemctl enable dployr-worker
    log_success "Systemd services created and enabled"
}

create_cloudflare_dns() {
    local subdomain="$RANDOM_SUBDOMAIN"
    
    log_info "Creating DNS record for $SERVER_IP"
    
    local response
    response=$(curl -s -X POST "https://dployr.dev/api/dns/create" \
        -H "Content-Type: application/json" \
        --data '{
            "subdomain": "'"$subdomain"'",
            "host": "'"$SERVER_IP"'"
        }')
    
    local success
    success=$(echo $response | jq -r '.success // false')
    
    if [ "$success" = "true" ]; then
        log_success "DNS record created successfully for $subdomain.dployr.dev"
        return 0
    else
        local error_message
        error_message=$(echo $response | jq -r '.errors.message // "Unknown error"')
        handle_error "Setup error" $error_message
        return 1
    fi
}

setup_caddy() {

    log_info "Setting up caddy configuration..."
    
    if ! create_cloudflare_dns; then
        handle_error "DNS setup error" "Failed to create DNS record"
        exit 1
    fi
    
    DOMAIN="$RANDOM_SUBDOMAIN.dployr.dev"
    APP_FOLDER="/home/dployr"
    log_info "Using domain: $DOMAIN"
    
    # Behind NAT - serve on both public and private IP
    if [ "$PUBLIC_IP" != "$PRIVATE_IP" ]; then
        cat > /etc/caddy/Caddyfile << EOF
{
    auto_https disable_redirects
}

http://$PUBLIC_IP:8000  {
    root * $APP_FOLDER/public
    php_fastcgi unix//run/php/php8.3-fpm.sock
    try_files {path} {path}/ /index.php?{query}
    file_server
    
    # Enable gzip compression
    encode gzip
    
    # Security headers
    header {
        X-Content-Type-Options nosniff
        X-Frame-Options DENY
        X-XSS-Protection "1; mode=block"
        Referrer-Policy strict-origin-when-cross-origin
    }
}

http://$PRIVATE_IP:8000 {
    root * $APP_FOLDER/public
    php_fastcgi unix//run/php/php8.3-fpm.sock
    try_files {path} {path}/ /index.php?{query}
    file_server
    
    # Enable gzip compression
    encode gzip
    
    # Security headers
    header {
        X-Content-Type-Options nosniff
        X-Frame-Options DENY
        X-XSS-Protection "1; mode=block"
        Referrer-Policy strict-origin-when-cross-origin
    }
}
EOF
        log_info "Configured caddy successfully for $PUBLIC_IP and $PRIVATE_IP"
    else
        cat > /etc/caddy/Caddyfile << EOF
$PRIVATE_IP {
    root * $APP_FOLDER/public
    php_fastcgi unix//run/php/php8.3-fpm.sock
    try_files {path} {path}/ /index.php?{query}
    file_server
    
    # Enable gzip compression
    encode gzip
    
    # Security headers
    header {
        X-Content-Type-Options nosniff
        X-Frame-Options DENY
        X-XSS-Protection "1; mode=block"
        Referrer-Policy strict-origin-when-cross-origin
    }
}
EOF
        log_info "Configured Caddy for direct internet access on $DOMAIN"
    fi
    
    log_info "Setting up permissions..."
    if ! chown caddy:caddy /etc/caddy/Caddyfile; then
        handle_error "Permission error" "Failed to set Caddyfile ownership"
        return 1
    fi
    
    if ! chmod 644 /etc/caddy/Caddyfile; then
        handle_error "Permission error" "Failed to set Caddyfile permissions"
        return 1
    fi
    
    log_info "Opening firewall ports..."
    ufw allow 80 2>/dev/null || true
    ufw allow 443 2>/dev/null || true
    ufw allow 8000 2>/dev/null || true
    ufw allow 22 2>/dev/null || true
    
    log_info "Restarting Caddy service..."
    if ! systemctl restart caddy; then
        handle_error "Caddy restart error" "Failed to restart Caddy service"
        return 1
    fi
    
    if systemctl is-active --quiet caddy; then
        log_success "Caddy configured successfully for domain: $DOMAIN"
        if [ "$PUBLIC_IP" != "$PRIVATE_IP" ]; then
            log_info "Available at: https://$PUBLIC_IP and http://$PRIVATE_IP"
        else
            log_info "Available at: https://$PRIVATE_IP"
        fi
    else
        handle_error "Caddy setup error" "Caddy service failed to start properly"
        return 1
    fi
    
    touch "$flag_file"
}

start_dployr() {
    log_info "Starting dployr..."
    
    if [ "$INSTALL_TYPE" = "docker" ]; then
        cd /data/dployr
        docker compose up -d
        sleep 5
        if docker compose ps | grep -q "Up"; then
            log_success "Dployr started successfully"
        else
            handle_error "Program error" "Failed to start dployr"
            exit 1
        fi
    else
        systemctl start dployr
        sleep 2
        
        systemctl start dployr-worker
        sleep 2
        
        if systemctl is-active --quiet dployr && systemctl is-active --quiet dployr-worker; then
            log_success "dployr started successfully"
        else
            handle_error "Program error" "Failed to start dployr services"
            exit 1
        fi
    fi
}

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
    
    if [ "$PUBLIC_IP" = "$PRIVATE_IP" ]; then
        echo "  http://$PRIVATE_IP:8000"
    else
        echo "  https://$PUBLIC_IP:8000"
        echo "  http://$PRIVATE_IP:8000"
    fi
    
    echo ""
    echo "Service management:"
    if [ "$INSTALL_TYPE" = "docker" ]; then
        echo "  Start:   cd /data/dployr && docker compose up -d"
        echo "  Stop:    cd /data/dployr && docker compose down"
        echo "  Logs:    cd /data/dployr && docker compose logs -f"
    else
        echo "  Start:   sudo systemctl start dployr dployr-worker"
        echo "  Stop:    sudo systemctl stop dployr dployr-worker"
        echo "  Status:  sudo systemctl status dployr dployr-worker"
        echo "  Logs:    tail -f /var/log/dployr.log"
    fi
    echo ""
}

exec 3>&1 4>&2

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

main() {
    echo "Starting dployr installer..."
    echo ""

    LOG_FILE="/tmp/dployr-install-$(date +%Y%m%d-%H%M%S).log"
    echo "Installation started at $(date)" > "$LOG_FILE"
    
    check_sudo
    select_install_type "$1"
    
    TOTAL_STEPS=10
    CURRENT_STEP=0
    
    show_progress $CURRENT_STEP $TOTAL_STEPS "Creating user..."
    if ! create_dployr_user >> "$LOG_FILE" 2>&1; then
        exit 1
    fi

    show_progress $CURRENT_STEP $TOTAL_STEPS "Determining IP address..."
    if ! determine_ip_addr >> "$LOG_FILE" 2>&1; then
        exit 1
    fi
    
    ((CURRENT_STEP++))
    show_progress $CURRENT_STEP $TOTAL_STEPS "Installing requirements..."
    if ! install_requirements >> "$LOG_FILE" 2>&1; then
        exit 1
    fi

    ((CURRENT_STEP++))
    show_progress $CURRENT_STEP $TOTAL_STEPS "Downloading archive..."
    if ! download_dployr >> "$LOG_FILE" 2>&1; then
        exit 1
    fi
    
    ((CURRENT_STEP++))
    show_progress $CURRENT_STEP $TOTAL_STEPS "Configuring Environment..."
    if ! configure_dployr >> "$LOG_FILE" 2>&1; then
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
    show_progress $CURRENT_STEP $TOTAL_STEPS "Setting up caddy..."
    if ! setup_caddy >> "$LOG_FILE" 2>&1; then
        exit 1
    fi
    
    ((CURRENT_STEP++))
    show_progress $CURRENT_STEP $TOTAL_STEPS "Starting services..."
    if ! start_dployr >> "$LOG_FILE" 2>&1; then
        exit 1
    fi
    
    show_progress $TOTAL_STEPS $TOTAL_STEPS "Complete!"
    echo ""
    echo ""
    
    show_completion
}

main "$@"
