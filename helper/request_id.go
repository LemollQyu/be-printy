package helper

import "context"

const (
	HeaderRequestID = "X-Request-ID"
	RequestIDKey    = "request_id"
)

type ctxKey struct{}

var CtxRequestID = ctxKey{}

func RequestID(ctx context.Context) string {
	if v, ok := ctx.Value(CtxRequestID).(string); ok {
		return v
	}
	return ""
}
