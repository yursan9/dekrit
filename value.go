package dekrit

import (
	"encoding"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

var (
	durationType        = reflect.TypeOf(time.Duration(0))
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

// rawValue is a flag.Value that merely records the raw textual values passed
// on the command line. Resolution against environment variables and defaults
// happens afterwards so that precedence is explicit.
type rawValue struct {
	provided bool
	vals     []string
	last     string
	isBool   bool
}

func (r *rawValue) String() string { return r.last }

func (r *rawValue) Set(s string) error {
	r.provided = true
	r.vals = append(r.vals, s)
	r.last = s
	return nil
}

func (r *rawValue) IsBoolFlag() bool { return r.isBool }

func isBoolType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Kind() == reflect.Bool
}

func isSliceType(t reflect.Type) bool {
	return t.Kind() == reflect.Slice && !isTextType(t)
}

// isTextType reports whether t, or a pointer to t, implements
// encoding.TextUnmarshaler.
func isTextType(t reflect.Type) bool {
	if t.Implements(textUnmarshalerType) {
		return true
	}
	return reflect.PointerTo(t).Implements(textUnmarshalerType)
}

func setField(dst reflect.Value, f *fieldInfo, vals []string) error {
	if isSliceType(dst.Type()) {
		for _, v := range vals {
			for _, part := range splitList(v) {
				if err := setValue(dst, part); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if len(vals) == 0 {
		return fmt.Errorf("no value")
	}
	return setValue(dst, vals[len(vals)-1])
}

func splitList(s string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}

func setValue(dst reflect.Value, s string) error {
	if dst.Kind() == reflect.Ptr {
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		return setValue(dst.Elem(), s)
	}

	if dst.Type() == durationType {
		d, err := time.ParseDuration(s)
		if err != nil {
			return fmt.Errorf("invalid duration %q", s)
		}
		dst.SetInt(int64(d))
		return nil
	}

	if u, ok := textUnmarshaler(dst); ok {
		if err := u.UnmarshalText([]byte(s)); err != nil {
			return err
		}
		return nil
	}

	switch dst.Kind() {
	case reflect.String:
		dst.SetString(s)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return fmt.Errorf("invalid boolean %q", s)
		}
		dst.SetBool(b)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 0, dst.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid integer %q", s)
		}
		dst.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(s, 0, dst.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid unsigned integer %q", s)
		}
		dst.SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(s, dst.Type().Bits())
		if err != nil {
			return fmt.Errorf("invalid number %q", s)
		}
		dst.SetFloat(n)
	case reflect.Slice:
		return setSliceValue(dst, s)
	default:
		return fmt.Errorf("unsupported type %s", dst.Type())
	}
	return nil
}

// textUnmarshaler returns a usable encoding.TextUnmarshaler for dst. A
// pointer receiver is preferred; a value receiver is accepted as a fallback.
func textUnmarshaler(dst reflect.Value) (encoding.TextUnmarshaler, bool) {
	if dst.CanAddr() {
		if u, ok := dst.Addr().Interface().(encoding.TextUnmarshaler); ok {
			return u, true
		}
	}
	if dst.CanInterface() {
		if u, ok := dst.Interface().(encoding.TextUnmarshaler); ok {
			return u, true
		}
	}
	return nil, false
}

func setSliceValue(dst reflect.Value, s string) error {
	ev := reflect.New(dst.Type().Elem()).Elem()
	if err := setValue(ev, s); err != nil {
		return err
	}
	dst.Set(reflect.Append(dst, ev))
	return nil
}
