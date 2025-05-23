# Installation Guide

gh-notif provides multiple installation methods to suit different platforms and preferences. Choose the method that works best for your environment.

## 🚀 Quick Install (Recommended)

### Unix/Linux/macOS
```bash
curl -fsSL https://raw.githubusercontent.com/SharanRP/gh-notif/main/scripts/install.sh | bash
```

### Windows (PowerShell)
```powershell
iwr https://raw.githubusercontent.com/SharanRP/gh-notif/main/scripts/install.ps1 | iex
```

## 📦 Package Managers

### Homebrew (macOS/Linux)
```bash
# Add the tap
brew tap SharanRP/homebrew-tap

# Install gh-notif
brew install gh-notif

# Update
brew upgrade gh-notif
```

### Scoop (Windows)
```powershell
# Add the bucket
scoop bucket add SharanRP https://github.com/SharanRP/scoop-bucket

# Install gh-notif
scoop install gh-notif

# Update
scoop update gh-notif
```

### Chocolatey (Windows)
```powershell
# Install gh-notif
choco install gh-notif

# Update
choco upgrade gh-notif
```

### Snap (Linux)
```bash
# Install gh-notif
sudo snap install gh-notif

# Update
sudo snap refresh gh-notif
```

### APT (Debian/Ubuntu)
```bash
# Add repository
curl -fsSL https://github.com/SharanRP/gh-notif/releases/latest/download/gh-notif_amd64.deb -o gh-notif.deb

# Install
sudo dpkg -i gh-notif.deb

# Fix dependencies if needed
sudo apt-get install -f
```

### YUM/DNF (RHEL/Fedora/CentOS)
```bash
# Download RPM
curl -fsSL https://github.com/SharanRP/gh-notif/releases/latest/download/gh-notif-1.0.0-1.x86_64.rpm -o gh-notif.rpm

# Install with YUM
sudo yum localinstall gh-notif.rpm

# Or install with DNF
sudo dnf install gh-notif.rpm
```

### AUR (Arch Linux)
```bash
# Using yay
yay -S gh-notif

# Using paru
paru -S gh-notif

# Manual installation
git clone https://aur.archlinux.org/gh-notif.git
cd gh-notif
makepkg -si
```

## 🐳 Container/Docker

### Docker Hub
```bash
# Run directly
docker run --rm -it ghcr.io/SharanRP/gh-notif:latest --help

# With volume for config persistence
docker run --rm -it -v ~/.gh-notif:/home/gh-notif/.gh-notif ghcr.io/SharanRP/gh-notif:latest

# Create alias for easy use
alias gh-notif='docker run --rm -it -v ~/.gh-notif:/home/gh-notif/.gh-notif ghcr.io/SharanRP/gh-notif:latest'
```

### GitHub Container Registry
```bash
# Pull latest image
docker pull ghcr.io/SharanRP/gh-notif:latest

# Pull specific version
docker pull ghcr.io/SharanRP/gh-notif:v1.0.0

# Run with authentication
docker run --rm -it \
  -v ~/.gh-notif:/home/gh-notif/.gh-notif \
  -e GITHUB_TOKEN=$GITHUB_TOKEN \
  ghcr.io/SharanRP/gh-notif:latest list
```

## 📥 Direct Download

