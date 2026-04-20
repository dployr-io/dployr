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

        # Build docker create command
        create_cmd=(docker create --name "$name" --restart unless-stopped)

        # Port mapping if port is set and non-zero
        if [ -n "$port" ] && [ "$port" -ne 0 ] 2>/dev/null; then
            create_cmd+=(-p "${port}:${port}")
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

        echo "Creating Docker container: $name"
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
