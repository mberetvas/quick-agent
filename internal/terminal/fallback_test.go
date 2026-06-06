package terminal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFallbackOutput_writeFailure(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, err := FallbackOutput(blocker, "hello", nil)
	if err == nil {
		t.Fatal("expected error when output dir path is an existing file")
	}
}

func TestFallbackOutput_openerFailure(t *testing.T) {
	dir := t.TempDir()
	path, err := FallbackOutput(dir, "hello", func(string) error {
		return os.ErrPermission
	})
	if err == nil {
		t.Fatal("expected opener error")
	}
	if path == "" {
		t.Fatal("expected path even when opener fails")
	}
}

func TestFallbackOutput_writes_and_opens(t *testing.T) {
	dir := t.TempDir()
	var opened string

	path, err := FallbackOutput(dir, "hello clipboard", func(p string) error {
		opened = p
		return nil
	})
	if err != nil {
		t.Fatalf("FallbackOutput() error = %v", err)
	}
	if opened != path {
		t.Fatalf("opened %q, path %q", opened, path)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello clipboard" {
		t.Errorf("file content = %q", string(data))
	}
	if !strings.HasPrefix(filepath.Base(path), "clipboard-") {
		t.Errorf("unexpected filename %q", filepath.Base(path))
	}
}
