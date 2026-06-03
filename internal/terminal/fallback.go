package terminal

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type fileOpener func(path string) error

// FallbackOutput writes text to dir and opens it with the OS default handler.
func FallbackOutput(dir, text string, opener fileOpener) (string, error) {
	if dir == "" {
		dir = filepath.Join(os.TempDir(), "clipboard-tui-output")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}

	name := fmt.Sprintf("clipboard-%d.txt", time.Now().Unix())
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		return "", err
	}

	if opener == nil {
		opener = defaultFileOpener
	}
	if err := opener(path); err != nil {
		return path, err
	}
	return path, nil
}

func defaultFileOpener(path string) error {
	name, args, err := openFileCommand(path)
	if err != nil {
		return err
	}
	return runCommand(name, args...)
}
