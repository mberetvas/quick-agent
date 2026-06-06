//go:build darwin

package terminal

const (
	terminalAppBin = "/Applications/Utilities/Terminal.app/Contents/MacOS/Terminal"
	itermAppBin    = "/Applications/iTerm.app/Contents/MacOS/iTerm2"
)

func platformProfiles(lookPath lookPathFunc) []TerminalProfile {
	return []TerminalProfile{
		{
			ID:       "terminal",
			Binaries: []string{terminalAppBin},
			BuildLaunch: func(resolvedBin, innerCmd string) (string, []string, error) {
				openBin, err := lookPath("open")
				if err != nil {
					return "", nil, err
				}
				return openBin, []string{"-a", "Terminal", "--args", "-e", "sh", "-c", innerCmd}, nil
			},
		},
		{
			ID:       "iterm",
			Binaries: []string{itermAppBin},
			BuildLaunch: func(resolvedBin, innerCmd string) (string, []string, error) {
				openBin, err := lookPath("open")
				if err != nil {
					return "", nil, err
				}
				return openBin, []string{"-a", "iTerm", "--args", "sh", "-c", innerCmd}, nil
			},
		},
	}
}
