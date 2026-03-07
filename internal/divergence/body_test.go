package divergence

import (
	"testing"
)

// TestIdenticalBodiesNoDiff checks that two identical JSON payloads produce no diffs.
func TestIdenticalBodiesNoDiff(t *testing.T) {
	live := []byte(`{"user_id": 42, "name": "alice", "active": true}`)
	shadow := []byte(`{"user_id": 42, "name": "alice", "active": true}`)

	diffs, err := DiffBodies(live, shadow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs for identical bodies, got %d: %+v", len(diffs), diffs)
	}
}

// TestChangedValueDetected checks that a different field value is caught.
func TestChangedValueDetected(t *testing.T) {
	live := []byte(`{"user_id": 42, "name": "alice"}`)
	shadow := []byte(`{"user_id": 99, "name": "alice"}`) // user_id changed

	diffs, err := DiffBodies(live, shadow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diffs) == 0 {
		t.Fatal("expected at least 1 diff for changed user_id, got 0")
	}

	// At least one diff should reference the user_id path
	found := false
	for _, d := range diffs {
		if d.Path == "/user_id" || d.Path == "user_id" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected a diff at path 'user_id', got diffs: %+v", diffs)
	}
}

// TestMissingFieldDetected checks that a field present in live but absent in shadow is caught.
func TestMissingFieldDetected(t *testing.T) {
	live := []byte(`{"user_id": 42, "premium": true}`)
	shadow := []byte(`{"user_id": 42}`) // "premium" field is gone

	diffs, err := DiffBodies(live, shadow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diffs) == 0 {
		t.Fatal("expected at least 1 diff for missing 'premium' field, got 0")
	}
}

// TestAddedFieldDetected checks that a field present in shadow but not in live is caught.
func TestAddedFieldDetected(t *testing.T) {
	live := []byte(`{"user_id": 42}`)
	shadow := []byte(`{"user_id": 42, "extra_field": "oops"}`) // shadow has extra field

	diffs, err := DiffBodies(live, shadow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diffs) == 0 {
		t.Fatal("expected at least 1 diff for added 'extra_field', got 0")
	}
}

// TestNestedChangeDetected checks that changes inside nested objects are caught.
func TestNestedChangeDetected(t *testing.T) {
	live := []byte(`{"user": {"id": 1, "address": {"city": "New York"}}}`)
	shadow := []byte(`{"user": {"id": 1, "address": {"city": "London"}}}`) // city changed

	diffs, err := DiffBodies(live, shadow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diffs) == 0 {
		t.Fatal("expected at least 1 diff for nested city change, got 0")
	}
}

// TestTypeChangeDetected checks that a field changing type (e.g. string → int) is caught.
func TestTypeChangeDetected(t *testing.T) {
	live := []byte(`{"count": "five"}`)   // string
	shadow := []byte(`{"count": 5}`)      // now a number

	diffs, err := DiffBodies(live, shadow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diffs) == 0 {
		t.Fatal("expected at least 1 diff for type change string→int, got 0")
	}
}

// TestInvalidJSONLiveReturnsError checks that malformed live JSON causes an error.
func TestInvalidJSONLiveReturnsError(t *testing.T) {
	live := []byte(`{this is not json`)
	shadow := []byte(`{"valid": true}`)

	_, err := DiffBodies(live, shadow)
	if err == nil {
		t.Error("expected error for invalid live JSON, got nil")
	}
}

// TestInvalidJSONShadowReturnsError checks that malformed shadow JSON causes an error.
func TestInvalidJSONShadowReturnsError(t *testing.T) {
	live := []byte(`{"valid": true}`)
	shadow := []byte(`not json at all`)

	_, err := DiffBodies(live, shadow)
	if err == nil {
		t.Error("expected error for invalid shadow JSON, got nil")
	}
}

// TestEmptyBodiesNoDiff checks that two empty JSON objects are equal.
func TestEmptyBodiesNoDiff(t *testing.T) {
	live := []byte(`{}`)
	shadow := []byte(`{}`)

	diffs, err := DiffBodies(live, shadow)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diffs) != 0 {
		t.Errorf("expected 0 diffs for empty objects, got %d", len(diffs))
	}
}
