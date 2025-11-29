#!/usr/bin/env bash

# Copyright 2025 Emmanuel Madehin
# SPDX-License-Identifier: Apache-2.0

# systemd.sh — manages systemd system services
# Usage: systemd.sh <action> <name> [description] [exec_start] [working_dir]

set -e

action="$1"
name="$2"
description="${3:-}"
exec_start="${4:-}"
working_dir="${5:-}"

case "$action" in
    install)
        sudo mkdir -p "/etc/systemd/system"
        mkdir -p "$HOME/.dployr/logs"
        mkdir -p "$HOME/.dployr/scripts"
        
        # Create wrapper script that loads vfox environment
        exe_script="$HOME/.dployr/scripts/${name}.sh"
        cat > "$exe_script" <<EOF
#!/usr/bin/env bash
set -e

export VFOX_HOME="\$HOME/.version-fox"
[ -f "\$VFOX_HOME/bin/vfox" ] && eval "\$(\$VFOX_HOME/bin/vfox activate bash)"

cd "$working_dir" || exit 1
exec $exec_start
EOF
        
        chmod +x "$exe_script"
        
        # Create systemd service file
        sudo tee "/etc/systemd/system/${name}.service" > /dev/null <<EOF
[Unit]
Description=$description
After=network.target

[Service]
Type=simple
ExecStart=/bin/bash $exe_script
WorkingDirectory=$working_dir
Restart=always
RestartSec=10
StandardOutput=append:$HOME/.dployr/logs/${name}.log
StandardError=append:$HOME/.dployr/logs/${name}.log

[Install]
WantedBy=multi-user.target
EOF
        
        sudo systemctl daemon-reload
        sudo systemctl enable "$name"
        echo "Service $name installed successfully"
        ;;
        
    start)
        sudo systemctl start "$name"
        echo "Service $name started successfully"
        ;;
        
    stop)
        sudo systemctl stop "$name"
        echo "Service $name stopped successfully"
        ;;
        
    remove)
        sudo systemctl stop "$name" 2>/dev/null || true
        sudo systemctl disable "$name" 2>/dev/null || true
        sudo rm -f "/etc/systemd/system/${name}.service"
        rm -f "$HOME/.dployr/scripts/${name}.sh"
        sudo systemctl daemon-reload
        echo "Service $name removed successfully"
        ;;
        
    status)
        if [ ! -f "/etc/systemd/system/${name}.service" ]; then
            echo "stopped"
            exit 1
        fi
        
        if sudo systemctl is-active --quiet "$name" 2>/dev/null; then
            echo "running"
        else
            echo "stopped"
        fi
        ;;
        
    *)
        echo "Usage: $0 {install|start|stop|remove|status} <name> [description] [exec_start] [working_dir]"
        exit 1
        ;;
esac
exit 0
