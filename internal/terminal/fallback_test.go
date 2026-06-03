package terminal

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
