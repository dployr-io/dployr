#!/bin/bash

# dployr Unix Installer (Linux/macOS)
# Downloads and installs dployr, dployrd, and Caddy

set -e

INSTALL_DIR="/usr/local/bin"
VERSION="${1:-latest}"
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
    echo "Usage: $0 [VERSION]"
    echo ""
    echo "Examples:"
    echo "  $0                    # Install latest version"
    echo "  $0 v0.1.1-beta.17     # Install specific version"
    echo ""
    echo "Available versions: https://github.com/dployr-io/dployr/releases"
    exit 0
fi

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
cp "$EXTRACT_DIR/dployr" "$INSTALL_DIR/" && chmod +x "$INSTALL_DIR/dployr" || error "Failed to install dployr"
cp "$EXTRACT_DIR/dployrd" "$INSTALL_DIR/" && chmod +x "$INSTALL_DIR/dployrd" || error "Failed to install dployrd"

# Install Caddy
info "Installing Caddy..."
if command -v caddy &> /dev/null; then
    info "Caddy already installed"
else
    case $OS in
        linux)
            if command -v apt &> /dev/null; then
                info "Installing Caddy via apt..."
                if [[ $EUID -eq 0 ]]; then
                    apt update && apt install -y debian-keyring debian-archive-keyring apt-transport-https
                    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
                    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | tee /etc/apt/sources.list.d/caddy-stable.list
                    apt update && apt install caddy
                else
                    warn "Need root privileges to install Caddy via apt. Installing binary instead..."
                    CADDY_URL="https://github.com/caddyserver/caddy/releases/latest/download/caddy_${OS}_${ARCH}.tar.gz"
                    curl -L "$CADDY_URL" -o "$TEMP_DIR/caddy.tar.gz"
                    tar -xzf "$TEMP_DIR/caddy.tar.gz" -C "$TEMP_DIR"
                    cp "$TEMP_DIR/caddy" "$INSTALL_DIR/" && chmod +x "$INSTALL_DIR/caddy"
                fi
            elif command -v yum &> /dev/null || command -v dnf &> /dev/null; then
                info "Installing Caddy via dnf/yum..."
                if [[ $EUID -eq 0 ]]; then
                    dnf copr enable @caddy/caddy -y && dnf install caddy -y
                else
                    warn "Need root privileges to install Caddy via dnf. Installing binary instead..."
                    CADDY_URL="https://github.com/caddyserver/caddy/releases/latest/download/caddy_${OS}_${ARCH}.tar.gz"
                    curl -L "$CADDY_URL" -o "$TEMP_DIR/caddy.tar.gz"
                    tar -xzf "$TEMP_DIR/caddy.tar.gz" -C "$TEMP_DIR"
                    cp "$TEMP_DIR/caddy" "$INSTALL_DIR/" && chmod +x "$INSTALL_DIR/caddy"
                fi
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
    # Generate a secure random secret
    SECRET=$(openssl rand -hex 32 2>/dev/null || head -c 32 /dev/urandom | base64 | tr -d '=+/' | cut -c1-32)
    
    cat > "$CONFIG_FILE" << EOF
# dployr configuration file
address = "localhost"
port = 7879
max-workers = 5

# Secret key
secret = "$SECRET"
EOF
    # Set proper permissions (root-owned, readable by all)
    chmod 644 "$CONFIG_FILE"
    chmod 755 "$CONFIG_DIR"
    info "Created system config at $CONFIG_FILE"
    SHOW_SECRET="$SECRET"
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
            useradd --system --create-home --shell /bin/false -G dployr-admin dployrd
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
        ;;
esac

# Reload vfox environment for current shell
info "Reloading vfox environment..."
if [[ -n "$SHELL" && -f "$HOME/.bashrc" ]]; then
    # shellcheck source=/dev/null
    source "$HOME/.bashrc"
    eval "$(vfox activate bash)" || warn "Failed to reload vfox environment"
else
    eval "$(vfox activate bash)" || warn "Failed to reload vfox environment"
fi

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

# Show the generated secret once
if [[ -n "$SHOW_SECRET" ]]; then
    echo "=========================================="
    echo "YOUR SECRET KEY (SAVE THIS NOW!):"
    echo ""
    echo "  $SHOW_SECRET"
    echo ""
    echo "This secret will NOT be shown again!"
    echo "It's saved in: $CONFIG_FILE"
    echo "=========================================="
    echo ""
fi

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