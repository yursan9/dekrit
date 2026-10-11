package input

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Select presents options as a numbered list and returns the zero-based index
// of the chosen option.
func (p *Prompter) Select(prompt string, options []string) (int, error) {
	if len(options) == 0 {
		return -1, errors.New("input: no options to select from")
	}
	if prompt != "" {
		fmt.Fprintln(p.Out, prompt)
	}
	for i, opt := range options {
		fmt.Fprintf(p.Out, "  %d) %s\n", i+1, opt)
	}
	for {
		fmt.Fprint(p.Out, "> ")
		s, err := p.line()
		if err != nil {
			return -1, err
		}
		n, err := strconv.Atoi(strings.TrimSpace(s))
		if err == nil && n >= 1 && n <= len(options) {
			return n - 1, nil
		}
		fmt.Fprintf(p.Out, "enter a number between 1 and %d\n", len(options))
	}
}
