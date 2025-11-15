//go:build js

package jz_test

import (
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

	t.Run("read small at once", func(t *testing.T) {
		x := require.New(t)

		data := "Royale with Cheese"
		r, err := jz.NewReader(newReadableStream(t, []byte(data)))
		x.NoError(err)

		buff := make([]byte, len(data)*2)
		n, err := r.Read(buff)
		x.NoError(err)
		x.Equal(len(data), n)
		x.Equal(data, string(buff[:n]))

		_, err = r.Read(buff)
		x.ErrorIs(err, io.EOF)
	})
	t.Run("read small by chunk", func(t *testing.T) {
		x := require.New(t)

		//       |<--->|<---->|<--|+++
		data := "Royale with Cheese"
		r, err := jz.NewReader(newReadableStream(t, []byte(data)))
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

		data := strings.Repeat("abcdefg", (BufSize/7 + 1))[:BufSize]
		r, err := jz.NewReader(newReadableStream(t, []byte(data)))
		x.NoError(err)

		buff := make([]byte, len(data)*2)
		n, err := r.Read(buff)
		x.NoError(err)
		x.Equal(len(data), n)
		x.Equal(data, string(buff[:n]))

		_, err = r.Read(buff)
		x.ErrorIs(err, io.EOF)
	})

	teat_read_by_chunk := func(t *testing.T, data_size, chunk_size int) {
		x := require.New(t)

		data := strings.Repeat("abcdefg", (data_size/7 + 1))[:data_size]
		r, err := jz.NewReader(newReadableStream(t, []byte(data)))
		x.NoError(err)

		pos := 0
		buff := make([]byte, chunk_size)
		for pos < (data_size - chunk_size) {
			n, err := r.Read(buff)
			x.NoError(err)
			x.Equal(min(chunk_size, BufSize), n)
			x.Equal(data[pos:pos+n], string(buff[:n]))
			pos += n
		}

		n, err := r.Read(buff)
		x.NoError(err)
		x.Equal(data_size-pos, n)
		x.Equal(data[pos:], string(buff[:n]))

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
		r, err := jz.NewReader(newReadableStream(t, []byte(data)))
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
		r, err := jz.NewReader(newReadableStream(t, []byte(data)))
		x.NoError(err)

		buff, err := io.ReadAll(r)
		x.NoError(err)
		x.Equal(data, string(buff))
	})
	t.Run("close", func(t *testing.T) {
		x := require.New(t)

		r, err := jz.NewReader(newHangStream(t))
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

func newReadableStream(t *testing.T, data []byte) js.Value {
	t.Helper()

	pull := js.FuncOf(func(this js.Value, args []js.Value) any {
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
		if l >= m {
			view.Call("set", jz.BytesToJs(data))
			req.Call("respond", m)
			ctrl.Call("close")
		} else {
			view.Call("set", jz.BytesToJs(data[:l]))
			req.Call("respond", l)
			data = data[l:]
		}
		return js.Undefined()
	})
	t.Cleanup(pull.Release)

	opt := map[string]any{
		"type": "bytes",
		"pull": pull,
	}
	return js.Global().Get("ReadableStream").New(opt)
}

func newHangStream(t *testing.T) js.Value {
	t.Helper()

	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	pull := js.FuncOf(func(this js.Value, args []js.Value) any {
		return jz.Promise(func() (any, any) {
			<-ctx.Done()
			return nil, ctx.Err().Error()
		})
	})
	cancel_read := js.FuncOf(func(this js.Value, args []js.Value) any {
		cancel()
		return js.Undefined()
	})
	t.Cleanup(pull.Release)

	opt := map[string]any{
		"type":   "bytes",
		"pull":   pull,
		"cancel": cancel_read,
	}
	return js.Global().Get("ReadableStream").New(opt)
}
