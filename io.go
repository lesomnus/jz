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
