package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func resetCLIState(t *testing.T) {
	t.Helper()
	daemonStop = false
	configPath = ""
	logLevel = "info"
	rootCmd.SetArgs(nil)
}

func TestRootHelp(t *testing.T) {
	resetCLIState(t)
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"--help"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("Execute() error = %v", err)
	}
	if !bytes.Contains(buf.Bytes(), []byte("quick-agent")) {
		t.Fatalf("help output = %q", buf.String())
	}
}

func TestVersionCommand(t *testing.T) {
	rootCmd.SetArgs([]string{"version"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestConfigShow(t *testing.T) {
	dir := t.TempDir()
	configPath = filepath.Join(dir, "config.json")

	rootCmd.SetArgs([]string{"config", "show"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("config show: %v", err)
	}
}

func TestConfigValidate(t *testing.T) {
	dir := t.TempDir()
	configPath = filepath.Join(dir, "config.json")

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"config", "validate"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("config validate: %v", err)
	}
}

func TestDaemonStopNotRunning(t *testing.T) {
	dir := t.TempDir()
	configPath = filepath.Join(dir, "config.json")

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"daemon", "--stop"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when daemon not running")
	}
}

func TestConfigGetKeyMissing(t *testing.T) {
	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"config", "get-key", "openrouter"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatal(err)
	}
}

func TestDebugSpawnTerminalRequiresCommand(t *testing.T) {
	rootCmd.SetArgs([]string{"debug", "spawn-terminal"})
	if err := rootCmd.Execute(); err == nil {
		t.Fatal("expected error when --command missing")
	}
}

func TestDebugSubcommandHelp(t *testing.T) {
	for _, args := range [][]string{
		{"debug", "watch-clipboard", "--help"},
		{"debug", "llm", "--help"},
		{"debug", "hotkey", "--help"},
		{"debug", "spawn-terminal", "--help"},
	} {
		t.Run(args[1], func(t *testing.T) {
			buf := &bytes.Buffer{}
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs(args)
			if err := rootCmd.Execute(); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestConfigInvalidBackend(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "set-key", args: []string{"config", "set-key", "badbackend"}},
		{name: "get-key", args: []string{"config", "get-key", "badbackend"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs(tt.args)
			err := rootCmd.Execute()
			if err == nil {
				t.Fatal("expected error for invalid backend")
			}
			if !strings.Contains(err.Error(), "invalid backend") {
				t.Fatalf("error = %v, want invalid backend", err)
			}
		})
	}
}

func TestConfigSetKeyEmptyStdin(t *testing.T) {
	dir := t.TempDir()
	configPath = filepath.Join(dir, "config.json")

	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdin = r
	_, _ = w.WriteString("\n")
	w.Close()
	t.Cleanup(func() {
		os.Stdin = oldStdin
		r.Close()
	})

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"config", "set-key", "openrouter"})
	err = rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for empty API key")
	}
	if !strings.Contains(err.Error(), "cannot set an empty API key") {
		t.Fatalf("error = %v", err)
	}
}

func TestDaemonBadConfigPath(t *testing.T) {
	resetCLIState(t)
	dir := t.TempDir()
	configPath = filepath.Join(dir, "bad.json")
	if err := os.WriteFile(configPath, []byte("{invalid"), 0o644); err != nil {
		t.Fatal(err)
	}

	buf := &bytes.Buffer{}
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"daemon"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid config")
	}
	if !strings.Contains(err.Error(), "failed to load configuration") {
		t.Fatalf("error = %v", err)
	}
}

func TestConfigShowLogLevelDebug(t *testing.T) {
	resetCLIState(t)
	dir := t.TempDir()
	configPath = filepath.Join(dir, "config.json")

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	t.Cleanup(func() {
		os.Stdout = oldStdout
		_ = r.Close()
	})

	rootCmd.SetArgs([]string{"--log-level", "debug", "config", "show"})
	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("config show: %v", err)
	}
	_ = w.Close()

	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	if !strings.Contains(buf.String(), `"level": "debug"`) {
		t.Fatalf("output = %q, want debug log level", buf.String())
	}
}

func TestSubcommandHelp(t *testing.T) {
	for _, args := range [][]string{
		{"daemon", "--help"},
		{"tui", "--help"},
		{"config", "--help"},
		{"debug", "--help"},
	} {
		t.Run(args[0], func(t *testing.T) {
			buf := &bytes.Buffer{}
			rootCmd.SetOut(buf)
			rootCmd.SetErr(buf)
			rootCmd.SetArgs(args)
			if err := rootCmd.Execute(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
