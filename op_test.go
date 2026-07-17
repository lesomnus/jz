//go:build js

package jz_test

import (
	"syscall/js"
	"testing"

	"github.com/lesomnus/jz"
	"github.com/lesomnus/jz/internal/x"
)

func TestLookup(t *testing.T) {
	obj := js.ValueOf(map[string]any{
		"a": map[string]any{"b": 42},
	})

	t.Run("found", func(t *testing.T) {
		v, ok := jz.Lookup(obj, "a", "b")
		x.True(t, ok)
		x.Equal(t, 42, v.Int())
	})
	t.Run("missing final property", func(t *testing.T) {
		v, ok := jz.Lookup(obj, "a", "nope")
		x.False(t, ok)
		x.True(t, v.IsUndefined())
	})
	t.Run("missing intermediate does not panic", func(t *testing.T) {
		v, ok := jz.Lookup(obj, "x", "y", "z")
		x.False(t, ok)
		x.True(t, v.IsUndefined())
	})
	t.Run("null intermediate does not panic", func(t *testing.T) {
		o := js.ValueOf(map[string]any{"a": nil})
		v, ok := jz.Lookup(o, "a", "b")
		x.False(t, ok)
		x.True(t, v.IsUndefined())
	})
}

func TestGet(t *testing.T) {
	t.Run("existing path", func(t *testing.T) {
		v := jz.Get(js.Global(), "JSON", "stringify")
		x.Equal(t, js.TypeFunction, v.Type())
	})
	t.Run("missing path does not panic", func(t *testing.T) {
		v := jz.Get(js.Global(), "nope", "deeper", "still")
		x.True(t, v.IsUndefined())
	})
	t.Run("non-object intermediate does not panic", func(t *testing.T) {
		o := js.ValueOf(map[string]any{"n": 42})
		v := jz.Get(o, "n", "toFixed") // n is a number, not navigable
		x.True(t, v.IsUndefined())
	})
}

func TestObject(t *testing.T) {
	v := jz.Object(map[string]any{"foo": "bar", "n": 42})
	x.Equal(t, js.TypeObject, v.Type())
	x.Equal(t, "bar", v.Get("foo").String())
	x.Equal(t, 42, v.Get("n").Int())
}
