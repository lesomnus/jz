//go:build js

package jz

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"syscall/js"
	"unicode"
)

func Unmarshal(data js.Value, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Pointer || rv.IsNil() {
		return errors.New("out value must a non-nil pointer")
	}

	return unmarshal(data, rv.Elem())
}

func unmarshal(data js.Value, v reflect.Value) error {
	if !v.CanSet() {
		return nil
	}
	if v.Type() == reflect.TypeOf(js.Value{}) {
		v.Set(reflect.ValueOf(data))
		return nil
	}

	t := data.Type()
	k := v.Kind()
	invalid_unmarshal_err := func() error {
		return fmt.Errorf("invalid unmarshal from %q to %q", t.String(), k.String())
	}

	switch k {
	case reflect.Pointer:
		next := reflect.New(v.Type().Elem())
		if err := unmarshal(data, next.Elem()); err != nil {
			return err
		}

		v.Set(next)
		return nil

	case reflect.Interface:
		switch t {
		case js.TypeNull, js.TypeUndefined:
			v.Set(reflect.Zero(v.Type()))
		case js.TypeNumber:
			v.Set(reflect.ValueOf(data.Float()))
		case js.TypeString:
			v.Set(reflect.ValueOf(data.String()))
		case js.TypeBoolean:
			v.Set(reflect.ValueOf(data.Bool()))
		case js.TypeObject:
			if js.Global().Get("Array").Call("isArray", data).Truthy() {
				// JS Array -> []any
				l := data.Length()
				c := make([]any, l)
				w := reflect.ValueOf(c)
				for i := range l {
					src := data.Index(i)
					dst := w.Index(i)
					if err := unmarshal(src, dst); err != nil {
						return fmt.Errorf(".%s[%d]: %w", v.Type().Name(), i, err)
					}
				}
				v.Set(w)

			} else {
				// JS Object -> map[string]any
				ks := js.Global().Get("Object").Call("keys", v)

				l := ks.Length()
				c := map[string]any{}
				w := reflect.ValueOf(c)
				for i := range l {
					k := ks.Index(i).String()
					var v_ any

					src := data.Get(k)
					dst := reflect.ValueOf(&v_).Elem()
					if err := unmarshal(src, dst); err != nil {
						return fmt.Errorf(".%s[%d]: %w", k, i, err)
					}

					w.SetMapIndex(reflect.ValueOf(k), dst)
				}
				v.Send(w)
			}
		}

		return nil
	}

	switch t {
	case js.TypeNull, js.TypeUndefined:
		v.Set(reflect.Zero(v.Type()))

	case js.TypeNumber:
		switch k {
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			v.SetInt(int64(data.Int()))

		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			v.SetUint(uint64(data.Int()))

		case reflect.Float32,
			reflect.Float64:
			v.SetFloat(data.Float())

		default:
			return invalid_unmarshal_err()
		}

	case js.TypeString:
		switch k {
		case reflect.String:
			v.SetString(data.String())

		default:
			return invalid_unmarshal_err()
		}

	case js.TypeBoolean:
		switch k {
		case reflect.Bool:
			v.SetBool(data.Bool())

		default:
			return invalid_unmarshal_err()
		}

	case js.TypeObject:
		switch k {
		case reflect.Struct:
			for i := range v.NumField() {
				f := v.Field(i)
				if !f.CanSet() {
					continue
				}

				vt := v.Type()
				vf := vt.Field(i)
				tag := vf.Tag.Get("json")
				if tag == "-" {
					continue
				}

				name, _, _ := strings.Cut(tag, ",")
				if name == "" {
					name = toLowerCamel(vf.Name)
				}

				next := data.Get(name)
				if err := unmarshal(next, f); err != nil {
					return fmt.Errorf(".%s: %w", vt.Name(), err)
				}
			}

		case reflect.Array, reflect.Slice:
			if !js.Global().Get("Array").Call("isArray", data).Truthy() {
				return fmt.Errorf("invalid unmarshal from non-array object to %q", k.String())
			}

			l := data.Length()
			if k == reflect.Array {
				l = min(l, v.Len())
			} else {
				v.Set(reflect.MakeSlice(v.Type(), l, l))
			}

			for i := range l {
				src := data.Index(i)
				dst := v.Index(i)
				if err := unmarshal(src, dst); err != nil {
					return fmt.Errorf(".%s[%d]: %w", v.Type().Name(), i, err)
				}
			}

		default:
			return invalid_unmarshal_err()
		}

	case js.TypeFunction, js.TypeSymbol:
		// Do nothing.

	default:
		panic("unknown type")
	}

	return nil
}

func toLowerCamel(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}
