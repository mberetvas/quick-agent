// Package hotkey provides cross-platform hotkey listening functionality.
// Tests use a mock detector to avoid requiring CGO for testing.
package hotkey

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/yourname/clipboard-tui/internal/config"
)

// MockDetector is a mock implementation of KeyDetector for testing
type MockDetector struct {
	mu             sync.Mutex
	registeredKeys [][]string
	onPress        func()
}

// AddEvents mock - always returns true
func (m *MockDetector) AddEvents(keys ...string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.registeredKeys = append(m.registeredKeys, keys)
	return true
}

// StartEventLoop mock - stores the callback and blocks until context is cancelled
func (m *MockDetector) StartEventLoop(ctx context.Context, onPress func()) {
	m.onPress = onPress
	<-ctx.Done()
}

// TriggerPress simulates a key press by calling the onPress callback
func (m *MockDetector) TriggerPress() {
	if m.onPress != nil {
		m.onPress()
	}
}

func TestListener_Debounce(t *testing.T) {
	cfg := config.HotkeyConfig{
		Modifiers:  []string{"ctrl", "alt"},
		Key:        "v",
		DebounceMS: 100,
	}

	// Use mock detector
	mockDetector := &MockDetector{}
	SetDetector(mockDetector)
	defer SetDetector(nil) // Reset after test

	listener := NewListener(cfg)
	events := make(chan struct{}, 10)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = listener.Start(ctx, events)
	}()

	// Wait a bit for Start to set up the detector
	time.Sleep(10 * time.Millisecond)

	// First trigger - should send event
	mockDetector.TriggerPress()

	select {
	case <-events:
		// Success - first event received
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected first event to be sent")
	}

	// Second trigger immediately - should be debounced
	mockDetector.TriggerPress()

	// Check no second event
	select {
	case <-events:
		t.Fatal("Second event should have been debounced")
	case <-time.After(50 * time.Millisecond):
		// Good - no event
	}

	// Wait for debounce period to pass
	time.Sleep(110 * time.Millisecond)

	// Third trigger after debounce - should send event
	mockDetector.TriggerPress()

	select {
	case <-events:
		// Success - third event received
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected third event to be sent after debounce")
	}

	// Clean up
	cancel()
	wg.Wait()
}

func TestListener_Start_SendsEventOnChannel(t *testing.T) {
	cfg := config.HotkeyConfig{
		Modifiers:  []string{"ctrl"},
		Key:        "c",
		DebounceMS: 10,
	}

	mockDetector := &MockDetector{}
	SetDetector(mockDetector)
	defer SetDetector(nil)

	listener := NewListener(cfg)
	events := make(chan struct{}, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = listener.Start(ctx, events)
	}()

	// Wait a bit for setup
	time.Sleep(10 * time.Millisecond)

	// Trigger an event
	mockDetector.TriggerPress()

	// Wait for event
	select {
	case <-events:
		// Success
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected event on channel")
	}

	cancel()
	wg.Wait()
}

func TestListener_CleanShutdown(t *testing.T) {
	cfg := config.HotkeyConfig{
		Modifiers:  []string{"alt"},
		Key:        "tab",
		DebounceMS: 50,
	}

	mockDetector := &MockDetector{}
	SetDetector(mockDetector)
	defer SetDetector(nil)

	listener := NewListener(cfg)
	events := make(chan struct{}, 1)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = listener.Start(ctx, events)
	}()

	// Let it run for a bit
	time.Sleep(10 * time.Millisecond)

	// Cancel context
	cancel()

	// Wait for Start to return
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - clean shutdown
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected clean shutdown")
	}
}

func TestListener_ChannelFull_DropsEvent(t *testing.T) {
	cfg := config.HotkeyConfig{
		Modifiers:  []string{},
		Key:        "a",
		DebounceMS: 0, // No debounce
	}

	mockDetector := &MockDetector{}
	SetDetector(mockDetector)
	defer SetDetector(nil)

	listener := NewListener(cfg)

	// Use buffered channel of size 1
	events := make(chan struct{}, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = listener.Start(ctx, events)
	}()

	// Wait a bit for setup
	time.Sleep(10 * time.Millisecond)

	// First trigger - fills the buffer
	mockDetector.TriggerPress()
	time.Sleep(10 * time.Millisecond)

	// Second trigger - buffer is full, should be dropped
	mockDetector.TriggerPress()

	// Drain the first event
	select {
	case <-events:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Expected to drain first event")
	}

	// No second event should be available
	select {
	case <-events:
		t.Fatal("Second event should have been dropped")
	case <-time.After(50 * time.Millisecond):
		// Success - no second event
	}

	cancel()
	wg.Wait()
}

func TestNewListener(t *testing.T) {
	cfg := config.DefaultHotkeyConfig()
	mockDetector := &MockDetector{}
	SetDetector(mockDetector)
	defer SetDetector(nil)

	listener := NewListener(cfg)

	if listener == nil {
		t.Fatal("NewListener returned nil")
	}

	if listener.active {
		t.Error("Listener should not be active before Start is called")
	}
}

func TestListener_PlatformConfigs(t *testing.T) {
	tests := []struct {
		name     string
		modifiers []string
		key      string
	}{
		{"Windows Ctrl+Alt+V", []string{"ctrl", "alt"}, "v"},
		{"macOS Cmd+Option+V", []string{"cmd", "option"}, "v"},
		{"Linux Ctrl+Shift+C", []string{"ctrl", "shift"}, "c"},
		{"Single key F1", []string{}, "f1"},
		{"NumPad 0", []string{"ctrl"}, "num0"},
	}

	mockDetector := &MockDetector{}
	SetDetector(mockDetector)
	defer SetDetector(nil)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.HotkeyConfig{
				Modifiers:  tt.modifiers,
				Key:        tt.key,
				DebounceMS: 300,
			}
			listener := NewListener(cfg)
			if listener == nil {
				t.Fatalf("NewListener returned nil for config: %+v", cfg)
			}
		})
	}
}

func TestListener_WithoutDetector_Fails(t *testing.T) {
	cfg := config.HotkeyConfig{
		Modifiers:  []string{"ctrl"},
		Key:        "v",
		DebounceMS: 100,
	}

	// Ensure no detector is set
	SetDetector(nil)

	listener := NewListener(cfg)
	events := make(chan struct{}, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start should fail without a detector
	err := listener.Start(ctx, events)
	if err == nil {
		t.Fatal("Expected error when no detector is set")
	}
	if err.Error() != "no key detector available. Build with CGO_ENABLED=1 for robotgo support" {
		t.Errorf("Unexpected error message: %v", err)
	}
}
