package types

func Bool[T any](p *T) bool {
	if p == nil {
		return false
	}
	return true
}
