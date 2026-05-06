#!/usr/bin/env bash

# Copyright 2025 Emmanuel Madehin
# SPDX-License-Identifier: Apache-2.0

# docker.sh — manages Docker containers for service lifecycle
# Usage: docker.sh <action> <name> [image] [workdir] [run_cmd] [port] [description]

set -e

action="$1"
name="$2"
image="${3:-}"
workdir="${4:-}"
run_cmd="${5:-}"
port="${6:-3000}"
description="${7:-}"

# Function to check if a port is in use
is_port_available() {
    local check_port="$1"
    ! timeout 1 bash -c "echo > /dev/tcp/127.0.0.1/${check_port}" 2>/dev/null
}

# Function to find an available host port starting from container name hash (61000-64999 range)
get_host_port() {
    local container_name="$1"
    local hash
    local hash_dec
    local port_range
    local host_port

    hash=$(echo -n "$container_name" | md5sum | cut -c1-8)
    hash_dec=$((16#$hash))
    port_range=$((64999 - 61000 + 1))
    host_port=$((hash_dec % port_range + 61000))

    local attempts=0

    # Find an available port starting from calculated base port
    while [ $attempts -lt $port_range ]; do
        if is_port_available "$host_port"; then
            echo "$host_port"
            return 0
        fi

        # Try next port in range, wrapping around
        host_port=$(( host_port + 1 ))
        if [ $host_port -gt 64999 ]; then
            host_port=61000
        fi

        attempts=$((attempts + 1))
    done

    # Fallback: return original base port if all are in use
    hash=$(echo -n "$container_name" | md5sum | cut -c1-8)
    hash_dec=$((16#$hash))
    echo $((hash_dec % port_range + 61000))
}

case "$action" in
    install)
        [ -z "$image" ] && { echo "ERROR: image name required"; exit 1; }
        [ -z "$workdir" ] && { echo "ERROR: workdir required"; exit 1; }
        [ -z "$run_cmd" ] && { echo "ERROR: run_cmd required"; exit 1; }

        # Check Docker availability
        if ! command -v docker &> /dev/null; then
            echo "ERROR: docker command not found. Please install Docker."
            exit 1
        fi

        # Pull image if not present
        if ! docker image inspect "$image" > /dev/null 2>&1; then
            echo "Pulling Docker image: $image"
            docker pull "$image"
        fi

        # Prepare env file path
        env_file="${workdir}/.env"

        # Handle PORT environment variable - remove any existing PORT and set correct one
        if [ -f "$env_file" ]; then
            grep -v '^PORT=' "$env_file" > "${env_file}.tmp" || true
            mv "${env_file}.tmp" "$env_file"
        fi

        echo "PORT=$port" >> "$env_file"

        # Build docker create command
        create_cmd=(docker create --name "$name" --restart unless-stopped)

        # Port mapping - use host port in 61000-64999 range, container port is the specified port
        if [ -n "$port" ] && [ "$port" -ne 0 ] 2>/dev/null; then
            host_port=$(get_host_port "$name")
            create_cmd+=(-p "${host_port}:${port}")
        fi

        # Env file if exists
        if [ -f "$env_file" ]; then
            create_cmd+=(--env-file "$env_file")
        fi

        # Label with description if provided
        if [ -n "$description" ]; then
            create_cmd+=(--label "description=$description")
        fi

        # Append image and command
        if [ -n "$run_cmd" ]; then
            create_cmd+=("$image" bash -c "$run_cmd")
        else
            create_cmd+=("$image")
        fi

        host_port=$(get_host_port "$name")
        echo "Creating Docker container: $name (host port: $host_port, container port: $port)"
        "${create_cmd[@]}"

        echo "Container $name created successfully"
        ;;

    start)
        echo "Starting Docker container: $name"
        docker start "$name"
        ;;

    stop)
        echo "Stopping Docker container: $name"
        docker stop "$name"
        ;;

    remove)
        echo "Removing Docker container: $name"
        docker rm -f "$name" 2>/dev/null || true
        ;;

    status)
        if ! docker inspect "$name" > /dev/null 2>&1; then
            echo "stopped"
            exit 1
        fi
        state=$(docker inspect -f '{{.State.Status}}' "$name")
        if [ "$state" = "running" ]; then
            echo "running"
        else
            echo "stopped"
        fi
        ;;

    *)
        echo "Usage: $0 {install|start|stop|remove|status} <name> [image] [workdir] [run_cmd] [port] [description]"
        exit 1
        ;;
esac

exit 0
