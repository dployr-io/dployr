#!/usr/bin/env bash

# Publish image to local registry

./scripts/docker/build.sh

# Load environment variables from .env file
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="$SCRIPT_DIR/../../.env.dev"
if [ -f "$ENV_FILE" ]; then
  export $(grep -v '^#' "$ENV_FILE" | xargs)
fi

: ${REGISTRY_URL:=$HOST:$PORT}
: ${IMAGE:=$REGISTRY_URL/$NAME:$TAG}

if [ -z "$REGISTRY_URL" ]; then
  echo "REGISTRY_URL is not set. Please set it in your environment variables."
  echo "Example: export REGISTRY_URL=ghcr.io/yourusername"
  echo ""
  echo "If you're building locally, you can run ./scripts/dev/local-registry.sh"
  echo "first to start a local Docker registry."
  exit 1
fi

if [ -z "$IMAGE" ]; then
    echo "No image found. Please build the image first using ./scripts/docker/build.sh"
    exit 1
fi

docker push $IMAGE

echo "Image $IMAGE published successfully."