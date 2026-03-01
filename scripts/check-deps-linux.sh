#!/usr/bin/env bash
# check-deps-linux.sh — verifies required system dependencies for Linux builds.
# Exits 0 silently on non-Linux systems.
set -euo pipefail

[[ "$(uname -s)" == "Linux" ]] || exit 0

MISSING_PKGS=()
MISSING_LABELS=()

check_pkgconfig() {
    local pc_name="$1"
    local apt_pkg="$2"
    local label="$3"
    if ! pkg-config --exists "$pc_name" 2>/dev/null; then
        MISSING_PKGS+=("$apt_pkg")
        MISSING_LABELS+=("$apt_pkg  ($label)")
    fi
}

check_command() {
    local cmd="$1"
    local apt_pkg="$2"
    local label="$3"
    if ! command -v "$cmd" &>/dev/null; then
        MISSING_PKGS+=("$apt_pkg")
        MISSING_LABELS+=("$apt_pkg  ($label)")
    fi
}

# pkg-config itself is needed to locate C libraries at compile time.
if ! command -v pkg-config &>/dev/null; then
    echo "error: pkg-config not found. Install it with:"
    echo "  sudo apt-get install -y pkg-config"
    exit 1
fi

# CGO build-time: systray requires ayatana-appindicator3.
check_pkgconfig "ayatana-appindicator3-0.1" "libayatana-appindicator3-dev" \
    "system tray — github.com/getlantern/systray"

# Runtime: desktop notifications via notify-send.
check_command "notify-send" "libnotify-bin" \
    "desktop notifications"

# Runtime: clipboard support via xclip or xsel.
if ! command -v xclip &>/dev/null && ! command -v xsel &>/dev/null; then
    MISSING_PKGS+=("xclip")
    MISSING_LABELS+=("xclip  (clipboard support; alternatively: xsel)")
fi

if [[ ${#MISSING_PKGS[@]} -eq 0 ]]; then
    exit 0
fi

echo "error: missing Linux system dependencies required to build/run squrl:"
echo ""
for label in "${MISSING_LABELS[@]}"; do
    echo "  - $label"
done
echo ""
echo "Install with:"
echo "  sudo apt-get install -y ${MISSING_PKGS[*]}"
echo ""
exit 1
