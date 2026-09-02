#!/usr/bin/env pwsh
$ErrorActionPreference = 'Stop'

$Repo = "sarner/envbox"
$InstallDir = "$HOME\.envbox\bin"

function Detect-Architecture {
    switch ($env:PROCESSOR_ARCHITECTURE) {
        "AMD64"   { return "amd64" }
        "ARM64"   { return "arm64" }
        "X86"     { return "386" }
        default   { return $null }
    }
}

function Add-ToPath {
    param([string]$Path)
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$Path*") {
        [Environment]::SetEnvironmentVariable("Path", "$userPath;$Path", "User")
        $env:Path = "$env:Path;$Path"
        Write-Host "Added $Path to user PATH"
    }
}

Write-Host "Installing envbox..."

$Arch = Detect-Architecture
if (-not $Arch) {
    Write-Error "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE"
    exit 1
}

$VersionUrl = "https://api.github.com/repos/$Repo/releases/latest"
try {
    $Version = (Invoke-RestMethod $VersionUrl).tag_name
} catch {
    Write-Error "Failed to fetch latest release: $_"
    exit 1
}

if (-not $Version) {
    Write-Error "Could not determine latest version"
    exit 1
}

$Filename = "envbox_windows_${Arch}.zip"
$DownloadUrl = "https://github.com/$Repo/releases/download/$Version/$Filename"
$TempZip = Join-Path $env:TEMP "envbox_install.zip"
$TempDir = Join-Path $env:TEMP "envbox_install"

Write-Host "Downloading $Filename..."

try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TempZip -UseBasicParsing
} catch {
    Write-Error "Failed to download: $_"
    exit 1
}

if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

Expand-Archive -Path $TempZip -DestinationPath $TempDir -Force
Move-Item -Path (Join-Path $TempDir "envbox.exe") -Destination $InstallDir -Force

Remove-Item $TempZip -Force -ErrorAction SilentlyContinue
Remove-Item $TempDir -Recurse -Force -ErrorAction SilentlyContinue

Add-ToPath -Path $InstallDir

$envboxCmd = Join-Path $InstallDir "envbox.exe"
if (Test-Path $envboxCmd) {
    Write-Host "Successfully installed envbox"
    & $envboxCmd --version
} else {
    Write-Error "Installation failed"
    exit 1
}
