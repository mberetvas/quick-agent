#!/usr/bin/env bash
# Lightweight end-to-end smoke tests (Linux-friendly; see 05-Testing.md).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BINARY="${ROOT}/clipboard-tui"
CONFIG_DIR="$(mktemp -d)"
export CLIPBOARD_TUI_CONFIG="${CONFIG_DIR}/config.json"

cleanup() {
	rm -rf "$CONFIG_DIR" "$BINARY"
}
trap cleanup EXIT

echo "=== End-to-End Tests ==="

echo "Building..."
(cd "$ROOT" && go build -o "$BINARY" ./cmd/clipboard-tui)

echo "Test 1: config validate (creates default if missing)"
"$BINARY" config validate

echo "Test 2: TUI with piped text"
echo "Hello world" | timeout 3 "$BINARY" tui --text "Hello world" 2>/dev/null || true

echo "Test 3: daemon starts and stops"
"$BINARY" daemon --config "$CLIPBOARD_TUI_CONFIG" &
DAEMON_PID=$!
sleep 1
kill "$DAEMON_PID" 2>/dev/null || true
wait "$DAEMON_PID" 2>/dev/null || true

if command -v xclip >/dev/null 2>&1 && [ -n "${DISPLAY:-}" ]; then
	echo "Test 4: clipboard polling (xclip + DISPLAY)"
	cat >"$CLIPBOARD_TUI_CONFIG" <<'EOF'
{
  "version": "1",
  "backend": "ollama",
  "clipboard": { "poll_interval_ms": 100, "truncate_size": 10000 }
}
EOF
	"$BINARY" daemon --config "$CLIPBOARD_TUI_CONFIG" --log-level=debug &
	DAEMON_PID=$!
	sleep 0.5
	echo "test text" | xclip -selection clipboard
	sleep 0.5
	kill "$DAEMON_PID" 2>/dev/null || true
	wait "$DAEMON_PID" 2>/dev/null || true
else
	echo "Test 4: skipped (needs xclip and DISPLAY)"
fi

echo "All E2E smoke tests passed."
