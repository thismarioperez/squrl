#!/usr/bin/env bash
# make-icns.sh — converts assets/icon.svg into assets/AppIcon.icns for macOS .app bundles.
# Requires: rsvg-convert (brew install librsvg), iconutil (built-in macOS)
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(dirname "$SCRIPT_DIR")"
SVG="$REPO_ROOT/assets/icon.svg"
ICONSET="$REPO_ROOT/assets/AppIcon.iconset"
ICNS="$REPO_ROOT/assets/AppIcon.icns"

if ! command -v rsvg-convert &>/dev/null; then
    echo "Error: rsvg-convert not found. Install with: brew install librsvg" >&2
    exit 1
fi

rm -rf "$ICONSET"
mkdir -p "$ICONSET"

render() {
    local size=$1
    local name=$2
    rsvg-convert -w "$size" -h "$size" "$SVG" -o "$ICONSET/$name"
}

render 16    icon_16x16.png
render 32    icon_16x16@2x.png
render 32    icon_32x32.png
render 64    icon_32x32@2x.png
render 128   icon_128x128.png
render 256   icon_128x128@2x.png
render 256   icon_256x256.png
render 512   icon_256x256@2x.png
render 512   icon_512x512.png
render 1024  icon_512x512@2x.png

iconutil -c icns "$ICONSET" -o "$ICNS"
rm -rf "$ICONSET"

echo "Created: $ICNS"

# Render menu bar template icon from menubar.svg.
TRAY_SVG="$REPO_ROOT/assets/menubar.svg"
rsvg-convert -w 22 -h 22 "$TRAY_SVG" -o "$REPO_ROOT/assets/menubar_22.png"
rsvg-convert -w 44 -h 44 "$TRAY_SVG" -o "$REPO_ROOT/assets/menubar_44.png"
echo "Created: assets/menubar_22.png, assets/menubar_44.png"

# Render 64×64 notification icon from notification.svg (used by desktop notifications).
NOTIF_SVG="$REPO_ROOT/assets/notification.svg"
rsvg-convert -w 64 -h 64 "$NOTIF_SVG" -o "$REPO_ROOT/assets/notification_64.png"
echo "Created: assets/notification_64.png"

# Render CLI ANSI icon from cli.svg.
# Renders at native resolution then nearest-neighbour samples to 32×32 so
# each SVG pixel maps 1-to-1 to a terminal half-block without any blending.
CLI_SVG="$REPO_ROOT/assets/cli.svg"
CLI_ANSI="$REPO_ROOT/assets/cli_ansi.txt"
CLI_PNG_FULL=$(mktemp)
CLI_PNG=$(mktemp)
rsvg-convert -w 512 -h 512 "$CLI_SVG" -o "$CLI_PNG_FULL"
magick "$CLI_PNG_FULL" -sample 16x16 "PNG32:$CLI_PNG"
python3 "$SCRIPT_DIR/gen-cli-icon.py" "$CLI_PNG" > "$CLI_ANSI"
rm "$CLI_PNG_FULL" "$CLI_PNG"
echo "Created: assets/cli_ansi.txt"
