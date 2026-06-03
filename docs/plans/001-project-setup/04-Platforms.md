# 04 - Platform-Specific Details

## Overview

This document details how the application handles platform differences for Windows, macOS, and Linux. Each platform has unique requirements for clipboard access, hotkey detection, terminal spawning, and auto-start configuration.

---

## Platform Detection

```go
// internal/platform/platform.go
package platform

import "runtime"

const (
    Windows = "windows"
    Darwin  = "darwin"
    Linux   = "linux"
)

// OS returns the current operating system
func OS() string {
    return runtime.GOOS
}

// IsWindows returns true if running on Windows
func IsWindows() bool {
    return OS() == Windows
}

// IsMac returns true if running on macOS
func IsMac() bool {
    return OS() == Darwin
}

// IsLinux returns true if running on Linux
func IsLinux() bool {
    return OS() == Linux
}
```

---

## Clipboard Access

### Library: github.com/atotto/clipboard

This library provides cross-platform clipboard access. It handles the platform differences internally.

| Platform | Implementation | Notes |
|----------|----------------|-------|
| Windows | Win32 API | Works with standard clipboard |
| macOS | NSPasteboard | Requires GUI context |
| Linux | X11 | Requires X11 server, not Wayland |

### Wayland Support (Linux)

**Issue**: The `atotto/clipboard` library only supports X11, not Wayland.

**Solutions**:
1. **Recommended for v1**: Document that Wayland is not supported, require X11
2. **v2**: Add Wayland support using `wl-clipboard` CLI tool
3. **v2**: Use `github.com/linusg/skus` which supports both X11 and Wayland

```go
// internal/clipboard/wayland.go (v2)
package clipboard

import (
    "os/exec"
)

// ReadAll attempts to read clipboard, falls back to wl-copy on Wayland
func ReadAll() (string, error) {
    // Try atotto/clipboard first
    text, err := atottoClipboard.ReadAll()
    if err == nil {
        return text, nil
    }
    
    // Fallback to wl-copy for Wayland
    cmd := exec.Command("wl-copy")
    output, err := cmd.Output()
    if err != nil {
        return "", err
    }
    return string(output), nil
}
```

---

## Hotkey Detection

### Library: github.com/go-vgo/robotgo

Robotgo provides cross-platform hotkey detection using cgo. It wraps native APIs for each platform.

**Implementation note:** The shipped hotkey listener uses `github.com/robotn/gohook` directly (`hook.Register` + `hook.Start` + `hook.Process` + `hook.End`); see [`docs/hotkey.md`](../../hotkey.md).

### Implementation

```go
// internal/hotkey/listener.go
package hotkey

import (
    "context"
    "github.com/go-vgo/robotgo"
    "time"
)

type Listener struct {
    modifiers   []string
    key         string
    debounceMS  int
    onPress     func()
    lastPressed time.Time
}

func NewListener(cfg Config) *Listener {
    return &Listener{
        modifiers:  cfg.Modifiers,
        key:        cfg.Key,
        debounceMS: cfg.DebounceMS,
    }
}

func (l *Listener) Start(ctx context.Context, onPress func()) error {
    l.onPress = onPress
    
    // Register hotkey
    keycode := l.getKeyCode()
    for _, mod := range l.modifiers {
        keycode |= l.getModifierCode(mod)
    }
    
    robotgo.AddEvent(keycode, "hotkey", func() {
        l.handlePress()
    })
    
    // Wait for context cancellation
    <-ctx.Done()
    robotgo.RemoveEvent("hotkey")
    return nil
}

func (l *Listener) handlePress() {
    now := time.Now()
    if now.Sub(l.lastPressed) < time.Duration(l.debounceMS)*time.Millisecond {
        return // Debounced
    }
    l.lastPressed = now
    l.onPress()
}

func (l *Listener) getKeyCode() robotgo.Keycode {
    // Map key string to robotgo keycode
    switch l.key {
    case "a": return robotgo.KeyA
    case "b": return robotgo.KeyB
    // ... all keys
    default: return 0
    }
}

func (l *Listener) getModifierCode(mod string) robotgo.Keycode {
    switch mod {
    case "ctrl": return robotgo.KeyControl
    case "alt": return robotgo.KeyAlt
    case "shift": return robotgo.KeyShift
    case "cmd": return robotgo.KeyMeta // macOS Command key
    default: return 0
    }
}
```

