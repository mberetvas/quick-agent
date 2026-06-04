#!/usr/bin/env bash
# Build clipboard-tui binaries for the current OS (all supported archs).
# gohook/robotn require CGO, so cross-OS builds must use CI (.github/workflows/release.yml).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
PKG="./cmd/clipboard-tui"
OUTPUT_DIR="${ROOT}/dist"
VERSION="${VERSION:-dev}"
LDFLAGS="-s -w -X github.com/yourname/clipboard-tui/internal/version.Version=${VERSION}"

mkdir -p "$OUTPUT_DIR"

current_os="$(go env GOOS)"
archs=(amd64 arm64)

for arch in "${archs[@]}"; do
  if [ "$current_os" = "windows" ] && [ "$arch" = "arm64" ]; then
    continue
  fi

  output="clipboard-tui-${current_os}-${arch}"
  if [ "$current_os" = "windows" ]; then
    output="${output}.exe"
  fi

  echo "Building ${output}..."
  CGO_ENABLED=1 GOOS="$current_os" GOARCH="$arch" go build \
    -ldflags "$LDFLAGS" \
    -o "${OUTPUT_DIR}/${output}" \
    "$PKG"
done

echo "Build complete. Binaries in: ${OUTPUT_DIR}"
