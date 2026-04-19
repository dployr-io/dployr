#!/usr/bin/env bash

# Copyright 2025 Emmanuel Madehin
# SPDX-License-Identifier: Apache-2.0

# deploy_app.sh — unified deployment script
# Handles runtime setup, build, and service installation in one go
# Usage: deploy_app.sh <action> <service_name> <source> <type> <runtime> <version> <workdir> <run_cmd> <description> <build_cmd> <port> [image] [static_dir]
# Environment variables are read from config.toml in the workdir

set -euo pipefail

# --- arguments ---
ACTION="$1"
SERVICE_NAME="$2"
SOURCE="${3:-}"
TYPE="${4:-}"
RUNTIME="${5:-}"
VERSION="${6:-}"
WORKDIR="${7:-}"
RUN_CMD="${8:-}"
DESCRIPTION="${9:-}"
BUILD_CMD="${10:-}"
PORT="${11:-3000}"
IMAGE="${12:-}"
STATIC_DIR="${13:-}"

# --- logging ---
log() { echo "[INFO] $*"; }
warn() { echo "[WARN] $*"; }
abort() { echo "[ERROR] $*" >&2; exit 1; }

# --- detect runtime backend ---
detect_backend() {
    if [ "$TYPE" = "job" ]; then
        echo "systemd"
    else
        echo "docker"
    fi
}

BACKEND=$(detect_backend)
log "Detected backend: $BACKEND (source=$SOURCE, type=$TYPE)"

# ==== SYSTEMD BACKEND ====

setup_vfox_env() {
    local vfox_bin="/usr/local/bin/vfox"
    
    if [ ! -f "$vfox_bin" ]; then
        abort "vfox not found at $vfox_bin"
    fi
    
    log "Using vfox at: $vfox_bin"
    
    if ! eval "$($vfox_bin activate bash)" 2>/dev/null; then
        abort "failed to activate vfox environment"
    fi
}

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

