package proxy

import (
	"bytes"
	"context"
	"io"
	"log"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

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

const interNodeForwardTimeout = 3 * time.Second

var interNodeClient = &http.Client{
	Timeout: interNodeForwardTimeout,
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
	routingKeyPresent := userID != ""
	if !routingKeyPresent {
		userID = "unknown"
	}

	owner := p.nodeName
	if p.ring != nil && userID != "unknown" {
		if resolved := p.ring.GetOwner(userID); resolved != "" {
			owner = resolved
		}
	}

	log.Printf("[node: %s] handling request for user %s (owner: %s)", p.nodeName, userID, owner)

	if !routingKeyPresent {
		log.Printf("specter: missing routing key header %q, handling locally", p.routingKey)
	} else if owner != p.nodeName && r.Header.Get("X-Specter-Forwarded-By") == "" {
		if p.forwardToOwner(w, r, body, owner) {
			return
		}
		log.Printf("specter: owner %s unavailable, falling back to local handling", owner)
	}

	if p.shadowURL != "" {
		go p.forkToShadow(r, body)
	}
	log.Printf("specter: %s %s → live", r.Method, r.URL.Path)
	p.ReverseProxy.ServeHTTP(w, r)
}

func (p *Proxy) forwardToOwner(w http.ResponseWriter, r *http.Request, body []byte, owner string) bool {
	baseURL := interNodeBaseURL(owner, r.Host)
	target, err := url.Parse(baseURL)
	if err != nil {
		log.Printf("specter: invalid owner URL %q for node %s: %v", baseURL, owner, err)
		return false
	}

	ctx, cancel := context.WithTimeout(r.Context(), interNodeForwardTimeout)
	defer cancel()

	forwardedReq := r.Clone(ctx)
	forwardedReq.URL.Scheme = target.Scheme
	forwardedReq.URL.Host = target.Host
	forwardedReq.Host = target.Host
	forwardedReq.RequestURI = ""
	forwardedReq.Body = io.NopCloser(bytes.NewReader(body))
	forwardedReq.ContentLength = int64(len(body))
	forwardedReq.Header.Set("X-Specter-Forwarded-By", p.nodeName)

	start := time.Now()
	resp, err := interNodeClient.Do(forwardedReq)
	if err != nil {
		log.Printf("specter: forwarding to owner %s failed after %v: %v", owner, time.Since(start), err)
		return false
	}
	defer resp.Body.Close()

	for name, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(name, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	if _, err := io.Copy(w, resp.Body); err != nil {
		log.Printf("specter: failed to copy owner %s response body: %v", owner, err)
	}

	log.Printf("specter: %s %s → owner %s (%v)", r.Method, r.URL.Path, owner, time.Since(start))
	return true
}

func interNodeBaseURL(owner, requestHost string) string {
	if strings.HasPrefix(owner, "http://") || strings.HasPrefix(owner, "https://") {
		return owner
	}

	if _, _, err := net.SplitHostPort(owner); err == nil {
		return "http://" + owner
	}

	if _, requestPort, err := net.SplitHostPort(requestHost); err == nil && requestPort != "" {
		return "http://" + net.JoinHostPort(owner, requestPort)
	}

	return "http://" + owner
}
