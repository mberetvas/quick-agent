package main

import (
	"bytes"
	"path/filepath"
	"testing"
)

func TestRootHelp(t *testing.T) {
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
