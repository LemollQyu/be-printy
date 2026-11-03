package helper

import (
	"net/http"
	"time"

	"github.com/LemollQyu/be-printy/models"
	"github.com/gin-gonic/gin"
)

// =========================
// Internal
// =========================

func newMeta(c *gin.Context) models.Meta {
	return models.Meta{
		RequestID: c.GetString(RequestIDKey), // diisi oleh middleware RequestLogger
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// =========================
// Success
// =========================

func Success(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, models.Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    newMeta(c),
	})
}

func OK(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusOK, message, data)
}

func Created(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusCreated, message, data)
}

// Paginated: sama seperti OK, tapi dengan info pagination di meta.
func Paginated(c *gin.Context, message string, data interface{}, page, perPage int, total int64) {
	meta := newMeta(c)
	meta.Page = page
	meta.PerPage = perPage
	meta.Total = total

	c.JSON(http.StatusOK, models.Response{
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
		Success: false,
		Message: message,
		Error:   &models.ErrorDetail{Code: code, Details: details},
		Meta:    newMeta(c),
	})
}

func BadRequest(c *gin.Context, message string, details map[string]string) {
	Error(c, http.StatusBadRequest, "BAD_REQUEST", message, details)
}

func ValidationError(c *gin.Context, details map[string]string) {
	Error(c, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Validasi gagal", details)
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, "UNAUTHORIZED", message, nil)
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, "FORBIDDEN", message, nil)
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, "NOT_FOUND", message, nil)
}

func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, "CONFLICT", message, nil)
}

// InternalError: jangan kirim detail error ke client, cukup di-log di server.
func InternalError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, "INTERNAL_SERVER_ERROR", "Terjadi kesalahan pada server", nil)
}
