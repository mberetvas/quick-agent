package terminal

import (
	"errors"
	"testing"
)

func TestFindOnPath_firstMatch(t *testing.T) {
	path, err := findOnPath([]string{"alpha", "beta"}, func(name string) (string, error) {
		if name == "alpha" {
			return "/usr/bin/alpha", nil
		}
		return "", errors.New("not found")
	})
	if err != nil {
		t.Fatalf("findOnPath() error = %v", err)
	}
	if path != "/usr/bin/alpha" {
		t.Errorf("path = %q, want /usr/bin/alpha", path)
	}
}

func TestFindOnPath_laterMatch(t *testing.T) {
	path, err := findOnPath([]string{"missing", "gamma"}, func(name string) (string, error) {
		if name == "gamma" {
			return "/opt/bin/gamma", nil
		}
		return "", errors.New("not found")
	})
	if err != nil {
		t.Fatalf("findOnPath() error = %v", err)
	}
	if path != "/opt/bin/gamma" {
		t.Errorf("path = %q, want /opt/bin/gamma", path)
	}
}

func TestFindOnPath_noneReturnsErrNoTerminal(t *testing.T) {
	_, err := findOnPath([]string{"a", "b"}, func(string) (string, error) {
		return "", errors.New("not found")
	})
	if !errors.Is(err, ErrNoTerminal) {
		t.Fatalf("error = %v, want ErrNoTerminal", err)
	}
}
