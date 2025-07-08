#!/usr/bin/env bash

set -e

echo "Launching dployr.io..."

# Check if Docker is installed and running
if ! command -v docker &> /dev/null; then
    echo "ERROR: Docker is not installed. Please run the install script first."
    exit 1
fi

if ! systemctl is-active --quiet docker; then
    echo "ERROR: Docker service is not running. Please start Docker first."
    exit 1
fi

# Check if the dployr directory exists
if [ ! -d "/data/dployr" ]; then
    echo "ERROR: /data/dployr directory does not exist. Please run the install script first."
    exit 1
fi

cd /data/dployr

# Check if docker-compose.yml exists
if [ ! -f "docker-compose.yml" ]; then
    echo "ERROR: docker-compose.yml not found in /data/dployr. Please run the install script first."
    exit 1
fi

echo "Starting dployr.io services..."
docker compose up -d --pull always

# Wait for health check with timeout
echo "Waiting for dployr.io to start..."
TIMEOUT=60
COUNTER=0

while [ $COUNTER -lt $TIMEOUT ]; do
    if curl -f http://localhost:7879/health &> /dev/null; then
        break
    fi
    echo "Waiting for dployr.io to start... ($COUNTER/$TIMEOUT seconds)"
    sleep 5
    COUNTER=$((COUNTER + 5))
done

if [ $COUNTER -ge $TIMEOUT ]; then
    echo "ERROR: dployr.io failed to start within $TIMEOUT seconds"
    echo "Checking container logs..."
    docker compose logs
    exit 1
fi

cat << 'EOF'
     _       _                        _       
  __| |_ __ | | ___  _   _ _ __      (_) ___  
 / _` | '_ \| |/ _ \| | | | '__|     | |/ _ \ 
| (_| | |_) | | (_) | |_| | |     _  | | (_) |
 \__,_| .__/|_|\___/ \__, |_|    (_) |_|\___/ 
      |_|            |___/                    
EOF

echo "Dashboard: http://$(curl -s ifconfig.me):7879"

echo "dployr.io started successfully!"
