package handler

import (
	"errors"
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

	helper.Created(c, "Registration successful", gin.H{"id": id})
}
