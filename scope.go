//go:build js

package jz

import (
	"sync"
	"syscall/js"
)

var globalScope = Scope{}

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
	return js.Global().Get("Promise").New(s.FuncOf(func(this js.Value, args []js.Value) any {
		resolve := args[0]
		reject := args[1]

		v, err := f()
		if err != nil {
			reject.Invoke(err)
		} else {
			resolve.Invoke(v)
		}

		return nil
	}))
}

func (s *Scope) Await(p js.Value) (js.Value, error) {
	s.wg.Add(1)
	defer s.wg.Done()

	c := make(chan js.Value)
	caught := false
	p.Call("then",
		js.FuncOf(func(this js.Value, args []js.Value) any {
			c <- args[0]
			return js.Undefined()
		}),
		js.FuncOf(func(this js.Value, args []js.Value) any {
			caught = true
			c <- args[0]
			return js.Undefined()
		}),
	)

	v := <-c
	if caught {
		return js.Undefined(), RejectedError{v}
	} else {
		return v, nil
	}
}
