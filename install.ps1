# Copyright 2025 Emmanuel Madehin
# SPDX-License-Identifier: Apache-2.0

# Downloads and installs dployr, dployrd, Caddy, and NSSM

param(
    [string]$InstallDir = "$env:ProgramFiles\dployr",
    [Alias('v')]
    [string]$Version = "latest",
    [Alias('t')]
    [string]$Token,
    [Alias('i')]
    [string]$Instance,
    [switch]$Help
)

$ErrorActionPreference = "Stop"

$logDir = Join-Path $env:PROGRAMDATA "dployr"
if (!(Test-Path $logDir)) { New-Item -ItemType Directory -Path $logDir -Force | Out-Null }
$logFile = Join-Path $logDir "install.log"
Start-Transcript -Path $logFile -Append | Out-Null
Write-Host "Logging installer output to $logFile" -ForegroundColor Gray

function Invoke-Fatal {
    param(
        [string]$Message
    )

    Write-Error $Message
    if (Test-Path $logFile) {
        Write-Host "[LOG] Last 20 lines from $logFile:" -ForegroundColor Yellow
        Get-Content $logFile -Tail 20 | Write-Host
        Write-Host "[INFO] Full log available at: $logFile" -ForegroundColor Yellow
    }
    exit 1
}

Write-Host "dployr Windows Installer" -ForegroundColor Green
Write-Host "=========================" -ForegroundColor Green

# Show usage if help requested
if ($Help) {
    Write-Host "" 
    Write-Host "Usage: .\install.ps1 [-Version <VERSION>] -Token <TOKEN> [-Instance <ID>] [-InstallDir <PATH>]" 
    Write-Host "" 
    Write-Host "Arguments:" 
    Write-Host "  -Version    Optional dployr version tag (default: latest)" 
    Write-Host "  -Token      Required install token issued by dployr base" 
    Write-Host "  -Instance   Optional custom instance ID (default: my-instance-id)" 
    Write-Host "  -InstallDir Optional install directory (default: $env:ProgramFiles\\dployr)" 
    Write-Host "" 
    Write-Host "Examples:" 
    Write-Host "  .\install.ps1 -Token $env:DPLOYR_INSTALL_TOKEN" 
    Write-Host "  .\install.ps1 -Version v0.1.1-beta.17 -Token $env:DPLOYR_INSTALL_TOKEN" 
    Write-Host "  .\install.ps1 -Token $env:DPLOYR_INSTALL_TOKEN -Instance prod-server-01" 
    Write-Host "  .\install.ps1 -Version latest -Token $env:DPLOYR_INSTALL_TOKEN -InstallDir C:\dployr" 
    Write-Host "" 
    Write-Host "Available versions: https://github.com/dployr-io/dployr/releases" 
    exit 0
}

if (-not $Token) {
    Invoke-Fatal "Missing required -Token parameter. Run with -Help for usage."
}

function Install-Git {
    if (Get-Command git -ErrorAction SilentlyContinue) {
        Write-Host "✓ git already installed"
        return
    }

    Write-Host "Installing git..."
    try {
        & winget install Git.Git --silent --accept-source-agreements
        Write-Host "✓ git installed successfully"
    } catch {
        Invoke-Fatal "Failed to install git: $_"
    }
}

function Register-Instance {
    param(
        [string]$Token
    )

    if (-not $Token) {
        Write-Warning "No install token provided; skipping /system/register call"
        return
    }

    Write-Host "Registering instance with local dployrd via /system/register..."
    try {
        $body = @{ claim = $Token } | ConvertTo-Json
        Invoke-RestMethod -Method Post -Uri "http://localhost:7879/system/register" -ContentType "application/json" -Body $body | Out-Null
        Write-Host "✓ Instance registration request sent successfully"
    } catch {
        Write-Warning "Failed to register instance with /system/register: $_"
    }
}

