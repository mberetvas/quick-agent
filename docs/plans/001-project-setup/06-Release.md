# 06 - Release & Distribution

## Overview

This document covers the build, packaging, and release process for Clipboard TUI. The goal is to provide users with an easy way to install and update the application on all supported platforms.

---

## Build Process

### Local Build

```bash
# Build for current platform
go build -o clipboard-tui ./cmd/clipboard-tui

# Build for specific platform
GOOS=linux GOARCH=amd64 go build -o clipboard-tui-linux-amd64 ./cmd/clipboard-tui
GOOS=darwin GOARCH=amd64 go build -o clipboard-tui-darwin-amd64 ./cmd/clipboard-tui
GOOS=windows GOARCH=amd64 go build -o clipboard-tui-windows-amd64.exe ./cmd/clipboard-tui

# Build all platforms
./scripts/build.sh
```

### Build Scripts

```bash
#!/bin/bash
# scripts/build.sh

set -euo pipefail

PKG="github.com/yourname/clipboard-tui"
OUTPUT_DIR="./dist"
mkdir -p "$OUTPUT_DIR"

# Build for all platform/arch combinations
for os in linux darwin windows; do
  for arch in amd64 arm64; do
    # Skip ARM64 on Windows (not well supported)
    if [ "$os" = "windows" ] && [ "$arch" = "arm64" ]; then
      continue
    fi
    
    output="clipboard-tui-${os}-${arch}"
    if [ "$os" = "windows" ]; then
      output="${output}.exe"
    fi
    
    echo "Building ${output}..."
    GOOS=$os GOARCH=$arch CGO_ENABLED=1 go build -o "${OUTPUT_DIR}/${output}" "$PKG"
  done
done

echo "Build complete. Binaries in: $OUTPUT_DIR"
```

### Cross-Compilation Notes

**robotgo requires cgo**, which means:
- Cannot cross-compile from Linux to Windows/macOS
- Cannot cross-compile from macOS to Linux/Windows
- Must build on each platform natively

**Options**:
1. **Recommended**: Use GitHub Actions matrix to build on each OS
2. **Alternative**: Use Docker with cross-compilation toolchain (complex)
3. **Local**: Build on each machine you have access to

---

## GitHub Actions Workflows

### CI Pipeline (ci.yml)

```yaml
# .github/workflows/ci.yml
name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
    runs-on: ${{ matrix.os }}
    
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true
      
      - name: Install dependencies (Linux)
        if: matrix.os == 'ubuntu-latest'
        run: |
          sudo apt-get update
          sudo apt-get install -y xvfb
      
      - name: Test
        run: |
          go vet ./...
          go test -v -race ./...
      
      - name: Integration tests (Linux)
        if: matrix.os == 'ubuntu-latest'
        run: |
          Xvfb :99 -screen 0 1024x768x16 &
          export DISPLAY=:99
          go test -v -tags=integration ./...
  
  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
      - uses: golangci/golangci-lint-action@v6
        with:
          version: latest
          skip-cache: true
```

### Release Pipeline (release.yml)

