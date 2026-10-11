package dekrit

import (
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"
)

type upper string

func (u *upper) UnmarshalText(b []byte) error {
	*u = upper(strings.ToUpper(string(b)))
	return nil
}

type semiList []string

func (s *semiList) UnmarshalText(b []byte) error {
	*s = strings.Split(string(b), ";")
	return nil
}

type rangeVal struct {
	Lo, Hi int
}

func (r *rangeVal) UnmarshalText(b []byte) error {
	parts := strings.SplitN(string(b), "-", 2)
	if len(parts) != 2 {
		return errors.New("want lo-hi")
	}
	lo, err := strconv.Atoi(parts[0])
	if err != nil {
		return err
	}
	hi, err := strconv.Atoi(parts[1])
	if err != nil {
		return err
	}
	r.Lo, r.Hi = lo, hi
	return nil
}

type textCmd struct {
	Name   upper     `dekrit:"flag,env" name:"name" short:"n" env:"APP_NAME" default:"fallback"`
	Tags   semiList  `dekrit:"flag" name:"tags"`
	Window rangeVal  `dekrit:"flag" name:"window" default:"1-2"`
	When   time.Time `dekrit:"flag" name:"when"`
}

func (t *textCmd) Run(*Context) error { return nil }

func TestTextUnmarshalerFlag(t *testing.T) {
	a, _, _, _ := newTestApp(t)
	cmd := &textCmd{}
	a.Command("run", cmd)
	args := []string{"run", "-n", "hello", "--tags", "a;b;c", "--window", "3-9", "--when", "2020-01-02T15:04:05Z"}
	if err := a.Run(args); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if cmd.Name != "HELLO" {
		t.Errorf("name = %q, want HELLO", cmd.Name)
	}
	if len(cmd.Tags) != 3 || cmd.Tags[0] != "a" || cmd.Tags[2] != "c" {
		t.Errorf("tags = %v, want [a b c]", cmd.Tags)
	}
	if cmd.Window.Lo != 3 || cmd.Window.Hi != 9 {
		t.Errorf("window = %+v, want {3 9}", cmd.Window)
	}
	want := time.Date(2020, 1, 2, 15, 4, 5, 0, time.UTC)
	if !cmd.When.Equal(want) {
		t.Errorf("when = %v, want %v", cmd.When, want)
	}
}

func TestTextUnmarshalerEnvAndDefault(t *testing.T) {
	a, _, _, env := newTestApp(t)
	env["APP_NAME"] = "from-env"
	cmd := &textCmd{}
	a.Command("run", cmd)
	if err := a.Run([]string{"run"}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if cmd.Name != "FROM-ENV" {
		t.Errorf("name = %q, want FROM-ENV", cmd.Name)
	}
	if cmd.Window.Lo != 1 || cmd.Window.Hi != 2 {
		t.Errorf("window default = %+v, want {1 2}", cmd.Window)
	}
}

type ptrTextCmd struct {
	Addr *upper `dekrit:"flag" name:"addr"`
}

func (p *ptrTextCmd) Run(*Context) error { return nil }

func TestTextUnmarshalerPointer(t *testing.T) {
	a, _, _, _ := newTestApp(t)
	cmd := &ptrTextCmd{}
	a.Command("run", cmd)
	if err := a.Run([]string{"run", "--addr", "x"}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if cmd.Addr == nil || *cmd.Addr != "X" {
		t.Fatalf("addr = %v, want X", cmd.Addr)
	}
}

type badTextCmd struct {
	Window rangeVal `dekrit:"flag" name:"window"`
}

func (b *badTextCmd) Run(*Context) error { return nil }

func TestTextUnmarshalerHelpPlaceholder(t *testing.T) {
	a, out, _, _ := newTestApp(t)
	a.Command("run", &textCmd{})
	if err := a.Run([]string{"run", "--help"}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out.String(), "--window value") {
		t.Errorf("help missing value placeholder\n%s", out.String())
	}
}

func TestTextUnmarshalerError(t *testing.T) {
	a, _, errb, _ := newTestApp(t)
	cmd := &badTextCmd{}
	a.Command("run", cmd)
	err := a.Run([]string{"run", "--window", "nope"})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(errb.String(), "want lo-hi") {
		t.Errorf("stderr = %q", errb.String())
	}
}
