package input

import (
	"io"
	"os"
	"strings"
)

// Prompter reads user input from In and writes prompts to Out.
type Prompter struct {
	In  io.Reader
	Out io.Writer
}

// Option customizes a Prompter.
type Option func(*Prompter)

// WithInput sets the reader prompts read from.
func WithInput(r io.Reader) Option {
	return func(p *Prompter) {
		if r != nil {
			p.In = r
		}
	}
}

// WithOutput sets the writer prompts are written to.
func WithOutput(w io.Writer) Option {
	return func(p *Prompter) {
		if w != nil {
			p.Out = w
		}
	}
}

// New returns a Prompter writing to os.Stdout and reading from os.Stdin by
// default, overridden by options.
func New(opts ...Option) *Prompter {
	p := &Prompter{In: os.Stdin, Out: os.Stdout}
	for _, opt := range opts {
		opt(p)
	}
	return p
}

// line reads a single line without consuming any input beyond it, so multiple
// prompts can safely share the same reader (including non-interactive pipes).
// A partial line at EOF is returned without error.
func (p *Prompter) line() (string, error) {
	var sb strings.Builder
	buf := make([]byte, 1)
	for {
		n, err := p.In.Read(buf)
		if n > 0 {
			if buf[0] == '\n' {
				return strings.TrimRight(sb.String(), "\r"), nil
			}
			sb.WriteByte(buf[0])
		}
		if err != nil {
			if sb.Len() == 0 {
				return "", err
			}
			return strings.TrimRight(sb.String(), "\r"), nil
		}
	}
}

// Confirm asks a yes/no question. An empty answer returns def.
func Confirm(prompt string, def bool) (bool, error) {
	return New().Confirm(prompt, def)
}

// Text asks for a non-empty line of text.
func Text(prompt string) (string, error) {
	return New().Text(prompt)
}

// TextDefault asks for a line of text, returning def for an empty answer.
func TextDefault(prompt, def string) (string, error) {
	return New().TextDefault(prompt, def)
}

// Password asks for a password. Input is hidden when reading from a terminal.
func Password(prompt string) (string, error) {
	return New().Password(prompt)
}

// Select asks the user to pick one of options by number, returning its index.
func Select(prompt string, options []string) (int, error) {
	return New().Select(prompt, options)
}
