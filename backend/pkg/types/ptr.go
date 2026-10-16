// Package types provides generic utility functions for pointer and value transformations.
package types

// Ptr returns a pointer to a heap-allocated copy of the passed-in value of type T.
// This is useful for constructing pointers from literal values or function returns inline
// (e.g., when populating optional struct fields).
//
// Parameters:
//   - v: The value of type T to reference.
//
// Returns:
//   - *T: A pointer referencing a copy of the given value.
func Ptr[T any](v T) *T {
	// Allocate a new copy of v and return its memory address.
	return &v
}
