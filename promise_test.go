//go:build js

package jz_test

import (
	"context"
	"errors"
	"syscall/js"
	"testing"
	"time"

	"github.com/lesomnus/jz"
	"github.com/lesomnus/jz/internal/x"
)

func TestPromiseAwait(t *testing.T) {
	t.Run("resolve", func(t *testing.T) {
		p := jz.Promise(func() (any, any) {
			return "foo", nil
		})
		v, err := jz.Await(p)
		x.NoError(t, err)
		x.Equal(t, "foo", v.String())
	})
	t.Run("reject", func(t *testing.T) {
		p := jz.Promise(func() (any, any) {
			return nil, "foo"
		})
		_, err := jz.Await(p)
		rejected := jz.RejectedError{}
		x.ErrorAs(t, err, &rejected)
		x.Equal(t, "foo", rejected.Value.String())
	})
}

func TestPromiseAwaitContext(t *testing.T) {
	p := jz.Promise(func() (any, any) {
		time.Sleep(time.Second)
		return js.Undefined(), nil
	})

	t0 := time.Now()
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Millisecond)
	defer cancel()

	_, err := jz.AwaitContext(ctx, p)
	t1 := time.Now()
	x.WithinRange(t, t1, t0, t0.Add(50*time.Millisecond))
	x.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestResolve(t *testing.T) {
	p := jz.Resolve(js.ValueOf(42))
	v, err := jz.Await(p)
	x.NoError(t, err)
	x.Equal(t, js.TypeNumber, v.Type())
	x.Equal(t, 42, v.Int())
}

func TestReject(t *testing.T) {
	p := jz.Reject(js.ValueOf(42))
	_, err := jz.Await(p)
	rejected := jz.RejectedError{}
	x.ErrorAs(t, err, &rejected)
	x.Equal(t, js.TypeNumber, rejected.Value.Type())
	x.Equal(t, 42, rejected.Value.Int())
}

func TestPromisePanicBecomesRejection(t *testing.T) {

	p := jz.Promise(func() (any, any) {
		panic("kaboom")
	})

	_, err := jz.Await(p)
	x.Error(t, err)
	var rej jz.RejectedError
	x.ErrorAs(t, err, &rej)
	x.Contains(t, rej.Value.Get("message").String(), "kaboom")
}

func TestPromiseGoErrorBecomesJSError(t *testing.T) {

	sentinel := errors.New("db down")
	p := jz.Promise(func() (any, any) {
		return nil, sentinel
	})

	_, err := jz.Await(p)
	x.Error(t, err)
	var rej jz.RejectedError
	x.ErrorAs(t, err, &rej)
	// A Go error must be converted to a JS Error (not crash the module).
	x.Equal(t, js.TypeObject, rej.Value.Type())
	x.Contains(t, rej.Value.Get("message").String(), "db down")
}

func TestPromiseResolvesBytes(t *testing.T) {

	p := jz.Promise(func() (any, any) {
		return []byte("hi"), nil
	})

	v, err := jz.Await(p)
	x.NoError(t, err)
	x.Equal(t, []byte("hi"), jz.BytesToGo(v))
}

func TestAwaitManyDoesNotLeakOrPanic(t *testing.T) {
	// Each Await releases its two handlers; doing many in a row must stay
	// correct and never panic on the self-release.
	for i := range 200 {
		v, err := jz.Await(jz.Resolve(js.ValueOf(i)))
		x.NoError(t, err)
		x.Equal(t, i, v.Int())
	}
}

func TestAwaitContextCancelThenLateSettle(t *testing.T) {
	// The context deadline wins; the promise settles afterwards. The orphaned
	// handler firing later must not panic.

	p := jz.Promise(func() (any, any) {
		time.Sleep(40 * time.Millisecond)
		return js.ValueOf(1), nil
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	_, err := jz.AwaitContext(ctx, p)
	x.ErrorIs(t, err, context.DeadlineExceeded)

	// Give the promise time to settle and fire its now-orphaned callback.
	time.Sleep(60 * time.Millisecond)
}
