package clipboard

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mberetvas/quick-agent/internal/config"
)

type errorOnGetClipboard struct{}

func (errorOnGetClipboard) Get() (string, error) { return "", errors.New("clipboard read failed") }
func (errorOnGetClipboard) Set(string) error     { return nil }

func TestPollerStart_emitsGetErrors(t *testing.T) {
	cfg := config.Default()
	cfg.Clipboard.PollIntervalMS = 100
	poller := NewPoller(errorOnGetClipboard{}, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	poller.Start(ctx)

	select {
	case err := <-poller.Errors():
		if err == nil {
			t.Fatal("expected error from poller")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for poller error")
	}
}

func TestPollerStart_doubleStartIsNoop(t *testing.T) {
	cfg := config.Default()
	poller := NewPoller(mockClipboard{text: "x"}, cfg)
	ctx := context.Background()
	poller.Start(ctx)
	poller.Start(ctx) // second call should return immediately
}

type mockClipboard struct {
	text string
}

func (m mockClipboard) Get() (string, error) { return m.text, nil }
func (m mockClipboard) Set(string) error     { return nil }
