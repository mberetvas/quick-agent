package terminal

import (
	"reflect"
	"testing"
)

func TestOpenFileCommand(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		path     string
		wantCmd  string
		wantArgs []string
	}{
		{
			name:     "windows",
			goos:     "windows",
			path:     `C:\Temp\out.txt`,
			wantCmd:  "cmd",
			wantArgs: []string{"/c", "start", "", `C:\Temp\out.txt`},
		},
		{
			name:     "darwin",
			goos:     "darwin",
			path:     "/tmp/out.txt",
			wantCmd:  "open",
			wantArgs: []string{"/tmp/out.txt"},
		},
		{
			name:     "linux",
			goos:     "linux",
			path:     "/tmp/out.txt",
			wantCmd:  "xdg-open",
			wantArgs: []string{"/tmp/out.txt"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			old := goos
			goos = tt.goos
			defer func() { goos = old }()

			cmd, args, err := openFileCommand(tt.path)
			if err != nil {
				t.Fatalf("openFileCommand() error = %v", err)
			}
			if cmd != tt.wantCmd {
				t.Errorf("cmd = %q, want %q", cmd, tt.wantCmd)
			}
			if !reflect.DeepEqual(args, tt.wantArgs) {
				t.Errorf("args = %v, want %v", args, tt.wantArgs)
			}
		})
	}
}
