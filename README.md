# dployr

*Your app, your server, your rules!*

[![Tests](https://github.com/dployr-io/dployr/actions/workflows/tests.yml/badge.svg)](https://github.com/dployr-io/dployr/actions/workflows/tests.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/dployr-io/dployr.svg)](https://pkg.go.dev/github.com/dployr-io/dployr)
[![Go Report Card](https://goreportcard.com/badge/github.com/dployr-io/dployr)](https://goreportcard.com/report/github.com/dployr-io/dployr)
[![Release](https://img.shields.io/github/v/release/dployr-io/dployr)](https://github.com/dployr-io/dployr/releases)
[![License: Apache License, Version 2.0](https://img.shields.io/badge/License-Apache%20License%2C%20Version%202.0-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.24%2B-00ADD8?logo=go)](https://go.dev/dl/)

## Overview

dployr is a self‑hosted platform with a globally distributed control‑plane (“base”) and lightweight agents that run on your infrastructure.

You interact with base through:
- The web dashboard (dployr‑app), or
- The CLI (dployr‑cli)

Both use the same programmatic, RBAC‑aware API with full auditing. Anything you do in the UI can be scripted with the CLI or called directly.

It consists of four components:

- dployr‑base — Globally distributed control‑plane (API, scheduling, storage).
- dployrd — Daemon on each instance. Connects to base over mTLS, executes tasks, and reports status.
- dployr‑cli — RBAC‑aware command‑line client that talks to base from anywhere.
- dployr‑app — Web dashboard built on the same API for managing projects, deployments, and environments.

## Quickstart (5 minutes)

Linux/macOS

```bash
# First time install
curl -sSL https://raw.githubusercontent.com/dployr-io/dployr/master/install.sh \
  | bash -s -- --token "<bootstrap_token>"

# Install latest version
curl -sSL https://raw.githubusercontent.com/dployr-io/dployr/master/install.sh | bash

# Start the daemon
dployrd
```

Windows (PowerShell as Administrator)

```powershell
iwr "https://raw.githubusercontent.com/dployr-io/dployr/master/install.ps1" -OutFile install.ps1
.\install.ps1  # add -Token $env:DPLOYR_INSTALL_TOKEN (first time install)

dployrd.exe
```

## Verify

- Version: `dployrd --version`
- Logs (JSON): `/var/log/dployrd/app.log` (Linux/macOS) or ProgramData on Windows
- Daemon should start and log a websocket mTLS connection attempt to Base.

## Quick deploy

```bash
dployr deploy \
  --name hello-world \
  --source remote \
  --runtime nodejs \
  --version 20 \
  --remote https://github.com/dployr-io/dployr-examples \
  --branch master \
  --build-cmd "npm install" \
  --run-cmd "npm start" \
  --working-dir "nodejs" \
  --port 3000
```

## Concepts

- CLI vs Daemon: CLI issues commands; daemon (`dployrd`) executes and syncs with Base.
- Sync: long‑lived WSS + mTLS. Daemon generates a client cert and publishes it to Base.
- Tokens: `bootstrap_token` (long‑lived, stored in DB) → exchanged for short‑lived `access_token` (auto‑refreshed).
- Persistence: SQLite for instance metadata, tokens, deployments, services, task results.
- Logging: structured JSON to stdout + `/var/log/dployrd/app.log` for remote debugging.

## Troubleshooting 

- No bootstrap token: set it in config or rerun installer with `--token`.
- WS auth errors (401/403): daemon clears `access_token` and reacquires; check logs.
- mTLS/cert issues: ensure pinned CA/cert path is correct if you customized certs.
- Permissions: service managers may require admin/root for install/start.
- Where are logs? `/var/log/dployrd/app.log` (Linux/macOS). On Windows, see ProgramData/dployr.

---

## Dependencies

`dployr` relies on a few external tools to provide core functionality:
| Dependency | Purpose | Platform |
|-------------|----------|-----------|
| [Caddy](https://caddyserver.com) | Reverse proxy and automatic HTTPS management | All |
| [vfox](https://version-fox.dev) | Runtime and version management (Node, Python, Go, etc.) | All |
| [SQLite](https://sqlite.org) | Embedded database for persistence and data storage | All |
| [systemd](https://systemd.io) | Linux service manager for running `dployrd` as a background daemon | Linux |
| [launchd](https://developer.apple.com/library/archive/documentation/MacOSX/Conceptual/BPSystemStartup/Chapters/CreatingLaunchdJobs.html) | macOS service manager for running `dployrd` as a background daemon | macOS |
| [NSSM](https://nssm.cc) | Windows service manager for running `dployrd` as a background daemon | Windows |
> **Note:** Each operating system uses its native service manager: `systemd` on Linux, `launchd` on macOS, and `NSSM` on Windows.
These binaries are automatically downloaded and configured during installation when possible.

---

## Supported Runtimes

- Static files
- Node.js
- Python
- Go
- PHP
- Ruby
- .NET
- Java
- Docker containers
- K3s clusters *(roadmap)*
- Custom runtimes

---

## Installation

### Quick Install

**Linux/macOS**
```bash
# First-time setup 
curl -sSL https://raw.githubusercontent.com/dployr-io/dployr/master/install.sh \
  | bash -s -- --token "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."

# Install with custom instance ID
curl -sSL https://raw.githubusercontent.com/dployr-io/dployr/master/install.sh \
  | bash -s -- --token "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..." --instance "prod-server-01"

# Install latest version 
curl -sSL https://raw.githubusercontent.com/dployr-io/dployr/master/install.sh \
  | bash 

# Install specific version
curl -sSL https://raw.githubusercontent.com/dployr-io/dployr/master/install.sh \
  | bash -s -- --version v0.1.1-beta.17 
```
> **Note:** You’ll need administrator privileges (`sudo`) to install, but the dployr daemon itself runs as `dployr` user, not as `root`. For details about permissions and setup, visit https://docs.dployr.dev/permissions.

> **Config locations:**
> - **Linux**: `/etc/dployr/config.toml`
> - **macOS**: `/usr/local/etc/dployr/config.toml`  
> - **Windows**: `C:\ProgramData\dployr\config.toml`

### Manual Installation

**Linux**
```bash
# Download the latest release
curl -L https://github.com/dployr-io/dployr/releases/latest/download/dployr-Linux-x86_64.tar.gz -o dployr.tar.gz

# Extract and install
tar -xzf dployr.tar.gz
sudo mv dployr dployrd /usr/local/bin/
chmod +x /usr/local/bin/dployr /usr/local/bin/dployrd

# Install Caddy 
sudo apt update && sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update && sudo apt install caddy
```

**Windows (PowerShell as Administrator)**

```powershell
# Install latest version
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/dployr-io/dployr/master/install.ps1" -OutFile "install.ps1"
.\install.ps1

# Install specific version
Invoke-WebRequest -Uri "https://raw.githubusercontent.com/dployr-io/dployr/master/install.ps1" -OutFile "install.ps1"
.\install.ps1 -Version v0.1.1-beta.17 -Token $env:DPLOYR_INSTALL_TOKEN

# Install with custom instance ID
.\install.ps1 -Token $env:DPLOYR_INSTALL_TOKEN -Instance "prod-server-01"

# Custom install directory
.\install.ps1 -Version latest -InstallDir "C:\dployr" -Token $env:DPLOYR_INSTALL_TOKEN
```

---

### Package Managers

**Homebrew (macOS)**

```bash
brew install dployr-io/dployr/dployr
```

**Chocolatey (Windows)**

```powershell
choco install dployr
```

**Scoop (Windows)**

```powershell
scoop bucket add dployr https://github.com/dployr-io/scoop-dployr
scoop install dployr
```

---

## Configuration

Configuration files are stored at `~/.dployr/config.toml`:

```toml
address = "localhost"
port = 7879
max-workers = 5
secret = "your-secret-key"
```

---

## Usage

### Start the Daemon

```bash
dployrd
```

### Authenticate

```bash
dployr login --email user@example.com
```

### Deploy an Application

```bash
dployr deploy \
  --name old-county-times \
  --source remote \
  --runtime php \
  --version 8.3 \
  --remote https://github.com/dployr-io/dployr-examples \
  --branch master \
  --build-cmd "composer install" \
  --run-cmd "php -S localhost:3000" \
  --working-dir "php" \
  --port 3000
```
For detailed deployment examples—including interactive guides—visit: https://docs.dployr.dev/runtimes/quick-deploy

### List Deployments

```bash
dployr list
```

### Manage Services

```bash
dployr services list
dployr services get <service-id>
```

### Configure Proxy

```bash
dployr proxy add --domain myapp.com --upstream http://localhost:3000
dployr proxy list
dployr proxy status
```

---

## API

The daemon exposes a REST API on port `7879` (configurable).

* Full spec: [api/openapi.yaml](api/openapi.yaml)
* Interactive docs: [docs.dployr.dev](https://docs.dployr.dev)

### Authentication

Most endpoints require JWT authentication:

```
Authorization: Bearer <jwt-token>
```

---

## Development

### Prerequisites

* Go **1.24+**
* Git

### Build

```bash
make build          # Build both binaries
make build-cli      # Build CLI only
make build-daemon   # Build daemon only
```

### Test

```bash
make test
```

### Release

```bash
./scripts/release.sh patch    # Patch release
./scripts/release.sh minor    # Minor release
./scripts/release.sh major    # Major release
```

---

## License

Apache License, Version 2.0 — see [LICENSE](LICENSE) for details.

---

## Documentation

Complete documentation available at
 [https://docs.dployr.dev](https://docs.dployr.dev)

---

