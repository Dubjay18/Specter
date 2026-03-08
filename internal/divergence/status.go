package divergence

import "github.com/Dubjay/specter/internal/types"

func DiffStatus(live, shadow int) *types.StatusDiff {
	if live == shadow {
		return nil
	}
	return &types.StatusDiff{
		Live: live,
		Shadow: shadow,
	}
}
