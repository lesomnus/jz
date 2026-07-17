//go:build js

package jz

import "syscall/js"

// Get navigates through nested JavaScript object properties, like optional
// chaining, and returns the final value. It returns js.Undefined if any value
// along the path is not an object, so it never panics. Use Lookup instead when
// you need to know whether the whole path existed.
func Get(v js.Value, ps ...string) js.Value {
	r, _ := Lookup(v, ps...)
	return r
}

// Lookup navigates through nested JavaScript object properties.
// It returns the final value and true if every property along the path exists,
// or js.Undefined and false otherwise. It never panics: if an intermediate
// value is not an object (e.g. undefined or null), navigation stops.
func Lookup(v js.Value, ps ...string) (js.Value, bool) {
	for _, p := range ps {
		if !isNavigable(v) {
			return js.Undefined(), false
		}

		v = v.Get(p)
	}

	return v, !v.IsUndefined()
}

// isNavigable reports whether v.Get can be called without panicking, i.e. v is
// an object or a function.
func isNavigable(v js.Value) bool {
	switch v.Type() {
	case js.TypeObject, js.TypeFunction:
		return true
	default:
		return false
	}
}

// Object converts a Go map to a JavaScript object.
func Object(v map[string]any) js.Value {
	return js.ValueOf(v)
}
