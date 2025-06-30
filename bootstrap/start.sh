#!/bin/bash

echo "Launching dployr.io..."

cd /data/dployr
# Use modern "docker compose" (space, not hyphen)
docker compose up -d

# Wait for health check
until curl -f http://localhost:7879/health; do
    echo "Waiting for dployr.io to start..."
    sleep 5
done

cat << 'EOF'
     _       _                        _       
  __| |_ __ | | ___  _   _ _ __      (_) ___  
 / _` | '_ \| |/ _ \| | | | '__|     | |/ _ \ 
| (_| | |_) | | (_) | |_| | |     _  | | (_) |
 \__,_| .__/|_|\___/ \__, |_|    (_) |_|\___/ 
      |_|            |___/                    
EOF

echo "Dashboard: http://$(curl -s ifconfig.me):7879"

echo "dployr.io installed successfully!"
