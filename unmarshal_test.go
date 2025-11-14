//go:build js

package jz_test

import (
	"fmt"
	"reflect"
	"syscall/js"
	"testing"

	"github.com/lesomnus/jz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUnmarshal(t *testing.T) {
	t.Run("type match boolean", func(t *testing.T) {
		DoTestType[bool, bool](t, true)

		DoTestType[bool, int](t, false)
		DoTestType[bool, int8](t, false)
		DoTestType[bool, int16](t, false)
		DoTestType[bool, int32](t, false)
		DoTestType[bool, int64](t, false)

		DoTestType[bool, uint](t, false)
		DoTestType[bool, uint8](t, false)
		DoTestType[bool, uint16](t, false)
		DoTestType[bool, uint32](t, false)
		DoTestType[bool, uint64](t, false)
		DoTestType[bool, uintptr](t, false)

		DoTestType[bool, float32](t, false)
		DoTestType[bool, float64](t, false)
		DoTestType[bool, complex64](t, false)
		DoTestType[bool, complex128](t, false)

		DoTestType[bool, [3]int](t, false)
		DoTestType[bool, chan int](t, false)
		DoTestType[bool, func()](t, false)
		DoTestType[bool, any](t, true)
		DoTestType[bool, map[string]any](t, false)
		DoTestType[bool, []int](t, false)
		DoTestType[bool, string](t, false)
		DoTestType[bool, struct{}](t, false)

		DoTestType[bool, byte](t, false)
		DoTestType[bool, rune](t, false)
	})
	t.Run("type match number", func(t *testing.T) {
		DoTestType[int, bool](t, false)

		DoTestType[int, int](t, true)
		DoTestType[int, int8](t, true)
		DoTestType[int, int16](t, true)
		DoTestType[int, int32](t, true)
		DoTestType[int, int64](t, true)

		DoTestType[int, uint](t, true)
		DoTestType[int, uint8](t, true)
		DoTestType[int, uint16](t, true)
		DoTestType[int, uint32](t, true)
		DoTestType[int, uint64](t, true)
		DoTestType[int, uintptr](t, true)

		DoTestType[int, float32](t, true)
		DoTestType[int, float64](t, true)
		DoTestType[int, complex64](t, true)
		DoTestType[int, complex128](t, true)

		DoTestType[int, [3]int](t, false)
		DoTestType[int, chan int](t, false)
		DoTestType[int, func()](t, false)
		DoTestType[int, any](t, true)
		DoTestType[int, map[string]any](t, false)
		DoTestType[int, []any](t, false)
		DoTestType[int, string](t, false)
		DoTestType[int, struct{}](t, false)

		DoTestType[int, byte](t, true)
		DoTestType[int, rune](t, true)
	})
	t.Run("type match string", func(t *testing.T) {
		DoTestType[string, bool](t, false)

		DoTestType[string, int](t, false)
		DoTestType[string, int8](t, false)
		DoTestType[string, int16](t, false)
		DoTestType[string, int32](t, false)
		DoTestType[string, int64](t, false)

		DoTestType[string, uint](t, false)
		DoTestType[string, uint8](t, false)
		DoTestType[string, uint16](t, false)
		DoTestType[string, uint32](t, false)
		DoTestType[string, uint64](t, false)
		DoTestType[string, uintptr](t, false)

		DoTestType[string, float32](t, false)
		DoTestType[string, float64](t, false)
		DoTestType[string, complex64](t, false)
		DoTestType[string, complex128](t, false)

		DoTestType[string, [3]int](t, false)
		DoTestType[string, chan int](t, false)
		DoTestType[string, func()](t, false)
		DoTestType[string, any](t, true)
		DoTestType[string, map[string]any](t, false)
		DoTestType[string, []any](t, false)
		DoTestType[string, string](t, true)
		DoTestType[string, struct{}](t, false)

		DoTestType[string, byte](t, false)
		DoTestType[string, rune](t, false)
	})
	t.Run("type match null", func(t *testing.T) {
		DoTestType[any, bool](t, true)

		DoTestType[any, int](t, true)
		DoTestType[any, int8](t, true)
		DoTestType[any, int16](t, true)
		DoTestType[any, int32](t, true)
		DoTestType[any, int64](t, true)

		DoTestType[any, uint](t, true)
		DoTestType[any, uint8](t, true)
		DoTestType[any, uint16](t, true)
		DoTestType[any, uint32](t, true)
		DoTestType[any, uint64](t, true)
		DoTestType[any, uintptr](t, true)

		DoTestType[any, float32](t, true)
		DoTestType[any, float64](t, true)
		DoTestType[any, complex64](t, true)
		DoTestType[any, complex128](t, true)

		DoTestType[any, [3]int](t, true)
		DoTestType[any, chan int](t, true)
		DoTestType[any, func()](t, true)
		DoTestType[any, any](t, true)
		DoTestType[any, map[string]any](t, true)
		DoTestType[any, []any](t, true)
		DoTestType[any, string](t, true)
		DoTestType[any, struct{}](t, true)

		DoTestType[any, byte](t, true)
		DoTestType[any, rune](t, true)
	})
	t.Run("type match Array", func(t *testing.T) {
		DoTestType[[]any, bool](t, false)

		DoTestType[[]any, int](t, false)
		DoTestType[[]any, int8](t, false)
		DoTestType[[]any, int16](t, false)
		DoTestType[[]any, int32](t, false)
		DoTestType[[]any, int64](t, false)

		DoTestType[[]any, uint](t, false)
		DoTestType[[]any, uint8](t, false)
		DoTestType[[]any, uint16](t, false)
		DoTestType[[]any, uint32](t, false)
		DoTestType[[]any, uint64](t, false)
		DoTestType[[]any, uintptr](t, false)

		DoTestType[[]any, float32](t, false)
		DoTestType[[]any, float64](t, false)
		DoTestType[[]any, complex64](t, false)
		DoTestType[[]any, complex128](t, false)

		DoTestType[[]any, [3]int](t, true)
		DoTestType[[]any, chan int](t, false)
		DoTestType[[]any, func()](t, false)
		DoTestType[[]any, any](t, true)
		DoTestType[[]any, map[string]any](t, true) // Note that Array is an Object.
		DoTestType[[]any, []any](t, true)
		DoTestType[[]any, string](t, false)
		DoTestType[[]any, struct{}](t, true)

		DoTestType[[]any, byte](t, false)
		DoTestType[[]any, rune](t, false)
	})
	t.Run("type match Object", func(t *testing.T) {
		DoTestType[map[string]any, bool](t, false)

		DoTestType[map[string]any, int](t, false)
		DoTestType[map[string]any, int8](t, false)
		DoTestType[map[string]any, int16](t, false)
		DoTestType[map[string]any, int32](t, false)
		DoTestType[map[string]any, int64](t, false)

		DoTestType[map[string]any, uint](t, false)
		DoTestType[map[string]any, uint8](t, false)
		DoTestType[map[string]any, uint16](t, false)
		DoTestType[map[string]any, uint32](t, false)
		DoTestType[map[string]any, uint64](t, false)
		DoTestType[map[string]any, uintptr](t, false)

		DoTestType[map[string]any, float32](t, false)
		DoTestType[map[string]any, float64](t, false)
		DoTestType[map[string]any, complex64](t, false)
		DoTestType[map[string]any, complex128](t, false)

		DoTestType[map[string]any, [3]int](t, false)
		DoTestType[map[string]any, chan int](t, false)
		DoTestType[map[string]any, func()](t, false)
		DoTestType[map[string]any, any](t, true)
		DoTestType[map[string]any, map[string]any](t, true)
		DoTestType[map[string]any, []any](t, false)
		DoTestType[map[string]any, string](t, false)
		DoTestType[map[string]any, struct{}](t, true)

		DoTestType[map[string]any, byte](t, false)
		DoTestType[map[string]any, rune](t, false)
	})
	t.Run("bool", func(t *testing.T) {
		DoTestV[any](t, false)
		DoTestV[any](t, true)

		DoTestV(t, false)
		DoTestV(t, true)
	})
	t.Run("number", func(t *testing.T) {
		DoTestV[any](t, float64(3.14))

		DoTestV(t, int(42))
		DoTestV(t, int8(42))
		DoTestV(t, int16(42))
		DoTestV(t, int32(42))
		DoTestV(t, int64(42))

		DoTestV(t, uint(42))
		DoTestV(t, uint8(42))
		DoTestV(t, uint16(42))
		DoTestV(t, uint32(42))
		DoTestV(t, uint64(42))

		DoTestV(t, float32(3.14))
		DoTestV(t, float64(3.14))
		DoTestJ(t, complex64(3.14), js.ValueOf(3.14))
		DoTestJ(t, complex128(3.14), js.ValueOf(3.14))
	})
	t.Run("Array -> array", func(t *testing.T) {
		DoTest(t, [5]int{1, 2, 3, 4, 5}, [5]int{}, js.ValueOf([]any{1, 2, 3, 4, 5}))
		DoTest(t, [5]int{1, 2, 3, 4, 5}, [5]int{6, 7, 8}, js.ValueOf([]any{1, 2, 3, 4, 5}))
		DoTest(t, [5]int{1, 2, 3, 4, 5}, [5]int{6, 7, 8, 9, 0}, js.ValueOf([]any{1, 2, 3, 4, 5}))
		DoTest(t, [5]int{0, 0, 0, 0, 0}, [5]int{}, js.ValueOf([]any{}))
		DoTest(t, [5]int{6, 7, 8, 0, 0}, [5]int{6, 7, 8}, js.ValueOf([]any{}))
		DoTest(t, [5]int{6, 7, 8, 9, 0}, [5]int{6, 7, 8, 9, 0}, js.ValueOf([]any{}))
		DoTest(t, [5]int{1, 2, 3, 9, 0}, [5]int{6, 7, 8, 9, 0}, js.ValueOf([]any{1, 2, 3}))
		DoTest(t, [5]int{1, 2, 3, 4, 5}, [5]int{6, 7, 8, 9, 0}, js.ValueOf([]any{1, 2, 3, 4, 5}))
		DoTest(t, [5]int{1, 2, 3, 4, 5}, [5]int{6, 7, 8, 9, 0}, js.ValueOf([]any{1, 2, 3, 4, 5, 6, 7}))
	})
	t.Run("Array -> slice", func(t *testing.T) {
		DoTest(t, []int{1, 2, 3, 4, 5}, []int{}, js.ValueOf([]any{1, 2, 3, 4, 5}))
		DoTest(t, []int{1, 2, 3, 4, 5}, []int{6, 7, 8}, js.ValueOf([]any{1, 2, 3, 4, 5}))
		DoTest(t, []int{1, 2, 3, 4, 5}, []int{6, 7, 8, 9, 0}, js.ValueOf([]any{1, 2, 3, 4, 5}))
		DoTest(t, []int{}, []int{}, js.ValueOf([]any{}))
		DoTest(t, []int{}, []int{6, 7, 8}, js.ValueOf([]any{}))
		DoTest(t, []int{}, []int{6, 7, 8, 9, 0}, js.ValueOf([]any{}))
		DoTest(t, []int{1, 2, 3}, []int{6, 7, 8, 9, 0}, js.ValueOf([]any{1, 2, 3}))
		DoTest(t, []int{1, 2, 3, 4, 5}, []int{6, 7, 8, 9, 0}, js.ValueOf([]any{1, 2, 3, 4, 5}))
		DoTest(t, []int{1, 2, 3, 4, 5, 6, 7}, []int{6, 7, 8, 9, 0}, js.ValueOf([]any{1, 2, 3, 4, 5, 6, 7}))
	})
	t.Run("Array of numbers -> []number", func(t *testing.T) {
		jv := js.ValueOf([]any{1, float32(2), 3.14, 4, 5})

		DoTestJ(t, []int{1, 2, 3, 4, 5}, jv)
		DoTestJ(t, []int8{1, 2, 3, 4, 5}, jv)
		DoTestJ(t, []int16{1, 2, 3, 4, 5}, jv)
		DoTestJ(t, []int32{1, 2, 3, 4, 5}, jv)
		DoTestJ(t, []int64{1, 2, 3, 4, 5}, jv)

		DoTestJ(t, []uint{1, 2, 3, 4, 5}, jv)
		DoTestJ(t, []uint8{1, 2, 3, 4, 5}, jv)
		DoTestJ(t, []uint16{1, 2, 3, 4, 5}, jv)
		DoTestJ(t, []uint32{1, 2, 3, 4, 5}, jv)
		DoTestJ(t, []uint64{1, 2, 3, 4, 5}, jv)

		DoTestJ(t, []float32{1, 2, 3.14, 4, 5}, jv)
		DoTestJ(t, []float64{1, 2, 3.14, 4, 5}, jv)
		DoTestJ(t, []complex64{1, 2, 3.14, 4, 5}, jv)
		DoTestJ(t, []complex128{1, 2, 3.14, 4, 5}, jv)
	})
	t.Run("Array of mixed types -> []any", func(t *testing.T) {
		DoTestV[any](t, []any{true, false, float64(2), 3.14, "foo"})
		DoTestV(t, []any{true, false, float64(2), 3.14, "foo"})
		DoTestV(t, []any{[]any{true, false}, []any{float64(2), 3.14, []any{"foo", float64(42)}}, "bar"})
	})
	t.Run("Array of mixed types -> []type", func(t *testing.T) {
		v := []int{}
		err := jz.Unmarshal(js.ValueOf([]any{42, "foo"}), &v)
		require.ErrorContains(t, err, ".[1]")
		require.ErrorContains(t, err, fmt.Sprintf("from %q to %q", "string", "int"))
	})
	t.Run("Uint8Array -> slice", func(t *testing.T) {
		jv := js.Global().Get("Uint8Array").Call("from", []any{1, 2, 3, 4, 5})

		DoTestJ(t, []byte{1, 2, 3, 4, 5}, jv)

		DoTestJ(t, []int{1, 2, 3, 4, 5}, jv)
		DoTestJ(t, []int8{1, 2, 3, 4, 5}, jv)
		DoTestJ(t, []int16{1, 2, 3, 4, 5}, jv)
		DoTestJ(t, []int32{1, 2, 3, 4, 5}, jv)
		DoTestJ(t, []int64{1, 2, 3, 4, 5}, jv)

		DoTestJ(t, []uint{1, 2, 3, 4, 5}, jv)
		DoTestJ(t, []uint8{1, 2, 3, 4, 5}, jv)
		DoTestJ(t, []uint16{1, 2, 3, 4, 5}, jv)
		DoTestJ(t, []uint32{1, 2, 3, 4, 5}, jv)
		DoTestJ(t, []uint64{1, 2, 3, 4, 5}, jv)
	})
	t.Run("Object -> map[string]any", func(t *testing.T) {
		var v any = map[string]any{
			"foo": "bar",
			"baz": float64(42),
		}
		DoTestV(t, v)
	})
	t.Run("Object -> map[string]any with existing fields", func(t *testing.T) {
		DoTest(t,
			map[string]any{
				"foo": "bar",
				"baz": float64(42),
			},
			map[string]any{
				"foo": "bar",
			},
			map[string]any{
				"baz": float64(42),
			},
		)
	})
	t.Run("Object -> struct", func(t *testing.T) {
		type A struct {
			Bool   bool
			Int    int
			String string
		}

		DoTestJ(t,
			A{
				Bool:   true,
				Int:    42,
				String: "Le Big Mac",
			},
			map[string]any{
				"bool":   true,
				"int":    42,
				"string": "Le Big Mac",
			},
		)
	})
	t.Run("Object -> struct with unknown field", func(t *testing.T) {
		type A struct {
			Unknown string
		}

		// Unknown fields are preserved.
		v := A{"foo"}
		DoTest(t, v, v, map[string]any{"bar": 42})
	})
	t.Run("Object -> struct with rename", func(t *testing.T) {
		type A struct {
			Rename string `json:"Abc"`
		}

		DoTestJ(t, A{"foo"}, map[string]any{"Abc": "foo"})
	})
	t.Run("Object -> struct with skip", func(t *testing.T) {
		type A struct {
			Skip string `json:"-"`
		}

		DoTestJ(t, A{}, map[string]any{"skip": "foo"})
	})
	t.Run("string", func(t *testing.T) {
		DoTestV(t, "Royale with Cheese")
	})
	t.Run("undefined noop", func(t *testing.T) {
		j := js.Undefined()

		DoTest(t, false, false, j)
		DoTest(t, true, true, j)

		DoTest(t, int(42), int(42), j)
		DoTest(t, int8(42), int8(42), j)
		DoTest(t, int16(42), int16(42), j)
		DoTest(t, int32(42), int32(42), j)
		DoTest(t, int64(42), int64(42), j)

		DoTest(t, uint(42), uint(42), j)
		DoTest(t, uint8(42), uint8(42), j)
		DoTest(t, uint16(42), uint16(42), j)
		DoTest(t, uint32(42), uint32(42), j)
		DoTest(t, uint64(42), uint64(42), j)
		DoTest(t, uintptr(42), uintptr(42), j)

		DoTest(t, float32(3.14), float32(3.14), j)
		DoTest(t, float64(3.14), float64(3.14), j)
		DoTest(t, complex64(3.14i), complex64(3.14i), j)
		DoTest(t, complex128(3.14i), complex128(3.14i), j)

		DoTest(t, [3]int{1, 2, 3}, [3]int{1, 2, 3}, j)
		{
			v := make(chan int)
			DoTest(t, v, v, j)
		}
		{
			v := map[string]any{"foo": "bar", "baz": 42}
			DoTest(t, v, v, j)
		}
		{
			v := []any{true, 42, "foo"}
			DoTest(t, v, v, j)
		}
		{
			v := struct{ A int }{42}
			DoTest(t, v, v, j)
		}
	})
	t.Run("null -> [zero]", func(t *testing.T) {
		j := js.Null()

		DoTest(t, false, false, j)
		DoTest(t, false, true, j)

		DoTest(t, int(0), int(42), j)
		DoTest(t, int8(0), int8(42), j)
		DoTest(t, int16(0), int16(42), j)
		DoTest(t, int32(0), int32(42), j)
		DoTest(t, int64(0), int64(42), j)

		DoTest(t, uint(0), uint(42), j)
		DoTest(t, uint8(0), uint8(42), j)
		DoTest(t, uint16(0), uint16(42), j)
		DoTest(t, uint32(0), uint32(42), j)
		DoTest(t, uint64(0), uint64(42), j)
		DoTest(t, uintptr(0), uintptr(42), j)

		DoTest(t, float32(0), float32(3.14), j)
		DoTest(t, float64(0), float64(3.14), j)
		DoTest(t, complex64(0), complex64(3.14i), j)
		DoTest(t, complex128(0), complex128(3.14i), j)

		DoTest(t, [3]int{0, 0, 0}, [3]int{1, 2, 3}, j)
		DoTest(t, nil, make(chan int), j)
		DoTest(t, nil, map[string]any{"foo": "bar", "baz": 42}, j)
		DoTest(t, nil, []any{true, 42, "foo"}, j)
		DoTest(t, struct{ A int }{0}, struct{ A int }{42}, j)
	})
	t.Run("nil pointer", func(t *testing.T) {
		var v *int
		err := jz.Unmarshal(js.ValueOf(42), v)
		require.ErrorContains(t, err, "non-nil pointer")
	})
	t.Run("non-pointer", func(t *testing.T) {
		var v int
		err := jz.Unmarshal(js.ValueOf(42), v)
		require.ErrorContains(t, err, "non-nil pointer")
	})
	t.Run("fail if string -> int", func(t *testing.T) {
		var v int
		err := jz.Unmarshal(js.ValueOf("foo"), &v)
		require.ErrorContains(t, err, `"string" to "int"`)
	})
}

func DoTestV[T any](t *testing.T, v T) {
	DoTestJ(t, v, v)
}

func DoTestJ[T any, U any](t *testing.T, v T, j U) {
	var z T
	DoTest(t, v, z, j)
}

func DoTest[T any, U any](t *testing.T, expected T, actual T, j U) {
	err := jz.Unmarshal(js.ValueOf(j), &actual)
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}

func DoTestType[J any, T any](t *testing.T, ok bool) {
	DoTestType_[J, T](t, ok)
	DoTestType_[J, *T](t, ok)
}

func DoTestType_[J any, T any](t *testing.T, ok bool) {
	var a J
	var b T

	j := js.ValueOf(a)
	err := jz.Unmarshal(j, &b)
	if ok {
		assert.NoError(t, err)
	} else {
		dst_t := reflect.TypeOf(b)
		dst := dst_t.Kind().String()
		if dst_t.Kind() == reflect.Pointer {
			dst = dst_t.Elem().Kind().String()
		}

		msg := fmt.Sprintf(" to %q", dst)
		assert.ErrorContains(t, err, msg)
	}
}
