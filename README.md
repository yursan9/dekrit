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

Planned (not yet implemented):

- Helpers to pretty print a struct as JSON or a tab based table
- Helpers to create loading progress with a loading bar or spinner

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
returned error to a process exit code with `dekrit.ExitCode`. Returning
`&dekrit.ExitError{Code: n, Err: err}` from a command sets a custom exit code.
