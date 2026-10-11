package dekrit

import (
	"reflect"
	"testing"
)

func TestParseFieldRoles(t *testing.T) {
	type S struct {
		A string `dekrit:"flag" name:"a"`
		B string `dekrit:"pos"`
		C string `dekrit:"env"`
		D string `dekrit:"flag,env" name:"d"`
	}
	fields, err := collectFields(reflect.TypeOf(S{}), "app")
	if err != nil {
		t.Fatalf("collectFields: %v", err)
	}
	if len(fields) != 4 {
		t.Fatalf("got %d fields, want 4", len(fields))
	}
	if !fields[0].isFlag() {
		t.Error("field A should be a flag")
	}
	if !fields[1].isPos() {
		t.Error("field B should be positional")
	}
	if !fields[2].isEnv() || fields[2].Env != "APP_C" {
		t.Errorf("field C env = %q, want APP_C", fields[2].Env)
	}
	if !fields[3].isFlag() || !fields[3].isEnv() {
		t.Error("field D should be flag+env")
	}
	if fields[1].Name != "b" {
		t.Errorf("positional default name = %q, want b", fields[1].Name)
	}
}

func TestParseFieldErrors(t *testing.T) {
	tests := []struct {
		name string
		typ  reflect.Type
	}{
		{"unknown role", reflect.TypeOf(struct {
			A string `dekrit:"nope"`
		}{})},
		{"flag without name", reflect.TypeOf(struct {
			A string `dekrit:"flag"`
		}{})},
		{"pos with short", reflect.TypeOf(struct {
			A string `dekrit:"pos" short:"a"`
		}{})},
		{"short without flag", reflect.TypeOf(struct {
			A string `dekrit:"env" short:"a"`
		}{})},
		{"short too long", reflect.TypeOf(struct {
			A string `dekrit:"flag" name:"a" short:"ab"`
		}{})},
		{"env tag without role", reflect.TypeOf(struct {
			A string `dekrit:"flag" name:"a" env:"X"`
		}{})},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := collectFields(tt.typ, "app"); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestEmbeddedStruct(t *testing.T) {
	type Common struct {
		Region string `dekrit:"flag" name:"region" default:"us"`
	}
	type Outer struct {
		Common
		Name string `dekrit:"flag" name:"name"`
	}
	fields, err := collectFields(reflect.TypeOf(Outer{}), "app")
	if err != nil {
		t.Fatalf("collectFields: %v", err)
	}
	if len(fields) != 2 {
		t.Fatalf("got %d fields, want 2", len(fields))
	}
	if len(fields[0].Index) != 2 {
		t.Errorf("embedded field index = %v, want length 2", fields[0].Index)
	}
}
