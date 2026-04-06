#!/usr/bin/env bash
set -euo pipefail

# ── Platform detection ─────────────────────────────────────────────
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)  ARCH="amd64" ;;
  aarch64) ARCH="arm64" ;;
esac

BINARY_NAME="rdt-${OS}-${ARCH}"
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
BIN_DIR="$SCRIPT_DIR/bin"
BINARY="$BIN_DIR/$BINARY_NAME"
REPO="Huafucius/rdt-cli-go"

# ── Download or use cached binary ──────────────────────────────────
if [ ! -f "$BINARY" ]; then
  mkdir -p "$BIN_DIR"
  RELEASE_URL="https://github.com/${REPO}/releases/latest/download/${BINARY_NAME}"
  echo "Downloading rdt binary for ${OS}/${ARCH}..." >&2
  if curl -fsSL -o "$BINARY" "$RELEASE_URL"; then
    chmod +x "$BINARY"
    echo "Installed to $BINARY" >&2
  else
    echo "Error: failed to download from $RELEASE_URL" >&2
    echo "You can build from source: cd $(dirname "$SCRIPT_DIR") && go build -o $BINARY ." >&2
    exit 1
  fi
fi

# ── Run ────────────────────────────────────────────────────────────
exec "$BINARY" "$@"
