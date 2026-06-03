//go:build windows

package terminal

import (
	"strings"
	"testing"

	"github.com/yourname/clipboard-tui/internal/config"
)

func mockWindowsLookPath(t *testing.T) lookPathFunc {
	t.Helper()
	return func(name string) (string, error) {
		switch name {
		case "wt.exe":
			return `C:\Windows\System32\wt.exe`, nil
		case "pwsh.exe":
			return `C:\Program Files\PowerShell\7\pwsh.exe`, nil
		case "powershell.exe":
			return `C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`, nil
		case "cmd.exe":
			return `C:\Windows\System32\cmd.exe`, nil
		default:
			return "", ErrNoTerminal
		}
	}
}

func TestBuildLaunch_wt_uses_pwsh(t *testing.T) {
	lookPath := mockWindowsLookPath(t)
	p, err := ProfileByID("wt", lookPath)
	if err != nil {
		t.Fatal(err)
	}

	name, args, err := p.BuildLaunch(`C:\Windows\System32\wt.exe`, "echo hello")
	if err != nil {
		t.Fatal(err)
	}
	if name != `C:\Windows\System32\wt.exe` {
		t.Errorf("name = %q", name)
	}
	if len(args) < 4 || args[0] != "new-tab" || args[2] != `C:\Program Files\PowerShell\7\pwsh.exe` {
		t.Fatalf("args = %v, want new-tab with pwsh", args)
	}
	if args[len(args)-1] != "echo hello" {
		t.Errorf("last arg = %q, want echo hello", args[len(args)-1])
	}
}

func TestBuildLaunch_wt_falls_back_to_cmd_shell(t *testing.T) {
	lookPath := func(name string) (string, error) {
		switch name {
		case "wt.exe":
			return `C:\Windows\System32\wt.exe`, nil
		case "cmd.exe":
			return `C:\Windows\System32\cmd.exe`, nil
		default:
			return "", ErrNoTerminal
		}
	}
	p, err := ProfileByID("wt", lookPath)
	if err != nil {
		t.Fatal(err)
	}

	_, args, err := p.BuildLaunch(`C:\Windows\System32\wt.exe`, "echo hello")
	if err != nil {
		t.Fatal(err)
	}
	if len(args) < 3 || args[1] != "cmd" {
		t.Fatalf("args = %v, want cmd fallback when pwsh missing", args)
	}
}

func TestBuildLaunch_powershell(t *testing.T) {
	lookPath := mockWindowsLookPath(t)
	p, err := ProfileByID("powershell", lookPath)
	if err != nil {
		t.Fatal(err)
	}

	name, args, err := p.BuildLaunch(`C:\Windows\System32\WindowsPowerShell\v1.0\powershell.exe`, "echo hello")
	if err != nil {
		t.Fatal(err)
	}
	if name == "" || len(args) == 0 {
		t.Fatal("expected launcher args")
	}
	joined := strings.Join(args, " ")
	if !strings.Contains(joined, "Start-Process") {
		t.Errorf("args should use Start-Process, got %v", args)
	}
}

func TestBuildLaunch_cmd(t *testing.T) {
	lookPath := mockWindowsLookPath(t)
	p, err := ProfileByID("cmd", lookPath)
	if err != nil {
		t.Fatal(err)
	}

	name, args, err := p.BuildLaunch(`C:\Windows\System32\cmd.exe`, "echo hello")
	if err != nil {
		t.Fatal(err)
	}
	if name != `C:\Windows\System32\cmd.exe` {
		t.Errorf("name = %q", name)
	}
	want := []string{"/c", "start", "cmd", "/k", "echo hello"}
	if len(args) != len(want) {
		t.Fatalf("args = %v, want %v", args, want)
	}
	for i := range want {
		if args[i] != want[i] {
			t.Fatalf("args[%d] = %q, want %q", i, args[i], want[i])
		}
	}
}

func TestResolveProfile_auto_prefers_wt(t *testing.T) {
	lookPath := mockWindowsLookPath(t)
	profile, _, err := ResolveProfile(config.TerminalConfig{Emulator: "auto"}, lookPath)
	if err != nil {
		t.Fatal(err)
	}
	if profile.ID != "wt" {
		t.Errorf("profile ID = %q, want wt", profile.ID)
	}
}
