package dekrit

import (
	"bytes"
	"errors"
	"testing"
)

type echoCmd struct{}

func (e *echoCmd) Run(ctx *Context) error {
	for _, a := range ctx.Args {
		ctx.Stdout.Write([]byte(a))
	}
	return nil
}

func (e *echoCmd) Describe() string { return "echo arguments" }

func TestRouting(t *testing.T) {
	a, out, _, _ := newTestApp(t)
	a.Command("echo", &echoCmd{})
	if err := a.Run([]string{"echo", "hi"}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if out.String() != "hi" {
		t.Fatalf("stdout = %q, want hi", out.String())
	}
}

func TestUnknownCommand(t *testing.T) {
	a, _, errb, _ := newTestApp(t)
	a.Command("echo", &echoCmd{})
	err := a.Run([]string{"nope"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !bytes.Contains(errb.Bytes(), []byte("unknown command")) {
		t.Errorf("stderr = %q", errb.String())
	}
}

func TestVersion(t *testing.T) {
	a, out, _, _ := newTestApp(t)
	a = New("app", WithStdout(out), WithVersion("1.2.3"))
	a.Command("echo", &echoCmd{})
	if err := a.Run([]string{"--version"}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if out.String() != "app 1.2.3\n" {
		t.Fatalf("stdout = %q", out.String())
	}
}

type badCmd struct {
	A string `dekrit:"flag"`
}

func (b *badCmd) Run(*Context) error { return nil }

func TestInvalidRegistration(t *testing.T) {
	a, _, errb, _ := newTestApp(t)
	a.Command("bad", &badCmd{})
	err := a.Run([]string{"bad"})
	if err == nil {
		t.Fatal("expected registration error")
	}
	if !bytes.Contains(errb.Bytes(), []byte("name tag")) {
		t.Errorf("stderr = %q", errb.String())
	}
}

type hCmd struct {
	Help string `dekrit:"flag" name:"h"`
}

func (h *hCmd) Run(*Context) error { return nil }

func TestReservedFlag(t *testing.T) {
	a, _, _, _ := newTestApp(t)
	a.Command("x", &hCmd{})
	if err := a.Run([]string{"x"}); err == nil {
		t.Fatal("expected reserved flag error")
	}
}

type failCmd struct{}

func (f *failCmd) Run(*Context) error { return &ExitError{Code: 2, Err: errors.New("boom")} }

func TestRunErrorLogged(t *testing.T) {
	a, _, errb, _ := newTestApp(t)
	a.Command("fail", &failCmd{})
	err := a.Run([]string{"fail"})
	if err == nil {
		t.Fatal("expected error")
	}
	if ExitCode(err) != 2 {
		t.Errorf("exit code = %d, want 2", ExitCode(err))
	}
	if !bytes.Contains(errb.Bytes(), []byte("boom")) {
		t.Errorf("stderr = %q", errb.String())
	}
}
