//go:build js

package jz_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"strings"
	"syscall/js"
	"testing"
	"time"

	"github.com/lesomnus/jz"
	"github.com/lesomnus/jz/internal/x"
)

func TestNewReader(t *testing.T) {
	const BufSize = 1024

	new_readable_stream := func(t *testing.T, data []byte) js.Value {
		t.Helper()

		_pull := js.FuncOf(func(this js.Value, args []js.Value) any {
			ctrl := args[0]
			req := ctrl.Get("byobRequest")
			if !req.Truthy() {
				ctrl.Call("enqueue", jz.BytesToJs(data))
				ctrl.Call("close")
				return js.Undefined()
			}

			view := req.Get("view")
			l := view.Length()
			m := len(data)
			if l < m {
				view.Call("set", jz.BytesToJs(data[:l]))
				req.Call("respond", l)
				data = data[l:]
			} else {
				view.Call("set", jz.BytesToJs(data))
				req.Call("respond", m)
				ctrl.Call("close")
			}
			return js.Undefined()
		})
		t.Cleanup(_pull.Release)

		opt := map[string]any{
			"type": "bytes",
			"pull": _pull,
		}
		return js.Global().Get("ReadableStream").New(opt)
	}

	new_hang_stream := func(t *testing.T) js.Value {
		t.Helper()

		ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
		_pull := js.FuncOf(func(this js.Value, args []js.Value) any {
			return jz.Promise(func() (any, any) {
				<-ctx.Done()
				return nil, ctx.Err().Error()
			})
		})
		_cancel := js.FuncOf(func(this js.Value, args []js.Value) any {
			cancel()
			return js.Undefined()
		})
		t.Cleanup(_pull.Release)

		opt := map[string]any{
			"type":   "bytes",
			"pull":   _pull,
			"cancel": _cancel,
		}
		return js.Global().Get("ReadableStream").New(opt)
	}

	t.Run("read small at once", func(t *testing.T) {

		data := []byte("Royale with Cheese")
		r, err := jz.NewReader(new_readable_stream(t, data))
		x.NoError(t, err)

		buff := make([]byte, len(data)*2)
		n, err := r.Read(buff)
		x.NoError(t, err)
		x.Equal(t, len(data), n)
		x.Equal(t, data, buff[:n])

		_, err = r.Read(buff)
		x.ErrorIs(t, err, io.EOF)
	})
	t.Run("read small by chunk", func(t *testing.T) {

		//              |<---->|<---->|<--+++
		data := []byte("Royale with Cheese")
		r, err := jz.NewReader(new_readable_stream(t, data))
		x.NoError(t, err)

		buff := make([]byte, 7)
		n, err := r.Read(buff)
		x.NoError(t, err)
		x.Equal(t, 7, n)
		x.Equal(t, "Royale ", string(buff))

		n, err = r.Read(buff)
		x.NoError(t, err)
		x.Equal(t, 7, n)
		x.Equal(t, "with Ch", string(buff))

		n, err = r.Read(buff)
		x.NoError(t, err)
		x.Equal(t, 4, n)
		x.Equal(t, "eese", string(buff[:n]))

		_, err = r.Read(buff)
		x.ErrorIs(t, err, io.EOF)
	})
	t.Run("read fit at once", func(t *testing.T) {

		data := []byte(strings.Repeat("abcdefg", (BufSize/7 + 1))[:BufSize])
		r, err := jz.NewReader(new_readable_stream(t, data))
		x.NoError(t, err)

		buff := make([]byte, len(data)*2)
		n, err := r.Read(buff)
		x.NoError(t, err)
		x.Equal(t, len(data), n)
		x.Equal(t, data, buff[:n])

		_, err = r.Read(buff)
		x.ErrorIs(t, err, io.EOF)
	})

	teat_read_by_chunk := func(t *testing.T, data_size, chunk_size int) {

		data := []byte(strings.Repeat("abcdefg", (data_size/7 + 1))[:data_size])
		r, err := jz.NewReader(new_readable_stream(t, data))
		x.NoError(t, err)

		pos := 0
		buff := make([]byte, chunk_size)
		for pos < (data_size - chunk_size) {
			n, err := r.Read(buff)
			x.NoError(t, err)
			x.Equal(t, min(chunk_size, BufSize), n)
			x.Equal(t, data[pos:pos+n], buff[:n])
			pos += n
		}

		n, err := r.Read(buff)
		x.NoError(t, err)
		x.Equal(t, data_size-pos, n)
		x.Equal(t, data[pos:], buff[:n])

		_, err = r.Read(buff)
		x.ErrorIs(t, err, io.EOF)
	}
	t.Run("read fit by chunk", func(t *testing.T) {
		teat_read_by_chunk(t, BufSize, 13)
	})
	t.Run("read large by chunk with sub-size of buff", func(t *testing.T) {
		teat_read_by_chunk(t, 5000, BufSize-23)
	})
	t.Run("read large by chunk with exact size of buff", func(t *testing.T) {
		teat_read_by_chunk(t, 5000, BufSize)
	})
	t.Run("read large by chunk with sup-size of buff", func(t *testing.T) {
		teat_read_by_chunk(t, 5000, BufSize+23)
	})

	t.Run("read by io.ReadFull", func(t *testing.T) {

		const data_size = 5000
		const chunk_size = BufSize + 23

		data := strings.Repeat("abcdefg", (data_size/7 + 1))[:data_size]
		r, err := jz.NewReader(new_readable_stream(t, []byte(data)))
		x.NoError(t, err)

		pos := 0
		buff := make([]byte, chunk_size)
		for pos < (data_size - chunk_size) {
			n, err := io.ReadFull(r, buff)
			x.NoError(t, err)
			x.Equal(t, chunk_size, n)
			x.Equal(t, data[pos:pos+n], string(buff[:n]))
			pos += n
		}

		n, err := io.ReadFull(r, buff)
		x.ErrorIs(t, err, io.ErrUnexpectedEOF)
		x.Equal(t, data_size-pos, n)
		x.Equal(t, data[pos:], string(buff[:n]))

		_, err = r.Read(buff)
		x.ErrorIs(t, err, io.EOF)
	})
	t.Run("read by io.ReadAll", func(t *testing.T) {

		const data_size = 5000

		data := strings.Repeat("abcdefg", (data_size/7 + 1))[:data_size]
		r, err := jz.NewReader(new_readable_stream(t, []byte(data)))
		x.NoError(t, err)

		buff, err := io.ReadAll(r)
		x.NoError(t, err)
		x.Equal(t, data, string(buff))
	})
	t.Run("close", func(t *testing.T) {

		r, err := jz.NewReader(new_hang_stream(t))
		x.NoError(t, err)

		t0 := time.Now()
		go func() {
			time.Sleep(30 * time.Millisecond)
			r.Close()
		}()

		buff := make([]byte, 42)
		_, err = r.Read(buff)
		x.ErrorIs(t, err, io.ErrClosedPipe)

		t1 := time.Now()
		x.WithinRange(t, t1, t0, t0.Add(50*time.Millisecond))
	})
}

