//go:build js

package jz

import "syscall/js"

func Stringify(v any) string {
	jv := js.ValueOf(v)
	if f, ok := Get(jv, "toString"); ok && f.Type() == js.TypeFunction {
		return f.Invoke(jv).String()
	}

	return GetX(js.Global(), "JSON", "stringify").Invoke(v).String()
}
