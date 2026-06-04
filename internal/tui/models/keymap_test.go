package models

import (
	"testing"

	"github.com/yourname/clipboard-tui/internal/config"
)

func TestKeyMap_Match(t *testing.T) {
	km := NewKeyMap(config.DefaultTUIConfig())

	if !km.Match("navigate_down", "j") {
		t.Error("expected j to match navigate_down")
	}
	if !km.Match("select", "enter") {
		t.Error("expected enter to match select")
	}
	if km.Match("navigate_down", "enter") {
		t.Error("enter should not match navigate_down")
	}
}

func TestKeyMap_empty_bindings_uses_defaults(t *testing.T) {
	km := NewKeyMap(config.TUIConfig{})
	if !km.Match("back", "esc") {
		t.Error("expected default back binding for esc")
	}
}
