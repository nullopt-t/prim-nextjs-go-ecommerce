package types

func Bool(p *any) bool {
	if p == nil {
		return false
	}
	return true
}
