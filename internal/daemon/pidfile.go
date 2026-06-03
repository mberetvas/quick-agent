package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// WritePIDFile writes the current process ID to path.
func WritePIDFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	pid := os.Getpid()
	return os.WriteFile(path, []byte(strconv.Itoa(pid)), 0o644)
}

// ReadPIDFile reads a PID from path.
func ReadPIDFile(path string) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, fmt.Errorf("invalid pid in %s: %w", path, err)
	}
	return pid, nil
}

// RemovePIDFile deletes the PID file at path.
func RemovePIDFile(path string) error {
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// IsRunning reports whether the PID in path refers to a live process.
func IsRunning(path string) (bool, error) {
	pid, err := ReadPIDFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return processAlive(pid), nil
}

// AcquirePIDFile ensures no other live instance is running, then writes this process PID.
// The returned release function removes the PID file.
func AcquirePIDFile(path string) (func(), error) {
	running, err := IsRunning(path)
	if err != nil {
		return nil, fmt.Errorf("check running instance: %w", err)
	}
	if running {
		return nil, fmt.Errorf("another instance is already running")
	}
	if err := RemovePIDFile(path); err != nil {
		return nil, err
	}
	if err := WritePIDFile(path); err != nil {
		return nil, fmt.Errorf("write pid file: %w", err)
	}
	return func() { _ = RemovePIDFile(path) }, nil
}
