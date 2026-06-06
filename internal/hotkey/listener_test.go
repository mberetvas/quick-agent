// Package hotkey provides cross-platform hotkey listening functionality.
// Tests use a mock detector to avoid requiring CGO for testing.
package hotkey

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/mberetvas/quick-agent/internal/config"
)

// MockDetector is a mock implementation of KeyDetector for testing
type MockDetector struct {
	mu             sync.Mutex
	registeredKeys [][]string
	onPress        func()
	ready          chan struct{}
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
	if m.ready != nil {
		close(m.ready)
	}
	<-ctx.Done()
}

// TriggerPress simulates a key press by calling the onPress callback
func (m *MockDetector) TriggerPress() {
	if m.onPress != nil {
		m.onPress()
	}
}

func waitForMockReady(t *testing.T, mock *MockDetector) {
	t.Helper()
	select {
	case <-mock.ready:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("mock detector did not become ready")
	}
}

func TestListener_RegistersExpectedCombo(t *testing.T) {
	cfg := config.HotkeyConfig{
		Modifiers:  []string{"ctrl", "alt"},
		Key:        "v",
		DebounceMS: 100,
	}

	mockDetector := &MockDetector{ready: make(chan struct{})}
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

	waitForMockReady(t, mockDetector)

	mockDetector.mu.Lock()
	if len(mockDetector.registeredKeys) != 1 {
		t.Fatalf("expected 1 registration, got %d", len(mockDetector.registeredKeys))
	}
	got := mockDetector.registeredKeys[0]
	mockDetector.mu.Unlock()

	want := []string{"ctrl", "alt", "v"}
	if len(got) != len(want) {
		t.Fatalf("registered keys = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("registered keys = %v, want %v", got, want)
		}
	}

	cancel()
	wg.Wait()
}

func TestListener_Debounce(t *testing.T) {
	cfg := config.HotkeyConfig{
		Modifiers:  []string{"ctrl", "alt"},
		Key:        "v",
		DebounceMS: 100,
	}

	mockDetector := &MockDetector{ready: make(chan struct{})}
	SetDetector(mockDetector)
	defer SetDetector(nil)

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

	waitForMockReady(t, mockDetector)

	mockDetector.TriggerPress()

	select {
	case <-events:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected first event to be sent")
	}

	mockDetector.TriggerPress()

	select {
	case <-events:
		t.Fatal("Second event should have been debounced")
	case <-time.After(50 * time.Millisecond):
	}

	time.Sleep(110 * time.Millisecond)

	mockDetector.TriggerPress()

	select {
	case <-events:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected third event to be sent after debounce")
	}

	cancel()
	wg.Wait()
}

func TestListener_Start_SendsEventOnChannel(t *testing.T) {
	cfg := config.HotkeyConfig{
		Modifiers:  []string{"ctrl"},
		Key:        "c",
		DebounceMS: 10,
	}

	mockDetector := &MockDetector{ready: make(chan struct{})}
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

	waitForMockReady(t, mockDetector)

	mockDetector.TriggerPress()

	select {
	case <-events:
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

	mockDetector := &MockDetector{ready: make(chan struct{})}
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

	waitForMockReady(t, mockDetector)

	cancel()

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("Expected clean shutdown")
	}
}

func TestListener_ChannelFull_DropsEvent(t *testing.T) {
	cfg := config.HotkeyConfig{
		Modifiers:  []string{"ctrl"},
		Key:        "a",
		DebounceMS: 0,
	}

	mockDetector := &MockDetector{ready: make(chan struct{})}
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

	waitForMockReady(t, mockDetector)

	mockDetector.TriggerPress()
	time.Sleep(10 * time.Millisecond)

	mockDetector.TriggerPress()

	select {
	case <-events:
	case <-time.After(50 * time.Millisecond):
		t.Fatal("Expected to drain first event")
	}

	select {
	case <-events:
		t.Fatal("Second event should have been dropped")
	case <-time.After(50 * time.Millisecond):
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
		name      string
		modifiers []string
		key       string
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

	SetDetector(nil)

	listener := NewListener(cfg)
	events := make(chan struct{}, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := listener.Start(ctx, events)
	if err == nil {
		t.Fatal("Expected error when no detector is set")
	}
	if err.Error() != "no key detector available. Build with CGO_ENABLED=1 for gohook (CGO) support" {
		t.Errorf("Unexpected error message: %v", err)
	}
}
