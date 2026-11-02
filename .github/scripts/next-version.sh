#!/usr/bin/env bash
# Prints the next release tag.
#
# Usage: git tag --list | next-version.sh <patch|minor>
#
# Reads tags on stdin and bumps the highest stable vX.Y.Z tag. Pre-release
# tags (v1.2.0-rc.1) and other tag names are ignored. With no stable tag at
# all, prints v0.1.0 for either bump.
set -euo pipefail

bump="${1:-}"
case "$bump" in
  patch | minor) ;;
  *)
    echo "usage: next-version.sh <patch|minor>" >&2
    exit 2
    ;;
esac

latest=$(grep -E '^v[0-9]+\.[0-9]+\.[0-9]+$' | sort -t. -k1.2,1n -k2,2n -k3,3n | tail -n 1 || true)

if [ -z "$latest" ]; then
  echo "v0.1.0"
  exit 0
fi

IFS=. read -r major minor patch <<<"${latest#v}"
case "$bump" in
  patch) patch=$((10#$patch + 1)) ;;
  minor)
    minor=$((10#$minor + 1))
    patch=0
    ;;
esac

echo "v$((10#$major)).${minor}.${patch}"
