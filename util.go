//go:build js

package jz

import (
	"syscall/js"
)

// Stringify converts a value to its string representation.
func Stringify(jv js.Value) string {
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

		return GetX(js.Global(), "JSON", "stringify").Invoke(jv).String()
	default:
		return "<unknown>"
	}
}
