package dekrit

import (
	"context"
	"testing"
	"time"
)

type cancelCmd struct {
	started chan struct{}
	errored error
	hasDead bool
	done    bool
}

func (c *cancelCmd) Run(ctx *Context) error {
	if c.started != nil {
		close(c.started)
	}
	_, c.hasDead = ctx.Deadline()
	<-ctx.Done()
	c.done = true
	c.errored = ctx.Err()
	return nil
}

func TestRunContextCancellation(t *testing.T) {
	a, _, _, _ := newTestApp(t)
	cmd := &cancelCmd{started: make(chan struct{})}
	a.Command("wait", cmd)

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- a.RunContext(ctx, []string{"wait"}) }()

	<-cmd.started
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("RunContext: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("RunContext did not return after cancel")
	}
	if !cmd.done {
		t.Fatal("command did not observe cancellation")
	}
	if cmd.errored != context.Canceled {
		t.Fatalf("ctx.Err() = %v, want context.Canceled", cmd.errored)
	}
}

func TestRunContextDeadline(t *testing.T) {
	a, _, _, _ := newTestApp(t)
	cmd := &cancelCmd{}
	a.Command("wait", cmd)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := a.RunContext(ctx, []string{"wait"}); err != nil {
		t.Fatalf("RunContext: %v", err)
	}
	if !cmd.hasDead {
		t.Fatal("command did not receive a deadline")
	}
	if cmd.errored != context.DeadlineExceeded {
		t.Fatalf("ctx.Err() = %v, want context.DeadlineExceeded", cmd.errored)
	}
}

func TestRunContextCanceledBeforeRun(t *testing.T) {
	a, _, _, _ := newTestApp(t)
	cmd := &cancelCmd{}
	a.Command("wait", cmd)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := a.RunContext(ctx, []string{"wait"}); err != nil {
		t.Fatalf("RunContext: %v", err)
	}
	if !cmd.done || cmd.errored != context.Canceled {
		t.Fatalf("command did not observe pre-canceled context: done=%v err=%v", cmd.done, cmd.errored)
	}
}

type probeCmd struct {
	errored error
	hasCtx  bool
}

func (p *probeCmd) Run(ctx *Context) error {
	p.errored = ctx.Err()
	p.hasCtx = ctx.Context != nil
	return nil
}

func TestRunUsesBackgroundContext(t *testing.T) {
	a, _, _, _ := newTestApp(t)
	cmd := &probeCmd{}
	a.Command("probe", cmd)
	if err := a.Run([]string{"probe"}); err != nil {
		t.Fatalf("Run: %v", err)
	}
	if cmd.errored != nil {
		t.Errorf("ctx.Err() = %v, want nil", cmd.errored)
	}
	if !cmd.hasCtx {
		t.Error("Context did not carry a context.Context")
	}
}
