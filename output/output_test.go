package output

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
)

func TestJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := JSON(&buf, map[string]any{"a": 1, "b": []int{2, 3}}); err != nil {
		t.Fatalf("JSON: %v", err)
	}
	want := "{\n  \"a\": 1,\n  \"b\": [\n    2,\n    3\n  ]\n}\n"
	if buf.String() != want {
		t.Fatalf("JSON = %q, want %q", buf.String(), want)
	}
}

func TestTable(t *testing.T) {
	var buf bytes.Buffer
	tbl := NewTable("NAME", "AGE")
	tbl.AddRow("alice", 30)
	tbl.AddRow("bob", 5)
	if err := tbl.Write(&buf); err != nil {
		t.Fatalf("Write: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"NAME", "AGE", "alice", "30", "bob", "5"} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing %q\n%s", want, out)
		}
	}
	if lines := strings.Split(strings.TrimRight(out, "\n"), "\n"); len(lines) != 3 {
		t.Errorf("expected 3 lines, got %d\n%s", len(lines), out)
	}
}

type BaseRow struct {
	ID int
}

type itemRow struct {
	BaseRow
	Title string
	Skip  string `col:"-"`
	Age   int    `col:"Years"`
}

func TestPrintTableStruct(t *testing.T) {
	var buf bytes.Buffer
	rows := []itemRow{
		{BaseRow{1}, "first", "hidden", 30},
		{BaseRow{2}, "second", "hidden", 5},
	}
	if err := PrintTable(&buf, rows); err != nil {
		t.Fatalf("PrintTable: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"ID", "Title", "Years", "first", "30", "second"} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing %q\n%s", want, out)
		}
	}
	if strings.Contains(out, "Skip") || strings.Contains(out, "hidden") {
		t.Errorf("col:\"-\" field should be excluded\n%s", out)
	}
}

func TestPrintTableEmptySliceHeaders(t *testing.T) {
	var buf bytes.Buffer
	if err := PrintTable(&buf, []itemRow{}); err != nil {
		t.Fatalf("PrintTable: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"ID", "Title", "Years"} {
		if !strings.Contains(out, want) {
			t.Errorf("empty table missing header %q\n%q", want, out)
		}
	}
}

func TestPrintTableErrors(t *testing.T) {
	var buf bytes.Buffer
	if err := PrintTable(&buf, 42); err == nil {
		t.Error("expected error for non-slice")
	}
}

func TestStyle(t *testing.T) {
	NoColor = true
	defer func() { NoColor = false }()
	if got := Bold.Sprint("x"); got != "x" {
		t.Errorf("NoColor Sprint = %q, want x", got)
	}
}

func TestStyleEnabled(t *testing.T) {
	NoColor = false
	if got := Bold.Sprint("x"); got != "\x1b[1mx\x1b[0m" {
		t.Errorf("Sprint = %q", got)
	}
	if Enabled(&bytes.Buffer{}) {
		t.Error("buffer should not be color-enabled")
	}
}

func TestProgress(t *testing.T) {
	p := NewProgress(WithWriter(&bytes.Buffer{}))
	p.SetTotal(100)
	p.Set(50)
	s := p.String()
	if !strings.Contains(s, "50%") || !strings.Contains(s, "(50/100)") {
		t.Fatalf("progress = %q", s)
	}
	if !strings.HasPrefix(s, "[===============") {
		t.Fatalf("unexpected bar fill: %q", s)
	}
}

func TestProgressPrefix(t *testing.T) {
	p := NewProgress(WithWriter(&bytes.Buffer{}), WithPrefix("Downloading"))
	p.SetTotal(4)
	p.Set(1)
	if got := p.String(); !strings.HasPrefix(got, "Downloading [") {
		t.Fatalf("progress with prefix = %q", got)
	}
}

func TestSpinner(t *testing.T) {
	var buf bytes.Buffer
	s := NewSpinner(WithWriter(&buf))
	s.Start("loading")
	if got := s.String(); got != "| loading" {
		t.Fatalf("spinner = %q, want \"| loading\"", got)
	}
	s.Success("done")
	if !strings.Contains(buf.String(), "[ok] done") {
		t.Fatalf("success output = %q", buf.String())
	}
}

func TestSpinnerPrefix(t *testing.T) {
	var buf bytes.Buffer
	s := NewSpinner(WithWriter(&buf), WithPrefix("Uploading"))
	s.Start("file.txt")
	if got := s.String(); got != "Uploading | file.txt" {
		t.Fatalf("spinner with prefix = %q", got)
	}
	s.Success("done")
	if !strings.Contains(buf.String(), "Uploading [ok] done") {
		t.Fatalf("success output = %q", buf.String())
	}
}

func TestOptionDefaults(t *testing.T) {
	p := NewProgress()
	if p.w != os.Stdout {
		t.Errorf("progress writer = %v, want os.Stdout", p.w)
	}
	s := NewSpinner()
	if s.w != os.Stdout {
		t.Errorf("spinner writer = %v, want os.Stdout", s.w)
	}
	if got := NewProgress(WithWriter(nil)).w; got != io.Writer(os.Stdout) {
		t.Errorf("nil writer should keep default, got %v", got)
	}
}
