# jz

[![Go Reference](https://pkg.go.dev/badge/github.com/lesomnus/jz.svg)](https://pkg.go.dev/github.com/lesomnus/jz)
[![Go Report Card](https://goreportcard.com/badge/github.com/lesomnus/jz)](https://goreportcard.com/report/github.com/lesomnus/jz)

A utility package that makes JavaScript bindings in WASM-compiled Go code feel like native Go code. Write idiomatic Go code that seamlessly interacts with browser JavaScript APIs.

I hope this package becomes obsolete soon when WASI becomes natively supported in browsers, eliminating the need for these JavaScript bindings altogether.

## Features

- **Promise Handling**: Await JavaScript Promises with Go's context support
- **HTTP Client**: Use `net/http` with `fetch` API through a `FetchTransport`
- **Stream I/O**: Convert JavaScript ReadableStreams to `io.Reader`
- **Type Conversion**: Unmarshal JavaScript types to Go.

## Installation

```bash
go get github.com/lesomnus/jz
```

## Usage

This library is designed for Go code compiled to WebAssembly targeting the browser (`GOOS=js GOARCH=wasm`).

### Promise Handling

Await JavaScript Promises with Go's familiar error handling:

```go
import "github.com/lesomnus/jz"

// Await a Promise.
result, err := jz.Await(jsPromise)
if err != nil {
	// Handle rejection.
}

// Await with context support.
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

result, err := jz.AwaitContext(ctx, jsPromise)
if err != nil {
	// Handle rejection or timeout.
}

// Create a Promise from Go code.
promise := jz.Promise(func() (any, any) {
	result, err := doSomething()
	return result, err
})
```

### HTTP with Fetch API

Use the standard `net/http` package with the browser's `fetch` API:

```go
import (
	"net/http"
	"github.com/lesomnus/jz"
)

client := &http.Client{
	Transport: jz.FetchTransport{},
}

res, err := client.Get("https://api.example.com/data")
if err != nil {
	// Handle error.
}
defer res.Body.Close()

body, err := io.ReadAll(res.Body)
```

### Stream Reading

Convert JavaScript ReadableStreams to Go's `io.Reader`:

```go
// jsStream is a JavaScript ReadableStream.
reader, err := jz.NewReader(jsStream)
if err != nil {
	// Handle error
}
defer reader.Close()

// Use like any io.Reader.
data, err := io.ReadAll(reader)
```

### Type Conversion

Unmarshal JavaScript objects to Go structs:

```go
type User struct {
	Name  string
	Email string
	Age   int
}

var user User
err := jz.Unmarshal(jsValue, &user)
if err != nil {
	// Handle error.
}
```

### Utility Functions

```go
// Convert JavaScript values to readable strings.
str := jz.Stringify(jsValue)

// Access nested properties like optional chaining.
window := js.Global()
fetch := jz.GetX(window, "fetch")
json := jz.GetX(window, "JSON", "stringify")
```

## Testing

For full tests, including `FetchTransport`, you need an HTTP server to target.
```bash
$ go run ./internal/httptest
```

Then run tests with wasm/js.
The required environment variables are defined at [./scripts/setup.sh](./scripts/setup.sh).
```bash
$ source ./scripts/setup.sh
$ go test
```

## Related Works

- [gowebapi](https://github.com/gowebapi)
- [dominikh/go-js-dom](https://github.com/dominikh/go-js-dom)
- [Nigel2392/jsext](https://github.com/Nigel2392/jsext)