create_env_file() {
    local workdir="$1"
    local port="$2"
    
    local env_file="${workdir}/.env"
    local config_file="${workdir}/config.toml"
    
    log "Creating .env file at: $env_file"
    log "Working directory: $workdir"
    log "Port: $port"
    
    declare -A written_keys
    
    echo "PORT=${port}" > "$env_file" || abort "Failed to write to $env_file"
    written_keys["PORT"]=1
    
    if [ -f "$config_file" ]; then
        log "Reading environment configuration from: $config_file"
        
        local current_section=""
        while IFS= read -r line || [ -n "$line" ]; do
            [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
            
            if [[ "$line" =~ ^\[([a-zA-Z_]+)\]$ ]]; then
                current_section="${BASH_REMATCH[1]}"
                log "Processing section: [$current_section]"
                continue
            fi
            
            if [[ "$line" =~ ^[[:space:]]*([a-zA-Z_][a-zA-Z0-9_]*)[[:space:]]*=[[:space:]]*\"(.*)\"[[:space:]]*$ ]]; then
                local key="${BASH_REMATCH[1]}"
                local value="${BASH_REMATCH[2]}"
                
                if [ -n "${written_keys[$key]:-}" ]; then
                    log "Skipping duplicate key: $key (already set)"
                    continue
                fi
                
                echo "${key}=${value}" >> "$env_file"
                written_keys[$key]=1
                
                if [ "$current_section" = "secrets" ]; then
                    log "Adding secret: $key=***"
                else
                    log "Adding env var: $key=$value"
                fi
            fi
        done < "$config_file"
    else
        log "No config.toml found at $config_file, using PORT only"
    fi
    
    log "Environment variables written to .env file"
}

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

systemd_install() {
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

systemd_start() {
    local service_name="$1"
    local log_file="${HOME}/.dployr/logs/${service_name}.log"
    
    log "Starting service: $service_name"
    
    local log_size_before=0
    if [ -f "$log_file" ]; then
        log_size_before=$(stat -c%s "$log_file" 2>/dev/null || echo 0)
    fi
    
    if ! sudo systemctl start "$service_name"; then
        systemd_capture_logs "$service_name" "$log_file" "$log_size_before"
        abort "failed to start service: $service_name"
    fi
    
    sleep 2
    
    if sudo systemctl is-active --quiet "$service_name"; then
        log "Service $service_name started successfully"
    else
        systemd_capture_logs "$service_name" "$log_file" "$log_size_before"
        abort "service $service_name failed to start"
    fi
}

systemd_capture_logs() {
    local service_name="$1"
    local log_file="$2"
    local log_size_before="$3"
    
    log "Service startup failed. Capturing runtime logs..."
    
    if sudo systemctl status "$service_name" --no-pager --lines=5 2>/dev/null | grep -q "Active: failed"; then
        log "Service status: failed"
    fi
    
    if [ -f "$log_file" ]; then
        local log_size_after
        log_size_after=$(stat -c%s "$log_file" 2>/dev/null || echo 0)
        
        if [ "$log_size_after" -gt "$log_size_before" ]; then
            log "--- Service Runtime Logs (last 50 lines) ---"
            tail -n 50 "$log_file" 2>/dev/null || true
            log "--- End of Runtime Logs ---"
        else
            log "No new runtime logs found in $log_file"
        fi
    else
        log "Service log file not found: $log_file"
    fi
}

systemd_stop() {
    local service_name="$1"
    
    log "Stopping service: $service_name"
    sudo systemctl stop "$service_name" 2>/dev/null || true
    log "Service $service_name stopped"
}

systemd_remove() {
    local service_name="$1"
    
    log "Removing service: $service_name"
    sudo systemctl stop "$service_name" 2>/dev/null || true
    sudo systemctl disable "$service_name" 2>/dev/null || true
    sudo rm -f "/etc/systemd/system/${service_name}.service"
    rm -f "${HOME}/.dployr/scripts/${service_name}.sh"
    sudo systemctl daemon-reload
    log "Service $service_name removed"
}

systemd_status() {
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

# ==== DOCKER BACKEND ====

runtime_to_image() {
    local runtime="$1"
    local version="$2"
    
    case "$runtime" in
        golang)
            echo "golang:${version}"
            ;;
        php)
            echo "php:${version}"
            ;;
        python)
            echo "python:${version}"
            ;;
        nodejs)
            echo "node:${version}"
            ;;
        ruby)
            echo "ruby:${version}"
            ;;
        dotnet)
            echo "mcr.microsoft.com/dotnet:${version}"
            ;;
        java)
            echo "eclipse-temurin:${version}"
            ;;
        *)
            echo "${runtime}:${version}"
            ;;
    esac
}

ensure_docker() {
    if ! command -v docker &> /dev/null; then
        log "Docker not found. Installing..."
        if [ "$(id -u)" -ne 0 ]; then
            abort "Docker installation requires root privileges"
        fi
        
        curl -fsSL https://get.docker.com -o /tmp/get-docker.sh
        sh /tmp/get-docker.sh
        rm -f /tmp/get-docker.sh
        
        if ! command -v docker &> /dev/null; then
            abort "Docker installation failed"
        fi
        
        log "Docker installed successfully"
    fi
    
    if ! docker info &> /dev/null; then
        abort "Docker daemon not running. Please start docker and try again."
    fi
    
    log "Docker is available and running"
}

docker_build_image() {
    local workdir="$1"
    local service_name="$2"
    local runtime="$3"
    local version="$4"
    local port="$5"
    
    local image_ref
    image_ref=$(runtime_to_image "$runtime" "$version")
    
    local image_name="dployr/${service_name}:latest"
    
    log "Building Docker image: $image_name"
    
    local dockerfile="${workdir}/Dockerfile"
    if [ ! -f "$dockerfile" ]; then
        cat > "$dockerfile" <<EOF
FROM ${image_ref}

WORKDIR /app

COPY . .

ENV PORT=${port}

CMD ["${RUN_CMD:-}"]
EOF
        log "Created Dockerfile from template"
    fi
    
    cd "$workdir" || abort "cannot cd into workdir: $workdir"
    
    if ! docker build -t "$image_name" .; then
        abort "Docker build failed"
    fi
    
    log "Docker image built: $image_name"
    echo "$image_name"
}

docker_pull_image() {
    local image="$1"
    
    if [ -z "$image" ]; then
        abort "Image name required for pull"
    fi
    
    if ! docker image inspect "$image" > /dev/null 2>&1; then
        log "Pulling Docker image: $image"
        if ! docker pull "$image"; then
            abort "Docker pull failed for $image"
        fi
        log "Image pulled: $image"
    else
        log "Image already present: $image"
    fi
    
    echo "$image"
}

