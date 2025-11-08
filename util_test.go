//go:build js

package jz_test

import (
	"syscall/js"
	"testing"

	"github.com/lesomnus/jz"
	"github.com/stretchr/testify/require"
)

func TestStringify(t *testing.T) {
	t.Run("scalars", func(t *testing.T) {
		x := require.New(t)
		x.Equal(`undefined`, jz.Stringify(js.Undefined()))
		x.Equal(`null`, jz.Stringify(js.Null()))
		x.Equal(`<boolean: true>`, jz.Stringify(js.ValueOf(true)))
		x.Equal(`<boolean: false>`, jz.Stringify(js.ValueOf(false)))
		x.Equal(`<number: 0>`, jz.Stringify(js.ValueOf(0)))
		x.Equal(`<number: 42>`, jz.Stringify(js.ValueOf(42)))
		x.Equal(`<number: 3.14>`, jz.Stringify(js.ValueOf(3.14)))
		x.Equal(`"foo"`, jz.Stringify(js.ValueOf("foo")))
	})
	t.Run("Symbol", func(t *testing.T) {
		v := js.Global().Call("Symbol", "foo")
		s := jz.Stringify(v)
		require.Equal(t, "Symbol(foo)", s)
	})
	t.Run("Function", func(t *testing.T) {
		v := js.Global().Get("parseInt")
		s := jz.Stringify(v)
		require.Equal(t, "[Function: parseInt]", s)
	})
	t.Run("Object", func(t *testing.T) {
		v := js.ValueOf(map[string]any{"foo": "bar", "baz": 42})
		s := jz.Stringify(v)
		require.Equal(t, `{"foo":"bar","baz":42}`, s)
	})
	t.Run("custom toString", func(t *testing.T) {
		v := js.Global().Get("Object").New()
		v.Set("toString", js.FuncOf(func(this js.Value, args []js.Value) any {
			return "custom"
		}))
		s := jz.Stringify(v)
		require.Equal(t, "custom", s)
	})
}
