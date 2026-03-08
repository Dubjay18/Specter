package divergence

import "log"

func DiffLatency(liveMs, shadowMs int64) LatencyDiff {
	delta := shadowMs - liveMs
	return LatencyDiff{
		LiveMs: liveMs,
		ShadowMs: shadowMs,
		DeltaMs: delta,
	}
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