#!/bin/bash

# dployr Unix Installer (Linux/macOS)
# Downloads and installs dployr, dployrd, and Caddy

set -e

INSTALL_DIR="/usr/local/bin"
VERSION="latest"
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

# Install binaries
info "Installing dployr binaries..."
if [[ $EUID -eq 0 ]]; then
    cp "$TEMP_DIR/dployr" "$INSTALL_DIR/" && chmod +x "$INSTALL_DIR/dployr"
    cp "$TEMP_DIR/dployrd" "$INSTALL_DIR/" && chmod +x "$INSTALL_DIR/dployrd"
else
    cp "$TEMP_DIR/dployr" "$INSTALL_DIR/" && chmod +x "$INSTALL_DIR/dployr"
    cp "$TEMP_DIR/dployrd" "$INSTALL_DIR/" && chmod +x "$INSTALL_DIR/dployrd"
fi

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
    echo "1. Add $INSTALL_DIR to your PATH if not already done"
fi
echo "2. Start the daemon: dployrd"
echo "3. Use the CLI: dployr --help"