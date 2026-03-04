# SandwichLab ahcli Installation Script for Windows
# Usage: irm https://raw.githubusercontent.com/sandwichlab-ai/sandwichlab-skills/main/scripts/install-ahcli.ps1 | iex

$ErrorActionPreference = "Stop"

# Configuration
$Repo = "sandwichlab-ai/sandwichlab-skills"
$BinaryName = "ahcli"
$InstallDir = "$env:USERPROFILE\bin"

# Colors
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Green
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Write-Error-Custom {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

function Get-LatestVersion {
    Write-Info "Fetching latest version..."
    
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
        $version = $response.tag_name
        Write-Info "Latest version: $version"
        return $version
    }
    catch {
        Write-Error-Custom "Failed to fetch latest version: $_"
        exit 1
    }
}

function Download-Binary {
    param([string]$Version)
    
    $binaryName = "$BinaryName-windows-amd64.exe"
    $downloadUrl = "https://github.com/$Repo/releases/download/$Version/$binaryName"
    $tempFile = "$env:TEMP\$binaryName"
    
    Write-Info "Downloading $binaryName..."
    
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $tempFile
        Write-Info "Download complete"
        return $tempFile
    }
    catch {
        Write-Error-Custom "Failed to download binary from $downloadUrl : $_"
        exit 1
    }
}

function Verify-Checksum {
    param([string]$BinaryFile, [string]$Version)
    
    Write-Info "Verifying checksum..."
    
    try {
        $checksumsUrl = "https://github.com/$Repo/releases/download/$Version/checksums.txt"
        $checksums = Invoke-RestMethod -Uri $checksumsUrl
        
        $binaryName = Split-Path $BinaryFile -Leaf
        $expectedChecksum = ($checksums -split "`n" | Where-Object { $_ -match $binaryName } | ForEach-Object { ($_ -split "\s+")[0] })
        
        if (-not $expectedChecksum) {
            Write-Warn "Checksum not found for $binaryName, skipping verification"
            return
        }
        
        $actualChecksum = (Get-FileHash -Path $BinaryFile -Algorithm SHA256).Hash.ToLower()
        
        if ($expectedChecksum -ne $actualChecksum) {
            Write-Error-Custom "Checksum verification failed!"
            Write-Error-Custom "Expected: $expectedChecksum"
            Write-Error-Custom "Actual:   $actualChecksum"
            exit 1
        }
        
        Write-Info "Checksum verified ✓"
    }
    catch {
        Write-Warn "Failed to verify checksum: $_"
    }
}

function Install-Binary {
    param([string]$BinaryFile)
    
    # Create install directory if it doesn't exist
    if (-not (Test-Path $InstallDir)) {
        Write-Info "Creating directory $InstallDir..."
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }
    
    $installPath = Join-Path $InstallDir "$BinaryName.exe"
    
    Write-Info "Installing to $installPath..."
    
    # Remove existing binary if it exists
    if (Test-Path $installPath) {
        Remove-Item $installPath -Force
    }
    
    Move-Item $BinaryFile $installPath -Force
    
    Write-Info "Installation complete ✓"
}

function Add-ToPath {
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    
    if ($currentPath -notlike "*$InstallDir*") {
        Write-Info "Adding $InstallDir to PATH..."
        
        $newPath = "$currentPath;$InstallDir"
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        
        # Update current session PATH
        $env:Path = "$env:Path;$InstallDir"
        
        Write-Info "PATH updated ✓"
    }
    else {
        Write-Info "$InstallDir is already in PATH"
    }
}

function Verify-Installation {
    $installPath = Join-Path $InstallDir "$BinaryName.exe"
    
    if (-not (Test-Path $installPath)) {
        Write-Error-Custom "$BinaryName.exe not found at $installPath"
        exit 1
    }
    
    try {
        $version = & $installPath --version 2>&1
        Write-Info "Installed version: $version"
    }
    catch {
        Write-Warn "Could not verify version: $_"
    }
}

function Main {
    Write-Host ""
    Write-Host "╔═══════════════════════════════════════╗"
    Write-Host "║   SandwichLab ahcli Installer        ║"
    Write-Host "╚═══════════════════════════════════════╝"
    Write-Host ""
    
    Write-Info "Detected platform: Windows (amd64)"
    
    $version = Get-LatestVersion
    $binaryFile = Download-Binary -Version $version
    Verify-Checksum -BinaryFile $binaryFile -Version $version
    Install-Binary -BinaryFile $binaryFile
    Add-ToPath
    Verify-Installation
    
    Write-Host ""
    Write-Info "🎉 Installation successful!"
    Write-Host ""
    Write-Host "Get started with:"
    Write-Host "  $BinaryName auth login"
    Write-Host ""
    Write-Host "Note: You may need to restart your terminal for PATH changes to take effect."
    Write-Host ""
    Write-Host "For more information, visit:"
    Write-Host "  https://github.com/$Repo"
    Write-Host ""
}

Main
