#!/bin/bash

# Copyright 2025 Emmanuel Madehin
# SPDX-License-Identifier: Apache-2.0

# dployr Unix Installer (Linux/macOS)

set -e

START_TIME=$(date +%s)

LOG_DIR="/var/log/dployrd"

if [[ $EUID -eq 0 ]]; then
    mkdir -p "$LOG_DIR"
    LOG_FILE="$LOG_DIR/install.log"
else
    LOG_DIR="$HOME/.dployr"
    mkdir -p "$LOG_DIR"
    LOG_FILE="$LOG_DIR/install.log"
fi

# 3 = stderr for user-facing messages
# 4 = log file for structured JSON logs
exec 3>&2
exec 4>>"$LOG_FILE"

log_json() {
    local level="$1"
    local message="$2"
    local timestamp
    timestamp=$(date -Iseconds 2>/dev/null || date "+%Y-%m-%dT%H:%M:%S%z")
    printf '{"timestamp":"%s","level":"%s","message":"%s","pid":%d,"user":"%s"}\n' \
        "$timestamp" "$level" "$message" "$$" "${USER:-unknown}" >&4
}

log_json "info" "Installation started"
log_json "info" "Logging to $LOG_FILE"

export DEBIAN_FRONTEND=noninteractive

INSTALL_DIR="/usr/local/bin"
VERSION="latest"
TOKEN=""
REPO="dployr-io/dployr"
DPLOYR_DOMAIN=""
BASE_URL=""
INSTANCE_ID=""

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info() {
    echo -e "${GREEN}[INFO]${NC} $1" >&3
    log_json "info" "$1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1" >&3
    log_json "warn" "$1"
}

error() {
    local msg="$1"
    echo -e "${RED}[ERROR]${NC} $msg" >&3
    log_json "error" "$msg"

    if command -v tail >/dev/null 2>&1 && [[ -f "$LOG_FILE" ]]; then
        echo -e "${YELLOW}[LOG]${NC} Last 20 lines from $LOG_FILE:" >&3
        tail -n 20 "$LOG_FILE" >&3 || true
    fi

    echo -e "${YELLOW}[INFO]${NC} Full log available at: $LOG_FILE" >&3
    exit 1
}

install_git() {
    if command -v git &>/dev/null; then
        return 0
    fi

    info "Installing git..."
    
    case $OS in
        linux)
            if command -v apt &>/dev/null; then
                while sudo fuser /var/lib/apt/lists/lock >/dev/null 2>&1; do
                    sleep 1
                done

                apt update -qq < /dev/null || warn "apt update failed while installing git"

                while sudo fuser /var/lib/apt/lists/lock >/dev/null 2>&1; do
                    sleep 1
                done

                apt install -y -qq git < /dev/null || error "Failed to install git via apt"
            elif command -v yum &>/dev/null; then
                yum install -y -q git || error "Failed to install git via yum"
            else
                error "Unable to install git: no supported package manager found (apt or yum required)"
            fi
            ;;
        darwin)
            if command -v brew &>/dev/null; then
                brew install -q git || error "Failed to install git via Homebrew"
            else
                error "Homebrew not found. Please install Homebrew first: https://brew.sh"
            fi
            ;;
    esac

    if ! command -v git &>/dev/null; then
        error "Failed to install git"
    fi
}

