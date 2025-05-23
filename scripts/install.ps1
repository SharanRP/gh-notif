# gh-notif installation script for Windows
# This script installs the latest version of gh-notif

param(
    [string]$InstallDir = "$env:LOCALAPPDATA\gh-notif",
    [switch]$AddToPath,
    [switch]$Force
)

# Configuration
$Repo = "SharanRP/gh-notif"
$BinaryName = "gh-notif.exe"
$TempDir = [System.IO.Path]::GetTempPath() + [System.Guid]::NewGuid().ToString()

# Functions
function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Blue
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
    exit 1
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor Green
}

# Detect architecture
function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "x86_64" }
        "ARM64" { return "arm64" }
        default { Write-Error "Unsupported architecture: $arch" }
    }
}

# Get the latest release version
function Get-LatestVersion {
    try {
        $response = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
        return $response.tag_name
    }
    catch {
        Write-Error "Failed to get latest version: $_"
    }
}

# Download file
function Download-File {
    param(
        [string]$Url,
        [string]$OutputPath
    )

    try {
        Write-Info "Downloading from: $Url"
        Invoke-WebRequest -Uri $Url -OutFile $OutputPath -UseBasicParsing
    }
    catch {
        Write-Error "Failed to download file: $_"
    }
}

# Extract archive
function Extract-Archive {
    param(
        [string]$ArchivePath,
        [string]$DestinationPath
    )

    try {
        Write-Info "Extracting archive..."
        Expand-Archive -Path $ArchivePath -DestinationPath $DestinationPath -Force
    }
    catch {
        Write-Error "Failed to extract archive: $_"
    }
}

# Add to PATH
function Add-ToPath {
    param([string]$Directory)

    $currentPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    if ($currentPath -notlike "*$Directory*") {
        Write-Info "Adding $Directory to PATH..."
        $newPath = "$currentPath;$Directory"
        [Environment]::SetEnvironmentVariable("PATH", $newPath, "User")
        Write-Success "Added to PATH. Restart your terminal to use gh-notif from anywhere."
    }
    else {
        Write-Info "$Directory is already in PATH"
    }
}

# Main installation function
function Install-GhNotif {
    $arch = Get-Architecture
    $version = Get-LatestVersion
    $platform = "Windows_$arch"

    Write-Info "Detected platform: $platform"
    Write-Info "Latest version: $version"

    # Create temp directory
    New-Item -ItemType Directory -Path $TempDir -Force | Out-Null

    # Construct download URL
    $filename = "gh-notif_${platform}.zip"
    $downloadUrl = "https://github.com/$Repo/releases/download/$version/$filename"
    $archivePath = Join-Path $TempDir $filename

    # Download
    Download-File -Url $downloadUrl -OutputPath $archivePath

    # Extract
    Extract-Archive -ArchivePath $archivePath -DestinationPath $TempDir

    # Find the binary
    $binaryPath = Get-ChildItem -Path $TempDir -Name $BinaryName -Recurse | Select-Object -First 1
    if (-not $binaryPath) {
        Write-Error "Binary not found in archive"
    }

    $fullBinaryPath = Join-Path $TempDir $binaryPath

    # Create install directory
    if (-not (Test-Path $InstallDir)) {
        Write-Info "Creating install directory: $InstallDir"
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    }

    # Check if binary already exists
    $targetPath = Join-Path $InstallDir $BinaryName
    if ((Test-Path $targetPath) -and -not $Force) {
        $response = Read-Host "gh-notif is already installed. Overwrite? (y/N)"
        if ($response -ne "y" -and $response -ne "Y") {
            Write-Info "Installation cancelled"
            return
        }
    }

    # Copy binary
    Write-Info "Installing to $InstallDir"
    Copy-Item -Path $fullBinaryPath -Destination $targetPath -Force

    Write-Success "gh-notif installed successfully!"

    # Add to PATH if requested
    if ($AddToPath) {
        Add-ToPath -Directory $InstallDir
    }

    # Verify installation
    try {
        $versionOutput = & $targetPath --version
        Write-Info "Verifying installation..."
        Write-Host $versionOutput
    }
    catch {
        Write-Warn "Installation completed but verification failed: $_"
    }
}

# Cleanup function
function Cleanup {
    if (Test-Path $TempDir) {
        Remove-Item -Path $TempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}

# Main execution
try {
    Write-Info "Installing gh-notif..."

    # Check PowerShell version
    if ($PSVersionTable.PSVersion.Major -lt 5) {
        Write-Error "PowerShell 5.0 or later is required"
    }

    # Install
    Install-GhNotif

    # Show next steps
    Write-Host ""
    Write-Info "Next steps:"
    Write-Host "  1. Run 'gh-notif --help' to see available commands"
    Write-Host "  2. Run 'gh-notif auth login' to authenticate with GitHub"
    Write-Host "  3. Run 'gh-notif tutorial' for an interactive tutorial"
    Write-Host "  4. Run 'gh-notif wizard' to configure your preferences"
    Write-Host ""

    if (-not $AddToPath) {
        Write-Info "To use gh-notif from anywhere, add it to your PATH:"
        Write-Host "  Add-ToPath '$InstallDir'"
        Write-Host "  Or run this script with -AddToPath"
    }

    Write-Host ""
    Write-Info "For more information, visit: https://github.com/$Repo"
}
finally {
    Cleanup
}
