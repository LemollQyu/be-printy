package middleware

import (
	"context"
	"errors"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/LemollQyu/be-printy/helper"
	herr "github.com/LemollQyu/be-printy/helper/error"
	"github.com/LemollQyu/be-printy/infrastructure/log"
	models "github.com/LemollQyu/be-printy/models/authentikasi"
)

const CtxUserKey = "user"

type SessionValidator func(ctx context.Context, rawToken string) (*models.User, error)

func AuthMiddleware(validate SessionValidator) gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.SplitN(c.GetHeader("Authorization"), " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			helper.Unauthorized(c, "Unauthorized")
			c.Abort()
			return
		}

		user, err := validate(c.Request.Context(), strings.TrimSpace(parts[1]))
		if err != nil {
			if errors.Is(err, herr.ErrUnauthorized) {
				helper.Unauthorized(c, "Unauthorized")
				c.Abort()
				return
			}
			log.Logger.Error("AuthMiddleware: ", err.Error())
			helper.InternalError(c)
			c.Abort()
			return
		}

		c.Set(CtxUserKey, user)
		c.Next()
	}
}
