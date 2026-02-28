package proxy

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

type Proxy struct {
	liveURL      string
	shadowURL    string
	ReverseProxy *httputil.ReverseProxy
}

func New(liveURL string) *Proxy {
	u, err := url.Parse(liveURL)
	if err != nil {
		log.Fatalf("specter: invalid live URL %q: %v", liveURL, err)
	}
	ph := httputil.NewSingleHostReverseProxy(u)
	// Custom error handler so a broken live service logs clearly
	// instead of returning an empty response to the client.
	ph.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("specter: error forwarding to live (%s): %v", liveURL, err)
		http.Error(w, "bad gateway", http.StatusBadGateway)
	}
	return &Proxy{
		liveURL:      liveURL,
		ReverseProxy: ph,
	}
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Buffer the body ONCE so both live and shadow can read it.
	// HTTP bodies are streams — once read they're gone.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "error reading request body", http.StatusInternalServerError)
		return
	}
	r.Body.Close()
	// Restore the body for the reverse proxy (live).
	r.Body = io.NopCloser(bytes.NewReader(body))
	if p.shadowURL != "" {
		go p.forkToShadow(r, body)
	}
	log.Printf("specter: %s %s → live", r.Method, r.URL.Path)
	p.ReverseProxy.ServeHTTP(w, r)
}