### Platform-Specific Notes

| Platform | Behavior | Limitations |
|----------|----------|-------------|
| Windows | Uses RegisterHotKey API | Works globally |
| macOS | Uses NSEvent.addLocalMonitor | Requires app to be running |
| Linux | Uses X11 GrabKey | Requires X11, not Wayland |

### Hotkey Conflict Detection

```go
// Check if hotkey is already registered
func CheckHotkeyConflict(modifiers []string, key string) (bool, error) {
    // This is platform-specific and difficult to check programmatically
    // Recommendation: Try to register, handle error if it fails
    
    // For Windows: RegisterHotKey returns ERROR_HOTKEY_ALREADY_REGISTERED
    // For macOS: No built-in way, just try
    // For Linux: No built-in way
    
    // Alternative: Warn user on first press if hotkey doesn't work
    return false, nil
}
```

---

## Terminal Spawning

### Strategy

Each platform requires different commands to spawn a new terminal window with our TUI process running inside it.

### Terminal Detection

```go
// internal/terminal/detect.go
package terminal

import (
    "os/exec"
    "runtime"
)

// terminalCommands maps OS to list of terminal commands to try
var terminalCommands = map[string][]string{
    "windows": {
        "cmd /c start /B cmd /c %s",
        "powershell -Command Start-Process -NoNewWindow -FilePath cmd -ArgumentList /c,%s",
    },
    "darwin": {
        "open -a Terminal --args -e %s",
        "open -a iTerm --args %s",
    },
    "linux": {
        "x-terminal-emulator -e %s",
        "gnome-terminal -- %s",
        "konsole -e %s",
        "xfce4-terminal -e %s",
        "alacritty -e %s",
        "kitty %s",
    },
}

// DetectTerminal finds an available terminal command
func DetectTerminal() (string, error) {
    os := runtime.GOOS
    commands := terminalCommands[os]
    
    for _, cmd := range commands {
        // Extract the binary name (first word)
        parts := strings.Fields(cmd)
        if len(parts) == 0 {
            continue
        }
        binary := parts[0]
        
        // Check if binary exists
        _, err := exec.LookPath(binary)
        if err == nil {
            return cmd, nil
        }
    }
    
    return "", errors.New("no terminal emulator found")
}
```

### Spawning with stdin

```go
// internal/terminal/spawner.go
package terminal

import (
    "fmt"
    "os"
    "os/exec"
)

// SpawnTUI spawns a new terminal running the TUI with clipboard text piped to stdin
func SpawnTUI(executable string, text string) error {
    // Get terminal command
    terminalCmd, err := DetectTerminal()
    if err != nil {
        return err
    }
    
    switch runtime.GOOS {
    case "windows":
        return spawnWindows(terminalCmd, executable, text)
    case "darwin":
        return spawnDarwin(terminalCmd, executable, text)
    default: // linux
        return spawnLinux(terminalCmd, executable, text)
    }
}

func spawnWindows(terminalCmd, executable, text string) error {
    // On Windows, use a temp file because stdin piping is unreliable
    tempFile, err := os.CreateTemp("", "clipboard-tui-*.txt")
    if err != nil {
        return err
    }
    defer os.Remove(tempFile.Name())
    
    if _, err := tempFile.WriteString(text); err != nil {
        return err
    }
    tempFile.Close()
    
    // Build command: start cmd /c executable < tempFile
    cmd := exec.Command("cmd", "/c", "start", "", "/B", "cmd", "/c", executable, "<", tempFile.Name())
    return cmd.Start()
}

func spawnDarwin(terminalCmd, executable, text string) error {
    // macOS: Use open command
    fullCmd := fmt.Sprintf(terminalCmd, fmt.Sprintf("sh -c \"echo '%s' | %s\"", escapeShell(text), executable))
    cmd := exec.Command("sh", "-c", fullCmd)
    return cmd.Start()
}

func spawnLinux(terminalCmd, executable, text string) error {
    // Linux: Most terminals support -e flag
    fullCmd := fmt.Sprintf(terminalCmd, fmt.Sprintf("sh -c \"echo '%s' | %s\"", escapeShell(text), executable))
    cmd := exec.Command("sh", "-c", fullCmd)
    return cmd.Start()
}

func escapeShell(s string) string {
    // Escape special characters for shell
    s = strings.ReplaceAll(s, "'", "'\\''")
    s = strings.ReplaceAll(s, "\n", "\\n")
    s = strings.ReplaceAll(s, "\r", "\\r")
    return s
}
```

