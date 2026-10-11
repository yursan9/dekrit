package dekrit

import (
	"strings"
	"testing"
)

type helpCmd struct {
	Host string `dekrit:"flag,env" name:"host" short:"H" env:"APP_HOST" default:"localhost" help:"host to bind"`
	Verb bool   `dekrit:"flag" name:"verbose" short:"v" help:"verbose output"`
	File string `dekrit:"pos" name:"file" required:"true" help:"input file"`
}

func (h *helpCmd) Run(*Context) error { return nil }
func (h *helpCmd) Describe() string   { return "start the server" }

func TestAppHelp(t *testing.T) {
	a, out, _, _ := newTestApp(t)
	a = New("app", WithStdout(out), WithDescription("my tool"))
	a.Command("server", &helpCmd{})
	a.Command("other", &echoCmd{})
	if err := a.Run([]string{"help"}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	s := out.String()
	for _, want := range []string{"Usage: app", "my tool", "Commands:", "server", "start the server"} {
		if !strings.Contains(s, want) {
			t.Errorf("help missing %q\n%s", want, s)
		}
	}
}

func TestCommandHelp(t *testing.T) {
	a, out, _, _ := newTestApp(t)
	a = New("app", WithStdout(out))
	a.Command("server", &helpCmd{})
	if err := a.Run([]string{"server", "--help"}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	s := out.String()
	for _, want := range []string{
		"Usage: app server [flags] <file>",
		"-H, --host string",
		`(default "localhost")`,
		"[env: APP_HOST]",
		"--verbose",
		"<file>",
		"(required)",
		"-h, --help",
	} {
		if !strings.Contains(s, want) {
			t.Errorf("command help missing %q\n%s", want, s)
		}
	}
}

func TestInvalidFlagValue(t *testing.T) {
	a, _, errb, _ := newTestApp(t)
	cmd := &serverCmd{}
	a.Command("server", cmd)
	err := a.Run([]string{"server", "--port", "notanumber"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errb.String(), "invalid integer") {
		t.Errorf("stderr = %q", errb.String())
	}
}
