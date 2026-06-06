// Package hotkey provides cross-platform hotkey listening functionality.
// It uses github.com/robotn/gohook (CGO) for detecting key presses and emits events on a channel.
//
// REQUIREMENTS FOR PRODUCTION USE:
//   - CGO_ENABLED=1 must be set
//   - A C compiler must be installed:
//     * Windows: MinGW-w64 or MSVC (via Visual Studio Build Tools)
//     * Linux: GCC (sudo apt-get install gcc libx11-dev libxtst-dev)
//     * macOS: Xcode command line tools (xcode-select --install)
//
// Build with: CGO_ENABLED=1 go build
//
// For testing without CGO, the tests use the mock detector and Trigger method.
//
// See: docs/hotkey.md
package hotkey

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/mberetvas/quick-agent/internal/config"
)

// KeyDetector is an interface for detecting hotkey presses.
// This allows for mocking in tests and potential alternative implementations.
type KeyDetector interface {
	// AddEvents registers a key combination to listen for.
	// Returns true if successful.
	AddEvents(keys ...string) bool
	// StartEventLoop starts the event detection loop.
	// It should block until the context is cancelled.
	StartEventLoop(ctx context.Context, onPress func())
}

// defaultDetector uses gohook for actual hotkey detection (when CGO is enabled).
// Set via init() in gohook_detector.go when CGO_ENABLED=1.
var defaultDetector KeyDetector

// SetDetector sets a custom key detector. Useful for testing.
func SetDetector(d KeyDetector) {
	defaultDetector = d
}

// Listener monitors for hotkey presses and emits events on a channel.
// It supports platform-specific modifier keys and debouncing.
type Listener struct {
	cfg         config.HotkeyConfig
	events      chan<- struct{}
	lastPressed time.Time
	mu          sync.Mutex
	active      bool
	detector    KeyDetector
}

// NewListener creates a new hotkey listener with the given configuration.
func NewListener(cfg config.HotkeyConfig) *Listener {
	return &Listener{
		cfg:      cfg,
		active:   false,
		detector: defaultDetector,
	}
}

// Start begins listening for hotkey presses and sends an empty struct on the
// events channel each time the hotkey combination is detected.
// It respects the configured debounce interval to prevent duplicate events.
// Blocks until the context is cancelled, at which point it cleans up resources.
//
// The hotkey combination uses the detector's AddEvents which expects modifier keys
// followed by the main key as separate string arguments.
// Supported modifiers: "ctrl", "alt", "shift", "cmd" (macOS), "option" (alias for alt), "win" (Windows)
// Supported keys: letters (a-z), numbers (0-9), F1-F12, "up", "down", "left", "right",
// "space", "enter", "tab", "esc", "backspace", "delete", "home", "end", "pageup", "pagedown"
func (l *Listener) Start(ctx context.Context, events chan<- struct{}) error {
	l.events = events
	l.lastPressed = time.Time{}
	l.active = true

	// Build arguments for AddEvents: modifiers + key
	args := make([]string, 0, len(l.cfg.Modifiers)+1)
	args = append(args, l.cfg.Modifiers...)
	args = append(args, l.cfg.Key)

	if l.detector == nil {
		return fmt.Errorf("no key detector available. Build with CGO_ENABLED=1 for gohook (CGO) support")
	}

	if !l.detector.AddEvents(args...) {
		l.active = false
		return fmt.Errorf("failed to register hotkey: %s", strings.Join(args, "+"))
	}

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		l.detector.StartEventLoop(ctx, l.handlePress)
	}()

	<-ctx.Done()
	l.active = false
	wg.Wait()
	return nil
}

// Trigger simulates a hotkey press event. This is primarily for testing
// when a real detector cannot be used (e.g., without CGO).
func (l *Listener) Trigger() {
	if !l.active {
		return
	}
	l.handlePress()
}

// handlePress handles a hotkey press event with debouncing.
func (l *Listener) handlePress() {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	if !l.lastPressed.IsZero() {
		elapsed := now.Sub(l.lastPressed)
		if elapsed < time.Duration(l.cfg.DebounceMS)*time.Millisecond {
			return // Debounced
		}
	}

	l.lastPressed = now
	// Non-blocking send to avoid blocking the caller
	select {
	case l.events <- struct{}{}:
	default:
		// Channel full, drop the event
	}
}
