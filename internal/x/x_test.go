package x

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

type myError struct{}

func (*myError) Error() string { return "typed error" }

var (
	errSentinel = errors.New("sentinel")
	errWrapped  = fmt.Errorf("wrap: %w", errSentinel)
	errTyped    = fmt.Errorf("outer: %w", &myError{})
)

func TestObjectsAreEqual(t *testing.T) {
	cases := []struct {
		name     string
		a, b     any
		expected bool
	}{
		{"equal ints", 42, 42, true},
		{"unequal ints", 42, 43, false},
		{"different int types", int8(42), int16(42), false},
		{"equal strings", "foo", "foo", true},
		{"unequal strings", "foo", "bar", false},
		{"equal byte slices", []byte("hi"), []byte("hi"), true},
		{"nil vs empty byte slice", []byte(nil), []byte{}, true},
		{"byte vs non-byte", []byte("hi"), "hi", false},
		{"both untyped nil", nil, nil, true},
		{"nil vs value", nil, 1, false},
		{"equal slices", []int{1, 2}, []int{1, 2}, true},
		{"equal maps", map[string]int{"a": 1}, map[string]int{"a": 1}, true},
		{"typed nil chans", (chan int)(nil), (chan int)(nil), true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := objectsAreEqual(c.a, c.b); got != c.expected {
				t.Fatalf("objectsAreEqual(%#v, %#v) = %v, want %v", c.a, c.b, got, c.expected)
			}
		})
	}
}

func TestContainsElement(t *testing.T) {
	cases := []struct {
		name            string
		container, elem any
		expected        bool
	}{
		{"substring present", "alpha beta", "beta", true},
		{"substring absent", "alpha beta", "gamma", false},
		{"map key present", map[string]int{"foo": 1}, "foo", true},
		{"map key absent", map[string]int{"foo": 1}, "bar", false},
		{"slice element present", []int{1, 2, 3}, 2, true},
		{"slice element absent", []int{1, 2, 3}, 9, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := containsElement(c.container, c.elem); got != c.expected {
				t.Fatalf("containsElement(%#v, %#v) = %v, want %v", c.container, c.elem, got, c.expected)
			}
		})
	}
}

func TestIsEmpty(t *testing.T) {
	var nilPtr *int
	cases := []struct {
		name     string
		obj      any
		expected bool
	}{
		{"nil", nil, true},
		{"empty string", "", true},
		{"non-empty string", "x", false},
		{"empty slice", []int{}, true},
		{"non-empty slice", []int{1}, false},
		{"empty map", map[string]int{}, true},
		{"zero int", 0, true},
		{"non-zero int", 1, false},
		{"nil pointer", nilPtr, true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isEmpty(c.obj); got != c.expected {
				t.Fatalf("isEmpty(%#v) = %v, want %v", c.obj, got, c.expected)
			}
		})
	}
}

func TestIsNil(t *testing.T) {
	var nilPtr *int
	var nilSlice []int
	cases := []struct {
		name     string
		obj      any
		expected bool
	}{
		{"nil", nil, true},
		{"nil pointer", nilPtr, true},
		{"nil slice", nilSlice, true},
		{"non-nil value", 42, false},
		{"empty non-nil slice", []int{}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isNil(c.obj); got != c.expected {
				t.Fatalf("isNil(%#v) = %v, want %v", c.obj, got, c.expected)
			}
		})
	}
}

// TestAssertionsPassPath exercises the happy path of each exported assertion:
// none of these should fail the test. If any wrongly reported a failure, this
// test would fail, proving the assertions are not vacuous no-ops.
func TestAssertionsPassPath(t *testing.T) {
	NoError(t, nil)
	Error(t, errSentinel)
	Equal(t, 1, 1)
	True(t, true)
	False(t, false)
	ErrorIs(t, errWrapped, errSentinel)
	var target *myError
	ErrorAs(t, errTyped, &target)
	ErrorContains(t, errSentinel, "sentinel")
	Contains(t, "hello", "ell")
	NotContains(t, "hello", "xyz")
	Len(t, []int{1, 2, 3}, 3)
	Empty(t, []int{})
	NotNil(t, 5)
	Greater(t, 2, 1)
	now := time.Now()
	WithinRange(t, now, now.Add(-time.Second), now.Add(time.Second))
	WithinDuration(t, now, now.Add(10*time.Millisecond), time.Second)
	NotPanics(t, func() {})
}
