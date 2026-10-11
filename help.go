package dekrit

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

// Describer can be implemented by a Command to provide a one-line summary
// shown in help output.
type Describer interface {
	Describe() string
}

func typeName(t reflect.Type) string {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t == durationType {
		return "duration"
	}
	if isTextType(t) {
		return "value"
	}
	switch t.Kind() {
	case reflect.Bool:
		return ""
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "int"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "uint"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.Slice:
		if et := typeName(t.Elem()); et != "" {
			return et + "..."
		}
		return ""
	default:
		return t.String()
	}
}

func (a *App) writeHelp(w io.Writer) {
	if len(a.commands) == 1 {
		a.commands[0].writeHelp(w, a, true)
		return
	}

	name := a.Name
	if name == "" {
		name = "app"
	}
	fmt.Fprintf(w, "Usage: %s <command> [flags]\n", name)
	if a.Description != "" {
		fmt.Fprintf(w, "\n%s\n", a.Description)
	}

	if len(a.commands) > 0 {
		fmt.Fprintf(w, "\nCommands:\n")
		width := 0
		for _, c := range a.commands {
			if len(c.name) > width {
				width = len(c.name)
			}
		}
		for _, c := range a.commands {
			desc := ""
			if d, ok := c.cmd.(Describer); ok {
				desc = d.Describe()
			}
			fmt.Fprintf(w, "  %-*s  %s\n", width, c.name, desc)
		}
	}

	fmt.Fprintf(w, "\nFlags:\n")
	fmt.Fprintf(w, "  -h, --help  show this help\n")
	if len(a.commands) > 0 {
		fmt.Fprintf(w, "\nUse \"%s <command> --help\" for more information about a command.\n", name)
	}
}

// writeHelp renders a command's help. asRoot is true when the help represents
// the application itself (a single-command app invoked without a subcommand),
// in which case the program label is the app name rather than "app <command>".
func (c *command) writeHelp(w io.Writer, a *App, asRoot bool) {
	usage := a.commandPrefix(c.name)
	if asRoot {
		usage = a.Name
	}
	if d, ok := c.cmd.(Describer); ok {
		if s := d.Describe(); s != "" {
			fmt.Fprintf(w, "%s\n\n", s)
		}
	}
	fmt.Fprintf(w, "Usage: %s [flags]", usage)
	for _, f := range c.pos {
		fmt.Fprintf(w, " %s", f.label())
	}
	fmt.Fprintf(w, "\n")

	if len(c.pos) > 0 {
		fmt.Fprintf(w, "\nArguments:\n")
		width := 0
		for _, f := range c.pos {
			if l := len(f.label()); l > width {
				width = l
			}
		}
		for _, f := range c.pos {
			help := f.Help
			var extra []string
			if f.Required {
				extra = append(extra, "(required)")
			}
			if f.Default != "" {
				extra = append(extra, fmt.Sprintf("(default %q)", f.Default))
			}
			if len(extra) > 0 {
				help = strings.TrimSpace(help + " " + strings.Join(extra, " "))
			}
			fmt.Fprintf(w, "  %-*s  %s\n", width, f.label(), help)
		}
	}

	fmt.Fprintf(w, "\nFlags:\n")
	type row struct {
		left   string
		suffix string
	}
	var rows []row
	for _, f := range c.flags {
		if f.Hidden {
			continue
		}
		left := "      --" + f.Name
		if f.Short != "" {
			left = "  -" + f.Short + ", --" + f.Name
		}
		if tn := typeName(f.Type); tn != "" {
			left += " " + tn
		}
		var extra []string
		if f.Required {
			extra = append(extra, "(required)")
		}
		if f.Default != "" {
			extra = append(extra, fmt.Sprintf("(default %q)", f.Default))
		}
		if f.isEnv() {
			extra = append(extra, fmt.Sprintf("[env: %s]", f.Env))
		}
		help := f.Help
		if len(extra) > 0 {
			help = strings.TrimSpace(help + " " + strings.Join(extra, " "))
		}
		rows = append(rows, row{left: left, suffix: help})
	}
	rows = append(rows, row{left: "  -h, --help", suffix: "show this help"})

	width := 0
	for _, r := range rows {
		if len(r.left) > width {
			width = len(r.left)
		}
	}
	for _, r := range rows {
		fmt.Fprintf(w, "%-*s  %s\n", width, r.left, r.suffix)
	}
}
