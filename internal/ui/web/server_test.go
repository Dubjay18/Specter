package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Dubjay/specter/internal/divergence"
)

type fakeStatsProvider struct {
	snapshot divergence.StatsSnapshot
}

func (f *fakeStatsProvider) StatsSnapshot() divergence.StatsSnapshot {
	return f.snapshot
}

func TestGetStatsReturnsJSONSnapshot(t *testing.T) {
	now := time.Now().UTC()
	provider := &fakeStatsProvider{
		snapshot: divergence.StatsSnapshot{
			StartedAt:     now,
			TotalRequests: 10,
			Divergences:   3,
			Matches:       7,
			LatencyBuckets: map[string]uint64{
				"faster":      2,
				"similar":     3,
				"2x slower":   1,
				"5x slower":   1,
				"10x slower":  2,
				"10x+ slower": 1,
			},
			DivergenceRate: 0.3,
		},
	}

	server := NewServer(provider)
	mux := http.NewServeMux()
	server.RegisterRoutes(mux)

	req := httptest.NewRequest(http.MethodGet, "/api/stats", nil)
	rr := httptest.NewRecorder()
	mux.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("expected Content-Type application/json, got %q", got)
	}

	var payload divergence.StatsSnapshot
	if err := json.NewDecoder(rr.Body).Decode(&payload); err != nil {
		t.Fatalf("failed to decode JSON payload: %v", err)
	}
	if payload.TotalRequests != 10 {
		t.Fatalf("expected total_requests=10, got %d", payload.TotalRequests)
	}
	if payload.Divergences != 3 {
		t.Fatalf("expected divergences=3, got %d", payload.Divergences)
	}
}
