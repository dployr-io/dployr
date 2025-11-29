#!/usr/bin/env bash

# Copyright 2025 Emmanuel Madehin
# SPDX-License-Identifier: Apache-2.0

set -euo pipefail

log() { echo "[doctor] $*"; }
warn() { echo "[doctor][WARN] $*" >&2; }
err()  { echo "[doctor][ERROR] $*" >&2; }

DATA_DIR="/var/lib/dployrd"
INSTALL_LOG="/var/log/dployrd/install.log"
INSTALL_SCRIPT_URL="https://raw.githubusercontent.com/dployr-io/dployr/master/install.sh"
INSTALL_VERSION="${DPLOYR_VERSION:-latest}"

ensure_data_dir() {
  if mkdir -p "$DATA_DIR" 2>/dev/null; then
    log "data directory ready at $DATA_DIR"
  else
    err "unable to create data directory $DATA_DIR"
    return 1
  fi
}

check_vfox() {
  if command -v vfox >/dev/null 2>&1; then
    log "vfox found: $(command -v vfox)"
    vfox --version 2>&1 | sed 's/^/[doctor] vfox: /'
  else
    err "vfox not found in PATH"
    return 1
  fi
}

check_systemd_caddy() {
  if ! command -v systemctl >/dev/null 2>&1; then
    warn "systemctl not found; skipping caddy service check"
    return 0
  fi

  if ! systemctl list-unit-files caddy.service >/dev/null 2>&1; then
    warn "caddy.service not registered with systemd"
  fi

  status=$(systemctl is-active caddy 2>/dev/null || echo "unknown")
  log "caddy service status: $status"
  if [ "$status" != "active" ]; then
    warn "caddy service is not active"
  fi

  if command -v caddy >/dev/null 2>&1; then
    log "caddy binary found: $(command -v caddy)"
  else
    err "caddy binary not found in PATH"
    return 1
  fi
}

snapshot_resources() {
  if command -v free >/dev/null 2>&1; then
    echo "[doctor] ===== free -h ====="; free -h || true
  fi
  if command -v df >/dev/null 2>&1; then
    echo "[doctor] ===== df -h ====="; df -h || true
  fi
  if command -v lsblk >/dev/null 2>&1; then
    echo "[doctor] ===== lsblk ====="; lsblk || true
  fi
}

install_if_needed() {
	warn "attempting remote install to fix issues"
	if command -v curl >/dev/null 2>&1; then
		if [ -n "$INSTALL_VERSION" ]; then
			curl -sSL "$INSTALL_SCRIPT_URL" \
				| bash -s "$INSTALL_VERSION" >>"$INSTALL_LOG" 2>&1 \
				|| warn "remote installer reported errors; see $INSTALL_LOG"
		else
			curl -sSL "$INSTALL_SCRIPT_URL" \
				| bash >>"$INSTALL_LOG" 2>&1 \
				|| warn "remote installer reported errors; see $INSTALL_LOG"
		fi
	else
		warn "curl not available; cannot run remote installer"
	fi
}

main() {
  local failed=0

  ensure_data_dir || failed=1
  check_vfox || failed=1

  if [ "$(uname -s)" != "Darwin" ]; then
    check_systemd_caddy || failed=1
  fi

  snapshot_resources || true

  if [ $failed -ne 0 ]; then
    install_if_needed
    exit 1
  fi

  log "system doctor completed successfully"
}

main "$@"
