//go:build integration && cgo

package hotkey

import (
	"context"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/mberetvas/quick-agent/internal/config"
)

func TestListener_RealDetectorStarts(t *testing.T) {
	if runtime.GOOS == "linux" && os.Getenv("DISPLAY") == "" {
		t.Skip("DISPLAY not set; hotkey integration requires a display server")
	}
	if defaultDetector == nil {
		t.Skip("gohook detector not available (CGO disabled or unsupported platform)")
	}

	cfg := config.Default().Hotkey
	listener := NewListener(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	events := make(chan struct{}, 1)
	errCh := make(chan error, 1)
	go func() {
		errCh <- listener.Start(ctx, events)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			t.Fatalf("Start failed: %v", err)
		}
	case <-events:
		t.Fatal("unexpected hotkey event without user input")
	case <-ctx.Done():
		// Listener should exit cleanly after context timeout.
		select {
		case err := <-errCh:
			if err != nil {
				t.Fatalf("Start returned error: %v", err)
			}
		case <-time.After(2 * time.Second):
			t.Fatal("listener did not stop after context cancellation")
		}
	}
}
