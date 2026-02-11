#!/usr/bin/env sh
set -eu

REPO="ZiaoLiu-1/pskill"
VERSION="${VERSION:-latest}"

OS="$(uname | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64) ARCH="amd64" ;;
  arm64|aarch64) ARCH="arm64" ;;
esac

if [ "$VERSION" = "latest" ]; then
  VERSION="$(curl -fsSL "https://api.github.com/repos/$REPO/releases/latest" | sed -n 's/.*"tag_name": "\(.*\)".*/\1/p' | head -n 1)"
fi

ASSET="pskill_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/$VERSION/$ASSET"

TMP="$(mktemp -d)"
curl -fsSL "$URL" -o "$TMP/pskill.tgz"
tar -xzf "$TMP/pskill.tgz" -C "$TMP"
install "$TMP/pskill" /usr/local/bin/pskill
echo "Installed pskill to /usr/local/bin/pskill"
