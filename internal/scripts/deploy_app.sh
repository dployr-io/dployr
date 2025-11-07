#!/usr/bin/env bash
# deploy_app.sh — unified deployment script
# Handles runtime setup, build, and service installation in one go
# Usage: deploy_app.sh <action> <service_name> <runtime> <version> <workdir> <run_cmd> <description> <build_cmd> <port> [env_vars...]

set -euo pipefail

# --- arguments ---
ACTION="$1"
SERVICE_NAME="$2"
RUNTIME="${3:-}"
VERSION="${4:-}"
WORKDIR="${5:-}"
RUN_CMD="${6:-}"
DESCRIPTION="${7:-}"
BUILD_CMD="${8:-}"
PORT="${9:-3000}"
shift 9 2>/dev/null || true
# Store remaining args (env vars) in array
declare -a ENV_VARS_ARRAY
ENV_VARS_ARRAY=("$@")

# --- logging ---
log() { echo "[$(date -u +"%Y-%m-%dT%H:%M:%SZ")] $*"; }
abort() { log "ERROR: $*"; exit 1; }

# --- vfox setup ---
setup_vfox_env() {
    local vfox_bin="/usr/local/bin/vfox"
    
    if [ ! -f "$vfox_bin" ]; then
        abort "vfox not found at $vfox_bin"
    fi
    
    log "Using vfox at: $vfox_bin"
    
    # Activate vfox in current shell
    if ! eval "$($vfox_bin activate bash)" 2>/dev/null; then
        abort "failed to activate vfox environment"
    fi
}

# --- runtime installation ---
install_runtime() {
    local runtime="$1"
    local version="$2"
    
    log "Installing runtime: ${runtime}@${version}"
    
    if ! vfox install "${runtime}@${version}" -y 2>&1; then
        abort "failed to install ${runtime}@${version}"
    fi
    
    log "Switching to runtime: ${runtime}@${version}"
    if ! vfox use --global "${runtime}@${version}" 2>&1; then
        abort "failed to use ${runtime}@${version}"
    fi

    mkdir -p "${HOME}/.dployr/envs"
    ENV_FILE="${HOME}/.dployr/envs/${runtime}-${version}.env"

    log "Capturing environment for systemd"

    eval "$(/usr/local/bin/vfox activate bash)" || abort "failed to activate vfox in current shell"

    vfox use "${runtime}@${version}" || abort "failed to switch to ${runtime}@${version}"

    {
        echo "PATH=$PATH"
        echo "HOME=$HOME"
        echo "USER=$USER"
    } > "$ENV_FILE"

    log "Environment snapshot saved: $ENV_FILE"
}

# --- verify runtime ---
verify_runtime() {
    local runtime="$1"
    
    case "$runtime" in
        nodejs)
            if ! command -v node >/dev/null 2>&1; then
                abort "node executable not found after activation"
            fi
            log "Node version: $(node --version)"
            ;;
        python)
            if ! command -v python >/dev/null 2>&1 && ! command -v python3 >/dev/null 2>&1; then
                abort "python executable not found after activation"
            fi
            ;;
        golang)
            if ! command -v go >/dev/null 2>&1; then
                abort "go executable not found after activation"
            fi
            ;;
    esac
}

# --- build application ---
run_build() {
    local workdir="$1"
    local build_cmd="$2"
    
    cd "$workdir" || abort "cannot cd into workdir: $workdir"
    
    if [ -n "$build_cmd" ]; then
        log "Running build command: $build_cmd"
        if ! eval "$build_cmd" 2>&1; then
            abort "build command failed: $build_cmd"
        fi
        log "Build completed successfully"
    else
        log "No build command specified, skipping build"
    fi
}

# --- create .env file ---
create_env_file() {
    local workdir="$1"
    local port="$2"
    shift 2
    
    local env_file="${workdir}/.env"
    
    log "Creating .env file at: $env_file"
    log "Working directory: $workdir"
    log "Port: $port"
    log "Additional env vars count: $#"
    
    # Start with PORT
    echo "PORT=${port}" > "$env_file" || abort "Failed to write to $env_file"
    
    # Add any additional env vars passed as KEY=VALUE pairs
    for env_var in "$@"; do
        if [ -n "$env_var" ]; then
            log "Adding env var: $env_var"
            echo "$env_var" >> "$env_file"
        fi
    done
    
    log "Environment variables written to .env file"
    
    # TODO [DEBUG] - remove this
    log "Contents of .env file:"
    cat "$env_file"
}

# --- create service exe script ---
create_service_exe() {
    local service_name="$1"
    local runtime="$2"
    local version="$3"
    local workdir="$4"
    local run_cmd="$5"

    local exe_script="${HOME}/.dployr/scripts/${service_name}.sh"
    mkdir -p "$(dirname "$exe_script")"

    local env_file="${HOME}/.dployr/envs/${runtime}-${version}.env"

    cat > "$exe_script" <<EOF
#!/usr/bin/env bash
set -euo pipefail

# Source the captured vfox environment
if [ -f "$env_file" ]; then
    set -a
    source "$env_file"
    set +a
else
    echo "[ERROR] ENV_FILE does not exist: $env_file" >&2
    exit 1
fi

cd "${workdir}" || exit 1

exec ${run_cmd}
EOF

    chmod +x "$exe_script"
    log "Created service exe: $exe_script"
    echo "$exe_script"
}

