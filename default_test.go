package dekrit

import (
	"strings"
	"testing"
)

type recCmd struct {
	Host string `dekrit:"flag" name:"host" default:"none"`
	In   string `dekrit:"pos" name:"in"`
	ran  bool
}

func (r *recCmd) Run(*Context) error {
	r.ran = true
	return nil
}

func TestDefaultSingleCommandEmptyArgs(t *testing.T) {
	a, _, _, _ := newTestApp(t)
	cmd := &recCmd{}
	a.Command("run", cmd)
	if err := a.Run(nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !cmd.ran || cmd.Host != "none" {
		t.Fatalf("default did not run: %+v", cmd)
	}
}

func TestDefaultSingleCommandFlagFallback(t *testing.T) {
	a, _, _, _ := newTestApp(t)
	cmd := &recCmd{}
	a.Command("run", cmd)
	if err := a.Run([]string{"--host", "example"}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !cmd.ran || cmd.Host != "example" {
		t.Fatalf("flags not routed to default: %+v", cmd)
	}
}

func TestDefaultSingleCommandPositionalFallback(t *testing.T) {
	a, _, _, _ := newTestApp(t)
	cmd := &recCmd{}
	a.Command("run", cmd)
	if err := a.Run([]string{"file.txt"}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if cmd.In != "file.txt" {
		t.Fatalf("positional not routed to default: %+v", cmd)
	}
}

func TestDefaultSingleCommandNamedInvocation(t *testing.T) {
	a, _, _, _ := newTestApp(t)
	cmd := &recCmd{}
	a.Command("run", cmd)
	if err := a.Run([]string{"run", "--host", "named"}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !cmd.ran || cmd.Host != "named" {
		t.Fatalf("named invocation failed: %+v", cmd)
	}
}

func TestDefaultSingleCommandHelp(t *testing.T) {
	a, out, _, _ := newTestApp(t)
	a.Command("run", &recCmd{})
	if err := a.Run([]string{"--help"}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	s := out.String()
	if !strings.Contains(s, "Usage: app [flags]") {
		t.Errorf("help missing app usage\n%s", s)
	}
	if !strings.Contains(s, "--host") {
		t.Errorf("help missing command flags\n%s", s)
	}
	if strings.Contains(s, "Commands:") {
		t.Errorf("single-command help should not list commands\n%s", s)
	}
}

func TestNamedHelpKeepsCommandName(t *testing.T) {
	a, out, _, _ := newTestApp(t)
	a.Command("run", &recCmd{})
	if err := a.Run([]string{"run", "--help"}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out.String(), "Usage: app run [flags]") {
		t.Errorf("named help must keep command name\n%s", out.String())
	}
}

func TestTwoCommandsNoDefault(t *testing.T) {
	a, out, _, _ := newTestApp(t)
	a.Command("run", &recCmd{})
	a.Command("other", &recCmd{})
	if err := a.Run(nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out.String(), "Commands:") {
		t.Errorf("expected app help with command list\n%s", out.String())
	}

	err := a.Run([]string{"--host", "x"})
	if err == nil {
		t.Fatal("expected unknown command error without a default")
	}
}

func TestZeroCommandsHelp(t *testing.T) {
	a, out, _, _ := newTestApp(t)
	if err := a.Run(nil); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out.String(), "Usage:") {
		t.Errorf("expected usage output\n%s", out.String())
	}
}
