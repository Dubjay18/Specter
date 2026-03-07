package types

import (
	"net/http"
	"time"
)

type CapturedResponse struct {
	StatusCode int
	Body []byte
	Headers http.Header
	Latency time.Duration
}