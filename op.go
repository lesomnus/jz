//go:build js

package jz

import "syscall/js"

// Get navigates through nested JavaScript object properties.
// Returns the final value and true if all properties exist, or js.Undefined and false otherwise.
func Get(v js.Value, ps ...string) (js.Value, bool) {
	for _, p := range ps {
		if v.IsUndefined() {
			return v, false
		}

		v = v.Get(p)
	}

	return v, true
}

// GetX navigates through nested JavaScript object properties without checking existence.
// Returns js.Undefined if any property in the path doesn't exist.
func GetX(v js.Value, ps ...string) js.Value {
	for _, p := range ps {
		v = v.Get(p)
	}
	return v
}

// Object converts a Go map to a JavaScript object.
func Object(v map[string]any) js.Value {
	return js.ValueOf(v)
}
