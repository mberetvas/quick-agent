package clipboard

import (
	"context"
	"sync"
	"time"

	"github.com/mberetvas/quick-agent/internal/config"
)

// Poller polls the clipboard periodically and emits sanitized changes on a channel.
type Poller struct {
	mu          sync.Mutex
	cb          Clipboard
	cfg         *config.Config
	changes     chan string
	errors      chan error
	lastVal     string
	isRunning   bool
	used        bool
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
	p.mu.Lock()
	if p.isRunning || p.used {
		p.mu.Unlock()
		return
	}
	p.isRunning = true
	p.used = true

	// Read initial value so we only emit actual new changes
	if initial, err := p.cb.Get(); err == nil {
		p.lastVal = Sanitize(initial, p.cfg.Clipboard.TruncateSize)
	}
	p.mu.Unlock()

	go func() {
		defer func() {
			p.mu.Lock()
			p.isRunning = false
			p.mu.Unlock()
			close(p.changes)
			close(p.errors)
		}()

		for {
			p.mu.Lock()
			interval := p.currentPoll
			p.mu.Unlock()

			select {
			case <-ctx.Done():
				return
			case <-time.After(interval):
				raw, err := p.cb.Get()
				if err != nil {
					select {
					case p.errors <- err:
					default:
					}
					// On error, let's keep polling at normal interval
					p.mu.Lock()
					p.currentPoll = p.slowPoll
					p.mu.Unlock()
					continue
				}

				sanitized := Sanitize(raw, p.cfg.Clipboard.TruncateSize)

				// Skip empty / whitespace-only (it is already sanitized to "")
				if sanitized == "" {
					p.mu.Lock()
					p.currentPoll = p.slowPoll
					p.mu.Unlock()
					continue
				}

				p.mu.Lock()
				last := p.lastVal
				p.mu.Unlock()

				if sanitized != last {
					p.mu.Lock()
					p.lastVal = sanitized
					p.currentPoll = p.fastPoll
					p.mu.Unlock()

					// Emit the change
					select {
					case p.changes <- sanitized:
					default:
					}
				} else {
					// No change - adaptively back off slowly
					p.mu.Lock()
					if p.currentPoll < p.slowPoll {
						p.currentPoll += 100 * time.Millisecond
						if p.currentPoll > p.slowPoll {
							p.currentPoll = p.slowPoll
						}
					}
					p.mu.Unlock()
				}
			}
		}
	}()
}

// LatestText returns the most recently seen sanitized clipboard value.
func (p *Poller) LatestText() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.lastVal
}

// CurrentPoll returns the current adaptive polling interval in a thread-safe manner.
func (p *Poller) CurrentPoll() time.Duration {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.currentPoll
}