func TestNewReadableStream(t *testing.T) {
	const BufSize = 1024

	get_reader := func(data []byte) js.Value {
		s := jz.NewReadableStream(io.NopCloser(bytes.NewReader(data)))
		return s.Call("getReader", map[string]any{"mode": "byob"})
	}
	read := func(t *testing.T, r js.Value, buff js.Value) (js.Value, js.Value, bool, error) {
		res, err := jz.AwaitContext(t.Context(), r.Call("read", js.Global().Get("Uint8Array").New(buff)))
		x.NoError(t, err)

		v := res.Get("value")
		return v.Get("buffer"), v, res.Get("done").Bool(), nil
	}

	t.Run("read small at once", func(t *testing.T) {

		data := []byte("Royale with Cheese")
		r := get_reader(data)

		buff := js.Global().Get("ArrayBuffer").New(len(data) * 2)
		buff, v, done, err := read(t, r, buff)
		x.NoError(t, err)
		x.False(t, done)
		x.Equal(t, []byte(data), jz.BytesToGo(v))

		_, _, done, err = read(t, r, buff)
		x.NoError(t, err)
		x.True(t, done)
	})
	t.Run("read small by chunk", func(t *testing.T) {

		//              |<---->|<---->|<--+++
		data := []byte("Royale with Cheese")
		r := get_reader(data)

		buff := js.Global().Get("ArrayBuffer").New(7)
		buff, v, done, err := read(t, r, buff)
		x.NoError(t, err)
		x.False(t, done)
		x.Equal(t, 7, v.Length())
		x.Equal(t, "Royale ", string(jz.BytesToGo(v)))

		buff, v, done, err = read(t, r, buff)
		x.NoError(t, err)
		x.False(t, done)
		x.Equal(t, 7, v.Length())
		x.Equal(t, "with Ch", string(jz.BytesToGo(v)))

		buff, v, done, err = read(t, r, buff)
		x.NoError(t, err)
		x.False(t, done)
		x.Equal(t, 4, v.Length())
		x.Equal(t, "eese", string(jz.BytesToGo(v)))

		_, _, done, err = read(t, r, buff)
		x.NoError(t, err)
		x.True(t, done)
	})
	t.Run("read fit at once", func(t *testing.T) {

		data := []byte(strings.Repeat("abcdefg", (BufSize/7 + 1))[:BufSize])
		r := get_reader(data)

		buff := js.Global().Get("ArrayBuffer").New(BufSize * 2)
		buff, v, done, err := read(t, r, buff)
		x.NoError(t, err)
		x.False(t, done)
		x.Equal(t, len(data), v.Length())
		x.Equal(t, data, jz.BytesToGo(v))

		_, _, done, err = read(t, r, buff)
		x.NoError(t, err)
		x.True(t, done)
	})

	teat_read_by_chunk := func(t *testing.T, data_size, chunk_size int) {

		data := []byte(strings.Repeat("abcdefg", (data_size/7 + 1))[:data_size])
		r := get_reader(data)

		pos := 0
		remain := BufSize
		buff := js.Global().Get("ArrayBuffer").New(chunk_size)
		for pos < (data_size - chunk_size) {
			expected_read_bytes := min(chunk_size, remain)

			buff_, v, done, err := read(t, r, buff)
			n := v.Length()
			x.NoError(t, err)
			x.False(t, done)
			x.Equal(t, expected_read_bytes, n)
			x.Equal(t, data[pos:pos+n], jz.BytesToGo(v))

			pos += n
			buff = buff_

			remain -= expected_read_bytes
			if remain <= 0 {
				remain = BufSize
			}
		}
		if chunk_size < BufSize {
			expected_read_bytes := min(data_size-pos, BufSize-chunk_size)

			buff_, v, done, err := read(t, r, buff)
			n := v.Length()
			x.NoError(t, err)
			x.False(t, done)
			x.Equal(t, expected_read_bytes, n)
			x.Equal(t, data[pos:pos+expected_read_bytes], jz.BytesToGo(v))

			pos += n
			buff = buff_
		}
		if pos < data_size {
			buff_, v, done, err := read(t, r, buff)
			n := v.Length()
			x.NoError(t, err)
			x.False(t, done)
			x.Equal(t, data_size-pos, n)
			x.Equal(t, data[pos:], jz.BytesToGo(v))

			pos += n
			buff = buff_
		}

		_, _, done, err := read(t, r, buff)
		x.NoError(t, err)
		x.True(t, done)
	}
	t.Run("read fit by chunk", func(t *testing.T) {
		teat_read_by_chunk(t, BufSize, 13)
	})
	t.Run("read large by chunk with sub-size of buff", func(t *testing.T) {
		teat_read_by_chunk(t, 5000, BufSize-23)
	})
	t.Run("read large by chunk with exact size of buff", func(t *testing.T) {
		teat_read_by_chunk(t, 5000, BufSize)
	})
	t.Run("read large by chunk with sup-size of buff", func(t *testing.T) {
		teat_read_by_chunk(t, 5000, BufSize+23)
	})
	t.Run("cancel", func(t *testing.T) {

		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()

		s := jz.NewReadableStream(hangReader{ctx, cancel})
		r := s.Call("getReader", map[string]any{"mode": "byob"})

		t0 := time.Now()
		go func() {
			time.Sleep(30 * time.Millisecond)
			r.Call("cancel")
		}()

		buff := js.Global().Get("ArrayBuffer").New(BufSize)
		res, err := jz.AwaitContext(t.Context(), r.Call("read", js.Global().Get("Uint8Array").New(buff)))
		x.NoError(t, err)
		x.True(t, res.Get("done").Bool())

		t1 := time.Now()
		x.WithinRange(t, t1, t0, t0.Add(50*time.Millisecond))
	})
}

