//go:build js

package jz_test

import (
	"syscall/js"
	"testing"

	"github.com/lesomnus/jz"
	"github.com/lesomnus/jz/internal/x"
)

func TestStringify(t *testing.T) {
	t.Run("scalars", func(t *testing.T) {
		x.Equal(t, `undefined`, jz.Stringify(js.Undefined()))
		x.Equal(t, `null`, jz.Stringify(js.Null()))
		x.Equal(t, `<boolean: true>`, jz.Stringify(js.ValueOf(true)))
		x.Equal(t, `<boolean: false>`, jz.Stringify(js.ValueOf(false)))
		x.Equal(t, `<number: 0>`, jz.Stringify(js.ValueOf(0)))
		x.Equal(t, `<number: 42>`, jz.Stringify(js.ValueOf(42)))
		x.Equal(t, `<number: 3.14>`, jz.Stringify(js.ValueOf(3.14)))
		x.Equal(t, `"foo"`, jz.Stringify(js.ValueOf("foo")))
	})
	t.Run("Symbol", func(t *testing.T) {
		v := js.Global().Call("Symbol", "foo")
		s := jz.Stringify(v)
		x.Equal(t, "Symbol(foo)", s)
	})
	t.Run("Function", func(t *testing.T) {
		v := js.Global().Get("parseInt")
		s := jz.Stringify(v)
		x.Equal(t, "[Function: parseInt]", s)
	})
	t.Run("Object", func(t *testing.T) {
		v := js.ValueOf(map[string]any{})
		v.Set("foo", "bar")
		v.Set("baz", 42)

		s := jz.Stringify(v)
		x.Equal(t, `{"foo":"bar","baz":42}`, s)
	})
	t.Run("custom toString", func(t *testing.T) {
		v := js.Global().Get("Object").New()
		v.Set("toString", js.FuncOf(func(this js.Value, args []js.Value) any {
			return "custom"
		}))
		s := jz.Stringify(v)
		x.Equal(t, "custom", s)
	})
}

func TestStringifyDoesNotPanicOnCircular(t *testing.T) {

	obj := js.Global().Get("Object").New()
	obj.Set("self", obj) // circular reference: JSON.stringify throws

	var s string
	x.NotPanics(t, func() { s = jz.Stringify(obj) })
	x.Equal(t, "[object Object]", s)
}

func TestRejectedErrorDoesNotPanicOnCircular(t *testing.T) {

	obj := js.Global().Get("Object").New()
	obj.Set("self", obj)

	err := jz.RejectedError{Value: obj}
	x.NotPanics(t, func() { _ = err.Error() })
}

func TestStringifyArray(t *testing.T) {
	// An Array stringifies via String() (comma-joined), like JS itself.
	v := js.ValueOf([]any{1, 2, 3})
	x.Equal(t, "1,2,3", jz.Stringify(v))
}
