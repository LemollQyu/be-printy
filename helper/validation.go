package helper

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// Call once in main() / when setting up the router.
// Makes error field names use the JSON name (full_name) instead of FullName.
func RegisterValidator() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterTagNameFunc(func(f reflect.StructField) string {
			name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})
	}
}

// toJSONName converts a Go field name (e.g. "Password", "ConfirmPassword")
// to snake_case ("password", "confirm_password") for eqfield-style messages.
func toJSONName(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "numeric":
		return "must be a number"
	case "min":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("must be at least %s characters", fe.Param())
		}
		return fmt.Sprintf("must be at least %s", fe.Param())
	case "max":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("must be at most %s characters", fe.Param())
		}
		return fmt.Sprintf("must be at most %s", fe.Param())
	case "eqfield":
		return fmt.Sprintf("must match %s", toJSONName(fe.Param()))
	default:
		return "is invalid"
	}
}

// BindJSON: bind + validate + send the error response automatically.
// Returns true on success, false if an error response has already been sent.
func BindJSON(c *gin.Context, obj interface{}) bool {
	err := c.ShouldBindJSON(obj)
	if err == nil {
		return true
	}

	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		details := make(map[string]string, len(ve))
		for _, fe := range ve {
			details[fe.Field()] = validationMessage(fe)
		}
		ValidationError(c, details)
		return false
	}

	BadRequest(c, "Invalid request body", nil)
	return false
}
