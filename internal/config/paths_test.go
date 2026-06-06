package config

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetConfigDir_platformPaths(t *testing.T) {
	base := t.TempDir()
	t.Setenv("HOME", base)
	t.Setenv("USERPROFILE", base)
	want := filepath.Join(base, ".quick-agent")
	if got := GetConfigDir(); got != want {
		t.Errorf("GetConfigDir() = %q, want %q", got, want)
	}
}

func TestDefaultHotkeyConfig_matchesPlatform(t *testing.T) {
	hk := DefaultHotkeyConfig()
	if runtime.GOOS == "darwin" {
		if len(hk.Modifiers) != 2 || hk.Modifiers[0] != "cmd" {
			t.Errorf("darwin hotkey = %+v", hk)
		}
	} else if hk.Modifiers[0] != "ctrl" {
		t.Errorf("default hotkey = %+v", hk)
	}
}

func TestDefaultLoggingConfig_usesConfigDir(t *testing.T) {
	base := t.TempDir()
	t.Setenv("HOME", base)
	t.Setenv("USERPROFILE", base)
	logCfg := DefaultLoggingConfig()
	want := filepath.Join(base, ".quick-agent", "quick-agent.log")
	if logCfg.File != want {
		t.Errorf("log file = %q, want %q", logCfg.File, want)
	}
}
