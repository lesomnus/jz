//go:build js

package jz

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"syscall/js"
)

type streamReader struct {
	s js.Value // ReadableStream
	a js.Value // ArrayBuffer

	buff []byte
	view []byte

	done   bool
	closed atomic.Bool
}

func NewReader(v js.Value) (io.ReadCloser, error) {
	const BuffSize = 1024
	if !v.InstanceOf(js.Global().Get("ReadableStream")) {
		return nil, errors.New("not a ReadableStream")
	}

	b := make([]byte, BuffSize)
	return &streamReader{
		s: v.Call("getReader", map[string]any{"mode": "byob"}),
		a: js.Global().Get("ArrayBuffer").New(BuffSize),

		buff: b,
		view: b[:0],
	}, nil
}

func (r *streamReader) Read(p []byte) (int, error) {
	if r.closed.Load() {
		return 0, io.ErrClosedPipe
	}
	if r.done {
		return 0, io.EOF
	}

	l := len(p)
	if l == 0 {
		return 0, nil
	}

	m := 0 // copied
	if len(r.view) > 0 {
		m = copy(p, r.view)
		r.view = r.view[m:]
		if l == m {
			return m, nil
		}

		p = p[m:]
	}

	res, err := Await(r.s.Call("read", js.Global().Get("Uint8Array").New(r.a)))
	if err != nil {
		if r.closed.Load() {
			return m, io.ErrClosedPipe
		}
		return m, err
	}
	if res.Get("done").Truthy() {
		r.done = true
		if m > 0 {
			return m, nil
		}
		if r.closed.Load() {
			return 0, io.ErrClosedPipe
		} else {
			return 0, io.EOF
		}
	}

	data := res.Get("value")
	r.a = data.Get("buffer")

	n := js.CopyBytesToGo(p, data)
	data = data.Call("subarray", n)

	size := js.CopyBytesToGo(r.buff, data)
	r.view = r.buff[:size]

	return m + n, nil
}

func (r *streamReader) Close() error {
	if r.closed.Swap(true) {
		return nil
	}

	// cancel() is best-effort cleanup: it may reject if the stream is already
	// errored or closed, which is not a meaningful failure for the caller.
	Await(r.s.Call("cancel", "close is called"))
	return nil
}

func NewReadableStream(r io.ReadCloser) js.Value {
	buff := make([]byte, 1024)
	view := buff[:0]

	closed := false
	pulling := false
	var pendingErr error

	var _pull, _cancel js.Func
	var releaseOnce sync.Once
	release := func() {
		releaseOnce.Do(func() {
			_pull.Release()
			_cancel.Release()
		})
	}

	_pull = js.FuncOf(func(this js.Value, args []js.Value) any {
		ctrl := args[0]
		req := ctrl.Get("byobRequest")

		if len(view) == 0 && pendingErr == nil {
			// r.Read may park this goroutine; _cancel can run meanwhile and
			// defers releasing the funcs to us (see the pulling guard).
			pulling = true
			var n int
			var err error
			for {
				// Skip spurious (0, nil) reads: responding 0 / enqueuing an
				// empty chunk both throw, so keep reading until there is data
				// or an error (io.Reader discourages repeated empty reads).
				n, err = r.Read(buff)
				if closed || n > 0 || err != nil {
					break
				}
			}
			pulling = false
			if closed {
				release()
				return js.Undefined()
			}
			view = buff[:n]
			pendingErr = err
		}

		// Terminate only once every buffered byte has been delivered, so that
		// bytes returned alongside io.EOF (a legal Read result) are not lost.
		if len(view) == 0 && pendingErr != nil {
			if errors.Is(pendingErr, io.EOF) {
				ctrl.Call("close")
				if req.Truthy() {
					// Valid (and required) only while the byob request is live,
					// i.e. after close but not after error.
					req.Call("respond", 0)
				}
			} else {
				ctrl.Call("error", NewError(pendingErr.Error()))
			}
			release()
			return js.Undefined()
		}

		if !req.Truthy() {
			ctrl.Call("enqueue", BytesToJs(view))
			view = view[:0]
			return js.Undefined()
		}

		dst := req.Get("view")
		l := dst.Length()
		m := len(view)
		if l < m {
			dst.Call("set", BytesToJs(view[:l]))
			req.Call("respond", l)
			view = view[l:]
		} else {
			dst.Call("set", BytesToJs(view))
			req.Call("respond", m)
			view = view[:0]
		}
		return js.Undefined()
	})
	_cancel = js.FuncOf(func(this js.Value, args []js.Value) any {
		closed = true
		r.Close()
		if !pulling {
			// If a pull is in flight (parked in r.Read), it releases the funcs
			// itself when it resumes; releasing here would drop a callback that
			// is still executing.
			release()
		}
		return js.Undefined()
	})

	opt := map[string]any{
		"type":   "bytes",
		"pull":   _pull,
		"cancel": _cancel,
	}
	return js.Global().Get("ReadableStream").New(opt)
}
