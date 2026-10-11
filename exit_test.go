package dekrit

import (
	"errors"
	"fmt"
	"testing"
)

type codeErr struct {
	code int
	msg  string
}

func (e *codeErr) Error() string { return e.msg }
func (e *codeErr) ExitCode() int { return e.code }

var _ ExitCoder = (*ExitError)(nil)

func TestExitCodeFromExitCoder(t *testing.T) {
	if got := ExitCode(&codeErr{code: 42, msg: "nope"}); got != 42 {
		t.Fatalf("ExitCode = %d, want 42", got)
	}
}

func TestExitCodeWrappedExitCoder(t *testing.T) {
	err := fmt.Errorf("context: %w", &codeErr{code: 3, msg: "boom"})
	if got := ExitCode(err); got != 3 {
		t.Fatalf("ExitCode = %d, want 3", got)
	}
}

func TestExitCodeExitError(t *testing.T) {
	if got := ExitCode(&ExitError{Code: 7}); got != 7 {
		t.Fatalf("ExitCode = %d, want 7", got)
	}
}

func TestExitCodeDefaults(t *testing.T) {
	if got := ExitCode(nil); got != 0 {
		t.Fatalf("ExitCode(nil) = %d, want 0", got)
	}
	if got := ExitCode(errors.New("plain")); got != 1 {
		t.Fatalf("ExitCode(plain) = %d, want 1", got)
	}
}

type codedCmd struct{ err error }

func (c *codedCmd) Run(*Context) error { return c.err }

func TestRunHonorsExitCoder(t *testing.T) {
	a, _, _, _ := newTestApp(t)
	a.Command("fail", &codedCmd{err: &codeErr{code: 5, msg: "custom"}})
	err := a.Run([]string{"fail"})
	if err == nil {
		t.Fatal("expected error")
	}
	if got := ExitCode(err); got != 5 {
		t.Fatalf("ExitCode = %d, want 5", got)
	}
}