create_caddyfile() {
    local workdir="$1"
    local port="$2"
    local static_dir="$3"
    
    local serve_root="${static_dir:-.}"
    local serve_path="${workdir}/${serve_root}"
    
    if [ ! -f "${serve_path}/index.html" ] && [ ! -f "${serve_path}/index.htm" ]; then
        abort "Static deployment requires index.html or index.htm in ${serve_path}"
    fi
    
    local caddyfile="${workdir}/Caddyfile"
    
    cat > "$caddyfile" <<EOF
:${port} {
    root * ${serve_root}
    file_server
    encode gzip
    
    handle_errors {
        respond "{err.status_text} {err.status_code}"
    }
}
EOF
    
    log "Created Caddyfile at: $caddyfile"
}

docker_create_container() {
    local name="$1"
    local image="$2"
    local workdir="$3"
    local port="$4"
    local run_cmd="$5"
    local description="$6"
    local type="$7"
    local static_dir="$8"
    
    log "Creating Docker container: $name (type: $type)"
    
    docker rm -f "$name" 2>/dev/null || true
    
    local create_cmd=(docker run -d --name "$name" --restart unless-stopped)
    
    if [ -n "$port" ] && [ "$port" -ne 0 ] 2>/dev/null; then
        create_cmd+=(-p "${port}:${port}")
    fi
    
    local env_file="${workdir}/.env"
    if [ -f "$env_file" ]; then
        create_cmd+=(--env-file "$env_file")
    fi
    
    if [ -n "$description" ]; then
        create_cmd+=(--label "description=$description")
    fi
    
    if [ "$type" = "static" ]; then
        local serve_dir="${static_dir:-${workdir}}"
        if [ -d "$serve_dir" ]; then
            create_cmd+=(-v "${serve_dir}:/srv")
            create_cmd+=(-v "${workdir}/Caddyfile:/etc/caddy/Caddyfile:ro")
            create_cmd+=("caddy:2-alpine" "caddy" "run" "--config" "/etc/caddy/Caddyfile")
        else
            warn "Static directory not found: $serve_dir, using workdir"
            create_cmd+=(-v "${workdir}:/srv")
            create_cmd+=("caddy:2-alpine" "caddy" "file-server" "--root-dir" "/srv")
        fi
    elif [ -n "$run_cmd" ]; then
        if [ -d "$workdir" ]; then
            create_cmd+=(-v "${workdir}:/app")
            create_cmd+=(-w /app)
        fi
        create_cmd+=("$image" bash -c "$run_cmd")
    else
        if [ -d "$workdir" ]; then
            create_cmd+=(-v "${workdir}:/app")
            create_cmd+=(-w /app)
        fi
        create_cmd+=("$image")
    fi
    
    "${create_cmd[@]}" || abort "Failed to create container"
    
    log "Container $name created successfully"
}

docker_start() {
    local name="$1"
    
    log "Starting Docker container: $name"
    docker start "$name" || abort "Failed to start container"
    log "Container $name started"
}

docker_stop() {
    local name="$1"
    
    log "Stopping Docker container: $name"
    docker stop "$name" 2>/dev/null || true
    log "Container $name stopped"
}

docker_remove() {
    local name="$1"
    
    log "Removing Docker container: $name"
    docker rm -f "$name" 2>/dev/null || true
    log "Container $name removed"
}

docker_status() {
    local name="$1"
    
    if ! docker inspect "$name" > /dev/null 2>&1; then
        echo "stopped"
        exit 1
    fi
    
    local state
    state=$(docker inspect -f '{{.State.Status}}' "$name")
    if [ "$state" = "running" ]; then
        echo "running"
    else
        echo "stopped"
    fi
}

# ==== DEPLOYMENT ACTION ====

