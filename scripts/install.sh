#!/usr/bin/env bash
# install.sh — Install or uninstall quick-agent as a background service.
#
# Usage:
#   ./install.sh             Install the service
#   ./install.sh --uninstall Remove the service

set -euo pipefail

BINARY_NAME="quick-agent"
SERVICE_NAME="quick-agent"
INSTALL_DIR="${HOME}/.local/bin"
BINARY_SRC="$(cd "$(dirname "$0")/.." && pwd)/quick-agent"

# ─── helpers ────────────────────────────────────────────────────────────────

die() { echo "error: $*" >&2; exit 1; }
info() { echo "• $*"; }

# ─── uninstall ──────────────────────────────────────────────────────────────

uninstall() {
  if [[ "$(uname)" == "Darwin" ]]; then
    PLIST="${HOME}/Library/LaunchAgents/${SERVICE_NAME}.plist"
    if launchctl list | grep -q "${SERVICE_NAME}" 2>/dev/null; then
      info "Unloading launchd service..."
      launchctl unload "${PLIST}" 2>/dev/null || true
    fi
    rm -f "${PLIST}"
    info "Removed ${PLIST}"
  else
    UNIT="${HOME}/.config/systemd/user/${SERVICE_NAME}.service"
    if systemctl --user is-active --quiet "${SERVICE_NAME}" 2>/dev/null; then
      info "Stopping systemd service..."
      systemctl --user stop "${SERVICE_NAME}"
    fi
    if systemctl --user is-enabled --quiet "${SERVICE_NAME}" 2>/dev/null; then
      systemctl --user disable "${SERVICE_NAME}"
    fi
    rm -f "${UNIT}"
    systemctl --user daemon-reload 2>/dev/null || true
    info "Removed ${UNIT}"
  fi
  rm -f "${INSTALL_DIR}/${BINARY_NAME}"
  info "Uninstall complete."
}

# ─── install ────────────────────────────────────────────────────────────────

install() {
  # Verify binary exists
  if [[ ! -f "${BINARY_SRC}" ]]; then
    # Try building it
    info "Binary not found at ${BINARY_SRC}; attempting build..."
    pushd "$(dirname "${BINARY_SRC}")" > /dev/null
    go build -o "${BINARY_SRC}" ./cmd/quick-agent/ || die "Build failed. Run 'go build ./cmd/quick-agent/' first."
    popd > /dev/null
  fi

  mkdir -p "${INSTALL_DIR}"
  cp "${BINARY_SRC}" "${INSTALL_DIR}/${BINARY_NAME}"
  chmod +x "${INSTALL_DIR}/${BINARY_NAME}"
  info "Installed binary to ${INSTALL_DIR}/${BINARY_NAME}"

  if [[ "$(uname)" == "Darwin" ]]; then
    _install_launchd
  else
    _install_systemd
  fi

  info "Install complete. The daemon will start automatically on login."
}

_install_systemd() {
  UNIT_DIR="${HOME}/.config/systemd/user"
  mkdir -p "${UNIT_DIR}"
  cat > "${UNIT_DIR}/${SERVICE_NAME}.service" <<EOF
[Unit]
Description=quick-agent background daemon
After=graphical-session.target

[Service]
Type=simple
ExecStart=${INSTALL_DIR}/${BINARY_NAME} daemon
Restart=on-failure
RestartSec=5

[Install]
WantedBy=default.target
EOF
  systemctl --user daemon-reload
  systemctl --user enable --now "${SERVICE_NAME}"
  info "Systemd user service enabled and started."
}

_install_launchd() {
  PLIST_DIR="${HOME}/Library/LaunchAgents"
  mkdir -p "${PLIST_DIR}"
  cat > "${PLIST_DIR}/${SERVICE_NAME}.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
    "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>${SERVICE_NAME}</string>
    <key>ProgramArguments</key>
    <array>
        <string>${INSTALL_DIR}/${BINARY_NAME}</string>
        <string>daemon</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardErrorPath</key>
    <string>${HOME}/.quick-agent/launchd.log</string>
    <key>StandardOutPath</key>
    <string>${HOME}/.quick-agent/launchd.log</string>
</dict>
</plist>
EOF
  launchctl load "${PLIST_DIR}/${SERVICE_NAME}.plist"
  info "LaunchAgent loaded: ${PLIST_DIR}/${SERVICE_NAME}.plist"
}

# ─── entry point ─────────────────────────────────────────────────────────────

if [[ "${1:-}" == "--uninstall" ]]; then
  uninstall
else
  install
fi
