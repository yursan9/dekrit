# Dekrit

Dekrit is a library for making command line applications for the Go
programming language. The core of dekrit is centralized config between
subcommands. You define config once on a struct, as CLI parameters or
environment variables, using struct tags.

Dekrit uses only the standard library.

## Features

- One struct to define parameters and environment variables
- Auto defined help parameter
- Struct tags to define flags, positional arguments, or environment variables
- Shared configuration between subcommands via embedded structs
- Explicit value precedence: CLI flag/positional > environment variable > default
- `context.Context` support for cancellation and deadlines
- Automatic default command for single-command apps
- `dekrit/output`: JSON, tables, styled text, progress bars, and spinners
- `dekrit/input`: interactive prompts for confirmation, text, passwords, and selections

## Install

```sh
go get github.com/yursan9/dekrit
```

## Usage

```go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/yursan9/dekrit"
)

type serveCmd struct {
	Host    string        `dekrit:"flag,env" name:"host" short:"H" env:"APP_HOST" default:"localhost" help:"host to bind"`
	Port    int           `dekrit:"flag,env" name:"port" short:"p" env:"APP_PORT" default:"8080" help:"port to bind"`
	Verbose bool          `dekrit:"flag" name:"verbose" short:"v" help:"verbose output"`
	Timeout time.Duration `dekrit:"flag" name:"timeout" default:"5s" help:"request timeout"`
	File    string        `dekrit:"pos" name:"file" required:"true" help:"config file"`
}

func (s *serveCmd) Describe() string { return "start the server" }

func (s *serveCmd) Run(ctx *dekrit.Context) error {
	fmt.Fprintf(ctx.Stdout, "serving on %s:%d (file=%s)\n", s.Host, s.Port, s.File)
	return nil
}

func main() {
	app := dekrit.New("app",
		dekrit.WithVersion("1.0.0"),
		dekrit.WithDescription("example application"),
	)
	app.Command("serve", &serveCmd{})

	if err := app.Run(os.Args[1:]); err != nil {
		os.Exit(dekrit.ExitCode(err))
	}
}
```

## Help output

Running `app help` (or `app --help`) prints the application overview:

```text
Usage: app <command> [flags]

example application

Commands:
  serve    start the server
  migrate  run database migrations

Flags:
  -h, --help  show this help

Use "app <command> --help" for more information about a command.
```

Running `app serve --help` prints command-specific usage, including
positionals, defaults, and environment variables:

```text
start the server

Usage: app serve [flags] <file>

Arguments:
  <file>  config file (required)

Flags:
  -H, --host string       host to bind (default "localhost") [env: APP_HOST]
  -p, --port int          port to bind (default "8080") [env: APP_PORT]
  -v, --verbose           verbose output
      --timeout duration  request timeout (default "5s")
  -h, --help              show this help
```

## Struct tags

The `dekrit` tag declares the role of a field. A field may have more than one
role.

| `dekrit` value   | Meaning                                        |
| ---------------- | ---------------------------------------------- |
| `flag`           | CLI flag (requires a `name`)                   |
| `pos`            | Positional argument (in field declaration order) |
| `env`            | Environment variable                           |
| `flag,env`       | CLI flag with environment fallback             |

Modifier tags:

| Tag                  | Applies to | Meaning                                              |
| -------------------- | ---------- | ---------------------------------------------------- |
| `name:"host"`        | flag, pos  | Flag name (required for flags) / label for positionals |
| `short:"H"`          | flag       | Single character shorthand                           |
| `env:"APP_HOST"`     | env        | Environment variable name (defaults to `APP_FIELD`)  |
| `help:"..."`         | any        | Description shown in help                            |
| `default:"8080"`     | any        | Default value                                        |
| `required:"true"`    | any        | Error when no value is provided                      |
| `hidden:"true"`      | flag       | Hide from help output                                |

### Value precedence

For every field: **CLI flag/positional > environment variable > default > zero
value**. Environment variables are resolved with `os.LookupEnv` by default and
can be overridden with `dekrit.WithEnvLookup` for testing.

### Supported types

`string`, `bool`, all `int`/`uint` sizes, `float32`/`float64`,
`time.Duration`, pointers to any of these, and slices of the scalar types
(comma separated or repeated flags).

Any type implementing `encoding.TextUnmarshaler` is also supported, which
covers types such as `time.Time`, `net.IP`, `netip.Addr`, `url.URL`, and your
own custom types. A custom type is always treated as a single value, even when
its underlying kind is a slice, and its help placeholder is shown as `value`.

