package divergence

import (
	"net/http/httptest"
	"testing"
)

func TestStatsSnapshotTracksAnalyzeResults(t *testing.T) {
	engine := newEngineForTest()

	identicalReq := httptest.NewRequest("GET", "/health", nil)
	engine.Analyze(identicalReq, makeCapture(200, `{"ok":true}`, 40), makeCapture(200, `{"ok":true}`, 45))

	divergedReq := httptest.NewRequest("GET", "/api/items", nil)
	engine.Analyze(divergedReq, makeCapture(200, `{"count":1}`, 50), makeCapture(500, `{"error":"boom"}`, 70))

	snapshot := engine.StatsSnapshot()

	if snapshot.TotalRequests != 2 {
		t.Fatalf("expected total_requests=2, got %d", snapshot.TotalRequests)
	}
	if snapshot.Matches != 1 {
		t.Fatalf("expected matches=1, got %d", snapshot.Matches)
	}
	if snapshot.Divergences != 1 {
		t.Fatalf("expected divergences=1, got %d", snapshot.Divergences)
	}
	if snapshot.StatusMismatches != 1 {
		t.Fatalf("expected status_mismatches=1, got %d", snapshot.StatusMismatches)
	}
	if snapshot.BodyMismatches != 1 {
		t.Fatalf("expected body_mismatches=1, got %d", snapshot.BodyMismatches)
	}
	if snapshot.LastEventAt == nil {
		t.Fatal("expected last_event_at to be set")
	}
	if snapshot.DivergenceRate != 0.5 {
		t.Fatalf("expected divergence_rate=0.5, got %v", snapshot.DivergenceRate)
	}
}