install_jq() {
    if command -v jq &>/dev/null; then
        return 0
    fi

    info "Installing jq..."
    
    case $OS in
        linux)
            if command -v apt &>/dev/null; then
                while sudo fuser /var/lib/apt/lists/lock >/dev/null 2>&1; do
                    sleep 1
                done

                apt update -qq < /dev/null || warn "apt update failed while installing jq; attempting fallback download"

                while sudo fuser /var/lib/apt/lists/lock >/dev/null 2>&1; do
                    sleep 1
                done

                if ! apt install -y -qq jq < /dev/null; then
                    warn "apt install jq failed; falling back to static jq binary download"
                    local jq_url="https://github.com/jqlang/jq/releases/latest/download/jq-linux-amd64"
                    curl -sL "$jq_url" -o "$INSTALL_DIR/jq" || error "Failed to download jq binary from $jq_url"
                    chmod +x "$INSTALL_DIR/jq"
                fi
            elif command -v yum &>/dev/null; then
                yum install -y -q jq
            else
                local jq_url="https://github.com/jqlang/jq/releases/latest/download/jq-linux-amd64"
                curl -sL "$jq_url" -o "$INSTALL_DIR/jq"
                chmod +x "$INSTALL_DIR/jq"
            fi
            ;;
        darwin)
            if command -v brew &>/dev/null; then
                brew install -q jq
            else
                local jq_url="https://github.com/jqlang/jq/releases/latest/download/jq-macos-amd64"
                curl -sL "$jq_url" -o "$INSTALL_DIR/jq"
                chmod +x "$INSTALL_DIR/jq"
            fi
            ;;
    esac

    if ! command -v jq &>/dev/null; then
        error "Failed to install jq"
    fi
}

# get_daemon_port tries to read the HTTP port from the config file,
# falling back to the default 7879.
get_daemon_port() {
    local cfg_file

    case $OS in
        darwin)
            cfg_file="/usr/local/etc/dployr/config.toml"
            ;;
        *)
            cfg_file="/etc/dployr/config.toml"
            ;;
    esac

    if [[ -r "$cfg_file" ]]; then
        local p
        p=$(grep -E '^port[[:space:]]*=' "$cfg_file" | head -1 | sed 's/[^0-9]*\([0-9][0-9]*\).*/\1/' || true)
        if [[ -n "$p" ]]; then
            echo "$p"
            return 0
        fi
    fi

    echo "7879"
}

# wait_for_pending_tasks blocks for a short period while there are
# in-progress deployments reported by the daemon. On first install or
# when the daemon is not running, it returns immediately.
wait_for_pending_tasks() {
    # If daemon is not running yet, nothing to wait for.
    if ! pgrep -x "dployrd" >/dev/null 2>&1; then
        return 0
    fi

    # jq is installed earlier; curl is expected to be available.
    if ! command -v curl >/dev/null 2>&1; then
        return 0
    fi

    local port
    port=$(get_daemon_port)

    info "Checking for pending tasks before upgrade (port $port)..."

    # Best-effort: tell the daemon we are entering an updating window.
    local auth_header=""
    if [[ -n "$TOKEN" ]]; then
        auth_header="Authorization: Bearer $TOKEN"
    fi

    curl -sS -X POST \
        -H "Content-Type: application/json" \
        ${auth_header:+-H "$auth_header"} \
        -d '{"mode":"updating"}' \
        "http://localhost:${port}/system/mode" >/dev/null 2>&1 || true

    local attempts=0
    local max_attempts=24 # ~2 minutes total at 5s intervals

    while (( attempts < max_attempts )); do
        attempts=$((attempts + 1))

        local resp
        resp=$(curl -sS ${auth_header:+-H "$auth_header"} "http://localhost:${port}/system/tasks?status=pending" || true)
        if [[ -z "$resp" ]]; then
            # If we can't talk to the daemon, do not block the install.
            warn "Could not query dployrd for pending deployments; continuing with install."
            return 0
        fi

        local count
        count=$(echo "$resp" | jq '.count' 2>/dev/null || echo "")
        if [[ -z "$count" ]]; then
            warn "Could not parse pending tasks response; continuing with install."
            return 0
        fi

        if [[ "$count" -eq 0 ]]; then
            info "No pending tasks detected. Proceeding with install."
            # Best-effort: mark daemon as ready again.
            curl -sS -X POST \
                -H "Content-Type: application/json" \
                ${auth_header:+-H "$auth_header"} \
                -d '{"mode":"ready"}' \
                "http://localhost:${port}/system/mode" >/dev/null 2>&1 || true
            return 0
        fi

        info "There are $count pending tasks. Waiting for them to finish before install (attempt $attempts/$max_attempts)..."
        sleep 1
    done

    warn "Timed out waiting for pending deployments to finish. Proceeding with install."
    # Even on timeout, attempt to return daemon to ready mode.
    curl -sS -X POST \
        -H "Content-Type: application/json" \
        ${auth_header:+-H "$auth_header"} \
        -d '{"mode":"ready"}' \
        "http://localhost:${port}/system/mode" >/dev/null 2>&1 || true
}

