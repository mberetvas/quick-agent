package hotkey

import (
	"testing"

	"github.com/mberetvas/quick-agent/internal/config"
)

func TestListenerTrigger_sendsEvent(t *testing.T) {
	l := NewListener(config.DefaultHotkeyConfig())
	ch := make(chan struct{}, 1)
	l.events = ch
	l.active = true

	l.Trigger()

	select {
	case <-ch:
	default:
		t.Fatal("Trigger should send on events channel")
	}
}

func TestListenerTrigger_inactiveNoop(t *testing.T) {
	l := NewListener(config.DefaultHotkeyConfig())
	ch := make(chan struct{}, 1)
	l.events = ch
	l.active = false

	l.Trigger()

	select {
	case <-ch:
		t.Fatal("inactive listener should not emit")
	default:
	}
}
