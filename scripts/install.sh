#!/usr/bin/env bash
set -e

REPO="sarner/envbox"
GH_REPO="github.com/${REPO}"
INSTALL_DIR="/usr/local/bin"
TEMP_DIR=""

detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        *)       echo "unsupported" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   echo "amd64" ;;
        aarch64|arm64)  echo "arm64" ;;
        armv7l|armhf)   echo "armv7" ;;
        *)              echo "unsupported" ;;
    esac
}

download() {
    local url="$1"
    local dest="$2"
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL "$url" -o "$dest"
    elif command -v wget >/dev/null 2>&1; then
        wget -q "$url" -O "$dest"
    else
        echo "Error: neither curl nor wget found"
        exit 1
    fi
}

cleanup() {
    if [ -n "$TEMP_DIR" ] && [ -d "$TEMP_DIR" ]; then
        rm -rf "$TEMP_DIR"
    fi
}

trap cleanup EXIT

echo "Installing envbox..."

OS=$(detect_os)
ARCH=$(detect_arch)

if [ "$OS" = "unsupported" ]; then
    echo "Error: unsupported operating system"
    exit 1
fi

if [ "$ARCH" = "unsupported" ]; then
    echo "Error: unsupported architecture: $(uname -m)"
    exit 1
fi

VERSION_URL="https://api.github.com/repos/${REPO}/releases/latest"
VERSION=$(download "$VERSION_URL" - | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
    echo "Error: could not fetch latest release"
    exit 1
fi

FILENAME="envbox_${OS}_${ARCH}.tar.gz"
DOWNLOAD_URL="https://${GH_REPO}/releases/download/${VERSION}/${FILENAME}"

echo "Downloading ${FILENAME}..."

TEMP_DIR=$(mktemp -d)
ARCHIVE="${TEMP_DIR}/envbox.tar.gz"

download "$DOWNLOAD_URL" "$ARCHIVE"

cd "$TEMP_DIR"
tar -xzf "$ARCHIVE"
chmod +x envbox

if [ -w "$INSTALL_DIR" ]; then
    mv envbox "$INSTALL_DIR/"
    echo "Installed to ${INSTALL_DIR}/envbox"
else
    echo "Elevation required..."
    sudo mv envbox "$INSTALL_DIR/"
    echo "Installed to ${INSTALL_DIR}/envbox"
fi

if command -v envbox >/dev/null 2>&1; then
    echo "Successfully installed envbox $(envbox --version)"
else
    echo "Please ensure ${INSTALL_DIR} is in your PATH"
fi