parse_json() {
    local expr="$1"
    local value
    value=$(jq -r "${expr} // empty" 2>/dev/null || echo "")
    if [[ "$value" == "null" ]]; then
        echo ""
    else
        echo "$value"
    fi
}

echo "dployr Unix Installer" >&3
echo "====================" >&3

show_help() {
    cat >&3 << EOF
Usage: $0 [options]

Options:
  -v, --version <tag>       Install a specific version (default: latest)
  -t, --token <token>       Instance registration token
  -b, --base <url>          Base API URL (overrides --env)
  -i, --instance <id>       Instance ID for config
  -e, --env <env>           Environment: prod (default), dev
  -h, --help                Show this help

Environment:
  prod → https://base.dployr.io
  dev  → https://base.dployr.dev

Examples:
  $0
  $0 -e dev
  $0 -v v0.3.1 -t <token>
  $0 -e dev -b https://custom.internal
EOF
}

if [[ "$1" == "--help" || "$1" == "-h" ]]; then
    show_help
    exit 0
fi

ENVIRONMENT="prod"
BASE_URL_EXPLICIT=0

while [[ $# -gt 0 ]]; do
    case "$1" in
        -v|--version)
            [[ -z "$2" ]] && error "Missing value for $1"
            VERSION="$2"
            shift 2
            ;;
        -t|--token)
            [[ -z "$2" ]] && error "Missing value for $1"
            TOKEN="$2"
            shift 2
            ;;
        -b|--base)
            [[ -z "$2" ]] && error "Missing value for $1"
            BASE_URL="$2"
            BASE_URL_EXPLICIT=1
            shift 2
            ;;
        -i|--instance)
            [[ -z "$2" ]] && error "Missing value for $1"
            INSTANCE_ID="$2"
            shift 2
            ;;
        -e|--env)
            [[ -z "$2" ]] && error "Missing value for $1"
            ENVIRONMENT="$2"
            shift 2
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            error "Unknown argument: $1"
            ;;
    esac
done

case "$ENVIRONMENT" in
    prod)
        DEFAULT_BASE_URL="https://base.dployr.io"
        ;;
    dev)
        DEFAULT_BASE_URL="https://base.dployr.dev"
        ;;
    *)
        error "Invalid environment: $ENVIRONMENT (expected: prod or dev)"
        ;;
esac

if [[ $BASE_URL_EXPLICIT -eq 0 ]]; then
    BASE_URL="$DEFAULT_BASE_URL"
fi

info "Environment: $ENVIRONMENT"
info "Base URL: $BASE_URL"

