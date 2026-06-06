//go:build !windows

package daemon

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"testing"
	"time"
)

func TestStop_signalsLiveProcess(t *testing.T) {
	cmd := exec.Command("sleep", "10")
	if err := cmd.Start(); err != nil {
		t.Fatalf("start sleep: %v", err)
	}
	defer func() {
		_ = cmd.Process.Kill()
	}()

	pidFile := filepath.Join(t.TempDir(), "daemon.pid")
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(cmd.Process.Pid)), 0644); err != nil {
		t.Fatal(err)
	}

	if err := Stop(pidFile); err != nil {
		t.Fatalf("Stop() error = %v", err)
	}

	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()

	select {
	case err := <-done:
		if err == nil {
			t.Log("process exited cleanly after SIGTERM")
		}
	case <-time.After(5 * time.Second):
		t.Fatal("process did not exit after SIGTERM")
	}
}
