#!/bin/bash

set -e

# Start timer for bootstrap process
START_TIME=$(date +%s)
echo "Bootstrap started at: $(date)"

. ./scripts/vm/clean-up.sh

# Get list of available VMs from Vagrantfile
AVAILABLE_VMS=($(grep -E '^  "[a-z]+[0-9]+"' Vagrantfile | cut -d'"' -f2))

# Prompt user to select VMs
echo "Available VMs: ${AVAILABLE_VMS[*]}"
read -p "Enter VM names to bootstrap (space separated, or 'all' for all VMs): " selected_vms

# Handle 'all' case
if [[ "$selected_vms" == "all" ]]; then
    selected_vms="${AVAILABLE_VMS[*]}"
fi

# Process each selected VM
for vm in $selected_vms; do
    if [[ " ${AVAILABLE_VMS[*]} " =~ " $vm " ]]; then
        echo "Processing VM: $vm"
        
        echo "Starting VM: $vm"
        vagrant up $vm
        
        echo "Installing dployr.io on VM: $vm"
        if ! vagrant ssh $vm -c "cd /vagrant && sudo ./scripts/bootstrap/install.sh"; then
            echo "ERROR: Installation failed on VM: $vm"
            exit 1
        fi
        
        echo "Building Docker image on VM: $vm"
        if ! vagrant ssh $vm -c "cd /vagrant && ./scripts/docker/build.sh"; then
            echo "ERROR: Docker build failed on VM: $vm"
            exit 1
        fi
        
        echo "Starting dployr.io on VM: $vm"
        if ! vagrant ssh $vm -c "cd /vagrant && ./scripts/bootstrap/start.sh"; then
            echo "ERROR: Start failed on VM: $vm"
            exit 1
        fi
        
        echo "VM $vm bootstrap completed successfully"
    else
        echo "Warning: '$vm' is not a valid VM name. Skipping..."
    fi
done

END_TIME=$(date +%s)
ELAPSED_TIME=$((END_TIME - START_TIME))

echo "Bootstrap completed in $((ELAPSED_TIME / 60)) minutes and $((ELAPSED_TIME % 60)) seconds"