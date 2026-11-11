package routes

import (
	"github.com/LemollQyu/be-printy/app/handler"
	"github.com/LemollQyu/be-printy/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(router *gin.Engine, printyHandler handler.PrintyHandler) {
	router.Use(middleware.RequestLogger())

	linkGroup := router.Group("/i")
	linkGroup.GET("/verify-email", printyHandler.VerifyAccountToken)

	groupV1 := router.Group("/api/v1")
	groupV1.GET("/test", printyHandler.Test)
	groupV1.POST("/register", printyHandler.Register)
	groupV1.POST("/resend-verification", printyHandler.SendTokenVerifyAccount)

	groupV1.POST("/forgot-password", printyHandler.ForgotPassword)
	groupV1.POST("/resend-otp", printyHandler.ResendOTP)
	groupV1.POST("/check-otp", printyHandler.CheckOTP)
	groupV1.POST("/reset-password", printyHandler.ResetPassword)

}