```yaml
# .github/workflows/release.yml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  build:
    strategy:
      matrix:
        os: [ubuntu-latest, macos-latest, windows-latest]
        arch: [amd64]
        # Add arm64 for macOS/Linux in future
    runs-on: ${{ matrix.os }}
    
    steps:
      - uses: actions/checkout@v4
      
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
          cache: true
      
      - name: Install dependencies (Linux)
        if: matrix.os == 'ubuntu-latest'
        run: |
          sudo apt-get update
          sudo apt-get install -y libx11-dev libxtst-dev
      
      - name: Build
        env:
          GOOS: ${{ matrix.os }}
          GOARCH: ${{ matrix.arch }}
          CGO_ENABLED: 1
        run: |
          mkdir -p dist
          output="clipboard-tui-${{ matrix.os }}-${{ matrix.arch }}"
          if [ "${{ matrix.os }}" = "windows" ]; then
            output="${output}.exe"
          fi
          go build -o "dist/${output}" ./cmd/clipboard-tui
      
      - name: Upload artifacts
        uses: actions/upload-artifact@v4
        with:
          name: clipboard-tui-${{ matrix.os }}-${{ matrix.arch }}
          path: dist/*
          retention-days: 30
  
  release:
    needs: build
    runs-on: ubuntu-latest
    
    steps:
      - uses: actions/download-artifact@v4
        with:
          path: artifacts
          pattern: clipboard-tui-*
      
      - name: Create checksums
        run: |
          cd artifacts
          for file in *; do
            sha256sum "$file" > "${file}.sha256"
          done
          cd ..
      
      - name: Create release
        uses: softprops/action-gh-release@v2
        with:
          tag_name: ${{ github.ref_name }}
          name: Clipboard TUI ${{ github.ref_name }}
          body: |
            ## Changes
            See [CHANGELOG.md](https://github.com/yourname/clipboard-tui/blob/${{ github.ref_name }}/CHANGELOG.md)
            
            ## Assets
            - Linux (amd64)
            - macOS (amd64)
            - Windows (amd64)
          files: |
            artifacts/**/*
          generate_release_notes: true
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

---

## Release Process

### Versioning

Use **Semantic Versioning** (SemVer):
- `MAJOR`: Breaking changes, incompatible API changes
- `MINOR`: Backwards-compatible new features
- `PATCH`: Backwards-compatible bug fixes

### Creating a Release

#### Manual Process

```bash
# 1. Update version in config schema
sed -i 's/"version": "[^"]*"/"version": "1.0.0"/' internal/config/config.go

# 2. Update CHANGELOG.md
# Add new entry at the top

# 3. Commit changes
git add CHANGELOG.md internal/config/config.go
git commit -m "Prepare release v1.0.0"

# 4. Tag the release
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0

# 5. GitHub Actions will build and release automatically
```

#### Using Script

```bash
#!/bin/bash
# scripts/release.sh

VERSION=$1

if [ -z "$VERSION" ]; then
    echo "Usage: $0 <version>"
    echo "Example: $0 v1.0.0"
    exit 1
fi

# Update version in config
sed -i "s/\"version\": \"[^\"]*\"/\"version\": \"$VERSION\"/" internal/config/config.go

# Update CHANGELOG
DATE=$(date +"%Y-%m-%d")
echo "
## $VERSION - $DATE" | cat - CHANGELOG.md > CHANGELOG.md.tmp && mv CHANGELOG.md.tmp CHANGELOG.md

# Commit
git add CHANGELOG.md internal/config/config.go
git commit -m "Prepare release $VERSION"

# Tag
git tag -a $VERSION -m "Release $VERSION"

# Push
git push origin main
git push origin $VERSION

echo "Release $VERSION created and pushed. GitHub Actions will build and release."
```

### Pre-Release Checklist

- [ ] All v1 features implemented (see 02-Implementation.md)
- [ ] All unit tests pass
- [ ] Integration tests pass
- [ ] Manual tests completed on all platforms
- [ ] Documentation complete (README, INSTALL, CONFIG)
- [ ] CHANGELOG.md updated
- [ ] Version updated in config schema
- [ ] No breaking changes since last release
- [ ] Dependencies up to date (`go mod tidy`)
- [ ] CI pipeline passes
- [ ] Binary builds successfully on all platforms

---

## Packaging

### Binary Packaging

Each release includes:
- Pre-built binaries for all supported platforms
- SHA256 checksums for each binary
- Release notes

**File naming convention**:
- `clipboard-tui-linux-amd64`
- `clipboard-tui-linux-amd64.sha256`
- `clipboard-tui-darwin-amd64`
- `clipboard-tui-darwin-amd64.sha256`
- `clipboard-tui-windows-amd64.exe`
- `clipboard-tui-windows-amd64.exe.sha256`

### Archive Packaging (Optional)

```bash
# scripts/package.sh
#!/bin/bash

VERSION=$1
OUTPUT_DIR="./release"

mkdir -p "$OUTPUT_DIR"

for os in linux darwin windows; do
  for arch in amd64; do
    file="clipboard-tui-${os}-${arch}"
    if [ "$os" = "windows" ]; then
      file="${file}.exe"
    fi
    
    # Create tarball
    tar czvf "${OUTPUT_DIR}/${file}-${VERSION}.tar.gz" dist/${file} README.md LICENSE
    
    # Create zip for Windows
    if [ "$os" = "windows" ]; then
      zip "${OUTPUT_DIR}/${file}-${VERSION}.zip" dist/${file} README.md LICENSE
    fi
  done
