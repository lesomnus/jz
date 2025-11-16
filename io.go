//go:build js

package jz

import (
	"errors"
	"io"
	"syscall/js"
)

type streamReader struct {
	s js.Value // ReadableStream
	a js.Value // ArrayBuffer

	buff []byte
	view []byte

	done   bool
	closed bool // Atomic?
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
	if r.closed {
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
		return 0, err
	}
	if res.Get("done").Truthy() {
		r.done = true
		if m > 0 {
			return m, nil
		}
		if r.closed {
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
	if r.closed {
		return nil
	}

	r.closed = true
	if _, err := Await(r.s.Call("cancel", "close is called")); err != nil {
		return err
	}

	return nil
}

func NewReadableStream(r io.ReadCloser) js.Value {
	buff := make([]byte, 1024)
	view := buff[:0]

	closed := false

	_pull := js.FuncOf(func(this js.Value, args []js.Value) any {
		ctrl := args[0]
		req := ctrl.Get("byobRequest")
		if len(view) == 0 {
			n, err := r.Read(buff)
			if closed {
				return js.Undefined()
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					ctrl.Call("close")
				} else {
					ctrl.Call("error", NewError(err.Error()))
				}
				req.Call("respond", 0)
				return js.Undefined()
			}

			view = buff[:n]
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
	_cancel := js.FuncOf(func(this js.Value, args []js.Value) any {
		closed = true
		r.Close()
		return js.Undefined()
	})

	opt := map[string]any{
		"type":   "bytes",
		"pull":   _pull,
		"cancel": _cancel,
	}
	return js.Global().Get("ReadableStream").New(opt)
}
