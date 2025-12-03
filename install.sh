#!/bin/sh
# react-analyzer installer script
# Usage: curl -fsSL https://raw.githubusercontent.com/rautio/react-analyzer/main/install.sh | sh

set -e

# Configuration
REPO="rautio/react-analyzer"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"
VERSION="${VERSION:-latest}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Helper functions
info() {
    printf "${GREEN}==>${NC} %s\n" "$1"
}

error() {
    printf "${RED}Error:${NC} %s\n" "$1" >&2
    exit 1
}

warn() {
    printf "${YELLOW}Warning:${NC} %s\n" "$1"
}

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Darwin*)
            echo "darwin"
            ;;
        Linux*)
            echo "linux"
            ;;
        MINGW*|MSYS*|CYGWIN*)
            echo "windows"
            ;;
        *)
            error "Unsupported operating system: $(uname -s)"
            ;;
    esac
}

# Detect architecture
detect_arch() {
    arch="$(uname -m)"
    case "$arch" in
        x86_64|amd64)
            echo "amd64"
            ;;
        aarch64|arm64)
            echo "arm64"
            ;;
        *)
            error "Unsupported architecture: $arch"
            ;;
    esac
}

# Get latest release version from GitHub API
get_latest_version() {
    if command -v curl > /dev/null 2>&1; then
        curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
    else
        error "curl is required but not installed"
    fi
}

# Download file
download() {
    url="$1"
    output="$2"

    if command -v curl > /dev/null 2>&1; then
        curl -fsSL "$url" -o "$output"
    elif command -v wget > /dev/null 2>&1; then
        wget -q "$url" -O "$output"
    else
        error "Neither curl nor wget is installed"
    fi
}

# Main installation
main() {
    info "Installing react-analyzer..."

    # Detect system
    OS=$(detect_os)
    ARCH=$(detect_arch)
    info "Detected platform: ${OS}/${ARCH}"

    # Get version
    if [ "$VERSION" = "latest" ]; then
        VERSION=$(get_latest_version)
        if [ -z "$VERSION" ]; then
            error "Failed to fetch latest version"
        fi
        info "Latest version: ${VERSION}"
    fi

    # Remove 'v' prefix if present
    VERSION_NUM="${VERSION#v}"

    # Determine file extension
    if [ "$OS" = "windows" ]; then
        EXT="zip"
        BINARY="react-analyzer.exe"
    else
        EXT="tar.gz"
        BINARY="react-analyzer"
    fi

    # Construct download URL
    FILENAME="react-analyzer_${VERSION_NUM}_${OS}-${ARCH}.${EXT}"
    DOWNLOAD_URL="https://github.com/${REPO}/releases/download/${VERSION}/${FILENAME}"

    info "Downloading from: ${DOWNLOAD_URL}"

    # Create temp directory
    TMP_DIR=$(mktemp -d)
    trap "rm -rf $TMP_DIR" EXIT

    # Download archive
    download "$DOWNLOAD_URL" "${TMP_DIR}/${FILENAME}" || error "Failed to download ${FILENAME}"

    # Extract archive
    info "Extracting archive..."
    cd "$TMP_DIR"

    if [ "$EXT" = "zip" ]; then
        if command -v unzip > /dev/null 2>&1; then
            unzip -q "$FILENAME" || error "Failed to extract archive"
        else
            error "unzip is required but not installed"
        fi
    else
        tar -xzf "$FILENAME" || error "Failed to extract archive"
    fi

    # Find binary
    if [ ! -f "$BINARY" ] && [ ! -f "react-analyzer/$BINARY" ]; then
        error "Binary not found in archive"
    fi

    # Move binary to install directory (handle both archive structures)
    if [ -f "react-analyzer/$BINARY" ]; then
        BINARY_PATH="react-analyzer/$BINARY"
    else
        BINARY_PATH="$BINARY"
    fi

    info "Installing to ${INSTALL_DIR}..."

    # Check if we need sudo
    if [ -w "$INSTALL_DIR" ]; then
        mv "$BINARY_PATH" "${INSTALL_DIR}/${BINARY}"
        chmod +x "${INSTALL_DIR}/${BINARY}"
    else
        info "Installing to ${INSTALL_DIR} requires elevated privileges"
        sudo mv "$BINARY_PATH" "${INSTALL_DIR}/${BINARY}"
        sudo chmod +x "${INSTALL_DIR}/${BINARY}"
    fi

    # Verify installation
    if command -v react-analyzer > /dev/null 2>&1; then
        INSTALLED_VERSION=$(react-analyzer --version | awk '{print $2}')
        info "Successfully installed react-analyzer ${INSTALLED_VERSION}"
        info "Run 'react-analyzer --help' to get started"
    else
        warn "Installation complete, but ${INSTALL_DIR} may not be in your PATH"
        info "Add it to your PATH with: export PATH=\"${INSTALL_DIR}:\$PATH\""
        info "Then run 'react-analyzer --help' to get started"
    fi
}

main "$@"
