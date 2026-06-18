package types

func BoolFromPtr[T any](p *T) bool {
	return p != nil
}
