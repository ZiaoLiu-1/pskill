#!/usr/bin/env sh
set -eu

REPO="ZiaoLiu-1/pskill"
VERSION="${VERSION:-latest}"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
OS="$(uname | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)        ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
  *)
    echo "Error: unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

# Verify supported OS
case "$OS" in
  darwin|linux) ;;
  *)
    echo "Error: unsupported OS: $OS (use Windows installer or npm)" >&2
    exit 1
    ;;
esac

# Resolve latest version from GitHub
if [ "$VERSION" = "latest" ]; then
  echo "Fetching latest release..."
  VERSION="$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" \
    | sed -n 's/.*"tag_name": "\(.*\)".*/\1/p' \
    | head -n 1)"
  if [ -z "$VERSION" ]; then
    echo "Error: could not determine latest version" >&2
    exit 1
  fi
fi

ASSET="pskill_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$VERSION/$ASSET"

echo "Downloading pskill $VERSION for $OS/$ARCH..."

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT

if ! curl -fsSL "$URL" -o "$TMP/pskill.tgz"; then
  echo "Error: download failed — check version and platform" >&2
  echo "  URL: $URL" >&2
  exit 1
fi

tar -xzf "$TMP/pskill.tgz" -C "$TMP"

if [ ! -f "$TMP/pskill" ]; then
  echo "Error: binary not found in archive" >&2
  exit 1
fi

# Install binary
if [ -w "$INSTALL_DIR" ]; then
  install -m 755 "$TMP/pskill" "$INSTALL_DIR/pskill"
else
  echo "Installing to $INSTALL_DIR (requires sudo)..."
  sudo install -m 755 "$TMP/pskill" "$INSTALL_DIR/pskill"
fi

echo "Installed pskill $VERSION to $INSTALL_DIR/pskill"
echo ""
echo "Run 'pskill' to start the interactive TUI."
