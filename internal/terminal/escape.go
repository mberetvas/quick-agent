package terminal

import "strings"

// QuotePOSIXSingle quotes s for use inside a single-quoted POSIX sh -c argument.
func QuotePOSIXSingle(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// BuildDebugInnerCommand wraps a user shell one-liner for execution inside a new terminal.
func BuildDebugInnerCommand(command string) string {
	switch goos {
	case "windows":
		return "cmd /c " + command
	default:
		return "sh -c " + QuotePOSIXSingle(command)
	}
}

// BuildTUIInnerCommand builds the command that runs the TUI with stdin from a temp file.
func BuildTUIInnerCommand(executable, tempPath string) string {
	switch goos {
	case "windows":
		// Redirection is cmd.exe syntax; wrap so wt/powershell hosts run it correctly.
		line := quoteWindowsCmdArg(executable) + " tui < " + quoteWindowsCmdArg(tempPath)
		return "cmd /c " + quoteWindowsCmdCArg(line)
	default:
		return QuotePOSIXSingle(executable) + " tui < " + QuotePOSIXSingle(tempPath)
	}
}

func quoteWindowsCmdArg(s string) string {
	if strings.ContainsAny(s, " \t\"") {
		return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}
	return s
}

// quoteWindowsCmdCArg quotes the entire command line passed to cmd /c.
func quoteWindowsCmdCArg(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}
