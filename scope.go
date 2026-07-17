//go:build js

package jz

import (
	"context"
	"fmt"
	"sync"
	"syscall/js"
)

var globalScope = Scope{}

// GlobalScope returns the global Scope instance that can be used to manage
// JavaScript function lifecycles and wait for their completion.
func GlobalScope() *Scope {
	return &globalScope
}

type Scope struct {
	wg sync.WaitGroup
}

func (s *Scope) Wait() {
	s.wg.Wait()
}

func (s *Scope) FuncOf(f func(this js.Value, args []js.Value) any) js.Func {
	return js.FuncOf(func(this js.Value, args []js.Value) any {
		s.wg.Add(1)
		defer s.wg.Done()

		return f(this, args)
	})
}

func (s *Scope) Promise(f func() (any, any)) js.Value {
	// The executor is invoked exactly once, synchronously, by the Promise
	// constructor; release it as soon as it returns so it is not retained in
	// syscall/js's callback table forever.
	var executor js.Func
	executor = js.FuncOf(func(this js.Value, args []js.Value) any {
		defer executor.Release()

		resolve := args[0]
		reject := args[1]

		s.wg.Go(func() {
			// A panic in the user function must not abort the whole wasm
			// module; surface it as a rejection instead.
			defer func() {
				if r := recover(); r != nil {
					reject.Invoke(toJS(fmt.Errorf("promise: panic: %v", r)))
				}
			}()

			v, err := f()
			if err != nil {
				reject.Invoke(toJS(err))
			} else {
				resolve.Invoke(toJS(v))
			}
		})

		return nil
	})

	return js.Global().Get("Promise").New(executor)
}

func (s *Scope) Await(p js.Value) (js.Value, error) {
	return s.AwaitContext(context.Background(), p)
}

func (s *Scope) AwaitContext(ctx context.Context, p js.Value) (js.Value, error) {
	s.wg.Add(1)
	defer s.wg.Done()

	type outcome struct {
		value    js.Value
		rejected bool
	}
	c := make(chan outcome, 1)

	// The two handlers are released as soon as either one fires. A settled
	// Promise invokes exactly one of them exactly once, so releasing both
	// (including the sibling that will never fire) is safe and frees the
	// callbacks even when the caller has already given up via ctx.
	var onResolve, onReject js.Func
	release := func() {
		onResolve.Release()
		onReject.Release()
	}
	arg := func(args []js.Value) js.Value {
		if len(args) > 0 {
			return args[0]
		}
		return js.Undefined()
	}
	onResolve = js.FuncOf(func(this js.Value, args []js.Value) any {
		c <- outcome{value: arg(args)}
		release()
		return js.Undefined()
	})
	onReject = js.FuncOf(func(this js.Value, args []js.Value) any {
		c <- outcome{value: arg(args), rejected: true}
		release()
		return js.Undefined()
	})
	p.Call("then", onResolve, onReject)

	select {
	case <-ctx.Done():
		return js.Undefined(), ctx.Err()

	case o := <-c:
		if o.rejected {
			return js.Undefined(), RejectedError{o.value}
		}
		return o.value, nil
	}
}

// toJS converts a value produced by a Promise function into a js.Value suitable
// for resolve/reject. A Go error becomes a JS Error and a []byte becomes a
// Uint8Array; anything syscall/js cannot convert falls back to its string form
// so that the conversion never panics (which on js/wasm would abort the whole
// module).
func toJS(v any) (out js.Value) {
	switch x := v.(type) {
	case nil:
		return js.Undefined()
	case js.Value:
		return x
	case js.Func:
		return x.Value
	case error:
		return NewError(x.Error())
	case []byte:
		return BytesToJs(x)
	}

	defer func() {
		if recover() != nil {
			out = NewError(fmt.Sprint(v))
		}
	}()
	return js.ValueOf(v)
}
