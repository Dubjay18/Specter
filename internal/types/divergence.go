package types

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type DivergenceEvent struct {
	ID          string
	Timestamp   time.Time
	RequestPath string
	Method      string
	Diverged    bool
	BodyDiff    []BodyDiffEntry
	StatusDiff  *StatusDiff
	LatencyDiff LatencyDiff
}
type BodyDiffEntry struct {
	Op          string
	Path        string
	LiveValue   any
	ShadowValue any
}
type StatusDiff struct {
	Live   int
	Shadow int
}
type LatencyDiff struct {
	LiveMs   int64
	ShadowMs int64
	DeltaMs  int64
}


func (d *DivergenceEvent) String() string {
	return fmt.Sprintf("DivergenceEvent(ID=%s, Path=%s, Method=%s, Diverged=%t, StatusDiff=%v, LatencyDiff=%v, BodyDiff=%v)",
		d.ID, d.RequestPath, d.Method, d.Diverged, d.StatusDiff, d.LatencyDiff, d.BodyDiff)
}

func (d *DivergenceEvent) Marshal()	([]byte, error) {
	return json.Marshal(d)
}

func (d *DivergenceEvent) Unmarshal(data []byte) error {
	return json.Unmarshal(data, d)
}

func (l *LatencyDiff) BucketLabel() string {
    if l.ShadowMs <= l.LiveMs {
        return "faster"
    }

    ratio := float64(l.ShadowMs) / float64(l.LiveMs)
// convert to the nearest integer
	r :=int(ratio*100) / 100
	log.Printf("Latency ratio: %d (shadow %d ms vs live %d ms)", r, l.ShadowMs, l.LiveMs)
    switch {
    case r <= 1:
        return "similar"
    case r <= 2.0:
        return "2x slower"
    case r <= 5.0:
        return "5x slower"
    case r <= 10.0:
        return "10x slower"
    default:
        return "10x+ slower"
    }
}