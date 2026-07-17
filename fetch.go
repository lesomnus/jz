//go:build js

package jz

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"syscall/js"
)

type FetchTransport struct{}

func (t FetchTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// The RoundTripper contract requires the request body to always be closed,
	// including on error paths.
	if req.Body != nil {
		defer req.Body.Close()
	}

	if req.URL == nil {
		return nil, errors.New("nil URL")
	}

	ctx := req.Context()
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	opts := map[string]any{
		"method": req.Method,
	}

	if len(req.Header) > 0 {
		// Use a Headers object with append so that headers carrying multiple
		// values are all forwarded (a plain object would keep only one).
		h := js.Global().Get("Headers").New()
		for k, vs := range req.Header {
			for _, v := range vs {
				h.Call("append", k, v)
			}
		}
		opts["headers"] = h
	}

	// GET/HEAD requests must not carry a body; http.NoBody means "no body".
	if req.Body != nil && req.Body != http.NoBody &&
		req.Method != http.MethodGet && req.Method != http.MethodHead {
		// Seems that browser does not support stream upload yet.
		// opts["body"] = NewReadableStream(req.Body)
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("read body: %w", err)
		}
		if len(b) > 0 {
			opts["body"] = BytesToJs(b)
		}
	}

	ac := js.Global().Get("AbortController").New()
	// Abort the fetch when the request context is cancelled. The hook is kept
	// alive until the response body is closed so that cancellation also
	// interrupts body reads that happen after RoundTrip returns; ownership of
	// stop() is transferred to the body (or run here when there is no body).
	stop := context.AfterFunc(ctx, func() {
		ac.Call("abort")
	})
	opts["signal"] = ac.Get("signal")

	handedOff := false
	defer func() {
		if !handedOff {
			stop()
		}
	}()

	js_res, err := Await(js.Global().Call("fetch", req.URL.String(), opts))
	if err != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("fetch: %w", err)
	}

	code := js_res.Get("status").Int()
	res := &http.Response{
		Request: req,

		Status:     strconv.Itoa(code) + " " + http.StatusText(code),
		StatusCode: code,
		Proto:      "HTTP/1.1",
		ProtoMajor: 1,
		ProtoMinor: 1,
		Header:     http.Header{},
	}

	// Parse response headers.
	{
		it := js_res.Get("headers").Call("entries")
		for {
			next := it.Call("next")
			if next.Get("done").Bool() {
				break
			}
			entry := next.Get("value")
			key := entry.Index(0).String()
			val := entry.Index(1).String()
			res.Header.Add(key, val)
		}
		// The Headers iterator collapses multiple Set-Cookie values into a
		// single comma-joined entry; recover them individually when the
		// runtime exposes getSetCookie().
		if hdr := js_res.Get("headers"); hdr.Get("getSetCookie").Truthy() {
			res.Header.Del("Set-Cookie")
			cookies := hdr.Call("getSetCookie")
			for i := range cookies.Length() {
				res.Header.Add("Set-Cookie", cookies.Index(i).String())
			}
		}
	}

	// Content length is known only from the header; otherwise it is unknown.
	res.ContentLength = -1
	if cl := res.Header.Get("Content-Length"); cl != "" {
		if n, err := strconv.ParseInt(cl, 10, 64); err == nil {
			res.ContentLength = n
		}
	}

	js_body := js_res.Get("body")
	switch {
	case js_body.IsNull():
		// There is no body.
		res.Body = http.NoBody
		res.ContentLength = 0

	case js_body.InstanceOf(js.Global().Get("ReadableStream")):
		body, err := NewReader(js_body)
		if err == nil {
			res.Body = &fetchBody{ReadCloser: body, stop: stop}
			handedOff = true
			break
		}

		// Fallback to `bytes`.
		fallthrough
	default:
		data, err := AwaitContext(ctx, js_res.Call("bytes"))
		if err != nil {
			return nil, fmt.Errorf("res.bytes: %w", err)
		}

		body := BytesToGo(data)
		res.Body = io.NopCloser(bytes.NewReader(body))
		res.ContentLength = int64(len(body))
	}

	return res, nil
}

// fetchBody wraps a streaming response body so that closing it also releases
// the context-cancellation hook created in RoundTrip.
type fetchBody struct {
	io.ReadCloser
	stop func() bool
	done bool
}

func (b *fetchBody) Close() error {
	if b.done {
		return nil
	}
	b.done = true

	b.stop()
	return b.ReadCloser.Close()
}