register_instance() {
    local token="$1"

    info "Checking if instance is already registered..."

    local reg
    reg=$(curl -sS "http://localhost:7879/system/registered" 2>/dev/null || true)
    local is_registered
    is_registered=$(echo "$reg" | parse_json '.registered')
    if [[ "$is_registered" == "true" ]]; then
        info "Instance already registered with base; skipping registration."

        # Update local copy of bootstrap token
        if [[ -n "$token" ]]; then
            curl -sS -X POST \
                -H "Content-Type: application/json" \
                -d "{\"token\":\"$token\"}" \
                "http://localhost:7879/system/token/rotate" >/dev/null 2>&1 || true
        fi

        return 0
    fi

    info "Registering instance with base..."

    local response
    response=$(curl -sS -X POST \
        -H "Content-Type: application/json" \
        -d "{\"token\":\"$token\"}" \
        "http://localhost:7879/system/domain" 2>&1)
    local status=$?

    if [[ $status -ne 0 ]]; then
        warn "Failed to register instance (curl exit $status). Visit https://dployr.io/docs/quickstart.html"
        log_json "error" "$response"
        return 1
    fi

    log_json "info" "Registration response received"

    local success instance_id audience
    success=$(echo "$response" | parse_json '.success')
    DPLOYR_DOMAIN=$(echo "$response" | parse_json '.domain')
    instance_id=$(echo "$response" | parse_json '.instanceId')
    audience=$(echo "$response" | parse_json '.audience')

    local error_msg error_code help_link display_msg
    error_msg=$(echo "$response" | parse_json '.message')
    error_code=$(echo "$response" | parse_json '.code')
    help_link=$(echo "$response" | parse_json '.helpLink')
    display_msg="$error_msg"

    if [[ "$success" == "true" && -n "$DPLOYR_DOMAIN" ]]; then
        info "Instance registered successfully. URL: https://$DPLOYR_DOMAIN (instance: $instance_id, audience: $audience)"
        log_json "info" "Instance registered with domain: $DPLOYR_DOMAIN, instanceId: $instance_id, audience: $audience"
        return 0
    fi

    if [[ -n "$error_code" || -n "$error_msg" ]]; then
        local help_suffix=""
        if [[ -n "$help_link" ]]; then
            help_suffix=" See $help_link for more information."
        fi

        if [[ "$error_code" == "auth.bad_token" ]]; then
            error "Invalid or expired token. Error: $display_msg (code: $error_code).$help_suffix"
        else
            log_json "error" "Registration failed: $display_msg (code: $error_code, helpLink: $help_link)"
            error "Instance registration failed. Error: $display_msg (code: $error_code).$help_suffix"
        fi
    fi

    log_json "error" "Registration failed, unexpected response: $response"
    error "Instance registration failed with unexpected response: $response"
}

if [[ $EUID -eq 0 ]]; then
    info "Installing system-wide to $INSTALL_DIR"
else
    INSTALL_DIR="$HOME/.local/bin"
    info "Installing to user directory: $INSTALL_DIR"
    mkdir -p "$INSTALL_DIR"
fi

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case $ARCH in
    x86_64) ARCH="x86_64" ;;
    aarch64|arm64) ARCH="arm64" ;;
    *) error "Unsupported architecture: $ARCH" ;;
esac

case $OS in
    linux) PLATFORM="Linux" ;;
    darwin) PLATFORM="Darwin" ;;
    *) error "Unsupported OS: $OS" ;;
esac

info "Detected platform: $PLATFORM-$ARCH"

install_git
install_jq

if [[ "$VERSION" == "latest" ]]; then
    info "Fetching latest release..."
    VERSION=$(curl -sS "https://api.github.com/repos/$REPO/releases/latest" | parse_json '.tag_name')
    [[ -z "$VERSION" ]] && error "Failed to get latest version"
    info "Latest version: $VERSION"
fi

info "Downloading dployr binaries..."
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/dployr-$PLATFORM-$ARCH.tar.gz"
TEMP_DIR=$(mktemp -d)

curl -sL "$DOWNLOAD_URL" -o "$TEMP_DIR/dployr.tar.gz" || error "Failed to download dployr"
tar -xzf "$TEMP_DIR/dployr.tar.gz" -C "$TEMP_DIR" || error "Failed to extract dployr"

EXTRACT_DIR="$TEMP_DIR/dployr-$PLATFORM-$ARCH"

if pgrep -x "dployrd" > /dev/null; then
    # Give any in-flight deployments a chance to complete before
    # stopping the daemon for upgrade.
    wait_for_pending_tasks

    info "Stopping dployrd..."
    case $OS in
        linux)
            if systemctl is-active --quiet dployrd 2>/dev/null; then
                sudo systemctl stop dployrd
            else
                pkill -x dployrd || sudo pkill -x dployrd || true
            fi
            ;;
        darwin)
            if launchctl list | grep -q io.dployr.dployrd; then
                launchctl stop io.dployr.dployrd
            else
                pkill -x dployrd
            fi
            ;;
    esac
    sleep 2
fi

