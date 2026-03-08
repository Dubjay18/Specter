package divergence

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Dubjay/specter/internal/types"
)

type mockStore struct {
	saved []types.DivergenceEvent
}

func (m *mockStore) Save(event types.DivergenceEvent) error {
	m.saved = append(m.saved, event)
	return nil
}

func (m *mockStore) List(limit int) ([]types.DivergenceEvent, error) {
	if limit <= 0 || limit >= len(m.saved) {
		return m.saved, nil
	}
	return m.saved[:limit], nil
}

func (m *mockStore) Close() error {
	return nil
}

func newEngineForTest() *DivergenceEngine {
	return NewEngine(&mockStore{})
}

// makeCapture is a test helper to build a CapturedResponse quickly.
// Import your actual CapturedResponse type from the proxy package.
// If it lives in the same package, just use it directly.
func makeCapture(status int, body string, latencyMs int64) *types.CapturedResponse {
	return &types.CapturedResponse{
		StatusCode: status,
		Body:       []byte(body),
		Latency:    time.Duration(latencyMs) * time.Millisecond,
	}
}

// TestAnalyzePopulatesEvent checks that Analyze always returns a fully populated event.
func TestAnalyzePopulatesEvent(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/users/42", nil)
	live := makeCapture(200, `{"id":42}`, 80)
	shadow := makeCapture(200, `{"id":42}`, 90)
	engine := newEngineForTest()

	event := engine.Analyze(req, live, shadow)

	if event.ID == "" {
		t.Error("expected non-empty event ID")
	}
	if event.Timestamp.IsZero() {
		t.Error("expected non-zero Timestamp")
	}
	if event.RequestPath != "/api/users/42" {
		t.Errorf("expected RequestPath '/api/users/42', got %q", event.RequestPath)
	}
	if event.Method != "GET" {
		t.Errorf("expected Method 'GET', got %q", event.Method)
	}
}

// TestAnalyzeNoDivergence checks that Diverged=false for identical responses.
func TestAnalyzeNoDivergence(t *testing.T) {
	req := httptest.NewRequest("GET", "/health", nil)
	live := makeCapture(200, `{"status":"ok"}`, 50)
	shadow := makeCapture(200, `{"status":"ok"}`, 55)
	engine := newEngineForTest()

	event := engine.Analyze(req, live, shadow)

	if event.Diverged {
		t.Errorf("expected Diverged=false for identical responses, got true. Diffs: %+v", event.BodyDiff)
	}
	if event.StatusDiff != nil {
		t.Errorf("expected nil StatusDiff, got %+v", event.StatusDiff)
	}
}

// TestAnalyzeBodyDivergence checks Diverged=true when bodies differ.
func TestAnalyzeBodyDivergence(t *testing.T) {
	req := httptest.NewRequest("GET", "/api/cart", nil)
	live := makeCapture(200, `{"total": 99.99}`, 60)
	shadow := makeCapture(200, `{"total": 0.00}`, 65) // total is wrong in shadow
	engine := newEngineForTest()

	event := engine.Analyze(req, live, shadow)

	if !event.Diverged {
		t.Error("expected Diverged=true when body totals differ, got false")
	}
	if len(event.BodyDiff) == 0 {
		t.Error("expected BodyDiff to be populated, got empty slice")
	}
}

// TestAnalyzeStatusDivergence checks Diverged=true when status codes differ.
func TestAnalyzeStatusDivergence(t *testing.T) {
	req := httptest.NewRequest("DELETE", "/api/items/1", nil)
	live := makeCapture(204, ``, 40)                   // 204 No Content
	shadow := makeCapture(500, `{"error":"oops"}`, 45) // 500 Internal Error
	engine := newEngineForTest()

	event := engine.Analyze(req, live, shadow)

	if !event.Diverged {
		t.Error("expected Diverged=true when status codes differ, got false")
	}
	if event.StatusDiff == nil {
		t.Fatal("expected non-nil StatusDiff when codes differ")
	}
	if event.StatusDiff.Live != 204 {
		t.Errorf("expected StatusDiff.Live=204, got %d", event.StatusDiff.Live)
	}
	if event.StatusDiff.Shadow != 500 {
		t.Errorf("expected StatusDiff.Shadow=500, got %d", event.StatusDiff.Shadow)
	}
}

// TestAnalyzeLatencyAlwaysRecorded checks that latency info is always recorded,
// even when there's no divergence.
func TestAnalyzeLatencyAlwaysRecorded(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	live := makeCapture(200, `{}`, 100)
	shadow := makeCapture(200, `{}`, 300)
	engine := newEngineForTest()

	event := engine.Analyze(req, live, shadow)

	if event.LatencyDiff.LiveMs == 0 {
		t.Error("expected LiveMs to be populated, got 0")
	}
	if event.LatencyDiff.ShadowMs == 0 {
		t.Error("expected ShadowMs to be populated, got 0")
	}
	if event.LatencyDiff.DeltaMs == 0 {
		t.Error("expected DeltaMs to be populated, got 0")
	}
}

// TestAnalyzePOSTRequest checks that POST method requests are handled correctly.
func TestAnalyzePOSTRequest(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/orders", nil)
	live := makeCapture(201, `{"order_id":"xyz"}`, 120)
	shadow := makeCapture(201, `{"order_id":"xyz"}`, 130)
	engine := newEngineForTest()

	event := engine.Analyze(req, live, shadow)

	if event.Method != "POST" {
		t.Errorf("expected Method 'POST', got %q", event.Method)
	}
	if event.Diverged {
		t.Error("expected no divergence for identical POST responses")
	}
}
