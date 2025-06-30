#!/bin/bash

# Start timer for bootstrap process
START_TIME=$(date +%s)
echo "Bootstrap started at: $(date)"

. ./bootstrap/clean-up.sh

vagrant up ubuntu20

vagrant ssh ubuntu20 -c "cd /vagrant && sudo ./bootstrap/install.sh"

vagrant ssh ubuntu20 -c "cd /vagrant && ./bootstrap/build.sh"

vagrant ssh ubuntu20 -c "cd /vagrant && ./bootstrap/start.sh"

END_TIME=$(date +%s)
ELAPSED_TIME=$((END_TIME - START_TIME))

echo "Bootstrap completed in $((ELAPSED_TIME / 60)) minutes and $((ELAPSED_TIME % 60)) seconds"