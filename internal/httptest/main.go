package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strconv"
)

func main() {
	host := flag.String("host", "127.0.0.1", "host to bind")
	port := flag.Int("port", 7743, "port to bind")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/header", handleHeader)
	mux.HandleFunc("/hang", handleHang)
	mux.HandleFunc("POST /echo", handleEcho)
	mux.HandleFunc("/set-cookie", handleSetCookie)
	mux.HandleFunc("/status", handleStatus)
	mux.HandleFunc("/nobody", handleNoBody)
	mux.HandleFunc("/drip", handleDrip)

	addr := net.JoinHostPort(*host, fmt.Sprintf("%d", *port))
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("listen error on %s: %v", addr, err)
	}

	log.Printf("http server listening on %s", ln.Addr())
	if err := http.Serve(ln, logRequests(mux)); err != nil {
		log.Fatalf("server error: %v", err)
	}
}

func handleHeader(w http.ResponseWriter, r *http.Request) {
	headers := make(map[string][]string, len(r.Header)+1)
	for k, v := range r.Header {
		vv := make([]string, len(v))
		copy(vv, v)
		headers[k] = vv
	}
	if r.Host != "" {
		headers["Host"] = []string{r.Host}
	}

	resp := struct {
		Method string              `json:"method"`
		URL    string              `json:"url"`
		Proto  string              `json:"proto"`
		Header map[string][]string `json:"header"`
	}{
		Method: r.Method,
		URL:    r.URL.String(),
		Proto:  r.Proto,
		Header: headers,
	}

	b, err := json.Marshal(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(b); err != nil {
		return
	}
	fmt.Println(string(b))
}

func handleHang(w http.ResponseWriter, r *http.Request) {
	<-r.Context().Done()
}

// handleSetCookie emits multiple Set-Cookie headers so clients can verify that
// they are not collapsed into a single value.
func handleSetCookie(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Set-Cookie", "a=1; Path=/")
	w.Header().Add("Set-Cookie", "b=2; Path=/")
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprint(w, "ok")
}

// handleStatus responds with the status code given by ?code= (default 200).
func handleStatus(w http.ResponseWriter, r *http.Request) {
	code := http.StatusOK
	if v := r.URL.Query().Get("code"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			code = n
		}
	}
	w.WriteHeader(code)
	fmt.Fprint(w, http.StatusText(code))
}

// handleNoBody responds with 204 No Content (no response body).
func handleNoBody(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

// handleDrip flushes response headers and a first chunk, then holds the
// connection open until the client disconnects. It lets clients exercise
// cancellation of a streaming response body after the headers have arrived.
func handleDrip(w http.ResponseWriter, r *http.Request) {
	fl, ok := w.(http.Flusher)
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	if ok {
		fl.Flush()
	}
	io.WriteString(w, "hello")
	if ok {
		fl.Flush()
	}
	<-r.Context().Done()
}

func handleEcho(w http.ResponseWriter, r *http.Request) {
	if ct := r.Header.Get("Content-Type"); ct != "" {
		w.Header().Set("Content-Type", ct)
	}
	if r.Body == nil {
		return
	}

	defer r.Body.Close()
	if _, err := io.Copy(w, r.Body); err != nil {
		return
	}
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.String())
		next.ServeHTTP(w, r)
	})
}
