package output

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"time"
)

var defaultFrames = []string{"|", "/", "-", "\\"}

// Spinner renders a rotating animation next to a message. Animation only runs
// when the writer is a terminal.
type Spinner struct {
	mu       sync.Mutex
	w        io.Writer
	prefix   string
	frames   []string
	idx      int
	msg      string
	interval time.Duration
	enabled  bool
	running  bool
	stop     chan struct{}
	done     chan struct{}
}

// NewSpinner returns a spinner. It writes to os.Stdout by default.
func NewSpinner(opts ...Option) *Spinner {
	o := resolveOptions(opts)
	frames := append([]string(nil), defaultFrames...)
	return &Spinner{
		w:        o.Writer,
		prefix:   o.Prefix,
		frames:   frames,
		interval: 100 * time.Millisecond,
		enabled:  isTerminal(o.Writer),
	}
}

// Start begins animating with the given message.
func (s *Spinner) Start(msg string) {
	s.mu.Lock()
	s.msg = msg
	if s.running || !s.enabled {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.stop = make(chan struct{})
	s.done = make(chan struct{})
	stop, done := s.stop, s.done
	s.mu.Unlock()

	go func() {
		t := time.NewTicker(s.interval)
		defer t.Stop()
		defer close(done)
		for {
			select {
			case <-stop:
				return
			case <-t.C:
				s.mu.Lock()
				s.idx = (s.idx + 1) % len(s.frames)
				s.mu.Unlock()
				s.render()
			}
		}
	}()
}

// SetMessage replaces the displayed message.
func (s *Spinner) SetMessage(msg string) {
	s.mu.Lock()
	s.msg = msg
	s.mu.Unlock()
}

// Success stops the spinner and prints a success marker.
func (s *Spinner) Success(msg string) { s.stopWith(msg, "[ok]") }

// Fail stops the spinner and prints a failure marker.
func (s *Spinner) Fail(msg string) { s.stopWith(msg, "[x]") }

// Stop halts the spinner, clearing the line on a terminal.
func (s *Spinner) Stop() {
	s.stopAnimation()
	if s.enabled {
		fmt.Fprint(s.w, "\r\x1b[K")
	}
}

// String returns the current line: an optional prefix, the frame, and the
// message.
func (s *Spinner) String() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	parts := make([]string, 0, 3)
	if s.prefix != "" {
		parts = append(parts, s.prefix)
	}
	if len(s.frames) > 0 {
		parts = append(parts, s.frames[s.idx])
	}
	if s.msg != "" {
		parts = append(parts, s.msg)
	}
	return strings.Join(parts, " ")
}

func (s *Spinner) stopWith(msg, marker string) {
	s.stopAnimation()
	s.mu.Lock()
	if msg == "" {
		msg = s.msg
	}
	prefix := s.prefix
	s.mu.Unlock()

	parts := make([]string, 0, 3)
	if prefix != "" {
		parts = append(parts, prefix)
	}
	parts = append(parts, marker)
	if msg != "" {
		parts = append(parts, msg)
	}
	line := strings.Join(parts, " ")

	if s.enabled {
		fmt.Fprint(s.w, "\r"+line+"\x1b[K\n")
		return
	}
	fmt.Fprintln(s.w, line)
}

func (s *Spinner) render() {
	if s.enabled {
		fmt.Fprint(s.w, "\r"+s.String()+"\x1b[K")
	}
}

func (s *Spinner) stopAnimation() {
	s.mu.Lock()
	running := s.running
	s.running = false
	stop, done := s.stop, s.done
	s.mu.Unlock()
	if running {
		close(stop)
		<-done
	}
}