type hangReader struct {
	ctx    context.Context
	cancel context.CancelFunc
}

func (r hangReader) Read(p []byte) (int, error) {
	<-r.ctx.Done()
	return 0, r.ctx.Err()
}

func (r hangReader) Close() error {
	r.cancel()
	return nil
}

// readCloser adapts a Read func into an io.ReadCloser.
type funcReadCloser struct {
	read  func(p []byte) (int, error)
	close func() error
}

func (r funcReadCloser) Read(p []byte) (int, error) { return r.read(p) }
func (r funcReadCloser) Close() error {
	if r.close != nil {
		return r.close()
	}
	return nil
}

// jsStream wraps data into a JS ReadableStream via the package under test.
func jsStream(data []byte) js.Value {
	return jz.NewReadableStream(io.NopCloser(bytes.NewReader(data)))
}

func TestStreamReaderClose(t *testing.T) {
	t.Run("idempotent", func(t *testing.T) {
		r, err := jz.NewReader(jsStream([]byte("abc")))
		x.NoError(t, err)
		x.NoError(t, r.Close())
		x.NoError(t, r.Close())
	})
	t.Run("read after close", func(t *testing.T) {
		r, err := jz.NewReader(jsStream([]byte("abc")))
		x.NoError(t, err)
		x.NoError(t, r.Close())

		_, err = r.Read(make([]byte, 4))
		x.ErrorIs(t, err, io.ErrClosedPipe)
	})
	t.Run("zero-length read", func(t *testing.T) {
		r, err := jz.NewReader(jsStream([]byte("abc")))
		x.NoError(t, err)
		defer r.Close()

		n, err := r.Read(nil)
		x.NoError(t, err)
		x.Equal(t, 0, n)
	})
	t.Run("double read at EOF", func(t *testing.T) {
		r, err := jz.NewReader(jsStream([]byte("abc")))
		x.NoError(t, err)
		defer r.Close()

		got, err := io.ReadAll(r)
		x.NoError(t, err)
		x.Equal(t, "abc", string(got))

		_, err = r.Read(make([]byte, 4))
		x.ErrorIs(t, err, io.EOF)
		_, err = r.Read(make([]byte, 4))
		x.ErrorIs(t, err, io.EOF)
	})
}

