package output

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

// Progress renders a simple text progress bar. Animation only runs when the
// writer is a terminal; otherwise only the final state is printed.
type Progress struct {
	mu       sync.Mutex
	w        io.Writer
	prefix   string
	total    int64
	current  int64
	width    int
	interval time.Duration
	enabled  bool
	running  bool
	stop     chan struct{}
	done     chan struct{}
}

// NewProgress returns a progress bar. It writes to os.Stdout by default.
func NewProgress(opts ...Option) *Progress {
	o := resolveOptions(opts)
	width := 30
	if tw := terminalWidth(o.Writer); tw > 0 && tw-20 < width {
		width = tw - 20
		if width < 5 {
			width = 5
		}
	}
	return &Progress{
		w:        o.Writer,
		prefix:   o.Prefix,
		width:    width,
		interval: 100 * time.Millisecond,
		enabled:  isTerminal(o.Writer),
	}
}

// SetTotal sets the total number of units.
func (p *Progress) SetTotal(n int64) {
	p.mu.Lock()
	p.total = n
	p.mu.Unlock()
}

// Set sets the current number of units.
func (p *Progress) Set(n int64) {
	p.mu.Lock()
	p.current = n
	p.mu.Unlock()
}

// Add increments the current count by n.
func (p *Progress) Add(n int64) {
	p.mu.Lock()
	p.current += n
	p.mu.Unlock()
}

// Start begins periodic rendering. It is a no-op when not writing to a
// terminal or when already running.
func (p *Progress) Start() {
	p.mu.Lock()
	if p.running || !p.enabled {
		p.mu.Unlock()
		return
	}
	p.running = true
	p.stop = make(chan struct{})
	p.done = make(chan struct{})
	stop, done := p.stop, p.done
	p.mu.Unlock()

	go func() {
		t := time.NewTicker(p.interval)
		defer t.Stop()
		defer close(done)
		for {
			select {
			case <-stop:
				return
			case <-t.C:
				p.render()
			}
		}
	}()
}

// Finish stops rendering and prints the final frame.
func (p *Progress) Finish() {
	p.stopAnimation()
	if p.enabled {
		fmt.Fprint(p.w, "\r"+p.String()+"\x1b[K\n")
		return
	}
	fmt.Fprintln(p.w, p.String())
}

// String returns the current bar as a string.
func (p *Progress) String() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	bar := p.bar()
	if p.prefix != "" {
		return p.prefix + " " + bar
	}
	return bar
}

func (p *Progress) bar() string {
	if p.total <= 0 {
		return fmt.Sprintf("%d", p.current)
	}
	frac := float64(p.current) / float64(p.total)
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	filled := int(frac * float64(p.width))
	bar := strings.Repeat("=", filled) + strings.Repeat(" ", p.width-filled)
	return fmt.Sprintf("[%s] %3.0f%% (%d/%d)", bar, frac*100, p.current, p.total)
}

func (p *Progress) render() {
	if p.enabled {
		fmt.Fprint(p.w, "\r"+p.String()+"\x1b[K")
	}
}

func (p *Progress) stopAnimation() {
	p.mu.Lock()
	running := p.running
	p.running = false
	stop, done := p.stop, p.done
	p.mu.Unlock()
	if running {
		close(stop)
		<-done
	}
}
