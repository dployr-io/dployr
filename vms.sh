#!/bin/bash

# Configuration
RESOURCE_GROUP="TestSuite"
LOCATION="eastus"  # Change to your preferred region
ADMIN_USERNAME="azureuser"
SSH_KEY_PATH="$HOME/.ssh/id_rsa.pub"  # Existing SSH public key

# VM Configuration - Minimal B1s size (1 vCPU, 1GB RAM)
declare -A VMS=(
    ["ubuntu20"]="Canonical:0001-com-ubuntu-server-focal:20_04-lts-gen2:latest"
    ["debian11"]="Debian:debian-11:11:latest"
    ["centos"]="OpenLogic:CentOS:7_9-gen2:latest"
    ["suse"]="SUSE:sles-15-sp5:gen2:latest"
)

# Create Resource Group
az group create --name $RESOURCE_GROUP --location $LOCATION

# Create Network Infrastructure
az network vnet create \
    --resource-group $RESOURCE_GROUP \
    --name VNet \
    --address-prefix 10.0.0.0/16 \
    --subnet-name default \
    --subnet-prefix 10.0.0.0/24

# Create each VM with minimal configuration
for vm_name in "${!VMS[@]}"; do
    IMAGE_URN=${VMS[$vm_name]}
    
    echo "Creating $vm_name..."
    az vm create \
        --resource-group $RESOURCE_GROUP \
        --name $vm_name \
        --image $IMAGE_URN \
        --size Standard_B1s \
        --admin-username $ADMIN_USERNAME \
        --ssh-key-values @$SSH_KEY_PATH \
        --public-ip-sku Basic \
        --vnet-name VNet \
        --subnet default \
        --storage-sku Standard_LRS \
        --os-disk-size-gb 32 \
        --no-wait
done

# Add HTTP/HTTPS rules to each VM's NSG
for vm_name in "${!VMS[@]}"; do
    echo "Opening ports 80/443 for $vm_name..."
    az vm open-port \
        --resource-group $RESOURCE_GROUP \
        --name $vm_name \
        --port 80 \
        --priority 300
    
    az vm open-port \
        --resource-group $RESOURCE_GROUP \
        --name $vm_name \
        --port 443 \
        --priority 301
done

echo "All VMs are being deployed..."
echo "Monitor creation progress with:"
echo "az vm list -g $RESOURCE_GROUP -o table --show-details"
echo ""
echo "Connect using:"
echo "ssh -i ${SSH_KEY_PATH%.pub} $ADMIN_USERNAME@<PUBLIC_IP>"
