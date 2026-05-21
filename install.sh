#!/bin/sh
set -e

REPO="iamy4n-dev/dpod-seed"
BIN_NAME="dpod-seed"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

OS=$(uname -s | tr '[:upper:]' '[:lower:]')
case "$OS" in
  linux|darwin) ;;
  *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
esac

ARCH=$(uname -m)
case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  aarch64|arm64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

if [ -z "$VERSION" ]; then
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' \
    | sed 's/.*"tag_name": *"\(.*\)".*/\1/')
fi

ASSET="${BIN_NAME}_${OS}_${ARCH}"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ASSET}"

echo "Installing ${BIN_NAME} ${VERSION} (${OS}/${ARCH}) -> ${INSTALL_DIR}/${BIN_NAME}"
curl -fsSL "$URL" -o "/tmp/${BIN_NAME}"
chmod +x "/tmp/${BIN_NAME}"
mv "/tmp/${BIN_NAME}" "${INSTALL_DIR}/${BIN_NAME}"
echo "Done. Run: ${BIN_NAME} --version"
