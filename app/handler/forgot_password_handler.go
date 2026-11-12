package handler

import (
	"errors"

	"github.com/LemollQyu/be-printy/helper"
	herr "github.com/LemollQyu/be-printy/helper/error"
	"github.com/LemollQyu/be-printy/infrastructure/log"
	models "github.com/LemollQyu/be-printy/models/authentikasi"
	"github.com/gin-gonic/gin"
)

// lupa password
func (h *PrintyHandler) ForgotPassword(c *gin.Context) {
	var param models.ForgotPasswordParam

	if err := c.ShouldBindJSON(&param); err != nil {
		log.Logger.Info("Invalid parameter: ", err.Error())
		helper.BadRequest(c, "Invalid request body", nil)
		return
	}

	if err := h.PrintyUsecase.ForgotPassword(c, param.Email); err != nil {
		log.Logger.Error("printyHandler: ForgotPassword ", err.Error())
		helper.InternalError(c)
		return
	}

	helper.OK(c, "If the email is registered, an OTP code has been sent", nil)
}

func (h *PrintyHandler) ResendOTP(c *gin.Context) {
	var param models.ResendOTPParam

	if err := c.ShouldBindJSON(&param); err != nil {
		log.Logger.Info("Invalid parameter: ", err.Error())
		helper.BadRequest(c, "Invalid request body", nil)
		return
	}

	if err := h.PrintyUsecase.ResendOTP(c, param.Email); err != nil {
		log.Logger.Error("printyHandler: ResendOTP ", err.Error())
		helper.InternalError(c)
		return
	}

	helper.OK(c, "If the email is registered, an OTP code has been sent", nil)
}

func (h *PrintyHandler) CheckOTP(c *gin.Context) {
	var param models.VerifyOTPParam

	if err := c.ShouldBindJSON(&param); err != nil {
		log.Logger.Info("Invalid parameter: ", err.Error())
		helper.BadRequest(c, "Invalid request body", nil)
		return
	}

	if err := h.PrintyUsecase.CheckOTP(c, param); err != nil {
		switch {
		case errors.Is(err, herr.ErrInvalidOTP):
			helper.BadRequest(c, "Invalid or expired OTP", nil)
		case errors.Is(err, herr.ErrTooManyAttempts):
			helper.TooManyRequests(c)
		default:
			log.Logger.Error("printyHandler: CheckOTP ", err.Error())
			helper.InternalError(c)
		}
		return
	}

	helper.OK(c, "OTP valid", nil)
}

func (h *PrintyHandler) ResetPassword(c *gin.Context) {
	var param models.ResetPasswordParam

	if err := c.ShouldBindJSON(&param); err != nil {
		log.Logger.Info("Invalid parameter: ", err.Error())
		helper.BadRequest(c, "Invalid request body", nil)
		return
	}

	if err := h.PrintyUsecase.ResetPassword(c, param); err != nil {
		switch {
		case errors.Is(err, herr.ErrInvalidOTP):
			helper.BadRequest(c, "Invalid or expired OTP", nil)
		case errors.Is(err, herr.ErrTooManyAttempts):
			helper.TooManyRequests(c)
		default:
			log.Logger.Error("printyHandler: ResetPassword ", err.Error())
			helper.InternalError(c)
		}
		return
	}

	helper.OK(c, "Password has been reset", nil)
}
