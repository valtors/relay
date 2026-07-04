# Relay installer for Windows
# Usage: irm https://raw.githubusercontent.com/valtors/relay/main/scripts/install.ps1 | iex

$ErrorActionPreference = "Stop"
$Repo = "valtors/relay"

# Detect architecture
$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
$OS = "windows"

# Get latest release
$Release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
$Version = $Release.tag_name -replace '^v', ''

if (-not $Version) {
    Write-Error "Could not determine latest version."
    exit 1
}

$FileName = "relay_${Version}_${OS}_${Arch}.zip"
$URL = "https://github.com/$Repo/releases/download/v${Version}/$FileName"

Write-Host "Downloading relay v${Version} for ${OS}/${Arch}..." -ForegroundColor Cyan

# Download and extract
$TmpDir = Join-Path $env:TEMP "relay-install"
New-Item -ItemType Directory -Path $TmpDir -Force | Out-Null
$ZipPath = Join-Path $TmpDir $FileName

Invoke-WebRequest -Uri $URL -OutFile $ZipPath
Expand-Archive -Path $ZipPath -DestinationPath $TmpDir -Force

# Install to user's local bin
$InstallDir = Join-Path $env:LOCALAPPDATA "relay"
New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
Move-Item -Path (Join-Path $TmpDir "relay.exe") -Destination (Join-Path $InstallDir "relay.exe") -Force

# Add to PATH if not already there
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
    Write-Host "Added $InstallDir to PATH (restart your terminal)" -ForegroundColor Yellow
}

Remove-Item -Recurse -Force $TmpDir

Write-Host ""
Write-Host "relay v${Version} installed to $InstallDir\relay.exe" -ForegroundColor Green
Write-Host "Run: relay" -ForegroundColor White
