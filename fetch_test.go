//go:build js

package jz_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/lesomnus/jz"
	"github.com/lesomnus/jz/internal/x"
)

func TestFetchTransport(t *testing.T) {
	base_url, err := testFetchBaseUrl()
	if err != nil {
		t.Fail()
		t.Skipf("invalid target URL for fetch test: %s", err)
	}
	if base_url == nil {
		t.Fail()
		t.Skip("set JZ_TEST_FETCH_TARGET_URL or JZ_TEST_FETCH_TARGET_HOST/JZ_TEST_FETCH_TARGET_PORT to run")
	}

	t.Run("GET", func(t *testing.T) {

		target := *base_url
		target.Path = "/header"
		req, err := http.NewRequest(http.MethodGet, target.String(), nil)
		x.NoError(t, err)

		req = req.WithContext(t.Context())
		req.Header.Set("X-Foo", "bar")
		res, err := jz.FetchTransport{}.RoundTrip(req)
		x.NoError(t, err)
		x.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		x.NoError(t, err)

		var got struct {
			Method string              `json:"method"`
			URL    string              `json:"url"`
			Proto  string              `json:"proto"`
			Header map[string][]string `json:"header"`
		}
		x.NoError(t, json.Unmarshal(body, &got))
		x.Equal(t, "GET", got.Method)
		x.Equal(t, []string{"bar"}, got.Header["X-Foo"])
	})
	t.Run("POST", func(t *testing.T) {

		target := *base_url
		target.Path = "/echo"
		req, err := http.NewRequest(http.MethodPost, target.String(),
			bytes.NewReader([]byte("Royale with Cheese")),
		)
		x.NoError(t, err)

		req = req.WithContext(t.Context())
		res, err := jz.FetchTransport{}.RoundTrip(req)
		x.NoError(t, err)
		x.Equal(t, http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		x.NoError(t, err)
		x.Equal(t, "Royale with Cheese", string(body))
	})
	t.Run("with context", func(t *testing.T) {

		target := *base_url
		target.Path = "/hang"
		req, err := http.NewRequest(http.MethodGet, target.String(), nil)
		x.NoError(t, err)

		t0 := time.Now()
		ctx, cancel := context.WithTimeout(t.Context(), 30*time.Millisecond)
		defer cancel()

		req = req.WithContext(ctx)
		_, err = jz.FetchTransport{}.RoundTrip(req)
		t1 := time.Now()
		x.WithinRange(t, t1, t0, t0.Add(50*time.Millisecond))
		x.ErrorIs(t, err, context.DeadlineExceeded)
	})
	t.Run("404", func(t *testing.T) {

		target := *base_url
		target.Path = "/not-found"
		req, err := http.NewRequest(http.MethodGet, target.String(), nil)
		x.NoError(t, err)

		req = req.WithContext(t.Context())
		res, err := jz.FetchTransport{}.RoundTrip(req)
		x.NoError(t, err)
		x.Equal(t, http.StatusNotFound, res.StatusCode)
	})
}

func testFetchBaseUrl() (*url.URL, error) {
	if v := os.Getenv("JZ_TEST_FETCH_TARGET_URL"); v != "" {
		return url.Parse(v)
	}

	host := os.Getenv("JZ_TEST_FETCH_TARGET_HOST")
	port := os.Getenv("JZ_TEST_FETCH_TARGET_PORT")
	if host == "" && port == "" {
		return nil, nil
	}

	if host == "" {
		host = "127.0.0.1"
	}
	if port == "" {
		port = "7743"
	}
	return url.Parse(fmt.Sprintf("http://%s:%s/", host, port))
}

// flagReadCloser records whether Close was called.
type flagReadCloser struct {
	io.Reader
	closed *bool
}

func (b flagReadCloser) Close() error { *b.closed = true; return nil }

func TestFetchTransportStatusAndProto(t *testing.T) {
	base, err := testFetchBaseUrl()
	if err != nil || base == nil {
		t.Skip("set JZ_TEST_FETCH_TARGET_URL to run fetch tests")
	}

	target := *base
	target.Path = "/status"
	target.RawQuery = "code=418"
	req, err := http.NewRequest(http.MethodGet, target.String(), nil)
	x.NoError(t, err)
	req = req.WithContext(t.Context())

	res, err := jz.FetchTransport{}.RoundTrip(req)
	x.NoError(t, err)
	defer res.Body.Close()

	x.Equal(t, 418, res.StatusCode)
	x.Equal(t, "418 I'm a teapot", res.Status)
	x.Equal(t, "HTTP/1.1", res.Proto)
	x.Equal(t, 1, res.ProtoMajor)
	x.Equal(t, 1, res.ProtoMinor)
}

func TestFetchTransportClosesRequestBody(t *testing.T) {
	base, err := testFetchBaseUrl()
	if err != nil || base == nil {
		t.Skip("set JZ_TEST_FETCH_TARGET_URL to run fetch tests")
	}

	t.Run("success path", func(t *testing.T) {
		closed := false
		target := *base
		target.Path = "/echo"
		req, err := http.NewRequest(http.MethodPost, target.String(), nil)
		x.NoError(t, err)
		req.Body = flagReadCloser{Reader: bytes.NewReader([]byte("hi")), closed: &closed}
		req = req.WithContext(t.Context())

		res, err := jz.FetchTransport{}.RoundTrip(req)
		x.NoError(t, err)
		res.Body.Close()
		x.True(t, closed, "req.Body must be closed by RoundTrip")
	})

	t.Run("error path", func(t *testing.T) {
		closed := false
		req, err := http.NewRequest(http.MethodPost, "http://127.0.0.1:1/echo", nil)
		x.NoError(t, err)
		req.Body = flagReadCloser{Reader: bytes.NewReader([]byte("hi")), closed: &closed}
		req = req.WithContext(t.Context())

		_, err = jz.FetchTransport{}.RoundTrip(req)
		x.Error(t, err)
		x.True(t, closed, "req.Body must be closed even on error")
	})
}

func TestFetchTransportMultiValueRequestHeader(t *testing.T) {
	base, err := testFetchBaseUrl()
	if err != nil || base == nil {
		t.Skip("set JZ_TEST_FETCH_TARGET_URL to run fetch tests")
	}

	target := *base
	target.Path = "/header"
	req, err := http.NewRequest(http.MethodGet, target.String(), nil)
	x.NoError(t, err)
	req.Header.Add("X-Multi", "alpha")
	req.Header.Add("X-Multi", "beta")
	req = req.WithContext(t.Context())

	res, err := jz.FetchTransport{}.RoundTrip(req)
	x.NoError(t, err)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	x.NoError(t, err)

	var got struct {
		Header map[string][]string `json:"header"`
	}
	x.NoError(t, json.Unmarshal(body, &got))
	joined := strings.Join(got.Header["X-Multi"], " ")
	x.Contains(t, joined, "alpha")
	x.Contains(t, joined, "beta", "both header values must survive")
}

func TestFetchTransportSetCookie(t *testing.T) {
	base, err := testFetchBaseUrl()
	if err != nil || base == nil {
		t.Skip("set JZ_TEST_FETCH_TARGET_URL to run fetch tests")
	}

	target := *base
	target.Path = "/set-cookie"
	req, err := http.NewRequest(http.MethodGet, target.String(), nil)
	x.NoError(t, err)
	req = req.WithContext(t.Context())

	res, err := jz.FetchTransport{}.RoundTrip(req)
	x.NoError(t, err)
	defer res.Body.Close()

	cookies := res.Header.Values("Set-Cookie")
	joined := strings.Join(cookies, "||")
	x.Contains(t, joined, "a=1")
	x.Contains(t, joined, "b=2")
	// When the runtime exposes getSetCookie the two cookies stay separate.
	x.Len(t, cookies, 2)
}

func TestFetchTransportNoBody(t *testing.T) {
	base, err := testFetchBaseUrl()
	if err != nil || base == nil {
		t.Skip("set JZ_TEST_FETCH_TARGET_URL to run fetch tests")
	}

	target := *base
	target.Path = "/nobody"
	req, err := http.NewRequest(http.MethodGet, target.String(), nil)
	x.NoError(t, err)
	req = req.WithContext(t.Context())

	res, err := jz.FetchTransport{}.RoundTrip(req)
	x.NoError(t, err)
	x.Equal(t, http.StatusNoContent, res.StatusCode)
	x.NotNil(t, res.Body, "Body must never be nil")
	x.Equal(t, int64(0), res.ContentLength)

	body, err := io.ReadAll(res.Body)
	x.NoError(t, err)
	x.Empty(t, body)
	x.NoError(t, res.Body.Close())
}

func TestFetchTransportContentLength(t *testing.T) {
	base, err := testFetchBaseUrl()
	if err != nil || base == nil {
		t.Skip("set JZ_TEST_FETCH_TARGET_URL to run fetch tests")
	}

	target := *base
	target.Path = "/header"
	req, err := http.NewRequest(http.MethodGet, target.String(), nil)
	x.NoError(t, err)
	req = req.WithContext(t.Context())

	res, err := jz.FetchTransport{}.RoundTrip(req)
	x.NoError(t, err)
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	x.NoError(t, err)
	// The server sends Content-Length, so it must be reflected (never left 0
	// for a non-empty streamed body).
	x.Equal(t, int64(len(body)), res.ContentLength)
}

func TestFetchTransportCancelStreamingBody(t *testing.T) {
	base, err := testFetchBaseUrl()
	if err != nil || base == nil {
		t.Skip("set JZ_TEST_FETCH_TARGET_URL to run fetch tests")
	}

	target := *base
	target.Path = "/drip"
	req, err := http.NewRequest(http.MethodGet, target.String(), nil)
	x.NoError(t, err)

	// Generous enough for the first fetch to deliver headers + first chunk,
	// but the server then holds the body open until the deadline aborts it.
	ctx, cancel := context.WithTimeout(t.Context(), 500*time.Millisecond)
	defer cancel()
	req = req.WithContext(ctx)

	res, err := jz.FetchTransport{}.RoundTrip(req)
	x.NoError(t, err) // headers + first chunk arrive before the server hangs
	defer res.Body.Close()

	// The body hangs after "hello"; the context deadline must abort the read
	// rather than blocking forever.
	t0 := time.Now()
	_, err = io.ReadAll(res.Body)
	x.Error(t, err, "streaming body read must be cancelled by the request context")
	x.WithinDuration(t, time.Now(), t0, 3*time.Second)
}