deploy_systemd() {
    log "Deploying with systemd backend"
    
    [ -z "$RUNTIME" ] && abort "runtime name required"
    [ -z "$VERSION" ] && abort "runtime version required"
    [ -z "$WORKDIR" ] && abort "working directory required"
    [ -z "$RUN_CMD" ] && abort "run command required"
    [ -z "$DESCRIPTION" ] && abort "service description required"
    
    log "Setting up vfox environment"
    setup_vfox_env
    
    log "Installing runtime: $RUNTIME@$VERSION"
    install_runtime "$RUNTIME" "$VERSION"
    
    verify_runtime "$RUNTIME"
    
    create_env_file "$WORKDIR" "$PORT"
    
    run_build "$WORKDIR" "$BUILD_CMD"
    
    exe_script=$(create_service_exe "$SERVICE_NAME" "$RUNTIME" "$VERSION" "$WORKDIR" "$RUN_CMD")
    
    systemd_install "$SERVICE_NAME" "$DESCRIPTION" "$exe_script" "$WORKDIR" "$RUNTIME" "$VERSION"
    
    systemd_start "$SERVICE_NAME"
    
    log "Deployment completed for: $SERVICE_NAME"
}

deploy_docker() {
    log "Deploying with docker backend"
    
    [ -z "$WORKDIR" ] && abort "working directory required"
    [ -z "$SERVICE_NAME" ] && abort "service name required"
    
    ensure_docker
    
    create_env_file "$WORKDIR" "$PORT"
    
    if [ "$TYPE" = "static" ]; then
        log "Setting up static deployment with Caddy"
        create_caddyfile "$WORKDIR" "$PORT" "$STATIC_DIR"
    fi
    
    local image=""
    if [ "$SOURCE" = "image" ]; then
        log "Using provided image"
        image=$(docker_pull_image "$IMAGE")
    else
        log "Building image from workdir"
        run_build "$WORKDIR" "$BUILD_CMD"
        image=$(docker_build_image "$WORKDIR" "$SERVICE_NAME" "$RUNTIME" "$VERSION" "$PORT")
    fi
    
    docker_create_container "$SERVICE_NAME" "$image" "$WORKDIR" "$PORT" "$RUN_CMD" "$DESCRIPTION" "$TYPE" "$STATIC_DIR"
    
    docker_start "$SERVICE_NAME"
    
    log "Deployment completed for: $SERVICE_NAME"
}

deploy() {
    case "$BACKEND" in
        systemd)
            deploy_systemd
            ;;
        docker)
            deploy_docker
            ;;
        *)
            abort "Unknown backend: $BACKEND"
            ;;
    esac
}

start() {
    [ -z "$SERVICE_NAME" ] && abort "service name required"
    
    case "$BACKEND" in
        systemd)
            systemd_start "$SERVICE_NAME"
            ;;
        docker)
            docker_start "$SERVICE_NAME"
            ;;
    esac
}

stop() {
    [ -z "$SERVICE_NAME" ] && abort "service name required"
    
    case "$BACKEND" in
        systemd)
            systemd_stop "$SERVICE_NAME"
            ;;
        docker)
            docker_stop "$SERVICE_NAME"
            ;;
    esac
}

remove() {
    [ -z "$SERVICE_NAME" ] && abort "service name required"
    
    case "$BACKEND" in
        systemd)
            systemd_remove "$SERVICE_NAME"
            ;;
        docker)
            docker_remove "$SERVICE_NAME"
            ;;
    esac
}

status() {
    [ -z "$SERVICE_NAME" ] && abort "service name required"
    
    case "$BACKEND" in
        systemd)
            systemd_status "$SERVICE_NAME"
            ;;
        docker)
            docker_status "$SERVICE_NAME"
            ;;
    esac
}

# ==== ACTION DISPATCHER ====

case "$ACTION" in
    deploy)
        deploy
        ;;
    start)
        start
        ;;
    stop)
        stop
        ;;
    remove)
        remove
        ;;
    status)
        status
        ;;
    *)
        echo "Usage: $0 {deploy|start|stop|remove|status} <service_name> [source] [type] [runtime] [version] [workdir] [run_cmd] [description] [build_cmd] [port] [image] [static_dir]"
        echo ""
        echo "Actions:"
        echo "  deploy  - Full deployment: setup runtime, build, install and start service"
        echo "  start   - Start an existing service"
        echo "  stop    - Stop a running service"
        echo "  remove  - Remove a service completely"
        echo "  status  - Check service status"
        echo ""
        echo "Backend selection (automatic):"
        echo "  type=job → systemd"
        echo "  other types → docker (including static)"
        exit 1
        ;;
esac

exit 0