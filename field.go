package dekrit

import (
	"reflect"
)

// collectFields walks a struct type and extracts every field carrying a
// dekrit tag. Embedded structs that are not themselves dekrit fields are
// traversed, which is how configuration can be shared across subcommands.
func collectFields(t reflect.Type, appName string) ([]*fieldInfo, error) {
	var out []*fieldInfo
	if err := walkStruct(t, nil, appName, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func walkStruct(t reflect.Type, prefix []int, appName string, out *[]*fieldInfo) error {
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)

		index := make([]int, 0, len(prefix)+1)
		index = append(index, prefix...)
		index = append(index, i)

		if sf.Anonymous {
			ft := sf.Type
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				if _, ok := sf.Tag.Lookup(tagKey); !ok {
					if err := walkStruct(ft, index, appName, out); err != nil {
						return err
					}
					continue
				}
			}
		}

		if sf.PkgPath != "" {
			continue
		}

		fi, err := parseField(sf, index)
		if err != nil {
			return err
		}
		if fi == nil {
			continue
		}
		if fi.isEnv() && fi.Env == "" {
			fi.Env = envName(appName, sf.Name)
		}
		*out = append(*out, fi)
	}
	return nil
}

// fieldByIndexAlloc resolves a field index path, allocating any nil embedded
// pointers along the way so the returned value is settable.
func fieldByIndexAlloc(root reflect.Value, index []int) reflect.Value {
	v := root
	for _, i := range index {
		if v.Kind() == reflect.Ptr {
			if v.IsNil() {
				v.Set(reflect.New(v.Type().Elem()))
			}
			v = v.Elem()
		}
		v = v.Field(i)
	}
	return v
}
