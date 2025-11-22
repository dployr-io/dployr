#!/bin/bash

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

INSTALL_DIR="/usr/local/bin"
VERSION="latest"
TOKEN=""
REPO="dployr-io/dployr"
DPLOYR_DOMAIN=""

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

install_jq() {
    if command -v jq &>/dev/null; then
        return 0
    fi

    info "Installing jq..."
    
    case $OS in
        linux)
            if command -v apt &>/dev/null; then
                apt update -qq && apt install -y -qq jq
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

if [[ "$1" == "--help" || "$1" == "-h" ]]; then
    cat >&3 << EOF
Usage: $0 [--version <VERSION>] [--token <TOKEN>]

Arguments:
  --version, -v  Optional dployr version tag (default: latest)
  --token, -t    Optional install token obtained from dployr base

Examples:
  $0 --version v0.3.1-beta.9
  $0 --version v0.3.1-beta.9 --token eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...

Available versions: https://github.com/dployr-io/dployr/releases
EOF
    exit 0
fi

while [[ $# -gt 0 ]]; do
    case "$1" in
        --version|-v)
            [[ -z "$2" ]] && error "Missing value for $1"
            VERSION="$2"
            shift 2
            ;;
        --token|-t)
            [[ -z "$2" ]] && error "Missing value for $1"
            TOKEN="$2"
            shift 2
            ;;
        *)
            error "Unknown argument: $1"
            ;;
    esac
done

register_instance() {
    local token="$1"

    info "Registering instance with base..."

    local response
    response=$(curl -sS -X POST \
        -H "Content-Type: application/json" \
        -d "{\"token\":\"$token\"}" \
        "http://localhost:7879/system/domain" 2>&1)
    local status=$?

    if [[ $status -ne 0 ]]; then
        warn "Failed to register instance (curl exit $status). Visit https://docs.dployr.dev/installation"
        log_json "error" "$response"
        return 1
    fi

    log_json "info" "Registration response received"

    DPLOYR_DOMAIN=$(echo "$response" | parse_json '.domain')
    if [[ -n "$DPLOYR_DOMAIN" ]]; then
        info "Instance registered successfully. URL: https://$DPLOYR_DOMAIN"
        log_json "info" "Instance registered with domain: $DPLOYR_DOMAIN"
    else
        error "No domain received from base. Please check your token or see https://docs.dployr.dev/installation for help."
        log_json "error" "Registration failed, domain not present in response: $response"
        return 1
    fi
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
    info "Stopping running dployrd daemon..."
    case $OS in
        linux)
            if systemctl is-active --quiet dployrd 2>/dev/null; then
                systemctl stop dployrd
            else
                pkill -x dployrd
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

                apt update -qq
                apt install -y -qq debian-keyring debian-archive-keyring apt-transport-https

                curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' \
                    | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg

                curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' \
                    | tee /etc/apt/sources.list.d/caddy-stable.list >/dev/null

                while sudo fuser /var/lib/apt/lists/lock >/dev/null 2>&1; do
                    sleep 2
                done

                apt update -qq
                apt install -y -qq caddy
            else
                CADDY_URL="https://github.com/caddyserver/caddy/releases/latest/download/caddy_${OS}_${ARCH}.tar.gz"
                curl -sL "$CADDY_URL" -o "$TEMP_DIR/caddy.tar.gz"
                tar -xzf "$TEMP_DIR/caddy.tar.gz" -C "$TEMP_DIR"
                cp "$TEMP_DIR/caddy" "$INSTALL_DIR/"
                chmod +x "$INSTALL_DIR/caddy"
            fi
            ;;
        darwin)
            if command -v brew &> /dev/null; then
                info "Installing Caddy via Homebrew..."
                brew install -q caddy
            else
                CADDY_URL="https://github.com/caddyserver/caddy/releases/latest/download/caddy_${OS}_${ARCH}.tar.gz"
                curl -sL "$CADDY_URL" -o "$TEMP_DIR/caddy.tar.gz"
                tar -xzf "$TEMP_DIR/caddy.tar.gz" -C "$TEMP_DIR"
                cp "$TEMP_DIR/caddy" "$INSTALL_DIR/"
                chmod +x "$INSTALL_DIR/caddy"
            fi
            ;;
    esac
fi

info "Installing vfox..."
if command -v vfox &> /dev/null; then
    info "vfox already installed"
