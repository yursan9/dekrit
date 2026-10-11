package input

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func newTest(in string) (*Prompter, *bytes.Buffer) {
	out := &bytes.Buffer{}
	return New(WithInput(strings.NewReader(in)), WithOutput(out)), out
}

func TestNewDefaults(t *testing.T) {
	p := New()
	if p.In != os.Stdin {
		t.Errorf("In = %v, want os.Stdin", p.In)
	}
	if p.Out != os.Stdout {
		t.Errorf("Out = %v, want os.Stdout", p.Out)
	}
}

func TestOptionsIgnoreNil(t *testing.T) {
	p := New(WithInput(nil), WithOutput(nil))
	if p.In != os.Stdin || p.Out != os.Stdout {
		t.Errorf("nil options should not override defaults: %+v", p)
	}
}

func TestConfirm(t *testing.T) {
	tests := []struct {
		name  string
		input string
		def   bool
		want  bool
	}{
		{"yes", "y\n", false, true},
		{"no", "n\n", true, false},
		{"default yes", "\n", true, true},
		{"default no", "\n", false, false},
		{"word yes", "yes\n", false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, _ := newTest(tt.input)
			got, err := p.Confirm("continue?", tt.def)
			if err != nil {
				t.Fatalf("Confirm: %v", err)
			}
			if got != tt.want {
				t.Fatalf("Confirm = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConfirmRetries(t *testing.T) {
	p, out := newTest("maybe\nn\n")
	got, err := p.Confirm("continue?", true)
	if err != nil {
		t.Fatalf("Confirm: %v", err)
	}
	if got {
		t.Fatal("Confirm = true, want false")
	}
	if !strings.Contains(out.String(), "please answer y or n") {
		t.Errorf("expected reprompt, got %q", out.String())
	}
}

func TestText(t *testing.T) {
	p, _ := newTest("\nhello\n")
	got, err := p.Text("name:")
	if err != nil {
		t.Fatalf("Text: %v", err)
	}
	if got != "hello" {
		t.Fatalf("Text = %q, want hello", got)
	}
}

func TestTextDefault(t *testing.T) {
	p, _ := newTest("\n")
	got, err := p.TextDefault("name:", "anon")
	if err != nil {
		t.Fatalf("TextDefault: %v", err)
	}
	if got != "anon" {
		t.Fatalf("TextDefault = %q, want anon", got)
	}
}

func TestPasswordFallback(t *testing.T) {
	p, out := newTest("secret\n")
	got, err := p.Password("password:")
	if err != nil {
		t.Fatalf("Password: %v", err)
	}
	if got != "secret" {
		t.Fatalf("Password = %q, want secret", got)
	}
	if !strings.Contains(out.String(), "password:") {
		t.Errorf("prompt not written: %q", out.String())
	}
}

func TestSelect(t *testing.T) {
	p, _ := newTest("2\n")
	idx, err := p.Select("pick", []string{"a", "b", "c"})
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if idx != 1 {
		t.Fatalf("Select = %d, want 1", idx)
	}
}

func TestSelectRetries(t *testing.T) {
	p, out := newTest("9\n1\n")
	idx, err := p.Select("pick", []string{"a", "b"})
	if err != nil {
		t.Fatalf("Select: %v", err)
	}
	if idx != 0 {
		t.Fatalf("Select = %d, want 0", idx)
	}
	if !strings.Contains(out.String(), "enter a number between 1 and 2") {
		t.Errorf("expected range hint: %q", out.String())
	}
}

func TestSelectNoOptions(t *testing.T) {
	p, _ := newTest("")
	if _, err := p.Select("pick", nil); err == nil {
		t.Fatal("expected error for empty options")
	}
}
