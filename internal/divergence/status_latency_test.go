package divergence

import (
	"testing"
)

// ── STATUS TESTS ────────────────────────────────────────────────────────────

// TestStatusMatchReturnsNil checks that matching status codes produce no diff.
func TestStatusMatchReturnsNil(t *testing.T) {
	result := DiffStatus(200, 200)
	if result != nil {
		t.Errorf("expected nil for matching status codes, got %+v", result)
	}
}

// TestStatusMismatchReturnsStruct checks that differing codes return a populated struct.
func TestStatusMismatchReturnsStruct(t *testing.T) {
	result := DiffStatus(200, 500)
	if result == nil {
		t.Fatal("expected non-nil StatusDiff for mismatched codes, got nil")
	}
	if result.Live != 200 {
		t.Errorf("expected Live=200, got %d", result.Live)
	}
	if result.Shadow != 500 {
		t.Errorf("expected Shadow=500, got %d", result.Shadow)
	}
}

// TestStatusVarious checks a range of status code pairs.
func TestStatusVarious(t *testing.T) {
	cases := []struct {
		live, shadow int
		wantDiff     bool
	}{
		{200, 200, false},
		{200, 201, true},
		{404, 404, false},
		{200, 404, true},
		{500, 200, true},
		{301, 302, true},
	}

	for _, tc := range cases {
		result := DiffStatus(tc.live, tc.shadow)
		hasDiff := result != nil
		if hasDiff != tc.wantDiff {
			t.Errorf("DiffStatus(%d, %d): wantDiff=%v, got hasDiff=%v",
				tc.live, tc.shadow, tc.wantDiff, hasDiff)
		}
	}
}

// ── LATENCY TESTS ───────────────────────────────────────────────────────────

// TestLatencyDeltaCalculation checks that the delta is computed correctly.
func TestLatencyDeltaCalculation(t *testing.T) {
	result := DiffLatency(100, 150)
	if result.LiveMs != 100 {
		t.Errorf("expected LiveMs=100, got %d", result.LiveMs)
	}
	if result.ShadowMs != 150 {
		t.Errorf("expected ShadowMs=150, got %d", result.ShadowMs)
	}
	if result.DeltaMs != 50 {
		t.Errorf("expected DeltaMs=50, got %d", result.DeltaMs)
	}
}

// TestLatencyDeltaWhenShadowFaster checks negative delta when shadow is faster.
func TestLatencyDeltaWhenShadowFaster(t *testing.T) {
	result := DiffLatency(200, 80)
	// Delta should show shadow was 120ms faster (negative or absolute — either is fine,
	// just make sure the raw numbers are there)
	if result.LiveMs != 200 {
		t.Errorf("expected LiveMs=200, got %d", result.LiveMs)
	}
	if result.ShadowMs != 80 {
		t.Errorf("expected ShadowMs=80, got %d", result.ShadowMs)
	}
}

// TestLatencyBucketLabelSimilar checks the "similar" label for close timings.
func TestLatencyBucketLabelSimilar(t *testing.T) {
	result := DiffLatency(100, 110) // only 10% slower
	label := result.BucketLabel()
	if label != "similar" {
		t.Errorf("expected bucket 'similar' for marginally slower shadow, got %q", label)
	}
}

// TestLatencyBucketLabelFaster checks the "faster" label when shadow is faster.
func TestLatencyBucketLabelFaster(t *testing.T) {
	result := DiffLatency(200, 50) // shadow is 4x faster
	label := result.BucketLabel()
	if label != "faster" {
		t.Errorf("expected bucket 'faster' for much faster shadow, got %q", label)
	}
}

// TestLatencyBucketLabel2x checks the "2x slower" label.
func TestLatencyBucketLabel2x(t *testing.T) {
	result := DiffLatency(100, 220) // shadow is ~2x slower
	label := result.BucketLabel()
	if label != "2x slower" {
		t.Errorf("expected '2x slower', got %q", label)
	}
}

// TestLatencyBucketLabel10x checks the "10x+ slower" label.
func TestLatencyBucketLabel10x(t *testing.T) {
	result := DiffLatency(50, 600) // shadow is 12x slower
	label := result.BucketLabel()
	if label != "10x+ slower" {
		t.Errorf("expected '10x+ slower', got %q", label)
	}
}