info "Installing dployr binaries..."
cp "$EXTRACT_DIR/dployr" "$INSTALL_DIR/" || error "Failed to copy dployr"
chmod +x "$INSTALL_DIR/dployr"

cp "$EXTRACT_DIR/dployrd" "$INSTALL_DIR/" || error "Failed to copy dployrd"
chmod +x "$INSTALL_DIR/dployrd"

if [[ $EUID -eq 0 ]]; then
    # Create dployrd system user and required groups 
    for group in dployr-owner dployr-admin dployr-dev dployr-viewer; do
        if ! getent group "$group" &>/dev/null; then
            groupadd "$group"
            log_json "info" "Created group: $group"
        fi
    done

    if ! id "dployrd" &>/dev/null; then
        _groups="dployr-admin"
        getent group docker &>/dev/null && _groups="dployr-admin,docker"
        useradd --system --create-home --shell /bin/bash -G "$_groups" dployrd
        log_json "info" "Created dployrd system user"
    fi
    mkdir -p /var/log/dployrd /var/lib/dployrd
    chown dployrd:dployrd /var/log/dployrd /var/lib/dployrd
    mkdir -p /var/lib/dployrd/.dployr/caddy
    touch /var/lib/dployrd/.dployr/caddy/Caddyfile
    chown -R dployrd:dployrd /var/lib/dployrd/.dployr
fi

info "Installing Caddy..."
if command -v caddy &> /dev/null; then
    info "Caddy already installed"
else
    case $OS in
        linux)
            if [[ $EUID -eq 0 ]] && command -v apt &> /dev/null; then
                info "Installing Caddy via apt..."

                while sudo fuser /var/lib/apt/lists/lock >/dev/null 2>&1; do
                    sleep 2
                done

                apt update -qq < /dev/null
                apt install -y -qq debian-keyring debian-archive-keyring apt-transport-https < /dev/null

                curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' \
                    | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg \
                    || error "Failed to import Caddy GPG key"

                curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' \
                    | tee /etc/apt/sources.list.d/caddy-stable.list >/dev/null \
                    || error "Failed to add Caddy apt repository"

                while sudo fuser /var/lib/apt/lists/lock >/dev/null 2>&1; do
                    sleep 2
                done

                apt update -qq < /dev/null || error "apt update failed after adding Caddy repository"
                apt install -y -qq caddy < /dev/null || error "Failed to install Caddy via apt"

                info "Configuring Caddy systemd service..."
                # Stop and disable only if we need to replace the service file
                systemctl stop caddy 2>/dev/null || true
                systemctl disable caddy 2>/dev/null || true
                
                cat > /var/lib/dployrd/.dployr/caddy/Caddyfile <<'EOF'
:80 {
    respond "dployr bootstrapping"
}
EOF
                chown dployrd:dployrd /var/lib/dployrd/.dployr/caddy/Caddyfile

                info "Granting Caddy capability to bind to privileged ports..."
                setcap cap_net_bind_service=+ep /usr/bin/caddy || warn "setcap failed; Caddy may not bind to ports 80/443 without root"

                mkdir -p /etc/systemd/system/caddy.service.d
                cat > /etc/systemd/system/caddy.service.d/override.conf << 'EOF'
[Unit]
Description=Caddy web server
After=network.target

[Service]
User=dployrd
Group=dployrd
ExecStart=
ExecStart=/usr/bin/caddy run --config /var/lib/dployrd/.dployr/caddy/Caddyfile
WorkingDirectory=/var/lib/dployrd
ReadWritePaths=/var/lib/dployrd/.dployr
Restart=always
RestartSec=5
EOF

                systemctl daemon-reload || error "systemctl daemon-reload failed"
                systemctl enable caddy || warn "Failed to enable Caddy service"
                systemctl start caddy || error "Failed to start Caddy service"
            else
                info "Installing Caddy via binary download..."
                CADDY_URL="https://github.com/caddyserver/caddy/releases/latest/download/caddy_${OS}_${ARCH}.tar.gz"
                curl -sL "$CADDY_URL" -o "$TEMP_DIR/caddy.tar.gz"
                tar -xzf "$TEMP_DIR/caddy.tar.gz" -C "$TEMP_DIR"
                cp "$TEMP_DIR/caddy" "$INSTALL_DIR/"
                chmod +x "$INSTALL_DIR/caddy"

                # Create systemd service for binary installation (if systemd exists)
                if command -v systemctl &>/dev/null && [[ $EUID -eq 0 ]]; then
                    info "Creating systemd service for Caddy..."
                    cat > /etc/systemd/system/caddy.service << 'EOF'
