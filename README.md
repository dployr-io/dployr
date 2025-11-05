# dployr

*Your app, your server, your rules!*

Turn any machine into a deployment platform. Deploy applications from Git repositories or Docker images with automatic reverse proxy, SSL certificates, and service management.

---

## Overview

`dployr` gives developers a self-hosted alternative to managed platforms.  
It combines a lightweight daemon, a CLI client, and powerful integrations to automate deployment pipelines across operating systems.

---

## Architecture

`dployr` consists of two main components:

- **dployr** — Command-line client  
- **dployrd** — Background daemon that handles deployment execution, service management, and API endpoints

- **SQLite** for persistence  
- **Caddy** for automatic HTTPS and reverse proxy  
- **systemd** on Linux, **launchd** on macOS, or **NSSM** on Windows for service management  
All components are written in Go and packaged as standalone binaries.

---

## Dependencies

`dployr` relies on a few external tools to provide core functionality:

| Dependency | Purpose | Platform |
|-------------|----------|-----------|
| [Caddy](https://caddyserver.com) | Reverse proxy and automatic HTTPS management | All |
| [vfox](https://version-fox.dev) | Runtime and version management (Node, Python, Go, etc.) | All |
| [systemd](https://systemd.io) | Linux service manager for running `dployrd` as a background daemon | Linux |
| [launchd](https://developer.apple.com/library/archive/documentation/MacOSX/Conceptual/BPSystemStartup/Chapters/CreatingLaunchdJobs.html) | macOS service manager for running `dployrd` as a background daemon | macOS |
| [NSSM](https://nssm.cc) | Windows service manager for running `dployrd` as a background daemon | Windows |
> **Note:** Each operating system uses its native service manager: `systemd` on Linux, `launchd` on macOS, and `NSSM` on Windows (uses Windows Process API).
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
# Install latest version
curl -sSL https://raw.githubusercontent.com/dployr-io/dployr/master/install.sh | bash

# Install specific version
curl -sSL https://raw.githubusercontent.com/dployr-io/dployr/master/install.sh | bash -s v0.1.1-beta.17
```

> **Note**: The installer automatically generates a secure secret key and creates a system-wide config file. The secret is shown once during installation - save it securely!
> 
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
.\install.ps1 -Version v0.1.1-beta.17

# Custom install directory
.\install.ps1 -Version latest -InstallDir "C:\dployr"
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
github-token = "ghp_xxxx"  # Optional: for private repositories
```

Environment variables can override these values:

* `ADDRESS` — Server bind address
* `PORT` — Server port
* `MAX_WORKERS` — Concurrent deployment limit
* `SECRET` — JWT signing secret *(required)*
* `GITHUB_TOKEN` — GitHub access token

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
  --name my-app \
  --source remote \
  --runtime nodejs \
  --version 18 \
  --remote https://github.com/user/repo.git \
  --branch main \
  --build-cmd "npm install" \
  --run-cmd "npm start" \
  --port 3000
```

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

MIT License — see [LICENSE](LICENSE) for details.

---

## Documentation

Complete documentation available at
 [https://docs.dployr.dev](https://docs.dployr.dev)

---

