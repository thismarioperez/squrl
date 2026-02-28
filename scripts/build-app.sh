#!/usr/bin/env bash
# build-app.sh — packages the binary into a macOS .app bundle.
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
APP_NAME="Squrl"
APP_DIR="$REPO_ROOT/$APP_NAME.app"
BINARY_NAME="squrl"

echo "Building $APP_NAME.app..."

# 1. Build the binary.
cd "$REPO_ROOT"
LDFLAGS="${LDFLAGS:--X main.version=dev}"
CGO_ENABLED=1 go build -ldflags "$LDFLAGS" -o "bin/$BINARY_NAME" ./cmd/squrl/

# 2. Create bundle structure.
mkdir -p "$APP_DIR/Contents/MacOS"
mkdir -p "$APP_DIR/Contents/Resources"

# 3. Copy binary, plist, and icon.
cp "bin/$BINARY_NAME" "$APP_DIR/Contents/MacOS/$BINARY_NAME"
cp "$REPO_ROOT/Info.plist" "$APP_DIR/Contents/Info.plist"
if [ -f "$REPO_ROOT/assets/AppIcon.icns" ]; then
    cp "$REPO_ROOT/assets/AppIcon.icns" "$APP_DIR/Contents/Resources/AppIcon.icns"
fi

echo "Done: $APP_DIR"
echo ""
echo "Launch with:"
echo "  open \"$APP_DIR\""
echo ""
echo "NOTE: On first launch macOS will prompt for Screen Recording permission."
echo "      Grant it in System Settings → Privacy & Security → Screen Recording."
