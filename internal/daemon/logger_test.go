package daemon

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/mberetvas/quick-agent/internal/config"
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

func TestNewLogger_consoleAndDiscard(t *testing.T) {
	t.Run("console", func(t *testing.T) {
		log, closeFn, err := NewLogger(config.LoggingConfig{Level: "debug", Console: true})
		if err != nil {
			t.Fatal(err)
		}
		defer closeFn()
		log.Debug("console only")
	})

	t.Run("discard when no outputs", func(t *testing.T) {
		log, closeFn, err := NewLogger(config.LoggingConfig{Level: "warn"})
		if err != nil {
			t.Fatal(err)
		}
		defer closeFn()
		log.Warn("discarded")
	})
}

func TestNewLogger_invalidLevel(t *testing.T) {
	_, _, err := NewLogger(config.LoggingConfig{Level: "verbose"})
	if err == nil {
		t.Fatal("expected error for invalid log level")
	}
}

func TestParseLogLevel_allLevels(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"debug", "DEBUG"},
		{"info", "INFO"},
		{"warn", "WARN"},
		{"error", "ERROR"},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			level, err := parseLogLevel(tt.in)
			if err != nil {
				t.Fatal(err)
			}
			if level.String() != tt.want {
				t.Errorf("level = %s, want %s", level.String(), tt.want)
			}
		})
	}
}

func TestNewLogger_createsParentDirectory(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "nested", "logs", "daemon.log")

	log, closeFn, err := NewLogger(config.LoggingConfig{
		Level: "info",
		File:  logPath,
	})
	if err != nil {
		t.Fatal(err)
	}
	log.Info("nested log path")
	closeFn()

	if _, err := os.Stat(logPath); err != nil {
		t.Fatalf("log file not created: %v", err)
	}
}
