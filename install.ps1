# dployr Windows Installer
# Downloads and installs dployr, dployrd, Caddy, and NSSM

param(
    [string]$InstallDir = "$env:ProgramFiles\dployr",
    [string]$Version = "latest",
    [switch]$Help
)

$ErrorActionPreference = "Stop"

Write-Host "dployr Windows Installer" -ForegroundColor Green
Write-Host "=========================" -ForegroundColor Green

# Show usage if help requested
if ($Help) {
    Write-Host ""
    Write-Host "Usage: .\install.ps1 [-Version <VERSION>] [-InstallDir <PATH>]"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\install.ps1                           # Install latest version"
    Write-Host "  .\install.ps1 -Version v0.1.1-beta.17   # Install specific version"
    Write-Host "  .\install.ps1 -InstallDir C:\dployr     # Custom install directory"
    Write-Host ""
    Write-Host "Available versions: https://github.com/dployr-io/dployr/releases"
    exit 0
}

# Create install directory
if (!(Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Write-Host "Created directory: $InstallDir"
}

# Detect architecture
$arch = if ([Environment]::Is64BitOperatingSystem) { "x86_64" } else { "i386" }
Write-Host "Detected architecture: $arch"

# Get latest version if not specified
if ($Version -eq "latest") {
    try {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/dployr-io/dployr/releases/latest"
        $Version = $release.tag_name
        Write-Host "Latest version: $Version"
    } catch {
        Write-Error "Failed to get latest version: $_"
        exit 1
    }
}

# Download dployr binaries
$dployrUrl = "https://github.com/dployr-io/dployr/releases/download/$Version/dployr-Windows-$arch.zip"
$dployrZip = "$env:TEMP\dployr.zip"

Write-Host "Downloading dployr binaries..."
try {
    Invoke-WebRequest -Uri $dployrUrl -OutFile $dployrZip
    
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
    Write-Error "Failed to download dployr: $_"
    exit 1
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
    Write-Error "Failed to install Caddy: $_"
    exit 1
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
    Write-Error "Failed to install NSSM: $_"
    exit 1
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
    # Generate a secure random secret
    $secret = -join ((1..32) | ForEach {'{0:X}' -f (Get-Random -Max 16)})
    
    $defaultConfig = @"
# dployr configuration file
address = "localhost"
port = 7879
max-workers = 5

# Secret key for JWT signing (auto-generated)
secret = "$secret"
"@
    $defaultConfig | Out-File -FilePath $configFile -Encoding UTF8
    Write-Host "✓ Created system config at $configFile"
    $global:ShowSecret = $secret
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
Write-Host ""

# Show the generated secret once
if ($global:ShowSecret) {
    Write-Host "==========================================" -ForegroundColor Yellow
    Write-Host "YOUR SECRET KEY (SAVE THIS NOW!):" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "  $($global:ShowSecret)" -ForegroundColor White
    Write-Host ""
    Write-Host "This secret will NOT be shown again!" -ForegroundColor Red
    Write-Host "It's saved in: $configFile" -ForegroundColor Gray
    Write-Host "==========================================" -ForegroundColor Yellow
    Write-Host ""
}

Write-Host "Next steps:"
Write-Host "1. Restart your terminal to use the new PATH"
Write-Host "2. Dployrd is now running"
Write-Host "3. Use the CLI: dployr --help"
Write-Host ""
Write-Host "Service management:"
Write-Host "- Status: nssm status dployrd"
Write-Host "- Stop: nssm stop dployrd"
Write-Host "- Restart: nssm restart dployrd"