package api

type PagedResult[T any] struct {
	Items []*T  `json:"items"`
	Page  *Page `json:"page"`
}

func NewPagedResult[T any](items []*T, page *Page) *PagedResult[T] {
	return &PagedResult[T]{
		Items: items,
		Page:  page,
	}
}
