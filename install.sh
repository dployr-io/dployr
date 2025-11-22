#!/bin/bash

# dployr Unix Installer (Linux/macOS)
# Downloads and installs dployr, dployrd, and Caddy

set -e

LOG_DIR="/var/log/dployrd"
if [[ $EUID -eq 0 ]]; then
    mkdir -p "$LOG_DIR"
    LOG_FILE="$LOG_DIR/install.log"
else
    LOG_DIR="$HOME/.dployr"
    mkdir -p "$LOG_DIR"
    LOG_FILE="$LOG_DIR/install.log"
fi

exec >>"$LOG_FILE" 2>&1
echo "Logging installer output to $LOG_FILE"

INSTALL_DIR="/usr/local/bin"
VERSION="${1:-latest}"
TOKEN="$2"
PUBLIC_IP_V4="$3"
REPO="dployr-io/dployr"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info() { echo -e "${GREEN}[INFO]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

echo "dployr Unix Installer"
echo "===================="

# Show usage if help requested
if [[ "$1" == "--help" || "$1" == "-h" ]]; then
    echo "Usage: $0 [version] token [public_ip_v4]"
    echo ""
    echo "Arguments:"
    echo "  version       Optional dployr version tag (default: latest)"
    echo "  token         Install token obtained from dployr base"
    echo "  public_ip_v4  Optional public IP address of the instance"
    echo ""
    echo "Examples:"
    echo "  $0 latest eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..." 
    echo "  $0 v0.1.1-beta.17 eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..." 
    echo ""
    echo "Available versions: https://github.com/dployr-io/dployr/releases"
    exit 0
fi

if [[ -z "$TOKEN" ]]; then
    error "Missing install token argument. Run with: $0 [VERSION] token.\n\n Visit https://docs.dployr.dev/installation for more information."
fi

# Simplified best effort for Public IP detection
# If behind a Proxy, this might fail and would generally
# be recommended to provide the IP during installation
# See https://docs.dployr.dev/installation for more information
detect_public_ip_address() {
    local ip=""
    
    # Try external service for public IP (handles NAT/VPS)
    ip=$(curl -s --connect-timeout 2 --max-time 3 "https://api.ipify.org" 2>/dev/null)
    
    # Validate it's a public IP
    if [[ "$ip" =~ ^[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}$ ]] && is_public_ip "$ip"; then
        PUBLIC_IP_V4="$ip"
        return 0
    fi
    
    # Fallback to local interface
    if [[ $OS == "linux" ]]; then
        ip=$(hostname -I 2>/dev/null | awk '{print $1}')
    elif [[ $OS == "darwin" ]]; then
        ip=$(ifconfig 2>/dev/null | grep "inet " | grep -v "127.0.0.1" | awk '{print $2}' | head -n1)
    fi
    
    PUBLIC_IP_V4="${ip:-127.0.0.1}"
}

register_instance() {
    local token="$1"
    
    if [[ -z "$PUBLIC_IP_V4" ]]; then
        info "Public IP was not provided. Attempting auto-detection..."

        detect_public_ip_address
    fi

    info "Registering instance with base..."
    if ! curl -sS -X POST \
        -H "Content-Type: application/json" \
        -d "{\"token\":\"$token\", \"address\":\"$PUBLIC_IP_V4\"}" \
        "http://localhost:7879/system/domain"; then
        warn "Failed to register instance with base. Visit https://docs.dployr.dev/installation for more information."
        return 1
    fi

    echo
    info "Instance registration request sent successfully"
}

# Check if running as root for system install
if [[ $EUID -eq 0 ]]; then
    info "Installing system-wide to $INSTALL_DIR"
else
    INSTALL_DIR="$HOME/.local/bin"
    info "Installing to user directory: $INSTALL_DIR"
    mkdir -p "$INSTALL_DIR"
fi

# Detect platform and architecture
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

# Get latest version
if [[ "$VERSION" == "latest" ]]; then
    info "Fetching latest release..."
    VERSION=$(curl -s "https://api.github.com/repos/$REPO/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    [[ -z "$VERSION" ]] && error "Failed to get latest version"
    info "Latest version: $VERSION"
fi

# Download dployr binaries
info "Downloading dployr binaries..."
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/dployr-$PLATFORM-$ARCH.tar.gz"
TEMP_DIR=$(mktemp -d)

curl -L "$DOWNLOAD_URL" -o "$TEMP_DIR/dployr.tar.gz" || error "Failed to download dployr"
tar -xzf "$TEMP_DIR/dployr.tar.gz" -C "$TEMP_DIR" || error "Failed to extract dployr"

# The archive contains a directory named dployr-$PLATFORM-$ARCH
EXTRACT_DIR="$TEMP_DIR/dployr-$PLATFORM-$ARCH"

# Stop running daemon if it exists
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
    sleep 2  # Give it time to stop
fi

# Install binaries
info "Installing dployr binaries..."
if ! cp "$EXTRACT_DIR/dployr" "$INSTALL_DIR/"; then
    error "Failed to copy dployr"
fi
chmod +x "$INSTALL_DIR/dployr" || error "Failed to make dployr executable"

if ! cp "$EXTRACT_DIR/dployrd" "$INSTALL_DIR/"; then
    error "Failed to copy dployrd"
fi
chmod +x "$INSTALL_DIR/dployrd" || error "Failed to make dployrd executable"

# Install Caddy
info "Installing Caddy..."
if command -v caddy &> /dev/null; then
    info "Caddy already installed"
else
    case $OS in
        linux)
            if [[ $EUID -eq 0 ]] && command -v apt &> /dev/null; then
                info "Installing Caddy via apt..."

                # Wait for any other apt processes
                info "Waiting for other apt processes to finish..."
                while sudo fuser /var/lib/apt/lists/lock >/dev/null 2>&1; do
                    sleep 2
                done

                apt update
                apt install -y debian-keyring debian-archive-keyring apt-transport-https

                curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' \
                    | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg

                curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' \
                    | tee /etc/apt/sources.list.d/caddy-stable.list

                # Wait again for apt to be free before installing Caddy
                while sudo fuser /var/lib/apt/lists/lock >/dev/null 2>&1; do
                    sleep 2
                done

                apt update
                apt install -y caddy
            else
                info "Installing Caddy binary..."
                CADDY_URL="https://github.com/caddyserver/caddy/releases/latest/download/caddy_${OS}_${ARCH}.tar.gz"
                curl -L "$CADDY_URL" -o "$TEMP_DIR/caddy.tar.gz"
                tar -xzf "$TEMP_DIR/caddy.tar.gz" -C "$TEMP_DIR"
                cp "$TEMP_DIR/caddy" "$INSTALL_DIR/" && chmod +x "$INSTALL_DIR/caddy"
            fi
            ;;
        darwin)
            if command -v brew &> /dev/null; then
                info "Installing Caddy via Homebrew..."
                brew install caddy
            else
                info "Installing Caddy binary..."
                CADDY_URL="https://github.com/caddyserver/caddy/releases/latest/download/caddy_${OS}_${ARCH}.tar.gz"
                curl -L "$CADDY_URL" -o "$TEMP_DIR/caddy.tar.gz"
                tar -xzf "$TEMP_DIR/caddy.tar.gz" -C "$TEMP_DIR"
                cp "$TEMP_DIR/caddy" "$INSTALL_DIR/" && chmod +x "$INSTALL_DIR/caddy"
            fi
            ;;
    esac
