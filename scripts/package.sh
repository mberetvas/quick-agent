#!/usr/bin/env bash
# Package dist/ binaries into release archives with README and LICENSE.
set -euo pipefail

VERSION="${1:-}"
if [ -z "$VERSION" ]; then
  echo "Usage: $0 <version>"
  echo "Example: $0 v1.0.0"
  exit 1
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DIST="${ROOT}/dist"
OUTPUT_DIR="${ROOT}/release"
README="${ROOT}/README.md"
LICENSE="${ROOT}/LICENSE"

if [ ! -d "$DIST" ]; then
  echo "Run scripts/build.sh or CI release first (missing dist/)"
  exit 1
fi

mkdir -p "$OUTPUT_DIR"

shopt -s nullglob
for bin in "$DIST"/quick-agent-*; do
  base="$(basename "$bin")"
  stem="${base%.exe}"
  staging="$(mktemp -d)"
  trap 'rm -rf "$staging"' RETURN

  cp "$bin" "$staging/"
  [ -f "$README" ] && cp "$README" "$staging/"
  [ -f "$LICENSE" ] && cp "$LICENSE" "$staging/"

  if [[ "$base" == *.exe ]]; then
    (cd "$staging" && zip -j "${OUTPUT_DIR}/${stem}-${VERSION}.zip" ./*)
  else
    tar czf "${OUTPUT_DIR}/${stem}-${VERSION}.tar.gz" -C "$staging" .
  fi

  rm -rf "$staging"
  trap - RETURN
done

echo "Packages written to: ${OUTPUT_DIR}"
