package divergence

import "time"

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
