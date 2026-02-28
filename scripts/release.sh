#!/usr/bin/env bash
# release.sh — bumps the version tag and optionally pushes to trigger CI.
#
# Usage:
#   scripts/release.sh patch          # v1.2.3 → v1.2.4
#   scripts/release.sh minor          # v1.2.3 → v1.3.0
#   scripts/release.sh major          # v1.2.3 → v2.0.0
#   scripts/release.sh beta           # v1.2.3 → v1.3.0-beta.1  |  v1.3.0-beta.1 → v1.3.0-beta.2
#   scripts/release.sh rc             # v1.3.0-beta.2 → v1.3.0-rc.1  |  v1.3.0-rc.1 → v1.3.0-rc.2
#   scripts/release.sh v1.5.0         # explicit stable version
#   scripts/release.sh v1.5.0-beta.1  # explicit pre-release version
#
# Tip: add a local git alias with:
#   git config alias.release '!bash scripts/release.sh'
# Then use: git release patch
set -euo pipefail

usage() {
    echo "Usage: $(basename "$0") <major|minor|patch|beta|rc|vX.Y.Z[-pre.N]>" >&2
    exit 1
}

[[ $# -ne 1 ]] && usage

# Ensure clean working tree.
if [[ -n "$(git status --porcelain)" ]]; then
    echo "error: working tree is not clean — commit or stash changes first" >&2
    exit 1
fi

# Ensure on main branch.
branch="$(git rev-parse --abbrev-ref HEAD)"
if [[ "$branch" != "main" ]]; then
    echo "error: releases must be cut from main (current branch: $branch)" >&2
    exit 1
fi

# Fetch latest tags from origin so we don't miss anything.
git fetch --tags --quiet

# Latest stable tag (no pre-release suffix).
latest_stable="$(git tag --list "v*" --sort=-version:refname | grep -v -- '-' | head -n1)"
latest_stable="${latest_stable:-v0.0.0}"

# Latest tag overall (may be a pre-release).
latest="$(git tag --list "v*" --sort=-version:refname | head -n1)"
latest="${latest:-v0.0.0}"

# Compute the new version.
case "$1" in
    major | minor | patch)
        # Stable bumps always operate on the latest stable tag, not a pre-release.
        version="${latest_stable#v}"
        IFS='.' read -r major minor patch_num <<< "$version"
        case "$1" in
            major) ((major++)); minor=0; patch_num=0 ;;
            minor) ((minor++));           patch_num=0 ;;
            patch) ((patch_num++))                    ;;
        esac
        new_version="v${major}.${minor}.${patch_num}"
        ;;
    beta | rc)
        pre_type="$1"
        if [[ "$latest" =~ ^v([0-9]+)\.([0-9]+)\.([0-9]+)-([a-z]+)\.([0-9]+)$ ]]; then
            # Latest is a pre-release — reuse the same base version.
            base_major="${BASH_REMATCH[1]}"
            base_minor="${BASH_REMATCH[2]}"
            base_patch="${BASH_REMATCH[3]}"
            existing_type="${BASH_REMATCH[4]}"
            pre_num="${BASH_REMATCH[5]}"
            if [[ "$existing_type" == "$pre_type" ]]; then
                # Same type: increment the pre-release number.
                ((pre_num++))
            else
                # Different type (e.g. beta → rc): same base, reset to 1.
                pre_num=1
            fi
            new_version="v${base_major}.${base_minor}.${base_patch}-${pre_type}.${pre_num}"
        else
            # Latest is stable: start a new pre-release series off the next minor.
            version="${latest_stable#v}"
            IFS='.' read -r major minor patch_num <<< "$version"
            ((minor++)); patch_num=0
            new_version="v${major}.${minor}.${patch_num}-${pre_type}.1"
        fi
        ;;
    v[0-9]*)
        new_version="$1"
        if ! [[ "$new_version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-z]+\.[0-9]+)?$ ]]; then
            echo "error: invalid version format '$new_version' — expected vX.Y.Z or vX.Y.Z-pre.N" >&2
            exit 1
        fi
        ;;
    *)
        usage
        ;;
esac

echo "Latest stable   : $latest_stable"
[[ "$latest" != "$latest_stable" ]] && echo "Latest pre-release: $latest"
echo "New version     : $new_version"
echo ""

read -rp "Create annotated tag $new_version? [y/N] " confirm
[[ "$confirm" != [yY] ]] && { echo "Aborted."; exit 0; }

git tag -a "$new_version" -m "Release $new_version"
echo "Created tag $new_version"
echo ""

read -rp "Push $new_version to origin and trigger release? [y/N] " push_confirm
if [[ "$push_confirm" == [yY] ]]; then
    git push origin "$new_version"
    echo ""
    echo "Tag pushed — release workflow is running:"
    echo "  https://github.com/thismarioperez/squrl/actions"
else
    echo ""
    echo "Tag created locally. Push when ready with:"
    echo "  git push origin $new_version"
fi
