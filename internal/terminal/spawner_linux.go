//go:build linux

package terminal

func platformProfiles(lookPath lookPathFunc) []TerminalProfile {
	return []TerminalProfile{
		{
			ID:       "x-terminal-emulator",
			Binaries: []string{"x-terminal-emulator"},
			BuildLaunch: func(resolvedBin, innerCmd string) (string, []string, error) {
				return resolvedBin, []string{"-e", "sh", "-c", innerCmd}, nil
			},
		},
		{
			ID:       "gnome-terminal",
			Binaries: []string{"gnome-terminal"},
			BuildLaunch: func(resolvedBin, innerCmd string) (string, []string, error) {
				return resolvedBin, []string{"--", "sh", "-c", innerCmd}, nil
			},
		},
		{
			ID:       "konsole",
			Binaries: []string{"konsole"},
			BuildLaunch: func(resolvedBin, innerCmd string) (string, []string, error) {
				return resolvedBin, []string{"-e", "sh", "-c", innerCmd}, nil
			},
		},
		{
			ID:       "xfce4-terminal",
			Binaries: []string{"xfce4-terminal"},
			BuildLaunch: func(resolvedBin, innerCmd string) (string, []string, error) {
				return resolvedBin, []string{"-e", "sh", "-c", innerCmd}, nil
			},
		},
		{
			ID:       "alacritty",
			Binaries: []string{"alacritty"},
			BuildLaunch: func(resolvedBin, innerCmd string) (string, []string, error) {
				return resolvedBin, []string{"-e", "sh", "-c", innerCmd}, nil
			},
		},
		{
			ID:       "kitty",
			Binaries: []string{"kitty"},
			BuildLaunch: func(resolvedBin, innerCmd string) (string, []string, error) {
				return resolvedBin, []string{"sh", "-c", innerCmd}, nil
			},
		},
	}
}
