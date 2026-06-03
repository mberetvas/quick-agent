//go:build darwin

package terminal

func platformProfiles(lookPath lookPathFunc) []TerminalProfile {
	return []TerminalProfile{
		{
			ID:       "terminal",
			Binaries: []string{"open"},
			BuildLaunch: func(resolvedBin, innerCmd string) (string, []string, error) {
				return resolvedBin, []string{"-a", "Terminal", "--args", "-e", "sh", "-c", innerCmd}, nil
			},
		},
		{
			ID:       "iterm",
			Binaries: []string{"open"},
			BuildLaunch: func(resolvedBin, innerCmd string) (string, []string, error) {
				return resolvedBin, []string{"-a", "iTerm", "--args", "sh", "-c", innerCmd}, nil
			},
		},
	}
}
