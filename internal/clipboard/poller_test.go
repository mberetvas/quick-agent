package clipboard

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/mberetvas/quick-agent/internal/config"
)

// MockClipboard is an in-memory mock implementation of Clipboard for unit testing.
type MockClipboard struct {
	mu   sync.Mutex
	text string
	err  error
}

func (m *MockClipboard) Get() (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return "", m.err
	}
	return m.text, nil
}

func (m *MockClipboard) Set(text string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	m.text = text
	return nil
}

func (m *MockClipboard) SetError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.err = err
}

func TestPollerLatestText(t *testing.T) {
	mock := &MockClipboard{text: "initial"}
	cfg := config.Default()
	cfg.Clipboard.TruncateSize = 100
	cfg.Clipboard.PollIntervalMS = 50

	poller := NewPoller(mock, cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	poller.Start(ctx)

	deadline := time.After(2 * time.Second)
	for poller.LatestText() != "initial" {
		select {
		case <-deadline:
			t.Fatalf("LatestText() = %q, want initial", poller.LatestText())
		case <-time.After(10 * time.Millisecond):
		}
	}

	mock.text = "updated"
	for poller.LatestText() != "updated" {
		select {
		case <-deadline:
			t.Fatalf("LatestText() = %q, want updated", poller.LatestText())
		case <-time.After(10 * time.Millisecond):
		}
	}
}

func TestPollerDetectsChanges(t *testing.T) {
	cfg := config.Default()
	cfg.Clipboard.PollIntervalMS = 200

	mock := &MockClipboard{text: "initial"}
	poller := NewPoller(mock, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	poller.Start(ctx)

	// Update the clipboard text in mock
	time.Sleep(50 * time.Millisecond)
	mock.Set("next change")

	select {
	case change := <-poller.Changes():
		if change != "next change" {
			t.Errorf("expected change to be 'next change', got '%q'", change)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for clipboard change notification")
	}
}

func TestPollerAdaptiveBackoff(t *testing.T) {
	cfg := config.Default()
	cfg.Clipboard.PollIntervalMS = 500

	mock := &MockClipboard{text: "initial"}
	poller := NewPoller(mock, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	poller.Start(ctx)

	// Since initial matches "initial", currentPoll starts slow at 500ms
	if poller.CurrentPoll() != 500*time.Millisecond {
		t.Errorf("expected initial currentPoll to be 500ms, got %v", poller.CurrentPoll())
	}

	// Trigger a change
	mock.Set("second")

	// Wait for change to be processed and poller to adapt to fast polling
	select {
	case change := <-poller.Changes():
		if change != "second" {
			t.Errorf("expected change 'second', got '%q'", change)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for change")
	}

	// At this point, a change was detected, currentPoll should drop to fastPoll (100ms)
	if poller.CurrentPoll() != 100*time.Millisecond {
		t.Errorf("expected currentPoll to be 100ms (fast poll) after change, got %v", poller.CurrentPoll())
	}

	// Let's do another cycle with no change. Poller should adaptively back off (add 100ms)
	// We wait 150ms (enough to trigger one tick of 100ms)
	time.Sleep(150 * time.Millisecond)

	if poller.CurrentPoll() < 150*time.Millisecond {
		t.Errorf("expected currentPoll to adaptively back off (be > 100ms), got %v", poller.CurrentPoll())
	}
}

func TestPollerEmitsErrors(t *testing.T) {
	cfg := config.Default()
	cfg.Clipboard.PollIntervalMS = 200

	mock := &MockClipboard{}
	mock.SetError(errors.New("clipboard locked"))

	poller := NewPoller(mock, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	poller.Start(ctx)

	select {
	case err := <-poller.Errors():
		if !strings.Contains(err.Error(), "clipboard locked") {
			t.Errorf("expected error 'clipboard locked', got: %v", err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for poller error report")
	}
}

func BenchmarkPollerSanitizePath(b *testing.B) {
	mock := &MockClipboard{text: "hello world from clipboard"}
	cfg := config.Default()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		raw, err := mock.Get()
		if err != nil {
			b.Fatal(err)
		}
		_ = Sanitize(raw, cfg.Clipboard.TruncateSize)
	}
}

func TestPollerUsesTruncateSize(t *testing.T) {
	cfg := config.Default()
	cfg.Clipboard.MaxSize = 20
	cfg.Clipboard.TruncateSize = 10
	cfg.Clipboard.PollIntervalMS = 100

	mock := &MockClipboard{text: "initial"}
	poller := NewPoller(mock, cfg)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	poller.Start(ctx)

	// Set value larger than TruncateSize but within MaxSize
	val := "abcdefghijklm" // 13 chars
	mock.Set(val)

	select {
	case change := <-poller.Changes():
		expected := "abcdefghij" + TruncateMarker
		if change != expected {
			t.Errorf("expected truncated change %q, got %q", expected, change)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for clipboard change notification")
	}
}