### Fallback: Direct Execution

If terminal spawning fails, fall back to writing to a file and opening it:

```go
// internal/terminal/fallback.go
package terminal

import (
    "fmt"
    "os"
    "os/exec"
    "path/filepath"
)

func FallbackOutput(text, result string) error {
    // Write result to temp file
    home := os.Getenv("HOME")
    if home == "" && runtime.GOOS == "windows" {
        home = os.Getenv("USERPROFILE")
    }
    tempDir := filepath.Join(home, ".config", "clipboard-tui", "output")
    if err := os.MkdirAll(tempDir, 0755); err != nil {
        return err
    }
    
    outputFile := filepath.Join(tempDir, fmt.Sprintf("output-%d.txt", time.Now().Unix()))
    if err := os.WriteFile(outputFile, []byte(result), 0644); err != nil {
        return err
    }
    
    // Open file with default application
    cmd := exec.Command("open", outputFile) // macOS
    if runtime.GOOS == "windows" {
        cmd = exec.Command("cmd", "/c", "start", "", outputFile)
    } else if runtime.GOOS == "linux" {
        cmd = exec.Command("xdg-open", outputFile)
    }
    
    return cmd.Start()
}
```

---

## Auto-Start Configuration

### Windows: Scheduled Task

```powershell
# scripts/install.ps1
param(
    [string]$BinaryPath
)

# Create Scheduled Task to run on login
$Action = New-ScheduledTaskAction -Execute $BinaryPath -Argument "daemon"
$Trigger = New-ScheduledTaskTrigger -AtLogOn
$Settings = New-ScheduledTaskSettingsSet -StartWhenAvailable -DontStopOnIdleEnd
$Principal = New-ScheduledTaskPrincipal -UserId (New-SchedulablePrincipal -LogonType S4U -UserId "NT AUTHORITY\SYSTEM" -GroupId "BUILTIN\Administrators")

Register-ScheduledTask -TaskName "ClipboardTUI-Daemon" -Action $Action -Trigger $Trigger -Settings $Settings -Principal $Principal -Force

Write-Host "Clipboard TUI daemon installed. It will start automatically on login."
```

```powershell
# scripts/uninstall.ps1
Unregister-ScheduledTask -TaskName "ClipboardTUI-Daemon" -Confirm:$false -ErrorAction SilentlyContinue
Write-Host "Clipboard TUI daemon uninstalled."
```

### macOS: Login Item

```bash
# scripts/install.sh (macOS section)
#!/bin/bash

# Get the path to this script and the binary
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="$SCRIPT_DIR/../clipboard-tui"

# Create AppleScript to run the daemon
cat > /tmp/clipboard-tui.scpt << 'EOF'
tell application "System Events"
    do shell script "$BINARY daemon &"
end tell
EOF

# Add to login items
osascript -e 'tell application "System Events" to make login item at end with properties {path:"/tmp/clipboard-tui.scpt", hidden:true}'

# Cleanup
rm /tmp/clipboard-tui.scpt

echo "Clipboard TUI daemon installed. It will start automatically on login."
```

```bash
# scripts/uninstall.sh (macOS section)
osascript -e 'tell application "System Events" to delete login item "clipboard-tui.scpt"'
```

### Linux: systemd User Service

```bash
# scripts/install.sh (Linux section)
#!/bin/bash

# Get the path to this script and the binary
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BINARY="$SCRIPT_DIR/../clipboard-tui"

# Create systemd service file
SERVICE_DIR="$HOME/.config/systemd/user"
mkdir -p "$SERVICE_DIR"

cat > "$SERVICE_DIR/clipboard-tui.service" << EOF
[Unit]
Description=Clipboard TUI Daemon
After=graphical-session.target

[Service]
ExecStart=$BINARY daemon
Restart=always
RestartSec=5s

[Install]
WantedBy=default.target
EOF

# Enable and start the service
systemctl --user daemon-reload
systemctl --user enable clipboard-tui.service
systemctl --user start clipboard-tui.service

echo "Clipboard TUI daemon installed. It will start automatically on login."
echo "To start now: systemctl --user start clipboard-tui.service"
echo "To view logs: journalctl --user -u clipboard-tui.service -f"
```

