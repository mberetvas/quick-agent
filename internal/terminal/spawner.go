package terminal

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/yourname/clipboard-tui/internal/config"
)

var (
	// ErrNoTerminal is returned when no terminal emulator can be resolved.
	ErrNoTerminal = errors.New("no terminal emulator found")
	// ErrUsedFallback indicates SpawnTUI opened a fallback file instead of a terminal.
	ErrUsedFallback = errors.New("terminal spawn failed; opened fallback file")
)

type commandStarter func(ctx context.Context, name string, args ...string) error

// Spawner launches commands in a new terminal window.
type Spawner struct {
	cfg      config.TerminalConfig
	lookPath lookPathFunc
	startCmd commandStarter
	opener   fileOpener
}

// NewSpawner creates a Spawner for the given terminal configuration.
func NewSpawner(cfg config.TerminalConfig) *Spawner {
	return &Spawner{
		cfg:      cfg,
		lookPath: defaultLookPath,
		startCmd: defaultStartCmd,
		opener:   defaultFileOpener,
	}
}

func defaultStartCmd(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	return cmd.Start()
}

// Spawn launches innerCmd in a new terminal window via the resolved profile.
// Returns the profile ID used. Does not use fallback on error.
func (s *Spawner) Spawn(ctx context.Context, innerCmd string) (string, error) {
	profile, resolved, err := ResolveProfile(s.cfg, s.lookPath)
	if err != nil {
		return "", err
	}

	name, args, err := profile.BuildLaunch(resolved, innerCmd)
	if err != nil {
		return "", err
	}

	if err := s.startCmd(ctx, name, args...); err != nil {
		return "", err
	}
	return profile.ID, nil
}

// SpawnTUI launches executable with clipboard text via a secure temp file redirect.
// On failure, writes fallback output and returns ErrUsedFallback.
func (s *Spawner) SpawnTUI(ctx context.Context, executable, text string) error {
	tempFile, err := os.CreateTemp("", "clipboard-tui-input-*.txt")
	if err != nil {
		return s.useFallback(text, err)
	}
	tempPath := tempFile.Name()

	if _, err := tempFile.WriteString(text); err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempPath)
		return s.useFallback(text, err)
	}
	if err := tempFile.Close(); err != nil {
		_ = os.Remove(tempPath)
		return s.useFallback(text, err)
	}
	if err := os.Chmod(tempPath, 0o600); err != nil {
		_ = os.Remove(tempPath)
		return s.useFallback(text, err)
	}

	inner := BuildTUIInnerCommand(executable, tempPath)

	profile, resolved, err := ResolveProfile(s.cfg, s.lookPath)
	if err != nil {
		return s.useFallback(text, err)
	}

	name, args, err := profile.BuildLaunch(resolved, inner)
	if err != nil {
		return s.useFallback(text, err)
	}

	if err := s.startCmd(ctx, name, args...); err != nil {
		return s.useFallback(text, err)
	}
	return nil
}

func (s *Spawner) useFallback(text string, originalErr error) error {
	dir := s.cfg.FallbackDir
	if dir == "" {
		dir = config.DefaultTerminalConfig().FallbackDir
	}
	if _, err := FallbackOutput(dir, text, s.opener); err != nil {
		return fmt.Errorf("%w: fallback failed: %v (original: %v)", ErrUsedFallback, err, originalErr)
	}
	return fmt.Errorf("%w: %v", ErrUsedFallback, originalErr)
}
