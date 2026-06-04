//go:build integration

package clipboard

import (
	"context"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/yourname/clipboard-tui/internal/config"
)

func requireDisplay(t *testing.T) {
	t.Helper()
	if runtime.GOOS == "linux" && os.Getenv("DISPLAY") == "" {
		t.Skip("DISPLAY not set; start Xvfb for clipboard integration tests")
	}
}

func TestPoller_SystemClipboardDetectsChanges(t *testing.T) {
	requireDisplay(t)

	sys := SystemClipboard{}
	if err := sys.Set("clipboard-tui-initial"); err != nil {
		t.Skipf("system clipboard unavailable: %v", err)
	}

	cfg := config.Default()
	cfg.Clipboard.PollIntervalMS = 100
	cfg.Clipboard.TruncateSize = 10000

	poller := NewPoller(&sys, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	poller.Start(ctx)

	// Wait until poller has read the initial value.
	deadline := time.After(2 * time.Second)
	for poller.LatestText() != "clipboard-tui-initial" {
		select {
		case <-deadline:
			t.Fatalf("LatestText() = %q, want clipboard-tui-initial", poller.LatestText())
		case <-time.After(20 * time.Millisecond):
		}
	}

	if err := sys.Set("clipboard-tui-changed"); err != nil {
		t.Fatalf("set clipboard: %v", err)
	}

	select {
	case text := <-poller.Changes():
		if text != "clipboard-tui-changed" {
			t.Errorf("change = %q, want clipboard-tui-changed", text)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for clipboard change")
	}
}