[Unit]
Description=Caddy web server
After=network.target

[Service]
User=dployrd
Group=dployrd
ExecStart=/usr/local/bin/caddy run --config /var/lib/dployrd/.dployr/caddy/Caddyfile
WorkingDirectory=/var/lib/dployrd
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
                    systemctl daemon-reload
                    systemctl enable caddy || warn "Failed to enable Caddy service"
                    systemctl start caddy || error "Failed to start Caddy service"
                else
                    warn "systemd not available or not root – Caddy binary installed but not started automatically"
                fi
            fi
            ;;
        darwin)
            if command -v brew &> /dev/null; then
                info "Installing Caddy via Homebrew..."
                brew install -q caddy
                info "Starting Caddy as a background service..."
                brew services start caddy || warn "Failed to start Caddy via brew services"
            else
                info "Installing Caddy via binary download (macOS)..."
                CADDY_URL="https://github.com/caddyserver/caddy/releases/latest/download/caddy_${OS}_${ARCH}.tar.gz"
                curl -sL "$CADDY_URL" -o "$TEMP_DIR/caddy.tar.gz"
                tar -xzf "$TEMP_DIR/caddy.tar.gz" -C "$TEMP_DIR"
                cp "$TEMP_DIR/caddy" "$INSTALL_DIR/"
                chmod +x "$INSTALL_DIR/caddy"

                # Create launchd plist for binary installation
                if [[ $EUID -eq 0 ]]; then
                    info "Creating launchd plist for Caddy..."
                    cat > /Library/LaunchDaemons/com.caddyserver.caddy.plist << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.caddyserver.caddy</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/caddy</string>
        <string>run</string>
        <string>--config</string>
        <string>/var/lib/dployrd/.dployr/caddy/Caddyfile</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>WorkingDirectory</key>
    <string>/var/lib/dployrd</string>
    <key>StandardOutPath</key>
    <string>/var/log/dployrd/caddy.log</string>
    <key>StandardErrorPath</key>
    <string>/var/log/dployrd/caddy.log</string>
    <key>UserName</key>
    <string>_dployrd</string>
</dict>
</plist>
EOF
                    launchctl load /Library/LaunchDaemons/com.caddyserver.caddy.plist
                    launchctl start com.caddyserver.caddy || warn "Failed to start Caddy via launchctl"
                else
                    warn "Root privileges required to install launchd plist – Caddy binary installed but not started"
                fi
            fi
            ;;
    esac
fi

# info "Installing vfox..."
# if command -v vfox &> /dev/null; then
#     info "vfox already installed"
# else
#     curl -sSL https://raw.githubusercontent.com/version-fox/vfox/main/install.sh | bash || error "Failed to install vfox"
# fi

info "Installing docker..."
if ! command -v docker &> /dev/null; then
    info "Docker not found. Installing..."
    if [ "$(id -u)" -ne 0 ]; then
        error "Docker installation requires root privileges"
    fi
    
    curl -fsSL https://get.docker.com -o /tmp/get-docker.sh
    sh /tmp/get-docker.sh
    rm -f /tmp/get-docker.sh
    
    if ! command -v docker &> /dev/null; then
        error "Docker installation failed"
    fi
    
    info "Docker installed successfully"
fi

case $OS in
    darwin)
        CONFIG_DIR="/usr/local/etc/dployr"
        ;;
    *)
        CONFIG_DIR="/etc/dployr"
        ;;
