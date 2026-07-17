// Package x provides the small set of test assertions used by this module.
// It exists to keep the module dependency-free (replacing testify). Every
// assertion is fatal: on mismatch it reports the failure and stops the test
// via t.FailNow, matching testify's require semantics.
package x

import (
	"bytes"
	"cmp"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"
)

func report(t testing.TB, msgAndArgs []any, format string, args ...any) {
	t.Helper()
	msg := fmt.Sprintf(format, args...)
	if extra := formatMsg(msgAndArgs); extra != "" {
		msg += ": " + extra
	}
	t.Fatal(msg)
}

func formatMsg(msgAndArgs []any) string {
	switch len(msgAndArgs) {
	case 0:
		return ""
	case 1:
		if s, ok := msgAndArgs[0].(string); ok {
			return s
		}
		return fmt.Sprint(msgAndArgs[0])
	default:
		if s, ok := msgAndArgs[0].(string); ok && strings.Contains(s, "%") {
			return fmt.Sprintf(s, msgAndArgs[1:]...)
		}
		return fmt.Sprint(msgAndArgs...)
	}
}

// NoError asserts that err is nil.
func NoError(t testing.TB, err error, msgAndArgs ...any) {
	t.Helper()
	if err != nil {
		report(t, msgAndArgs, "expected no error, got: %v", err)
	}
}

// Error asserts that err is non-nil.
func Error(t testing.TB, err error, msgAndArgs ...any) {
	t.Helper()
	if err == nil {
		report(t, msgAndArgs, "expected an error, got nil")
	}
}

// Equal asserts that expected and actual are deeply equal.
func Equal(t testing.TB, expected, actual any, msgAndArgs ...any) {
	t.Helper()
	if !objectsAreEqual(expected, actual) {
		report(t, msgAndArgs, "not equal:\n  expected: %#v\n  actual:   %#v", expected, actual)
	}
}

// True asserts that value is true.
func True(t testing.TB, value bool, msgAndArgs ...any) {
	t.Helper()
	if !value {
		report(t, msgAndArgs, "expected true, got false")
	}
}

// False asserts that value is false.
func False(t testing.TB, value bool, msgAndArgs ...any) {
	t.Helper()
	if value {
		report(t, msgAndArgs, "expected false, got true")
	}
}

// ErrorIs asserts that errors.Is(err, target) holds.
func ErrorIs(t testing.TB, err, target error, msgAndArgs ...any) {
	t.Helper()
	if !errors.Is(err, target) {
		report(t, msgAndArgs, "error is not the target:\n  err:    %v\n  target: %v", err, target)
	}
}

// ErrorAs asserts that errors.As(err, target) holds.
func ErrorAs(t testing.TB, err error, target any, msgAndArgs ...any) {
	t.Helper()
	if !errors.As(err, target) {
		report(t, msgAndArgs, "error does not match target type: %v", err)
	}
}

// ErrorContains asserts that err is non-nil and its message contains contains.
func ErrorContains(t testing.TB, err error, contains string, msgAndArgs ...any) {
	t.Helper()
	if err == nil {
		report(t, msgAndArgs, "expected an error containing %q, got nil", contains)
	}
	if !strings.Contains(err.Error(), contains) {
		report(t, msgAndArgs, "error %q does not contain %q", err.Error(), contains)
	}
}

// Contains asserts that container contains element. container may be a string
// (substring), a map (key), or a slice/array (element).
func Contains(t testing.TB, container, element any, msgAndArgs ...any) {
	t.Helper()
	if !containsElement(container, element) {
		report(t, msgAndArgs, "%#v does not contain %#v", container, element)
	}
}

// NotContains asserts that container does not contain element.
func NotContains(t testing.TB, container, element any, msgAndArgs ...any) {
	t.Helper()
	if containsElement(container, element) {
		report(t, msgAndArgs, "%#v should not contain %#v", container, element)
	}
}

// Len asserts that object has the given length.
func Len(t testing.TB, object any, length int, msgAndArgs ...any) {
	t.Helper()
	v := reflect.ValueOf(object)
	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		if v.Len() != length {
			report(t, msgAndArgs, "expected length %d, got %d", length, v.Len())
		}
	default:
		report(t, msgAndArgs, "value of type %T has no length", object)
	}
}

// Empty asserts that object is empty (nil, zero value, or an empty
// string/slice/map/chan).
func Empty(t testing.TB, object any, msgAndArgs ...any) {
	t.Helper()
	if !isEmpty(object) {
		report(t, msgAndArgs, "expected empty, got %#v", object)
	}
}

// NotNil asserts that object is not nil.
func NotNil(t testing.TB, object any, msgAndArgs ...any) {
	t.Helper()
	if isNil(object) {
		report(t, msgAndArgs, "expected non-nil value")
	}
}

// Greater asserts that a > b.
func Greater[T cmp.Ordered](t testing.TB, a, b T, msgAndArgs ...any) {
	t.Helper()
	if !(a > b) {
		report(t, msgAndArgs, "expected %v > %v", a, b)
	}
}

// WithinRange asserts that actual lies within [start, end], inclusive.
func WithinRange(t testing.TB, actual, start, end time.Time, msgAndArgs ...any) {
	t.Helper()
	if actual.Before(start) || actual.After(end) {
		report(t, msgAndArgs, "time %v is not within [%v, %v]", actual, start, end)
	}
}

// WithinDuration asserts that expected and actual are within delta of each other.
func WithinDuration(t testing.TB, expected, actual time.Time, delta time.Duration, msgAndArgs ...any) {
	t.Helper()
	diff := expected.Sub(actual)
	if diff < 0 {
		diff = -diff
	}
	if diff > delta {
		report(t, msgAndArgs, "time difference %v exceeds %v", diff, delta)
	}
}

// NotPanics asserts that f does not panic.
func NotPanics(t testing.TB, f func(), msgAndArgs ...any) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			report(t, msgAndArgs, "unexpected panic: %v", r)
		}
	}()
	f()
}

func objectsAreEqual(expected, actual any) bool {
	if expected == nil || actual == nil {
		return expected == actual
	}
	if exp, ok := expected.([]byte); ok {
		act, ok := actual.([]byte)
		if !ok {
			return false
		}
		return bytes.Equal(exp, act)
	}
	return reflect.DeepEqual(expected, actual)
}

func containsElement(container, element any) bool {
	cv := reflect.ValueOf(container)
	switch cv.Kind() {
	case reflect.String:
		return strings.Contains(cv.String(), fmt.Sprintf("%v", element))
	case reflect.Map:
		for _, k := range cv.MapKeys() {
			if objectsAreEqual(k.Interface(), element) {
				return true
			}
		}
	case reflect.Slice, reflect.Array:
		for i := range cv.Len() {
			if objectsAreEqual(cv.Index(i).Interface(), element) {
				return true
			}
		}
	}
	return false
}

func isEmpty(object any) bool {
	if object == nil {
		return true
	}
	v := reflect.ValueOf(object)
	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Pointer:
		if v.IsNil() {
			return true
		}
		return isEmpty(v.Elem().Interface())
	default:
		return reflect.DeepEqual(object, reflect.Zero(v.Type()).Interface())
	}
}

func isNil(object any) bool {
	if object == nil {
		return true
	}
	v := reflect.ValueOf(object)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	}
	return false
}
