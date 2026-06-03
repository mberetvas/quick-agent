package terminal

import (
	"errors"
	"fmt"

	"github.com/yourname/clipboard-tui/internal/config"
)

// TerminalProfile describes a terminal emulator launcher on the current OS.
type TerminalProfile struct {
	ID          string
	Binaries    []string
	BuildLaunch func(resolvedBin, innerCmd string) (name string, args []string, err error)
}

// ResolveProfile picks a profile from cfg and verifies its launcher exists on PATH.
func ResolveProfile(cfg config.TerminalConfig, lookPath lookPathFunc) (*TerminalProfile, string, error) {
	if lookPath == nil {
		lookPath = defaultLookPath
	}

	profiles := platformProfiles(lookPath)

	if cfg.Emulator == "auto" {
		for i := range profiles {
			p := &profiles[i]
			resolved, err := findOnPath(p.Binaries, lookPath)
			if err != nil {
				continue
			}
			return p, resolved, nil
		}
		return nil, "", ErrNoTerminal
	}

	for i := range profiles {
		p := &profiles[i]
		if p.ID != cfg.Emulator {
			continue
		}
		resolved, err := findOnPath(p.Binaries, lookPath)
		if err != nil {
			return nil, "", fmt.Errorf("%w: profile %q", ErrNoTerminal, cfg.Emulator)
		}
		return p, resolved, nil
	}

	return nil, "", fmt.Errorf("%w: unknown profile %q", ErrNoTerminal, cfg.Emulator)
}

// ProfileByID returns a profile for tests and golden argv checks.
func ProfileByID(id string, lookPath lookPathFunc) (*TerminalProfile, error) {
	if lookPath == nil {
		lookPath = defaultLookPath
	}
	for _, p := range platformProfiles(lookPath) {
		if p.ID == id {
			cp := p
			return &cp, nil
		}
	}
	return nil, errors.New("profile not found")
}
