package terminal

import "testing"

func TestQuotePOSIXSingle(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "simple",
			input: "echo hello",
			want:  "'echo hello'",
		},
		{
			name:  "empty",
			input: "",
			want:  "''",
		},
		{
			name:  "single quote",
			input: "it's fine",
			want:  "'it'\\''s fine'",
		},
		{
			name:  "newline",
			input: "line1\nline2",
			want:  "'line1\nline2'",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := QuotePOSIXSingle(tt.input)
			if got != tt.want {
				t.Errorf("QuotePOSIXSingle() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBuildTUIInnerCommand_unix(t *testing.T) {
	if goos == "windows" {
		t.Skip("unix quoting")
	}
	got := BuildTUIInnerCommand("/usr/bin/clipboard-tui", "/tmp/input.txt")
	want := "'/usr/bin/clipboard-tui' tui < '/tmp/input.txt'"
	if got != want {
		t.Errorf("BuildTUIInnerCommand() = %q, want %q", got, want)
	}
}

func TestBuildTUIInnerCommand_windows(t *testing.T) {
	if goos != "windows" {
		t.Skip("windows quoting")
	}
	got := BuildTUIInnerCommand(`C:\Program Files\clipboard-tui.exe`, `C:\Temp Files\input.txt`)
	want := `cmd /c "\"C:\Program Files\clipboard-tui.exe\" tui < \"C:\Temp Files\input.txt\""`
	if got != want {
		t.Errorf("BuildTUIInnerCommand() = %q, want %q", got, want)
	}
}

func TestBuildTUIInnerCommand_windows_no_spaces(t *testing.T) {
	if goos != "windows" {
		t.Skip("windows quoting")
	}
	got := BuildTUIInnerCommand(`C:\bin\clipboard-tui.exe`, `C:\Temp\input.txt`)
	want := `cmd /c "C:\bin\clipboard-tui.exe tui < C:\Temp\input.txt"`
	if got != want {
		t.Errorf("BuildTUIInnerCommand() = %q, want %q", got, want)
	}
}

func TestBuildDebugInnerCommand_windows(t *testing.T) {
	if goos != "windows" {
		t.Skip("windows only")
	}
	got := BuildDebugInnerCommand("echo hello")
	want := "cmd /c echo hello"
	if got != want {
		t.Errorf("BuildDebugInnerCommand() = %q, want %q", got, want)
	}
}

func TestBuildDebugInnerCommand_unix(t *testing.T) {
	if goos == "windows" {
		t.Skip("unix only")
	}
	got := BuildDebugInnerCommand("echo hello")
	want := "'echo hello'"
	if got != "sh -c "+want {
		t.Errorf("BuildDebugInnerCommand() = %q, want sh -c %q", got, want)
	}
}

func TestBuildDebugInnerCommand_unix_singleQuote(t *testing.T) {
	if goos == "windows" {
		t.Skip("unix only")
	}
	got := BuildDebugInnerCommand("it's fine")
	want := "sh -c " + QuotePOSIXSingle("it's fine")
	if got != want {
		t.Errorf("BuildDebugInnerCommand() = %q, want %q", got, want)
	}
}
