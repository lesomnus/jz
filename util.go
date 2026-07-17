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
		// A custom toString or JSON.stringify can throw (circular references,
		// BigInt, ...); never let that panic escape, since RejectedError.Error
		// relies on Stringify.
		if s, ok := tryString(func() js.Value { return js.Global().Call("String", jv) }); ok && s != "[object Object]" {
			return s
		}
		if s, ok := tryString(func() js.Value { return Get(js.Global(), "JSON", "stringify").Invoke(jv) }); ok {
			return s
		}
		return "[object Object]"
	default:
		return "<unknown>"
	}
}

// tryString evaluates f and returns its value as a string, reporting false if
// the underlying JavaScript threw (which syscall/js surfaces as a panic).
func tryString(f func() js.Value) (s string, ok bool) {
	defer func() {
		if recover() != nil {
			s, ok = "", false
		}
	}()
	return f().String(), true
}
