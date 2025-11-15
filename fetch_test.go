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
	"testing"
	"time"

	"github.com/lesomnus/jz"
	"github.com/stretchr/testify/require"
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
		x := require.New(t)

		target := *base_url
		target.Path = "/header"
		req, err := http.NewRequest(http.MethodGet, target.String(), nil)
		x.NoError(err)

		req = req.WithContext(t.Context())
		req.Header.Set("X-Foo", "bar")
		res, err := jz.FetchTransport{}.RoundTrip(req)
		x.NoError(err)
		x.Equal(http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		x.NoError(err)

		var got struct {
			Method string              `json:"method"`
			URL    string              `json:"url"`
			Proto  string              `json:"proto"`
			Header map[string][]string `json:"header"`
		}
		x.NoError(json.Unmarshal(body, &got))
		x.Equal("GET", got.Method)
		x.Equal([]string{"bar"}, got.Header["X-Foo"])
	})
	t.Run("POST", func(t *testing.T) {
		x := require.New(t)

		target := *base_url
		target.Path = "/echo"
		req, err := http.NewRequest(http.MethodPost, target.String(),
			bytes.NewReader([]byte("Royale with Cheese")),
		)
		x.NoError(err)

		req = req.WithContext(t.Context())
		res, err := jz.FetchTransport{}.RoundTrip(req)
		x.NoError(err)
		x.Equal(http.StatusOK, res.StatusCode)

		body, err := io.ReadAll(res.Body)
		x.NoError(err)
		x.Equal("Royale with Cheese", string(body))
	})
	t.Run("with context", func(t *testing.T) {
		x := require.New(t)

		target := *base_url
		target.Path = "/hang"
		req, err := http.NewRequest(http.MethodGet, target.String(), nil)
		x.NoError(err)

		t0 := time.Now()
		ctx, cancel := context.WithTimeout(t.Context(), 30*time.Millisecond)
		defer cancel()

		req = req.WithContext(ctx)
		_, err = jz.FetchTransport{}.RoundTrip(req)
		t1 := time.Now()
		x.WithinRange(t1, t0, t0.Add(50*time.Millisecond))
		x.ErrorIs(err, context.DeadlineExceeded)
	})
	t.Run("404", func(t *testing.T) {
		x := require.New(t)

		target := *base_url
		target.Path = "/not-found"
		req, err := http.NewRequest(http.MethodGet, target.String(), nil)
		x.NoError(err)

		req = req.WithContext(t.Context())
		res, err := jz.FetchTransport{}.RoundTrip(req)
		x.NoError(err)
		x.Equal(http.StatusNotFound, res.StatusCode)
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
