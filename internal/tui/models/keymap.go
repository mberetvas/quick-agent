package models

import "github.com/mberetvas/quick-agent/internal/config"

// KeyMap resolves configured TUI keybindings.
type KeyMap struct {
	bindings map[string][]string
}

// NewKeyMap builds a KeyMap from config, falling back to defaults for missing entries.
func NewKeyMap(cfg config.TUIConfig) KeyMap {
	bindings := cfg.Keybindings
	if bindings == nil {
		bindings = config.DefaultTUIConfig().Keybindings
	}
	return KeyMap{bindings: bindings}
}

// Match reports whether key is bound to the given action name.
func (k KeyMap) Match(action, key string) bool {
	keys, ok := k.bindings[action]
	if !ok {
		return false
	}
	for _, bound := range keys {
		if bound == key {
			return true
		}
	}
	return false
}