```bash
# scripts/uninstall.sh (Linux section)
systemctl --user stop clipboard-tui.service 2>/dev/null
systemctl --user disable clipboard-tui.service 2>/dev/null
rm -f "$HOME/.config/systemd/user/clipboard-tui.service"
systemctl --user daemon-reload
```

---

## Process Management

### PID File

```go
// internal/daemon/pidfile.go
package daemon

import (
    "fmt"
    "os"
    "path/filepath"
    "strconv"
    "syscall"
)

func WritePIDFile(path string) error {
    pid := os.Getpid()
    return os.WriteFile(path, []byte(strconv.Itoa(pid)), 0644)
}

func ReadPIDFile(path string) (int, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return 0, err
    }
    return strconv.Atoi(string(data))
}

func RemovePIDFile(path string) error {
    return os.Remove(path)
}

// CheckIfRunning checks if another instance is running
func CheckIfRunning(pidFile string) (bool, error) {
    pid, err := ReadPIDFile(pidFile)
    if err != nil {
        if os.IsNotExist(err) {
            return false, nil // No PID file, not running
        }
        return false, err
    }
    
    // Check if process exists
    process, err := os.FindProcess(pid)
    if err != nil {
        return false, nil // Process doesn't exist, stale PID file
    }
    
    // Send signal 0 to check if process is alive
    err = process.Signal(syscall.Signal(nil))
    if err != nil {
        if err.Error() == "no such process" {
            return false, nil // Stale PID file
        }
        return false, err
    }
    
    return true, nil // Process is running
}
```

### Single Instance Check

```go
// internal/daemon/single_instance.go
package daemon

import (
    "fmt"
    "os"
)

func EnsureSingleInstance(pidFile string) error {
    running, err := CheckIfRunning(pidFile)
    if err != nil {
        return fmt.Errorf("failed to check for running instance: %w", err)
    }
    if running {
        return fmt.Errorf("another instance is already running")
    }
    
    // Write our PID
    if err := WritePIDFile(pidFile); err != nil {
        return fmt.Errorf("failed to write PID file: %w", err)
    }
    
    return nil
}
```

---

## Platform-Specific Quirks

### Windows

1. **Hotkey Modifiers**: Use `KeyControl`, `KeyAlt`, `KeyShift`, `KeyMeta` (for Windows key)
2. **Terminal**: `cmd.exe` is always available, but PowerShell is more modern
3. **Clipboard**: Works with standard Win32 API
4. **Auto-start**: Scheduled Task is most reliable
5. **cgo**: Required for robotgo, needs GCC installed

### macOS

1. **Hotkey Modifiers**: Use `KeyMeta` for Command (⌘) key
2. **Terminal**: `open -a Terminal` works, but iTerm is popular
3. **Clipboard**: Requires app to have GUI context (may need `--with-gui` flag)
4. **Auto-start**: Login Items via AppleScript
5. **cgo**: Required for robotgo, should work with Xcode CLI tools
6. **Permissions**: May need accessibility permissions for hotkey detection

### Linux

1. **Terminal**: Many options, need to detect which is available
2. **Clipboard**: X11 only (not Wayland in v1)
3. **Auto-start**: systemd user service
4. **cgo**: Required for robotgo, needs X11 development libraries
5. **Dependencies**: May need to install `libx11-dev`, `libxtst-dev` for robotgo

---

## Build Tags

Use Go build tags to handle platform-specific code:

```go
// +build windows

// Windows-specific code here
```

```go
// +build linux

// Linux-specific code here
```

```go
// +build darwin

// macOS-specific code here
```

### Example: Platform-Specific Terminal Commands

```go
// internal/terminal/spawner_linux.go
// +build linux

package terminal

func getTerminalCommands() []string {
    return []string{
        "x-terminal-emulator -e %s",
        "gnome-terminal -- %s",
        "konsole -e %s",
        // ...
    }
}
```

```go
// internal/terminal/spawner_darwin.go
// +build darwin

package terminal

func getTerminalCommands() []string {
    return []string{
        "open -a Terminal --args -e %s",
        "open -a iTerm --args %s",
    }
}
```

```go
// internal/terminal/spawner_windows.go
// +build windows

package terminal

func getTerminalCommands() []string {
    return []string{
        "cmd /c start /B cmd /c %s",
        "powershell -Command Start-Process -NoNewWindow -FilePath cmd -ArgumentList /c,%s",
    }
}
```
