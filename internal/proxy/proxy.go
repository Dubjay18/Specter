package proxy

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/Dubjay/specter/internal/divergence"
	"github.com/Dubjay/specter/internal/ring"
)

type Proxy struct {
	liveURL      string
	shadowURL    string
	nodeName     string
	routingKey   string
	ring         *ring.Ring
	ReverseProxy *httputil.ReverseProxy
	divergence   *divergence.DivergenceEngine
}

func New(liveURL string, shadowURL string, nodeName string, routingKey string, hashRing *ring.Ring, engine *divergence.DivergenceEngine) *Proxy {
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
		shadowURL:    shadowURL,
		nodeName:     nodeName,
		routingKey:   routingKey,
		ring:         hashRing,
		ReverseProxy: ph,
		divergence:   engine,
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

	userID := r.Header.Get(p.routingKey)
	if userID == "" {
		userID = "unknown"
	}

	owner := p.nodeName
	if p.ring != nil && userID != "unknown" {
		if resolved := p.ring.GetOwner(userID); resolved != "" {
			owner = resolved
		}
	}

	log.Printf("[node: %s] handling request for user %s (owner: %s)", p.nodeName, userID, owner)

	if p.shadowURL != "" {
		go p.forkToShadow(r, body)
	}
	log.Printf("specter: %s %s → live", r.Method, r.URL.Path)
	p.ReverseProxy.ServeHTTP(w, r)
}
