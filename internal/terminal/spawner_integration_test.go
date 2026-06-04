//go:build integration

package terminal

import (
	"context"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/yourname/clipboard-tui/internal/config"
)

func TestSpawner_ResolveProfileOnSystem(t *testing.T) {
	if runtime.GOOS == "linux" && os.Getenv("DISPLAY") == "" {
		t.Skip("DISPLAY not set; terminal resolution may require a display on Linux")
	}

	cfg := config.Default().Terminal
	cfg.Emulator = "auto"

	profile, resolved, err := ResolveProfile(cfg, defaultLookPath)
	if err != nil {
		t.Fatalf("ResolveProfile: %v", err)
	}
	if profile.ID == "" {
		t.Fatal("expected non-empty profile ID")
	}
	if resolved == "" {
		t.Fatal("expected non-empty resolved emulator name")
	}
	t.Logf("resolved profile %q (%s)", profile.ID, resolved)
}

func TestSpawner_SpawnEchoCommand(t *testing.T) {
	if os.Getenv("CLIPBOARD_TUI_SPAWN_INTEGRATION") != "1" {
		t.Skip("set CLIPBOARD_TUI_SPAWN_INTEGRATION=1 to run terminal spawn integration test")
	}
	if runtime.GOOS == "linux" && os.Getenv("DISPLAY") == "" {
		t.Skip("DISPLAY not set")
	}

	cfg := config.Default().Terminal
	cfg.Emulator = "auto"
	spawner := NewSpawner(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	inner := "echo clipboard-tui-spawn-test"
	profileID, err := spawner.Spawn(ctx, inner)
	if err != nil {
		t.Fatalf("Spawn: %v", err)
	}
	if profileID == "" {
		t.Fatal("expected profile ID")
	}
}