done
```

---

## Package Managers

### Homebrew (macOS/Linux)

**Tap repository**: `yourname/homebrew-tap`

**Formula** (`Formula/clipboard-tui.rb`):
```ruby
class ClipboardTui < Formula
  desc "AI-powered clipboard supercharger"
  homepage "https://github.com/yourname/clipboard-tui"
  url "https://github.com/yourname/clipboard-tui/archive/refs/tags/v1.0.0.tar.gz"
  sha256 "<SHA256_OF_SOURCE_TARBALL>"
  license "MIT"
  head "https://github.com/yourname/clipboard-tui.git", branch: "main"

  def install
    system "go", "build", *std_go_args(ldflags: -s -w)
  end

  service do
    run [opt_bin/"clipboard-tui", "daemon"]
    run_type :user
    keep_alive true
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/clipboard-tui --version")
  end
end
```

**Release workflow for Homebrew**:
```yaml
# .github/workflows/homebrew.yml
name: Update Homebrew

on:
  release:
    types: [published]

jobs:
  update-homebrew:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          repository: yourname/homebrew-tap
          token: ${{ secrets.HOMEBREW_TAP_TOKEN }}
      
      - name: Update formula
        run: |
          # Update version and SHA256 in formula
          # Commit and push
          git config user.name "GitHub Actions"
          git config user.email "actions@github.com"
          git add Formula/clipboard-tui.rb
          git commit -m "clipboard-tui: update to ${{ github.event.release.tag_name }}"
          git push
```

### Scoop (Windows)

**Bucket**: `yourname/scoop-bucket`

**Manifest** (`bucket/clipboard-tui.json`):
```json
{
  "version": "1.0.0",
  "homepage": "https://github.com/yourname/clipboard-tui",
  "license": "MIT",
  "url": "https://github.com/yourname/clipboard-tui/releases/download/v1.0.0/clipboard-tui-windows-amd64.exe",
  "hash": "<SHA256_OF_BINARY>",
  "extract_dir": "",
  "bin": ["clipboard-tui-windows-amd64.exe"],
  "shortcuts": [
    ["clipboard-tui.exe", "Clipboard TUI"]
  ],
  "checkver": "github",
  "autoupdate": {
    "url": "https://github.com/yourname/clipboard-tui/releases.atom",
    "extract_dir": ""
  }
}
```

---

## Distribution Channels

| Channel | Platform | Auto-Update | Notes |
|---------|----------|-------------|-------|
| GitHub Releases | All | ❌ | Manual download |
| Homebrew | macOS/Linux | ✅ | `brew install yourname/tap/clipboard-tui` |
| Scoop | Windows | ✅ | `scoop install yourname/clipboard-tui` |
| Chocolatey | Windows | ✅ | v2 |
| Snap/Flatpak | Linux | ✅ | v2 |

---

## Auto-Update (v2)

### Strategy

Check for new versions on daemon startup:
1. Fetch latest release from GitHub API
2. Compare with current version
3. Notify user if update available
4. Provide option to download and install

### Implementation

```go
// internal/updater/updater.go
package updater

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

const (
    githubAPI = "https://api.github.com/repos/yourname/clipboard-tui/releases/latest"
    checkInterval = 24 * time.Hour
)

type GitHubRelease struct {
    TagName string `json:"tag_name"`
}

func CheckForUpdate(ctx context.Context, currentVersion string) (string, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", githubAPI, nil)
    if err != nil {
        return "", err
    }
    
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        return "", fmt.Errorf("GitHub API returned %d", resp.StatusCode)
    }
    
    var release GitHubRelease
    if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
        return "", err
    }
    
    if release.TagName > currentVersion {
        return release.TagName, nil
    }
    
    return "", nil
}
```

### User Notification

```go
// In daemon startup
func (d *Daemon) Start() {
    // Check for updates in background
    go d.checkForUpdates()
}

