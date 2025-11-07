#!/usr/bin/env pwsh
# setup_runtime.ps1 — sets up and uses a specific runtime using vfox
# Usage: setup_runtime.ps1 <runtime> <version> <workdir> <build_command>

param(
    [string]$Runtime,
    [string]$Version,
    [string]$Workdir,
    [string]$BuildCmd
)

function Log($msg) {
    Write-Host "[$(Get-Date -Format o)] $msg"
}

function Abort($msg) {
    Log "ERROR: $msg"
    exit 1
}

if (-not (Get-Command vfox -ErrorAction SilentlyContinue)) {
    Abort "vfox is not installed or not in PATH"
}

Log "Installing runtime: $Runtime@$Version"
vfox install "$Runtime@$Version" -y | Out-Null

Log "Activating runtime: $Runtime@$Version"
Invoke-Expression "$(vfox env -s pwsh $Runtime@$Version)"

switch ($Runtime) {
    'nodejs' {
        if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
            Abort "node not found after activation"
        }
        Log "Node version: $(node -v)"
    }
    'go' {
        if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
            Abort "go not found after activation"
        }
        Log "Go version: $(go version)"
    }
}

Set-Location $Workdir

if ($BuildCmd) {
    Log "Running build command: $BuildCmd"
    Invoke-Expression $BuildCmd
} else {
    Log "No build command specified"
}

Log "Runtime $Runtime@$Version setup complete"
exit 0
