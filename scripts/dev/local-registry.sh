#!/usr/bin/env bash
set -e

# Load environment variables from .env file
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ENV_FILE="$SCRIPT_DIR/../../.env.dev"
if [ -f "$ENV_FILE" ]; then
  export $(grep -v '^#' "$ENV_FILE" | xargs)
fi

# Set default values if not set in .env
: ${PORT:=4444}
: ${HOST:=192.168.100.4}
: ${IMAGE:=registry}

# Check for port conflict
if netstat -tln 2>/dev/null \
   | awk '{print $4}' \
   | grep -Eq ":${PORT}\$"; then

  echo "Port $PORT is in use. Checking if it's our container…"

  # 2) If our container is already running, we’re done
  if docker ps \
       --filter "name=^/${CT_NAME}$" \
       --filter "status=running" \
       -q | grep -q .; then
    echo "Container '$CT_NAME' already bound to port $PORT."
    exit 0
  fi

  # 3) Otherwise it must be someone else on that port → error
  echo "Port $PORT is occupied by another process; aborting." >&2
  exit 1
fi


# If container exists but is stopped, start it
if docker ps -a --filter "name=$HOST" -q | grep -q .; then
  echo "Starting existing container $HOST..."
  docker start $HOST
  exit 0
fi

# Otherwise, create & run it
echo "Creating & launching $HOST on port $PORT..."
docker run -d \
  --name $HOST \
  -p $PORT:5000 \
  --restart unless-stopped \
  $IMAGE

echo "Waiting for container '$HOST' to start…"
until [ "$(docker inspect -f '{{.State.Running}}' $HOST 2>/dev/null)" = "true" ]; do
  sleep 0.5
done

echo "Registry is now running at $HOST:$PORT"