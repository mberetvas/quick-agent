package daemon

import (
	"context"
	"testing"
	"time"

	"github.com/mberetvas/quick-agent/internal/config"
)

func TestRunWithRuntime_immediateCancel(t *testing.T) {
	cfg := config.Default()
	poller := newFakePoller("")

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	rt := &Runtime{
		AcquirePID: func(string) (func(), error) {
			return func() {}, nil
		},
		NewPoller: func(context.Context, *config.Config) ClipboardPoller { return poller },
		NewHotkey: func(config.HotkeyConfig) HotkeyListener {
			return &blockingHotkey{}
		},
		NewSpawner: func(config.TerminalConfig) TUISpawner { return &fakeSpawner{} },
		Executable: func() (string, error) { return "/bin/quick-agent", nil },
	}

	err := RunWithRuntime(ctx, cfg, rt)
	if err != nil && err != context.DeadlineExceeded && err != context.Canceled {
		t.Fatalf("RunWithRuntime() error = %v", err)
	}
	_ = err
}

func TestRun_delegatesToRuntime(t *testing.T) {
	cfg := config.Default()
	ctx, cancel := context.WithCancel(context.Background())

	rt := &Runtime{
		AcquirePID: func(string) (func(), error) { return func() {}, nil },
		NewPoller:  func(context.Context, *config.Config) ClipboardPoller { return newFakePoller("") },
		NewHotkey:  func(config.HotkeyConfig) HotkeyListener { return &blockingHotkey{} },
		NewSpawner: func(config.TerminalConfig) TUISpawner { return &fakeSpawner{} },
		Executable: func() (string, error) { return "/bin/quick-agent", nil },
	}

	done := make(chan error, 1)
	go func() {
		done <- RunWithRuntime(ctx, cfg, rt)
	}()

	time.Sleep(20 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil && err != context.Canceled {
			t.Fatalf("error = %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("RunWithRuntime did not exit after cancel")
	}
}

type blockingHotkey struct{}

func (b *blockingHotkey) Start(ctx context.Context, _ chan<- struct{}) error {
	<-ctx.Done()
	return nil
}
