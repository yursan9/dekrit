package input

import (
	"fmt"
	"strings"
)

// Text asks for a non-empty line of text, reprompting until one is given.
func (p *Prompter) Text(prompt string) (string, error) {
	for {
		fmt.Fprint(p.Out, prompt+" ")
		s, err := p.line()
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(s) != "" {
			return s, nil
		}
		fmt.Fprintln(p.Out, "value is required")
	}
}

// TextDefault asks for a line of text, returning def for an empty answer.
func (p *Prompter) TextDefault(prompt, def string) (string, error) {
	fmt.Fprintf(p.Out, "%s [%s] ", prompt, def)
	s, err := p.line()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(s) == "" {
		return def, nil
	}
	return s, nil
}
