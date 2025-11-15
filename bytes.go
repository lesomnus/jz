//go:build js

package jz

import "syscall/js"

// BytesToGo converts a JavaScript Uint8Array to a Go byte slice.
func BytesToGo(v js.Value) []byte {
	b := make([]byte, v.Length())
	js.CopyBytesToGo(b, v)
	return b
}

// BytesToJs converts a Go byte slice to a JavaScript Uint8Array.
func BytesToJs(v []byte) js.Value {
	a := js.Global().Get("Uint8Array").New(len(v))
	js.CopyBytesToJS(a, v)
	return a
}
