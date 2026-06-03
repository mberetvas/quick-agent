//go:build cgo

package hotkey

import (
	"context"

	hook "github.com/robotn/gohook"
)

// GohookDetector is the production implementation using github.com/robotn/gohook.
// Requires CGO_ENABLED=1 and a C compiler.
type GohookDetector struct {
	combo []string
}

// AddEvents validates and stores the normalized key combination for registration.
func (g *GohookDetector) AddEvents(keys ...string) bool {
	combo, err := normalizeKeys(keys...)
	if err != nil {
		return false
	}
	for _, k := range combo {
		if _, ok := hook.Keycode[k]; !ok {
			return false
		}
	}
	g.combo = combo
	return true
}

// StartEventLoop registers the combo, runs the gohook event loop until ctx is cancelled.
func (g *GohookDetector) StartEventLoop(ctx context.Context, onPress func()) {
	hook.Register(hook.KeyDown, g.combo, func(hook.Event) {
		onPress()
	})
	evChan := hook.Start()
	defer hook.End()
	procDone := hook.Process(evChan)
	select {
	case <-ctx.Done():
	case <-procDone:
	}
}

func init() {
	SetDetector(&GohookDetector{})
}
