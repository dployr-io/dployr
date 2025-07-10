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

BASE=$(cat "$SCRIPT_DIR/../../version.txt")
DEVBUMP_FILE="$SCRIPT_DIR/../../.devbump"

if [ -n "$CI" ]; then
  # CI build: use nearest Git tag (must be pushed already)
  TAG="$(git describe --tags --abbrev=0)"
else
  # Local build: bump .devbump
  B0=$(cat "$DEVBUMP_FILE")
  B=$((B0 + 1))
  echo "$B" > "$DEVBUMP_FILE"
  TAG="${BASE}+dev.${B}"
fi

echo "→ Building version $TAG"

if [ -z "$REGISTRY_URL" ]; then
  echo "REGISTRY_URL is not set. Please set it in your environment variables."
  echo "Example: export REGISTRY_URL=ghcr.io/yourusername"
  echo ""
  echo "If you're building locally, you can run ./scripts/dev/local-registry.sh"
  echo "first to start a local Docker registry."
  exit 1
fi

cd server

# Pass VERSION into Go via ldflags
LDFLAGS="-X 'main.Version=$TAG'"

docker build \
  --build-arg LDFLAGS="$LDFLAGS" \
  -t "$REGISTRY_URL/$NAME:$TAG" \
  . --file Dockerfile

echo "Docker image built successfully: $REGISTRY_URL/$NAME:$TAG"
