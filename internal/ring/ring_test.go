// DAY 16 TEST — Consistent hash ring
// Run with: go test ./internal/ring/ -run TestRing -v
//
// What this tests:
//   - GetOwner returns a node for any key
//   - The same key always maps to the same node (determinism)
//   - Removing a node only reassigns ~1/N keys (core consistent hashing guarantee)
//   - Adding a node redistributes some but not all keys
//   - An empty ring returns an error or empty string (not a panic)

package ring

import (
	"fmt"
	"testing"
)

// TestRingGetOwnerIsDeterministic checks that the same key always returns the same node.
func TestRingGetOwnerIsDeterministic(t *testing.T) {
	r := NewRing(100) // 100 virtual nodes per real node
	r.AddNode("node-a")
	r.AddNode("node-b")
	r.AddNode("node-c")

	key := "user-12345"
	first := r.GetOwner(key)
	if first == "" {
		t.Fatal("GetOwner returned empty string")
	}

	// Call 100 times — must always return the same node
	for i := 0; i < 100; i++ {
		got := r.GetOwner(key)
		if got != first {
			t.Errorf("GetOwner(%q) returned %q on iteration %d, expected %q (not deterministic)", key, got, i, first)
		}
	}
}

// TestRingDistribution checks that keys are distributed reasonably across nodes.
// No node should own more than 60% of keys (roughly even distribution).
func TestRingDistribution(t *testing.T) {
	r := NewRing(150)
	nodes := []string{"node-a", "node-b", "node-c"}
	for _, n := range nodes {
		r.AddNode(n)
	}

	counts := make(map[string]int)
	const total = 300
	for i := 0; i < total; i++ {
		key := fmt.Sprintf("user-%d", i)
		owner := r.GetOwner(key)
		counts[owner]++
	}

	for node, count := range counts {
		pct := float64(count) / float64(total) * 100
		if pct > 60 {
			t.Errorf("node %q owns %.0f%% of keys — distribution is too uneven", node, pct)
		}
	}
}

// TestRingRemoveNodeReassignsMinimumKeys is the core consistent hashing guarantee.
// When a node is removed, only the keys it owned should move — not all keys.
func TestRingRemoveNodeReassignsMinimumKeys(t *testing.T) {
	r := NewRing(150)
	r.AddNode("node-a")
	r.AddNode("node-b")
	r.AddNode("node-c")

	const total = 300
	// Record who owns each key before removal
	before := make(map[string]string, total)
	for i := 0; i < total; i++ {
		key := fmt.Sprintf("user-%d", i)
		before[key] = r.GetOwner(key)
	}

	// Remove one node
	r.RemoveNode("node-c")

	// Count how many keys changed owner
	changed := 0
	for i := 0; i < total; i++ {
		key := fmt.Sprintf("user-%d", i)
		if r.GetOwner(key) != before[key] {
			changed++
		}
	}

	// With 3 nodes, removing 1 should move ~1/3 of keys
	// Allow a generous range: between 20% and 50%
	pct := float64(changed) / float64(total) * 100
	if pct < 20 || pct > 50 {
		t.Errorf("removing 1 of 3 nodes moved %.0f%% of keys (expected 20%%–50%%)", pct)
	}
}

// TestRingAddNodeReassignsMinimumKeys checks that adding a node only pulls
// keys from existing nodes, not all of them.
func TestRingAddNodeReassignsMinimumKeys(t *testing.T) {
	r := NewRing(150)
	r.AddNode("node-a")
	r.AddNode("node-b")

	const total = 300
	before := make(map[string]string, total)
	for i := 0; i < total; i++ {
		key := fmt.Sprintf("user-%d", i)
		before[key] = r.GetOwner(key)
	}

	r.AddNode("node-c") // add a third node

	changed := 0
	for i := 0; i < total; i++ {
		key := fmt.Sprintf("user-%d", i)
		after := r.GetOwner(key)
		if after != before[key] {
			// Any change must be to "node-c" (the new node)
			if after != "node-c" {
				t.Errorf("key %q moved from %q to %q — should only move to new node 'node-c'",
					key, before[key], after)
			}
			changed++
		}
	}

	pct := float64(changed) / float64(total) * 100
	// With 3 nodes, new node should take ~33%
	if pct < 20 || pct > 50 {
		t.Errorf("adding 1 node to 2-node ring pulled %.0f%% of keys (expected ~20%%–50%%)", pct)
	}
}

// TestRingEmptyRing checks that an empty ring doesn't panic.
func TestRingEmptyRing(t *testing.T) {
	r := NewRing(100)

	defer func() {
		if rec := recover(); rec != nil {
			t.Errorf("GetOwner on empty ring panicked: %v", rec)
		}
	}()

	owner := r.GetOwner("any-key")
	// Either empty string or an error is acceptable — just no panic
	_ = owner
}

// TestRingOneNode checks that a single-node ring always returns that node.
func TestRingOneNode(t *testing.T) {
	r := NewRing(100)
	r.AddNode("only-node")

	for i := 0; i < 50; i++ {
		key := fmt.Sprintf("key-%d", i)
		owner := r.GetOwner(key)
		if owner != "only-node" {
			t.Errorf("single-node ring: expected 'only-node', got %q for key %q", owner, key)
		}
	}
}
