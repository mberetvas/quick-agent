package daemon

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
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

func TestRun_handleHotkey_executableError(t *testing.T) {
	cfg := config.Default()
	spawner := &fakeSpawner{}
	hotkeyCh := make(chan struct{}, 1)
	rt := testRuntime(spawner, "clipboard body", hotkeyCh)
	rt.Executable = func() (string, error) { return "", errors.New("no exe") }

	ctx, cancel := context.WithCancel(context.Background())
	go func() { _ = RunWithRuntime(ctx, cfg, &rt) }()

	hotkeyCh <- struct{}{}
	time.Sleep(100 * time.Millisecond)

	spawner.mu.Lock()
	n := len(spawner.calls)
	spawner.mu.Unlock()
	if n != 0 {
		t.Fatalf("expected no spawn when executable fails, got %d calls", n)
	}
	cancel()
}

func TestRun_handleHotkey_spawnError(t *testing.T) {
	cfg := config.Default()
	spawner := &fakeSpawner{err: errors.New("spawn failed")}
	hotkeyCh := make(chan struct{}, 1)
	rt := testRuntime(spawner, "clipboard body", hotkeyCh)

	ctx, cancel := context.WithCancel(context.Background())
	go func() { _ = RunWithRuntime(ctx, cfg, &rt) }()

	hotkeyCh <- struct{}{}
	time.Sleep(100 * time.Millisecond)

	spawner.mu.Lock()
	n := len(spawner.calls)
	spawner.mu.Unlock()
	if n != 1 {
		t.Fatalf("expected one spawn attempt, got %d", n)
	}
	cancel()
}

func TestRunWithRuntime_clipboardChange(t *testing.T) {
	cfg := config.Default()
	poller := newFakePoller("text")
	rt := &Runtime{
		Logger: testLogger(),
		AcquirePID: func(string) (func(), error) {
			return func() {}, nil
		},
		NewPoller: func(context.Context, *config.Config) ClipboardPoller { return poller },
		NewHotkey: func(config.HotkeyConfig) HotkeyListener {
			return &blockingHotkey{}
		},
		NewSpawner: func(config.TerminalConfig) TUISpawner { return &fakeSpawner{} },
		Executable: func() (string, error) { return "/bin/clipboard-tui", nil },
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- RunWithRuntime(ctx, cfg, rt) }()

	poller.changes <- "updated clipboard"
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatalf("RunWithRuntime() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunWithRuntime did not exit")
	}
}

func TestRunWithRuntime_hotkeyErrorOnShutdown(t *testing.T) {
	cfg := config.Default()
	rt := &Runtime{
		Logger: testLogger(),
		AcquirePID: func(string) (func(), error) {
			return func() {}, nil
		},
		NewPoller:  func(context.Context, *config.Config) ClipboardPoller { return newFakePoller("") },
		NewHotkey:  func(config.HotkeyConfig) HotkeyListener { return &errorHotkey{} },
		NewSpawner: func(config.TerminalConfig) TUISpawner { return &fakeSpawner{} },
		Executable: func() (string, error) { return "/bin/clipboard-tui", nil },
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- RunWithRuntime(ctx, cfg, rt) }()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err == nil || !strings.Contains(err.Error(), "hotkey listener") {
			t.Fatalf("RunWithRuntime() error = %v, want hotkey listener error", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunWithRuntime did not exit")
	}
}

func TestRun_delegatesToRunWithRuntime(t *testing.T) {
	cfg := config.Default()
	ctx, cancel := context.WithCancel(context.Background())

	rt := &Runtime{
		Logger: testLogger(),
		AcquirePID: func(string) (func(), error) {
			return func() {}, nil
		},
		NewPoller:  func(context.Context, *config.Config) ClipboardPoller { return newFakePoller("") },
		NewHotkey:  func(config.HotkeyConfig) HotkeyListener { return &blockingHotkey{} },
		NewSpawner: func(config.TerminalConfig) TUISpawner { return &fakeSpawner{} },
		Executable: func() (string, error) { return "/bin/clipboard-tui", nil },
	}

	done := make(chan error, 1)
	go func() { done <- RunWithRuntime(ctx, cfg, rt) }()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatalf("RunWithRuntime() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunWithRuntime did not exit after cancel")
	}
}

type errorHotkey struct{}

func (e *errorHotkey) Start(ctx context.Context, _ chan<- struct{}) error {
	<-ctx.Done()
	return errors.New("listener died")
}

type exitHotkey struct{}

func (exitHotkey) Start(context.Context, chan<- struct{}) error { return nil }

func TestRunWithRuntime_hotkeyExitsCleanly(t *testing.T) {
	cfg := config.Default()
	rt := &Runtime{
		Logger: testLogger(),
		AcquirePID: func(string) (func(), error) {
			return func() {}, nil
		},
		NewPoller:  func(context.Context, *config.Config) ClipboardPoller { return newFakePoller("") },
		NewHotkey:  func(config.HotkeyConfig) HotkeyListener { return exitHotkey{} },
		NewSpawner: func(config.TerminalConfig) TUISpawner { return &fakeSpawner{} },
		Executable: func() (string, error) { return "/bin/clipboard-tui", nil },
	}

	if err := RunWithRuntime(context.Background(), cfg, rt); err != nil {
		t.Fatalf("RunWithRuntime() error = %v", err)
	}
}

func TestRunWithRuntime_pollerError(t *testing.T) {
	cfg := config.Default()
	poller := newFakePoller("text")
	rt := &Runtime{
		Logger: testLogger(),
		AcquirePID: func(string) (func(), error) {
			return func() {}, nil
		},
		NewPoller: func(context.Context, *config.Config) ClipboardPoller { return poller },
		NewHotkey: func(config.HotkeyConfig) HotkeyListener {
			return &blockingHotkey{}
		},
		NewSpawner: func(config.TerminalConfig) TUISpawner { return &fakeSpawner{} },
		Executable: func() (string, error) { return "/bin/clipboard-tui", nil },
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- RunWithRuntime(ctx, cfg, rt) }()

	poller.errors <- errors.New("poll failed")
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil && !errors.Is(err, context.Canceled) {
			t.Fatalf("RunWithRuntime() error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunWithRuntime did not exit")
	}
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
