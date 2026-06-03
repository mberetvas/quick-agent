package terminal

import (
	"context"
	"errors"
	"testing"

	"github.com/yourname/clipboard-tui/internal/config"
)

func TestResolveProfile_auto_first_available(t *testing.T) {
	lookPath := mockLookPathPreferLastProfile(t)
	cfg := config.TerminalConfig{Emulator: "auto"}

	profile, resolved, err := ResolveProfile(cfg, lookPath.fn)
	if err != nil {
		t.Fatalf("ResolveProfile() error = %v", err)
	}
	if profile.ID != lookPath.wantID {
		t.Fatalf("profile ID = %q, want %q", profile.ID, lookPath.wantID)
	}
	if resolved == "" {
		t.Fatal("expected resolved binary path")
	}
}

type mockLookPathState struct {
	fn     lookPathFunc
	wantID string
}

func mockLookPathPreferLastProfile(t *testing.T) mockLookPathState {
	t.Helper()
	profiles := platformProfiles(func(string) (string, error) { return "", errors.New("missing") })
	if len(profiles) == 0 {
		t.Fatal("no platform profiles")
	}
	want := profiles[len(profiles)-1].ID
	lastBin := profiles[len(profiles)-1].Binaries[len(profiles[len(profiles)-1].Binaries)-1]

	fn := func(name string) (string, error) {
		if name == lastBin {
			return "/mock/" + name, nil
		}
		return "", errors.New("missing")
	}
	return mockLookPathState{fn: fn, wantID: want}
}

func TestResolveProfile_explicit_missing_binary(t *testing.T) {
	lookPath := func(string) (string, error) {
		return "", errors.New("not found")
	}
	profiles := platformProfiles(lookPath)
	cfg := config.TerminalConfig{Emulator: profiles[0].ID}

	_, _, err := ResolveProfile(cfg, lookPath)
	if !errors.Is(err, ErrNoTerminal) {
		t.Fatalf("ResolveProfile() error = %v, want ErrNoTerminal", err)
	}
}

func TestSpawner_Spawn_starts_command(t *testing.T) {
	state := mockLookPathPreferLastProfile(t)
	var gotName string
	var gotArgs []string

	s := NewSpawner(config.TerminalConfig{Emulator: state.wantID})
	s.lookPath = state.fn
	s.startCmd = func(ctx context.Context, name string, args ...string) error {
		gotName = name
		gotArgs = append([]string(nil), args...)
		return nil
	}

	id, err := s.Spawn(context.Background(), "echo hello")
	if err != nil {
		t.Fatalf("Spawn() error = %v", err)
	}
	if id != state.wantID {
		t.Fatalf("profile id = %q, want %q", id, state.wantID)
	}
	if gotName == "" {
		t.Fatal("expected command name")
	}
	if len(gotArgs) == 0 {
		t.Fatal("expected args")
	}
}

func TestSpawner_SpawnTUI_uses_fallback_when_no_terminal(t *testing.T) {
	dir := t.TempDir()
	var opened string

	s := NewSpawner(config.TerminalConfig{
		Emulator:    "wt",
		FallbackDir: dir,
	})
	if goos != "windows" {
		profiles := platformProfiles(func(string) (string, error) { return "", errors.New("missing") })
		s.cfg.Emulator = profiles[0].ID
	}
	s.lookPath = func(string) (string, error) {
		return "", errors.New("missing")
	}
	s.opener = func(path string) error {
		opened = path
		return nil
	}

	err := s.SpawnTUI(context.Background(), "/bin/clipboard-tui", "clip text")
	if !errors.Is(err, ErrUsedFallback) {
		t.Fatalf("SpawnTUI() error = %v, want ErrUsedFallback", err)
	}
	if opened == "" {
		t.Fatal("expected fallback file to be opened")
	}
}

func TestBuildLaunch_last_profile(t *testing.T) {
	state := mockLookPathPreferLastProfile(t)
	profile, err := ProfileByID(state.wantID, state.fn)
	if err != nil {
		t.Fatal(err)
	}
	name, args, err := profile.BuildLaunch("/mock/bin", "echo hello")
	if err != nil {
		t.Fatal(err)
	}
	if name == "" {
		t.Fatal("expected launcher name")
	}
	if len(args) == 0 {
		t.Fatal("expected args")
	}
	last := args[len(args)-1]
	if last != "echo hello" && !containsArg(args, "echo hello") {
		t.Errorf("args = %v, expected inner command echo hello", args)
	}
}

func containsArg(args []string, want string) bool {
	for _, a := range args {
		if a == want {
			return true
		}
	}
	return false
}
