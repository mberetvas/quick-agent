#!/usr/bin/env bash
# Prepare a release: bump app version, update CHANGELOG, commit, tag, and push.
set -euo pipefail

VERSION="${1:-}"
if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version>"
  echo "Example: $0 v1.0.0"
  exit 1
fi

if [[ ! "$VERSION" =~ ^v[0-9]+\.[0-9]+\.[0-9]+(-[a-zA-Z0-9.]+)?$ ]]; then
  echo "Version must look like v1.0.0 (optional pre-release suffix)"
  exit 1
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
VERSION_FILE="${ROOT}/internal/version/version.go"
CHANGELOG="${ROOT}/CHANGELOG.md"
TAG_VERSION="${VERSION#v}"

if ! grep -q 'Version = "dev"' "$VERSION_FILE" && ! grep -q "Version = \"${TAG_VERSION}\"" "$VERSION_FILE"; then
  echo "Refusing to release: update internal/version/version.go manually or reset Version to dev first"
  exit 1
fi

tmp="$(mktemp)"
sed 's/Version = "[^"]*"/Version = "'"${TAG_VERSION}"'"/' "$VERSION_FILE" > "$tmp"
mv "$tmp" "$VERSION_FILE"

DATE="$(date +%Y-%m-%d)"
ENTRY="## ${VERSION} - ${DATE}

### Added
- (describe changes)

"
printf '%s' "$ENTRY" | cat - "$CHANGELOG" > "${CHANGELOG}.tmp"
mv "${CHANGELOG}.tmp" "$CHANGELOG"

git -C "$ROOT" add "$CHANGELOG" "$VERSION_FILE"
git -C "$ROOT" commit -m "Prepare release ${VERSION}"
git -C "$ROOT" tag -a "$VERSION" -m "Release ${VERSION}"

echo "Created commit and tag ${VERSION}."
echo "Push when ready:"
echo "  git push origin main"
echo "  git push origin ${VERSION}"
echo "GitHub Actions will build and publish the release."
