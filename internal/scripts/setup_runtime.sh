#!/usr/bin/env bash
# setup_runtime.sh — sets up and uses a specific runtime using vfox
# Usage: setup_runtime.sh <runtime> <version> <workdir> <build_command>

set -euo pipefail

runtime="$1"
version="$2"
workdir="$3"
shift 3 || true
build_cmd="$*"

log() { echo "[$(date -u +"%Y-%m-%dT%H:%M:%SZ")] $*"; }

abort() {
    log "ERROR: $*"
    exit 1
}

# --- validation ---
[ -z "$runtime" ] && abort "runtime name required"
[ -z "$version" ] && abort "runtime version required"
[ -z "$workdir" ] && abort "working directory required"

# --- install runtime ---
log "Installing runtime: ${runtime}@${version}"
if ! vfox install "${runtime}@${version}" -y >/dev/null 2>&1; then
    abort "failed to install ${runtime}@${version}"
fi

# --- activate vfox ---
if ! eval "$(vfox activate bash)" >/dev/null 2>&1; then
    abort "failed to activate vfox environment"
fi

# --- use runtime ---
log "Switching to runtime: ${runtime}@${version}"
if ! vfox use --global "${runtime}@${version}" >/dev/null 2>&1; then
    abort "failed to use ${runtime}@${version}"
fi

# --- verify runtime ---
if ! command -v node >/dev/null 2>&1 && [ "$runtime" = "nodejs" ]; then
    abort "node executable not found after activation"
fi

# --- run build command ---
cd "$workdir" || abort "cannot cd into workdir: $workdir"

if [ -n "$build_cmd" ]; then
    log "Running build command: $build_cmd"
    if ! eval "$build_cmd"; then
        abort "build command failed: $build_cmd"
    fi
else
    log "No build command specified"
fi

log "Runtime ${runtime}@${version} setup complete"
exit 0