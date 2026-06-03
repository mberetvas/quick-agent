package daemon

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/yourname/clipboard-tui/internal/config"
)

func TestNewLogger_writes_to_file(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "daemon.log")
	cfg := config.LoggingConfig{
		Level:   "info",
		File:    logPath,
		Console: false,
	}

	log, closeFn, err := NewLogger(cfg)
	if err != nil {
		t.Fatalf("NewLogger() error = %v", err)
	}
	defer closeFn()

	log.Info("daemon test message")
	closeFn()

	data, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("read log file: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected log output in file")
	}
}
