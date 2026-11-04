package handler

import (
	"strings"

	"github.com/LemollQyu/be-printy/helper"
	models "github.com/LemollQyu/be-printy/models/authentikasi"
	"github.com/gin-gonic/gin"
)

func (h *PrintyHandler) Register(c *gin.Context) {
	var param models.RegisterParameter
	if !helper.BindJSON(c, &param) {
		return
	}

	user := &models.User{
		FullName:     strings.TrimSpace(param.FullName),
		Email:        strings.ToLower(strings.TrimSpace(param.Email)),
		Phone:        &param.Phone,
		PasswordHash: &param.Password, // pastikan di-hash di usecase!
		Role:         models.RoleCustomer,
	}

	id, err := h.PrintyUsecase.InsertUser(c.Request.Context(), user)
	if err != nil {
		helper.HandleError(c, err)
		return
	}

	helper.Created(c, "Register success", gin.H{"id": id})
}
