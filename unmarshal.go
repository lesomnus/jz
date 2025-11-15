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
		return fmt.Errorf("invalid unmarshal from %q to %q", t.String(), v.Type().String())
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
			switch {
			case data.InstanceOf(js.Global().Get("ArrayBuffer")):
				data = js.Global().Get("Uint8Array").New(data)

				fallthrough
			case data.InstanceOf(js.Global().Get("Uint8Array")):
				c := unmarshalByteArray(data)
				v.Set(reflect.ValueOf(c))

			case js.Global().Get("Array").Call("isArray", data).Truthy():
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

			default:
				// JS Object -> map[string]any
				ks := js.Global().Get("Object").Call("keys", data)

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
				v.Set(w)
			}
		}

		return nil
	}

	switch t {
	case js.TypeUndefined:
		// Noop.

	case js.TypeNull:
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
			reflect.Uint64,
			reflect.Uintptr:
			v.SetUint(uint64(data.Int()))

		case reflect.Float32,
			reflect.Float64:
			v.SetFloat(data.Float())

		case reflect.Complex64,
			reflect.Complex128:
			v.SetComplex(complex(data.Float(), 0))

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
		case reflect.Map:
			// JS Object -> map[string]any
			ks := js.Global().Get("Object").Call("keys", data)
			l := ks.Length()
			if l == 0 {
				break
			}
			if v.IsNil() {
				v.Set(reflect.MakeMap(v.Type()))
			}

			dst := reflect.New(v.Type().Elem())
			for i := range l {
				k := ks.Index(i).String()

				src := data.Get(k)
				if err := unmarshal(src, dst.Elem()); err != nil {
					return fmt.Errorf(".%s: %w", k, err)
				}

				v.SetMapIndex(reflect.ValueOf(k), dst.Elem())
			}

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
			src_typename := ""
			switch {
			case data.InstanceOf(js.Global().Get("ArrayBuffer")):
				src_typename = "ArrayBuffer"
			case data.InstanceOf(js.Global().Get("Uint8Array")):
				src_typename = "Uint8Array"
			case js.Global().Get("Array").Call("isArray", data).Truthy():
				src_typename = "Array"
			default:
				return fmt.Errorf("invalid unmarshal from non-array object to %q", v.Type().String())
			}

			l := data.Length()
			if k == reflect.Array {
				l = min(l, v.Len())
			} else {
				v.Set(reflect.MakeSlice(v.Type(), l, l))
			}
			if l == 0 {
				break
			}

			switch src_typename {
			case "ArrayBuffer":
				data = js.Global().Get("Uint8Array").New(data)
				fallthrough
			case "Uint8Array":
				switch v.Type().Elem().Kind() {
				case reflect.Uint8:
					// fast path for []byte?
					c := unmarshalByteArray(data)
					v.SetBytes(c)
					return nil

				case reflect.Int,
					reflect.Int8,
					reflect.Int16,
					reflect.Int32,
					reflect.Int64:
					for i := range l {
						v.Index(i).SetInt(int64(data.Index(i).Int()))
					}

				case reflect.Uint,
					reflect.Uint16,
					reflect.Uint32,
					reflect.Uint64:
					for i := range l {
						v.Index(i).SetUint(uint64(data.Index(i).Int()))
					}

				case reflect.Float32,
					reflect.Float64:
					for i := range l {
						v.Index(i).SetFloat(data.Index(i).Float())
					}

				case reflect.Interface:
					for i := range l {
						v.Index(i).Set(reflect.ValueOf(data.Index(i).Float()))
					}

				default:
					// number -> non-number
					return fmt.Errorf("invalid unmarshal from %q to %q", src_typename, v.Type().String())
				}

			case "Array":
				for i := range l {
					src := data.Index(i)
					dst := v.Index(i)
					if err := unmarshal(src, dst); err != nil {
						return fmt.Errorf(".%s[%d]: %w", v.Type().Name(), i, err)
					}
				}

			default:
				panic("unreachable")
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

func unmarshalByteArray(data js.Value) []byte {
	l := data.Length()
	c := make([]byte, l)
	for i := range l {
		src := data.Index(i)
		c[i] = byte(src.Int())
	}

	return c
}
