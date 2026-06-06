package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestTerminalConfig_validate(t *testing.T) {
	valid := validTerminalEmulators()
	if len(valid) == 0 {
		t.Fatal("validTerminalEmulators() returned empty slice")
	}

	tests := []struct {
		name     string
		emulator string
		wantErr  bool
	}{
		{name: "empty", emulator: "", wantErr: true},
		{name: "auto", emulator: "auto", wantErr: false},
		{name: "valid for platform", emulator: valid[0], wantErr: false},
		{name: "invalid", emulator: "nonexistent-terminal-xyz", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := TerminalConfig{Emulator: tt.emulator}
			err := cfg.validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidTerminalEmulators_matchesGOOS(t *testing.T) {
	emus := validTerminalEmulators()
	if len(emus) == 0 {
		t.Fatal("expected at least one emulator")
	}

	switch runtime.GOOS {
	case "windows":
		if emus[0] != "wt" {
			t.Errorf("windows emulators = %v, want wt first", emus)
		}
	case "darwin":
		if emus[0] != "terminal" {
			t.Errorf("darwin emulators = %v", emus)
		}
	default:
		if emus[0] != "x-terminal-emulator" {
			t.Errorf("linux emulators = %v", emus)
		}
	}
}

func TestCheckPermissions_missingDirOK(t *testing.T) {
	base := t.TempDir()
	if runtime.GOOS == "windows" {
		t.Setenv("APPDATA", base)
	} else {
		t.Setenv("HOME", base)
	}

	if err := CheckPermissions(); err != nil {
		t.Errorf("missing config dir should be ok, got %v", err)
	}
}

func TestCheckPermissions_insecurePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission check skipped on Windows")
	}

	base := t.TempDir()
	t.Setenv("HOME", base)

	configDir := filepath.Join(base, ".quick-agent")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := CheckPermissions(); err == nil {
		t.Fatal("expected insecure permissions error")
	}
}

func TestCheckPermissions_securePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("permission check skipped on Windows")
	}

	base := t.TempDir()
	t.Setenv("HOME", base)

	configDir := filepath.Join(base, ".quick-agent")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatal(err)
	}

	if err := CheckPermissions(); err != nil {
		t.Errorf("secure dir should pass, got %v", err)
	}
}
