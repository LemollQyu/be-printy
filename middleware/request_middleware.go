package middleware

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/LemollQyu/be-printy/helper"
	"github.com/LemollQyu/be-printy/infrastructure/log"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

const maxLogPayload = 2048

var sensitiveRe = regexp.MustCompile(
	`(?i)("(?:password|current_password|new_password|token|access_token|refresh_token|secret|otp|pin)"\s*:\s*)"[^"]*"`,
)

type limitedWriter struct {
	buf bytes.Buffer
	max int
}

func (w *limitedWriter) Write(p []byte) (int, error) {
	if remain := w.max - w.buf.Len(); remain > 0 {
		if len(p) > remain {
			w.buf.Write(p[:remain])
		} else {
			w.buf.Write(p)
		}
	}
	return len(p), nil
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {

		requestID := c.GetHeader(helper.HeaderRequestID)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Set(helper.RequestIDKey, requestID)
		c.Header(helper.HeaderRequestID, requestID)

		ctx := context.WithValue(c.Request.Context(), helper.CtxRequestID, requestID)
		c.Request = c.Request.WithContext(ctx) //
		method := c.Request.Method
		contentType := c.GetHeader("Content-Type")
		hasBody := method == http.MethodPost || method == http.MethodPut || method == http.MethodPatch

		payload := ""
		var captured *limitedWriter

		if hasBody && c.Request.Body != nil {
			if strings.HasPrefix(contentType, "multipart/form-data") {

				payload = "[multipart/form-data hidden]"
			} else {

				captured = &limitedWriter{max: maxLogPayload}
				c.Request.Body = struct {
					io.Reader
					io.Closer
				}{
					Reader: io.TeeReader(c.Request.Body, captured),
					Closer: c.Request.Body,
				}
			}
		}

		start := time.Now()
		c.Next()
		latency := time.Since(start)

		if captured != nil {
			payload = sensitiveRe.ReplaceAllString(captured.buf.String(), `$1"***"`)
			if captured.buf.Len() >= maxLogPayload {
				payload += "...(truncated)"
			}
		}

		status := c.Writer.Status()

		logFields := logrus.Fields{
			"request_id": requestID,
			"method":     method,
			"path":       c.Request.URL.Path,
			"query":      c.Request.URL.RawQuery,
			"status":     status,
			"latency_ms": latency.Milliseconds(),
			"client_ip":  c.ClientIP(),
			"payload":    payload,
		}
		if len(c.Errors) > 0 {
			logFields["errors"] = c.Errors.String()
		}

		entry := log.Logger.WithFields(logFields)
		switch {
		case status >= 500:
			entry.Error("Request failed")
		case status >= 400:
			entry.Warn("Request rejected")
		default:
			entry.Info("Request completed")
		}
	}
}
