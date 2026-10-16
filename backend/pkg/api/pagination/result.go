// Package pagination provides generic paginated container models for API responses.
package pagination

// PagedResult is a generic response wrapper encapsulating a slice of items alongside its pagination metadata.
type PagedResult[T any] struct {
	// Items contains the page slice of typed entity pointers.
	Items []*T  `json:"items"`

	// Page contains page numbers, item counts, and boundary navigation flags.
	Page  *Page `json:"page"`
}

// NewPagedResult constructs a PagedResult container from the provided item slice and page metadata.
// If items is nil, it initializes an empty slice to ensure JSON marshaling outputs [] instead of null.
// If page is nil, it falls back to an empty Page struct.
func NewPagedResult[T any](items []*T, page *Page) *PagedResult[T] {
	if items == nil {
		items = []*T{}
	}
	if page == nil {
		page = &Page{}
	}
	return &PagedResult[T]{
		Items: items,
		Page:  page,
	}
}

// PagedValueResult is a generic response wrapper for value slices ([]T) alongside its pagination metadata.
type PagedValueResult[T any] struct {
	// Items contains the page slice of typed entity values.
	Items []T   `json:"items"`

	// Page contains page numbers, item counts, and boundary navigation flags.
	Page  *Page `json:"page"`
}

// NewPagedValueResult constructs a PagedValueResult container from the provided value slice and page metadata.
// If items is nil, it initializes an empty slice to ensure JSON marshaling outputs [] instead of null.
// If page is nil, it falls back to an empty Page struct.
func NewPagedValueResult[T any](items []T, page *Page) *PagedValueResult[T] {
	if items == nil {
		items = []T{}
	}
	if page == nil {
		page = &Page{}
	}
	return &PagedValueResult[T]{
		Items: items,
		Page:  page,
	}
}
