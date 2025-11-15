//go:build js

package jz

import (
	"context"
	"syscall/js"
)

// Promise creates a JavaScript Promise from a Go function.
// The function f should return a result value and an error.
// If the error is not nil, the Promise will be rejected.
//
// Note: Unlike a JavaScript `new Promise((resolve, reject) => { ... })` executor,
// the function you pass is not run synchronously during construction.
// It is deferred and executed in a separate goroutine after the current call returns.
// Therefore, any side effects or panics inside the function will not be observable
// in the immediate calling context.
func Promise(f func() (any, any)) js.Value {
	return globalScope.Promise(f)
}

// Await waits for a JavaScript Promise to resolve and returns its result.
// If the Promise is rejected, an error is returned.
func Await(p js.Value) (js.Value, error) {
	return globalScope.Await(p)
}

// AwaitContext waits for a JavaScript Promise to resolve with context support.
// The wait can be cancelled via the provided context.
func AwaitContext(ctx context.Context, p js.Value) (js.Value, error) {
	return globalScope.AwaitContext(ctx, p)
}

// Resolve creates a Promise that resolves to the given value.
func Resolve(v js.Value) js.Value {
	return js.Global().Get("Promise").Call("resolve", v)
}

// Reject creates a Promise that rejects with the given value.
func Reject(v js.Value) js.Value {
	return js.Global().Get("Promise").Call("reject", v)
}
