package daemon

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/mberetvas/quick-agent/internal/config"
	"github.com/mberetvas/quick-agent/internal/terminal"
)

type fakePoller struct {
	text    string
	changes chan string
	errors  chan error
}

func newFakePoller(text string) *fakePoller {
	return &fakePoller{
		text:    text,
		changes: make(chan string, 1),
		errors:  make(chan error, 1),
	}
}

func (f *fakePoller) Start(context.Context) {}

func (f *fakePoller) LatestText() string { return f.text }

func (f *fakePoller) Changes() <-chan string { return f.changes }

func (f *fakePoller) Errors() <-chan error { return f.errors }

type fakeSpawner struct {
	mu    sync.Mutex
	calls []spawnCall
	err   error
}

type spawnCall struct {
	exe  string
	text string
}

func (s *fakeSpawner) SpawnTUI(_ context.Context, executable, text string) error {
	s.mu.Lock()
	s.calls = append(s.calls, spawnCall{executable, text})
	err := s.err
	s.mu.Unlock()
	return err
}

type immediateHotkey struct {
	ch chan struct{}
}

func (h *immediateHotkey) Start(ctx context.Context, events chan<- struct{}) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case ev := <-h.ch:
			select {
			case events <- ev:
			case <-ctx.Done():
				return ctx.Err()
			}
		}
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

func testRuntime(spawner *fakeSpawner, text string, hotkeyCh chan struct{}) Runtime {
	return Runtime{
		Logger: testLogger(),
		AcquirePID: func(string) (func(), error) {
			return func() {}, nil
		},
		NewPoller: func(context.Context, *config.Config) ClipboardPoller {
			return newFakePoller(text)
		},
		NewHotkey: func(config.HotkeyConfig) HotkeyListener {
			return &immediateHotkey{ch: hotkeyCh}
		},
		NewSpawner: func(config.TerminalConfig) TUISpawner {
			return spawner
		},
		Executable: func() (string, error) { return "/bin/clipboard-tui", nil },
	}
}

func TestRun_spawnsTUIOnHotkey(t *testing.T) {
	cfg := config.Default()
	spawner := &fakeSpawner{}
	hotkeyCh := make(chan struct{}, 1)
	rt := testRuntime(spawner, "clipboard body", hotkeyCh)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		_ = RunWithRuntime(ctx, cfg, &rt)
		close(done)
	}()

	hotkeyCh <- struct{}{}

	deadline := time.After(2 * time.Second)
	for {
		spawner.mu.Lock()
		n := len(spawner.calls)
		spawner.mu.Unlock()
		if n > 0 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("timeout waiting for SpawnTUI")
		case <-time.After(10 * time.Millisecond):
		}
	}

	spawner.mu.Lock()
	call := spawner.calls[0]
	spawner.mu.Unlock()
	if call.text != "clipboard body" {
		t.Errorf("spawn text = %q, want clipboard body", call.text)
	}
	if call.exe != "/bin/clipboard-tui" {
		t.Errorf("spawn exe = %q", call.exe)
	}

	cancel()
	<-done
}

func TestRun_skipsEmptyClipboard(t *testing.T) {
	cfg := config.Default()
	spawner := &fakeSpawner{}
	hotkeyCh := make(chan struct{}, 1)
	rt := testRuntime(spawner, "", hotkeyCh)

	ctx, cancel := context.WithCancel(context.Background())
	go func() { _ = RunWithRuntime(ctx, cfg, &rt) }()

	hotkeyCh <- struct{}{}
	time.Sleep(100 * time.Millisecond)

	spawner.mu.Lock()
	n := len(spawner.calls)
	spawner.mu.Unlock()
	if n != 0 {
		t.Fatalf("expected no spawn for empty clipboard, got %d calls", n)
	}
	cancel()
}

func TestRun_fallbackIsInfoNotFatal(t *testing.T) {
	cfg := config.Default()
	spawner := &fakeSpawner{err: errors.Join(terminal.ErrUsedFallback, errors.New("no terminal"))}
	hotkeyCh := make(chan struct{}, 1)
	rt := testRuntime(spawner, "text", hotkeyCh)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	errCh := make(chan error, 1)
	go func() { errCh <- RunWithRuntime(ctx, cfg, &rt) }()

	hotkeyCh <- struct{}{}
	time.Sleep(200 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		if err != nil && !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("Run() error = %v, want cancellation only", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit")
	}
}
