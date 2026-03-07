package divergence

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

    switch {
    case ratio < 1.2:
        return "similar"
    case ratio < 2.0:
        return "2x slower"
    case ratio < 5.0:
        return "5x slower"
    case ratio < 10.0:
        return "10x slower"
    default:
        return "10x+ slower"
    }
}