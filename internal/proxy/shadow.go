package proxy

import (
	"io"
	"net/http"
	"time"
)

type CapturedResponse struct {
	StatusCode int
	Body []byte
	Headers http.Header
	Latency time.Duration
}


// to save shadow's response
func captureResponse(resp *http.Response, latency time.Duration) (*CapturedResponse, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &CapturedResponse{
		StatusCode: resp.StatusCode,
		Body:       body,
		Headers:    resp.Header,
		Latency:    latency,
	}, nil
}
