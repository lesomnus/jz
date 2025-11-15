package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
)

func main() {
	host := flag.String("host", "127.0.0.1", "host to bind")
	port := flag.Int("port", 7743, "port to bind")
	flag.Parse()

	mux := http.NewServeMux()
	mux.HandleFunc("/header", handleHeader)
	mux.HandleFunc("/hang", handleHang)
	mux.HandleFunc("POST /echo", handleEcho)

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
