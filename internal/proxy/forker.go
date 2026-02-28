package proxy

import (
	"bytes"
	"context"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"
)

var shadowClient = &http.Client{
	Timeout: 10 * time.Second,
}

func (p *Proxy) forkToShadow(r *http.Request, body []byte) {
	// Parse the shadow URL
	shadowTarget, err := url.Parse(p.shadowURL)
	if err != nil {
		log.Printf("specter: invalid shadow URL %q: %v", p.shadowURL, err)
		return
	}
	cloned := r.Clone(context.Background())
	// Rewrite the URL to point at the shadow server.
	// Keep the original path, query string, everything — just swap the host.
	cloned.URL.Scheme = shadowTarget.Scheme
	cloned.URL.Host = shadowTarget.Host
	cloned.RequestURI = ""

	// Restore the body — the original was already read and buffered.
	// Without this, the shadow request body would be empty.
	cloned.Body = io.NopCloser(bytes.NewReader(body))
	cloned.ContentLength = int64(len(body))
	start := time.Now()
	resp, err := shadowClient.Do(cloned)
	latency := time.Since(start)
	if err != nil {
		log.Printf("specter: shadow error (%s %s): %v", r.Method, r.URL.Path, err)
		return
	}

	captured, err := captureResponse(resp, latency)
	if err != nil {
		log.Printf("specter: failed to capture shadow response: %v", err)
		return
	}



log.Printf("specter: shadow → status=%d latency=%v body=%s",
		captured.StatusCode, captured.Latency, captured.Body)

}