func (d *Daemon) checkForUpdates() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    latest, err := updater.CheckForUpdate(ctx, version)
    if err != nil {
        log.Debug("Failed to check for updates", "error", err)
        return
    }
    
    if latest != "" {
        log.Info("Update available", "current", version, "latest", latest)
        // Notify user via:
        // - Desktop notification (v2)
        // - Log message
        // - TUI notification on next open (v2)
    }
}
```

---

## Installation Instructions

### macOS

**Option 1: Homebrew (Recommended)**
```bash
brew tap yourname/tap
brew install clipboard-tui
```

**Option 2: Manual Install**
```bash
# Download binary
curl -LO https://github.com/yourname/clipboard-tui/releases/latest/download/clipboard-tui-darwin-amd64
chmod +x clipboard-tui-darwin-amd64
sudo mv clipboard-tui-darwin-amd64 /usr/local/bin/clipboard-tui

# Install daemon
clipboard-tui daemon --install

# Start daemon
clipboard-tui daemon
```

### Linux

**Option 1: Homebrew (Recommended)**
```bash
brew tap yourname/tap
brew install clipboard-tui
```

**Option 2: Manual Install**
```bash
# Download binary
curl -LO https://github.com/yourname/clipboard-tui/releases/latest/download/clipboard-tui-linux-amd64
chmod +x clipboard-tui-linux-amd64
sudo mv clipboard-tui-linux-amd64 /usr/local/bin/clipboard-tui

# Install daemon (systemd)
clipboard-tui daemon --install

# Start daemon
clipboard-tui daemon
```

### Windows

**Option 1: Scoop (Recommended)**
```powershell
scoop bucket add yourname https://github.com/yourname/scoop-bucket.git
scoop install clipboard-tui
```

**Option 2: Manual Install**
```powershell
# Download binary
Invoke-WebRequest -Uri "https://github.com/yourname/clipboard-tui/releases/latest/download/clipboard-tui-windows-amd64.exe" -OutFile "clipboard-tui.exe"

# Install daemon
./clipboard-tui daemon --install

# Start daemon
./clipboard-tui daemon
```

---

## Post-Release

### Announcement

1. **GitHub Release**: Auto-created by GitHub Actions
2. **Twitter/X**: Post announcement with key features
3. **Reddit**: Post to r/golang, r/selfhosted, r/StableDiffusion
4. **Hacker News**: Submit to HN if significant release
5. **Discord/Slack**: Notify relevant communities

### Monitoring

1. **Download stats**: Track GitHub release downloads
2. **Issue tracker**: Monitor for bugs in new release
3. **User feedback**: Respond to issues and discussions

---

## Rollback Plan

If a release has critical bugs:

1. **Yank the release** on GitHub
2. **Revert the tag**:
   ```bash
   git tag -d v1.0.0
   git push origin :refs/tags/v1.0.0
   ```
3. **Create new patch release** with fix
4. **Notify users** via:
   - GitHub release notes
   - Issue tracker
   - Social media

---

## Version Compatibility

| Version | Go Version | Backend Support | Notes |
|---------|------------|------------------|-------|
| v1.0.0 | 1.22+ | Ollama, OpenRouter | Initial release |
| v1.x.x | 1.22+ | Ollama, OpenRouter | Bug fixes |
| v2.0.0 | 1.23+ | +Plugin system | Breaking changes |

### Backwards Compatibility

- **v1**: All v1.x releases are backwards compatible
- **Config**: Schema migrations handle version differences
- **API**: LLM client interface remains stable

---

## Checklist for Each Release

### Pre-Release
- [ ] All v1 features complete
- [ ] All tests pass
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Version updated in config schema
- [ ] Dependencies audited (`go mod tidy`, `go vet`)
- [ ] Security scan clean
- [ ] Manual testing on all platforms

### Release Day
- [ ] Create Git tag
- [ ] Push tag to GitHub
- [ ] GitHub Actions builds successfully
- [ ] Release created on GitHub
- [ ] Binaries uploaded
- [ ] Checksums generated
- [ ] Package managers updated (Homebrew, Scoop)

### Post-Release
- [ ] Announcement posted
- [ ] Download stats monitored
- [ ] Issues triaged
- [ ] User feedback collected
- [ ] Next version planned
