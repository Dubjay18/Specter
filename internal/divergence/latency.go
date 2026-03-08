package divergence

import (

	"github.com/Dubjay/specter/internal/types"
)

func DiffLatency(liveMs, shadowMs int64) types.LatencyDiff {
	delta := shadowMs - liveMs
	return types.LatencyDiff{
		LiveMs: liveMs,
		ShadowMs: shadowMs,
		DeltaMs: delta,
	}
}

