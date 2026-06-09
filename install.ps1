# Copyright 2025 Emmanuel Madehin
# SPDX-License-Identifier: Apache-2.0

# dployr Windows Installer (CLI only — daemon support is Linux/macOS only)

param(
    [string]$InstallDir = "$env:ProgramFiles\dployr",
    [Alias('v')]
    [string]$Version = "latest",
    [switch]$Help
)

$ErrorActionPreference = "Stop"

if ($Help) {
    Write-Host "Usage: .\install.ps1 [-Version <tag>] [-InstallDir <path>]"
    Write-Host ""
    Write-Host "  -Version    dployr version to install (default: latest)"
    Write-Host "  -InstallDir Installation directory (default: $env:ProgramFiles\dployr)"
    Write-Host ""
    Write-Host "Examples:"
    Write-Host "  .\install.ps1"
    Write-Host "  .\install.ps1 -Version v0.3.1"
    exit 0
}

Write-Host "dployr Windows Installer (CLI)" -ForegroundColor Green
Write-Host "==============================" -ForegroundColor Green
Write-Host "Note: Windows support is experimental. The daemon (dployrd) is not supported on Windows." -ForegroundColor Yellow
Write-Host ""

# Resolve version
if ($Version -eq "latest") {
    Write-Host "Fetching latest release..."
    try {
        $release = Invoke-RestMethod -Uri "https://api.github.com/repos/dployr-io/dployr/releases/latest"
        $Version = $release.tag_name
        Write-Host "Latest version: $Version"
    } catch {
        Write-Error "Failed to fetch latest version: $_"
        exit 1
    }
}

$arch = if ([Environment]::Is64BitOperatingSystem) { "x86_64" } else { "i386" }
$url  = "https://github.com/dployr-io/dployr/releases/download/$Version/dployr-Windows-$arch.zip"
$zip  = "$env:TEMP\dployr.zip"

Write-Host "Downloading dployr $Version..."
try {
    Invoke-WebRequest -Uri $url -OutFile $zip
} catch {
    Write-Error "Failed to download dployr: $_"
    exit 1
}

if (!(Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

$tmp = "$env:TEMP\dployr-extract"
if (Test-Path $tmp) { Remove-Item $tmp -Recurse -Force }
Expand-Archive -Path $zip -DestinationPath $tmp -Force
Copy-Item "$tmp\dployr-Windows-$arch\dployr.exe" $InstallDir -Force
Remove-Item $zip -Force
Remove-Item $tmp  -Recurse -Force
Write-Host "✓ dployr.exe installed to $InstallDir"

# Add to PATH
$machinePath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
if ($machinePath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$machinePath;$InstallDir", "Machine")
    Write-Host "✓ Added $InstallDir to system PATH"
}

# Write minimal config
$configDir  = "$env:PROGRAMDATA\dployr"
$configFile = "$configDir\config.toml"
if (!(Test-Path $configDir)) {
    New-Item -ItemType Directory -Path $configDir -Force | Out-Null
}
if (!(Test-Path $configFile)) {
    @"
base_url = "https://base.dployr.io"
"@ | Out-File -FilePath $configFile -Encoding UTF8
    Write-Host "✓ Created config at $configFile"
}

# Install PowerShell completion
$completionScript = & "$InstallDir\dployr.exe" completion powershell 2>$null
if ($completionScript) {
    if (!(Test-Path -Path (Split-Path $PROFILE))) {
        New-Item -ItemType Directory -Path (Split-Path $PROFILE) -Force | Out-Null
    }
    if (!(Test-Path $PROFILE)) {
        New-Item -ItemType File -Path $PROFILE -Force | Out-Null
    }
    $marker = "# dployr completion"
    if (!(Select-String -Path $PROFILE -Pattern "dployr completion" -Quiet 2>$null)) {
        Add-Content -Path $PROFILE -Value "`n$marker`n$completionScript"
        Write-Host "✓ PowerShell completion added to $PROFILE"
    }
}

Write-Host ""
Write-Host "Done! Restart your terminal then run: dployr --help" -ForegroundColor Green
