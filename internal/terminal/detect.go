package terminal

import "os/exec"

type lookPathFunc func(string) (string, error)

func defaultLookPath(name string) (string, error) {
	return exec.LookPath(name)
}

func findOnPath(binaries []string, lookPath lookPathFunc) (string, error) {
	for _, name := range binaries {
		path, err := lookPath(name)
		if err == nil {
			return path, nil
		}
	}
	return "", ErrNoTerminal
}
