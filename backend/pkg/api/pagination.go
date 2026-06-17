package api

type Page struct {
	Page     int `json:"page"`
	PageSize int `json:"pageSize"`
	Total    int `json:"total"`
}

type PageQuery struct {
	Page     int `form:"page"`
	PageSize int `form:"pageSize"`
}
