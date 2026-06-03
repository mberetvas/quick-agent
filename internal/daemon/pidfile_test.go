package daemon

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestWriteReadRemovePIDFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "daemon.pid")

	if err := WritePIDFile(path); err != nil {
		t.Fatalf("WritePIDFile() error = %v", err)
	}

	pid, err := ReadPIDFile(path)
	if err != nil {
		t.Fatalf("ReadPIDFile() error = %v", err)
	}
	if pid != os.Getpid() {
		t.Errorf("ReadPIDFile() = %d, want %d", pid, os.Getpid())
	}

	if err := RemovePIDFile(path); err != nil {
		t.Fatalf("RemovePIDFile() error = %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("pid file still exists after remove: %v", err)
	}
}

func TestAcquirePIDFile_rejects_running_instance(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "daemon.pid")

	release, err := AcquirePIDFile(path)
	if err != nil {
		t.Fatalf("AcquirePIDFile() error = %v", err)
	}
	defer release()

	_, err = AcquirePIDFile(path)
	if err == nil {
		t.Fatal("expected error when another instance holds pid file")
	}
}

func TestAcquirePIDFile_clears_stale_pid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "daemon.pid")

	if err := os.WriteFile(path, []byte("999999"), 0o644); err != nil {
		t.Fatal(err)
	}

	release, err := AcquirePIDFile(path)
	if err != nil {
		t.Fatalf("AcquirePIDFile() with stale pid error = %v", err)
	}
	release()

	pid, err := ReadPIDFile(path)
	if err == nil {
		t.Fatalf("expected removed pid file, read pid %d", pid)
	}
}

func TestStop_not_running(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.pid")
	if err := Stop(path); err == nil {
		t.Fatal("expected error when pid file missing")
	}
}

func TestStop_sends_signal_to_pid(t *testing.T) {
	if os.Getenv("DAEMON_STOP_INTEGRATION") == "" {
		t.Skip("set DAEMON_STOP_INTEGRATION=1 to run process signal test")
	}

	path := filepath.Join(t.TempDir(), "daemon.pid")
	if err := os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
		t.Fatal(err)
	}

	done := make(chan struct{})
	go func() {
		_ = Stop(path)
		close(done)
	}()

	select {
	case <-done:
	case <-t.Context().Done():
		t.Fatal("Stop did not return")
	}
}
