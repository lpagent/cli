#!/bin/bash
set -euo pipefail

REPO="lpagent/cli"
BINARY="lpagent"
INSTALL_DIR="${HOME}/.local/bin"

# Colors
GREEN='\033[32m'
YELLOW='\033[33m'
CYAN='\033[36m'
RESET='\033[0m'

echo ""

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64|amd64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *) echo "Unsupported architecture: $ARCH"; exit 1 ;;
esac

case "$OS" in
  darwin|linux) ;;
  *) echo "Unsupported OS: $OS"; exit 1 ;;
esac

# Get latest release tag
LATEST=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$LATEST" ]; then
  echo "Failed to fetch latest release"
  exit 1
fi

VERSION="${LATEST#v}"
ARCHIVE="cli_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${LATEST}/${ARCHIVE}"

echo -e "  → Downloading ${BINARY} ${CYAN}${LATEST}${RESET} for ${OS}_${ARCH}..."

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

curl -fsSL "$URL" -o "${TMP}/${ARCHIVE}"

echo "  → Extracting..."
tar xzf "${TMP}/${ARCHIVE}" -C "$TMP"

# Ensure install directory exists
mkdir -p "$INSTALL_DIR"
mv "${TMP}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
chmod +x "${INSTALL_DIR}/${BINARY}"

echo -e "  ${GREEN}✓${RESET} Installed ${BINARY} ${LATEST} to ${INSTALL_DIR}/${BINARY}"

# Check if INSTALL_DIR is in PATH
if ! echo "$PATH" | tr ':' '\n' | grep -qx "$INSTALL_DIR"; then
  echo ""
  echo -e "  ${YELLOW}⚠${RESET} ${INSTALL_DIR} is not in your PATH. Add it with:"
  SHELL_NAME=$(basename "$SHELL")
  case "$SHELL_NAME" in
    zsh)  echo "    echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.zshrc && source ~/.zshrc" ;;
    bash) echo "    echo 'export PATH=\"\$HOME/.local/bin:\$PATH\"' >> ~/.bashrc && source ~/.bashrc" ;;
    *)    echo "    export PATH=\"\$HOME/.local/bin:\$PATH\"" ;;
  esac
fi

echo ""
echo "  Next steps:"
echo -e "    ${CYAN}lpagent auth set-key${RESET}                          Set your API key"
echo -e "    ${CYAN}lpagent auth set-default-owner <wallet>${RESET}       Set default wallet"
echo -e "    ${CYAN}lpagent positions open -o table --native${RESET}      View open positions"
echo ""
