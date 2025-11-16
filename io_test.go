//go:build js

package jz_test

import (
	"bytes"
	"context"
	"io"
	"strings"
	"syscall/js"
	"testing"
	"time"

	"github.com/lesomnus/jz"
	"github.com/stretchr/testify/require"
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
		x := require.New(t)

		data := []byte("Royale with Cheese")
		r, err := jz.NewReader(new_readable_stream(t, data))
		x.NoError(err)

		buff := make([]byte, len(data)*2)
		n, err := r.Read(buff)
		x.NoError(err)
		x.Equal(len(data), n)
		x.Equal(data, buff[:n])

		_, err = r.Read(buff)
		x.ErrorIs(err, io.EOF)
	})
	t.Run("read small by chunk", func(t *testing.T) {
		x := require.New(t)

		//              |<---->|<---->|<--+++
		data := []byte("Royale with Cheese")
		r, err := jz.NewReader(new_readable_stream(t, data))
		x.NoError(err)

		buff := make([]byte, 7)
		n, err := r.Read(buff)
		x.NoError(err)
		x.Equal(7, n)
		x.Equal("Royale ", string(buff))

		n, err = r.Read(buff)
		x.NoError(err)
		x.Equal(7, n)
		x.Equal("with Ch", string(buff))

		n, err = r.Read(buff)
		x.NoError(err)
		x.Equal(4, n)
		x.Equal("eese", string(buff[:n]))

		_, err = r.Read(buff)
		x.ErrorIs(err, io.EOF)
	})
	t.Run("read fit at once", func(t *testing.T) {
		x := require.New(t)

		data := []byte(strings.Repeat("abcdefg", (BufSize/7 + 1))[:BufSize])
		r, err := jz.NewReader(new_readable_stream(t, data))
		x.NoError(err)

		buff := make([]byte, len(data)*2)
		n, err := r.Read(buff)
		x.NoError(err)
		x.Equal(len(data), n)
		x.Equal(data, buff[:n])

		_, err = r.Read(buff)
		x.ErrorIs(err, io.EOF)
	})

	teat_read_by_chunk := func(t *testing.T, data_size, chunk_size int) {
		x := require.New(t)

		data := []byte(strings.Repeat("abcdefg", (data_size/7 + 1))[:data_size])
		r, err := jz.NewReader(new_readable_stream(t, data))
		x.NoError(err)

		pos := 0
		buff := make([]byte, chunk_size)
		for pos < (data_size - chunk_size) {
			n, err := r.Read(buff)
			x.NoError(err)
			x.Equal(min(chunk_size, BufSize), n)
			x.Equal(data[pos:pos+n], buff[:n])
			pos += n
		}

		n, err := r.Read(buff)
		x.NoError(err)
		x.Equal(data_size-pos, n)
		x.Equal(data[pos:], buff[:n])

		_, err = r.Read(buff)
		x.ErrorIs(err, io.EOF)
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
		x := require.New(t)

		const data_size = 5000
		const chunk_size = BufSize + 23

		data := strings.Repeat("abcdefg", (data_size/7 + 1))[:data_size]
		r, err := jz.NewReader(new_readable_stream(t, []byte(data)))
		x.NoError(err)

		pos := 0
		buff := make([]byte, chunk_size)
		for pos < (data_size - chunk_size) {
			n, err := io.ReadFull(r, buff)
			x.NoError(err)
			x.Equal(chunk_size, n)
			x.Equal(data[pos:pos+n], string(buff[:n]))
			pos += n
		}

		n, err := io.ReadFull(r, buff)
		x.ErrorIs(err, io.ErrUnexpectedEOF)
		x.Equal(data_size-pos, n)
		x.Equal(data[pos:], string(buff[:n]))

		_, err = r.Read(buff)
		x.ErrorIs(err, io.EOF)
	})
	t.Run("read by io.ReadAll", func(t *testing.T) {
		x := require.New(t)

		const data_size = 5000

		data := strings.Repeat("abcdefg", (data_size/7 + 1))[:data_size]
		r, err := jz.NewReader(new_readable_stream(t, []byte(data)))
		x.NoError(err)

		buff, err := io.ReadAll(r)
		x.NoError(err)
		x.Equal(data, string(buff))
	})
	t.Run("close", func(t *testing.T) {
		x := require.New(t)

		r, err := jz.NewReader(new_hang_stream(t))
		x.NoError(err)

		t0 := time.Now()
		go func() {
			time.Sleep(30 * time.Millisecond)
			r.Close()
		}()

		buff := make([]byte, 42)
		_, err = r.Read(buff)
		x.ErrorIs(err, io.ErrClosedPipe)

		t1 := time.Now()
		x.WithinRange(t1, t0, t0.Add(50*time.Millisecond))
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
		require.NoError(t, err)

		v := res.Get("value")
		return v.Get("buffer"), v, res.Get("done").Bool(), nil
	}

	t.Run("read small at once", func(t *testing.T) {
		x := require.New(t)

		data := []byte("Royale with Cheese")
		r := get_reader(data)

		buff := js.Global().Get("ArrayBuffer").New(len(data) * 2)
		buff, v, done, err := read(t, r, buff)
		x.NoError(err)
		x.False(done)
		x.Equal([]byte(data), jz.BytesToGo(v))

		_, _, done, err = read(t, r, buff)
		x.NoError(err)
		x.True(done)
	})
	t.Run("read small by chunk", func(t *testing.T) {
		x := require.New(t)

		//              |<---->|<---->|<--+++
		data := []byte("Royale with Cheese")
		r := get_reader(data)

		buff := js.Global().Get("ArrayBuffer").New(7)
		buff, v, done, err := read(t, r, buff)
		x.NoError(err)
		x.False(done)
		x.Equal(7, v.Length())
		x.Equal("Royale ", string(jz.BytesToGo(v)))

		buff, v, done, err = read(t, r, buff)
		x.NoError(err)
		x.False(done)
		x.Equal(7, v.Length())
		x.Equal("with Ch", string(jz.BytesToGo(v)))

		buff, v, done, err = read(t, r, buff)
		x.NoError(err)
		x.False(done)
		x.Equal(4, v.Length())
		x.Equal("eese", string(jz.BytesToGo(v)))

		_, _, done, err = read(t, r, buff)
		x.NoError(err)
		x.True(done)
	})
	t.Run("read fit at once", func(t *testing.T) {
		x := require.New(t)

		data := []byte(strings.Repeat("abcdefg", (BufSize/7 + 1))[:BufSize])
		r := get_reader(data)

		buff := js.Global().Get("ArrayBuffer").New(BufSize * 2)
		buff, v, done, err := read(t, r, buff)
		x.NoError(err)
		x.False(done)
		x.Equal(len(data), v.Length())
		x.Equal(data, jz.BytesToGo(v))

		_, _, done, err = read(t, r, buff)
		x.NoError(err)
		x.True(done)
	})

	teat_read_by_chunk := func(t *testing.T, data_size, chunk_size int) {
		x := require.New(t)

		data := []byte(strings.Repeat("abcdefg", (data_size/7 + 1))[:data_size])
		r := get_reader(data)

		pos := 0
		remain := BufSize
		buff := js.Global().Get("ArrayBuffer").New(chunk_size)
		for pos < (data_size - chunk_size) {
			expected_read_bytes := min(chunk_size, remain)

			buff_, v, done, err := read(t, r, buff)
			n := v.Length()
			x.NoError(err)
			x.False(done)
			x.Equal(expected_read_bytes, n)
			x.Equal(data[pos:pos+n], jz.BytesToGo(v))

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
			x.NoError(err)
			x.False(done)
			x.Equal(expected_read_bytes, n)
			x.Equal(data[pos:pos+expected_read_bytes], jz.BytesToGo(v))

			pos += n
			buff = buff_
		}
		if pos < data_size {
			buff_, v, done, err := read(t, r, buff)
			n := v.Length()
			x.NoError(err)
			x.False(done)
			x.Equal(data_size-pos, n)
			x.Equal(data[pos:], jz.BytesToGo(v))

			pos += n
			buff = buff_
		}

		_, _, done, err := read(t, r, buff)
		x.NoError(err)
		x.True(done)
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
		x := require.New(t)

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
		x.NoError(err)
		x.True(res.Get("done").Bool())

		t1 := time.Now()
		x.WithinRange(t1, t0, t0.Add(50*time.Millisecond))
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