# --- install systemd service ---
install_systemd_service() {
    local service_name="$1"
    local description="$2"
    local exe_script="$3"
    local workdir="$4"
    local runtime="$5"
    local version="$6"

    log "Installing systemd service: $service_name"

    local log_dir="${HOME}/.dployr/logs"
    mkdir -p "$log_dir"
    sudo mkdir -p "/etc/systemd/system"
    local log_file="${log_dir}/${service_name}.log"

    sudo tee "/etc/systemd/system/${service_name}.service" > /dev/null <<EOF
[Unit]
Description=${description}
After=network.target

[Service]
Type=simple
User=dployrd
WorkingDirectory=${workdir}
ExecStart=${exe_script}
StandardOutput=append:${log_file}
StandardError=append:${log_file}

Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
    sudo systemctl enable "$service_name"
    log "Service $service_name installed and enabled"
}


# --- start systemd service ---
start_systemd_service() {
    local service_name="$1"
    
    log "Starting service: $service_name"
    if ! sudo systemctl start "$service_name"; then
        abort "failed to start service: $service_name"
    fi
    
    sleep 2
    
    if sudo systemctl is-active --quiet "$service_name"; then
        log "Service $service_name started successfully"
    else
        abort "service $service_name failed to start"
    fi
}

# --- stop systemd service ---
stop_systemd_service() {
    local service_name="$1"
    
    log "Stopping service: $service_name"
    sudo systemctl stop "$service_name" || true
    log "Service $service_name stopped"
}

# --- remove systemd service ---
remove_systemd_service() {
    local service_name="$1"
    
    log "Removing service: $service_name"
    sudo systemctl stop "$service_name" 2>/dev/null || true
    sudo systemctl disable "$service_name" 2>/dev/null || true
    sudo rm -f "/etc/systemd/system/${service_name}.service"
    rm -f "${HOME}/.dployr/scripts/${service_name}.sh"
    sudo systemctl daemon-reload
    log "Service $service_name removed"
}

# --- get service status ---
get_service_status() {
    local service_name="$1"
    
    if [ ! -f "/etc/systemd/system/${service_name}.service" ]; then
        echo "stopped"
        exit 1
    fi
    
    if sudo systemctl is-active --quiet "$service_name" 2>/dev/null; then
        echo "running"
    else
        echo "stopped"
    fi
}

# --- main deployment flow ---
deploy() {
    log "Starting deployment for: $SERVICE_NAME"
    
    # Validate inputs
    [ -z "$RUNTIME" ] && abort "runtime name required"
    [ -z "$VERSION" ] && abort "runtime version required"
    [ -z "$WORKDIR" ] && abort "working directory required"
    [ -z "$RUN_CMD" ] && abort "run command required"
    [ -z "$DESCRIPTION" ] && abort "service description required"
    
    # Setup vfox environment
    log "Setting up vfox environment"
    setup_vfox_env
    
    # Install and activate runtime
    install_runtime "$RUNTIME" "$VERSION"
    
    # Verify runtime is available
    verify_runtime "$RUNTIME"
    
    # Create .env file with PORT and environment variables
    if [ ${#ENV_VARS_ARRAY[@]} -gt 0 ]; then
        create_env_file "$WORKDIR" "$PORT" "${ENV_VARS_ARRAY[@]}"
    else
        create_env_file "$WORKDIR" "$PORT"
    fi
    
    # Run build if specified
    run_build "$WORKDIR" "$BUILD_CMD"
    
    # Create service exe script
    exe_script=$(create_service_exe "$SERVICE_NAME" "$RUNTIME" "$VERSION" "$WORKDIR" "$RUN_CMD" | tail -n1)
    
    # Install systemd service
    install_systemd_service "$SERVICE_NAME" "$DESCRIPTION" "$exe_script" "$WORKDIR" "$RUNTIME" "$VERSION"
    
    # Start the service
    start_systemd_service "$SERVICE_NAME"
    
    log "Deployment completed successfully for: $SERVICE_NAME"
}

# --- action dispatcher ---
case "$ACTION" in
    deploy)
        deploy
        ;;
    start)
        [ -z "$SERVICE_NAME" ] && abort "service name required"
        start_systemd_service "$SERVICE_NAME"
        ;;
    stop)
        [ -z "$SERVICE_NAME" ] && abort "service name required"
        stop_systemd_service "$SERVICE_NAME"
        ;;
    remove)
        [ -z "$SERVICE_NAME" ] && abort "service name required"
        remove_systemd_service "$SERVICE_NAME"
        ;;
    status)
        [ -z "$SERVICE_NAME" ] && abort "service name required"
        get_service_status "$SERVICE_NAME"
        ;;
    *)
        echo "Usage: $0 {deploy|start|stop|remove|status} <service_name> [runtime] [version] [workdir] [run_cmd] [description] [build_cmd]"
        echo ""
        echo "Actions:"
        echo "  deploy  - Full deployment: setup runtime, build, install and start service"
        echo "  start   - Start an existing service"
        echo "  stop    - Stop a running service"
        echo "  remove  - Remove a service completely"
        echo "  status  - Check service status"
        exit 1
        ;;
esac

exit 0