# Create install directory
if (!(Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Write-Host "Created directory: $InstallDir"
}

# Detect architecture
$arch = if ([Environment]::Is64BitOperatingSystem) { "x86_64" } else { "i386" }
Write-Host "Detected architecture: $arch"

# Install git
Install-Git

# Get latest version if not specified
if ($Version -eq "latest") {
    try {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/dployr-io/dployr/releases/latest"
        $Version = $release.tag_name
        Write-Host "Latest version: $Version"
    } catch {
        Invoke-Fatal "Failed to get latest version: $_"
    }
}

# Download dployr binaries
$dployrUrl = "https://github.com/dployr-io/dployr/releases/download/$Version/dployr-Windows-$arch.zip"
$dployrZip = "$env:TEMP\dployr.zip"

Write-Host "Downloading dployr binaries..."
try {
    Invoke-WebRequest -Uri $dployrUrl -OutFile $dployrZip
    
    # Stop running daemon if it exists
    try {
        $service = Get-Service -Name "dployrd" -ErrorAction SilentlyContinue
        if ($service -and $service.Status -eq "Running") {
            Write-Host "Stopping running dployrd service..."
            Stop-Service -Name "dployrd" -Force
            Start-Sleep -Seconds 2
        }
    } catch {
        # Service might not exist, that's fine
    }
    
    # Extract to temp directory
    $tempExtract = "$env:TEMP\dployr-extract"
    if (Test-Path $tempExtract) { Remove-Item $tempExtract -Recurse -Force }
    Expand-Archive -Path $dployrZip -DestinationPath $tempExtract -Force
    
    # The archive contains a directory named dployr-Windows-$arch
    $extractedDir = "$tempExtract\dployr-Windows-$arch"
    
    # Copy binaries to install directory
    Copy-Item "$extractedDir\*" $InstallDir -Recurse -Force
    
    # Cleanup
    Remove-Item $dployrZip -Force
    Remove-Item $tempExtract -Recurse -Force
    Write-Host "✓ dployr binaries installed"
} catch {
    Invoke-Fatal "Failed to download dployr: $_"
}

# Download and install Caddy
Write-Host "Installing Caddy..."
try {
    $caddyUrl = "https://github.com/caddyserver/caddy/releases/latest/download/caddy_windows_amd64.zip"
    $caddyZip = "$env:TEMP\caddy.zip"
    $caddyDir = "$InstallDir\caddy"
    
    if (!(Test-Path $caddyDir)) {
        New-Item -ItemType Directory -Path $caddyDir -Force | Out-Null
    }
    
    Invoke-WebRequest -Uri $caddyUrl -OutFile $caddyZip
    Expand-Archive -Path $caddyZip -DestinationPath $caddyDir -Force
    Remove-Item $caddyZip
    Write-Host "✓ Caddy installed"
} catch {
    Invoke-Fatal "Failed to install Caddy: $_"
}

# Download and install NSSM
Write-Host "Installing NSSM..."
try {
    $nssmUrl = "https://nssm.cc/release/nssm-2.24.zip"
    $nssmZip = "$env:TEMP\nssm.zip"
    $nssmDir = "$InstallDir\nssm"
    
    if (!(Test-Path $nssmDir)) {
        New-Item -ItemType Directory -Path $nssmDir -Force | Out-Null
    }
    
    Invoke-WebRequest -Uri $nssmUrl -OutFile $nssmZip
    Expand-Archive -Path $nssmZip -DestinationPath $env:TEMP -Force
    
    # Copy the appropriate architecture version
    $nssmArch = if ([Environment]::Is64BitOperatingSystem) { "win64" } else { "win32" }
    Copy-Item "$env:TEMP\nssm-2.24\$nssmArch\nssm.exe" "$nssmDir\nssm.exe"
    
    Remove-Item $nssmZip
    Remove-Item "$env:TEMP\nssm-2.24" -Recurse -Force
    Write-Host "✓ NSSM installed"
} catch {
    Invoke-Fatal "Failed to install NSSM: $_"
}

# Install vfox
Write-Host "Installing vfox..."
try {
    # Check if vfox is already installed
    $vfoxExists = $false
    try {
        & vfox --version | Out-Null
        $vfoxExists = $true
    } catch {
        # vfox not found, proceed with installation
    }
    
    if ($vfoxExists) {
        Write-Host "✓ vfox already installed"
    } else {
        Write-Host "Installing vfox using winget..."
        & winget install vfox --silent
        Write-Host "✓ vfox installed successfully"
    }
} catch {
    Write-Warning "Failed to install vfox: $_"
    Write-Host "You can install vfox manually with: winget install vfox"
}

# Add to PATH
$currentPath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
$pathsToAdd = @(
    $InstallDir,
    "$InstallDir\caddy",
    "$InstallDir\nssm"
)

$pathUpdated = $false
foreach ($path in $pathsToAdd) {
    if ($currentPath -notlike "*$path*") {
        $currentPath += ";$path"
        $pathUpdated = $true
    }
}

if ($pathUpdated) {
    [Environment]::SetEnvironmentVariable("PATH", $currentPath, "Machine")
    Write-Host "✓ Added to system PATH"
}

# Create and start dployr service using NSSM
Write-Host "Setting up dployrd service..."
try {
    $nssmPath = "$InstallDir\nssm\nssm.exe"
    $dployrdPath = "$InstallDir\dployrd.exe"
    
    # Install service
    & $nssmPath install dployrd $dployrdPath
    & $nssmPath set dployrd DisplayName "dployr Daemon"
    & $nssmPath set dployrd Description "dployr deployment management daemon"
    & $nssmPath set dployrd Start SERVICE_AUTO_START
    
    # Start the service
    & $nssmPath start dployrd
    Write-Host "✓ dployrd service created and started" 
    Write-Host "  Control with: nssm start/stop/restart dployrd" 

    Start-Sleep -Seconds 1
    Register-Instance -Token $Token
} catch {
    Write-Warning "Failed to create/start service: $_"
    Write-Host "You can create the service manually later using NSSM"
}

# Create system-wide config directory and file
$configDir = "$env:PROGRAMDATA\dployr"
$configFile = "$configDir\config.toml"

Write-Host "Creating default configuration..."
if (!(Test-Path $configDir)) {
    New-Item -ItemType Directory -Path $configDir -Force | Out-Null
}

if (!(Test-Path $configFile)) {
    # Use custom instance_id if provided, otherwise use default
    $instanceValue = if ($Instance) { $Instance } else { "my-instance-id" }
    $defaultConfig = @"
# dployr configuration file
address = "localhost"
port = 7879
max-workers = 5

# Base configuration
base_url = "https://base.dployr.io"
instance_id = "$instanceValue"
"@
    $defaultConfig | Out-File -FilePath $configFile -Encoding UTF8
    Write-Host "✓ Created system config at $configFile"
    if ($Instance) {
        Write-Host "✓ Using custom instance_id: $Instance"
    }
} else {
    Write-Host "✓ Config file already exists at $configFile"
}

Write-Host ""
Write-Host "Installation completed successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "Installed components:"
Write-Host "  - dployr.exe (CLI)"
Write-Host "  - dployrd.exe (daemon)"
Write-Host "  - caddy.exe (reverse proxy)"
Write-Host "  - nssm.exe (service manager)"
Write-Host "  - vfox.exe (version manager)"
Write-Host ""

Write-Host "Next steps:"
Write-Host "1. Restart your terminal to use the new PATH"
Write-Host "2. Dployrd is now running"
Write-Host "3. Use the CLI: dployr --help"
Write-Host ""
Write-Host "Service management:"
Write-Host "- Status: nssm status dployrd"
Write-Host "- Stop: nssm stop dployrd"
Write-Host "- Restart: nssm restart dployrd"
Stop-Transcript | Out-Null