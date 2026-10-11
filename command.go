package dekrit

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"reflect"
)

// Command is implemented by user subcommands. A command is typically a struct
// with dekrit-tagged fields and a Run method.
type Command interface {
	Run(ctx *Context) error
}

// Context carries per-invocation state to a command's Run method.
type Context struct {
	App    *App
	Args   []string
	Stdin  io.Reader
	Stdout io.Writer
	Stderr io.Writer
}

// ExitError wraps an error with a process exit code.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e.Err == nil {
		return fmt.Sprintf("exit status %d", e.Code)
	}
	return e.Err.Error()
}

func (e *ExitError) Unwrap() error { return e.Err }

// ExitCode reports the exit code associated with err, defaulting to 1.
func ExitCode(err error) int {
	if err == nil {
		return 0
	}
	var ee *ExitError
	if errors.As(err, &ee) {
		return ee.Code
	}
	return 1
}

type command struct {
	name  string
	cmd   Command
	flags []*fieldInfo
	pos   []*fieldInfo
	all   []*fieldInfo
}

func newCommand(appName, name string, cmd Command) (*command, error) {
	rv := reflect.ValueOf(cmd)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return nil, fmt.Errorf("command %q must be a non-nil pointer to a struct", name)
	}
	if rv.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("command %q must point to a struct", name)
	}

	fields, err := collectFields(rv.Elem().Type(), appName)
	if err != nil {
		return nil, fmt.Errorf("command %q: %w", name, err)
	}

	c := &command{name: name, cmd: cmd, all: fields}
	names := map[string]string{}
	posIndex := 0
	for _, f := range fields {
		if f.isFlag() {
			if f.Name == "h" || f.Name == "help" {
				return nil, fmt.Errorf("command %q: flag name %q is reserved for help", name, f.Name)
			}
			if prev, ok := names[f.Name]; ok {
				return nil, fmt.Errorf("command %q: duplicate flag %q (also %s)", name, f.Name, prev)
			}
			names[f.Name] = "--" + f.Name
			if f.Short != "" {
				if prev, ok := names[f.Short]; ok {
					return nil, fmt.Errorf("command %q: duplicate short flag %q (also %s)", name, f.Short, prev)
				}
				if f.Short == "h" {
					return nil, fmt.Errorf("command %q: short flag %q is reserved for help", name, f.Short)
				}
				names[f.Short] = "-" + f.Short
			}
			c.flags = append(c.flags, f)
		}
		if f.isPos() {
			f.posIndex = posIndex
			posIndex++
			c.pos = append(c.pos, f)
		}
	}
	return c, nil
}

func (c *command) run(a *App, args []string) error {
	fs := flag.NewFlagSet(c.name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}

	raw := map[string]*rawValue{}
	for _, f := range c.flags {
		rv := &rawValue{isBool: isBoolType(f.Type)}
		raw[f.Name] = rv
		fs.Var(rv, f.Name, f.Help)
		if f.Short != "" {
			raw[f.Short] = rv
			fs.Var(rv, f.Short, f.Help)
		}
	}

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			c.writeHelp(a.Stdout, a)
			return nil
		}
		fmt.Fprintf(a.Stderr, "%s: %v\n", a.commandPrefix(c.name), err)
		fmt.Fprintf(a.Stderr, "Run '%s %s --help' for usage.\n", a.Name, c.name)
		return &ExitError{Code: 1, Err: err}
	}

	if err := c.resolve(reflect.ValueOf(c.cmd).Elem(), raw, fs.Args(), a.LookupEnv); err != nil {
		fmt.Fprintf(a.Stderr, "%s: %v\n", a.commandPrefix(c.name), err)
		return &ExitError{Code: 1, Err: err}
	}

	ctx := &Context{
		App:    a,
		Args:   fs.Args(),
		Stdin:  a.Stdin,
		Stdout: a.Stdout,
		Stderr: a.Stderr,
	}
	if err := c.cmd.Run(ctx); err != nil {
		return a.logError(err)
	}
	return nil
}

// resolve populates the command struct using the precedence:
// CLI flag/positional > environment variable > default > zero value.
func (c *command) resolve(root reflect.Value, raw map[string]*rawValue, args []string, lookupEnv func(string) (string, bool)) error {
	for _, f := range c.all {
		var vals []string
		provided := false

		if f.isFlag() {
			if rv, ok := raw[f.Name]; ok && rv.provided {
				vals, provided = rv.vals, true
			}
		}
		if !provided && f.isPos() && f.posIndex < len(args) {
			vals, provided = []string{args[f.posIndex]}, true
		}
		if !provided && f.isEnv() {
			if v, ok := lookupEnv(f.Env); ok {
				vals, provided = []string{v}, true
			}
		}
		if !provided && f.Default != "" {
			vals, provided = []string{f.Default}, true
		}
		if !provided {
			if f.Required {
				return fmt.Errorf("missing required argument %s", f.label())
			}
			continue
		}

		dst := fieldByIndexAlloc(root, f.Index)
		if err := setField(dst, f, vals); err != nil {
			return fmt.Errorf("%s: %w", f.label(), err)
		}
	}
	return nil
}