esac
CONFIG_FILE="$CONFIG_DIR/config.toml"

info "Creating system configuration..."
[[ $EUID -ne 0 ]] && error "System-wide installation requires root privileges"

mkdir -p "$CONFIG_DIR"

if [[ ! -f "$CONFIG_FILE" ]]; then
    # Use custom instance_id if provided, otherwise use default
    instance_value="${INSTANCE_ID:-my-instance-id}"
    cat > "$CONFIG_FILE" << EOF
address = "localhost"
port = 7879
max-workers = 5

base_url = "$BASE_URL"
instance_id = "$instance_value"
EOF
    chmod 644 "$CONFIG_FILE"
    chmod 755 "$CONFIG_DIR"
    info "Created system config at $CONFIG_FILE"
    [[ -n "$INSTANCE_ID" ]] && info "Using custom instance_id: $INSTANCE_ID"
else
    info "Config file already exists at $CONFIG_FILE"
fi

info "Setting up dployrd service..."
case $OS in
    linux)
        if ! id "dployrd" &>/dev/null; then
            _groups="dployr-admin"
            getent group docker &>/dev/null && _groups="dployr-admin,docker"
            useradd --system --create-home --shell /bin/bash -G "$_groups" dployrd
            log_json "info" "Created dployrd system user"
        else
            usermod -aG docker dployrd 2>/dev/null || true
        fi
        
        mkdir -p /var/log/dployrd /var/lib/dployrd
        chown dployrd:dployrd /var/log/dployrd /var/lib/dployrd
        
        mkdir -p /var/lib/dployrd/.dployr/logs/caddy
        chown -R dployrd:dployrd /var/lib/dployrd/.dployr
        systemctl restart caddy || warn "Failed to restart Caddy"
        chown -R dployrd:dployrd /var/lib/dployrd/.dployr/logs/caddy
        
        mkdir -p /home/dployrd/.version-fox/temp
        chown -R dployrd:dployrd /home/dployrd/.version-fox
        chmod -R 755 /home/dployrd/.version-fox
        
        info "Setting up vfox plugins..."
        for plugin in nodejs python golang php java dotnet ruby; do
            sudo -u dployrd bash -c "cd /var/lib/dployrd && vfox add $plugin" || warn "Failed to add $plugin plugin"
        done

        cat > /etc/systemd/system/dployrd.service << 'EOF'
[Unit]
Description=Dployr Daemon
After=network.target

[Service]
Type=simple
User=dployrd
Group=dployrd
ExecStart=/usr/local/bin/dployrd
WorkingDirectory=/var/lib/dployrd
StandardOutput=append:/var/log/dployrd/dployrd.log
StandardError=append:/var/log/dployrd/dployrd.log
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
EOF
        systemctl daemon-reload
        systemctl enable dployrd
        systemctl start dployrd
        info "dployrd service started"

        [[ -n "$TOKEN" ]] && { sleep 1; register_instance "$TOKEN" || true; }
        ;;
    darwin)
        if ! dscl . -read /Users/_dployrd &>/dev/null; then
            dscl . -create /Users/_dployrd
            dscl . -create /Users/_dployrd UserShell /usr/bin/false
            dscl . -create /Users/_dployrd RealName "dployr Daemon"
            dscl . -create /Users/_dployrd UniqueID 501
            dscl . -create /Users/_dployrd PrimaryGroupID 20
            dscl . -create /Users/_dployrd NFSHomeDirectory /var/lib/dployrd
            dscl . -append /Groups/dployr-admin GroupMembership _dployrd
            log_json "info" "Created _dployrd system user"
        fi
        
        mkdir -p /var/log/dployrd /var/lib/dployrd
        chown _dployrd:staff /var/log/dployrd /var/lib/dployrd
        
        mkdir -p /var/lib/dployrd/.version-fox/temp
        chown -R _dployrd:staff /var/lib/dployrd/.version-fox
        chmod -R 755 /var/lib/dployrd/.version-fox
        
        info "Setting up vfox plugins..."
        for plugin in nodejs python golang php java dotnet ruby; do
            sudo -u _dployrd bash -c "cd /var/lib/dployrd && vfox add $plugin" || warn "Failed to add $plugin plugin"
        done
        
        cat > /Library/LaunchDaemons/io.dployr.dployrd.plist << 'EOF'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>io.dployr.dployrd</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/dployrd</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>WorkingDirectory</key>
    <string>/var/lib/dployrd</string>
    <key>StandardOutPath</key>
    <string>/var/log/dployrd/dployrd.log</string>
    <key>StandardErrorPath</key>
    <string>/var/log/dployrd/dployrd.log</string>
    <key>UserName</key>
    <string>_dployrd</string>
