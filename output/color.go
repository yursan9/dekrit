package output

import (
	"fmt"
	"io"
	"os"
)

// NoColor disables ANSI styling globally when true.
var NoColor = false

// Style is an ANSI SGR style applied to text.
type Style struct {
	code string
}

// Sprint formats its arguments with fmt.Sprint and wraps them in the style's
// ANSI sequence followed by a reset, or returns plain text when color output
// is disabled.
func (s Style) Sprint(a ...any) string {
	str := fmt.Sprint(a...)
	if NoColor || s.code == "" {
		return str
	}
	return "\x1b[" + s.code + "m" + str + "\x1b[0m"
}

// Sprintf is like Sprint but with a format string.
func (s Style) Sprintf(format string, a ...any) string {
	if NoColor || s.code == "" {
		return fmt.Sprintf(format, a...)
	}
	return "\x1b[" + s.code + "m" + fmt.Sprintf(format, a...) + "\x1b[0m"
}

// Predefined styles.
var (
	Bold      = Style{"1"}
	Faint     = Style{"2"}
	Italic    = Style{"3"}
	Underline = Style{"4"}
	FgRed     = Style{"31"}
	FgGreen   = Style{"32"}
	FgYellow  = Style{"33"}
	FgBlue    = Style{"34"}
	FgMagenta = Style{"35"}
	FgCyan    = Style{"36"}
	FgWhite   = Style{"37"}
)

// Enabled reports whether styled output should be emitted for w. It is false
// when NoColor is set, the NO_COLOR environment variable is present, or w is
// not a terminal.
func Enabled(w io.Writer) bool {
	if NoColor {
		return false
	}
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}
	return isTerminal(w)
}

// Fprint writes a styled by s to w when styling is enabled for w, and plain
// text otherwise.
func Fprint(w io.Writer, s Style, a ...any) (int, error) {
	if !Enabled(w) {
		return fmt.Fprint(w, a...)
	}
	return fmt.Fprint(w, s.Sprint(a...))
}

// Fprintf is like Fprint but with a format string.
func Fprintf(w io.Writer, s Style, format string, a ...any) (int, error) {
	if !Enabled(w) {
		return fmt.Fprintf(w, format, a...)
	}
	return fmt.Fprint(w, s.Sprintf(format, a...))
}