### GitHub Releases
1. Go to [GitHub Releases](https://github.com/SharanRP/gh-notif/releases)
2. Download the appropriate binary for your platform:
   - **Linux**: `gh-notif_Linux_x86_64.tar.gz`
   - **macOS**: `gh-notif_Darwin_x86_64.tar.gz` (Intel) or `gh-notif_Darwin_arm64.tar.gz` (Apple Silicon)
   - **Windows**: `gh-notif_Windows_x86_64.zip`

3. Extract and install:

#### Linux/macOS
```bash
# Extract
tar -xzf gh-notif_*.tar.gz

# Make executable
chmod +x gh-notif

# Move to PATH
sudo mv gh-notif /usr/local/bin/

# Verify installation
gh-notif --version
```

#### Windows
```powershell
# Extract the ZIP file
Expand-Archive gh-notif_Windows_x86_64.zip

# Move to a directory in your PATH
Move-Item gh-notif.exe C:\Windows\System32\

# Or add to a custom directory and update PATH
```

## 🛠 Build from Source

### Prerequisites
- Go 1.20 or later
- Git

### Build Steps
```bash
# Clone the repository
git clone https://github.com/SharanRP/gh-notif.git
cd gh-notif

# Build
go build -o gh-notif .

# Install to system PATH
sudo mv gh-notif /usr/local/bin/

# Or install with Go
go install github.com/SharanRP/gh-notif@latest
```

### Development Build
```bash
# Clone for development
git clone https://github.com/SharanRP/gh-notif.git
cd gh-notif

# Install dependencies
go mod download

# Build with debug info
go build -ldflags="-X main.versionString=dev" -o gh-notif .

# Run tests
go test ./...

# Run with development config
./gh-notif --config dev.yaml
```

## 🔧 Post-Installation Setup

### 1. Verify Installation
```bash
gh-notif --version
gh-notif --help
```

### 2. Authentication
```bash
# Login to GitHub
gh-notif auth login

# Check authentication status
gh-notif auth status
```

### 3. Configuration
```bash
# Run first-time setup
gh-notif firstrun

# Or configure manually
gh-notif config set notifications.default_limit 50
gh-notif config set ui.theme dark
```

### 4. Shell Completions
```bash
# Bash
gh-notif completion bash | sudo tee /etc/bash_completion.d/gh-notif

# Zsh
gh-notif completion zsh > "${fpath[1]}/_gh-notif"

# Fish
gh-notif completion fish > ~/.config/fish/completions/gh-notif.fish

# PowerShell (Windows)
gh-notif completion powershell | Out-String | Invoke-Expression
```

### 5. Man Pages (Unix/Linux/macOS)
```bash
# Generate and install man pages
gh-notif man generate
sudo gh-notif man install
```

## 🔄 Updates

### Package Managers
Most package managers provide update commands:
```bash
# Homebrew
brew upgrade gh-notif

# Scoop
scoop update gh-notif

# Snap
sudo snap refresh gh-notif

# Chocolatey
choco upgrade gh-notif
```

### Manual Updates
```bash
# Check current version
gh-notif version

# Check for updates
gh-notif version --check-updates

# Download and install latest version using install script
curl -fsSL https://raw.githubusercontent.com/SharanRP/gh-notif/main/scripts/install.sh | bash
```

## 🗑 Uninstallation

### Package Managers
```bash
# Homebrew
brew uninstall gh-notif

# Scoop
scoop uninstall gh-notif

# Snap
sudo snap remove gh-notif

# APT
sudo apt remove gh-notif

# YUM/DNF
sudo yum remove gh-notif
# or
sudo dnf remove gh-notif
```

### Manual Removal
```bash
# Remove binary
sudo rm /usr/local/bin/gh-notif

# Remove configuration
rm -rf ~/.gh-notif

# Remove completions
sudo rm /etc/bash_completion.d/gh-notif
rm ~/.oh-my-zsh/completions/_gh-notif
rm ~/.config/fish/completions/gh-notif.fish

# Remove man pages
sudo rm /usr/share/man/man1/gh-notif*
```

## 🐛 Troubleshooting

### Common Issues

#### Permission Denied
```bash
# Make sure the binary is executable
chmod +x /usr/local/bin/gh-notif

# Check PATH
echo $PATH
```

#### Command Not Found
```bash
# Add to PATH in your shell profile
echo 'export PATH="/usr/local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

#### Authentication Issues
```bash
# Clear and re-authenticate
gh-notif auth logout
gh-notif auth login
```

#### Configuration Issues
```bash
# Reset configuration
gh-notif config reset

# Check configuration
gh-notif config list
```

### Getting Help
- 📖 [Documentation](https://github.com/SharanRP/gh-notif/blob/main/README.md)
- 🐛 [Issue Tracker](https://github.com/SharanRP/gh-notif/issues)
- 💬 [Discussions](https://github.com/SharanRP/gh-notif/discussions)
- 📧 [Support Email](mailto:support@gh-notif.dev)

## 📋 System Requirements

### Minimum Requirements
- **OS**: Windows 10, macOS 10.15, or Linux (kernel 3.10+)
- **Architecture**: x86_64 (amd64) or ARM64
- **Memory**: 64MB RAM
- **Disk**: 50MB free space
- **Network**: Internet connection for GitHub API access

### Recommended Requirements
- **Memory**: 128MB RAM for optimal performance
- **Disk**: 100MB free space for cache and logs
- **Terminal**: Modern terminal with Unicode support for best UI experience

### Supported Platforms
- ✅ **Linux**: Ubuntu 18.04+, Debian 10+, CentOS 7+, Fedora 30+, Arch Linux
- ✅ **macOS**: 10.15+ (Intel and Apple Silicon)
- ✅ **Windows**: Windows 10, Windows 11, Windows Server 2019+
- ✅ **Docker**: Linux containers on any platform

Choose the installation method that best fits your workflow and platform. The package manager installations are recommended for automatic updates and easy management.