</dict>
</plist>
EOF
        launchctl load /Library/LaunchDaemons/io.dployr.dployrd.plist
        launchctl start io.dployr.dployrd
        info "dployrd service started"

        # Restart Caddy (launchd) after directories are owned by _dployrd
        launchctl kickstart -k system/com.caddyserver.caddy 2>/dev/null || launchctl start com.caddyserver.caddy || warn "Failed to restart Caddy"

        [[ -n "$TOKEN" ]] && { sleep 1; register_instance "$TOKEN" || true; }
        ;;
esac

if [[ -n "$SHELL" && -f "$HOME/.bashrc" ]]; then
    # shellcheck disable=SC1091
    source "$HOME/.bashrc" 2>&1 || true
    eval "$(vfox activate bash)" 2>&1 || true
else
    eval "$(vfox activate bash)" 2>&1 || true
fi

info "Setting up sudo permissions..."
SYSTEMCTL=$(command -v systemctl)
TEE=$(command -v tee)
MKDIR=$(command -v mkdir)
RM=$(command -v rm)
CP=$(command -v cp)
CHMOD=$(command -v chmod)
DOCKER=$(command -v docker)

for cmd in SYSTEMCTL TEE MKDIR RM CP CHMOD; do
    [[ -z "${!cmd}" ]] && error "Command $cmd not found"
done

REBOOT=$(command -v reboot)
[[ -z "$REBOOT" ]] && error "Command reboot not found"

cat > /etc/sudoers.d/dployr << EOF
dployrd ALL=(ALL) NOPASSWD: $SYSTEMCTL *
dployrd ALL=(ALL) NOPASSWD: $REBOOT
dployrd ALL=(ALL) NOPASSWD: $MKDIR *
dployrd ALL=(ALL) NOPASSWD: $RM *
dployrd ALL=(ALL) NOPASSWD: $CP *
dployrd ALL=(ALL) NOPASSWD: $CHMOD *
dployrd ALL=(ALL) NOPASSWD: $TEE *
dployrd ALL=(ALL) NOPASSWD: $DOCKER *
EOF
chmod 440 /etc/sudoers.d/dployr

rm -rf "$TEMP_DIR"

END_TIME=$(date +%s)
DURATION=$((END_TIME - START_TIME))
log_json "info" "Installation completed in ${DURATION}s"

cat >&3 << EOF

Installation completed successfully!

Installed components:
  - dployr (CLI)
  - dployrd (daemon)
  - caddy (reverse proxy)

Next steps:
EOF

[[ $EUID -ne 0 ]] && echo "- Add $INSTALL_DIR to your PATH" >&3
[[ -n "$DPLOYR_DOMAIN" ]] && echo "- Instance URL: https://$DPLOYR_DOMAIN" >&3
cat >&3 << EOF
- The dployrd daemon is running as a system service
- Use the CLI: dployr --help

Service management:
EOF

case $OS in
    linux)
        cat >&3 << EOF
- Status: systemctl status dployrd
- Stop: systemctl stop dployrd
- Restart: systemctl restart dployrd
EOF
        ;;
    darwin)
        cat >&3 << EOF
- Status: launchctl list | grep dployrd
- Stop: launchctl stop io.dployr.dployrd
- Restart: launchctl kickstart -k system/io.dployr.dployrd
EOF
        ;;
esac
