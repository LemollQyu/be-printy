package helper

import (
	"errors"
	"math"
	"net/http"
	"time"

	"github.com/LemollQyu/be-printy/infrastructure/log"
	"github.com/LemollQyu/be-printy/models"
	"github.com/gin-gonic/gin"
)

// Kode error untuk mesin/frontend
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
		RequestID: c.GetString(RequestIDKey),
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// ---------- Success ----------

func Success(c *gin.Context, status int, message string, data interface{}) {
	c.JSON(status, models.Response{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    newMeta(c),
	})
}

func OK(c *gin.Context, message string, data interface{}) { Success(c, http.StatusOK, message, data) }
func Created(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusCreated, message, data)
}

// NoContent: untuk delete/aksi yang tidak perlu mengembalikan data.
func NoContent(c *gin.Context, message string) { Success(c, http.StatusOK, message, nil) }

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
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// ---------- Error ----------

func Error(c *gin.Context, status int, code, message string, details map[string]string) {
	c.JSON(status, models.Response{
		Success: false,
		Message: message,
		Error:   &models.ErrorDetail{Code: code, Details: details},
		Meta:    newMeta(c),
	})
}

func BadRequest(c *gin.Context, message string, details map[string]string) {
	Error(c, http.StatusBadRequest, CodeBadRequest, message, details)
}
func ValidationError(c *gin.Context, details map[string]string) {
	Error(c, http.StatusUnprocessableEntity, CodeValidation, "Validasi gagal", details)
}
func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, CodeUnauthorized, message, nil)
}
func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, CodeForbidden, message, nil)
}
func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, CodeNotFound, message, nil)
}
func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, CodeConflict, message, nil)
}
func TooManyRequests(c *gin.Context) {
	Error(c, http.StatusTooManyRequests, CodeTooManyReq, "Terlalu banyak permintaan, coba lagi nanti", nil)
}
func InternalError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, CodeInternal, "Terjadi kesalahan pada server", nil)
}

// ---------- Error handler terpusat ----------

type errRule struct {
	err    error
	status int
	code   string
}

// Daftarkan semua sentinel error domain kamu di sini.
var errRules = []errRule{
	{herr.ErrPasswordRequired, http.StatusBadRequest, CodeBadRequest},
	{herr.ErrEmailExist, http.StatusConflict, CodeConflict},
	{herr.ErrPhoneExist, http.StatusConflict, CodeConflict},
	// {herr.ErrUserNotFound, http.StatusNotFound, CodeNotFound},
	// {herr.ErrInvalidCredentials, http.StatusUnauthorized, CodeUnauthorized},
}

// HandleError: cocokkan error dengan tabel; kalau tidak dikenal -> 500 + log.
func HandleError(c *gin.Context, err error) {
	for _, r := range errRules {
		if errors.Is(err, r.err) {
			Error(c, r.status, r.code, r.err.Error(), nil)
			return
		}
	}
	log.Logger.Error("unhandled error: ", err.Error())
	InternalError(c)
}