else
    curl -sSL https://raw.githubusercontent.com/version-fox/vfox/main/install.sh | bash || error "Failed to install vfox"
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
    cat > "$CONFIG_FILE" << EOF
address = "localhost"
port = 7879
max-workers = 5

base_url = "https://base.dployr.dev"
base_jwks_url = "https://base.dployr.dev/.well-known/jwks.json"
instance_id = "my-instance-id"
EOF
    chmod 644 "$CONFIG_FILE"
    chmod 755 "$CONFIG_DIR"
    info "Created system config at $CONFIG_FILE"
else
    info "Config file already exists at $CONFIG_FILE"
fi

info "Setting up dployrd service..."
case $OS in
    linux)
        for group in dployr-owner dployr-admin dployr-dev dployr-viewer; do
            if ! getent group "$group" &>/dev/null; then
                groupadd "$group"
                log_json "info" "Created group: $group"
            fi
        done
        
        if ! id "dployrd" &>/dev/null; then
            useradd --system --create-home --shell /bin/bash -G dployr-admin dployrd
            log_json "info" "Created dployrd system user"
        fi
        
        mkdir -p /var/log/dployrd /var/lib/dployrd
        chown dployrd:dployrd /var/log/dployrd /var/lib/dployrd
        
        mkdir -p /home/dployrd/.version-fox/temp
        chown -R dployrd:dployrd /home/dployrd/.version-fox
        chmod -R 755 /home/dployrd/.version-fox
        
        info "Setting up vfox plugins..."
        for plugin in nodejs python golang php java dotnet ruby; do
            sudo -u dployrd bash -c "cd /var/lib/dployrd && vfox add $plugin" || warn "Failed to add $plugin plugin"
        done

        cat > /etc/systemd/system/dployrd.service << 'EOF'
[Unit]
Description=dployr Daemon
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
        for group in dployr-owner dployr-admin dployr-dev dployr-viewer; do
            if ! dscl . -read /Groups/"$group" &>/dev/null; then
                local gid
                gid=$(dscl . -list /Groups PrimaryGroupID | awk '{print $2}' | sort -n | tail -1 | awk '{print $1+1}')
                dscl . -create /Groups/"$group"
                dscl . -create /Groups/"$group" PrimaryGroupID "$gid"
                log_json "info" "Created group: $group"
            fi
        done
        
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

        [[ -n "$TOKEN" ]] && { sleep 1; register_instance "$TOKEN" || true; }
        ;;
esac

if [[ -n "$SHELL" && -f "$HOME/.bashrc" ]]; then
    source "$HOME/.bashrc" 2>&1 || true
    eval "$(vfox activate bash)" 2>&1 || true
else
    eval "$(vfox activate bash)" 2>&1 || true
fi

info "Setting up sudo permissions..."
SYSTEMCTL=$(command -v systemctl)
TEE=$(command -v tee)
CADDY=$(command -v caddy)
MKDIR=$(command -v mkdir)
RM=$(command -v rm)

for cmd in SYSTEMCTL TEE CADDY MKDIR RM; do
    [[ -z "${!cmd}" ]] && error "Command $cmd not found"
done

cat > /etc/sudoers.d/dployr << EOF
dployrd ALL=(ALL) NOPASSWD: $SYSTEMCTL daemon-reload
dployrd ALL=(ALL) NOPASSWD: $SYSTEMCTL start *
dployrd ALL=(ALL) NOPASSWD: $SYSTEMCTL stop *
dployrd ALL=(ALL) NOPASSWD: $SYSTEMCTL restart *
dployrd ALL=(ALL) NOPASSWD: $SYSTEMCTL reload *
dployrd ALL=(ALL) NOPASSWD: $SYSTEMCTL enable *
dployrd ALL=(ALL) NOPASSWD: $SYSTEMCTL disable *
dployrd ALL=(ALL) NOPASSWD: $SYSTEMCTL is-active *
dployrd ALL=(ALL) NOPASSWD: $MKDIR -p /etc/systemd/system
dployrd ALL=(ALL) NOPASSWD: $RM -f /etc/systemd/system/*.service
dployrd ALL=(ALL) NOPASSWD: $TEE /etc/systemd/system/*.service
dployrd ALL=(ALL) NOPASSWD: $TEE /etc/caddy/Caddyfile
dployrd ALL=(ALL) NOPASSWD: $CADDY validate --config /etc/caddy/Caddyfile --adapter caddyfile
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
