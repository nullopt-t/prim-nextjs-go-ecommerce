package api

type Page struct {
	Page        int  `json:"page" example:"1"`
	PageSize    int  `json:"page_size" example:"10"`
	TotalItems  int  `json:"total_items" example:"20"`
	TotalPages  int  `json:"total_pages" example:"2"`
	HasPrevious bool `json:"has_previous" example:"false"`
	HasNext     bool `json:"has_next" example:"true"`
}

type ListQuery struct {
	Page     int `form:"page" example:"1"`
	PageSize int `form:"pageSize" example:"10"`

	Search string `form:"search" example:"television"`

	Sort []Sort `form:"-"`
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
