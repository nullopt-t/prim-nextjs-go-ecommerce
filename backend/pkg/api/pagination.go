package api

import "strings"

const (
	DefaultPageSize = 10
	MaxPageSize     = 100
)

type Page struct {
	Page        int  `json:"page" example:"1"`
	PageSize    int  `json:"page_size" example:"10"`
	TotalItems  int  `json:"total_items" example:"20"`
	TotalPages  int  `json:"total_pages" example:"2"`
	HasPrevious bool `json:"has_previous" example:"false"`
	HasNext     bool `json:"has_next" example:"true"`
}

type Sort struct {
	Field string    `example:"name"`
	Order SortOrder `example:"asc"`
}

type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

type ListQuery struct {
	Page     int `form:"page" example:"1"`
	PageSize int `form:"pageSize" example:"10"`

	Search string `form:"search" example:"television"`

	RawSort []string `form:"sort" example:"name,-age"`
	Sort    []Sort   `form:"-"`
}

type QueryOptions struct {
	DefaultPageSize int
	MaxPageSize     int
}

func (q *ListQuery) ApplyDefaults(opt QueryOptions) *ListQuery {
	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = opt.DefaultPageSize
	}
	if q.PageSize > opt.MaxPageSize {
		q.PageSize = opt.MaxPageSize
	}
	return q
}

func (q *ListQuery) Parse() {
	q.Sort = make([]Sort, 0)
	for _, raw := range q.RawSort {
		fields := strings.SplitSeq(raw, ",")
		for s := range fields {
			s = strings.TrimSpace(s)

			order := SortAsc
			field := s

			if strings.HasPrefix(s, "-") {
				order = SortDesc
				field = strings.TrimPrefix(s, "-")
			}

			q.Sort = append(q.Sort, Sort{
				Field: field,
				Order: order,
			})
		}
	}
}
