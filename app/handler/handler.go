package handler

import (
	"time"

	"github.com/LemollQyu/be-printy/app/usecase"
	"github.com/LemollQyu/be-printy/helper"
	"github.com/gin-gonic/gin"
)

type PrintyHandler struct {
	PrintyUsecase usecase.PrintyUsecase
}

func NewPrintyHandler(printyUsecase usecase.PrintyUsecase) *PrintyHandler {
	return &PrintyHandler{
		PrintyUsecase: printyUsecase,
	}
}

var startTime = time.Now()

func (h *PrintyHandler) Test(c *gin.Context) {
	helper.OK(c, "Printy service is running", gin.H{
		"status":  "ok",
		"service": "printy",
		"uptime":  time.Since(startTime).String(),
	})
}
