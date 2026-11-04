package routes

import (
	"github.com/LemollQyu/be-printy/app/handler"
	"github.com/LemollQyu/be-printy/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, printyHandler handler.PrintyHandler) {
	router.Use(middleware.RequestLogger())

	groupV1 := router.Group("/api/v1")
	groupV1.GET("/test", printyHandler.Test)
	groupV1.POST("/register", printyHandler.Register)
}
