package helper

import (
	"errors"
	"math"
	"net/http"
	"time"

	herr "github.com/LemollQyu/be-printy/helper/error"
	"github.com/LemollQyu/be-printy/infrastructure/log"
	"github.com/LemollQyu/be-printy/models"
	"github.com/gin-gonic/gin"
)

// Kode error teks (untuk mesin/frontend)
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
		Status:  status,
		Success: true,
		Message: message,
		Data:    data,
		Meta:    newMeta(c),
	})
}

func OK(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusOK, message, data) // 200
}

func Created(c *gin.Context, message string, data interface{}) {
	Success(c, http.StatusCreated, message, data) // 201
}

func NoContent(c *gin.Context, message string) {
	Success(c, http.StatusOK, message, nil) // 200 tanpa data
}

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

// ---------- Error ----------

func Error(c *gin.Context, status int, code, message string, details map[string]string) {
	c.JSON(status, models.Response{
		Status:  status,
		Success: false,
		Message: message,
		Error:   &models.ErrorDetail{Code: code, Details: details},
		Meta:    newMeta(c),
	})
}

func BadRequest(c *gin.Context, message string, details map[string]string) {
	Error(c, http.StatusBadRequest, CodeBadRequest, message, details) // 400
}

// Validasi sekarang 400 (bukan 422), tapi kode teksnya tetap VALIDATION_ERROR
// supaya frontend bisa membedakan dari BAD_REQUEST biasa.
func ValidationError(c *gin.Context, details map[string]string) {
	Error(c, http.StatusBadRequest, CodeValidation, "Validation failed", details) // 400
}

func Unauthorized(c *gin.Context, message string) {
	Error(c, http.StatusUnauthorized, CodeUnauthorized, message, nil) // 401
}

func Forbidden(c *gin.Context, message string) {
	Error(c, http.StatusForbidden, CodeForbidden, message, nil) // 403
}

func NotFound(c *gin.Context, message string) {
	Error(c, http.StatusNotFound, CodeNotFound, message, nil) // 404
}

func Conflict(c *gin.Context, message string) {
	Error(c, http.StatusConflict, CodeConflict, message, nil) // 409
}

func TooManyRequests(c *gin.Context) {
	Error(c, http.StatusTooManyRequests, CodeTooManyReq, "Too many requests, try again later", nil) // 429
}

func InternalError(c *gin.Context) {
	Error(c, http.StatusInternalServerError, CodeInternal, "A server error occurred", nil) // 500
}

// ---------- Error handler terpusat ----------

type errRule struct {
	err    error
	status int
	code   string
}

var errRules = []errRule{
	{herr.ErrPasswordRequired, http.StatusBadRequest, CodeBadRequest},
	{herr.ErrEmailExist, http.StatusConflict, CodeConflict},
	{herr.ErrPhoneExist, http.StatusConflict, CodeConflict},
	// {herr.ErrUserNotFound, http.StatusNotFound, CodeNotFound},
	// {herr.ErrInvalidCredentials, http.StatusUnauthorized, CodeUnauthorized},
}

func HandleError(c *gin.Context, err error) {
	for _, r := range errRules {
		if errors.Is(err, r.err) {
			Error(c, r.status, r.code, r.err.Error(), nil)
			return
		}
	}
	log.Logger.Error("Unhandled error: ", err.Error())
	InternalError(c)
}
