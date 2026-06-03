package terminal

import "os/exec"

func runCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Start()
}

func openFileCommand(path string) (string, []string, error) {
	switch goos {
	case "windows":
		return "cmd", []string{"/c", "start", "", path}, nil
	case "darwin":
		return "open", []string{path}, nil
	default:
		return "xdg-open", []string{path}, nil
	}
}
