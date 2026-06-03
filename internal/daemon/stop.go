package daemon

import (
	"fmt"
	"os"
	"syscall"
)

// Stop signals the daemon process recorded in pidFile to shut down.
func Stop(pidFile string) error {
	pid, err := ReadPIDFile(pidFile)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("daemon is not running")
		}
		return fmt.Errorf("read pid file: %w", err)
	}

	if !processAlive(pid) {
		_ = RemovePIDFile(pidFile)
		return fmt.Errorf("daemon is not running")
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("find process: %w", err)
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("signal daemon: %w", err)
	}
	return nil
}
