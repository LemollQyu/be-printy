package helper

import (
	"math"
	"net/http"
	"time"

	"github.com/LemollQyu/be-printy/models"
	"github.com/gin-gonic/gin"
)

// Machine-readable error codes
const (
	CodeBadRequest   = "BAD_REQUEST"
	CodeValidation   = "VALIDATION_ERROR"
	CodeUnauthorized = "UNAUTHORIZED"
	CodeForbidden    = "FORBIDDEN"
	CodeNotFound     = "NOT_FOUND"
	CodeConflict     = "CONFLICT"
	CodeTooManyReq   = "TOO_MANY_REQUESTS"
	CodeInternal     = "INTERNAL_SERVER_ERROR"
)

func newMeta(c *gin.Context) models.Meta {
	return models.Meta{
		RequestID: c.GetString(RequestIDKey), // set by RequestLogger middleware
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// =========================
// Success
// =========================

func Success(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, models.Response{
		Status:  status,
		Success: true,
		Message: message,
		Data:    data,
		Meta:    newMeta(c),
	})
}

// 200
func OK(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusOK, message, data)
}

// 201
func Created(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusCreated, message, data)
}

// 200 without data (delete / action endpoints)
func NoContent(c *gin.Context, message string) {
	Success(c, http.StatusOK, message, nil)
}

// 200 with pagination info in meta
func Paginated(c *gin.Context, message string, data interface{}, page, perPage int, total int64) {
	totalPages := 0
	if perPage > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(perPage)))
	}

	meta := newMeta(c)
	meta.Pagination = &models.Pagination{
		Page:       page,
		PerPage:    perPage,
		Total:      total,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}

	c.JSON(http.StatusOK, models.Response{
		Status:  http.StatusOK,
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// =========================
// Error
// =========================

func Error(c *gin.Context, status int, code, message string, details map[string]string) {
	c.JSON(status, models.Response{
		Status:  status,
		Success: false,
		Message: message,
		Error:   &models.ErrorDetail{Code: code, Details: details},
		Meta:    newMeta(c),
	})
}

// 400
func BadRequest(c *gin.Context, message string, details map[string]string) {
	Error(c, http.StatusBadRequest, CodeBadRequest, message, details)
}

// 400 (validation failures, distinguished by the VALIDATION_ERROR code)
func ValidationError(c *gin.Context, details map[string]string) {
	Error(c, http.StatusBadRequest, CodeValidation, "Validation failed", details)
}

// 401
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, CodeUnauthorized, message, nil)
}

// 403
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, CodeForbidden, message, nil)
}

// 404
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, CodeNotFound, message, nil)
}

// 409
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, CodeConflict, message, nil)
}

// 429
func TooManyRequests(c *gin.Context) {
	Error(c, http.StatusTooManyRequests, CodeTooManyReq, "Too many requests, please try again later", nil)
}

// 500: never send error details to the client, just log them on the server.
func InternalError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, CodeInternal, "An internal server error occurred", nil)
}
