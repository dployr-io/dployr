#!/bin/bash

echo "Cleaning up all VMs..."

# Destroy all vagrant VMs
echo "Destroying Vagrant VMs..."
vagrant global-status --prune
vagrant destroy -f

# Remove any remaining VirtualBox VMs with "dployr" in name
echo "Removing VirtualBox VMs..."
if command -v VBoxManage >/dev/null 2>&1; then
    VBoxManage list vms | grep "dployr" | cut -d'"' -f2 | while read vm; do
        echo "Removing VM: $vm"
        VBoxManage unregistervm "$vm" --delete 2>/dev/null || true
    done
else
    echo "VBoxManage not found in PATH - open VirtualBox GUI and manually delete dployr VMs"
fi

# Clean up VirtualBox VM directories
echo "Cleaning VM directories..."
if [ -d "$HOME/VirtualBox VMs" ]; then
    find "$HOME/VirtualBox VMs" -name "*dployr*" -type d -exec rm -rf {} + 2>/dev/null || true
fi

# Windows path cleanup
if [ -d "/c/Users/$USER/VirtualBox VMs" ]; then
    find "/c/Users/$USER/VirtualBox VMs" -name "*dployr*" -type d -exec rm -rf {} + 2>/dev/null || true
fi

# Vagrant cleanup
echo "Cleaning Vagrant cache..."
vagrant box prune -f
rm -rf .vagrant 2>/dev/null || true

# Remove dployr.io/dployr:latest docker images and containers
echo "Removing dployr.io/dployr:latest docker images and containers..."
containers=$(docker ps -aq -f "ancestor=dployr.io/dployr:latest") && [ -n "$containers" ] && docker rm -f $containers 2>/dev/null || true
docker rmi -f $(docker images -q dployr.io/dployr:latest) 2>/dev/null || true

echo "Removing all containers created from dployr.io/dployr:latest..."
docker ps -aq --filter "ancestor=dployr.io/dployr:latest" | xargs -r docker rm -f 2>/dev/null || true
