package dekrit

import (
	"context"
	"fmt"
	"io"
	"os"
)

// App is a command line application with one or more subcommands and shared
// configuration conventions.
type App struct {
	Name        string
	Description string
	Version     string
	Stdout      io.Writer
	Stderr      io.Writer
	Stdin       io.Reader

	LookupEnv func(string) (string, bool)

	commands []*command
	err      error
}

// Option customizes an App.
type Option func(*App)

// WithVersion sets the version reported by the version pseudo-command.
func WithVersion(v string) Option { return func(a *App) { a.Version = v } }

// WithDescription sets the short description shown in help output.
func WithDescription(d string) Option { return func(a *App) { a.Description = d } }

// WithStdin overrides the input reader.
func WithStdin(r io.Reader) Option { return func(a *App) { a.Stdin = r } }

// WithStdout overrides the standard output writer.
func WithStdout(w io.Writer) Option { return func(a *App) { a.Stdout = w } }

// WithStderr overrides the standard error writer.
func WithStderr(w io.Writer) Option { return func(a *App) { a.Stderr = w } }

// WithEnvLookup overrides how environment variables are resolved. It is
// primarily useful for testing.
func WithEnvLookup(fn func(string) (string, bool)) Option {
	return func(a *App) { a.LookupEnv = fn }
}

// New creates an App.
func New(name string, opts ...Option) *App {
	a := &App{
		Name:      name,
		Stdin:     os.Stdin,
		Stdout:    os.Stdout,
		Stderr:    os.Stderr,
		LookupEnv: os.LookupEnv,
	}
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// Command registers a subcommand under the given name.
func (a *App) Command(name string, cmd Command) *App {
	if a.err != nil {
		return a
	}
	if a.find(name) != nil {
		a.err = fmt.Errorf("command %q already registered", name)
		return a
	}
	c, err := newCommand(a.Name, name, cmd)
	if err != nil {
		a.err = err
		return a
	}
	a.commands = append(a.commands, c)
	return a
}

func (a *App) find(name string) *command {
	for _, c := range a.commands {
		if c.name == name {
			return c
		}
	}
	return nil
}

// defaultCommand returns the command that runs when no subcommand is selected.
// A lone command is treated as the default; otherwise there is none.
func (a *App) defaultCommand() *command {
	if len(a.commands) == 1 {
		return a.commands[0]
	}
	return nil
}

func (a *App) commandPrefix(name string) string {
	if a.Name == "" {
		return name
	}
	return a.Name + " " + name
}

// Run parses args (typically os.Args[1:]) and dispatches to a subcommand. It
// is equivalent to RunContext(context.Background(), args).
func (a *App) Run(args []string) error {
	return a.RunContext(context.Background(), args)
}

// RunContext is like Run but accepts a context.Context that is exposed to the
// invoked command through Context. Use it to propagate cancellation and
// deadlines, for example from signal.NotifyContext.
//
// Errors are logged to stderr and returned for the caller to map to an exit
// code via ExitCode.
func (a *App) RunContext(ctx context.Context, args []string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if a.err != nil {
		return a.logError(a.err)
	}

	if len(args) == 0 {
		if d := a.defaultCommand(); d != nil {
			return d.run(a, ctx, nil)
		}
		a.writeHelp(a.Stdout)
		return nil
	}

	switch args[0] {
	case "help", "-h", "--help", "-help":
		if len(args) > 1 {
			if c := a.find(args[1]); c != nil {
				c.writeHelp(a.Stdout, a, false)
				return nil
			}
			return a.logError(fmt.Errorf("unknown command %q", args[1]))
		}
		a.writeHelp(a.Stdout)
		return nil
	case "version", "--version", "-V":
		fmt.Fprintf(a.Stdout, "%s %s\n", a.Name, a.Version)
		return nil
	}

	if c := a.find(args[0]); c != nil {
		return c.run(a, ctx, args[1:])
	}
	if d := a.defaultCommand(); d != nil {
		return d.run(a, ctx, args)
	}
	return a.logError(fmt.Errorf("unknown command %q", args[0]))
}

func (a *App) logError(err error) error {
	if err == nil {
		return nil
	}
	fmt.Fprintf(a.Stderr, "%s: %v\n", a.Name, err)
	return err
}
