package input

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// Password asks for a password. When reading from a terminal the input is
// hidden; otherwise it falls back to reading a plain line.
func (p *Prompter) Password(prompt string) (string, error) {
	fmt.Fprint(p.Out, prompt+" ")
	f, ok := p.In.(*os.File)
	if ok && term.IsTerminal(int(f.Fd())) {
		b, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(p.Out)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
	return p.line()
}
