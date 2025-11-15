//go:build js && wasm

package jz_test

import (
	"context"
	"syscall/js"
	"testing"
	"time"

	"github.com/lesomnus/jz"
	"github.com/stretchr/testify/require"
)

func TestPromiseAwait(t *testing.T) {
	t.Run("resolve", func(t *testing.T) {
		p := jz.Promise(func() (any, any) {
			return "foo", nil
		})
		v, err := jz.Await(p)
		require.NoError(t, err)
		require.Equal(t, "foo", v.String())
	})
	t.Run("reject", func(t *testing.T) {
		p := jz.Promise(func() (any, any) {
			return nil, "foo"
		})
		_, err := jz.Await(p)
		rejected := jz.RejectedError{}
		require.ErrorAs(t, err, &rejected)
		require.Equal(t, "foo", rejected.Value.String())
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
	require.WithinRange(t, t1, t0, t0.Add(50*time.Millisecond))
	require.ErrorIs(t, err, context.DeadlineExceeded)
}

func TestResolve(t *testing.T) {
	p := jz.Resolve(js.ValueOf(42))
	v, err := jz.Await(p)
	require.NoError(t, err)
	require.Equal(t, js.TypeNumber, v.Type())
	require.Equal(t, 42, v.Int())
}

func TestReject(t *testing.T) {
	p := jz.Reject(js.ValueOf(42))
	_, err := jz.Await(p)
	rejected := jz.RejectedError{}
	require.ErrorAs(t, err, &rejected)
	require.Equal(t, js.TypeNumber, rejected.Value.Type())
	require.Equal(t, 42, rejected.Value.Int())
}
