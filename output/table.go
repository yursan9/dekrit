package output

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"strings"
	"text/tabwriter"
)

// Table accumulates rows and renders them as an aligned text table using
// tab-aligned columns.
type Table struct {
	headers []string
	rows    [][]string
}

// NewTable creates a table with the given headers.
func NewTable(headers ...string) *Table {
	return &Table{headers: append([]string(nil), headers...)}
}

// AddRow appends a row, converting each column with fmt.Sprint.
func (t *Table) AddRow(cols ...any) {
	row := make([]string, len(cols))
	for i, c := range cols {
		row[i] = fmt.Sprint(c)
	}
	t.rows = append(t.rows, row)
}

// AddStruct appends a row built from the exported fields of v. Headers are
// taken from the field names, or a `col:"Name"` tag when present. A `col:"-"`
// tag excludes a field. Anonymous exported structs are flattened. When the
// table has no headers yet, they are derived from v.
func (t *Table) AddStruct(v any) error {
	var cols, headers []string
	if err := collectRow(reflect.ValueOf(v), &cols, &headers); err != nil {
		return err
	}
	if len(t.headers) == 0 {
		t.headers = headers
	}
	t.rows = append(t.rows, cols)
	return nil
}

// Write renders the table to w.
func (t *Table) Write(w io.Writer) error {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	if len(t.headers) > 0 {
		fmt.Fprintln(tw, strings.Join(t.headers, "\t"))
	}
	for _, row := range t.rows {
		fmt.Fprintln(tw, strings.Join(row, "\t"))
	}
	return tw.Flush()
}

// PrintTable renders a slice of structs as a table, deriving the headers from
// the element type.
func PrintTable(w io.Writer, v any) error {
	rv := reflect.ValueOf(v)
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return errors.New("output: nil value")
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice && rv.Kind() != reflect.Array {
		return fmt.Errorf("output: PrintTable expects a slice or array, got %s", rv.Kind())
	}
	t := &Table{}
	for i := 0; i < rv.Len(); i++ {
		if err := t.AddStruct(rv.Index(i).Interface()); err != nil {
			return err
		}
	}
	if len(t.rows) == 0 {
		t.headers = typeHeaders(rv.Type().Elem())
	}
	return t.Write(w)
}

// typeHeaders derives column headers from a struct type, used when a slice has
// no rows to inspect.
func typeHeaders(t reflect.Type) []string {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}
	var cols, headers []string
	_ = collectRow(reflect.New(t).Elem(), &cols, &headers)
	return headers
}

func collectRow(rv reflect.Value, cols, headers *[]string) error {
	for rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return errors.New("output: nil struct")
		}
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return fmt.Errorf("output: AddStruct expects a struct, got %s", rv.Kind())
	}

	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		fv := rv.Field(i)

		if sf.Anonymous {
			ft := sf.Type
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			_, tagged := sf.Tag.Lookup("col")
			if sf.PkgPath == "" && ft.Kind() == reflect.Struct && !tagged {
				if fv.Kind() == reflect.Ptr && fv.IsNil() {
					continue
				}
				if err := collectRow(fv, cols, headers); err != nil {
					return err
				}
				continue
			}
		}

		if sf.PkgPath != "" {
			continue
		}
		name := sf.Name
		if tag, ok := sf.Tag.Lookup("col"); ok {
			if tag == "-" {
				continue
			}
			if tag != "" {
				name = tag
			}
		}
		*cols = append(*cols, formatCell(fv))
		*headers = append(*headers, name)
	}
	return nil
}

func formatCell(v reflect.Value) string {
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return ""
		}
		v = v.Elem()
	}
	if !v.CanInterface() {
		return ""
	}
	return fmt.Sprint(v.Interface())
}
