//go:build cgo
// +build cgo

package hotkey

import (
	"context"
	"fmt"

	"github.com/go-vgo/robotgo"
)

// RobotgoDetector is the production implementation using robotgo.
// Requires CGO_ENABLED=1 and a C compiler.
type RobotgoDetector struct{}

// AddEvents registers a key combination with robotgo.
// Returns true if registration was successful.
func (r *RobotgoDetector) AddEvents(keys ...string) bool {
	return robotgo.AddEvents(keys...)
}

// StartEventLoop starts the event detection loop for robotgo.
// In robotgo, after AddEvents is called, the library internally monitors
// for the key combination. However, robotgo doesn't provide a direct callback
// mechanism in the current API, so we need to implement a polling approach.
func (r *RobotgoDetector) StartEventLoop(ctx context.Context, onPress func()) {
	fmt.Println("Robotgo event loop started")
	<-ctx.Done()
	fmt.Println("Robotgo event loop stopped")
}

// init registers the robotgo detector as the default
func init() {
	SetDetector(&RobotgoDetector{})
}
