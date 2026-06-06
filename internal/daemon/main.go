package daemon

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/mberetvas/quick-agent/internal/clipboard"
	"github.com/mberetvas/quick-agent/internal/config"
	"github.com/mberetvas/quick-agent/internal/hotkey"
	"github.com/mberetvas/quick-agent/internal/terminal"
)

// ClipboardPoller is the subset of clipboard.Poller the daemon loop uses.
type ClipboardPoller interface {
	Start(ctx context.Context)
	LatestText() string
	Changes() <-chan string
	Errors() <-chan error
}

// HotkeyListener registers a global hotkey and emits press events.
type HotkeyListener interface {
	Start(ctx context.Context, events chan<- struct{}) error
}

// TUISpawner launches the TUI in a new terminal window.
type TUISpawner interface {
	SpawnTUI(ctx context.Context, executable, text string) error
}

// Runtime wires dependencies for Run. Nil fields use production defaults.
type Runtime struct {
	Logger     *slog.Logger
	AcquirePID func(string) (func(), error)
	NewPoller  func(ctx context.Context, cfg *config.Config) ClipboardPoller
	NewHotkey  func(cfg config.HotkeyConfig) HotkeyListener
	NewSpawner func(cfg config.TerminalConfig) TUISpawner
	Executable func() (string, error)
}

// Run starts the daemon main loop until ctx is cancelled.
func Run(ctx context.Context, cfg *config.Config) error {
	return RunWithRuntime(ctx, cfg, nil)
}

// RunWithRuntime runs the daemon loop with injectable dependencies (for tests).
func RunWithRuntime(ctx context.Context, cfg *config.Config, rt *Runtime) error {
	if rt == nil {
		rt = &Runtime{}
	}

	log := rt.Logger
	var closeLog func()
	if log == nil {
		var err error
		log, closeLog, err = NewLogger(cfg.Logging)
		if err != nil {
			return err
		}
		defer closeLog()
	}

	acquire := rt.AcquirePID
	if acquire == nil {
		acquire = AcquirePIDFile
	}
	release, err := acquire(cfg.Daemon.PIDFile)
	if err != nil {
		return err
	}
	defer release()

	log.Info("daemon started", "pid_file", cfg.Daemon.PIDFile)

	newPoller := rt.NewPoller
	if newPoller == nil {
		newPoller = func(ctx context.Context, cfg *config.Config) ClipboardPoller {
			p := clipboard.NewPoller(clipboard.SystemClipboard{}, cfg)
			p.Start(ctx)
			return p
		}
	}
	poller := newPoller(ctx, cfg)

	newHotkey := rt.NewHotkey
	if newHotkey == nil {
		newHotkey = func(cfg config.HotkeyConfig) HotkeyListener {
			return hotkey.NewListener(cfg)
		}
	}
	listener := newHotkey(cfg.Hotkey)

	hotkeyEvents := make(chan struct{}, 1)
	hotkeyDone := make(chan error, 1)
	go func() {
		hotkeyDone <- listener.Start(ctx, hotkeyEvents)
	}()

	newSpawner := rt.NewSpawner
	if newSpawner == nil {
		newSpawner = func(cfg config.TerminalConfig) TUISpawner {
			return terminal.NewSpawner(cfg)
		}
	}
	spawner := newSpawner(cfg.Terminal)

	exeFn := rt.Executable
	if exeFn == nil {
		exeFn = os.Executable
	}

	for {
		select {
		case <-ctx.Done():
			log.Info("daemon shutting down")
			if err := <-hotkeyDone; err != nil && !errors.Is(err, context.Canceled) {
				return fmt.Errorf("hotkey listener: %w", err)
			}
			return ctx.Err()
		case err := <-hotkeyDone:
			if err != nil {
				return fmt.Errorf("hotkey listener: %w", err)
			}
			return nil
		case text := <-poller.Changes():
			log.Debug("clipboard changed", "length", len(text))
		case err := <-poller.Errors():
			if err != nil {
				log.Error("clipboard poll error", "error", err)
			}
		case <-hotkeyEvents:
			handleHotkey(ctx, log, poller, spawner, exeFn)
		}
	}
}

func handleHotkey(
	ctx context.Context,
	log *slog.Logger,
	poller ClipboardPoller,
	spawner TUISpawner,
	exeFn func() (string, error),
) {
	text := poller.LatestText()
	if text == "" {
		log.Debug("hotkey pressed with empty clipboard, skipping spawn")
		return
	}

	exe, err := exeFn()
	if err != nil {
		log.Error("resolve executable", "error", err)
		return
	}

	spawnErr := spawner.SpawnTUI(ctx, exe, text)
	if errors.Is(spawnErr, terminal.ErrUsedFallback) {
		log.Info("TUI spawn used fallback", "error", spawnErr)
		return
	}
	if spawnErr != nil {
		log.Error("failed to spawn TUI", "error", spawnErr)
		return
	}
	log.Info("spawned TUI", "length", len(text))
}
