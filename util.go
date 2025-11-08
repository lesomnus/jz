//go:build js

package jz

import (
	"syscall/js"
)

func Stringify(v any) string {
	jv, ok := v.(js.Value)
	if !ok {
		jv = js.ValueOf(v)
	}

	switch jv.Type() {
	case js.TypeUndefined:
		return "undefined"
	case js.TypeNull:
		return "null"
	case js.TypeBoolean, js.TypeNumber:
		return jv.String()
	case js.TypeString:
		return `"` + jv.String() + `"`
	case js.TypeSymbol:
		return js.Global().Call("String", jv).String()
	case js.TypeFunction:
		name := jv.Get("name")
		if name.Truthy() {
			return "[Function: " + name.String() + "]"
		}
		return "[Function]"
	case js.TypeObject:
		s := js.Global().Call("String", jv).String()
		if s != "[object Object]" {
			return s
		}

		return GetX(js.Global(), "JSON", "stringify").Invoke(v).String()
	default:
		return "<unknown>"
	}
}
