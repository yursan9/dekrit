package dekrit_test

import (
	"context"
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
	ctx.Stdout.Write([]byte("serving\n"))
	return nil
}

func Example() {
	app := dekrit.New("app",
		dekrit.WithVersion("1.0.0"),
		dekrit.WithDescription("example application"),
	)
	app.Command("serve", &serveCmd{})
	_ = app.Run([]string{"serve", "--port", "9090", "config.yaml"})
}

// Example_runContext shows how to propagate a context to a command. In a real
// program the context would come from signal.NotifyContext.
func Example_runContext() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := dekrit.New("app")
	app.Command("serve", &serveCmd{})
	_ = app.RunContext(ctx, []string{"serve", "config.yaml"})
}
