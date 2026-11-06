package handler

import (
	"time"

	"github.com/LemollQyu/be-printy/app/usecase"
	"github.com/LemollQyu/be-printy/config"
	"github.com/LemollQyu/be-printy/helper"
	"github.com/gin-gonic/gin"
)

type PrintyHandler struct {
	PrintyUsecase usecase.PrintyUsecase
	ConfigApp     config.AppConfig
}

func NewPrintyHandler(printyUsecase usecase.PrintyUsecase, cfgApp config.AppConfig) *PrintyHandler {
	return &PrintyHandler{
		PrintyUsecase: printyUsecase,
		ConfigApp:     cfgApp,
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