func TestNewReadableStreamBytesWithEOF(t *testing.T) {
	// A reader whose final Read returns (n>0, io.EOF) in one call must not lose
	// its trailing bytes.

	data := []byte("Royale with Cheese tail")
	done := false
	src := funcReadCloser{read: func(p []byte) (int, error) {
		if done {
			return 0, io.EOF
		}
		done = true
		n := copy(p, data)
		return n, io.EOF
	}}

	r, err := jz.NewReader(jz.NewReadableStream(src))
	x.NoError(t, err)
	defer r.Close()

	got, err := io.ReadAll(r)
	x.NoError(t, err)
	x.Equal(t, data, got)
}

func TestNewReadableStreamSpuriousEmptyRead(t *testing.T) {
	// A (0, nil) read must not enqueue an empty chunk / respond(0) (both throw);
	// the stream should just pull again and still deliver the data.

	data := []byte("payload")
	calls := 0
	src := funcReadCloser{read: func(p []byte) (int, error) {
		calls++
		switch calls {
		case 1:
			return 0, nil
		case 2:
			return copy(p, data), io.EOF
		default:
			return 0, io.EOF
		}
	}}

	r, err := jz.NewReader(jz.NewReadableStream(src))
	x.NoError(t, err)
	defer r.Close()

	got, err := io.ReadAll(r)
	x.NoError(t, err)
	x.Equal(t, data, got)
}

func TestNewReadableStreamErrorPropagation(t *testing.T) {
	// A non-EOF error must surface to the consumer as a stream error, after the
	// bytes read before it, and must not panic (the old code called respond(0)
	// on an invalidated byob request).

	calls := 0
	src := funcReadCloser{read: func(p []byte) (int, error) {
		calls++
		if calls == 1 {
			return copy(p, []byte("data")), nil
		}
		return 0, errors.New("boom")
	}}

	r, err := jz.NewReader(jz.NewReadableStream(src))
	x.NoError(t, err)
	defer r.Close()

	got, err := io.ReadAll(r)
	x.Error(t, err)
	x.Contains(t, err.Error(), "boom")
	x.Equal(t, "data", string(got))
}

func TestNewReadableStreamDefaultReader(t *testing.T) {
	// The default (non-byob) reader path uses enqueue; previously untested and
	// prone to throwing on the terminal respond(0).

	data := []byte("default reader path")
	s := jsStream(data)
	reader := s.Call("getReader") // default reader, no byob mode

	var got []byte
	for {
		res, err := jz.AwaitContext(t.Context(), reader.Call("read"))
		x.NoError(t, err)
		if res.Get("done").Bool() {
			break
		}
		got = append(got, jz.BytesToGo(res.Get("value"))...)
	}
	x.Equal(t, data, got)
}
