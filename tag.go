package dekrit

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

const tagKey = "dekrit"

type role uint8

const (
	roleFlag role = 1 << iota
	rolePos
	roleEnv
)

func (r role) has(x role) bool { return r&x != 0 }

type fieldInfo struct {
	Index    []int
	GoName   string
	Name     string
	Short    string
	Env      string
	Help     string
	Default  string
	Required bool
	Hidden   bool
	Roles    role
	Type     reflect.Type

	posIndex int
	display  string
}

func (f *fieldInfo) isFlag() bool { return f.Roles.has(roleFlag) }
func (f *fieldInfo) isPos() bool  { return f.Roles.has(rolePos) }
func (f *fieldInfo) isEnv() bool  { return f.Roles.has(roleEnv) }

func (f *fieldInfo) label() string {
	if f.display != "" {
		return f.display
	}
	switch {
	case f.isFlag() && f.isPos():
		return "--" + f.Name + " <" + f.Name + ">"
	case f.isFlag():
		return "--" + f.Name
	case f.isPos():
		return "<" + f.Name + ">"
	default:
		return f.GoName
	}
}

func parseField(sf reflect.StructField, index []int) (*fieldInfo, error) {
	raw, ok := sf.Tag.Lookup(tagKey)
	if !ok {
		return nil, nil
	}

	fi := &fieldInfo{
		Index:   index,
		GoName:  sf.Name,
		Name:    strings.TrimSpace(sf.Tag.Get("name")),
		Short:   strings.TrimSpace(sf.Tag.Get("short")),
		Env:     strings.TrimSpace(sf.Tag.Get("env")),
		Help:    sf.Tag.Get("help"),
		Default: sf.Tag.Get("default"),
		Type:    sf.Type,
	}

	if v, ok := sf.Tag.Lookup("required"); ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("field %s: invalid required value %q", sf.Name, v)
		}
		fi.Required = b
	}
	if v, ok := sf.Tag.Lookup("hidden"); ok {
		b, err := strconv.ParseBool(v)
		if err != nil {
			return nil, fmt.Errorf("field %s: invalid hidden value %q", sf.Name, v)
		}
		fi.Hidden = b
	}

	for _, tok := range strings.Split(raw, ",") {
		switch strings.TrimSpace(tok) {
		case "":
		case "flag":
			fi.Roles |= roleFlag
		case "pos", "positional", "arg":
			fi.Roles |= rolePos
		case "env":
			fi.Roles |= roleEnv
		default:
			return nil, fmt.Errorf("field %s: unknown dekrit role %q", sf.Name, tok)
		}
	}

	if fi.Roles == 0 {
		return nil, fmt.Errorf("field %s: dekrit tag declares no role (want flag, pos, or env)", sf.Name)
	}
	if fi.isFlag() && fi.Name == "" {
		return nil, fmt.Errorf("field %s: flag requires a name tag", sf.Name)
	}
	if fi.isPos() && fi.Name == "" {
		fi.Name = strings.ToLower(sf.Name)
		fi.display = "<" + strings.ToLower(sf.Name) + ">"
	}
	if fi.isPos() && fi.Short != "" {
		return nil, fmt.Errorf("field %s: positional argument cannot have a short tag", sf.Name)
	}
	if fi.Short != "" && !fi.isFlag() {
		return nil, fmt.Errorf("field %s: short tag requires the flag role", sf.Name)
	}
	if fi.Short != "" && len([]rune(fi.Short)) != 1 {
		return nil, fmt.Errorf("field %s: short tag must be a single character", sf.Name)
	}
	if fi.Env != "" && !fi.isEnv() {
		return nil, fmt.Errorf("field %s: env tag requires the env role", sf.Name)
	}
	if fi.isFlag() {
		fi.display = "--" + fi.Name
	}

	return fi, nil
}

func envName(app, goName string) string {
	var b strings.Builder
	for _, r := range app {
		if r == '-' || r == ' ' {
			b.WriteRune('_')
			continue
		}
		b.WriteRune(r)
	}
	b.WriteByte('_')
	for _, r := range goName {
		b.WriteRune(r)
	}
	return strings.ToUpper(b.String())
}
