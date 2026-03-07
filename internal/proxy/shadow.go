package proxy

import (
	"io"
	"net/http"
	"time"

	"github.com/Dubjay/specter/internal/types"
)

// to save shadow's response
func captureResponse(resp *http.Response, latency time.Duration) (*types.CapturedResponse, error) {
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return &types.CapturedResponse{
		StatusCode: resp.StatusCode,
		Body:       body,
		Headers:    resp.Header,
		Latency:    latency,
	}, nil
}
