package handler

import (
	"errors"
	"net/http"
	"strings"

	"github.com/LemollQyu/be-printy/helper"
	herr "github.com/LemollQyu/be-printy/helper/error"
	"github.com/LemollQyu/be-printy/infrastructure/log"
	models "github.com/LemollQyu/be-printy/models/authentikasi"
	"github.com/gin-gonic/gin"
)

func (h *PrintyHandler) Register(c *gin.Context) {
	var param models.RegisterParameter

	if err := c.ShouldBindJSON(&param); err != nil {
		log.Logger.Info("Invalid parameter: ", err.Error())
		helper.BadRequest(c, "Invalid request body", nil)
		return
	}

	user := &models.User{
		FullName:     strings.TrimSpace(param.FullName),
		Email:        strings.ToLower(strings.TrimSpace(param.Email)),
		Phone:        &param.Phone,
		PasswordHash: &param.Password,
		Role:         models.RoleCustomer,
	}

	id, err := h.PrintyUsecase.InsertUser(c.Request.Context(), user)
	if err != nil {
		switch {
		case errors.Is(err, herr.ErrPasswordRequired):
			helper.BadRequest(c, err.Error(), nil)
		case errors.Is(err, herr.ErrEmailExist), errors.Is(err, herr.ErrPhoneExist):
			helper.Conflict(c, err.Error())
		default:
			log.Logger.Error("printyHandler: InsertUser ", err.Error())
			helper.InternalError(c)
		}
		return
	}

	// email verifikasi sudah dikirim dari usecase (background)
	helper.Created(c, "Registration successful, please check your email to verify your account", gin.H{"id": id})
}

func (h *PrintyHandler) VerifyAccountToken(c *gin.Context) {
	token := strings.TrimSpace(c.Query("token"))
	if token == "" {
		c.Redirect(http.StatusFound, h.ConfigApp.FrontendUrl+"/verify-email/failed?reason=invalid")
		return
	}

	err := h.PrintyUsecase.VerifyAccount(c.Request.Context(), token)
	if err != nil {
		reason := "error"

		switch {
		case errors.Is(err, herr.ErrInvalidToken):
			reason = "invalid"
		case errors.Is(err, herr.ErrTokenUsed):
			reason = "used"
		case errors.Is(err, herr.ErrTokenExpired):
			reason = "expired"
		default:
			log.Logger.Error("printyHandler: VerifyAccount ", err.Error())
		}

		log.Logger.Info("redirecting to: ", h.ConfigApp.FrontendUrl)
		c.Redirect(http.StatusFound, h.ConfigApp.FrontendUrl+"/verify-email/failed?reason="+reason)
		return
	}

	log.Logger.Info("redirecting to: ", h.ConfigApp.FrontendUrl)
	c.Redirect(http.StatusFound, h.ConfigApp.FrontendUrl+"/verify-email/success")
}

// send again verify account
func (h *PrintyHandler) SendTokenVerifyAccount(c *gin.Context) {
	var param models.ResendVerifyRequest

	if err := c.ShouldBindJSON(&param); err != nil {
		log.Logger.Info("Invalid parameter: ", err.Error())
		helper.BadRequest(c, "Invalid request body", nil)
		return
	}

	err := h.PrintyUsecase.SendTokenVerifyAccount(c.Request.Context(), param.Email)
	if err != nil {
		switch {
		case errors.Is(err, herr.ErrTooManyRequests):
			helper.BadRequest(c, "Please wait a minute before requesting another email", nil)
		case errors.Is(err, herr.ErrUserNotFound):
			helper.Conflict(c, herr.ErrUserNotFound.Error())
		case errors.Is(err, herr.ErrAccountIsVerified):
			helper.Conflict(c, herr.ErrAccountIsVerified.Error())
		default:
			log.Logger.Error("printyHandler: SendTokenVerifyAccount ", err.Error())
			helper.InternalError(c)
		}
		return
	}

	// pesan generik, tidak membocorkan apakah email terdaftar
	helper.OK(c, "If the email is registered and not yet verified, a verification link has been sent", nil)
}
