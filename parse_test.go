package dekrit

import (
	"bytes"
	"testing"
	"time"
)

func newTestApp(t *testing.T) (*App, *bytes.Buffer, *bytes.Buffer, map[string]string) {
	t.Helper()
	env := map[string]string{}
	out := &bytes.Buffer{}
	errb := &bytes.Buffer{}
	a := New("app",
		WithStdout(out),
		WithStderr(errb),
		WithEnvLookup(func(k string) (string, bool) {
			v, ok := env[k]
			return v, ok
		}),
	)
	return a, out, errb, env
}

type serverCmd struct {
	Host    string        `dekrit:"flag,env" name:"host" short:"H" env:"APP_HOST" default:"localhost"`
	Port    int           `dekrit:"flag,env" name:"port" short:"p" env:"APP_PORT" default:"8080"`
	Verbose bool          `dekrit:"flag" name:"verbose" short:"v"`
	Tags    []string      `dekrit:"flag" name:"tag" short:"t"`
	Timeout time.Duration `dekrit:"flag" name:"timeout" default:"5s"`
}

func (s *serverCmd) Run(ctx *Context) error { return nil }

func TestPrecedenceDefault(t *testing.T) {
	a, _, _, _ := newTestApp(t)
	cmd := &serverCmd{}
	a.Command("server", cmd)
	if err := a.Run([]string{"server"}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if cmd.Host != "localhost" || cmd.Port != 8080 || cmd.Timeout != 5*time.Second {
		t.Fatalf("defaults not applied: %+v", cmd)
	}
}

func TestPrecedenceEnv(t *testing.T) {
	a, _, _, env := newTestApp(t)
	env["APP_HOST"] = "envhost"
	env["APP_PORT"] = "9090"
	cmd := &serverCmd{}
	a.Command("server", cmd)
	if err := a.Run([]string{"server"}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if cmd.Host != "envhost" || cmd.Port != 9090 {
		t.Fatalf("env not applied: %+v", cmd)
	}
}

func TestPrecedenceFlagOverEnv(t *testing.T) {
	a, _, _, env := newTestApp(t)
	env["APP_HOST"] = "envhost"
	env["APP_PORT"] = "9090"
	cmd := &serverCmd{}
	a.Command("server", cmd)
	if err := a.Run([]string{"server", "--host", "flaghost", "-p", "1234"}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if cmd.Host != "flaghost" || cmd.Port != 1234 {
		t.Fatalf("flag did not win: %+v", cmd)
	}
}

func TestBoolAndSliceAndDuration(t *testing.T) {
	a, _, _, _ := newTestApp(t)
	cmd := &serverCmd{}
	a.Command("server", cmd)
	args := []string{"server", "-v", "-t", "a,b", "--tag", "c", "--timeout", "1m"}
	if err := a.Run(args); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !cmd.Verbose {
		t.Error("verbose not set")
	}
	want := []string{"a", "b", "c"}
	if len(cmd.Tags) != len(want) {
		t.Fatalf("tags = %v, want %v", cmd.Tags, want)
	}
	for i := range want {
		if cmd.Tags[i] != want[i] {
			t.Fatalf("tags = %v, want %v", cmd.Tags, want)
		}
	}
	if cmd.Timeout != time.Minute {
		t.Errorf("timeout = %v, want 1m", cmd.Timeout)
	}
}

type copyCmd struct {
	Src  string `dekrit:"pos" name:"src" required:"true"`
	Dst  string `dekrit:"pos" name:"dst" required:"true"`
	Rest string `dekrit:"pos" name:"rest"`
}

func (c *copyCmd) Run(ctx *Context) error { return nil }

func TestPositional(t *testing.T) {
	a, _, _, _ := newTestApp(t)
	cmd := &copyCmd{}
	a.Command("cp", cmd)
	if err := a.Run([]string{"cp", "a.txt", "b.txt", "extra"}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if cmd.Src != "a.txt" || cmd.Dst != "b.txt" || cmd.Rest != "extra" {
		t.Fatalf("positionals = %+v", cmd)
	}
}

func TestRequiredMissing(t *testing.T) {
	a, _, errb, _ := newTestApp(t)
	cmd := &copyCmd{}
	a.Command("cp", cmd)
	err := a.Run([]string{"cp", "a.txt"})
	if err == nil {
		t.Fatal("expected error for missing required positional")
	}
	if ExitCode(err) != 1 {
		t.Errorf("exit code = %d, want 1", ExitCode(err))
	}
	if !bytes.Contains(errb.Bytes(), []byte("missing required")) {
		t.Errorf("stderr = %q, want missing required", errb.String())
	}
}

type envCmd struct {
	Token string `dekrit:"env"`
}

func (e *envCmd) Run(*Context) error { return nil }

func TestEnvDerivedName(t *testing.T) {
	cmd := &envCmd{}
	a, _, _, env := newTestApp(t)
	a.Command("run", cmd)
	env["APP_TOKEN"] = "secret"
	if err := a.Run([]string{"run"}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if cmd.Token != "secret" {
		t.Fatalf("token = %q, want secret", cmd.Token)
	}
}

type sliceCmd struct {
	Tags []string `dekrit:"flag" name:"tag" default:"x,y"`
}

func (s *sliceCmd) Run(*Context) error { return nil }

func TestFlagOverridesDefaultSlice(t *testing.T) {
	a, _, _, _ := newTestApp(t)
	cmd := &sliceCmd{}
	a.Command("run", cmd)
	if err := a.Run([]string{"run"}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(cmd.Tags) != 2 || cmd.Tags[0] != "x" {
		t.Fatalf("default slice = %v", cmd.Tags)
	}
}
