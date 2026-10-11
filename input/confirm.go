package input

import (
	"fmt"
	"strings"
)

// Confirm asks a yes/no question. An empty answer returns def.
func (p *Prompter) Confirm(prompt string, def bool) (bool, error) {
	suffix := " [y/N] "
	if def {
		suffix = " [Y/n] "
	}
	for {
		fmt.Fprint(p.Out, prompt+suffix)
		s, err := p.line()
		if err != nil {
			return false, err
		}
		switch strings.ToLower(strings.TrimSpace(s)) {
		case "":
			return def, nil
		case "y", "yes":
			return true, nil
		case "n", "no":
			return false, nil
		default:
			fmt.Fprintln(p.Out, "please answer y or n")
		}
	}
}
