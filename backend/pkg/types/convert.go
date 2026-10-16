// Package types provides common generic helper functions for handling typed values and pointers.
package types

// BoolFromPtr safely dereferences a boolean pointer. If the pointer is nil,
// it returns false; otherwise, it returns the underlying dereferenced bool value (*p).
//
// Parameters:
//   - p: A pointer to a boolean value (*bool).
//
// Returns:
//   - bool: *p if p != nil, false if p == nil.
func BoolFromPtr(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}

// BoolValueFromPtr returns the dereferenced boolean value if p is non-nil,
// or defaultVal if p is nil.
func BoolValueFromPtr(p *bool, defaultVal bool) bool {
	if p == nil {
		return defaultVal
	}
	return *p
}

// HasValue returns true if the generic pointer p is non-nil, and false if nil.
func HasValue[T any](p *T) bool {
	return p != nil
}
