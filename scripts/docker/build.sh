#!/usr/bin/env bash

set -e

# Load environment variables from .env file
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="$SCRIPT_DIR/../../.env.dev"
if [ -f "$ENV_FILE" ]; then
  export $(grep -v '^#' "$ENV_FILE" | xargs)
fi

: ${REGISTRY_URL:=$HOST:$PORT}
: ${NAME:=dployr}
: ${TAG:=latest}

if [ -z "$REGISTRY_URL" ]; then
  echo "REGISTRY_URL is not set. Please set it in your environment variables."
  echo "Example: export REGISTRY_URL=ghcr.io/yourusername"
  echo ""
  echo "If you're building locally, you can run ./scripts/dev/local-registry.sh"
  echo "first to start a local Docker registry."
  exit 1
fi

# Change to server directory for build context to avoid problematic files
cd server
docker build -t $REGISTRY_URL/$NAME:$TAG . --file Dockerfile

echo "Docker image built successfully: $REGISTRY_URL/$NAME:$TAG"
