#!/usr/bin/env bash

# Copyright 2025 Emmanuel Madehin
# SPDX-License-Identifier: Apache-2.0

# install.sh — installs dployrd from GitHub releases
# 
# Usage: install.sh <version> [token]
#   version: GitHub release tag (e.g., "v1.2.3") or "latest"
#   token: Optional authentication token (reserved for future use)

set -e

VERSION="${1:-latest}"
TOKEN="${2:-}"
REPO="dployr-io/dployr"
INSTALL_DIR="/usr/local/bin"
TEMP_DIR=$(mktemp -d)

cleanup() {
    rm -rf "$TEMP_DIR"
}
trap cleanup EXIT

log() {
    echo "[$(date -Iseconds)] $*" >&2
}

error() {
    log "ERROR: $*"
    exit 1
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
# in-progress deployments reported by the daemon.
wait_for_pending_tasks() {
    if ! pgrep -x "dployrd" >/dev/null 2>&1; then
        return 0
    fi

    if ! command -v curl >/dev/null 2>&1; then
        return 0
    fi

    local port
    port=$(get_daemon_port)

    log "Checking for pending tasks before upgrade (port $port)..."

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
    local max_attempts=12 # ~1 minute total at 5s intervals

    while (( attempts < max_attempts )); do
        attempts=$((attempts + 1))

        local resp
        # exclude_system=true tells the daemon to not count system/* tasks (including this install)
        resp=$(curl -sS ${auth_header:+-H "$auth_header"} "http://localhost:${port}/system/tasks?status=pending&exclude_system=true" 2>/dev/null || true)
        if [[ -z "$resp" ]]; then
            log "Could not query dployrd for pending tasks; continuing with install."
            return 0
        fi

        # Simple count extraction without jq dependency
        local count
        count=$(echo "$resp" | grep -o '"count":[0-9]*' | grep -o '[0-9]*' || echo "")
        if [[ -z "$count" ]]; then
            log "Could not parse pending tasks response; continuing with install."
            return 0
        fi

        if [[ "$count" -eq 0 ]]; then
            log "No pending tasks detected. Proceeding with install."
            return 0
        fi

        log "There are $count pending tasks. Waiting for them to finish (attempt $attempts/$max_attempts)..."
        sleep 5
    done

    log "Timed out waiting for pending tasks. Proceeding with install."
}

# Detect OS and architecture
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

log "Installing dployrd version: $VERSION for $PLATFORM-$ARCH"

# Fetch latest version if not specified
if [[ "$VERSION" == "latest" ]]; then
    log "Fetching latest release version..."
    VERSION=$(curl -sS "https://api.github.com/repos/$REPO/releases/latest" | grep -o '"tag_name": "[^"]*' | cut -d'"' -f4)
    [[ -z "$VERSION" ]] && error "Failed to fetch latest version"
    log "Latest version: $VERSION"
fi

# Download the release
DOWNLOAD_URL="https://github.com/$REPO/releases/download/$VERSION/dployr-$PLATFORM-$ARCH.tar.gz"
log "Downloading from: $DOWNLOAD_URL"

curl -sL "$DOWNLOAD_URL" -o "$TEMP_DIR/dployr.tar.gz" || error "Failed to download dployr"
tar -xzf "$TEMP_DIR/dployr.tar.gz" -C "$TEMP_DIR" || error "Failed to extract dployr"

EXTRACT_DIR="$TEMP_DIR/dployr-$PLATFORM-$ARCH"

# Wait for any pending tasks to complete before upgrading
wait_for_pending_tasks

# Install binaries while daemon is still running
log "Installing dployrd binary..."
sudo rm -f "$INSTALL_DIR/dployrd"
sudo cp "$EXTRACT_DIR/dployrd" "$INSTALL_DIR/" || error "Failed to copy dployrd"
sudo chmod +x "$INSTALL_DIR/dployrd"

log "Installing dployr CLI binary..."
sudo rm -f "$INSTALL_DIR/dployr"
sudo cp "$EXTRACT_DIR/dployr" "$INSTALL_DIR/" || error "Failed to copy dployr"
sudo chmod +x "$INSTALL_DIR/dployr"

# Restart the daemon
log "Restarting dployrd daemon..."
case $OS in
    linux)
        if systemctl is-enabled --quiet dployrd 2>/dev/null; then
            nohup bash -c 'sleep 1; sudo systemctl restart dployrd' >/dev/null 2>&1 &
        fi
        ;;
    darwin)
        if launchctl list 2>/dev/null | grep -q io.dployr.dployrd; then
            nohup bash -c 'sleep 1; sudo launchctl kickstart -k system/io.dployr.dployrd' >/dev/null 2>&1 &
        fi
        ;;
esac

log "Installation completed successfully: $VERSION"
echo "$VERSION"