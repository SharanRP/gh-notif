#!/bin/bash

# gh-notif installation script
# This script installs the latest version of gh-notif

set -e

# Configuration
REPO="user/gh-notif"
BINARY_NAME="gh-notif"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
TEMP_DIR=$(mktemp -d)

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Functions
log() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

# Detect OS and architecture
detect_platform() {
    local os arch
    
    case "$(uname -s)" in
        Linux*)     os="Linux" ;;
        Darwin*)    os="Darwin" ;;
        CYGWIN*|MINGW*|MSYS*) os="Windows" ;;
        *)          error "Unsupported operating system: $(uname -s)" ;;
    esac
    
    case "$(uname -m)" in
        x86_64|amd64)   arch="x86_64" ;;
        arm64|aarch64)  arch="arm64" ;;
        *)              error "Unsupported architecture: $(uname -m)" ;;
    esac
    
    echo "${os}_${arch}"
}

# Get the latest release version
get_latest_version() {
    local version
    version=$(curl -s "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$version" ]; then
        error "Failed to get latest version"
    fi
    
    echo "$version"
}

# Download and install
install_gh_notif() {
    local platform version download_url filename
    
    platform=$(detect_platform)
    version=$(get_latest_version)
    
    log "Detected platform: $platform"
    log "Latest version: $version"
    
    # Construct download URL
    if [[ "$platform" == "Windows"* ]]; then
        filename="${BINARY_NAME}_${platform}.zip"
    else
        filename="${BINARY_NAME}_${platform}.tar.gz"
    fi
    
    download_url="https://github.com/${REPO}/releases/download/${version}/${filename}"
    
    log "Downloading from: $download_url"
    
    # Download
    cd "$TEMP_DIR"
    if ! curl -L -o "$filename" "$download_url"; then
        error "Failed to download $filename"
    fi
    
    # Extract
    log "Extracting $filename"
    if [[ "$filename" == *.zip ]]; then
        unzip -q "$filename"
    else
        tar -xzf "$filename"
    fi
    
    # Find the binary
    binary_path=$(find . -name "$BINARY_NAME" -type f | head -1)
    if [ -z "$binary_path" ]; then
        # Try with .exe extension for Windows
        binary_path=$(find . -name "${BINARY_NAME}.exe" -type f | head -1)
    fi
    
    if [ -z "$binary_path" ]; then
        error "Binary not found in archive"
    fi
    
    # Install
    log "Installing to $INSTALL_DIR"
    
    # Create install directory if it doesn't exist
    if [ ! -d "$INSTALL_DIR" ]; then
        if ! mkdir -p "$INSTALL_DIR"; then
            error "Failed to create install directory: $INSTALL_DIR"
        fi
    fi
    
    # Copy binary
    if ! cp "$binary_path" "$INSTALL_DIR/"; then
        error "Failed to install binary to $INSTALL_DIR"
    fi
    
    # Make executable
    chmod +x "$INSTALL_DIR/$BINARY_NAME"
    
    success "gh-notif installed successfully!"
    
    # Verify installation
    if command -v "$BINARY_NAME" >/dev/null 2>&1; then
        log "Verifying installation..."
        "$BINARY_NAME" --version
    else
        warn "gh-notif was installed to $INSTALL_DIR but is not in your PATH"
        warn "Add $INSTALL_DIR to your PATH to use gh-notif from anywhere"
    fi
}

# Cleanup
cleanup() {
    if [ -d "$TEMP_DIR" ]; then
        rm -rf "$TEMP_DIR"
    fi
}

# Main
main() {
    log "Installing gh-notif..."
    
    # Check dependencies
    if ! command -v curl >/dev/null 2>&1; then
        error "curl is required but not installed"
    fi
    
    if ! command -v tar >/dev/null 2>&1 && ! command -v unzip >/dev/null 2>&1; then
        error "tar or unzip is required but not installed"
    fi
    
    # Install
    install_gh_notif
    
    # Show next steps
    echo
    log "Next steps:"
    echo "  1. Run 'gh-notif --help' to see available commands"
    echo "  2. Run 'gh-notif auth login' to authenticate with GitHub"
    echo "  3. Run 'gh-notif tutorial' for an interactive tutorial"
    echo "  4. Run 'gh-notif wizard' to configure your preferences"
    echo
    log "For more information, visit: https://github.com/${REPO}"
}

# Trap cleanup
trap cleanup EXIT

# Run main function
main "$@"
