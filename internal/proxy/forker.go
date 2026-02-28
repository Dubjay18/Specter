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
	resp, err := shadowClient.Do(cloned)
	if err != nil {
		log.Printf("specter: shadow error (%s %s): %v", r.Method, r.URL.Path, err)
		return
	}


	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	log.Printf("specter: shadow responded %d (%s %s)", resp.StatusCode, r.Method, r.URL.Path)

}
