#!/bin/sh
set -e

# Saber CLI install script
# Usage: curl -sSL https://install.saber.app | sh

REPO="saberapp/cli"
BINARY="saber"
INSTALL_DIR="${INSTALL_DIR:-/usr/local/bin}"

# Detect OS and architecture
OS=$(uname -s | tr '[:upper:]' '[:lower:]')
ARCH=$(uname -m)

case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
  arm64)   ARCH="arm64" ;;
  *)
    echo "Unsupported architecture: $ARCH" >&2
    exit 1
    ;;
esac

case "$OS" in
  darwin|linux) ;;
  *)
    echo "Unsupported OS: $OS" >&2
    echo "For Windows, download from https://github.com/$REPO/releases" >&2
    exit 1
    ;;
esac

# Get latest stable CLI release version. This is a monorepo so we search recent
# releases and skip drafts and prereleases (e.g. rc, beta).
# python3 is used for reliable JSON parsing — it ships with macOS and is standard
# on most Linux developer machines.
if ! command -v python3 >/dev/null 2>&1; then
  echo "Error: python3 is required to install saber" >&2
  exit 1
fi

VERSION=$(curl -sSf "https://api.github.com/repos/$REPO/releases?per_page=20" \
  | python3 -c "
import json, sys
for r in json.load(sys.stdin):
    if r.get('draft') or r.get('prerelease'):
        continue
    tag = r.get('tag_name', '')
    if tag.startswith('v'):
        print(tag[len('v'):])
        break
")

if [ -z "$VERSION" ]; then
  echo "Failed to determine latest version" >&2
  exit 1
fi

ARCHIVE="saber_${VERSION}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/$REPO/releases/download/v${VERSION}/${ARCHIVE}"
CHECKSUM_URL="https://github.com/$REPO/releases/download/v${VERSION}/saber_${VERSION}_checksums.txt"

TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT

echo "Downloading saber v${VERSION} for ${OS}/${ARCH}..."
curl -sSfL "$URL" -o "$TMP/$ARCHIVE"
curl -sSfL "$CHECKSUM_URL" -o "$TMP/checksums.txt"

# Verify checksum (sha256sum on Linux, shasum on macOS)
cd "$TMP"
if command -v sha256sum >/dev/null 2>&1; then
  grep "$ARCHIVE" checksums.txt | sha256sum -c - || { echo "Checksum verification failed" >&2; exit 1; }
elif command -v shasum >/dev/null 2>&1; then
  grep "$ARCHIVE" checksums.txt | shasum -a 256 -c - || { echo "Checksum verification failed" >&2; exit 1; }
else
  echo "Warning: no sha256 tool found, skipping checksum verification" >&2
fi

tar -xzf "$ARCHIVE"

# Install
if [ -w "$INSTALL_DIR" ]; then
  mv "$BINARY" "$INSTALL_DIR/$BINARY"
else
  echo "Installing to $INSTALL_DIR (may require sudo)..."
  sudo mv "$BINARY" "$INSTALL_DIR/$BINARY"
fi

chmod +x "$INSTALL_DIR/$BINARY"

echo ""
echo "saber v${VERSION} installed to $INSTALL_DIR/$BINARY"
echo ""
echo "Get started:"
echo "  saber auth login"
echo "  saber --help"
