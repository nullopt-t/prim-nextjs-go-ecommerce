package api

type FieldError struct {
	Field string `json:"field" example:"name"`
	Tags  string `json:"tags" example:"string"`
}

type MessageResponse struct {
	Message string `json:"message" example:"the request sent successfully"`
}

type UnauthorizedResponse struct {
	Message string `json:"message"`
}

type SuccessResponse struct {
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

type PaginatedResponse struct {
	Data any `json:"data,omitempty"`
	Meta any `json:"meta,omitempty"`
}

type DataResponse struct {
	Data any `json:"data"`
}

type BadReqResponse struct {
	Code    string       `json:"code,omitempty" example:"BAD_REQUEST"`
	Message string       `json:"message,omitempty" example:"field name should be string"`
	Details []FieldError `json:"details,omitempty"`
}

type ErrorResponse struct {
	Code    string `json:"code,omitempty" example:"CODE_ERROR"`
	Message string `json:"message,omitempty" example:"an error occured"`
}
