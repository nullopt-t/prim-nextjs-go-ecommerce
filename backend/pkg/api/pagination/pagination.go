package pagination

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

func NewPage(page, pageSize, totalItems int) *Page {
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	totalPages := 0
	if totalItems > 0 {
		totalPages = (totalItems + pageSize - 1) / pageSize
	}

	return &Page{
		Page:        page,
		PageSize:    pageSize,
		TotalItems:  totalItems,
		TotalPages:  totalPages,
		HasPrevious: page > 1,
		HasNext:     page < totalPages,
	}
}

type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

type Sort struct {
	Field string    `example:"name"`
	Order SortOrder `example:"asc"`
}

type ListQuery struct {
	Page     int `form:"page" example:"1"`
	PageSize int `form:"pageSize" example:"10"`
	Offset   int `form:"-"`

	Search string `form:"search" example:"television"`

	RawSort []string `form:"sort" example:"name,-created_at"`
	Sort    []Sort   `form:"-"`
}

type QueryOptions struct {
	DefaultPageSize int
	MaxPageSize     int
}

// Process sanitizes input, applies default fallbacks, and computes offset/sort slices.
func (q *ListQuery) Process(opt QueryOptions) *ListQuery {
	defaultSize := opt.DefaultPageSize
	if defaultSize <= 0 {
		defaultSize = DefaultPageSize
	}

	maxSize := opt.MaxPageSize
	if maxSize <= 0 {
		maxSize = MaxPageSize
	}

	if q.Page <= 0 {
		q.Page = 1
	}
	if q.PageSize <= 0 {
		q.PageSize = defaultSize
	}
	if q.PageSize > maxSize {
		q.PageSize = maxSize
	}

	q.Offset = (q.Page - 1) * q.PageSize
	q.parseSort()

	return q
}

func (q *ListQuery) parseSort() {
	q.Sort = make([]Sort, 0, len(q.RawSort))

	for _, raw := range q.RawSort {
		for s := range strings.SplitSeq(raw, ",") {
			s = strings.TrimSpace(s)
			if s == "" || s == "-" {
				continue
			}

			order := SortAsc
			field := s

			if strings.HasPrefix(s, "-") {
				order = SortDesc
				field = strings.TrimPrefix(s, "-")
			}

			if field != "" {
				q.Sort = append(q.Sort, Sort{
					Field: field,
					Order: order,
				})
			}
		}
	}
}
