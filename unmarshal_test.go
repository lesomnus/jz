//go:build js

package jz_test

import (
	"syscall/js"
	"testing"

	"github.com/lesomnus/jz"
	"github.com/stretchr/testify/require"
)

func TestUnmarshal(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		DoTestV(t, int8(42))
		DoTestV(t, int16(42))
		DoTestV(t, int32(42))
		DoTestV(t, int64(42))
	})
	t.Run("uint", func(t *testing.T) {
		DoTestV(t, uint8(42))
		DoTestV(t, uint16(42))
		DoTestV(t, uint32(42))
		DoTestV(t, uint64(42))
	})
	t.Run("string", func(t *testing.T) {
		DoTestV(t, "Royale with Cheese")
	})
	t.Run("bool", func(t *testing.T) {
		DoTestV(t, false)
		DoTestV(t, true)
	})
	t.Run("pointer to scalar", func(t *testing.T) {
		var v *string

		err := jz.Unmarshal(js.ValueOf("foo"), &v)
		require.NoError(t, err)
		require.NotNil(t, v)
		require.Equal(t, "foo", *v)
	})
	t.Run("pointer to struct", func(t *testing.T) {
		var v *struct{ Foo string }

		err := jz.Unmarshal(js.ValueOf(map[string]any{"foo": "bar"}), &v)
		require.NoError(t, err)
		require.NotNil(t, v)
		require.Equal(t, "bar", v.Foo)
	})
	t.Run("pointer in struct", func(t *testing.T) {
		var v struct{ Foo *string }

		err := jz.Unmarshal(js.ValueOf(map[string]any{"foo": "bar"}), &v)
		require.NoError(t, err)
		require.NotNil(t, v.Foo)
		require.Equal(t, "bar", *v.Foo)
	})
	t.Run("any", func(t *testing.T) {
		var v any = float64(42)
		DoTestV(t, v)
	})
	t.Run("any with slice", func(t *testing.T) {
		var v any = []any{"foo", 3.14, true, false}
		DoTestV(t, v)
	})
	t.Run("slice with zero", func(t *testing.T) {
		DoTest(t, []string{"foo", "bar", "baz"}, []any{"foo", "bar", "baz"})
	})
	t.Run("slice of any", func(t *testing.T) {
		var vs = []any{"foo", 3.14, true, false}
		DoTestV(t, vs)
	})
	t.Run("struct", func(t *testing.T) {
		type A struct {
			Int    int
			String string
			Bool   bool
		}

		DoTest(t,
			A{
				Int:    42,
				String: "Le Big Mac",
				Bool:   true,
			},
			map[string]any{
				"int":    42,
				"string": "Le Big Mac",
				"bool":   true,
			},
		)
	})
	t.Run("struct with unknown fields", func(t *testing.T) {
		type A struct {
			Int    int
			String string
			Bool   bool
		}

		DoTest(t,
			A{
				Int:    42,
				String: "Le Big Mac",
				Bool:   true,
			},
			map[string]any{
				"int":    42,
				"string": "Le Big Mac",
				"bool":   true,

				"foo": "bar",
				"baz": nil,
			},
		)
	})
	t.Run("struct with tag", func(t *testing.T) {
		type A struct {
			Int    int    `json:"int_"`
			String string `json:"string_"`
			Bool   bool   `json:"bool_"`
		}

		DoTest(t,
			A{
				Int:    42,
				String: "Le Big Mac",
				Bool:   true,
			},
			map[string]any{
				"int_":    42,
				"string_": "Le Big Mac",
				"bool_":   true,
			},
		)
	})
	t.Run("nested struct", func(t *testing.T) {
		type A struct {
			Int    int
			String string
			Bool   bool
		}
		type B struct {
			Int    int
			String string
			Bool   bool

			A1 A
			A2 A
		}

		DoTest(t,
			B{
				Int:    42,
				String: "Le Big Mac",
				Bool:   true,

				A1: A{
					Int:    1,
					String: "first",
					Bool:   true,
				},
				A2: A{
					Int:    2,
					String: "second",
					Bool:   false,
				},
			},
			map[string]any{
				"int":    42,
				"string": "Le Big Mac",
				"bool":   true,
				"a1": map[string]any{
					"int":    1,
					"string": "first",
					"bool":   true,
				},
				"a2": map[string]any{
					"int":    2,
					"string": "second",
					"bool":   false,
				},
			},
		)
	})
}

func DoTestV[T any](t *testing.T, v T) {
	DoTest(t, v, v)
}

func DoTest[T any, U any](t *testing.T, v T, j U) {
	var actual T
	expected := v

	err := jz.Unmarshal(js.ValueOf(j), &actual)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}
