# dployr

[![Tests](https://github.com/dployr-io/dployr/actions/workflows/tests.yml/badge.svg)](https://github.com/dployr-io/dployr/actions/workflows/tests.yml)
[![Release](https://img.shields.io/github/v/release/dployr-io/dployr)](https://github.com/dployr-io/dployr/releases)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.24%2B-00ADD8?logo=go)](https://go.dev/dl/)

**Ship apps, not infrastructure.**

Your server is ready the moment you sign up. Deploy from the CLI, GitHub Actions, or the dashboard. No SSH. No config files. On Pro, connect your own server.

[dployr.io](https://dployr.io) · [Docs](https://dployr.io/docs/introduction) · [Live demo](https://dployr.io/demo)

---

## What's in this repo

**`dployr`** is the CLI. It talks to Base, the hosted control plane.

**`dployrd`** is the daemon that runs on your server. It opens an outbound WebSocket to Base and executes deploy instructions on arrival. Your server never accepts inbound connections.

On Hobby and Indie plans, dployr provisions and manages the server. On Pro, you run `dployrd` on your own Linux server.

## Install

**macOS**
```bash
brew install dployr-io/dployr/dployr
```

**Windows**
```powershell
scoop bucket add dployr https://github.com/dployr-io/scoop-dployr
scoop install dployr
```

**Linux / macOS (script)**
```bash
curl -sSL https://raw.githubusercontent.com/dployr-io/dployr/master/install.sh | bash
```

**Windows (script)**
```powershell
iwr https://raw.githubusercontent.com/dployr-io/dployr/master/install.ps1 | iex
```

## Connect your own server (Pro)

```bash
curl -sSL https://raw.githubusercontent.com/dployr-io/dployr/master/install.sh \
  | bash -s -- --token "<bootstrap-token>"
```

Get the bootstrap token from the **Instances** page in the dashboard: [dployr.io/docs/byos](https://dployr.io/docs/byos)

## Deploy

```bash
dployr auth login --email you@example.com

dployr deployments create \
  --name my-api \
  --source remote \
  --runtime nodejs \
  --runtime-version 20 \
  --remote https://github.com/your-org/your-repo \
  --branch main \
  --run-cmd "npm start" \
  --port 3000
```

## Common commands

```bash
# Follow build logs
dployr logs my-api --build --follow

# List services and deployments
dployr services list
dployr deployments list

# Stop and start a service
dployr services stop my-api
dployr services start my-api

# List instances
dployr instances list
```

Full reference: [dployr.io/docs/cli](https://dployr.io/docs/cli)

## Runtimes

Node.js, Python, Go, PHP, Ruby, .NET, Java, Static, Docker

## Build

Requires Go 1.24+.

```bash
make build
```

## Contributing

[CONTRIBUTING.md](CONTRIBUTING.md)

## License

Apache 2.0