```go
type upper string

func (u *upper) UnmarshalText(b []byte) error {
	*u = upper(strings.ToUpper(string(b)))
	return nil
}

type cmd struct {
	Name upper `dekrit:"flag" name:"name"`
}
```

## Help

Every command automatically responds to `-h` / `--help`. The app responds to
`help`, `--help`, `version`, and `--version`.

## Default command

When exactly **one** command is registered, it becomes the *default* command:
it runs both when no subcommand is given and when the first argument is not a
recognized subcommand. The default receives all arguments, so flags and
positional arguments work without naming the command:

```sh
app                       # runs the command
app --port 8080           # flag routed to the command
app config.yaml           # positional routed to the command
app run --port 8080       # explicit name still works
```

Help adapts to a single-command app: `app --help` shows the command's own flags
and usage (`Usage: app [flags] ...`) instead of a command list, while
`app run --help` keeps the command name.

With two or more commands there is no default: a bare `app` prints help and an
unrecognized first argument is reported as an unknown command. Registering a
second command therefore changes bare invocation from running the command to
printing help, which matches how multi-command tools behave.

## Context and cancellation

Every command receives a `*dekrit.Context` that embeds a `context.Context`, so
it can react to cancellation and deadlines with `ctx.Done()`, `ctx.Err()`, and
`ctx.Deadline()`.

Pass your own context to `App.RunContext`; `App.Run` is shorthand for
`RunContext(context.Background(), args)`. Signal handling is left to the caller
so each program can choose its own policy. A typical `main` wires
`signal.NotifyContext` to the app:

```go
func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app := dekrit.New("app")
	app.Command("serve", &serveCmd{})
	if err := app.RunContext(ctx, os.Args[1:]); err != nil {
		os.Exit(dekrit.ExitCode(err))
	}
}

func (s *serveCmd) Run(ctx *dekrit.Context) error {
	<-ctx.Done() // blocks until a signal cancels the context

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return s.server.Shutdown(shutdownCtx)
}
```

The command owns its grace period (here via `http.Server.Shutdown`), keeping the
library free of signal handling and exit-code policy.

## Error handling

`App.Run` and `App.RunContext` log errors to stderr and return them. Map the
returned error to a process exit code with `dekrit.ExitCode`.

A command controls its exit code by returning an error that implements
`dekrit.ExitCoder`:

```go
type ExitCoder interface {
	ExitCode() int
}
```

`dekrit.ExitError{Code: n, Err: err}` is the built-in implementation, but any
error type (including wrapped ones) with an `ExitCode() int` method is honored.
Errors without one map to exit code 1.

## Input and output

Two sub packages provide common terminal helpers. Both are optional and do not
depend on the root package.

### `dekrit/output`

```go
import "github.com/yursan9/dekrit/output"

// Indented JSON.
output.JSON(ctx.Stdout, result)

// An aligned table from structs. Headers come from field names; override with
// a col:"..." tag and exclude a field with col:"-".
output.PrintTable(ctx.Stdout, []row{
	{Name: "alice", Age: 30},
	{Name: "bob", Age: 5},
})

// ...or build a table row by row.
t := output.NewTable("NAME", "AGE")
t.AddRow("alice", 30)
t.Write(ctx.Stdout)

// Progress bar and spinner (animated only on a terminal). Both default to
// os.Stdout; use WithWriter and WithPrefix to customize.
p := output.NewProgress(output.WithWriter(ctx.Stdout), output.WithPrefix("Downloading"))
p.SetTotal(100)
p.Start()
p.Add(1)
p.Finish()

sp := output.NewSpinner(output.WithWriter(ctx.Stdout), output.WithPrefix("Uploading"))
sp.Start("file.txt")
sp.Success("done") // or sp.Fail("boom") / sp.Stop()

// Styled text; Fprint disables styling when the writer is not a terminal.
output.Fprintf(ctx.Stdout, output.FgGreen, "ok %d\n", n)
```

### `dekrit/input`

```go
import "github.com/yursan9/dekrit/input"

p := input.New(input.WithInput(ctx.Stdin), input.WithOutput(ctx.Stdout))

ok, err := p.Confirm("overwrite?", false)
name, err := p.Text("name:")
host, err := p.TextDefault("host:", "localhost")
secret, err := p.Password("password:") // hidden on a terminal
idx, err := p.Select("pick one", []string{"a", "b", "c"})
```

Package-level convenience functions (`input.Confirm`, `input.Text`, etc.) use
`os.Stdin` and `os.Stdout`.

## License

Released under the MIT License. See [LICENSE](LICENSE).
