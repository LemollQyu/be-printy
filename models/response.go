package models

type Meta struct {
	RequestID string `json:"request_id,omitempty"`
	Timestamp string `json:"timestamp"`
	Page      int    `json:"page,omitempty"`
	PerPage   int    `json:"per_page,omitempty"`
	Total     int64  `json:"total,omitempty"`
}

type ErrorDetail struct {
	Code    string            `json:"code"`              // untuk mesin, contoh: VALIDATION_ERROR
	Details map[string]string `json:"details,omitempty"` // contoh: {"email": "format tidak valid"}
}

type Response struct {
	Success bool         `json:"success"`
	Message string       `json:"message"`
	Data    interface{}  `json:"data,omitempty"`
	Error   *ErrorDetail `json:"error,omitempty"`
	Meta    Meta         `json:"meta"`
}
