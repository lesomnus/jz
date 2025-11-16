//go:build js

package jz

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"syscall/js"
)

type FetchTransport struct{}

func (t FetchTransport) RoundTrip(req *http.Request) (*http.Response, error) {
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
		h := map[string]any{}
		for k, vs := range req.Header {
			if len(vs) > 0 {
				h[k] = vs[0]
			}
		}
		opts["headers"] = h
	}
	if req.Body != nil {
		// Seems that browser does not support stream upload yet.
		// opts["body"] = NewReadableStream(req.Body)
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, fmt.Errorf("read body: %w", err)
		}
		opts["body"] = BytesToJs(b)
	}

	ac := js.Global().Get("AbortController").New()
	stop := context.AfterFunc(ctx, func() {
		ac.Call("abort")
	})
	defer stop()
	opts["signal"] = ac.Get("signal")

	js_res, err := Await(js.Global().Call("fetch", req.URL.String(), opts))
	if err != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		return nil, fmt.Errorf("fetch: %w", err)
	}

	res := &http.Response{
		Request: req,

		Status:     http.StatusText(js_res.Get("status").Int()),
		StatusCode: js_res.Get("status").Int(),
	}

	js_body := js_res.Get("body")
	switch {
	case js_body.IsNull():
		// There is no body.
	case js_body.InstanceOf(js.Global().Get("ReadableStream")):
		res.Body, err = NewReader(js_body)
		if err == nil {
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

	// Prase header.
	{
		h := http.Header{}
		it := js_res.Get("headers").Call("entries")
		for {
			next := it.Call("next")
			if next.Get("done").Bool() {
				break
			}
			entry := next.Get("value")
			key := entry.Index(0).String()
			val := entry.Index(1).String()
			h.Add(key, val)
		}
		res.Header = h
	}

	return res, nil
}
