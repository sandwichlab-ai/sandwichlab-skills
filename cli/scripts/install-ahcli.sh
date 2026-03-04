#!/bin/bash
# SandwichLab ahcli Installation Script
# Usage: curl -fsSL https://raw.githubusercontent.com/sandwichlab-ai/sandwichlab-skills/main/scripts/install-ahcli.sh | bash

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
REPO="sandwichlab-ai/sandwichlab-skills"
BINARY_NAME="ahcli"
INSTALL_DIR="/usr/local/bin"

# Functions
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case "$os" in
        linux*)
            OS="linux"
            ;;
        darwin*)
            OS="darwin"
            ;;
        *)
            print_error "Unsupported operating system: $os"
            exit 1
            ;;
    esac
    
    case "$arch" in
        x86_64|amd64)
            ARCH="amd64"
            ;;
        arm64|aarch64)
            ARCH="arm64"
            ;;
        *)
            print_error "Unsupported architecture: $arch"
            exit 1
            ;;
    esac
    
    print_info "Detected platform: ${OS}-${ARCH}"
}

get_latest_version() {
    print_info "Fetching latest version..."
    VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/')
    
    if [ -z "$VERSION" ]; then
        print_error "Failed to fetch latest version"
        exit 1
    fi
    
    print_info "Latest version: ${VERSION}"
}

download_binary() {
    local binary_name="${BINARY_NAME}-${OS}-${ARCH}"
    local download_url="https://github.com/${REPO}/releases/download/${VERSION}/${binary_name}"
    local temp_file="/tmp/${binary_name}"
    
    print_info "Downloading ${binary_name}..."
    
    if ! curl -fsSL -o "${temp_file}" "${download_url}"; then
        print_error "Failed to download binary from ${download_url}"
        exit 1
    fi
    
    print_info "Download complete"
    echo "${temp_file}"
}

verify_checksum() {
    local binary_file="$1"
    local checksums_url="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"
    
    print_info "Verifying checksum..."
    
    if ! command -v sha256sum &> /dev/null; then
        print_warn "sha256sum not found, skipping checksum verification"
        return 0
    fi
    
    local checksums=$(curl -fsSL "${checksums_url}")
    local binary_name=$(basename "${binary_file}")
    local expected_checksum=$(echo "${checksums}" | grep "${binary_name}" | awk '{print $1}')
    
    if [ -z "${expected_checksum}" ]; then
        print_warn "Checksum not found for ${binary_name}, skipping verification"
        return 0
    fi
    
    local actual_checksum=$(sha256sum "${binary_file}" | awk '{print $1}')
    
    if [ "${expected_checksum}" != "${actual_checksum}" ]; then
        print_error "Checksum verification failed!"
        print_error "Expected: ${expected_checksum}"
        print_error "Actual:   ${actual_checksum}"
        exit 1
    fi
    
    print_info "Checksum verified ✓"
}

install_binary() {
    local binary_file="$1"
    local install_path="${INSTALL_DIR}/${BINARY_NAME}"
    
    print_info "Installing to ${install_path}..."
    
    # Check if we need sudo
    if [ -w "${INSTALL_DIR}" ]; then
        mv "${binary_file}" "${install_path}"
        chmod +x "${install_path}"
    else
        print_info "Requesting sudo access for installation..."
        sudo mv "${binary_file}" "${install_path}"
        sudo chmod +x "${install_path}"
    fi
    
    print_info "Installation complete ✓"
}

verify_installation() {
    if ! command -v ${BINARY_NAME} &> /dev/null; then
        print_error "${BINARY_NAME} not found in PATH"
        print_info "Please add ${INSTALL_DIR} to your PATH"
        exit 1
    fi
    
    local installed_version=$(${BINARY_NAME} --version 2>&1 || echo "unknown")
    print_info "Installed version: ${installed_version}"
}

main() {
    echo ""
    echo "╔═══════════════════════════════════════╗"
    echo "║   SandwichLab ahcli Installer        ║"
    echo "╚═══════════════════════════════════════╝"
    echo ""
    
    detect_platform
    get_latest_version
    
    local binary_file=$(download_binary)
    verify_checksum "${binary_file}"
    install_binary "${binary_file}"
    verify_installation
    
    echo ""
    print_info "🎉 Installation successful!"
    echo ""
    echo "Get started with:"
    echo "  ${BINARY_NAME} auth login"
    echo ""
    echo "For more information, visit:"
    echo "  https://github.com/${REPO}"
    echo ""
}

main
