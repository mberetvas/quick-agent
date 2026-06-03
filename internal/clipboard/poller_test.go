package clipboard

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/yourname/clipboard-tui/internal/config"
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

func TestSanitize(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxChars int
		want     string
	}{
		{
			name:     "valid simple text",
			input:    "hello world",
			maxChars: 100,
			want:     "hello world",
		},
		{
			name:     "empty text",
			input:    "",
			maxChars: 100,
			want:     "",
		},
		{
			name:     "whitespace only",
			input:    "   \n \t  ",
			maxChars: 100,
			want:     "",
		},
		{
			name:     "non-UTF8 input cleaned",
			input:    "hello \xff world", // \xff is invalid UTF-8
			maxChars: 100,
			want:     "hello \ufffd world", // replaced by rune error replacement char
		},
		{
			name:     "truncation with marker",
			input:    "abcdefghij",
			maxChars: 5,
			want:     "abcde" + TruncateMarker,
		},
		{
			name:     "exactly max length no truncation",
			input:    "abcde",
			maxChars: 5,
			want:     "abcde",
		},
		{
			name:     "unicode character aware truncation",
			input:    "世界你好!", // 5 runes
			maxChars: 2,
			want:     "世界" + TruncateMarker,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Sanitize(tt.input, tt.maxChars)
			if got != tt.want {
				t.Errorf("Sanitize() got = %q, want = %q", got, tt.want)
			}
		})
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
	if poller.currentPoll != 500*time.Millisecond {
		t.Errorf("expected initial currentPoll to be 500ms, got %v", poller.currentPoll)
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
	if poller.currentPoll != 100*time.Millisecond {
		t.Errorf("expected currentPoll to be 100ms (fast poll) after change, got %v", poller.currentPoll)
	}

	// Let's do another cycle with no change. Poller should adaptively back off (add 100ms)
	// We wait 150ms (enough to trigger one tick of 100ms)
	time.Sleep(150 * time.Millisecond)

	if poller.currentPoll < 150*time.Millisecond {
		t.Errorf("expected currentPoll to adaptively back off (be > 100ms), got %v", poller.currentPoll)
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
