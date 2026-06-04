#!/usr/bin/env bash
# Manual test helper for clipboard-tui (see docs/plans/001-project-setup/05-Testing.md).
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
BINARY="${ROOT}/clipboard-tui"

echo "=== Manual Test Suite ==="
echo ""

echo "Building..."
(cd "$ROOT" && go build -o "$BINARY" ./cmd/clipboard-tui)

echo "Starting daemon..."
"$BINARY" daemon --log-level=debug &
DAEMON_PID=$!
sleep 2

cleanup() {
	echo "Cleaning up..."
	kill "$DAEMON_PID" 2>/dev/null || true
	pkill -f "$BINARY" 2>/dev/null || true
	rm -f "$BINARY"
}
trap cleanup EXIT

echo "Test 1: Press the configured hotkey (default Ctrl+Alt+V / Cmd+Option+V)"
echo "Expected: TUI appears with clipboard content"
read -r -p "Press Enter when done... "

echo ""
echo "Test 2: Copy empty text, press hotkey"
echo "Expected: empty clipboard handling in TUI/daemon"
read -r -p "Press Enter when done... "

echo ""
echo "Test 3: Copy text, hotkey, select Refine in options"
echo "Expected: streaming LLM response (requires configured backend)"
read -r -p "Press Enter when done... "

echo ""
echo "Test 4: Translate flow (language picker stub may apply)"
read -r -p "Press Enter when done... "

echo "Manual tests complete."
