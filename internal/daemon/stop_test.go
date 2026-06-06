package daemon

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStop_missingPIDFile(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "daemon.pid")

	err := Stop(pidFile)
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "not running") {
		t.Errorf("error = %q, want 'not running'", err.Error())
	}
}

func TestStop_invalidPIDFile(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "daemon.pid")
	if err := os.WriteFile(pidFile, []byte("not-a-pid"), 0644); err != nil {
		t.Fatal(err)
	}

	err := Stop(pidFile)
	if err == nil {
		t.Fatal("expected error for invalid pid file")
	}
	if !strings.Contains(err.Error(), "pid") {
		t.Errorf("error = %q", err.Error())
	}
}

func TestStop_stalePIDRemovesFile(t *testing.T) {
	pidFile := filepath.Join(t.TempDir(), "daemon.pid")
	stalePID := "999999999"
	if err := os.WriteFile(pidFile, []byte(stalePID), 0644); err != nil {
		t.Fatal(err)
	}

	err := Stop(pidFile)
	if err == nil {
		t.Fatal("expected not running error")
	}
	if !strings.Contains(err.Error(), "not running") {
		t.Errorf("error = %q", err.Error())
	}
	if _, statErr := os.Stat(pidFile); !os.IsNotExist(statErr) {
		t.Error("stale pid file should be removed")
	}
}
