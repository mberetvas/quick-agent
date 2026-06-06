package terminal

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFallbackOutput_emptyDirUsesTempDefault(t *testing.T) {
	path, err := FallbackOutput("", "fallback body", func(p string) error {
		if _, statErr := os.Stat(p); statErr != nil {
			t.Fatalf("file not written: %v", statErr)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if !filepath.IsAbs(path) {
		t.Errorf("path = %q, want absolute", path)
	}
}
