package divergence

func DiffStatus(live, shadow int) *StatusDiff {
	if live == shadow {
		return nil
	}
	return &StatusDiff{
		Live: live,
		Shadow: shadow,
	}
}
