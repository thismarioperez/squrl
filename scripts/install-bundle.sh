#!/usr/bin/env bash
# install-bundle.sh — installs Squrl.app into /Applications.
set -euo pipefail

echo "Removing existing app (if it exists)..."
rm -rf /Applications/Squrl.app

echo "Installing Squrl.app..."
mv bin/Squrl.app /Applications/

echo "Fixing permissions..."
chmod -R u+w /Applications/Squrl.app
xattr -dr com.apple.quarantine /Applications/Squrl.app

echo "Done. Squrl.app installed at /Applications/Squrl.app"