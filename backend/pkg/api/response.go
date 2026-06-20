package api

type FieldError struct {
	Field string `json:"field"`
	Tags  string `json:"tags"`
}

type MessageResponse struct {
	Message string `json:"message"`
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
	Code    string       `json:"code,omitempty"`
	Message string       `json:"message,omitempty"`
	Details []FieldError `json:"details,omitempty"`
}

type ErrorResponse struct {
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}
