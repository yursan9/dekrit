package output

import (
	"io"
	"os"

	"golang.org/x/term"
)

// isTerminal reports whether w is connected to a character device (a
// terminal).
func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// terminalWidth returns the terminal width in columns, or -1 when w is not a
// terminal or the size is unavailable.
func terminalWidth(w io.Writer) int {
	f, ok := w.(*os.File)
	if !ok {
		return -1
	}
	width, _, err := term.GetSize(int(f.Fd()))
	if err != nil {
		return -1
	}
	return width
}
