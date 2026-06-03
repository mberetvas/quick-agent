//go:build windows

package terminal

import (
	"fmt"
	"strings"
)

func platformProfiles(lookPath lookPathFunc) []TerminalProfile {
	return []TerminalProfile{
		{
			ID:       "wt",
			Binaries: []string{"wt.exe"},
			BuildLaunch: func(resolvedBin, innerCmd string) (string, []string, error) {
				if pwsh, err := lookPath("pwsh.exe"); err == nil {
					return resolvedBin, []string{"new-tab", "--", pwsh, "-NoExit", "-Command", innerCmd}, nil
				}
				if powershell, err := lookPath("powershell.exe"); err == nil {
					return resolvedBin, []string{"new-tab", "--", powershell, "-NoExit", "-Command", innerCmd}, nil
				}
				return resolvedBin, []string{"new-tab", "cmd", "/k", innerCmd}, nil
			},
		},
		{
			ID:       "powershell",
			Binaries: []string{"powershell.exe"},
			BuildLaunch: func(resolvedBin, innerCmd string) (string, []string, error) {
				script := fmt.Sprintf("Start-Process -FilePath %s -ArgumentList '-NoExit','-Command',%s",
					quotePowerShellString(resolvedBin),
					quotePowerShellString(innerCmd),
				)
				return resolvedBin, []string{"-NoExit", "-Command", script}, nil
			},
		},
		{
			ID:       "cmd",
			Binaries: []string{"cmd.exe"},
			BuildLaunch: func(resolvedBin, innerCmd string) (string, []string, error) {
				return resolvedBin, []string{"/c", "start", "cmd", "/k", innerCmd}, nil
			},
		},
	}
}

func quotePowerShellString(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
