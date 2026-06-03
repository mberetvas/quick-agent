package clipboard

import (
	"context"
	"time"

	"github.com/yourname/clipboard-tui/internal/config"
)

// Poller polls the clipboard periodically and emits sanitized changes on a channel.
type Poller struct {
	cb          Clipboard
	cfg         *config.Config
	changes     chan string
	errors      chan error
	lastVal     string
	isRunning   bool
	fastPoll    time.Duration
	slowPoll    time.Duration
	currentPoll time.Duration
}

// NewPoller initializes a new Clipboard Poller with given configuration and clipboard accessor.
func NewPoller(cb Clipboard, cfg *config.Config) *Poller {
	slow := time.Duration(cfg.Clipboard.PollIntervalMS) * time.Millisecond
	if slow > 1*time.Second {
		slow = 1 * time.Second
	}
	if slow < 100*time.Millisecond {
		slow = 100 * time.Millisecond
	}

	return &Poller{
		cb:          cb,
		cfg:         cfg,
		changes:     make(chan string, 10),
		errors:      make(chan error, 10),
		fastPoll:    100 * time.Millisecond,
		slowPoll:    slow,
		currentPoll: slow,
	}
}

// Changes returns the read-only channel where sanitized changes are emitted.
func (p *Poller) Changes() <-chan string {
	return p.changes
}

// Errors returns the read-only channel where polling errors are emitted.
func (p *Poller) Errors() <-chan error {
	return p.errors
}

// Start spawns a background polling goroutine.
func (p *Poller) Start(ctx context.Context) {
	if p.isRunning {
		return
	}
	p.isRunning = true

	// Read initial value so we only emit actual new changes
	if initial, err := p.cb.Get(); err == nil {
		p.lastVal = Sanitize(initial, p.cfg.Clipboard.MaxSize)
	}

	go func() {
		defer func() {
			p.isRunning = false
			close(p.changes)
			close(p.errors)
		}()

		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(p.currentPoll):
				raw, err := p.cb.Get()
				if err != nil {
					select {
					case p.errors <- err:
					default:
					}
					// On error, let's keep polling at normal interval
					p.currentPoll = p.slowPoll
					continue
				}

				sanitized := Sanitize(raw, p.cfg.Clipboard.MaxSize)

				// Skip empty / whitespace-only (it is already sanitized to "")
				if sanitized == "" {
					p.currentPoll = p.slowPoll
					continue
				}

				if sanitized != p.lastVal {
					p.lastVal = sanitized
					// Emit the change
					select {
					case p.changes <- sanitized:
					default:
					}

					// Speed up polling to capture fast updates
					p.currentPoll = p.fastPoll
				} else {
					// No change - adaptively back off slowly
					if p.currentPoll < p.slowPoll {
						p.currentPoll += 100 * time.Millisecond
						if p.currentPoll > p.slowPoll {
							p.currentPoll = p.slowPoll
						}
					}
				}
			}
		}
	}()
}
