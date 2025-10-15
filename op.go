//go:build js

package jz

import "syscall/js"

func Get(v js.Value, ps ...string) (js.Value, bool) {
	for _, p := range ps {
		if v.IsUndefined() {
			return v, false
		}

		v = v.Get(p)
	}

	return v, true
}

func GetX(v js.Value, ps ...string) js.Value {
	for _, p := range ps {
		v = v.Get(p)
	}
	return v
}

func Object(v map[string]any) js.Value {
	return js.ValueOf(v)
}