fi


# Install vfox
info "Installing vfox..."
if command -v vfox &> /dev/null; then
    info "vfox already installed"
else
    info "Installing vfox using official installer..."
    curl -sSL https://raw.githubusercontent.com/version-fox/vfox/main/install.sh | bash || error "Failed to install vfox"
    info "vfox installed successfully!"
fi

# Setup vfox plugins for dployrd user
info "Setting up vfox plugins..."


# Create system-wide config directory and file
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
if [[ $EUID -ne 0 ]]; then
    error "System-wide installation requires root privileges. Run with sudo."
fi

mkdir -p "$CONFIG_DIR"

if [[ ! -f "$CONFIG_FILE" ]]; then
    cat > "$CONFIG_FILE" << EOF
# dployr configuration file
address = "localhost"
port = 7879
max-workers = 5

# Base configuration
base_url = "https://base.dployr.dev"
base_jwks_url = "https://base.dployr.dev/.well-known/jwks.json"
instance_id = "my-instance-id"
EOF
    # Set proper permissions (root-owned, readable by all)
    chmod 644 "$CONFIG_FILE"
    chmod 755 "$CONFIG_DIR"
    info "Created system config at $CONFIG_FILE"
else
    info "Config file already exists at $CONFIG_FILE"
fi

# Install and start system service
info "Setting up dployrd service..."
case $OS in
    linux)
        # Create dployr system groups
        for group in dployr-owner dployr-admin dployr-dev dployr-viewer; do
            if ! getent group "$group" &>/dev/null; then
                groupadd "$group"
                info "Created group: $group"
            fi
        done
        
        # Create dployrd user if it doesn't exist
        if ! id "dployrd" &>/dev/null; then
            useradd --system --create-home --shell /bin/bash -G dployr-admin dployrd
            info "Created dployrd system user"
        fi
        
        # Create log and working directories
        mkdir -p /var/log/dployrd /var/lib/dployrd
        chown dployrd:dployrd /var/log/dployrd /var/lib/dployrd
        
        # Setup vfox directory for dployrd user
        mkdir -p /home/dployrd/.version-fox/temp
        chown -R dployrd:dployrd /home/dployrd/.version-fox
        chmod -R 755 /home/dployrd/.version-fox
        
        # Setup vfox plugins for common runtimes
        info "Setting up vfox plugins for dployrd user..."
        sudo -u dployrd bash -c 'cd /var/lib/dployrd && vfox add nodejs' || warn "Failed to add nodejs plugin"
        sudo -u dployrd bash -c 'cd /var/lib/dployrd && vfox add python' || warn "Failed to add python plugin"
        sudo -u dployrd bash -c 'cd /var/lib/dployrd && vfox add golang' || warn "Failed to add golang plugin"
        sudo -u dployrd bash -c 'cd /var/lib/dployrd && vfox add php' || warn "Failed to add php plugin"
        sudo -u dployrd bash -c 'cd /var/lib/dployrd && vfox add java' || warn "Failed to add java plugin"
        sudo -u dployrd bash -c 'cd /var/lib/dployrd && vfox add dotnet' || warn "Failed to add dotnet plugin"
        sudo -u dployrd bash -c 'cd /var/lib/dployrd && vfox add ruby' || warn "Failed to add ruby plugin"

        # Create systemd service file
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
        info "dployrd service started and enabled"

		sleep 1
		register_instance "$TOKEN" "$PUBLIC_IP_V4" || true
        ;;
    darwin)
        # Create dployr system groups (macOS)
        for group in dployr-owner dployr-admin dployr-dev dployr-viewer; do
            if ! dscl . -read /Groups/"$group" &>/dev/null; then
                dscl . -create /Groups/"$group"
                dscl . -create /Groups/"$group" PrimaryGroupID "$(dscl . -list /Groups PrimaryGroupID | awk '{print $2}' | sort -n | tail -1 | awk '{print $1+1}')"
                info "Created group: $group"
            fi
        done
        
        # Create dployrd user if it doesn't exist (macOS uses _username convention)
        if ! dscl . -read /Users/_dployrd &>/dev/null; then
            dscl . -create /Users/_dployrd
            dscl . -create /Users/_dployrd UserShell /usr/bin/false
            dscl . -create /Users/_dployrd RealName "dployr Daemon"
            dscl . -create /Users/_dployrd UniqueID 501
            dscl . -create /Users/_dployrd PrimaryGroupID 20
            dscl . -create /Users/_dployrd NFSHomeDirectory /var/lib/dployrd
            dscl . -append /Groups/dployr-admin GroupMembership _dployrd
            info "Created _dployrd system user"
        fi
        
        # Create log and working directories for macOS
        mkdir -p /var/log/dployrd /var/lib/dployrd
        chown _dployrd:staff /var/log/dployrd /var/lib/dployrd
        
        # Setup vfox directory for dployrd user
        mkdir -p /var/lib/dployrd/.version-fox/temp
        chown -R _dployrd:staff /var/lib/dployrd/.version-fox
        chmod -R 755 /var/lib/dployrd/.version-fox
        
        # Setup vfox plugins for common runtimes
        info "Setting up vfox plugins for _dployrd user..."
        sudo -u _dployrd bash -c 'cd /var/lib/dployrd && vfox add nodejs' || warn "Failed to add nodejs plugin"
        sudo -u _dployrd bash -c 'cd /var/lib/dployrd && vfox add python' || warn "Failed to add python plugin"
        sudo -u _dployrd bash -c 'cd /var/lib/dployrd && vfox add golang' || warn "Failed to add golang plugin"
        sudo -u _dployrd bash -c 'cd /var/lib/dployrd && vfox add php' || warn "Failed to add php plugin"
        sudo -u _dployrd bash -c 'cd /var/lib/dployrd && vfox add java' || warn "Failed to add java plugin"
        sudo -u _dployrd bash -c 'cd /var/lib/dployrd && vfox add dotnet' || warn "Failed to add dotnet plugin"
        sudo -u _dployrd bash -c 'cd /var/lib/dployrd && vfox add ruby' || warn "Failed to add ruby plugin"
        
        # Create launchd plist
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
        info "dployrd service started and enabled"

		sleep 1
		register_instance "$TOKEN" "$PUBLIC_IP_V4" || true
        ;;
