package config

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestGetConfigDir_platformPaths(t *testing.T) {
	base := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", base)
		want := filepath.Join(base, "quick-agent")
		if got := GetConfigDir(); got != want {
			t.Errorf("GetConfigDir() = %q, want %q", got, want)
		}
	} else {
		t.Setenv("HOME", base)
		want := filepath.Join(base, ".config", "quick-agent")
		if got := GetConfigDir(); got != want {
			t.Errorf("GetConfigDir() = %q, want %q", got, want)
		}
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
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", base)
	} else {
		t.Setenv("HOME", base)
	}
	logCfg := DefaultLoggingConfig()
	if !filepath.IsAbs(logCfg.File) {
		t.Fatalf("log file path not absolute: %q", logCfg.File)
	}
}