esac

# Reload vfox environment for current shell
info "Reloading vfox environment..."
if [[ -n "$SHELL" && -f "$HOME/.bashrc" ]]; then
    # shellcheck source=/dev/null
    source "$HOME/.bashrc" 2>&1 || warn "Failed to source ~/.bashrc: $?"
    eval "$(vfox activate bash)" || warn "Failed to reload vfox environment"
else
    eval "$(vfox activate bash)" || warn "Failed to reload vfox environment"
fi

# Add safe sudo permissions for dployrd user
info "Setting up sudo permissions for dployrd user..."
SYSTEMCTL=$(command -v systemctl)
TEE=$(command -v tee)
CADDY=$(command -v caddy)
MKDIR=$(command -v mkdir)
RM=$(command -v rm)

for cmd in SYSTEMCTL TEE CADDY MKDIR RM; do
    if [ -z "${!cmd}" ]; then
        error "Command $cmd not found. Cannot configure sudoers."
    fi
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
info "Sudo permissions configured for dployrd user"

# Cleanup
rm -rf "$TEMP_DIR"

echo ""
echo "Installation completed successfully!"
echo ""
echo "Installed components:"
echo "  - dployr (CLI)"
echo "  - dployrd (daemon)"
echo "  - caddy (reverse proxy)"
echo ""

echo "Next steps:"
if [[ $EUID -ne 0 ]]; then
    echo "- Add $INSTALL_DIR to your PATH if not already done"
fi
echo "- The dployrd daemon is now running as a system service"
echo "- Use the CLI: dployr --help"
echo ""
echo "Service management:"
case $OS in
    linux)
        echo "- Status: systemctl status dployrd"
        echo "- Stop: systemctl stop dployrd"
        echo "- Restart: systemctl restart dployrd"
        ;;
    darwin)
        echo "- Status: launchctl list | grep dployrd"
        echo "- Stop: launchctl stop io.dployr.dployrd"
        echo "- Restart: launchctl kickstart -k system/io.dployr.dployrd"
        ;;
esac